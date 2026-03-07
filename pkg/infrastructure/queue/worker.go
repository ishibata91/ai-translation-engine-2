package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// Worker coordinates processing of queued jobs.
type Worker struct {
	queue           *Queue
	llmManager      llm.LLMManager
	configStore     config.Config
	secretStore     config.SecretStore
	notifier        progress.ProgressNotifier
	logger          *slog.Logger
	pollingInterval time.Duration
}

// ProcessHooks allows callers to observe progress boundaries while processing one process/task.
type ProcessHooks struct {
	OnDispatch func(current int, total int)
	OnSaving   func(current int, total int)
	OnComplete func(completed int, total int, failed int)
}

// ProcessOptions customizes worker execution without introducing slice-specific logic.
type ProcessOptions struct {
	ConfigNamespace        string
	RequireProvider        string
	UseConfigProviderModel bool
	ConfigRead             ConfigReadOptions
	Hooks                  *ProcessHooks
}

// ConfigReadOptions describes how to resolve provider/model settings from Config.
type ConfigReadOptions struct {
	Namespace           string
	DefaultProvider     string
	SelectedProviderKey string
}

// NewWorker initializes a new Worker.
func NewWorker(
	queue *Queue,
	llmManager llm.LLMManager,
	configStore config.Config,
	secretStore config.SecretStore,
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
	res, err := w.queue.db.ExecContext(ctx, "UPDATE llm_jobs SET status = ?, request_state = ?, updated_at = ? WHERE status = ?", StatusPending, RequestStatePending, time.Now().UTC(), StatusInProgress)
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
	return w.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{})
}

// ProcessProcessIDWithOptions synchronously processes queue jobs for one process/task.
func (w *Worker) ProcessProcessIDWithOptions(ctx context.Context, processID string, opts ProcessOptions) error {
	w.logger.DebugContext(ctx, "ENTER ProcessProcessID", slog.String("process_id", processID))

	// Fetch llm config
	llmConfig, err := w.fetchLLMConfig(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to fetch llm config: %w", err)
	}

	cfgNamespace := resolveConfigNamespace(opts)
	strategyStr := w.getConfigString(ctx, cfgNamespace, llm.LLMBulkStrategyKey, string(llm.BulkStrategySync))
	strategy := w.llmManager.ResolveBulkStrategy(ctx, llm.BulkStrategy(strategyStr), llmConfig.Provider)

	if strategy == llm.BulkStrategySync {
		err = w.processSync(ctx, processID, llmConfig, opts)
	} else {
		err = w.processBatch(ctx, processID, llmConfig)
	}

	w.logger.DebugContext(ctx, "EXIT ProcessProcessID", slog.String("process_id", processID), slog.Any("error", err))
	return err
}

