package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow/pipeline"
	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

// MasterPersonaService orchestrates master persona request preparation and execution.
type MasterPersonaService struct {
	manager          *task2.Manager
	logger           *slog.Logger
	parser           skyrim.Parser
	personaGenerator persona.NPCPersonaGenerator
	notifier         runtimeprogress.ProgressNotifier
	queue            *runtimequeue.Queue
	worker           *runtimequeue.Worker
}

type personaSaveSummary struct {
	Attempted          int
	Saved              int
	Failed             int
	SkippedAlreadySave int
	SavedRequestIDs    []string
}

const (
	phaseRequestEnqueued      = "REQUEST_ENQUEUED"
	phaseRequestExecutingSync = "REQUEST_EXECUTING_SYNC"
	phaseBatchSubmitted       = "BATCH_SUBMITTED"
	phaseBatchPolling         = "BATCH_POLLING"
	phaseRequestSaving        = "REQUEST_SAVING"
	phaseRequestCompleted     = "REQUEST_COMPLETED"
)

// NewMasterPersonaService constructs the workflow implementation.
func NewMasterPersonaService(
	manager *task2.Manager,
	logger *slog.Logger,
	parser skyrim.Parser,
	personaGenerator persona.NPCPersonaGenerator,
	notifier runtimeprogress.ProgressNotifier,
	requestQueue *runtimequeue.Queue,
	requestWorker *runtimequeue.Worker,
) *MasterPersonaService {
	return &MasterPersonaService{
		manager:          manager,
		logger:           logger.With("module", "master_persona_workflow"),
		parser:           parser,
		personaGenerator: personaGenerator,
		notifier:         notifier,
		queue:            requestQueue,
		worker:           requestWorker,
	}
}

// StartMasterPersona starts phase-1 request generation.
func (s *MasterPersonaService) StartMasterPersona(ctx context.Context, input StartMasterPersonaInput) (string, error) {
	if strings.TrimSpace(input.SourceJSONPath) == "" {
		return "", fmt.Errorf("source_json_path is required")
	}

	metadata := task2.TaskMetadata{
		"source_json_path":   input.SourceJSONPath,
		"overwrite_existing": input.OverwriteExisting,
		"entrypoint":         "master_persona",
		"phase":              "prepare_requests",
	}

	taskID, err := s.manager.AddTaskWithCompletionStatusContext(
		ctx,
		"Master Persona Request Generation",
		task2.TypePersonaExtraction,
		"pending",
		metadata,
		task2.StatusRequestGenerated,
		func(runCtx context.Context, taskID string, update func(phase string, progress float64)) error {
			return s.executeRequestPreparation(runCtx, taskID, input, update)
		},
	)
	if err != nil {
		return "", fmt.Errorf("start master persona task source_json_path=%s: %w", input.SourceJSONPath, err)
	}
	return taskID, nil
}

// RunPersonaPhase bootstraps requests on first run and resumes runtime only when requests already exist.
func (s *MasterPersonaService) RunPersonaPhase(ctx context.Context, input PersonaExecutionInput) error {
	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if s.queue == nil {
		return fmt.Errorf("request queue is not configured")
	}

	runtimeEntries, err := s.ListPersonaRuntime(ctx, trimmedTaskID)
	if err != nil {
		return fmt.Errorf("list persona runtime task_id=%s: %w", trimmedTaskID, err)
	}
	if len(runtimeEntries) == 0 {
		sourceJSONPath := strings.TrimSpace(input.SourceJSONPath)
		if sourceJSONPath == "" {
			return fmt.Errorf("source_json_path is required for persona bootstrap task_id=%s", trimmedTaskID)
		}

		bootstrapInput := StartMasterPersonaInput{
			SourceJSONPath:    sourceJSONPath,
			OverwriteExisting: input.OverwriteExisting,
		}
		bootstrapCtx := withPersonaPhaseRunConfig(ctx, input.Request, input.Prompt)
		if err := s.executeRequestPreparation(bootstrapCtx, trimmedTaskID, bootstrapInput, func(_ string, _ float64) {}); err != nil {
			return fmt.Errorf("bootstrap persona requests task_id=%s: %w", trimmedTaskID, err)
		}
		return nil
	}

	if err := s.runPersonaRuntime(ctx, input); err != nil {
		return fmt.Errorf("resume persona runtime task_id=%s: %w", trimmedTaskID, err)
	}
	return nil
}

