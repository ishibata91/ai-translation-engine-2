package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/configaccess"
)

type configReader interface {
	Get(ctx context.Context, namespace string, key string) (string, error)
}

type secretReader interface {
	GetSecret(ctx context.Context, namespace string, key string) (string, error)
}

// Worker coordinates processing of queued jobs.
type Worker struct {
	queue           *Queue
	llmManager      gatewayllm.LLMManager
	configStore     configReader
	secretStore     secretReader
	configAccessor  *configaccess.TypedAccessor
	notifier        runtimeprogress.ProgressNotifier
	logger          *slog.Logger
	pollingInterval time.Duration
}

// ProcessHooks allows callers to observe progress boundaries while processing one process/task.
type ProcessHooks struct {
	OnDispatch       func(current int, total int)
	OnSaving         func(current int, total int)
	OnComplete       func(completed int, total int, failed int)
	OnBatchSubmitted func(batchJobID string, reconnected bool)
	OnBatchPolling   func(state string, progress float32)
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

// ExecutionProfile is the resolved runtime execution mode for one process.
type ExecutionProfile struct {
	Provider              string
	Model                 string
	RequestedBulkStrategy gatewayllm.BulkStrategy
	BulkStrategy          gatewayllm.BulkStrategy
}

// NewWorker initializes a new Worker.
func NewWorker(
	queue *Queue,
	llmManager gatewayllm.LLMManager,
	configStore configReader,
	secretStore secretReader,
	notifier runtimeprogress.ProgressNotifier,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		queue:           queue,
		llmManager:      llmManager,
		configStore:     configStore,
		secretStore:     secretStore,
		configAccessor:  configaccess.NewTypedAccessor(configStore),
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
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("recover rows affected failed: %w", err)
	}
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
	strategy := w.resolveBulkStrategy(ctx, cfgNamespace, llmConfig.Provider)

	if strategy == gatewayllm.BulkStrategySync {
		err = w.processSync(ctx, processID, llmConfig, opts)
	} else {
		err = w.processBatch(ctx, processID, llmConfig, opts)
	}

	w.logger.DebugContext(ctx, "EXIT ProcessProcessID", slog.String("process_id", processID), slog.Any("error", err))
	if err != nil {
		return fmt.Errorf("process queue jobs process_id=%s: %w", processID, err)
	}
	return nil
}

// ResolveExecutionProfile resolves provider/model/strategy for one execution attempt.
func (w *Worker) ResolveExecutionProfile(ctx context.Context, opts ProcessOptions) (ExecutionProfile, error) {
	llmConfig, err := w.fetchLLMConfig(ctx, opts)
	if err != nil {
		return ExecutionProfile{}, fmt.Errorf("resolve execution profile fetch config: %w", err)
	}

	ns := resolveConfigNamespace(opts)
	requested := w.resolveConfiguredBulkStrategy(ctx, ns, llmConfig.Provider)
	resolved := w.llmManager.ResolveBulkStrategy(ctx, requested, llmConfig.Provider)

	return ExecutionProfile{
		Provider:              llmConfig.Provider,
		Model:                 llmConfig.Model,
		RequestedBulkStrategy: requested,
		BulkStrategy:          resolved,
	}, nil
}

// ValidateExecutionProfile validates unsupported profile combinations before runtime execution.
func (w *Worker) ValidateExecutionProfile(ctx context.Context, opts ProcessOptions) (ExecutionProfile, error) {
	profile, err := w.ResolveExecutionProfile(ctx, opts)
	if err != nil {
		return ExecutionProfile{}, fmt.Errorf("validate execution profile resolve: %w", err)
	}

	if profile.RequestedBulkStrategy == gatewayllm.BulkStrategyBatch && !gatewayllm.ProviderSupportsBatch(profile.Provider) {
		return ExecutionProfile{}, fmt.Errorf("batch execution is not supported for provider=%s", profile.Provider)
	}
	return profile, nil
}

