package terminology

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
)

// TermTranslatorImpl implements Terminology.
type TermTranslatorImpl struct {
	inputRepo     TranslationInputRepository
	builder       TermRequestBuilder
	searcher      TermDictionarySearcher
	store         ModTermStore
	promptBuilder TermPromptBuilder
	logger        *slog.Logger
}

// NewTermTranslator creates a new TermTranslatorImpl.
func NewTermTranslator(
	inputRepo TranslationInputRepository,
	builder TermRequestBuilder,
	searcher TermDictionarySearcher,
	store ModTermStore,
	promptBuilder TermPromptBuilder,
	logger *slog.Logger,
) *TermTranslatorImpl {
	return &TermTranslatorImpl{
		inputRepo:     inputRepo,
		builder:       builder,
		searcher:      searcher,
		store:         store,
		promptBuilder: promptBuilder,
		logger:        logger.With("component", "TermTranslatorImpl"),
	}
}

// ID returns the unique identifier of the slice.
func (t *TermTranslatorImpl) ID() string {
	return "Terminology"
}

// PreparePrompts implementation for terminology phase.
func (t *TermTranslatorImpl) PreparePrompts(ctx context.Context, taskID string, options PhaseOptions) ([]llmio.Request, error) {
	artifactInput, err := t.inputRepo.LoadTerminologyInput(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("load terminology artifact input task_id=%s: %w", taskID, err)
	}
	input := toTerminologyInput(artifactInput)
	requests, err := t.preparePrompts(ctx, input, options)
	if err != nil {
		return nil, fmt.Errorf("prepare terminology prompts task_id=%s: %w", taskID, err)
	}
	status := "pending"
	if len(requests) > 0 {
		status = "running"
	}
	if err := t.store.UpdatePhaseSummary(ctx, PhaseSummary{
		TaskID:      taskID,
		Status:      status,
		TargetCount: len(requests),
	}); err != nil {
		return nil, fmt.Errorf("persist terminology phase running summary task_id=%s: %w", taskID, err)
	}
	return requests, nil
}

// preparePrompts builds LLM requests (Phase 1).
func (t *TermTranslatorImpl) preparePrompts(ctx context.Context, data TerminologyInput, options PhaseOptions) ([]llmio.Request, error) {
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

	llmRequests := make([]llmio.Request, 0, len(requests))
	for _, req := range requests {
		// Fetch reference terms for LLM context
		t.fetchReferenceTerms(ctx, &req)

		prompt, err := t.buildPrompt(ctx, req, options)
		if err != nil {
			return nil, fmt.Errorf("failed to build prompt for %s: %w", req.SourceText, err)
		}

		llmRequests = append(llmRequests, llmio.Request{
			SystemPrompt: prompt,
			UserPrompt:   userPromptOrDefault(options),
			Temperature:  options.Request.Temperature,
			Metadata: map[string]interface{}{
				"source_text":   req.SourceText,
				"form_id":       req.FormID,
				"editor_id":     req.EditorID,
				"record_type":   req.RecordType,
				"source_plugin": req.SourcePlugin,
				"source_file":   req.SourceFile,
				"short_name":    req.ShortName,
			},
		})
	}

	return llmRequests, nil
}