// ListPersonaRuntime returns task-scoped runtime snapshot without exposing queue internals.
func (s *MasterPersonaService) ListPersonaRuntime(ctx context.Context, taskID string) ([]PersonaRuntimeEntry, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if s.queue == nil {
		return nil, fmt.Errorf("request queue is not configured")
	}

	requests, err := s.queue.GetTaskRequests(ctx, trimmedTaskID)
	if err != nil {
		return nil, fmt.Errorf("get task requests task_id=%s: %w", trimmedTaskID, err)
	}

	entries := make([]PersonaRuntimeEntry, 0, len(requests))
	for _, request := range requests {
		sourcePlugin, speakerID, hasLookupKey := parsePersonaRuntimeLookupFromRequestJSON(request.RequestJSON)
		entry := PersonaRuntimeEntry{
			RequestID:    request.ID,
			SourcePlugin: sourcePlugin,
			SpeakerID:    speakerID,
			RequestState: request.RequestState,
			ResumeCursor: request.ResumeCursor,
			UpdatedAt:    request.UpdatedAt,
			HasResponse:  request.ResponseJSON != nil,
			HasLookupKey: hasLookupKey,
		}
		if request.ErrorMessage != nil {
			entry.ErrorMessage = strings.TrimSpace(*request.ErrorMessage)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// ResumeMasterPersona resumes queued work for one task.
func (s *MasterPersonaService) ResumeMasterPersona(ctx context.Context, taskID string) error {
	if err := s.manager.ResumeTaskWithContext(ctx, taskID); err != nil {
		return fmt.Errorf("resume master persona task_id=%s: %w", taskID, err)
	}
	return nil
}

// CancelMasterPersona requests cancellation for one task.
func (s *MasterPersonaService) CancelMasterPersona(ctx context.Context, taskID string) error {
	s.logger.InfoContext(ctx, "persona.task.cancel_requested", slog.String("task_id", taskID))
	s.manager.CancelTask(taskID)
	return nil
}

// GetTaskRequestState returns aggregate request state for one task.
func (s *MasterPersonaService) GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error) {
	if s.queue == nil {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("request queue is not configured")
	}
	state, err := s.queue.GetTaskRequestState(ctx, taskID)
	if err != nil {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("get task request state task_id=%s: %w", taskID, err)
	}
	return state, nil
}

// GetTaskRequests returns queued requests for one task.
func (s *MasterPersonaService) GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error) {
	if s.queue == nil {
		return nil, fmt.Errorf("request queue is not configured")
	}
	requests, err := s.queue.GetTaskRequests(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task requests task_id=%s: %w", taskID, err)
	}
	return requests, nil
}

// StartMasterPersonTask is the controller-facing wrapper used by task.Bridge.
func (s *MasterPersonaService) StartMasterPersonTask(ctx context.Context, input task2.StartMasterPersonTaskInput) (string, error) {
	return s.StartMasterPersona(ctx, StartMasterPersonaInput{
		SourceJSONPath:    input.SourceJSONPath,
		OverwriteExisting: input.OverwriteExisting,
	})
}

// ResumeMasterPersonaTask is the controller-facing wrapper used by task.Bridge.
func (s *MasterPersonaService) ResumeMasterPersonaTask(ctx context.Context, taskID string) error {
	return s.ResumeMasterPersona(ctx, taskID)
}

// CancelTask is the controller-facing wrapper used by task.Bridge.
func (s *MasterPersonaService) CancelTask(ctx context.Context, taskID string) {
	if err := s.CancelMasterPersona(ctx, taskID); err != nil {
		s.logger.WarnContext(ctx, "persona.task.cancel_failed",
			append(telemetry2.ErrorAttrs(err), slog.String("task_id", taskID))...)
	}
}

// CleanupCompletedTask removes queued requests after a MasterPersona task is confirmed completed.
func (s *MasterPersonaService) CleanupCompletedTask(ctx context.Context, currentTask *task2.Task) error {
	if currentTask.Type != task2.TypePersonaExtraction {
		return nil
	}
	if s.queue == nil {
		return fmt.Errorf("request queue is not configured")
	}
	if err := s.queue.DeleteTaskRequests(ctx, currentTask.ID); err != nil {
		return fmt.Errorf("delete completed task requests task_id=%s: %w", currentTask.ID, err)
	}
	return nil
}

// Run satisfies task.Runner and keeps task execution in workflow.
func (s *MasterPersonaService) Run(ctx context.Context, currentTask *task2.Task, update func(phase string, progress float64)) error {
	if currentTask.Type != task2.TypePersonaExtraction {
		return fmt.Errorf("unsupported task type for workflow runner")
	}
	return s.runPersonaExecution(ctx, currentTask, update)
}

func (s *MasterPersonaService) runPersonaRuntime(ctx context.Context, input PersonaExecutionInput) error {
	if s.manager == nil {
		return fmt.Errorf("task manager is not configured")
	}
	if s.queue == nil || s.worker == nil {
		return fmt.Errorf("request queue worker is not configured")
	}

	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return fmt.Errorf("task_id is required")
	}

	metadata := personaExecutionConfigMetadata(input.Request, input.Prompt)
	metadata["entrypoint"] = "translation_flow_persona_phase"
	metadata["phase"] = phaseRequestEnqueued
	if strings.TrimSpace(input.SourceJSONPath) != "" {
		metadata["source_json_path"] = strings.TrimSpace(input.SourceJSONPath)
	}
	metadata["overwrite_existing"] = input.OverwriteExisting

	if existing, err := s.manager.Store().GetMetadata(ctx, trimmedTaskID); err == nil {
		metadata = mergeTaskMetadata(existing, metadata)
	}

	currentTask := &task2.Task{
		ID:       trimmedTaskID,
		Type:     task2.TypePersonaExtraction,
		Metadata: metadata,
	}
	runCtx := withPersonaPhaseRunConfig(ctx, input.Request, input.Prompt)
	if err := s.runPersonaExecution(runCtx, currentTask, func(_ string, _ float64) {}); err != nil {
		return fmt.Errorf("run persona runtime task_id=%s: %w", trimmedTaskID, err)
	}
	return nil
}