func (w *Worker) processSync(ctx context.Context, processID string, llmConfig gatewayllm.LLMConfig, opts ProcessOptions) error {
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
		resolvedProvider = gatewayllm.NormalizeProvider(llmConfig.Provider)
		resolvedModel = llmConfig.Model
	} else {
		resolvedProvider, resolvedModel, err = w.resolveResumeProviderModel(ctx, jobs, llmConfig.Provider, llmConfig.Model)
		if err != nil {
			return fmt.Errorf("resolve resume provider/model process_id=%s: %w", processID, err)
		}
	}
	if required := gatewayllm.NormalizeProvider(opts.RequireProvider); required != "" && gatewayllm.NormalizeProvider(resolvedProvider) != required {
		return fmt.Errorf("provider %q is not allowed, required=%q", resolvedProvider, required)
	}
	llmConfig.Provider = resolvedProvider
	llmConfig.Model = resolvedModel
	if err := w.queue.UpdateProcessMetadata(ctx, processID, resolvedProvider, resolvedModel); err != nil {
		return fmt.Errorf("update process metadata process_id=%s: %w", processID, err)
	}

	var reqs []gatewayllm.Request
	for _, job := range jobs {
		if job.RequestFingerprint == "" || job.StructuredOutputSchemaVersion == "" {
			return fmt.Errorf("job %s missing required metadata fields for resume", job.ID)
		}
		var req gatewayllm.Request
		if err := json.Unmarshal([]byte(job.RequestJSON), &req); err != nil {
			return fmt.Errorf("failed to unmarshal request: %w", err)
		}
		reqs = append(reqs, req)

		// Mark as IN_PROGRESS immediately
		if err := w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, nil); err != nil {
			return fmt.Errorf("failed to mark job %s in progress: %w", job.ID, err)
		}
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
		return gatewayllm.ErrModelRequired
	}

	// Load model once per job process.
	instanceID := ""
	if lifecycleClient, ok := client.(gatewayllm.ModelLifecycleClient); ok {
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
			unloadCtx := context.WithoutCancel(ctx)
			if unloadErr := lifecycleClient.UnloadModel(unloadCtx, instanceID); unloadErr != nil {
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
		progressNotifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
			CorrelationID: processID,
			Total:         len(jobs),
			Completed:     0,
			Status:        runtimeprogress.StatusInProgress,
			Message:       "Starting sync processing",
		})
	}

	responses, err := gatewayllm.ExecuteBulkSync(ctx, progressClient, reqs, llmConfig.Concurrency)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			w.logger.InfoContext(ctx, "ExecuteBulkSync canceled", slog.String("error", err.Error()))
			cancelMsg := "task canceled"
			persistCtx := context.WithoutCancel(ctx)
			for i, job := range jobs {
				// Keep completed responses even when the overall run is canceled.
				if i < len(responses) {
					res := responses[i]
					if res.Success {
						respJSON, marshalErr := json.Marshal(res)
						if marshalErr != nil {
							w.logger.ErrorContext(persistCtx, "failed to marshal completed response during cancellation",
								slog.String("job_id", job.ID),
								slog.String("error", marshalErr.Error()),
							)
							continue
						}
						respStr := string(respJSON)
						if updateErr := w.queue.UpdateJob(persistCtx, job.ID, StatusCompleted, &respStr, nil, nil); updateErr != nil {
							w.logger.ErrorContext(persistCtx, "failed to persist completed job during cancellation",
								slog.String("job_id", job.ID),
								slog.String("error", updateErr.Error()),
							)
						}
						continue
					}
					// Empty result means this request was not processed yet.
					if res.Error != "" {
						errMsg := res.Error
						if updateErr := w.queue.UpdateJob(persistCtx, job.ID, StatusCancelled, nil, &errMsg, nil); updateErr != nil {
							w.logger.ErrorContext(persistCtx, "failed to persist errored cancellation status",
								slog.String("job_id", job.ID),
								slog.String("error", updateErr.Error()),
							)
						}
						continue
					}
				}
				if updateErr := w.queue.UpdateJob(persistCtx, job.ID, StatusCancelled, nil, &cancelMsg, nil); updateErr != nil {
					w.logger.ErrorContext(persistCtx, "failed to persist canceled job status",
						slog.String("job_id", job.ID),
						slog.String("error", updateErr.Error()),
					)
				}
			}
		} else {
			w.logger.ErrorContext(ctx, "ExecuteBulkSync failed", slog.String("error", err.Error()))
		}
		return fmt.Errorf("execute bulk sync: %w", err)
	}

	// Update DB with results
	failedCount := 0
	for i, res := range responses {
		jobID := jobs[i].ID
		if opts.Hooks != nil && opts.Hooks.OnSaving != nil {
			opts.Hooks.OnSaving(i+1, len(jobs))
		}
		if res.Success {
			respJSON, marshalErr := json.Marshal(res)
			if marshalErr != nil {
				return fmt.Errorf("marshal completed response job_id=%s: %w", jobID, marshalErr)
			}
			respStr := string(respJSON)
			if err := w.queue.UpdateJob(ctx, jobID, StatusCompleted, &respStr, nil, nil); err != nil {
				return fmt.Errorf("failed to store completed job %s: %w", jobID, err)
			}
		} else {
			errMsg := res.Error
			if err := w.queue.UpdateJob(ctx, jobID, StatusFailed, nil, &errMsg, nil); err != nil {
				return fmt.Errorf("failed to store failed job %s: %w", jobID, err)
			}
			failedCount++
		}
	}
	if opts.Hooks != nil && opts.Hooks.OnComplete != nil {
		opts.Hooks.OnComplete(len(jobs)-failedCount, len(jobs), failedCount)
	}

	// Emit final completion
	if progressNotifier != nil {
		status := runtimeprogress.StatusCompleted
		if err != nil {
			status = runtimeprogress.StatusFailed
		}
		progressNotifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
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

func (w *Worker) processBatch(ctx context.Context, processID string, llmConfig gatewayllm.LLMConfig, opts ProcessOptions) error {
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

	progressNotifier := w.notifier
	if opts.Hooks != nil {
		progressNotifier = nil
	}

	batchJobID, resumedFromExisting, err := w.resolveBatchJob(ctx, processID, llmConfig.Provider, jobs, batchClient)
	if err != nil {
		return err
	}

	if opts.Hooks != nil && opts.Hooks.OnBatchSubmitted != nil {
		opts.Hooks.OnBatchSubmitted(batchJobID.ID, resumedFromExisting)
	}

	if progressNotifier != nil {
		message := fmt.Sprintf("Batch job %s submitted, polling...", batchJobID.ID)
		if resumedFromExisting {
			message = fmt.Sprintf("Batch job %s reconnected, polling...", batchJobID.ID)
		}
		progressNotifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
			CorrelationID: processID,
			Total:         100,
			Completed:     0,
			Status:        runtimeprogress.StatusInProgress,
			Message:       message,
		})
	}

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

			if opts.Hooks != nil && opts.Hooks.OnBatchPolling != nil {
				opts.Hooks.OnBatchPolling(string(status.State), status.Progress)
			}
			if progressNotifier != nil {
				progressNotifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
					CorrelationID: processID,
					Total:         100,
					Completed:     int(status.Progress * 100),
					Status:        runtimeprogress.StatusInProgress,
					Message:       fmt.Sprintf("Batch status: %s", status.State),
				})
			}

			if !isBatchTerminalState(status.State) {
				continue
			}

			results, err := batchClient.GetBatchResults(ctx, batchJobID)
			if err != nil {
				return fmt.Errorf("get batch results failed: %w", err)
			}

			completedCount, failedCount, err := w.applyBatchResults(ctx, jobs, results, opts)
			if err != nil {
				return err
			}

			if opts.Hooks != nil && opts.Hooks.OnComplete != nil {
				opts.Hooks.OnComplete(completedCount, len(jobs), failedCount)
			}
			if progressNotifier != nil {
				msgStatus := runtimeprogress.StatusCompleted
				if status.State == gatewayllm.BatchStateFailed || status.State == gatewayllm.BatchStateCancelled {
					msgStatus = runtimeprogress.StatusFailed
				}
				progressNotifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
					CorrelationID: processID,
					Total:         len(jobs),
					Completed:     completedCount + failedCount,
					Status:        msgStatus,
					Message:       "Batch finished",
				})
			}

			w.logger.DebugContext(ctx, "EXIT processBatch",
				slog.String("process_id", processID),
				slog.Int("completed", completedCount),
				slog.Int("failed", failedCount),
			)
			return nil
		}
	}
}

