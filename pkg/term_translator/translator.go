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
	t.logger.InfoContext(ctx, "Starting term translation")

	// 1. Build translation requests
	requests, err := t.builder.BuildRequests(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to build requests: %w", err)
	}
	t.logger.InfoContext(ctx, "Built translation requests", "count", len(requests))

	if len(requests) == 0 {
		return nil, nil // Nothing to do
	}

	// Ensure DB schema is initialized
	if err := t.store.InitSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to init mod term schema: %w", err)
	}

	var results []TermTranslationResult
	var mu sync.Mutex

	// For progress tracking
	total := len(requests)
	completed := 0

	if t.notifier != nil {
		t.notifier.OnProgress(0, total)
	}

	// 2. Prepare workers channel
	reqChan := make(chan TermTranslationRequest, len(requests))
	for _, req := range requests {
		reqChan <- req
	}
	close(reqChan)

	var wg sync.WaitGroup
	// 3. Start worker pool
	for i := 0; i < t.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range reqChan {
				res := t.processRequest(ctx, req)

				mu.Lock()
				// Expanded paired NPC results
				if res.Status == "success" || res.Status == "cached" {
					if req.RecordType == "NPC_" && req.ShortName != "" {
						// Split the result for FULL and SHRT
						// In a real implementation, LLM would be instructed to return JSON
						// with both full and short name translations. For this simplified
						// implementation, we'll store the same translation or assume
						// LLM generated a combined string we could split.
						// Handling simplified here:

						fullRes := res
						fullRes.RecordType = "NPC_:FULL"
						results = append(results, fullRes)

						shortRes := res
						shortRes.RecordType = "NPC_:SHRT"
						shortRes.SourceText = req.ShortName
						// Just a heuristic for the mock:
						shortRes.TranslatedText = strings.Split(res.TranslatedText, " ")[0]
						results = append(results, shortRes)
					} else {
						// normal single result
						results = append(results, res)
					}
				} else {
					results = append(results, res)
				}

				completed++
				if t.notifier != nil {
					t.notifier.OnProgress(completed, total)
				}
				mu.Unlock()
			}
		}()
	}

	// 4. Wait for completion
	wg.Wait()

	t.logger.InfoContext(ctx, "Term translation completed", "results_count", len(results))

	// 5. Save results to Mod DB
	if err := t.store.SaveTerms(ctx, results); err != nil {
		return results, fmt.Errorf("failed to save terms: %w", err)
	}

	t.logger.InfoContext(ctx, "Saved term translations to mod DB")
	return results, nil
}

func (t *TermTranslatorImpl) processRequest(ctx context.Context, req TermTranslationRequest) TermTranslationResult {
	res := TermTranslationResult{
		FormID:       req.FormID,
		EditorID:     req.EditorID,
		RecordType:   req.RecordType,
		SourceText:   req.SourceText,
		SourcePlugin: req.SourcePlugin,
		SourceFile:   req.SourceFile,
	}

	// Check if already in Mod DB (Skip translation)
	existingTerm, err := t.store.GetTerm(ctx, req.SourceText)
	if err != nil {
		t.logger.WarnContext(ctx, "Failed to get term from Mod DB", "term", req.SourceText, "error", err)
	} else if existingTerm != "" {
		res.TranslatedText = existingTerm
		res.Status = "cached" // Actually "already translated" in this context
		return res
	}

	// Check Exact Match in Reference Dictionary Cache
	refs, err := t.searcher.SearchExact(ctx, req.SourceText)
	if err != nil {
		t.logger.WarnContext(ctx, "Failed to search exact match", "term", req.SourceText, "error", err)
	}

	if len(refs) > 0 {
		// Found exact translation in dictionary, use it without calling LLM
		res.TranslatedText = refs[0].Translation
		res.Status = "cached"
		return res
	}

	// Fetch Reference Terms (Partial / Keyword)
	keywords := strings.Split(req.SourceText, " ")

	var contextRefs []ReferenceTerm
	isNPC := strings.HasPrefix(req.RecordType, "NPC")

	if isNPC {
		// NPC Partial Match (First Name / Last Name logic)
		npcRefs, err := t.searcher.SearchNPCPartial(ctx, keywords, nil, true)
		if err == nil {
			contextRefs = append(contextRefs, npcRefs...)
		}
	} else {
		// Standard Keyword Search
		kwRefs, err := t.searcher.SearchKeywords(ctx, keywords)
		if err == nil {
			contextRefs = append(contextRefs, kwRefs...)
		}
	}

	req.ReferenceTerms = contextRefs

	// Build LLM Prompt
	prompt, err := t.promptBuilder.BuildPrompt(ctx, req)
	if err != nil {
		res.Status = "error"
		res.ErrorMessage = err.Error()
		return res
	}

	// Call LLM
	// Create an empty request body since we pass the prompt directly
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

	content := strings.TrimSpace(llmResp.Content)
	// Extract translation from TL: |...| format
	startIdx := strings.Index(content, "TL: |")
	if startIdx != -1 {
		startIdx += 5 // length of "TL: |"
		endIdx := strings.Index(content[startIdx:], "|")
		if endIdx != -1 {
			res.TranslatedText = strings.TrimSpace(content[startIdx : startIdx+endIdx])
		} else {
			// fallback if closing pipe is missing
			res.TranslatedText = strings.TrimSpace(content[startIdx:])
		}
	} else {
		// fallback if format is not followed
		res.TranslatedText = content
	}

	res.Status = "success"
	return res
}