func (s *MasterPersonaService) reportProgress(ctx context.Context, correlationID string, completed int, status string, message string) {
	if s.notifier == nil {
		return
	}
	s.notifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
		CorrelationID: correlationID,
		Total:         100,
		Completed:     completed,
		Status:        status,
		Message:       message,
	})
}

func (s *MasterPersonaService) reportTaskPhaseProgress(ctx context.Context, taskID string, taskType task2.TaskType, phase string, current int, total int, status string, message string) {
	if s.notifier == nil {
		return
	}
	s.notifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
		CorrelationID: taskID,
		TaskID:        taskID,
		TaskType:      string(taskType),
		Phase:         phase,
		Current:       current,
		Total:         total,
		Completed:     current,
		Status:        status,
		Message:       message,
	})
}

func (s *MasterPersonaService) runPersonaExecution(ctx context.Context, currentTask *task2.Task, update func(phase string, progress float64)) error {
	if s.queue == nil || s.worker == nil {
		return fmt.Errorf("request queue worker is not configured")
	}

	processOpts := masterPersonaProcessOptions()
	profile, err := s.worker.ValidateExecutionProfile(ctx, processOpts)
	if err != nil {
		return fmt.Errorf("validate execution profile task_id=%s: %w", currentTask.ID, err)
	}

	taskMetadata := s.persistTaskMetadata(ctx, currentTask.ID, currentTask.Metadata, executionProfileMetadata(profile))

	state, err := s.loadTaskRequestState(ctx, currentTask.ID)
	if err != nil {
		return err
	}

	s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseRequestEnqueued, state.Completed, state.Total, runtimeprogress.StatusInProgress, "リクエスト再開準備")
	update(phaseRequestEnqueued, toPercent(state.Completed, state.Total))

	baseCompleted := state.Completed
	overallTotal := state.Total

	if err := s.queue.PrepareTaskResume(ctx, currentTask.ID); err != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(ctx, currentTask.ID); cleanupErr != nil {
			s.logger.WarnContext(ctx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", currentTask.ID))...)
		}
		return fmt.Errorf("prepare task resume task_id=%s: %w", currentTask.ID, err)
	}

	if err := s.processQueuedRequests(ctx, currentTask, state, update, baseCompleted, overallTotal, profile, processOpts, &taskMetadata); err != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(ctx, currentTask.ID); cleanupErr != nil {
			s.logger.WarnContext(ctx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", currentTask.ID))...)
		}
		return err
	}

	updatedState, err := s.loadUpdatedTaskState(ctx, currentTask.ID)
	if err != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(ctx, currentTask.ID); cleanupErr != nil {
			s.logger.WarnContext(ctx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", currentTask.ID))...)
		}
		return err
	}

	return s.finalizePersonaExecution(ctx, currentTask, updatedState, taskMetadata)
}

func masterPersonaProcessOptions() runtimequeue.ProcessOptions {
	return runtimequeue.ProcessOptions{
		ConfigNamespace:        "master_persona.llm",
		UseConfigProviderModel: true,
		ConfigRead: runtimequeue.ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
	}
}