func (w *Worker) resolveBatchJob(
	ctx context.Context,
	processID string,
	provider string,
	jobs []JobRequest,
	batchClient gatewayllm.BatchClient,
) (gatewayllm.BatchJobID, bool, error) {
	existingBatchID, hasExisting, err := findExistingBatchJobID(jobs, provider)
	if err != nil {
		return gatewayllm.BatchJobID{}, false, fmt.Errorf("resolve existing batch job id process_id=%s: %w", processID, err)
	}

	if hasExisting {
		w.logger.InfoContext(ctx, "reconnected existing batch job",
			slog.String("process_id", processID),
			slog.String("batch_job_id", existingBatchID.ID),
			slog.Int("job_count", len(jobs)),
		)
		batchIDString, err := marshalBatchJobID(existingBatchID)
		if err != nil {
			return gatewayllm.BatchJobID{}, false, fmt.Errorf("marshal existing batch job id process_id=%s: %w", processID, err)
		}
		for _, job := range jobs {
			jobBatchID := job.BatchJobID
			if jobBatchID == nil || strings.TrimSpace(*jobBatchID) == "" {
				jobBatchID = &batchIDString
			}
			if err := w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, jobBatchID); err != nil {
				return gatewayllm.BatchJobID{}, false, fmt.Errorf("mark reconnected batch job in progress job_id=%s: %w", job.ID, err)
			}
		}
		return existingBatchID, true, nil
	}

	reqs := make([]gatewayllm.Request, 0, len(jobs))
	for _, job := range jobs {
		var req gatewayllm.Request
		if err := json.Unmarshal([]byte(job.RequestJSON), &req); err != nil {
			return gatewayllm.BatchJobID{}, false, fmt.Errorf("failed to unmarshal request job_id=%s: %w", job.ID, err)
		}
		req = withBatchCorrelationMetadata(req, job)
		w.logger.DebugContext(ctx, "prepared batch correlation metadata",
			slog.String("process_id", processID),
			slog.String("job_id", job.ID),
			slog.Int("queue_request_seq", job.ResumeCursor),
		)
		reqs = append(reqs, req)
	}

	batchJobID, err := batchClient.SubmitBatch(ctx, reqs)
	if err != nil {
		return gatewayllm.BatchJobID{}, false, fmt.Errorf("submit batch failed: %w", err)
	}

	batchIDString, err := marshalBatchJobID(batchJobID)
	if err != nil {
		return gatewayllm.BatchJobID{}, false, fmt.Errorf("marshal batch job id process_id=%s: %w", processID, err)
	}
	for _, job := range jobs {
		if err := w.queue.UpdateJob(ctx, job.ID, StatusInProgress, nil, nil, &batchIDString); err != nil {
			return gatewayllm.BatchJobID{}, false, fmt.Errorf("failed to mark batch job %s in progress: %w", job.ID, err)
		}
	}

	return batchJobID, false, nil
}

