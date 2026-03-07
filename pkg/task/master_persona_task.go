package task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/pipeline"
)

type StartMasterPersonTaskInput struct {
	SourceJSONPath string `json:"source_json_path"`
}

type PersonaRequestSummary struct {
	RequestCount int `json:"request_count"`
	NPCCount     int `json:"npc_count"`
}

func (b *Bridge) StartMasterPersonTask(input StartMasterPersonTaskInput) (string, error) {
	if strings.TrimSpace(input.SourceJSONPath) == "" {
		return "", fmt.Errorf("source_json_path is required")
	}

	metadata := TaskMetadata{
		"source_json_path": input.SourceJSONPath,
		"entrypoint":       "master_persona",
		"phase":            "prepare_requests",
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

			summary := PersonaRequestSummary{
				RequestCount: len(requests),
				NPCCount:     len(personaInput.NPCs),
			}
			taskMetadata := TaskMetadata{
				"source_json_path": input.SourceJSONPath,
				"entrypoint":       "master_persona",
				"phase":            "request_enqueued",
				"resume_cursor":    0,
				"request_count":    summary.RequestCount,
				"npc_count":        summary.NPCCount,
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
			b.manager.EmitPhaseCompleted(taskID, "REQUEST_GENERATED", summary)
			b.logger.InfoContext(runCtx, "persona.requests.generated",
				slog.Int("request_count", summary.RequestCount),
				slog.Int("npc_count", summary.NPCCount),
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
			"max_tokens":    req.MaxTokens,
			"temperature":   req.Temperature,
			"metadata":      req.Metadata,
		})
	}
	return payload
}