func executionProfileMetadata(profile runtimequeue.ExecutionProfile) task2.TaskMetadata {
	return task2.TaskMetadata{
		"execution_profile": map[string]interface{}{
			"provider":                profile.Provider,
			"model":                   profile.Model,
			"requested_bulk_strategy": string(profile.RequestedBulkStrategy),
			"bulk_strategy":           string(profile.BulkStrategy),
		},
		"execution_provider":                profile.Provider,
		"execution_model":                   profile.Model,
		"execution_requested_bulk_strategy": string(profile.RequestedBulkStrategy),
		"execution_bulk_strategy":           string(profile.BulkStrategy),
	}
}

func (s *MasterPersonaService) persistTaskMetadata(ctx context.Context, taskID string, base task2.TaskMetadata, updates task2.TaskMetadata) task2.TaskMetadata {
	metadata := mergeTaskMetadata(base, updates)
	if err := s.manager.Store().SaveMetadata(ctx, taskID, metadata); err != nil {
		wrappedErr := fmt.Errorf("save metadata task_id=%s: %w", taskID, err)
		s.logger.WarnContext(ctx, "persona.metadata_persist_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("task_id", taskID))...)
		return base
	}
	return metadata
}

func isBatchExecutionProfile(profile runtimequeue.ExecutionProfile) bool {
	return strings.EqualFold(string(profile.BulkStrategy), "batch")
}

func batchProgressCurrent(baseCompleted int, overallTotal int, progress float32) int {
	if overallTotal <= 0 {
		return 0
	}
	if baseCompleted < 0 {
		baseCompleted = 0
	}
	if baseCompleted > overallTotal {
		baseCompleted = overallTotal
	}

	ratio := float64(progress)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	remaining := overallTotal - baseCompleted
	current := baseCompleted + int(math.Round(ratio*float64(remaining)))
	if current < baseCompleted {
		return baseCompleted
	}
	if current > overallTotal {
		return overallTotal
	}
	return current
}

func (s *MasterPersonaService) executeRequestPreparation(ctx context.Context, taskID string, input StartMasterPersonaInput, update func(phase string, progress float64)) error {
	runCtx := persona.WithTaskID(telemetry2.WithTraceID(ctx), taskID)
	s.reportProgress(runCtx, taskID, 0, runtimeprogress.StatusInProgress, "マスターペルソナ生成タスクを開始")
	update("loading_json", 10)
	s.reportProgress(runCtx, taskID, 10, runtimeprogress.StatusInProgress, "JSONを読み込み中")

	parsed, err := s.parser.LoadExtractedJSON(runCtx, input.SourceJSONPath)
	if err != nil {
		wrappedErr := fmt.Errorf("load extracted json source_json_path=%s: %w", input.SourceJSONPath, err)
		s.reportProgress(runCtx, taskID, 10, runtimeprogress.StatusFailed, "JSON読み込みに失敗")
		s.logger.ErrorContext(runCtx, "persona.requests.failed",
			slog.String("task_id", taskID),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}

	update("building_persona_input", 40)
	s.reportProgress(runCtx, taskID, 40, runtimeprogress.StatusInProgress, "ペルソナ入力を構築中")
	personaInput := pipeline.ToPersonaGenInput(parsed)
	personaInput.SourceJSONPath = input.SourceJSONPath
	personaInput.OverwriteExisting = input.OverwriteExisting

	update("preparing_requests", 75)
	s.reportProgress(runCtx, taskID, 75, runtimeprogress.StatusInProgress, "リクエストを生成中")
	requests, err := s.personaGenerator.PreparePrompts(runCtx, personaInput)
	if err != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(runCtx, taskID); cleanupErr != nil {
			s.logger.WarnContext(runCtx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", taskID))...)
		}
		wrappedErr := fmt.Errorf("prepare prompts task_id=%s source_json_path=%s: %w", taskID, input.SourceJSONPath, err)
		s.reportProgress(runCtx, taskID, 75, runtimeprogress.StatusFailed, "リクエスト生成に失敗")
		s.logger.ErrorContext(runCtx, "persona.requests.failed",
			slog.String("task_id", taskID),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}
	requestCfg, promptCfg, hasExecutionOverrides := personaPhaseRunConfigFromContext(runCtx)
	if hasExecutionOverrides {
		requests = applyPersonaExecutionOverrides(requests, requestCfg, promptCfg)
	}

	taskMetadata := task2.TaskMetadata{
		"source_json_path":   input.SourceJSONPath,
		"overwrite_existing": input.OverwriteExisting,
		"entrypoint":         "master_persona",
		"phase":              phaseRequestEnqueued,
		"resume_cursor":      0,
		"request_count":      len(requests),
	}
	if hasExecutionOverrides {
		taskMetadata = mergeTaskMetadata(taskMetadata, personaExecutionConfigMetadata(requestCfg, promptCfg))
	}

	if s.queue == nil {
		return fmt.Errorf("request queue is not configured")
	}
	if err := s.queue.SubmitTaskSharedRequests(runCtx, taskID, string(task2.TypePersonaExtraction), requests); err != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(runCtx, taskID); cleanupErr != nil {
			s.logger.WarnContext(runCtx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", taskID))...)
		}
		wrappedErr := fmt.Errorf("submit task requests task_id=%s request_count=%d: %w", taskID, len(requests), err)
		s.reportProgress(runCtx, taskID, 80, runtimeprogress.StatusFailed, "キュー保存に失敗")
		s.logger.ErrorContext(runCtx, "persona.requests.queue_save_failed",
			slog.String("task_id", taskID),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}

	s.reportProgress(runCtx, taskID, 90, runtimeprogress.StatusInProgress, "リクエストをキューへ保存")

	if err := s.manager.Store().SaveMetadata(runCtx, taskID, taskMetadata); err != nil {
		wrappedErr := fmt.Errorf("save persona task metadata task_id=%s: %w", taskID, err)
		s.logger.WarnContext(runCtx, "persona.requests.metadata_persist_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("task_id", taskID))...)
	}

	update("REQUEST_GENERATED", 100)
	s.reportProgress(runCtx, taskID, 100, runtimeprogress.StatusCompleted, "リクエスト生成が完了")
	s.manager.EmitPhaseCompleted(taskID, "REQUEST_GENERATED", nil)
	s.logger.InfoContext(runCtx, "persona.requests.generated",
		slog.Int("request_count", len(requests)),
		slog.Int("npc_count", len(personaInput.NPCs)),
		slog.String("task_id", taskID),
		slog.Any("requests", buildRequestLogPayload(requests)),
	)
	return nil
}

func (s *MasterPersonaService) loadTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error) {
	state, err := s.queue.GetTaskRequestState(ctx, taskID)
	if err != nil {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("get task request state task_id=%s: %w", taskID, err)
	}
	if state.Total == 0 {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("task %s has no queued requests", taskID)
	}
	return state, nil
}

func (s *MasterPersonaService) processQueuedRequests(
	ctx context.Context,
	currentTask *task2.Task,
	state runtimequeue.TaskRequestState,
	update func(phase string, progress float64),
	baseCompleted int,
	overallTotal int,
	profile runtimequeue.ExecutionProfile,
	processOpts runtimequeue.ProcessOptions,
	taskMetadata *task2.TaskMetadata,
) error {
	executionMessage := "リクエスト実行中"
	if profile.Provider != "" && profile.Model != "" {
		executionMessage = fmt.Sprintf("リクエスト実行中 (%s / %s)", profile.Provider, profile.Model)
	} else if profile.Provider != "" {
		executionMessage = fmt.Sprintf("リクエスト実行中 (%s)", profile.Provider)
	}

	failurePhase := phaseRequestExecutingSync
	if isBatchExecutionProfile(profile) {
		failurePhase = phaseBatchPolling
	}

	lastBatchState := ""
	hooks := &runtimequeue.ProcessHooks{
		OnDispatch: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseRequestExecutingSync, progressCurrent, overallTotal, runtimeprogress.StatusInProgress, executionMessage)
			update(phaseRequestExecutingSync, toPercent(progressCurrent, overallTotal))
		},
		OnBatchSubmitted: func(batchJobID string, reconnected bool) {
			message := "Batchジョブを送信しました"
			if reconnected {
				message = "既存Batchジョブに再接続しました"
			}
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseBatchSubmitted, baseCompleted, overallTotal, runtimeprogress.StatusInProgress, message)
			update(phaseBatchSubmitted, toPercent(baseCompleted, overallTotal))
			if taskMetadata != nil {
				*taskMetadata = s.persistTaskMetadata(ctx, currentTask.ID, *taskMetadata, task2.TaskMetadata{
					"phase":             phaseBatchSubmitted,
					"batch_job_id":      batchJobID,
					"batch_reconnected": reconnected,
					"batch_state":       "submitted",
				})
			}
		},
		OnBatchPolling: func(state string, progress float32) {
			pollingCurrent := batchProgressCurrent(baseCompleted, overallTotal, progress)
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseBatchPolling, pollingCurrent, overallTotal, runtimeprogress.StatusInProgress, fmt.Sprintf("Batch状態を確認中 (%s)", state))
			update(phaseBatchPolling, toPercent(pollingCurrent, overallTotal))
			if taskMetadata != nil && state != "" && state != lastBatchState {
				lastBatchState = state
				*taskMetadata = s.persistTaskMetadata(ctx, currentTask.ID, *taskMetadata, task2.TaskMetadata{
					"phase":       phaseBatchPolling,
					"batch_state": state,
				})
			}
		},
		OnSaving: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseRequestSaving, progressCurrent, overallTotal, runtimeprogress.StatusInProgress, "レスポンス保存中")
			update(phaseRequestSaving, toPercent(progressCurrent, overallTotal))
		},
		OnComplete: func(completed int, _ int, failed int) {
			progressCompleted := baseCompleted + completed
			finalStatus := runtimeprogress.StatusCompleted
			finalMessage := "リクエスト実行が完了"
			if failed > 0 {
				finalStatus = runtimeprogress.StatusFailed
				finalMessage = "一部リクエストが失敗"
			}
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseRequestCompleted, progressCompleted, overallTotal, finalStatus, finalMessage)
			if failed == 0 {
				update(phaseRequestCompleted, 100)
			}
		},
	}

	processOpts.Hooks = hooks
	err := s.worker.ProcessProcessIDWithOptions(ctx, currentTask.ID, processOpts)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if cancelErr := s.queue.MarkTaskRequestsCanceled(ctx, currentTask.ID); cancelErr != nil {
				wrappedErr := fmt.Errorf("mark task requests canceled task_id=%s: %w", currentTask.ID, cancelErr)
				s.logger.WarnContext(ctx, "persona.requests.cancel_mark_failed",
					append(telemetry2.ErrorAttrs(wrappedErr), slog.String("task_id", currentTask.ID))...)
			}
		}
		s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, failurePhase, state.Completed, state.Total, runtimeprogress.StatusFailed, err.Error())
		return fmt.Errorf("process queued requests task_id=%s: %w", currentTask.ID, err)
	}
	return nil
}