func withBatchCorrelationMetadata(req gatewayllm.Request, job JobRequest) gatewayllm.Request {
	metadata := make(map[string]interface{}, len(req.Metadata)+2)
	for key, value := range req.Metadata {
		metadata[key] = value
	}
	metadata[gatewayllm.BatchMetadataQueueJobIDKey] = job.ID
	metadata[gatewayllm.BatchMetadataQueueRequestSeqKey] = job.ResumeCursor
	req.Metadata = metadata
	return req
}

func extractBatchQueueJobID(metadata map[string]interface{}) (string, bool) {
	if len(metadata) == 0 {
		return "", false
	}
	raw, exists := metadata[gatewayllm.BatchMetadataQueueJobIDKey]
	if !exists {
		return "", false
	}

	switch value := raw.(type) {
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return "", false
		}
		return trimmed, true
	default:
		trimmed := strings.TrimSpace(fmt.Sprintf("%v", value))
		if trimmed == "" || trimmed == "<nil>" {
			return "", false
		}
		return trimmed, true
	}
}

func (w *Worker) applySingleBatchResult(ctx context.Context, job JobRequest, res gatewayllm.Response) (bool, error) {
	if res.Success {
		respJSON, marshalErr := json.Marshal(res)
		if marshalErr != nil {
			return false, fmt.Errorf("marshal batch result job_id=%s: %w", job.ID, marshalErr)
		}
		respStr := string(respJSON)
		if err := w.queue.UpdateJob(ctx, job.ID, StatusCompleted, &respStr, nil, nil); err != nil {
			return false, fmt.Errorf("store batch result job_id=%s: %w", job.ID, err)
		}
		return true, nil
	}

	errMsg := strings.TrimSpace(res.Error)
	if errMsg == "" {
		errMsg = "batch request failed"
	}
	if err := w.queue.UpdateJob(ctx, job.ID, StatusFailed, nil, &errMsg, nil); err != nil {
		return false, fmt.Errorf("store batch failure job_id=%s: %w", job.ID, err)
	}
	return false, nil
}