func (w *Worker) processSync(ctx context.Context, processID string, llmConfig llm.LLMConfig, opts ProcessOptions) error {
	w.logger.DebugContext(ctx, "ENTER processSync", slog.String("process_id", processID))

	jobs, err := w.queue.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		w.logger.DebugContext(ctx, "EXIT processSync - no pending jobs", slog.String("process_id", processID))
		return nil
	}

	var resolvedProvider string
	var resolvedModel string
	if opts.UseConfigProviderModel {
		resolvedProvider = llm.NormalizeProvider(llmConfig.Provider)
		resolvedModel = llmConfig.Model
	} else {
		resolvedProvider, resolvedModel, err = w.resolveResumeProviderModel(ctx, jobs, llmConfig.Provider, llmConfig.Model)
		if err != nil {
			return err
		}
	}
	if required := llm.NormalizeProvider(opts.RequireProvider); required != "" && llm.NormalizeProvider(resolvedProvider) != required {
		return fmt.Errorf("provider %q is not allowed, required=%q", resolvedProvider, required)
	}
	llmConfig.Provider = resolvedProvider
	llmConfig.Model = resolvedModel
	if err := w.queue.UpdateProcessMetadata(ctx, processID, resolvedProvider, resolvedModel); err != nil {
		return err
	}

	var reqs []llm.Request
	for _, job := range jobs {
		if job.RequestFingerprint == "" || job.StructuredOutputSchemaVersion == "" {
			return fmt.Errorf("job %s missing required metadata fields for resume", job.ID)
		}
		var req llm.Request
		if err := json.Unmarshal([]byte(job.RequestJSON), &req); err != nil {
			return fmt.Errorf("failed to unmarshal request: %w", err)
		}
		reqs = append(reqs, req)

		// Mark as IN_PROGRESS immediately
		w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, nil)
	}
	if opts.Hooks != nil && opts.Hooks.OnDispatch != nil {
		// Dispatch phase starts here; progress increments are reported in saving phase.
		opts.Hooks.OnDispatch(0, len(jobs))
	}

	client, err := w.llmManager.GetClient(ctx, llmConfig)
	if err != nil {
		return fmt.Errorf("failed to get llm client: %w", err)
	}
	if llmConfig.Model == "" {
		return llm.ErrModelRequired
	}

	// Load model once per job process.
	instanceID := ""
	if lifecycleClient, ok := client.(llm.ModelLifecycleClient); ok {
		ctxLen := 0
		if v, ok := llmConfig.Parameters["context_length"]; ok {
			switch n := v.(type) {
			case int:
				ctxLen = n
			case float64:
				ctxLen = int(n)
			}
		}
		instanceID, err = lifecycleClient.LoadModel(ctx, llmConfig.Model, ctxLen)
		if err != nil {
			return fmt.Errorf("failed to load model: %w", err)
		}
		defer func() {
			if unloadErr := lifecycleClient.UnloadModel(context.Background(), instanceID); unloadErr != nil {
				w.logger.ErrorContext(ctx, "failed to unload model", slog.String("instance_id", instanceID), slog.String("error", unloadErr.Error()))
			}
		}()
	}

	// Wrap the client to report progress periodically
	progressNotifier := w.notifier
	// task 側で phase/current/total を通知する経路がある場合は、
	// worker 既定の件数通知を止めて進捗の分母ずれを防ぐ。
	if opts.Hooks != nil {
		progressNotifier = nil
	}
	progressClient := &progressReportingClient{
		LLMClient: client,
		notifier:  progressNotifier,
		processID: processID,
		total:     len(jobs),
		completed: new(int32), // starting at 0
		onEach: func(completed int, total int) {
			if opts.Hooks != nil && opts.Hooks.OnDispatch != nil {
				opts.Hooks.OnDispatch(completed, total)
			}
		},
	}

	// Emit initial progress event
	if progressNotifier != nil {
		progressNotifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: processID,
			Total:         len(jobs),
			Completed:     0,
			Status:        progress.StatusInProgress,
			Message:       "Starting sync processing",
		})
	}

	responses, err := llm.ExecuteBulkSync(ctx, progressClient, reqs, llmConfig.Concurrency)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			w.logger.InfoContext(ctx, "ExecuteBulkSync canceled", slog.String("error", err.Error()))
			cancelMsg := "task canceled"
			persistCtx := context.Background()
			for i, job := range jobs {
				// Keep completed responses even when the overall run is canceled.
				if i < len(responses) {
					res := responses[i]
					if res.Success {
						respJSON, _ := json.Marshal(res)
						respStr := string(respJSON)
						_ = w.queue.UpdateJob(persistCtx, job.ID, StatusCompleted, &respStr, nil, nil)
						continue
					}
					// Empty result means this request was not processed yet.
					if res.Error != "" {
						errMsg := res.Error
						_ = w.queue.UpdateJob(persistCtx, job.ID, StatusCancelled, nil, &errMsg, nil)
						continue
					}
				}
				_ = w.queue.UpdateJob(persistCtx, job.ID, StatusCancelled, nil, &cancelMsg, nil)
			}
		} else {
			w.logger.ErrorContext(ctx, "ExecuteBulkSync failed", slog.String("error", err.Error()))
		}
		return err
	}

	// Update DB with results
	failedCount := 0
	for i, res := range responses {
		jobID := jobs[i].ID
		if opts.Hooks != nil && opts.Hooks.OnSaving != nil {
			opts.Hooks.OnSaving(i+1, len(jobs))
		}
		if res.Success {
			respJSON, _ := json.Marshal(res)
			respStr := string(respJSON)
			w.queue.UpdateJob(ctx, jobID, StatusCompleted, &respStr, nil, nil)
		} else {
			errMsg := res.Error
			w.queue.UpdateJob(ctx, jobID, StatusFailed, nil, &errMsg, nil)
			failedCount++
		}
	}
	if opts.Hooks != nil && opts.Hooks.OnComplete != nil {
		opts.Hooks.OnComplete(len(jobs)-failedCount, len(jobs), failedCount)
	}

	// Emit final completion
	if progressNotifier != nil {
		status := progress.StatusCompleted
		if err != nil {
			status = progress.StatusFailed
		}
		progressNotifier.OnProgress(ctx, progress.ProgressEvent{
			CorrelationID: processID,
			Total:         len(jobs),
			Completed:     int(atomic.LoadInt32(progressClient.completed)),
			Status:        status,
			Message:       "Sync processing finished",
		})
	}

	w.logger.DebugContext(ctx, "EXIT processSync", slog.String("process_id", processID))
	return nil
}

