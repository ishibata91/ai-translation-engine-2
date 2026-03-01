package terminology

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
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
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching exact match", slog.String("text", text))

	if text == "" {
		return nil, nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT original_en, translated_ja 
		FROM dictionary_terms 
		WHERE original_en = ?
	`, text)
	if err != nil {
		s.logger.ErrorContext(ctx, "exact search failed", telemetry.ErrorAttrs(err)...)
		return nil, fmt.Errorf("exact search query failed: %w", err)
	}
	defer rows.Close()

	terms, err := s.scanReferenceTermRows(rows)
	if err == nil {
		s.logger.DebugContext(ctx, "exact search completed", slog.Int("match_count", len(terms)))
	}
	return terms, err
}

// SearchKeywords searches the dictionary using stemmed keywords.
func (s *SQLiteTermDictionarySearcher) SearchKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching keywords", slog.Int("keyword_count", len(keywords)))

	if len(keywords) == 0 {
		return nil, nil
	}

	stemmedKeywords := s.stemKeywords(keywords)
	matchStr := strings.Join(stemmedKeywords, " OR ")

	rows, err := s.db.QueryContext(ctx, `
		SELECT original_en, translated_ja 
		FROM dictionary_terms_fts 
		WHERE original_en MATCH ?
		LIMIT 20
	`, matchStr)

	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			s.logger.WarnContext(ctx, "dictionary_terms_fts table not found, falling back to exact search")
			return nil, nil
		}
		s.logger.ErrorContext(ctx, "keyword search failed", telemetry.ErrorAttrs(err)...)
		return nil, fmt.Errorf("keyword search query failed: %w", err)
	}
	defer rows.Close()

	terms, err := s.scanReferenceTermRows(rows)
	if err == nil {
		s.logger.DebugContext(ctx, "keyword search completed", slog.Int("match_count", len(terms)))
	}
	return terms, err
}

// SearchNPCPartial searches for NPC names using partial matches.
func (s *SQLiteTermDictionarySearcher) SearchNPCPartial(ctx context.Context, keywords []string, consumedKeywords []string, isNPC bool) ([]ReferenceTerm, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	s.logger.DebugContext(ctx, "searching NPC partial", slog.Int("keyword_count", len(keywords)))

	if !isNPC || len(keywords) == 0 {
		return nil, nil
	}

	availableKeywords := s.filterConsumedKeywords(keywords, consumedKeywords)
	if len(availableKeywords) == 0 {
		return nil, nil
	}

	matchStr := strings.Join(availableKeywords, " OR ")

	rows, err := s.db.QueryContext(ctx, `
		SELECT dt.original_en, dt.translated_ja 
		FROM dictionary_terms_fts fts
		JOIN dictionary_terms dt ON fts.rowid = dt.rowid
		WHERE fts.original_en MATCH ? AND dt.record_type LIKE 'NPC_%'
		LIMIT 10
	`, matchStr)

	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		s.logger.ErrorContext(ctx, "npc partial search failed", telemetry.ErrorAttrs(err)...)
		return nil, fmt.Errorf("npc partial search query failed: %w", err)
	}
	defer rows.Close()

	terms, err := s.scanReferenceTermRows(rows)
	if err == nil {
		s.logger.DebugContext(ctx, "npc partial search completed", slog.Int("match_count", len(terms)))
	}
	return terms, err
}

// SearchBatch executes batched searches for efficiency, returning a map of terms.
func (s *SQLiteTermDictionarySearcher) SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	s.logger.DebugContext(ctx, "executing batch search", slog.Int("text_count", len(texts)))

	if len(texts) == 0 {
		return nil, nil
	}

	resultMap := make(map[string][]ReferenceTerm)
	for _, t := range texts {
		res, err := s.SearchExact(ctx, t)
		if err != nil {
			return nil, err
		}
		resultMap[t] = res
	}

	s.logger.DebugContext(ctx, "batch search completed", slog.Int("result_map_size", len(resultMap)))
	return resultMap, nil
}

// Close closes the dictionary database connection.
func (s *SQLiteTermDictionarySearcher) Close() error {
	return s.db.Close()
}

// --- Private Helper Methods ---

// stemKeywords applies stemming to each keyword, falling back to the original on error.
func (s *SQLiteTermDictionarySearcher) stemKeywords(keywords []string) []string {
	s.logger.Debug("ENTER SQLiteTermDictionarySearcher.stemKeywords")

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
	s.logger.Debug("ENTER SQLiteTermDictionarySearcher.filterConsumedKeywords")

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
	return results, nil
}
