package terminology

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	dictionaryartifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/dictionary_artifact"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

// SQLiteTermDictionarySearcher implements TermDictionarySearcher via dictionary artifact repository.
type SQLiteTermDictionarySearcher struct {
	repo    dictionaryartifact.Repository
	logger  *slog.Logger
	stemmer KeywordStemmer
}

// NewSQLiteTermDictionarySearcher creates a new SQLiteTermDictionarySearcher.
func NewSQLiteTermDictionarySearcher(repo dictionaryartifact.Repository, logger *slog.Logger, stemmer KeywordStemmer) *SQLiteTermDictionarySearcher {
	return &SQLiteTermDictionarySearcher{
		repo:    repo,
		logger:  logger.With("component", "SQLiteTermDictionarySearcher"),
		stemmer: stemmer,
	}
}

// SearchExact searches for exact matches in the dictionary.
func (s *SQLiteTermDictionarySearcher) SearchExact(ctx context.Context, text string) ([]ReferenceTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching exact match", slog.String("text", text))

	if text == "" {
		return nil, nil
	}

	entries, err := s.repo.FindExactBySourceText(ctx, text)
	if err != nil {
		s.logger.ErrorContext(ctx, "exact search failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("find exact source text: %w", err)
	}

	terms := toReferenceTerms(entries)
	s.logger.DebugContext(ctx, "exact search completed", slog.Int("match_count", len(terms)))
	return terms, nil
}

// SearchExactKeywords searches exact keyword entries without stemming/LIKE fallback.
func (s *SQLiteTermDictionarySearcher) SearchExactKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching exact keywords", slog.Int("keyword_count", len(keywords)))

	if len(keywords) == 0 {
		return nil, nil
	}

	seenKeywords := make(map[string]struct{})
	results := make([]ReferenceTerm, 0, len(keywords))
	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if _, exists := seenKeywords[lower]; exists {
			continue
		}
		seenKeywords[lower] = struct{}{}

		entries, err := s.repo.FindExactBySourceTextCI(ctx, trimmed)
		if err != nil {
			return nil, fmt.Errorf("exact keyword search keyword=%q: %w", trimmed, err)
		}
		results = append(results, toReferenceTerms(entries)...)
	}

	return dedupeReferenceTerms(results), nil
}

// SearchKeywords searches the dictionary using stemmed keywords.
func (s *SQLiteTermDictionarySearcher) SearchKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching keywords", slog.Int("keyword_count", len(keywords)))

	if len(keywords) == 0 {
		return nil, nil
	}

	terms, err := s.searchByKeywords(ctx, s.stemKeywords(keywords), 20, false)
	if err != nil {
		s.logger.ErrorContext(ctx, "keyword search failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("keyword search query failed: %w", err)
	}
	s.logger.DebugContext(ctx, "keyword search completed", slog.Int("match_count", len(terms)))
	return terms, nil
}

// SearchNPCPartial searches for NPC names using partial matches.
func (s *SQLiteTermDictionarySearcher) SearchNPCPartial(ctx context.Context, keywords []string, consumedKeywords []string, isNPC bool) ([]ReferenceTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching NPC partial", slog.Int("keyword_count", len(keywords)))

	if !isNPC || len(keywords) == 0 {
		return nil, nil
	}

	availableKeywords := s.filterConsumedKeywords(keywords, consumedKeywords)
	if len(availableKeywords) == 0 {
		return nil, nil
	}

	terms, err := s.searchByKeywords(ctx, availableKeywords, 10, true)
	if err != nil {
		s.logger.ErrorContext(ctx, "npc partial search failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("npc partial search query failed: %w", err)
	}
	s.logger.DebugContext(ctx, "npc partial search completed", slog.Int("match_count", len(terms)))
	return terms, nil
}

// SearchBatch executes batched searches for efficiency, returning a map of terms.
func (s *SQLiteTermDictionarySearcher) SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	s.logger.DebugContext(ctx, "executing batch search", slog.Int("text_count", len(texts)))

	if len(texts) == 0 {
		return nil, nil
	}

	entries, err := s.repo.FindExactBySourceTexts(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("batch exact search by texts: %w", err)
	}

	resultMap := make(map[string][]ReferenceTerm, len(texts))
	requested := make(map[string]struct{}, len(texts))
	for _, text := range texts {
		resultMap[text] = nil
		requested[text] = struct{}{}
	}
	for _, entry := range entries {
		if _, ok := requested[entry.SourceText]; !ok {
			continue
		}
		resultMap[entry.SourceText] = append(resultMap[entry.SourceText], ReferenceTerm{
			Source:      entry.SourceText,
			Translation: entry.DestText,
		})
	}
	for text, terms := range resultMap {
		resultMap[text] = dedupeReferenceTerms(terms)
	}

	s.logger.DebugContext(ctx, "batch search completed", slog.Int("result_map_size", len(resultMap)))
	return resultMap, nil
}

