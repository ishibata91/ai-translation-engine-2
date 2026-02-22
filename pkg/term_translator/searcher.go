package term_translator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
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
	if text == "" {
		return nil, nil
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT original_en, translated_ja 
		FROM dictionary_terms 
		WHERE original_en = ?
	`, text)
	if err != nil {
		return nil, fmt.Errorf("exact search query failed: %w", err)
	}
	defer rows.Close()

	var results []ReferenceTerm
	for rows.Next() {
		var term ReferenceTerm
		if err := rows.Scan(&term.Source, &term.Translation); err != nil {
			return nil, fmt.Errorf("failed to scan exact match row: %w", err)
		}
		results = append(results, term)
	}

	return results, nil
}

// SearchKeywords searches the dictionary using stemmed keywords.
func (s *SQLiteTermDictionarySearcher) SearchKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	// Apply stemming to keywords
	var stemmedKeywords []string
	for _, kw := range keywords {
		if s.stemmer != nil {
			stemmed, err := s.stemmer.Stem(kw)
			if err == nil && stemmed != "" {
				stemmedKeywords = append(stemmedKeywords, stemmed)
				continue
			}
		}
		// Fallback if no stemmer or error
		stemmedKeywords = append(stemmedKeywords, kw)
	}

	// Use match query over FTS
	matchStr := strings.Join(stemmedKeywords, " OR ")

	rows, err := s.db.QueryContext(ctx, `
		SELECT original_en, translated_ja 
		FROM dictionary_terms_fts 
		WHERE original_en MATCH ?
		LIMIT 20
	`, matchStr)

	if err != nil {
		// Check if table doesn't exist (Dictionary Builder might not have created it)
		if strings.Contains(err.Error(), "no such table") {
			s.logger.WarnContext(ctx, "dictionary_terms_fts table not found, falling back to exact search", "error", err)
			return nil, nil
		}
		return nil, fmt.Errorf("keyword search query failed: %w", err)
	}
	defer rows.Close()

	var results []ReferenceTerm
	for rows.Next() {
		var term ReferenceTerm
		if err := rows.Scan(&term.Source, &term.Translation); err != nil {
			return nil, fmt.Errorf("failed to scan keyword match row: %w", err)
		}
		results = append(results, term)
	}

	return results, nil
}

// SearchNPCPartial searches for NPC names using partial matches.
func (s *SQLiteTermDictionarySearcher) SearchNPCPartial(ctx context.Context, keywords []string, consumedKeywords []string, isNPC bool) ([]ReferenceTerm, error) {
	if !isNPC || len(keywords) == 0 {
		return nil, nil
	}

	// Find available keywords that are not consumed
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

	if len(availableKeywords) == 0 {
		return nil, nil
	}

	matchStr := strings.Join(availableKeywords, " OR ")

	// Search specifically within NPC records
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
		return nil, fmt.Errorf("npc partial search query failed: %w", err)
	}
	defer rows.Close()

	var results []ReferenceTerm
	for rows.Next() {
		var term ReferenceTerm
		if err := rows.Scan(&term.Source, &term.Translation); err != nil {
			return nil, fmt.Errorf("failed to scan npc partial match row: %w", err)
		}
		results = append(results, term)
	}

	return results, nil
}

// SearchBatch executes batched searches for efficiency, returning a map of terms.
func (s *SQLiteTermDictionarySearcher) SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Not implemented optimally for large batches here, would need an IN clause,
	// but simplifies logic for demonstration.
	resultMap := make(map[string][]ReferenceTerm)
	for _, t := range texts {
		res, err := s.SearchExact(ctx, t)
		if err != nil {
			return nil, err
		}
		resultMap[t] = res
	}

	return resultMap, nil
}

// Close closes the dictionary database connection.
func (s *SQLiteTermDictionarySearcher) Close() error {
	return s.db.Close()
}
