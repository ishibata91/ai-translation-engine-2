package translator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
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

func (s *translatorSlice) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	typedInput, ok := input.(TranslatorInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Translator slice: %T", input)
	}
	return s.ProposeJobs(ctx, typedInput)
}

func (s *translatorSlice) ProposeJobs(ctx context.Context, input TranslatorInput) ([]llm.Request, error) {
	slog.DebugContext(ctx, "ENTER ProposeJobs",
		slog.String("plugin", input.OutputConfig.PluginName),
		slog.Int("dialogue_count", len(input.GameData.Dialogues)),
	)
	start := time.Now()

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
		pass2Ctx, terms, forced, err := s.contextEngine.BuildTranslationContext(ctx, dial, &input.GameData)
		if err != nil {
			slog.ErrorContext(ctx, "failed to build context", "id", dial.ID, "error", err)
			continue
		}

		// If forced translation is found in dictionary/terms, we can bypass LLM.
		if forced != nil {
			slog.InfoContext(ctx, "forced translation found", "id", dial.ID)
			result := TranslationResult{
				ID:             dial.ID,
				RecordType:     dial.Type,
				SourceText:     *dial.Text,
				TranslatedText: forced,
				Status:         "completed",
				SourcePlugin:   input.OutputConfig.PluginName,
			}
			if err := s.resultWriter.Write(result); err != nil {
				slog.ErrorContext(ctx, "failed to write forced result", "id", dial.ID, "error", err)
			}
			continue
		}

		// Tag protection
		processedText, tags := s.tagProcessor.Preprocess(*dial.Text)

		// Book Chunking (if needed)
		// Use MaxTokens as a character limit for now, or a reasonable default
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
				slog.ErrorContext(ctx, "failed to build prompt", "id", dial.ID, "chunk", i, "error", err)
				continue
			}

			requests = append(requests, llm.Request{
				SystemPrompt: systemPrompt,
				UserPrompt:   userPrompt,
				MaxTokens:    input.OutputConfig.MaxTokens,
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

	slog.DebugContext(ctx, "EXIT ProposeJobs",
		slog.Int("requests_count", len(requests)),
		slog.Duration("elapsed", time.Since(start)),
	)
	return requests, nil
}