// Close closes the dictionary database connection.
func (s *SQLiteTermDictionarySearcher) Close() error {
	// No-op: lifecycle is owned by artifact repository provider.
	return nil
}

// --- Private Helper Methods ---

// stemKeywords applies stemming to each keyword, falling back to the original on error.
func (s *SQLiteTermDictionarySearcher) stemKeywords(keywords []string) []string {
	s.logger.DebugContext(context.Background(), "ENTER SQLiteTermDictionarySearcher.stemKeywords")

	var stemmedKeywords []string
	for _, kw := range keywords {
		if s.stemmer != nil {
			stemmed, err := s.stemmer.Stem(kw)
			if err == nil && stemmed != "" {
				stemmedKeywords = append(stemmedKeywords, stemmed)
				continue
			}
		}
		stemmedKeywords = append(stemmedKeywords, kw)
	}
	return stemmedKeywords
}

// filterConsumedKeywords returns keywords that have not been consumed yet.
func (s *SQLiteTermDictionarySearcher) filterConsumedKeywords(keywords []string, consumedKeywords []string) []string {
	s.logger.DebugContext(context.Background(), "ENTER SQLiteTermDictionarySearcher.filterConsumedKeywords")

	consumedMap := make(map[string]bool)
	for _, cw := range consumedKeywords {
		consumedMap[strings.ToLower(cw)] = true
	}

	var availableKeywords []string
	for _, kw := range keywords {
		if !consumedMap[strings.ToLower(kw)] {
			availableKeywords = append(availableKeywords, kw)
		}
	}
	return availableKeywords
}

func toReferenceTerms(entries []dictionaryartifact.Entry) []ReferenceTerm {
	results := make([]ReferenceTerm, 0, len(entries))
	for _, entry := range entries {
		results = append(results, ReferenceTerm{
			Source:      entry.SourceText,
			Translation: entry.DestText,
		})
	}
	return results
}

func (s *SQLiteTermDictionarySearcher) searchByKeywords(ctx context.Context, keywords []string, limit int, npcOnly bool) ([]ReferenceTerm, error) {
	results := make([]ReferenceTerm, 0)
	seen := make(map[string]struct{})

	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			continue
		}

		entries, err := s.repo.SearchBySourceTextLike(ctx, trimmed, limit, npcOnly)
		if err != nil {
			return nil, fmt.Errorf("search by keyword=%q: %w", trimmed, err)
		}
		terms := toReferenceTerms(entries)

		for _, term := range terms {
			key := term.Source + "\x00" + term.Translation
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			results = append(results, term)
			if len(results) >= limit {
				return results, nil
			}
		}
	}

	return results, nil
}

func dedupeReferenceTerms(terms []ReferenceTerm) []ReferenceTerm {
	if len(terms) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(terms))
	deduped := make([]ReferenceTerm, 0, len(terms))
	for _, term := range terms {
		key := term.Source + "\x00" + term.Translation
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, term)
	}
	return deduped
}