func (w *Worker) applyBatchResults(ctx context.Context, jobs []JobRequest, results []gatewayllm.Response, opts ProcessOptions) (int, int, error) {
	completedCount := 0
	failedCount := 0

	jobsByID := make(map[string]JobRequest, len(jobs))
	for _, job := range jobs {
		jobsByID[job.ID] = job
	}

	resultsByJobID := make(map[string]gatewayllm.Response, len(results))
	duplicateJobIDs := make(map[string]struct{})
	fallbackResults := make([]gatewayllm.Response, 0)
	unknownIDCount := 0

	for _, res := range results {
		queueJobID, hasJobID := extractBatchQueueJobID(res.Metadata)
		if !hasJobID {
			fallbackResults = append(fallbackResults, res)
			continue
		}

		if _, exists := jobsByID[queueJobID]; !exists {
			unknownIDCount++
			w.logger.WarnContext(ctx, "batch result has unknown queue_job_id",
				slog.String("queue_job_id", queueJobID),
			)
			continue
		}

		if _, duplicated := duplicateJobIDs[queueJobID]; duplicated {
			continue
		}
		if _, exists := resultsByJobID[queueJobID]; exists {
			delete(resultsByJobID, queueJobID)
			duplicateJobIDs[queueJobID] = struct{}{}
			w.logger.WarnContext(ctx, "batch result has duplicate queue_job_id",
				slog.String("queue_job_id", queueJobID),
			)
			continue
		}
		resultsByJobID[queueJobID] = res
	}

	if len(fallbackResults) > 0 {
		w.logger.WarnContext(ctx, "batch results missing queue_job_id, using limited index fallback",
			slog.Int("fallback_result_count", len(fallbackResults)),
		)
	}

	fallbackIndex := 0
	for i, job := range jobs {
		if opts.Hooks != nil && opts.Hooks.OnSaving != nil {
			opts.Hooks.OnSaving(i+1, len(jobs))
		}

		if _, duplicated := duplicateJobIDs[job.ID]; duplicated {
			errMsg := "duplicate queue_job_id in batch results"
			if err := w.queue.UpdateJob(ctx, job.ID, StatusFailed, nil, &errMsg, nil); err != nil {
				return 0, 0, fmt.Errorf("store duplicate queue_job_id failure job_id=%s: %w", job.ID, err)
			}
			failedCount++
			continue
		}

		if res, exists := resultsByJobID[job.ID]; exists {
			succeeded, err := w.applySingleBatchResult(ctx, job, res)
			if err != nil {
				return 0, 0, err
			}
			if succeeded {
				completedCount++
			} else {
				failedCount++
			}
			continue
		}

		if fallbackIndex < len(fallbackResults) {
			res := fallbackResults[fallbackIndex]
			w.logger.WarnContext(ctx, "applying index fallback for result without queue_job_id",
				slog.String("job_id", job.ID),
				slog.Int("fallback_index", fallbackIndex),
			)
			fallbackIndex++
			succeeded, err := w.applySingleBatchResult(ctx, job, res)
			if err != nil {
				return 0, 0, err
			}
			if succeeded {
				completedCount++
			} else {
				failedCount++
			}
			continue
		}

		errMsg := "batch result missing for request"
		if err := w.queue.UpdateJob(ctx, job.ID, StatusFailed, nil, &errMsg, nil); err != nil {
			return 0, 0, fmt.Errorf("mark missing batch result as failed job_id=%s: %w", job.ID, err)
		}
		failedCount++
	}

	if unknownIDCount > 0 {
		w.logger.WarnContext(ctx, "batch results contained unknown queue_job_id values",
			slog.Int("unknown_id_count", unknownIDCount),
		)
	}
	if fallbackIndex < len(fallbackResults) {
		w.logger.WarnContext(ctx, "batch fallback results exceeded queued jobs",
			slog.Int("remaining_fallback_results", len(fallbackResults)-fallbackIndex),
		)
	}
	if len(results) > len(jobs) {
		w.logger.WarnContext(ctx, "batch results exceeded queued jobs",
			slog.Int("results_count", len(results)),
			slog.Int("jobs_count", len(jobs)),
		)
	}

	return completedCount, failedCount, nil
}

func isBatchTerminalState(state gatewayllm.BatchState) bool {
	switch state {
	case gatewayllm.BatchStateCompleted,
		gatewayllm.BatchStatePartialFailed,
		gatewayllm.BatchStateFailed,
		gatewayllm.BatchStateCancelled:
		return true
	default:
		return false
	}
}