// SaveResults implementation for terminology phase.
func (t *TermTranslatorImpl) SaveResults(ctx context.Context, taskID string, responses []llmio.Response) error {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.SaveResults")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.SaveResults")

	// Ensure DB schema is ready.
	if err := t.store.InitSchema(ctx); err != nil {
		return fmt.Errorf("failed to init mod term schema: %w", err)
	}

	var finalResults []TermTranslationResult
	failedCount := 0
	for i, res := range responses {
		// Identify Term from metadata
		sourceText, _ := res.Metadata["source_text"].(string)
		formID, _ := res.Metadata["form_id"].(string)
		editorID, _ := res.Metadata["editor_id"].(string)
		recordType, _ := res.Metadata["record_type"].(string)
		sourcePlugin, _ := res.Metadata["source_plugin"].(string)
		sourceFile, _ := res.Metadata["source_file"].(string)
		shortName, _ := res.Metadata["short_name"].(string)

		translationResult := TermTranslationResult{
			FormID:       formID,
			EditorID:     editorID,
			RecordType:   recordType,
			SourceText:   sourceText,
			SourcePlugin: sourcePlugin,
			SourceFile:   sourceFile,
			Status:       "success",
		}

		if !res.Success {
			t.logger.WarnContext(ctx, "LLM response failed",
				"index", i,
				"term", sourceText,
				"error", res.Error)
			translationResult.Status = "error"
			translationResult.ErrorMessage = res.Error
			failedCount++
			continue
		}

		translatedText := t.extractTranslationFromLLMResponse(res.Content)
		if translatedText == res.Content && !strings.Contains(res.Content, "TL: |") {
			t.logger.WarnContext(ctx, "LLM response missing expected format",
				"index", i,
				"term", sourceText)
			failedCount++
			continue
		}

		translationResult.TranslatedText = translatedText

		// Expand NPC if needed (FULL/SHRT)
		// We need to re-construct a partial request for expandResult to work
		dummyReq := TermTranslationRequest{
			RecordType: recordType,
			ShortName:  shortName,
		}
		expanded := t.expandResult(translationResult, dummyReq)
		finalResults = append(finalResults, expanded...)
	}

	if len(finalResults) > 0 {
		if err := t.store.SaveTerms(ctx, finalResults); err != nil {
			return fmt.Errorf("failed to save terms: %w", err)
		}
		t.logger.InfoContext(ctx, "Saved term translations to mod DB", "count", len(finalResults))
	}

	status := "completed"
	if len(responses) == 0 {
		status = "pending"
	} else if failedCount > 0 && len(finalResults) == 0 {
		status = "failed"
	}
	if err := t.store.UpdatePhaseSummary(ctx, PhaseSummary{
		TaskID:      taskID,
		Status:      status,
		TargetCount: len(responses),
		SavedCount:  len(finalResults),
		FailedCount: failedCount,
	}); err != nil {
		return fmt.Errorf("persist terminology phase summary task_id=%s: %w", taskID, err)
	}
	return nil
}

// GetPhaseSummary returns the persisted terminology phase summary.
func (t *TermTranslatorImpl) GetPhaseSummary(ctx context.Context, taskID string) (PhaseSummary, error) {
	summary, err := t.store.GetPhaseSummary(ctx, taskID)
	if err != nil {
		return PhaseSummary{}, fmt.Errorf("get terminology phase summary task_id=%s: %w", taskID, err)
	}
	return summary, nil
}

// LegacySaveResults parses LLM responses and persists to the mod term database (Phase 2).
func (t *TermTranslatorImpl) LegacySaveResults(ctx context.Context, data TerminologyInput, results []llmio.Response) error {
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
	keywords := strings.Fields(req.SourceText)
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

func (t *TermTranslatorImpl) buildPrompt(ctx context.Context, request TermTranslationRequest, options PhaseOptions) (string, error) {
	templateString := strings.TrimSpace(options.Prompt.SystemPrompt)
	if templateString == "" {
		return t.promptBuilder.BuildPrompt(ctx, request)
	}
	builder, err := NewTermPromptBuilder(templateString)
	if err != nil {
		return "", fmt.Errorf("create terminology prompt builder: %w", err)
	}
	return builder.BuildPrompt(ctx, request)
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

func userPromptOrDefault(options PhaseOptions) string {
	if strings.TrimSpace(options.Prompt.UserPrompt) != "" {
		return options.Prompt.UserPrompt
	}
	return "Translate the provided term."
}

func toTerminologyInput(input translationinput.TerminologyInput) TerminologyInput {
	entries := make([]TerminologyEntry, 0, len(input.Entries))
	for _, entry := range input.Entries {
		entries = append(entries, TerminologyEntry{
			ID:         entry.ID,
			EditorID:   entry.EditorID,
			RecordType: entry.RecordType,
			SourceText: entry.SourceText,
			SourceFile: entry.SourceFile,
			PairKey:    entry.PairKey,
			Variant:    entry.Variant,
		})
	}

	return TerminologyInput{
		TaskID:    input.TaskID,
		FileNames: append([]string(nil), input.FileNames...),
		Entries:   entries,
	}
}
