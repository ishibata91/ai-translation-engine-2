package term_translator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

// TermTranslatorImpl implements TermTranslator.
type TermTranslatorImpl struct {
	builder       TermRequestBuilder
	searcher      TermDictionarySearcher
	store         ModTermStore
	llmClient     llm_client.LLMClient
	promptBuilder TermPromptBuilder
	logger        *slog.Logger
	notifier      ProgressNotifier

	// Max parallel workers for LLM calls
	workerCount int
}

// NewTermTranslator creates a new TermTranslatorImpl.
func NewTermTranslator(
	builder TermRequestBuilder,
	searcher TermDictionarySearcher,
	store ModTermStore,
	llmClient llm_client.LLMClient,
	promptBuilder TermPromptBuilder,
	logger *slog.Logger,
	notifier ProgressNotifier,
) *TermTranslatorImpl {
	return &TermTranslatorImpl{
		builder:       builder,
		searcher:      searcher,
		store:         store,
		llmClient:     llmClient,
		promptBuilder: promptBuilder,
		logger:        logger.With("component", "TermTranslatorImpl"),
		notifier:      notifier,
		workerCount:   10, // Default concurrency
	}
}

// TranslateTerms orchestrates the term translation process.
func (t *TermTranslatorImpl) TranslateTerms(ctx context.Context, data models.ExtractedData) ([]TermTranslationResult, error) {
	t.logger.InfoContext(ctx, "ENTER TermTranslatorImpl.TranslateTerms")
	defer t.logger.InfoContext(ctx, "EXIT TermTranslatorImpl.TranslateTerms")

	requests, err := t.initializeTranslation(ctx, data)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, nil
	}

	results := t.runWorkerPool(ctx, requests)

	return t.saveResults(ctx, results)
}

// initializeTranslation builds requests and ensures DB schema is ready.
func (t *TermTranslatorImpl) initializeTranslation(ctx context.Context, data models.ExtractedData) ([]TermTranslationRequest, error) {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.initializeTranslation")

	requests, err := t.builder.BuildRequests(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build requests: %w", err)
	}
	t.logger.InfoContext(ctx, "Built translation requests", "count", len(requests))

	if len(requests) == 0 {
		return nil, nil
	}

	if err := t.store.InitSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to init mod term schema: %w", err)
	}

	return requests, nil
}

