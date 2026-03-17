package terminology

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

// SQLiteTermDictionarySearcher implements TermDictionarySearcher using an existing SQLite dictionary DB.
type SQLiteTermDictionarySearcher struct {
	db      *sql.DB
	logger  *slog.Logger
	stemmer KeywordStemmer
}

// NewSQLiteTermDictionarySearcher creates a new SQLiteTermDictionarySearcher.
func NewSQLiteTermDictionarySearcher(db *sql.DB, logger *slog.Logger, stemmer KeywordStemmer) *SQLiteTermDictionarySearcher {
	return &SQLiteTermDictionarySearcher{
		db:      db,
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

	rows, err := s.db.QueryContext(ctx, `
		SELECT source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_text = ?
	`, text)
	if err != nil {
		s.logger.ErrorContext(ctx, "exact search failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("exact search query failed: %w", err)
	}
	defer rows.Close()

	terms, err := s.scanReferenceTermRows(rows)
	if err == nil {
		s.logger.DebugContext(ctx, "exact search completed", slog.Int("match_count", len(terms)))
	}
	if err != nil {
		return nil, fmt.Errorf("scan exact search results: %w", err)
	}
	return terms, nil
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

	resultMap := make(map[string][]ReferenceTerm)
	for _, t := range texts {
		res, err := s.SearchExact(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("batch exact search text=%q: %w", t, err)
		}
		resultMap[t] = res
	}

	s.logger.DebugContext(ctx, "batch search completed", slog.Int("result_map_size", len(resultMap)))
	return resultMap, nil
}

// Close closes the dictionary database connection.
func (s *SQLiteTermDictionarySearcher) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close terminology dictionary db: %w", err)
	}
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

// scanReferenceTermRows scans rows of (source, translation) pairs into ReferenceTerm slice.
func (s *SQLiteTermDictionarySearcher) scanReferenceTermRows(rows *sql.Rows) ([]ReferenceTerm, error) {
	var results []ReferenceTerm
	for rows.Next() {
		var term ReferenceTerm
		if err := rows.Scan(&term.Source, &term.Translation); err != nil {
			return nil, fmt.Errorf("failed to scan reference term row: %w", err)
		}
		results = append(results, term)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reference term rows: %w", err)
	}
	return results, nil
}

func (s *SQLiteTermDictionarySearcher) searchByKeywords(ctx context.Context, keywords []string, limit int, npcOnly bool) ([]ReferenceTerm, error) {
	results := make([]ReferenceTerm, 0)
	seen := make(map[string]struct{})

	for _, keyword := range keywords {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			continue
		}

		args := []any{"%" + trimmed + "%"}
		args = append(args, limit)

		query := `
			SELECT source_text, dest_text
			FROM artifact_dictionary_entries
			WHERE source_text LIKE ?
			LIMIT ?
		`
		if npcOnly {
			query = `
				SELECT source_text, dest_text
				FROM artifact_dictionary_entries
				WHERE source_text LIKE ? AND record_type LIKE 'NPC_%'
				LIMIT ?
			`
		}

		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query keyword=%q: %w", trimmed, err)
		}

		terms, scanErr := s.scanReferenceTermRows(rows)
		closeErr := rows.Close()
		if scanErr != nil {
			return nil, fmt.Errorf("scan keyword=%q: %w", trimmed, scanErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close keyword=%q rows: %w", trimmed, closeErr)
		}

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
