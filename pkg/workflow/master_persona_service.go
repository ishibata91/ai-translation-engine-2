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

	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow/pipeline"
	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

// MasterPersonaService orchestrates master persona request preparation and execution.
type MasterPersonaService struct {
	manager          *task2.Manager
	logger           *slog.Logger
	parser           parser.Parser
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

// NewMasterPersonaService constructs the workflow implementation.
func NewMasterPersonaService(
	manager *task2.Manager,
	logger *slog.Logger,
	parser parser.Parser,
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

	state, err := s.loadTaskRequestState(ctx, currentTask.ID)
	if err != nil {
		return err
	}

	s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_ENQUEUED", state.Completed, state.Total, runtimeprogress.StatusInProgress, "リクエスト再開準備")
	update("REQUEST_ENQUEUED", toPercent(state.Completed, state.Total))

	baseCompleted := state.Completed
	overallTotal := state.Total

	if err := s.queue.PrepareTaskResume(ctx, currentTask.ID); err != nil {
		return fmt.Errorf("prepare task resume task_id=%s: %w", currentTask.ID, err)
	}

	if err := s.processQueuedRequests(ctx, currentTask, state, update, baseCompleted, overallTotal); err != nil {
		return err
	}

	updatedState, err := s.loadUpdatedTaskState(ctx, currentTask.ID)
	if err != nil {
		return err
	}

	return s.finalizePersonaExecution(ctx, currentTask, updatedState)
}

func (s *MasterPersonaService) executeRequestPreparation(ctx context.Context, taskID string, input StartMasterPersonaInput, update func(phase string, progress float64)) error {
	runCtx := telemetry2.WithTraceID(ctx)
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
		wrappedErr := fmt.Errorf("prepare prompts task_id=%s source_json_path=%s: %w", taskID, input.SourceJSONPath, err)
		s.reportProgress(runCtx, taskID, 75, runtimeprogress.StatusFailed, "リクエスト生成に失敗")
		s.logger.ErrorContext(runCtx, "persona.requests.failed",
			slog.String("task_id", taskID),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}

	taskMetadata := task2.TaskMetadata{
		"source_json_path":   input.SourceJSONPath,
		"overwrite_existing": input.OverwriteExisting,
		"entrypoint":         "master_persona",
		"phase":              "request_enqueued",
		"resume_cursor":      0,
		"request_count":      len(requests),
	}

	if s.queue == nil {
		return fmt.Errorf("request queue is not configured")
	}
	if err := s.queue.SubmitTaskRequests(runCtx, taskID, string(task2.TypePersonaExtraction), requests); err != nil {
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
) error {
	hooks := &runtimequeue.ProcessHooks{
		OnDispatch: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_DISPATCHING", progressCurrent, overallTotal, runtimeprogress.StatusInProgress, "LM Studioへリクエスト送信中")
			update("REQUEST_DISPATCHING", toPercent(progressCurrent, overallTotal))
		},
		OnSaving: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_SAVING", progressCurrent, overallTotal, runtimeprogress.StatusInProgress, "レスポンス保存中")
			update("REQUEST_SAVING", toPercent(progressCurrent, overallTotal))
		},
		OnComplete: func(completed int, _ int, failed int) {
			progressCompleted := baseCompleted + completed
			finalStatus := runtimeprogress.StatusCompleted
			finalMessage := "リクエスト実行が完了"
			if failed > 0 {
				finalStatus = runtimeprogress.StatusFailed
				finalMessage = "一部リクエストが失敗"
			}
			s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_COMPLETED", progressCompleted, overallTotal, finalStatus, finalMessage)
			if failed == 0 {
				update("REQUEST_COMPLETED", 100)
			}
		},
	}

	err := s.worker.ProcessProcessIDWithOptions(ctx, currentTask.ID, runtimequeue.ProcessOptions{
		ConfigNamespace:        "master_persona.llm",
		UseConfigProviderModel: true,
		ConfigRead: runtimequeue.ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
		Hooks: hooks,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if cancelErr := s.queue.MarkTaskRequestsCanceled(ctx, currentTask.ID); cancelErr != nil {
				wrappedErr := fmt.Errorf("mark task requests canceled task_id=%s: %w", currentTask.ID, cancelErr)
				s.logger.WarnContext(ctx, "persona.requests.cancel_mark_failed",
					append(telemetry2.ErrorAttrs(wrappedErr), slog.String("task_id", currentTask.ID))...)
			}
		}
		s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_DISPATCHING", state.Completed, state.Total, runtimeprogress.StatusFailed, err.Error())
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

func (s *MasterPersonaService) finalizePersonaExecution(ctx context.Context, currentTask *task2.Task, updatedState runtimequeue.TaskRequestState) error {
	saveSummary, saveErr := s.persistPersonaResponses(ctx, currentTask)
	metadata := mergeTaskMetadata(currentTask.Metadata, task2.TaskMetadata{
		"phase":                "REQUEST_COMPLETED",
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

	s.manager.EmitPhaseCompleted(currentTask.ID, "REQUEST_COMPLETED", map[string]int{
		"total":                updatedState.Total,
		"completed":            updatedState.Completed,
		"failed":               updatedState.Failed,
		"canceled":             updatedState.Canceled,
		"persona_save_attempt": saveSummary.Attempted,
		"persona_save_success": saveSummary.Saved,
		"persona_save_failed":  saveSummary.Failed,
	})

	if saveErr != nil {
		s.reportTaskPhaseProgress(ctx, currentTask.ID, currentTask.Type, "REQUEST_COMPLETED", updatedState.Completed, updatedState.Total, runtimeprogress.StatusFailed, saveErr.Error())
		return fmt.Errorf("persist persona responses task_id=%s: %w", currentTask.ID, saveErr)
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
			sum, saveErr := reporter.SaveResultsWithSummary(ctx, []gatewayllm.Response{resp})
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

		if err := s.personaGenerator.SaveResults(ctx, []gatewayllm.Response{resp}); err != nil {
			return out, fmt.Errorf("save persona response task_id=%s request_id=%s: %w", currentTask.ID, job.ID, err)
		}
		out.Saved++
		savedSet[job.ID] = struct{}{}
	}

	out.SavedRequestIDs = sortedStringSet(savedSet)
	return out, nil
}

func buildRequestLogPayload(requests []gatewayllm.Request) []map[string]interface{} {
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

func buildPersonaResponseFromJob(job runtimequeue.JobRequest) (gatewayllm.Response, error) {
	var resp gatewayllm.Response
	if err := json.Unmarshal([]byte(*job.ResponseJSON), &resp); err != nil {
		return gatewayllm.Response{}, fmt.Errorf("failed to decode response for job=%s: %w", job.ID, err)
	}

	var req gatewayllm.Request
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
