package translator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

type translatorSlice struct {
	contextEngine ContextEngine
	promptBuilder PromptBuilder
	resumeLoader  ResumeLoader
	resultWriter  ResultWriter
	tagProcessor  TagProcessor
}

func NewTranslatorSlice(
	ce ContextEngine,
	pb PromptBuilder,
	rl ResumeLoader,
	rw ResultWriter,
	tp TagProcessor,
) TranslatorSlice {
	return &translatorSlice{
		contextEngine: ce,
		promptBuilder: pb,
		resumeLoader:  rl,
		resultWriter:  rw,
		tagProcessor:  tp,
	}
}

func (s *translatorSlice) ID() string {
	return "Translator"
}

func (s *translatorSlice) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	typedInput, ok := input.(TranslatorInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Translator slice: %T", input)
	}
	return s.ProposeJobs(ctx, typedInput)
}

func (s *translatorSlice) ProposeJobs(ctx context.Context, input TranslatorInput) ([]llm.Request, error) {
	slog.DebugContext(ctx, "ProposeJobs started", "plugin", input.OutputConfig.PluginName)

	// 1. Load cached results for resume
	cached, err := s.resumeLoader.LoadCachedResults(input.OutputConfig.PluginName, input.OutputConfig.OutputBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load cached results: %w", err)
	}

	var requests []llm.Request

	// 2. Process all records in GameData to build context and generate jobs
	// Simplified loop over Dialogues as an example.
	for _, dial := range input.GameData.Dialogues {
		// Check if already translated
		if res, ok := cached[dial.ID]; ok && res.Status == "completed" {
			continue
		}

		// Phase 1: Context Building (Integrated Lore logic)
		pass2Ctx, terms, _, err := s.contextEngine.BuildTranslationContext(ctx, dial, &input.GameData)
		if err != nil {
			slog.ErrorContext(ctx, "failed to build context", "id", dial.ID, "error", err)
			continue
		}

		// Prepare internal request DTO
		req := Pass2TranslationRequest{
			ID:             dial.ID,
			RecordType:     dial.Type,
			SourceText:     *dial.Text,
			Context:        *pass2Ctx,
			ReferenceTerms: terms,
			EditorID:       dial.EditorID,
			SourcePlugin:   input.OutputConfig.PluginName,
			SourceFile:     input.Config.SourceFile,
			MaxTokens:      &input.OutputConfig.MaxTokens,
		}

		// Phase 2: Prompt Building
		systemPrompt, userPrompt, err := s.promptBuilder.Build(ctx, req)
		if err != nil {
			slog.ErrorContext(ctx, "failed to build prompt", "id", dial.ID, "error", err)
			continue
		}

		requests = append(requests, llm.Request{
			SystemPrompt: systemPrompt,
			UserPrompt:   userPrompt,
			MaxTokens:    input.OutputConfig.MaxTokens,
			Metadata: map[string]interface{}{
				"id": req.ID,
			},
		})
	}

	slog.DebugContext(ctx, "ProposeJobs completed", "requests_count", len(requests))
	return requests, nil
}

func (s *translatorSlice) SaveResults(ctx context.Context, responses []llm.Response) error {
	slog.DebugContext(ctx, "SaveResults started", "responses_count", len(responses))

	for _, resp := range responses {
		id, ok := resp.Metadata["id"].(string)
		if !ok {
			slog.WarnContext(ctx, "response missing record id in metadata", "content", resp.Content)
			continue
		}

		// 1. Tag Restoration & Validation could happen here

		// 2. Write to persistent storage
		result := TranslationResult{
			ID:             id,
			TranslatedText: &resp.Content,
			Status:         "completed",
		}
		if err := s.resultWriter.Write(result); err != nil {
			return fmt.Errorf("failed to write result for %s: %w", id, err)
		}
	}

	if err := s.resultWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush results: %w", err)
	}

	slog.DebugContext(ctx, "SaveResults completed")
	return nil
}
