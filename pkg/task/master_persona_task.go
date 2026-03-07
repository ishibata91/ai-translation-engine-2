package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/pipeline"
)

type StartMasterPersonTaskInput struct {
	SourceJSONPath    string `json:"source_json_path"`
	OverwriteExisting bool   `json:"overwrite_existing"`
}

type personaSaveSummary struct {
	Attempted          int
	Saved              int
	Failed             int
	SkippedAlreadySave int
	SavedRequestIDs    []string
}

func (b *Bridge) StartMasterPersonTask(input StartMasterPersonTaskInput) (string, error) {
	if strings.TrimSpace(input.SourceJSONPath) == "" {
		return "", fmt.Errorf("source_json_path is required")
	}

	metadata := TaskMetadata{
		"source_json_path":   input.SourceJSONPath,
		"overwrite_existing": input.OverwriteExisting,
		"entrypoint":         "master_persona",
		"phase":              "prepare_requests",
	}

	taskID, err := b.manager.AddTaskWithCompletionStatus(
		"Master Persona Request Generation",
		TypePersonaExtraction,
		"pending",
		metadata,
		StatusRequestGenerated,
		func(ctx context.Context, taskID string, update func(phase string, progress float64)) error {
			runCtx := telemetry.WithTraceID(ctx)
			b.reportProgress(runCtx, taskID, 0, progress.StatusInProgress, "マスターペルソナ生成タスクを開始")
			update("loading_json", 10)
			b.reportProgress(runCtx, taskID, 10, progress.StatusInProgress, "JSONを読み込み中")

			parsed, err := b.parser.LoadExtractedJSON(runCtx, input.SourceJSONPath)
			if err != nil {
				b.reportProgress(runCtx, taskID, 10, progress.StatusFailed, "JSON読み込みに失敗")
				b.logger.ErrorContext(runCtx, "persona.requests.failed",
					slog.String("task_id", taskID),
					slog.String("reason", err.Error()),
				)
				return err
			}

			update("building_persona_input", 40)
			b.reportProgress(runCtx, taskID, 40, progress.StatusInProgress, "ペルソナ入力を構築中")
			personaInput := pipeline.ToPersonaGenInput(parsed)
			personaInput.SourceJSONPath = input.SourceJSONPath
			personaInput.OverwriteExisting = input.OverwriteExisting

			update("preparing_requests", 75)
			b.reportProgress(runCtx, taskID, 75, progress.StatusInProgress, "リクエストを生成中")
			requests, err := b.personaGenerator.PreparePrompts(runCtx, personaInput)
			if err != nil {
				b.reportProgress(runCtx, taskID, 75, progress.StatusFailed, "リクエスト生成に失敗")
				b.logger.ErrorContext(runCtx, "persona.requests.failed",
					slog.String("task_id", taskID),
					slog.String("reason", err.Error()),
				)
				return err
			}

			taskMetadata := TaskMetadata{
				"source_json_path":   input.SourceJSONPath,
				"overwrite_existing": input.OverwriteExisting,
				"entrypoint":         "master_persona",
				"phase":              "request_enqueued",
				"resume_cursor":      0,
				"request_count":      len(requests),
			}

			if b.queue == nil {
				return fmt.Errorf("request queue is not configured")
			}
			if err := b.queue.SubmitTaskRequests(runCtx, taskID, string(TypePersonaExtraction), requests); err != nil {
				b.reportProgress(runCtx, taskID, 80, progress.StatusFailed, "キュー保存に失敗")
				b.logger.ErrorContext(runCtx, "persona.requests.queue_save_failed",
					slog.String("task_id", taskID),
					slog.String("reason", err.Error()),
				)
				return err
			}
			b.reportProgress(runCtx, taskID, 90, progress.StatusInProgress, "リクエストをキューへ保存")

			if err := b.manager.store.SaveMetadata(runCtx, taskID, taskMetadata); err != nil {
				b.logger.WarnContext(runCtx, "failed to persist persona task summary",
					slog.String("task_id", taskID),
					slog.String("reason", err.Error()),
				)
			}

			update("REQUEST_GENERATED", 100)
			b.reportProgress(runCtx, taskID, 100, progress.StatusCompleted, "リクエスト生成が完了")
			b.manager.EmitPhaseCompleted(taskID, "REQUEST_GENERATED", nil)
			b.logger.InfoContext(runCtx, "persona.requests.generated",
				slog.Int("request_count", len(requests)),
				slog.Int("npc_count", len(personaInput.NPCs)),
				slog.String("task_id", taskID),
				slog.Any("requests", buildRequestLogPayload(requests)),
			)
			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

func buildRequestLogPayload(requests []llm.Request) []map[string]interface{} {
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

func (b *Bridge) runPersonaExecution(ctx context.Context, task *Task, update func(phase string, progress float64)) error {
	if b.queue == nil || b.worker == nil {
		return fmt.Errorf("request queue worker is not configured")
	}

	state, err := b.queue.GetTaskRequestState(ctx, task.ID)
	if err != nil {
		return err
	}
	if state.Total == 0 {
		return fmt.Errorf("task %s has no queued requests", task.ID)
	}

	b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_ENQUEUED", state.Completed, state.Total, progress.StatusInProgress, "リクエスト再開準備")
	update("REQUEST_ENQUEUED", toPercent(state.Completed, state.Total))

	baseCompleted := state.Completed
	overallTotal := state.Total

	if err := b.queue.PrepareTaskResume(ctx, task.ID); err != nil {
		return err
	}

	hooks := &queue.ProcessHooks{
		OnDispatch: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_DISPATCHING", progressCurrent, overallTotal, progress.StatusInProgress, "LM Studioへリクエスト送信中")
			update("REQUEST_DISPATCHING", toPercent(progressCurrent, overallTotal))
		},
		OnSaving: func(current int, _ int) {
			progressCurrent := baseCompleted + current
			b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_SAVING", progressCurrent, overallTotal, progress.StatusInProgress, "レスポンス保存中")
			update("REQUEST_SAVING", toPercent(progressCurrent, overallTotal))
		},
		OnComplete: func(completed int, _ int, failed int) {
			progressCompleted := baseCompleted + completed
			finalStatus := progress.StatusCompleted
			finalMessage := "リクエスト実行が完了"
			if failed > 0 {
				finalStatus = progress.StatusFailed
				finalMessage = "一部リクエストが失敗"
			}
			b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_COMPLETED", progressCompleted, overallTotal, finalStatus, finalMessage)
			if failed == 0 {
				update("REQUEST_COMPLETED", 100)
			}
		},
	}

	err = b.worker.ProcessProcessIDWithOptions(ctx, task.ID, queue.ProcessOptions{
		ConfigNamespace:        "master_persona.llm",
		UseConfigProviderModel: true,
		ConfigRead: queue.ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
		Hooks: hooks,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_ = b.queue.MarkTaskRequestsCanceled(context.Background(), task.ID)
		}
		b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_DISPATCHING", state.Completed, state.Total, progress.StatusFailed, err.Error())
		return err
	}

	updatedState, stateErr := b.queue.GetTaskRequestState(ctx, task.ID)
	if stateErr != nil {
		return nil
	}

	saveSummary, saveErr := b.persistPersonaResponses(ctx, task)
	metadata := mergeTaskMetadata(task.Metadata, TaskMetadata{
		"phase":                "REQUEST_COMPLETED",
		"resume_cursor":        updatedState.Completed,
		"saved_request_ids":    saveSummary.SavedRequestIDs,
		"persona_saved_count":  saveSummary.Saved,
		"persona_failed_count": saveSummary.Failed,
	})
	_ = b.manager.store.SaveMetadata(context.Background(), task.ID, metadata)
	b.manager.EmitPhaseCompleted(task.ID, "REQUEST_COMPLETED", map[string]int{
		"total":                updatedState.Total,
		"completed":            updatedState.Completed,
		"failed":               updatedState.Failed,
		"canceled":             updatedState.Canceled,
		"persona_save_attempt": saveSummary.Attempted,
		"persona_save_success": saveSummary.Saved,
		"persona_save_failed":  saveSummary.Failed,
	})
	if saveErr != nil {
		b.reportTaskPhaseProgress(ctx, task.ID, task.Type, "REQUEST_COMPLETED", updatedState.Completed, updatedState.Total, progress.StatusFailed, saveErr.Error())
		return saveErr
	}
	if saveSummary.Failed > 0 {
		b.logger.WarnContext(ctx, "persona.save.partial_failure",
			slog.String("task_id", task.ID),
			slog.Int("attempted", saveSummary.Attempted),
			slog.Int("saved", saveSummary.Saved),
			slog.Int("failed", saveSummary.Failed),
		)
	}

	return nil
}

func (b *Bridge) persistPersonaResponses(ctx context.Context, task *Task) (personaSaveSummary, error) {
	out := personaSaveSummary{}
	if b.personaGenerator == nil {
		return out, fmt.Errorf("persona generator is not configured")
	}

	jobs, err := b.queue.GetTaskRequests(ctx, task.ID)
	if err != nil {
		return out, err
	}

	savedSet := metadataStringSet(task.Metadata["saved_request_ids"])
	reporter, hasReporter := b.personaGenerator.(persona.SaveResultsReporter)
	for _, job := range jobs {
		if job.RequestState != queue.RequestStateCompleted || job.ResponseJSON == nil {
			continue
		}
		if _, ok := savedSet[job.ID]; ok {
			out.SkippedAlreadySave++
			continue
		}

		resp, parseErr := buildPersonaResponseFromJob(job)
		out.Attempted++
		if parseErr != nil {
			out.Failed++
			continue
		}
		if resp.Metadata == nil {
			resp.Metadata = map[string]interface{}{}
		}
		if _, ok := resp.Metadata["overwrite_existing"]; !ok {
			resp.Metadata["overwrite_existing"] = metadataBool(task.Metadata["overwrite_existing"])
		}

		if hasReporter {
			sum, saveErr := reporter.SaveResultsWithSummary(ctx, []llm.Response{resp})
			if saveErr != nil {
				return out, saveErr
			}
			out.Saved += sum.SuccessCount
			out.Failed += sum.FailCount
			if sum.SuccessCount > 0 {
				savedSet[job.ID] = struct{}{}
			}
			continue
		}

		if err := b.personaGenerator.SaveResults(ctx, []llm.Response{resp}); err != nil {
			return out, err
		}
		out.Saved++
		savedSet[job.ID] = struct{}{}
	}

	out.SavedRequestIDs = sortedStringSet(savedSet)
	return out, nil
}

func buildPersonaResponseFromJob(job queue.JobRequest) (llm.Response, error) {
	var resp llm.Response
	if err := json.Unmarshal([]byte(*job.ResponseJSON), &resp); err != nil {
		return llm.Response{}, fmt.Errorf("failed to decode response for job=%s: %w", job.ID, err)
	}

	var req llm.Request
	if err := json.Unmarshal([]byte(job.RequestJSON), &req); err == nil {
		resp.Metadata = mergeMetadata(resp.Metadata, req.Metadata)
	}
	if resp.Metadata == nil {
		resp.Metadata = map[string]interface{}{}
	}
	return resp, nil
}

func mergeTaskMetadata(base TaskMetadata, updates TaskMetadata) TaskMetadata {
	out := TaskMetadata{}
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
	if total <= 0 {
		return 0
	}
	if current <= 0 {
		return 0
	}
	raw := (float64(current) / float64(total)) * 100
	if raw > 100 {
		return 100
	}
	return math.Round(raw*100) / 100
}