func findExistingBatchJobID(jobs []JobRequest, fallbackProvider string) (gatewayllm.BatchJobID, bool, error) {
	provider := gatewayllm.NormalizeProvider(fallbackProvider)
	var resolved gatewayllm.BatchJobID
	for _, job := range jobs {
		if job.BatchJobID == nil {
			continue
		}
		raw := strings.TrimSpace(*job.BatchJobID)
		if raw == "" {
			continue
		}

		parsed, err := decodeStoredBatchJobID(raw, provider)
		if err != nil {
			return gatewayllm.BatchJobID{}, false, fmt.Errorf("decode batch job id for job_id=%s: %w", job.ID, err)
		}
		if resolved.ID == "" {
			resolved = parsed
			continue
		}
		if resolved.ID != parsed.ID {
			return gatewayllm.BatchJobID{}, false, fmt.Errorf("inconsistent batch job ids in process: %q and %q", resolved.ID, parsed.ID)
		}
	}

	if resolved.ID == "" {
		return gatewayllm.BatchJobID{}, false, nil
	}
	if resolved.Provider == "" {
		resolved.Provider = provider
	}
	return resolved, true, nil
}

func decodeStoredBatchJobID(raw string, fallbackProvider string) (gatewayllm.BatchJobID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return gatewayllm.BatchJobID{}, fmt.Errorf("empty stored batch job id")
	}

	var parsed gatewayllm.BatchJobID
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		if strings.TrimSpace(parsed.ID) == "" {
			return gatewayllm.BatchJobID{}, fmt.Errorf("stored batch job id payload is missing id")
		}
		if strings.TrimSpace(parsed.Provider) == "" {
			parsed.Provider = gatewayllm.NormalizeProvider(fallbackProvider)
		}
		return parsed, nil
	}

	var textID string
	if err := json.Unmarshal([]byte(trimmed), &textID); err == nil {
		textID = strings.TrimSpace(textID)
		if textID == "" {
			return gatewayllm.BatchJobID{}, fmt.Errorf("stored batch job id string is empty")
		}
		return gatewayllm.BatchJobID{ID: textID, Provider: gatewayllm.NormalizeProvider(fallbackProvider)}, nil
	}

	return gatewayllm.BatchJobID{ID: trimmed, Provider: gatewayllm.NormalizeProvider(fallbackProvider)}, nil
}

func marshalBatchJobID(batchJobID gatewayllm.BatchJobID) (string, error) {
	encoded, err := json.Marshal(batchJobID)
	if err != nil {
		return "", fmt.Errorf("marshal batch job id: %w", err)
	}
	return string(encoded), nil
}

// progressReportingClient wraps LLMClient to emit progress events.
type progressReportingClient struct {
	gatewayllm.LLMClient
	notifier  runtimeprogress.ProgressNotifier
	processID string
	total     int
	completed *int32
	onEach    func(completed int, total int)
}

func (c *progressReportingClient) Complete(ctx context.Context, req gatewayllm.Request) (gatewayllm.Response, error) {
	resp, err := c.LLMClient.Complete(ctx, req)
	comp := atomic.AddInt32(c.completed, 1)
	if c.onEach != nil {
		c.onEach(int(comp), c.total)
	}
	if c.notifier != nil {
		c.notifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
			CorrelationID: c.processID,
			Total:         c.total,
			Completed:     int(comp),
			Status:        runtimeprogress.StatusInProgress,
			Message:       "Processing...",
		})
	}
	if err != nil {
		return resp, fmt.Errorf("reporting client completion failed: %w", err)
	}
	return resp, nil
}