func (w *Worker) processBatch(ctx context.Context, processID string, llmConfig llm.LLMConfig) error {
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

	var reqs []llm.Request
	for _, job := range jobs {
		var req llm.Request
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
	llm.LLMClient
	notifier  progress.ProgressNotifier
	processID string
	total     int
	completed *int32
	onEach    func(completed int, total int)
}

func (c *progressReportingClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	resp, err := c.LLMClient.Complete(ctx, req)
	comp := atomic.AddInt32(c.completed, 1)
	if c.onEach != nil {
		c.onEach(int(comp), c.total)
	}
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

func (w *Worker) fetchLLMConfig(ctx context.Context, opts ProcessOptions) (llm.LLMConfig, error) {
	read := opts.ConfigRead
	ns := resolveConfigNamespace(opts)
	defaultProvider := strings.TrimSpace(read.DefaultProvider)
	if defaultProvider == "" {
		defaultProvider = "gemini"
	}

	rawProvider := ""
	if key := strings.TrimSpace(read.SelectedProviderKey); key != "" {
		rawProvider = w.getConfigString(ctx, ns, key, "")
	}
	if rawProvider == "" {
		rawProvider = w.getConfigString(ctx, ns, llm.LLMDefaultProviderKey, "")
	}
	if rawProvider == "" {
		rawProvider = w.getConfigString(ctx, ns, "provider", defaultProvider)
	}
	provider := llm.NormalizeProvider(rawProvider)
	providerNS := ns + "." + provider
	model := w.getConfigString(ctx, ns, "model", "")
	if model == "" {
		model = w.getConfigString(ctx, providerNS, "model", "")
	}
	if model == "" {
		model = w.getConfigString(ctx, ns, provider+"_"+llm.LLMModelIDKeySuffix, "")
	}
	if model == "" {
		model = w.getConfigString(ctx, ns, provider+"_default_model", "")
	}
	// legacy compatibility
	if provider == "lmstudio" && model == "" {
		model = w.getConfigString(ctx, ns, "local_default_model", "")
		if model == "" {
			model = w.getConfigString(ctx, ns, "local-llm_default_model", "")
		}
	}

	endpoint := w.getConfigString(ctx, ns, "endpoint", "")
	if endpoint == "" {
		endpoint = w.getConfigString(ctx, providerNS, "endpoint", "")
	}
	if endpoint == "" {
		endpoint = w.getConfigString(ctx, ns, provider+"_endpoint", "")
	}
	if provider == "lmstudio" && endpoint == "" {
		endpoint = w.getConfigString(ctx, ns, "local_endpoint", "")
		if endpoint == "" {
			endpoint = w.getConfigString(ctx, ns, "local-llm_endpoint", "")
		}
	}

	apiKey := ""
	if provider != "lmstudio" {
		apiKey = w.getConfigString(ctx, ns, "api_key", "")
		if apiKey == "" {
			apiKey = w.getConfigString(ctx, providerNS, "api_key", "")
		}
	}
	if provider != "lmstudio" && apiKey == "" {
		if val, err := w.secretStore.GetSecret(ctx, ns, provider+"_api_key"); err == nil {
			apiKey = val
		}
	}

	strConcurrency := w.getConfigString(ctx, ns, llm.LLMSyncConcurrencyKeySuffix+"."+provider, "")
	var concurrency int
	if strConcurrency != "" {
		fmt.Sscanf(strConcurrency, "%d", &concurrency)
	}
	if concurrency <= 0 {
		concurrency = llm.DefaultConcurrency(provider)
	}
	strContextLength := w.getConfigString(ctx, ns, "context_length", "")
	if strContextLength == "" {
		strContextLength = w.getConfigString(ctx, providerNS, "context_length", "")
	}
	var contextLength int
	if strContextLength != "" {
		fmt.Sscanf(strContextLength, "%d", &contextLength)
	}
	if strings.TrimSpace(model) == "" {
		return llm.LLMConfig{}, llm.ErrModelRequired
	}

	params := map[string]interface{}{}
	if contextLength > 0 {
		params["context_length"] = contextLength
	}

	return llm.LLMConfig{
		Provider:    provider,
		APIKey:      apiKey,
		Endpoint:    endpoint,
		Model:       model,
		Parameters:  params,
		Concurrency: concurrency,
	}, nil
}

func (w *Worker) resolveResumeProviderModel(ctx context.Context, jobs []JobRequest, defaultProvider, defaultModel string) (string, string, error) {
	allEmpty := true
	for _, job := range jobs {
		if job.Provider != "" || job.Model != "" {
			allEmpty = false
			break
		}
	}
	if allEmpty {
		if defaultProvider == "" || defaultModel == "" {
			return "", "", llm.ErrModelRequired
		}
		return llm.NormalizeProvider(defaultProvider), defaultModel, nil
	}

	var provider string
	var model string
	for _, job := range jobs {
		if job.Provider == "" || job.Model == "" {
			return "", "", fmt.Errorf("job %s missing provider/model metadata; resume is not allowed", job.ID)
		}
		if provider == "" {
			provider = llm.NormalizeProvider(job.Provider)
			model = job.Model
			continue
		}
		if provider != llm.NormalizeProvider(job.Provider) || model != job.Model {
			return "", "", fmt.Errorf("inconsistent provider/model metadata in process jobs")
		}
	}
	return provider, model, nil
}

func (w *Worker) getConfigString(ctx context.Context, ns, key, defaultVal string) string {
	val, err := w.configStore.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

func resolveConfigNamespace(opts ProcessOptions) string {
	ns := strings.TrimSpace(opts.ConfigRead.Namespace)
	if ns == "" {
		ns = strings.TrimSpace(opts.ConfigNamespace)
	}
	if ns == "" {
		ns = llm.LLMConfigNamespace
	}
	return ns
}
