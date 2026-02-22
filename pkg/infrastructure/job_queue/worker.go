package job_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// Worker coordinates processing of queued jobs.
type Worker struct {
	queue           *Queue
	llmManager      llm_client.LLMManager
	configStore     config_store.ConfigStore
	secretStore     config_store.SecretStore
	notifier        progress.ProgressNotifier
	logger          *slog.Logger
	pollingInterval time.Duration
}

// NewWorker initializes a new Worker.
func NewWorker(
	queue *Queue,
	llmManager llm_client.LLMManager,
	configStore config_store.ConfigStore,
	secretStore config_store.SecretStore,
	notifier progress.ProgressNotifier,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		queue:           queue,
		llmManager:      llmManager,
		configStore:     configStore,
		secretStore:     secretStore,
		notifier:        notifier,
		logger:          logger.With("slice", "JobQueue"),
		pollingInterval: 30 * time.Second, // Default
	}
}

// SetPollingInterval overrides the default polling interval (useful for tests).
func (w *Worker) SetPollingInterval(d time.Duration) {
	w.pollingInterval = d
}

// Recover should be called at startup to reset any IN_PROGRESS jobs to PENDING.
func (w *Worker) Recover(ctx context.Context) error {
	w.logger.DebugContext(ctx, "ENTER Recover")
	res, err := w.queue.db.ExecContext(ctx, "UPDATE llm_jobs SET status = ?, updated_at = ? WHERE status = ?", StatusPending, time.Now().UTC(), StatusInProgress)
	if err != nil {
		return fmt.Errorf("recover fail: %w", err)
	}
	affected, _ := res.RowsAffected()
	w.logger.DebugContext(ctx, "EXIT Recover", slog.Int("recovered", int(affected)))
	return nil
}

// ProcessProcessID synchronously processes all PENDING jobs for a given processID in the background.
// Note: This method blocks until the process finishes or polling exhausts. Callers should
// execute it in a goroutine if they do not wish to block.
func (w *Worker) ProcessProcessID(ctx context.Context, processID string) error {
	w.logger.DebugContext(ctx, "ENTER ProcessProcessID", slog.String("process_id", processID))

	// Fetch llm config
	llmConfig, err := w.fetchLLMConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch llm config: %w", err)
	}

	strategyStr := w.getConfigString(ctx, "llm", llm_client.LLMBulkStrategyKey, string(llm_client.BulkStrategySync))
	strategy := w.llmManager.ResolveBulkStrategy(ctx, llm_client.BulkStrategy(strategyStr), llmConfig.Provider)

	if strategy == llm_client.BulkStrategySync {
		err = w.processSync(ctx, processID, llmConfig)
	} else {
		err = w.processBatch(ctx, processID, llmConfig)
	}

	w.logger.DebugContext(ctx, "EXIT ProcessProcessID", slog.String("process_id", processID), slog.Any("error", err))
	return err
}

func (w *Worker) processSync(ctx context.Context, processID string, llmConfig llm_client.LLMConfig) error {
	w.logger.DebugContext(ctx, "ENTER processSync", slog.String("process_id", processID))

	jobs, err := w.queue.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		w.logger.DebugContext(ctx, "EXIT processSync - no pending jobs", slog.String("process_id", processID))
		return nil
	}

	var reqs []llm_client.Request
	for _, job := range jobs {
		var req llm_client.Request
		if err := json.Unmarshal([]byte(job.RequestJSON), &req); err != nil {
			return fmt.Errorf("failed to unmarshal request: %w", err)
		}
		reqs = append(reqs, req)

		// Mark as IN_PROGRESS immediately
		w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, nil)
	}

	client, err := w.llmManager.GetClient(ctx, llmConfig)
	if err != nil {
		return fmt.Errorf("failed to get llm client: %w", err)
	}

	// Wrap the client to report progress periodically
	progressClient := &progressReportingClient{
		LLMClient: client,
		notifier:  w.notifier,
		processID: processID,
		total:     len(jobs),
		completed: new(int32), // starting at 0
	}

	// Emit initial progress event
	if w.notifier != nil {
		w.notifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: processID,
			Total:         len(jobs),
			Completed:     0,
			Status:        progress.StatusInProgress,
			Message:       "Starting sync processing",
		})
	}

	responses, err := llm_client.ExecuteBulkSync(ctx, progressClient, reqs, llmConfig.Concurrency)
	if err != nil {
		w.logger.ErrorContext(ctx, "ExecuteBulkSync failed", slog.String("error", err.Error()))
		// It only fails on context cancellation.
		// State updates for individual successes still occurred so we preserve DB.
	}

	// Update DB with results
	for i, res := range responses {
		jobID := jobs[i].ID
		if res.Success {
			respJSON, _ := json.Marshal(res)
			respStr := string(respJSON)
			w.queue.UpdateJob(ctx, jobID, StatusCompleted, &respStr, nil, nil)
		} else {
			errMsg := res.Error
			w.queue.UpdateJob(ctx, jobID, StatusFailed, nil, &errMsg, nil)
		}
	}

	// Emit final completion
	if w.notifier != nil {
		status := progress.StatusCompleted
		if err != nil {
			status = progress.StatusFailed
		}
		w.notifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: processID,
			Total:         len(jobs),
			Completed:     int(atomic.LoadInt32(progressClient.completed)),
			Status:        status,
			Message:       "Sync processing finished",
		})
	}

	w.logger.DebugContext(ctx, "EXIT processSync", slog.String("process_id", processID))
	return err
}

