package dictionaryartifact

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const artifactSchemaVersion = 1

// Migrate ensures dictionary artifact tables exist in artifact.db.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("create artifact schema_version: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS artifact_dictionary_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT NOT NULL,
			format TEXT NOT NULL DEFAULT 'xml',
			file_path TEXT NOT NULL,
			file_size INTEGER NOT NULL DEFAULT 0,
			entry_count INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'PENDING',
			error_message TEXT,
			imported_at DATETIME,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS artifact_dictionary_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL REFERENCES artifact_dictionary_sources(id) ON DELETE CASCADE,
			edid TEXT NOT NULL,
			record_type TEXT NOT NULL,
			source_text TEXT NOT NULL,
			dest_text TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_artifact_dictionary_entries_source_id ON artifact_dictionary_entries(source_id);
	`); err != nil {
		return fmt.Errorf("create dictionary artifact tables: %w", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)`, artifactSchemaVersion, time.Now().UTC()); err != nil {
		return fmt.Errorf("insert artifact schema version: %w", err)
	}
	return nil
}
