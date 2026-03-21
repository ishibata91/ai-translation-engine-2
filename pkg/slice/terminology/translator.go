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
	matcher       *GreedyLongestMatcher
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
		matcher:       NewGreedyLongestMatcher(),
		logger:        logger.With("component", "TermTranslatorImpl"),
	}
}

// ID returns the unique identifier of the slice.
func (t *TermTranslatorImpl) ID() string {
	return "Terminology"
}

// PreparePrompts implementation for terminology phase.
func (t *TermTranslatorImpl) PreparePrompts(ctx context.Context, taskID string, options PhaseOptions) ([]llmio.Request, error) {
	requests, summary, err := t.preparePrompts(ctx, taskID, options)
	if err != nil {
		return nil, fmt.Errorf("prepare terminology prompts task_id=%s: %w", taskID, err)
	}
	if err := t.store.UpdatePhaseSummary(ctx, summary); err != nil {
		return nil, fmt.Errorf("persist terminology phase running summary task_id=%s: %w", taskID, err)
	}
	return requests, nil
}

// preparePrompts builds LLM requests (Phase 1).
func (t *TermTranslatorImpl) preparePrompts(ctx context.Context, taskID string, options PhaseOptions) ([]llmio.Request, PhaseSummary, error) {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.PreparePrompts")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.PreparePrompts")

	artifactInput, err := t.inputRepo.LoadTerminologyInput(ctx, taskID)
	if err != nil {
		return nil, PhaseSummary{}, fmt.Errorf("load terminology artifact input task_id=%s: %w", taskID, err)
	}
	data := toTerminologyInput(artifactInput)
	requests, err := t.builder.BuildRequests(ctx, data)
	if err != nil {
		return nil, PhaseSummary{}, fmt.Errorf("failed to build requests: %w", err)
	}
	if len(requests) == 0 {
		return nil, PhaseSummary{
			TaskID:       taskID,
			Status:       "pending",
			ProgressMode: "hidden",
		}, nil
	}

	targetCount := len(requests)
	cachedResults := make([]TermTranslationResult, 0, targetCount)
	llmRequests := make([]llmio.Request, 0, len(requests))
	cachedGroupCount := 0
	for _, req := range requests {
		exactRefs, err := t.searcher.SearchExact(ctx, req.SourceText)
		if err != nil {
			return nil, PhaseSummary{}, fmt.Errorf("search exact references source=%q: %w", req.SourceText, err)
		}
		if len(exactRefs) > 0 {
			cachedResult := TermTranslationResult{
				FormID:         req.FormID,
				EditorID:       req.EditorID,
				RecordType:     req.RecordType,
				SourceText:     req.SourceText,
				TranslatedText: exactRefs[0].Translation,
				SourcePlugin:   req.SourcePlugin,
				SourceFile:     req.SourceFile,
				Status:         "cached",
			}
			cachedResults = append(cachedResults, t.expandResult(cachedResult, req)...)
			cachedGroupCount++
			continue
		}

		workingReq := req
		replacedSourceText, consumedKeywords, err := t.buildReplacedSourceText(ctx, req.SourceText)
		if err != nil {
			return nil, PhaseSummary{}, fmt.Errorf("build replaced source text source=%q: %w", req.SourceText, err)
		}
		if strings.TrimSpace(replacedSourceText) == "" {
			replacedSourceText = req.SourceText
		}
		workingReq.ReplacedSourceText = replacedSourceText
		workingReq.ReferenceTerms = t.fetchReferenceTerms(ctx, req, replacedSourceText, consumedKeywords)

		prompt, err := t.buildPrompt(ctx, workingReq, options)
		if err != nil {
			return nil, PhaseSummary{}, fmt.Errorf("failed to build prompt for %s: %w", req.SourceText, err)
		}

		llmRequests = append(llmRequests, llmio.Request{
			SystemPrompt: prompt,
			UserPrompt:   userPromptOrDefault(options),
			Temperature:  options.Request.Temperature,
			Metadata: map[string]interface{}{
				"source_text":          req.SourceText,
				"original_source_text": req.SourceText,
				"replaced_source_text": replacedSourceText,
				"form_id":              req.FormID,
				"editor_id":            req.EditorID,
				"record_type":          req.RecordType,
				"source_plugin":        req.SourcePlugin,
				"source_file":          req.SourceFile,
				"short_name":           req.ShortName,
			},
		})
	}

	if len(cachedResults) > 0 {
		if err := t.store.SaveTerms(ctx, cachedResults); err != nil {
			return nil, PhaseSummary{}, fmt.Errorf("save cached exact matches task_id=%s: %w", taskID, err)
		}
	}

	status := "running"
	if len(llmRequests) == 0 {
		status = "completed"
	}
	summary := PhaseSummary{
		TaskID:          taskID,
		Status:          status,
		TargetCount:     targetCount,
		SavedCount:      cachedGroupCount,
		FailedCount:     0,
		ProgressMode:    progressModeForStatus(status),
		ProgressCurrent: cachedGroupCount,
		ProgressTotal:   targetCount,
		ProgressMessage: progressMessageForStatus(status),
	}
	if status == "running" {
		summary.ProgressMessage = buildProgressMessageWithRemaining(cachedGroupCount, targetCount)
	}
	return llmRequests, summary, nil
}