func (s *MasterPersonaService) loadUpdatedTaskState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error) {
	state, err := s.queue.GetTaskRequestState(ctx, taskID)
	if err != nil {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("refresh task request state task_id=%s: %w", taskID, err)
	}
	return state, nil
}

func (s *MasterPersonaService) finalizePersonaExecution(ctx context.Context, currentTask *task2.Task, updatedState runtimequeue.TaskRequestState, baseMetadata task2.TaskMetadata) error {
	saveSummary, saveErr := s.persistPersonaResponses(ctx, currentTask)
	metadata := mergeTaskMetadata(baseMetadata, task2.TaskMetadata{
		"phase":                phaseRequestCompleted,
		"resume_cursor":        updatedState.Completed,
		"saved_request_ids":    saveSummary.SavedRequestIDs,
		"persona_saved_count":  saveSummary.Saved,
		"persona_failed_count": saveSummary.Failed,
	})

	if err := s.manager.Store().SaveMetadata(ctx, currentTask.ID, metadata); err != nil {
		wrappedErr := fmt.Errorf("save completed metadata task_id=%s: %w", currentTask.ID, err)
		s.logger.WarnContext(ctx, "persona.responses.metadata_persist_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("task_id", currentTask.ID))...)
	}

	s.manager.EmitPhaseCompleted(currentTask.ID, phaseRequestCompleted, map[string]int{
		"total":                updatedState.Total,
		"completed":            updatedState.Completed,
		"failed":               updatedState.Failed,
		"canceled":             updatedState.Canceled,
		"persona_save_attempt": saveSummary.Attempted,
		"persona_save_success": saveSummary.Saved,
		"persona_save_failed":  saveSummary.Failed,
	})

	if saveErr != nil {
		if cleanupErr := s.cleanupPersonaTaskArtifacts(ctx, currentTask.ID); cleanupErr != nil {
			s.logger.WarnContext(ctx, "persona.artifacts.cleanup_failed",
				append(telemetry2.ErrorAttrs(cleanupErr), slog.String("task_id", currentTask.ID))...)
		}
		s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, phaseRequestCompleted, updatedState.Completed, updatedState.Total, runtimeprogress.StatusFailed, saveErr.Error())
		return fmt.Errorf("persist persona responses task_id=%s: %w", currentTask.ID, saveErr)
	}
	if err := s.cleanupPersonaTaskArtifacts(ctx, currentTask.ID); err != nil {
		return fmt.Errorf("cleanup persona task artifacts task_id=%s: %w", currentTask.ID, err)
	}
	return nil
}

