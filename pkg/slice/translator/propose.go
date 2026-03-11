package translator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

type translatorSlice struct {
	contextEngine ContextEngine
	promptBuilder PromptBuilder
	resumeLoader  ResumeLoader
	resultWriter  ResultWriter
	tagProcessor  TagProcessor
	bookChunker   BookChunker
}

func NewTranslatorSlice(
	ce ContextEngine,
	pb PromptBuilder,
	rl ResumeLoader,
	rw ResultWriter,
	tp TagProcessor,
	bc BookChunker,
) TranslatorSlice {
	return &translatorSlice{
		contextEngine: ce,
		promptBuilder: pb,
		resumeLoader:  rl,
		resultWriter:  rw,
		tagProcessor:  tp,
		bookChunker:   bc,
	}
}

func (s *translatorSlice) ID() string {
	return "Translator"
}

func (s *translatorSlice) PreparePrompts(ctx context.Context, input any) ([]llmio.Request, error) {
	typedInput, ok := input.(TranslatorInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Translator slice: %T", input)
	}
	return s.ProposeJobs(ctx, typedInput)
}

func (s *translatorSlice) ProposeJobs(ctx context.Context, input TranslatorInput) ([]llmio.Request, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionProcessTranslation)()
	slog.DebugContext(ctx, "starting job proposal",
		slog.String("plugin", input.OutputConfig.PluginName),
		slog.Int("dialogue_count", len(input.GameData.Dialogues)),
	)

	// 1. Load cached results for resume
	cached, err := s.resumeLoader.LoadCachedResults(input.OutputConfig.PluginName, input.OutputConfig.OutputBaseDir)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load cached results", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("failed to load cached results: %w", err)
	}

	var requests []llmio.Request
	completedCount := 0
	forcedCount := 0

	// 2. Process all records in GameData to build context and generate jobs
	for _, dial := range input.GameData.Dialogues {
		// Check if already translated
		if res, ok := cached[dial.ID]; ok && res.Status == "completed" {
			completedCount++
			continue
		}

		// Phase 1: Context Building (Integrated Lore logic)
		pass2Ctx, terms, forced, err := s.contextEngine.BuildTranslationContext(ctx, dial, &input.GameData)
		if err != nil {
			slog.ErrorContext(ctx, "failed to build translation context",
				append(telemetry2.ErrorAttrs(err), slog.String("resource_id", dial.ID))...)
			continue
		}

		// If forced translation is found in dictionary/terms, we can bypass LLM.
		if forced != nil {
			forcedCount++
			slog.InfoContext(ctx, "forced translation found",
				slog.String("resource_id", dial.ID),
				slog.String("reason", "dictionary_match"),
			)
			result := TranslationResult{
				ID:             dial.ID,
				RecordType:     dial.Type,
				SourceText:     *dial.Text,
				TranslatedText: forced,
				Status:         "completed",
				SourcePlugin:   input.OutputConfig.PluginName,
			}
			if err := s.resultWriter.Write(result); err != nil {
				slog.ErrorContext(ctx, "failed to write forced result",
					append(telemetry2.ErrorAttrs(err), slog.String("resource_id", dial.ID))...)
			}
			continue
		}

		// Tag protection
		processedText, tags := s.tagProcessor.Preprocess(*dial.Text)

		// Book Chunking (if needed)
		maxChars := 4000
		if input.OutputConfig.MaxTokens > 0 {
			maxChars = input.OutputConfig.MaxTokens
		}
		chunks := s.bookChunker.Chunk(processedText, maxChars)

		for i, chunk := range chunks {
			// Prepare internal request DTO
			req := Pass2TranslationRequest{
				ID:             dial.ID,
				RecordType:     dial.Type,
				SourceText:     chunk,
				Context:        *pass2Ctx,
				ReferenceTerms: terms,
				EditorID:       dial.EditorID,
				SourcePlugin:   input.OutputConfig.PluginName,
				SourceFile:     input.Config.SourceFile,
				MaxTokens:      &input.OutputConfig.MaxTokens,
			}
			if len(chunks) > 1 {
				idx := i
				req.Index = &idx
			}

			// Phase 2: Prompt Building
			systemPrompt, userPrompt, err := s.promptBuilder.Build(ctx, req)
			if err != nil {
				slog.ErrorContext(ctx, "failed to build prompt",
					append(telemetry2.ErrorAttrs(err),
						slog.String("resource_id", dial.ID),
						slog.Int("chunk_index", i))...)
				continue
			}

			requests = append(requests, llmio.Request{
				SystemPrompt: systemPrompt,
				UserPrompt:   userPrompt,
				Metadata: map[string]interface{}{
					"id":            req.ID,
					"record_type":   req.RecordType,
					"source_plugin": req.SourcePlugin,
					"tags":          tags,
					"chunk_index":   i,
					"is_chunked":    len(chunks) > 1,
				},
			})
		}
	}

	slog.InfoContext(ctx, "job proposal completed",
		slog.Int("total_requests", len(requests)),
		slog.Int("skipped_already_completed", completedCount),
		slog.Int("forced_translations", forcedCount),
	)
	return requests, nil
}