// SaveResults implementation for terminology phase.
func (t *TermTranslatorImpl) SaveResults(ctx context.Context, taskID string, responses []llmio.Response) error {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.SaveResults")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.SaveResults")

	// Ensure DB schema is ready.
	if err := t.store.InitSchema(ctx); err != nil {
		return fmt.Errorf("failed to init mod term schema: %w", err)
	}

	summary, err := t.store.GetPhaseSummary(ctx, taskID)
	if err != nil {
		return fmt.Errorf("load terminology phase summary before save task_id=%s: %w", taskID, err)
	}

	var finalResults []TermTranslationResult
	failedCount := 0
	savedGroups := 0
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
		savedGroups++

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
	targetCount := summary.TargetCount
	if targetCount <= 0 {
		targetCount = len(responses)
	}
	savedCount := summary.SavedCount + savedGroups
	if savedCount > targetCount {
		savedCount = targetCount
	}
	remaining := targetCount - savedCount
	if remaining < 0 {
		remaining = 0
	}
	finalFailedCount := remaining
	if len(responses) == 0 && targetCount == 0 {
		status = "pending"
	} else if finalFailedCount > 0 || failedCount > 0 {
		status = "completed_partial"
	}
	if err := t.store.UpdatePhaseSummary(ctx, PhaseSummary{
		TaskID:          taskID,
		Status:          status,
		TargetCount:     targetCount,
		SavedCount:      savedCount,
		FailedCount:     finalFailedCount,
		ProgressMode:    progressModeForStatus(status),
		ProgressCurrent: targetCount,
		ProgressTotal:   targetCount,
		ProgressMessage: progressMessageForStatus(status),
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

// GetPreviewTranslations resolves translated text for one preview page.
func (t *TermTranslatorImpl) GetPreviewTranslations(ctx context.Context, entries []TerminologyEntry) (map[string]PreviewTranslation, error) {
	translations, err := t.store.GetPreviewTranslations(ctx, entries)
	if err != nil {
		return nil, fmt.Errorf("get preview translations: %w", err)
	}
	return translations, nil
}

// ListTargets returns normalized terminology preview targets.
func (t *TermTranslatorImpl) ListTargets(ctx context.Context, taskID string) ([]TerminologyEntry, error) {
	artifactInput, err := t.inputRepo.LoadTerminologyInput(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("load terminology artifact input task_id=%s: %w", taskID, err)
	}
	requests, err := t.builder.BuildRequests(ctx, toTerminologyInput(artifactInput))
	if err != nil {
		return nil, fmt.Errorf("build terminology targets task_id=%s: %w", taskID, err)
	}
	targets := make([]TerminologyEntry, 0, len(requests))
	for _, req := range requests {
		targets = append(targets, TerminologyEntry{
			ID:         req.FormID,
			EditorID:   req.EditorID,
			RecordType: req.RecordType,
			SourceText: req.SourceText,
			SourceFile: req.SourceFile,
			Variant:    req.Variant,
		})
	}
	return targets, nil
}

// UpdatePhaseSummary persists a workflow-owned phase snapshot.
func (t *TermTranslatorImpl) UpdatePhaseSummary(ctx context.Context, summary PhaseSummary) error {
	if err := t.store.UpdatePhaseSummary(ctx, summary); err != nil {
		return fmt.Errorf("update terminology phase summary task_id=%s: %w", summary.TaskID, err)
	}
	return nil
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
func (t *TermTranslatorImpl) fetchReferenceTerms(ctx context.Context, req TermTranslationRequest, replacedSourceText string, consumedKeywords []string) []ReferenceTerm {
	keywords := extractKeywords(replacedSourceText)
	contextRefs := make([]ReferenceTerm, 0)

	kwRefs, err := t.searcher.SearchKeywords(ctx, keywords)
	if err == nil {
		contextRefs = append(contextRefs, kwRefs...)
	}
	if strings.HasPrefix(req.RecordType, "NPC") {
		npcRefs, npcErr := t.searcher.SearchNPCPartial(ctx, keywords, consumedKeywords, true)
		if npcErr == nil {
			contextRefs = append(contextRefs, npcRefs...)
		}
	}
	return dedupeReferenceTerms(contextRefs)
}

func (t *TermTranslatorImpl) buildPrompt(ctx context.Context, request TermTranslationRequest, options PhaseOptions) (string, error) {
	promptRequest := request
	if strings.TrimSpace(request.ReplacedSourceText) != "" {
		promptRequest.SourceText = request.ReplacedSourceText
	}
	templateString := strings.TrimSpace(options.Prompt.SystemPrompt)
	if templateString == "" {
		return t.promptBuilder.BuildPrompt(ctx, promptRequest)
	}
	builder, err := NewTermPromptBuilder(templateString)
	if err != nil {
		return "", fmt.Errorf("create terminology prompt builder: %w", err)
	}
	return builder.BuildPrompt(ctx, promptRequest)
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

func progressModeForStatus(status string) string {
	if status == "running" {
		return "indeterminate"
	}
	return "hidden"
}

func progressMessageForStatus(status string) string {
	switch status {
	case "running":
		return "単語翻訳を実行中"
	case "completed":
		return "単語翻訳完了"
	case "completed_partial":
		return "単語翻訳完了（一部失敗あり）"
	case "run_error":
		return "単語翻訳の実行に失敗しました"
	default:
		return ""
	}
}

func (t *TermTranslatorImpl) buildReplacedSourceText(ctx context.Context, sourceText string) (string, []string, error) {
	keywords := extractKeywords(sourceText)
	exactKeywordRefs, err := t.searcher.SearchExactKeywords(ctx, keywords)
	if err != nil {
		return "", nil, err
	}
	matches := t.matcher.MatchSpans(sourceText, exactKeywordRefs)
	if len(matches) == 0 {
		return sourceText, nil, nil
	}

	var builder strings.Builder
	last := 0
	consumed := make([]string, 0, len(matches))
	for _, match := range matches {
		if match.StartIndex < last || match.EndIndex > len(sourceText) {
			continue
		}
		builder.WriteString(sourceText[last:match.StartIndex])
		builder.WriteString(match.Term.Translation)
		last = match.EndIndex
		consumed = append(consumed, match.Term.Source)
	}
	builder.WriteString(sourceText[last:])
	return builder.String(), consumed, nil
}

func extractKeywords(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	rawTokens := strings.Fields(text)
	tokens := make([]string, 0, len(rawTokens))
	for _, token := range rawTokens {
		cleaned := strings.TrimSpace(strings.Trim(token, ".,!?;:\"'()[]{}<>"))
		if cleaned == "" {
			continue
		}
		tokens = append(tokens, cleaned)
	}
	if len(tokens) == 0 {
		return nil
	}

	keywords := make([]string, 0, len(tokens)*(len(tokens)+1)/2)
	seen := make(map[string]struct{}, cap(keywords))
	addCandidate := func(candidate string) {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			return
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		keywords = append(keywords, trimmed)
	}

	// Add longer n-grams first so downstream matching always receives long-form candidates.
	for width := len(tokens); width >= 1; width-- {
		for start := 0; start+width <= len(tokens); start++ {
			addCandidate(strings.Join(tokens[start:start+width], " "))
		}
	}
	return keywords
}

func buildProgressMessageWithRemaining(current int, total int) string {
	if total <= 0 {
		return progressMessageForStatus("running")
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	remaining := total - current
	return fmt.Sprintf("%d / %d 件（残り %d 件）", current, total, remaining)
}