// runWorkerPool distributes translation requests across workers and collects results.
func (t *TermTranslatorImpl) runWorkerPool(ctx context.Context, requests []TermTranslationRequest) []TermTranslationResult {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.runWorkerPool", slog.Int("requestCount", len(requests)))

	var results []TermTranslationResult
	var mu sync.Mutex

	total := len(requests)
	completed := 0

	if t.notifier != nil {
		t.notifier.OnProgress(0, total)
	}

	reqChan := make(chan TermTranslationRequest, len(requests))
	for _, req := range requests {
		reqChan <- req
	}
	close(reqChan)

	var wg sync.WaitGroup
	for i := 0; i < t.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range reqChan {
				res := t.processRequest(ctx, req)

				mu.Lock()
				results = append(results, t.expandResult(res, req)...)
				completed++
				if t.notifier != nil {
					t.notifier.OnProgress(completed, total)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	t.logger.InfoContext(ctx, "Term translation completed", "results_count", len(results))
	return results
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
	t.logger.DebugContext(context.Background(), "ENTER TermTranslatorImpl.expandNPCResult", slog.String("editorID", req.EditorID))

	fullRes := res
	fullRes.RecordType = "NPC_:FULL"

	shortRes := res
	shortRes.RecordType = "NPC_:SHRT"
	shortRes.SourceText = req.ShortName
	shortRes.TranslatedText = strings.Split(res.TranslatedText, " ")[0]

	return []TermTranslationResult{fullRes, shortRes}
}

// saveResults persists translation results to the mod DB.
func (t *TermTranslatorImpl) saveResults(ctx context.Context, results []TermTranslationResult) ([]TermTranslationResult, error) {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.saveResults", slog.Int("count", len(results)))

	if err := t.store.SaveTerms(ctx, results); err != nil {
		return results, fmt.Errorf("failed to save terms: %w", err)
	}

	t.logger.InfoContext(ctx, "Saved term translations to mod DB")
	return results, nil
}

// processRequest handles a single translation request through the full pipeline.
func (t *TermTranslatorImpl) processRequest(ctx context.Context, req TermTranslationRequest) TermTranslationResult {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.processRequest", slog.String("sourceText", req.SourceText))

	res := TermTranslationResult{
		FormID:       req.FormID,
		EditorID:     req.EditorID,
		RecordType:   req.RecordType,
		SourceText:   req.SourceText,
		SourcePlugin: req.SourcePlugin,
		SourceFile:   req.SourceFile,
	}

	// Check Mod DB cache
	if cached, ok := t.checkModDBCache(ctx, req, &res); ok {
		return cached
	}

	// Check dictionary cache
	if cached, ok := t.checkDictionaryCache(ctx, req, &res); ok {
		return cached
	}

	// Fetch reference terms for LLM context
	t.fetchReferenceTerms(ctx, &req)

	// Call LLM and extract translation
	return t.callLLMAndExtract(ctx, req, res)
}

// checkModDBCache checks if the term is already translated in the Mod DB.
func (t *TermTranslatorImpl) checkModDBCache(ctx context.Context, req TermTranslationRequest, res *TermTranslationResult) (TermTranslationResult, bool) {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.checkModDBCache", slog.String("term", req.SourceText))

	existingTerm, err := t.store.GetTerm(ctx, req.SourceText)
	if err != nil {
		t.logger.WarnContext(ctx, "Failed to get term from Mod DB", "term", req.SourceText, "error", err)
		return *res, false
	}
	if existingTerm != "" {
		res.TranslatedText = existingTerm
		res.Status = "cached"
		return *res, true
	}
	return *res, false
}

// checkDictionaryCache checks for an exact match in the reference dictionary.
func (t *TermTranslatorImpl) checkDictionaryCache(ctx context.Context, req TermTranslationRequest, res *TermTranslationResult) (TermTranslationResult, bool) {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.checkDictionaryCache", slog.String("term", req.SourceText))

	refs, err := t.searcher.SearchExact(ctx, req.SourceText)
	if err != nil {
		t.logger.WarnContext(ctx, "Failed to search exact match", "term", req.SourceText, "error", err)
		return *res, false
	}

	if len(refs) > 0 {
		res.TranslatedText = refs[0].Translation
		res.Status = "cached"
		return *res, true
	}
	return *res, false
}

// fetchReferenceTerms retrieves context reference terms based on the record type.
func (t *TermTranslatorImpl) fetchReferenceTerms(ctx context.Context, req *TermTranslationRequest) {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.fetchReferenceTerms", slog.String("recordType", req.RecordType))

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

// callLLMAndExtract builds the prompt, calls LLM, and extracts the translation.
func (t *TermTranslatorImpl) callLLMAndExtract(ctx context.Context, req TermTranslationRequest, res TermTranslationResult) TermTranslationResult {
	t.logger.DebugContext(ctx, "ENTER TermTranslatorImpl.callLLMAndExtract", slog.String("term", req.SourceText))

	prompt, err := t.promptBuilder.BuildPrompt(ctx, req)
	if err != nil {
		res.Status = "error"
		res.ErrorMessage = err.Error()
		return res
	}

	llmReq := llm_client.Request{
		SystemPrompt: prompt,
		UserPrompt:   "Translate the provided term.",
	}

	llmResp, err := t.llmClient.Complete(ctx, llmReq)
	if err != nil {
		res.Status = "error"
		res.ErrorMessage = err.Error()
		t.logger.ErrorContext(ctx, "LLM translation failed", "term", req.SourceText, "error", err)
		return res
	}

	res.TranslatedText = t.extractTranslationFromLLMResponse(llmResp.Content)
	res.Status = "success"
	return res
}

// extractTranslationFromLLMResponse extracts the translation text from "TL: |...|" format.
func (t *TermTranslatorImpl) extractTranslationFromLLMResponse(content string) string {
	t.logger.Debug("ENTER TermTranslatorImpl.extractTranslationFromLLMResponse")

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