func (s *MasterPersonaService) persistPersonaResponses(ctx context.Context, currentTask *task2.Task) (personaSaveSummary, error) {
	out := personaSaveSummary{}
	if s.personaGenerator == nil {
		return out, fmt.Errorf("persona generator is not configured")
	}

	jobs, err := s.queue.GetTaskRequests(ctx, currentTask.ID)
	if err != nil {
		return out, fmt.Errorf("load task requests task_id=%s: %w", currentTask.ID, err)
	}

	savedSet := metadataStringSet(currentTask.Metadata["saved_request_ids"])
	reporter, hasReporter := s.personaGenerator.(persona.SaveResultsReporter)
	runCtx := persona.WithTaskID(ctx, currentTask.ID)
	for _, job := range jobs {
		if job.RequestState != runtimequeue.RequestStateCompleted || job.ResponseJSON == nil {
			continue
		}
		if _, ok := savedSet[job.ID]; ok {
			out.SkippedAlreadySave++
			continue
		}

		resp, parseErr := buildPersonaResponseFromJob(job)
		out.Attempted++
		if parseErr != nil {
			s.logger.WarnContext(ctx, "persona.responses.decode_failed",
				append(telemetry2.ErrorAttrs(parseErr),
					slog.String("task_id", currentTask.ID),
					slog.String("request_id", job.ID),
				)...,
			)
			out.Failed++
			continue
		}
		if resp.Metadata == nil {
			resp.Metadata = map[string]interface{}{}
		}
		if _, ok := resp.Metadata["overwrite_existing"]; !ok {
			resp.Metadata["overwrite_existing"] = metadataBool(currentTask.Metadata["overwrite_existing"])
		}

		if hasReporter {
			sum, saveErr := reporter.SaveResultsWithSummary(runCtx, []llmio.Response{resp})
			if saveErr != nil {
				return out, fmt.Errorf("save persona response with summary task_id=%s request_id=%s: %w", currentTask.ID, job.ID, saveErr)
			}
			out.Saved += sum.SuccessCount
			out.Failed += sum.FailCount
			if sum.SuccessCount > 0 {
				savedSet[job.ID] = struct{}{}
			}
			continue
		}

		if err := s.personaGenerator.SaveResults(runCtx, []llmio.Response{resp}); err != nil {
			return out, fmt.Errorf("save persona response task_id=%s request_id=%s: %w", currentTask.ID, job.ID, err)
		}
		out.Saved++
		savedSet[job.ID] = struct{}{}
	}

	out.SavedRequestIDs = sortedStringSet(savedSet)
	return out, nil
}