func (w *Worker) fetchLLMConfig(ctx context.Context, opts ProcessOptions) (gatewayllm.LLMConfig, error) {
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
		rawProvider = w.getConfigString(ctx, ns, gatewayllm.LLMDefaultProviderKey, "")
	}
	if rawProvider == "" {
		rawProvider = w.getConfigString(ctx, ns, "provider", defaultProvider)
	}
	provider := gatewayllm.NormalizeProvider(rawProvider)
	providerNS := ns + "." + provider
	model := w.getConfigString(ctx, ns, "model", "")
	if model == "" {
		model = w.getConfigString(ctx, providerNS, "model", "")
	}
	if model == "" {
		model = w.getConfigString(ctx, ns, provider+"_"+gatewayllm.LLMModelIDKeySuffix, "")
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
		val, err := w.secretStore.GetSecret(ctx, ns, provider+"_api_key")
		if err != nil {
			w.logger.WarnContext(ctx, "failed to load provider api key from secret store",
				slog.String("namespace", ns),
				slog.String("provider", provider),
				slog.String("error", err.Error()),
			)
		} else {
			apiKey = val
		}
	}

	strConcurrency := w.getConfigString(ctx, ns, gatewayllm.LLMSyncConcurrencyKeySuffix+"."+provider, "")
	var concurrency int
	if strConcurrency != "" {
		parsedConcurrency, err := strconv.Atoi(strConcurrency)
		if err != nil {
			w.logger.WarnContext(ctx, "invalid sync concurrency; using default",
				slog.String("value", strConcurrency),
				slog.String("provider", provider),
				slog.String("error", err.Error()),
			)
		} else {
			concurrency = parsedConcurrency
		}
	}
	if concurrency <= 0 {
		concurrency = gatewayllm.DefaultConcurrency(provider)
	}
	strContextLength := w.getConfigString(ctx, ns, "context_length", "")
	if strContextLength == "" {
		strContextLength = w.getConfigString(ctx, providerNS, "context_length", "")
	}
	var contextLength int
	if strContextLength != "" {
		parsedContextLength, err := strconv.Atoi(strContextLength)
		if err != nil {
			w.logger.WarnContext(ctx, "invalid context_length; ignoring value",
				slog.String("value", strContextLength),
				slog.String("provider", provider),
				slog.String("error", err.Error()),
			)
		} else {
			contextLength = parsedContextLength
		}
	}
	if strings.TrimSpace(model) == "" {
		return gatewayllm.LLMConfig{}, gatewayllm.ErrModelRequired
	}

	params := map[string]interface{}{}
	if contextLength > 0 {
		params["context_length"] = contextLength
	}

	return gatewayllm.LLMConfig{
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
			return "", "", gatewayllm.ErrModelRequired
		}
		return gatewayllm.NormalizeProvider(defaultProvider), defaultModel, nil
	}

	var provider string
	var model string
	for _, job := range jobs {
		if job.Provider == "" || job.Model == "" {
			return "", "", fmt.Errorf("job %s missing provider/model metadata; resume is not allowed", job.ID)
		}
		if provider == "" {
			provider = gatewayllm.NormalizeProvider(job.Provider)
			model = job.Model
			continue
		}
		if provider != gatewayllm.NormalizeProvider(job.Provider) || model != job.Model {
			return "", "", fmt.Errorf("inconsistent provider/model metadata in process jobs")
		}
	}
	return provider, model, nil
}

func (w *Worker) resolveBulkStrategy(ctx context.Context, ns, provider string) gatewayllm.BulkStrategy {
	configured := w.resolveConfiguredBulkStrategy(ctx, ns, provider)
	return w.llmManager.ResolveBulkStrategy(ctx, configured, provider)
}

func (w *Worker) resolveConfiguredBulkStrategy(ctx context.Context, ns, provider string) gatewayllm.BulkStrategy {
	provider = gatewayllm.NormalizeProvider(provider)
	providerNS := ns + "." + provider

	strategyStr := strings.TrimSpace(w.getConfigString(ctx, providerNS, gatewayllm.LLMBulkStrategyKey, ""))
	if strategyStr == "" {
		// Backward compatibility for legacy key layout.
		strategyStr = strings.TrimSpace(w.getConfigString(ctx, ns, gatewayllm.LLMBulkStrategyKey, ""))
	}
	if strategyStr == "" {
		return gatewayllm.BulkStrategySync
	}
	if strings.EqualFold(strategyStr, string(gatewayllm.BulkStrategyBatch)) {
		return gatewayllm.BulkStrategyBatch
	}
	return gatewayllm.BulkStrategySync
}

func (w *Worker) getConfigString(ctx context.Context, ns, key, defaultVal string) string {
	if w.configAccessor == nil {
		return defaultVal
	}
	return w.configAccessor.GetString(ctx, ns, key, defaultVal)
}

func resolveConfigNamespace(opts ProcessOptions) string {
	ns := strings.TrimSpace(opts.ConfigRead.Namespace)
	if ns == "" {
		ns = strings.TrimSpace(opts.ConfigNamespace)
	}
	if ns == "" {
		ns = gatewayllm.LLMConfigNamespace
	}
	return ns
}
