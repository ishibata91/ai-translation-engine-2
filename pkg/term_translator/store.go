package term_translator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// SQLiteModTermStore implements ModTermStore using SQLite.
type SQLiteModTermStore struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLiteModTermStore creates a new SQLiteModTermStore.
func NewSQLiteModTermStore(db *sql.DB, logger *slog.Logger) *SQLiteModTermStore {
	return &SQLiteModTermStore{
		db:     db,
		logger: logger.With("component", "SQLiteModTermStore"),
	}
}

// InitSchema creates the necessary tables and FTS5 virtual tables.
func (s *SQLiteModTermStore) InitSchema(ctx context.Context) error {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.InitSchema")
	defer s.logger.DebugContext(ctx, "EXIT SQLiteModTermStore.InitSchema")

	return s.executeSchemaInTransaction(ctx)
}

// SaveTerms inserts or updates translated terms in the database.
func (s *SQLiteModTermStore) SaveTerms(ctx context.Context, results []TermTranslationResult) error {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.SaveTerms", slog.Int("count", len(results)))
	defer s.logger.DebugContext(ctx, "EXIT SQLiteModTermStore.SaveTerms")

	if len(results) == 0 {
		return nil
	}

	return s.prepareAndExecuteUpserts(ctx, results)
}

// GetTerm retrieves a translation by its original English text.
func (s *SQLiteModTermStore) GetTerm(ctx context.Context, originalEN string) (string, error) {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.GetTerm", slog.String("originalEN", originalEN))
	defer s.logger.DebugContext(ctx, "EXIT SQLiteModTermStore.GetTerm")

	var translatedJA string
	err := s.db.QueryRowContext(ctx, `
		SELECT translated_ja FROM mod_terms
		WHERE original_en = ?
		LIMIT 1;
	`, originalEN).Scan(&translatedJA)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not found, return empty string
		}
		return "", fmt.Errorf("failed to query mod_terms: %w", err)
	}

	return translatedJA, nil
}

// Clear removes all data from the database.
func (s *SQLiteModTermStore) Clear(ctx context.Context) error {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.Clear")
	defer s.logger.DebugContext(ctx, "EXIT SQLiteModTermStore.Clear")

	_, err := s.db.ExecContext(ctx, "DELETE FROM mod_terms")
	if err != nil {
		return fmt.Errorf("failed to clear mod_terms table: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (s *SQLiteModTermStore) Close() error {
	return s.db.Close()
}

// --- Private Helper Methods ---

// executeSchemaInTransaction creates all schema objects within a single transaction.
func (s *SQLiteModTermStore) executeSchemaInTransaction(ctx context.Context) error {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.executeSchemaInTransaction")

	queries := s.buildSchemaQueries()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit schema transaction: %w", err)
	}

	s.logger.InfoContext(ctx, "Mod terms schema initialized successfully")
	return nil
}

// buildSchemaQueries returns the SQL statements for creating the mod_terms schema.
func (s *SQLiteModTermStore) buildSchemaQueries() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS mod_terms (
			original_en TEXT NOT NULL,
			record_type TEXT NOT NULL,
			translated_ja TEXT NOT NULL,
			status TEXT NOT NULL,
			PRIMARY KEY (original_en, record_type)
		);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS mod_terms_fts USING fts5(
			original_en,
			translated_ja,
			content='mod_terms',
			content_rowid='rowid'
		);`,
		`CREATE TRIGGER IF NOT EXISTS mod_terms_ai AFTER INSERT ON mod_terms BEGIN
			INSERT INTO mod_terms_fts(rowid, original_en, translated_ja) VALUES (new.rowid, new.original_en, new.translated_ja);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS mod_terms_ad AFTER DELETE ON mod_terms BEGIN
			INSERT INTO mod_terms_fts(mod_terms_fts, rowid, original_en, translated_ja) VALUES('delete', old.rowid, old.original_en, old.translated_ja);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS mod_terms_au AFTER UPDATE ON mod_terms BEGIN
			INSERT INTO mod_terms_fts(mod_terms_fts, rowid, original_en, translated_ja) VALUES('delete', old.rowid, old.original_en, old.translated_ja);
			INSERT INTO mod_terms_fts(rowid, original_en, translated_ja) VALUES (new.rowid, new.original_en, new.translated_ja);
		END;`,
	}
}

// prepareAndExecuteUpserts handles the full upsert flow within a transaction.
func (s *SQLiteModTermStore) prepareAndExecuteUpserts(ctx context.Context, results []TermTranslationResult) error {
	s.logger.DebugContext(ctx, "ENTER SQLiteModTermStore.prepareAndExecuteUpserts")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for SaveTerms: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO mod_terms (original_en, record_type, translated_ja, status)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(original_en, record_type) DO UPDATE SET
			translated_ja=excluded.translated_ja,
			status=excluded.status;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, res := range results {
		if res.Status != "success" && res.Status != "cached" {
			continue
		}

		if _, err := stmt.ExecContext(ctx, res.SourceText, res.RecordType, res.TranslatedText, res.Status); err != nil {
			return fmt.Errorf("failed to execute upsert for term %s: %w", res.SourceText, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit SaveTerms transaction: %w", err)
	}

	return nil
}