func (s *MasterPersonaService) cleanupPersonaTaskArtifacts(ctx context.Context, taskID string) error {
	cleaner, ok := s.personaGenerator.(persona.TaskArtifactCleaner)
	if !ok {
		return nil
	}
	if err := cleaner.CleanupTaskArtifacts(persona.WithTaskID(ctx, taskID), taskID); err != nil {
		return fmt.Errorf("cleanup persona task artifacts: %w", err)
	}
	return nil
}

func buildRequestLogPayload(requests []llmio.Request) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(requests))
	for _, req := range requests {
		payload = append(payload, map[string]interface{}{
			"system_prompt": req.SystemPrompt,
			"user_prompt":   req.UserPrompt,
			"temperature":   req.Temperature,
			"metadata":      req.Metadata,
		})
	}
	return payload
}

func buildPersonaResponseFromJob(job runtimequeue.JobRequest) (llmio.Response, error) {
	var resp llmio.Response
	if err := json.Unmarshal([]byte(*job.ResponseJSON), &resp); err != nil {
		return llmio.Response{}, fmt.Errorf("failed to decode response for job=%s: %w", job.ID, err)
	}

	var req llmio.Request
	if err := json.Unmarshal([]byte(job.RequestJSON), &req); err == nil {
		resp.Metadata = mergeMetadata(resp.Metadata, req.Metadata)
	}
	if resp.Metadata == nil {
		resp.Metadata = map[string]interface{}{}
	}
	return resp, nil
}