func (w *Worker) processBatch(ctx context.Context, processID string, llmConfig llm_client.LLMConfig) error {
	w.logger.DebugContext(ctx, "ENTER processBatch", slog.String("process_id", processID))

	jobs, err := w.queue.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil
	}

	batchClient, err := w.llmManager.GetBatchClient(ctx, llmConfig)
	if err != nil {
		return fmt.Errorf("failed to get batch client: %w", err)
	}

	var reqs []llm_client.Request
	for _, job := range jobs {
		var req llm_client.Request
		if err := json.Unmarshal([]byte(job.RequestJSON), &req); err != nil {
			return fmt.Errorf("failed to unmarshal request: %w", err)
		}
		reqs = append(reqs, req)
	}

	batchJobID, err := batchClient.SubmitBatch(ctx, reqs)
	if err != nil {
		return fmt.Errorf("submit batch failed: %w", err)
	}

	batchIDJSON, _ := json.Marshal(batchJobID)
	batchIDString := string(batchIDJSON)

	for _, job := range jobs {
		w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, &batchIDString)
	}

	if w.notifier != nil {
		w.notifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: processID,
			Total:         100, // percentage maybe?
			Completed:     0,
			Status:        progress.StatusInProgress,
			Message:       fmt.Sprintf("Batch job %s submitted, polling...", batchJobID.ID),
		})
	}

	// Poll loop
	ticker := time.NewTicker(w.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := batchClient.GetBatchStatus(ctx, batchJobID)
			if err != nil {
				w.logger.WarnContext(ctx, "Failed to get batch status", slog.String("error", err.Error()))
				continue
			}

			if w.notifier != nil {
				w.notifier.OnProgress(ctx, progress.ProgressEvent{
					CorrelationID: processID,
					Total:         100,
					Completed:     int(status.Progress * 100),
					Status:        progress.StatusInProgress,
					Message:       fmt.Sprintf("Batch status: %s", status.State),
				})
			}

			if status.State == "COMPLETED" || status.State == "FAILED" || status.State == "CANCELLED" {
				// Fetch results
				results, err := batchClient.GetBatchResults(ctx, batchJobID)
				if err != nil {
					return fmt.Errorf("get batch results failed: %w", err)
				}

				// Update DB
				for i, res := range results {
					if i >= len(jobs) {
						break
					}
					jobID := jobs[i].ID
					if res.Success {
						respJSON, _ := json.Marshal(res)
						respStr := string(respJSON)
						w.queue.UpdateJob(ctx, jobID, StatusCompleted, &respStr, nil, nil)
					} else {
						errMsg := res.Error
						w.queue.UpdateJob(ctx, jobID, StatusFailed, nil, &errMsg, nil)
					}
				}

				if w.notifier != nil {
					msgStatus := progress.StatusCompleted
					if status.State != "COMPLETED" {
						msgStatus = progress.StatusFailed
					}
					w.notifier.OnProgress(ctx, progress.ProgressEvent{
						CorrelationID: processID,
						Total:         len(jobs),
						Completed:     len(jobs),
						Status:        msgStatus,
						Message:       "Batch finished",
					})
				}
				w.logger.DebugContext(ctx, "EXIT processBatch", slog.String("process_id", processID))
				return nil
			}
		}
	}
}

// progressReportingClient wraps LLMClient to emit progress events.
type progressReportingClient struct {
	llm_client.LLMClient
	notifier  progress.ProgressNotifier
	processID string
	total     int
	completed *int32
}

func (c *progressReportingClient) Complete(ctx context.Context, req llm_client.Request) (llm_client.Response, error) {
	resp, err := c.LLMClient.Complete(ctx, req)
	comp := atomic.AddInt32(c.completed, 1)
	if c.notifier != nil {
		c.notifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: c.processID,
			Total:         c.total,
			Completed:     int(comp),
			Status:        progress.StatusInProgress,
			Message:       "Processing...",
		})
	}
	return resp, err
}

func (w *Worker) fetchLLMConfig(ctx context.Context) (llm_client.LLMConfig, error) {
	provider := w.getConfigString(ctx, llm_client.LLMConfigNamespace, "default_provider", "gemini")
	model := w.getConfigString(ctx, llm_client.LLMConfigNamespace, provider+"_default_model", "")
	endpoint := w.getConfigString(ctx, llm_client.LLMConfigNamespace, provider+"_endpoint", "")

	apiKey := ""
	if val, err := w.secretStore.GetSecret(ctx, llm_client.LLMConfigNamespace, provider+"_api_key"); err == nil {
		apiKey = val
	}

	strConcurrency := w.getConfigString(ctx, llm_client.LLMConfigNamespace, llm_client.LLMSyncConcurrencyKeySuffix+"."+provider, "")
	var concurrency int
	if strConcurrency != "" {
		fmt.Sscanf(strConcurrency, "%d", &concurrency)
	}
	if concurrency <= 0 {
		concurrency = llm_client.DefaultConcurrency(provider)
	}

	return llm_client.LLMConfig{
		Provider:    provider,
		APIKey:      apiKey,
		Endpoint:    endpoint,
		Model:       model,
		Concurrency: concurrency,
	}, nil
}

func (w *Worker) getConfigString(ctx context.Context, ns, key, defaultVal string) string {
	val, err := w.configStore.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}
