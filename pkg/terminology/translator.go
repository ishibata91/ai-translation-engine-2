package terminology

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

// TermTranslatorImpl implements Terminology.
type TermTranslatorImpl struct {
	builder       TermRequestBuilder
	searcher      TermDictionarySearcher
	store         ModTermStore
	promptBuilder TermPromptBuilder
	logger        *slog.Logger
}

// NewTermTranslator creates a new TermTranslatorImpl.
func NewTermTranslator(
	builder TermRequestBuilder,
	searcher TermDictionarySearcher,
	store ModTermStore,
	promptBuilder TermPromptBuilder,
	logger *slog.Logger,
) *TermTranslatorImpl {
	return &TermTranslatorImpl{
		builder:       builder,
		searcher:      searcher,
		store:         store,
		promptBuilder: promptBuilder,
		logger:        logger.With("component", "TermTranslatorImpl"),
	}
}

// PreparePrompts builds LLM requests (Phase 1).
func (t *TermTranslatorImpl) PreparePrompts(ctx context.Context, data TerminologyInput) ([]llm_client.Request, error) {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.PreparePrompts")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.PreparePrompts")

	requests, err := t.builder.BuildRequests(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build requests: %w", err)
	}
	t.logger.InfoContext(ctx, "Built translation requests", "count", len(requests))

	if len(requests) == 0 {
		return nil, nil
	}

	llmRequests := make([]llm_client.Request, 0, len(requests))
	for _, req := range requests {
		// Fetch reference terms for LLM context
		t.fetchReferenceTerms(ctx, &req)

		prompt, err := t.promptBuilder.BuildPrompt(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to build prompt for %s: %w", req.SourceText, err)
		}

		llmRequests = append(llmRequests, llm_client.Request{
			SystemPrompt: prompt,
			UserPrompt:   "Translate the provided term.",
		})
	}

	return llmRequests, nil
}

// SaveResults parses LLM responses and persists to the mod term database (Phase 2).
func (t *TermTranslatorImpl) SaveResults(ctx context.Context, data TerminologyInput, results []llm_client.Response) error {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.SaveResults")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.SaveResults")

	// Ensure DB schema is ready.
	if err := t.store.InitSchema(ctx); err != nil {
		return fmt.Errorf("failed to init mod term schema: %w", err)
	}

	// Re-build requests to match results (as per design mitigation: index mapping).
	requests, err := t.builder.BuildRequests(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to build requests for matching: %w", err)
	}

	if len(requests) != len(results) {
		return fmt.Errorf("request/response count mismatch: req=%d, res=%d", len(requests), len(results))
	}

	var finalResults []TermTranslationResult
	for i, res := range results {
		req := requests[i]

		translationResult := TermTranslationResult{
			FormID:       req.FormID,
			EditorID:     req.EditorID,
			RecordType:   req.RecordType,
			SourceText:   req.SourceText,
			SourcePlugin: req.SourcePlugin,
			SourceFile:   req.SourceFile,
			Status:       "success",
		}

		if !res.Success {
			t.logger.WarnContext(ctx, "LLM response failed",
				"index", i,
				"term", req.SourceText,
				"error", res.Error)
			translationResult.Status = "error"
			translationResult.ErrorMessage = res.Error
			// Skip saving if failed? Or save as error?
			// Spec says: "安全にスキップし、エラー詳細を構造化ログとして記録する"
			continue
		}

		translatedText := t.extractTranslationFromLLMResponse(res.Content)
		if translatedText == res.Content && !strings.Contains(res.Content, "TL: |") {
			t.logger.WarnContext(ctx, "LLM response missing expected format",
				"index", i,
				"term", req.SourceText)
			// According to scenario: "TL: |...| 形式が含まれていない ... 安全にスキップ"
			continue
		}

		translationResult.TranslatedText = translatedText

		// Expand NPC if needed (FULL/SHRT)
		expanded := t.expandResult(translationResult, req)
		finalResults = append(finalResults, expanded...)
	}

	if len(finalResults) > 0 {
		if err := t.store.SaveTerms(ctx, finalResults); err != nil {
			return fmt.Errorf("failed to save terms: %w", err)
		}
		t.logger.InfoContext(ctx, "Saved term translations to mod DB", "count", len(finalResults))
	}

	return nil
}

// fetchReferenceTerms retrieves context reference terms based on the record type.
func (t *TermTranslatorImpl) fetchReferenceTerms(ctx context.Context, req *TermTranslationRequest) {
	keywords := strings.Split(req.SourceText, " ")
	isNPC := strings.HasPrefix(req.RecordType, "NPC")

	var contextRefs []ReferenceTerm
	if isNPC {
		npcRefs, err := t.searcher.SearchNPCPartial(ctx, keywords, nil, true)
		if err == nil {
			contextRefs = append(contextRefs, npcRefs...)
		}
	} else {
		kwRefs, err := t.searcher.SearchKeywords(ctx, keywords)
		if err == nil {
			contextRefs = append(contextRefs, kwRefs...)
		}
	}

	req.ReferenceTerms = contextRefs
}

// expandResult handles NPC paired results or returns a single result.
func (t *TermTranslatorImpl) expandResult(res TermTranslationResult, req TermTranslationRequest) []TermTranslationResult {
	if (res.Status == "success" || res.Status == "cached") && req.RecordType == "NPC_" && req.ShortName != "" {
		return t.expandNPCResult(res, req)
	}
	return []TermTranslationResult{res}
}

// expandNPCResult splits a paired NPC translation into FULL and SHRT results.
func (t *TermTranslatorImpl) expandNPCResult(res TermTranslationResult, req TermTranslationRequest) []TermTranslationResult {
	fullRes := res
	fullRes.RecordType = "NPC_:FULL"

	shortRes := res
	shortRes.RecordType = "NPC_:SHRT"
	shortRes.SourceText = req.ShortName
	shortRes.TranslatedText = strings.Split(res.TranslatedText, " ")[0]

	return []TermTranslationResult{fullRes, shortRes}
}

// extractTranslationFromLLMResponse extracts the translation text from "TL: |...|" format.
func (t *TermTranslatorImpl) extractTranslationFromLLMResponse(content string) string {
	content = strings.TrimSpace(content)
	startIdx := strings.Index(content, "TL: |")
	if startIdx != -1 {
		startIdx += 5 // length of "TL: |"
		endIdx := strings.Index(content[startIdx:], "|")
		if endIdx != -1 {
			return strings.TrimSpace(content[startIdx : startIdx+endIdx])
		}
		return strings.TrimSpace(content[startIdx:])
	}
	return content
}