func mergeTaskMetadata(base task2.TaskMetadata, updates task2.TaskMetadata) task2.TaskMetadata {
	out := task2.TaskMetadata{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range updates {
		out[k] = v
	}
	return out
}

func metadataStringSet(raw interface{}) map[string]struct{} {
	out := map[string]struct{}{}
	switch v := raw.(type) {
	case []string:
		for _, item := range v {
			if strings.TrimSpace(item) == "" {
				continue
			}
			out[item] = struct{}{}
		}
	case []interface{}:
		for _, item := range v {
			s, ok := item.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			out[s] = struct{}{}
		}
	}
	return out
}

func sortedStringSet(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func mergeMetadata(primary map[string]interface{}, fallback map[string]interface{}) map[string]interface{} {
	if primary == nil && fallback == nil {
		return nil
	}
	merged := map[string]interface{}{}
	for k, v := range fallback {
		merged[k] = v
	}
	for k, v := range primary {
		merged[k] = v
	}
	return merged
}

func personaExecutionConfigMetadata(request TranslationRequestConfig, prompt TranslationPromptConfig) task2.TaskMetadata {
	requestMetadata := map[string]interface{}{
		"provider":         strings.TrimSpace(request.Provider),
		"model":            strings.TrimSpace(request.Model),
		"endpoint":         strings.TrimSpace(request.Endpoint),
		"temperature":      request.Temperature,
		"context_length":   request.ContextLength,
		"sync_concurrency": request.SyncConcurrency,
		"bulk_strategy":    strings.TrimSpace(request.BulkStrategy),
	}
	promptMetadata := map[string]interface{}{
		"user_prompt":   strings.TrimSpace(prompt.UserPrompt),
		"system_prompt": strings.TrimSpace(prompt.SystemPrompt),
	}
	return task2.TaskMetadata{
		"request_config": requestMetadata,
		"prompt_config":  promptMetadata,
	}
}

func applyPersonaExecutionOverrides(requests []llmio.Request, request TranslationRequestConfig, prompt TranslationPromptConfig) []llmio.Request {
	if len(requests) == 0 {
		return requests
	}

	overridden := make([]llmio.Request, 0, len(requests))
	userPromptOverride := strings.TrimSpace(prompt.UserPrompt)
	systemPromptOverride := strings.TrimSpace(prompt.SystemPrompt)
	for _, req := range requests {
		current := req
		if current.Metadata == nil {
			current.Metadata = map[string]interface{}{}
		}
		if userPromptOverride != "" {
			current.UserPrompt = mergePersonaUserPrompt(current.UserPrompt, userPromptOverride)
		}
		if systemPromptOverride != "" {
			current.SystemPrompt = systemPromptOverride
		}
		if request.Temperature > 0 {
			current.Temperature = request.Temperature
		}

		if provider := strings.TrimSpace(request.Provider); provider != "" {
			current.Metadata["execution_provider"] = provider
		}
		if model := strings.TrimSpace(request.Model); model != "" {
			current.Metadata["execution_model"] = model
		}
		if endpoint := strings.TrimSpace(request.Endpoint); endpoint != "" {
			current.Metadata["execution_endpoint"] = endpoint
		}
		if bulkStrategy := strings.TrimSpace(request.BulkStrategy); bulkStrategy != "" {
			current.Metadata["execution_bulk_strategy"] = bulkStrategy
		}
		if request.SyncConcurrency > 0 {
			current.Metadata["execution_sync_concurrency"] = request.SyncConcurrency
		}
		if request.ContextLength > 0 {
			current.Metadata["execution_context_length"] = request.ContextLength
		}
		current.Metadata["execution_temperature"] = current.Temperature
		if userPromptOverride != "" {
			current.Metadata["execution_user_prompt"] = userPromptOverride
		}
		if systemPromptOverride != "" {
			current.Metadata["execution_system_prompt"] = systemPromptOverride
		}
		overridden = append(overridden, current)
	}
	return overridden
}

func mergePersonaUserPrompt(original string, override string) string {
	trimmedOverride := strings.TrimSpace(override)
	if trimmedOverride == "" {
		return original
	}

	const profileDelimiter = "\n\nNPC Profile:\n"
	idx := strings.Index(original, profileDelimiter)
	if idx < 0 {
		return trimmedOverride
	}
	return trimmedOverride + original[idx:]
}

func parsePersonaRuntimeLookupFromRequestJSON(rawRequest string) (sourcePlugin string, speakerID string, ok bool) {
	var payload struct {
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(rawRequest), &payload); err != nil {
		return "", "", false
	}

	rawSpeakerID, hasSpeaker := payload.Metadata["speaker_id"]
	if !hasSpeaker {
		return "", "", false
	}
	speakerIDText, ok := rawSpeakerID.(string)
	if !ok {
		return "", "", false
	}
	trimmedSpeakerID := strings.TrimSpace(speakerIDText)
	if trimmedSpeakerID == "" {
		return "", "", false
	}

	sourcePluginRaw, _ := payload.Metadata["source_plugin"].(string)
	normalizedSourcePlugin := normalizePersonaSourcePlugin(sourcePluginRaw, "")
	return normalizedSourcePlugin, trimmedSpeakerID, true
}

func metadataBool(raw interface{}) bool {
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func toPercent(current int, total int) float64 {
	if total <= 0 || current <= 0 {
		return 0
	}
	raw := (float64(current) / float64(total)) * 100
	if raw > 100 {
		return 100
	}
	return math.Round(raw*100) / 100
}
