package masterpersonaartifact

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const artifactSchemaVersion = 1

// Migrate ensures master-persona artifact tables exist in artifact.db.
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
		CREATE TABLE IF NOT EXISTS artifact_master_persona_final (
			persona_id INTEGER PRIMARY KEY AUTOINCREMENT,
			form_id TEXT NOT NULL,
			source_plugin TEXT NOT NULL,
			speaker_id TEXT NOT NULL,
			npc_name TEXT,
			editor_id TEXT,
			race TEXT,
			sex TEXT,
			voice_type TEXT,
			updated_at DATETIME NOT NULL,
			persona_text TEXT NOT NULL DEFAULT '',
			generation_request TEXT NOT NULL DEFAULT '',
			dialogues_json TEXT NOT NULL DEFAULT '[]',
			UNIQUE(source_plugin, speaker_id)
		);
		CREATE INDEX IF NOT EXISTS idx_artifact_master_persona_final_lookup
			ON artifact_master_persona_final(source_plugin, speaker_id);

		CREATE TABLE IF NOT EXISTS artifact_master_persona_temp (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			source_plugin TEXT NOT NULL,
			speaker_id TEXT NOT NULL,
			editor_id TEXT,
			npc_name TEXT,
			race TEXT,
			sex TEXT,
			voice_type TEXT,
			generation_request TEXT NOT NULL DEFAULT '',
			dialogues_json TEXT NOT NULL DEFAULT '[]',
			updated_at DATETIME NOT NULL,
			UNIQUE(task_id, source_plugin, speaker_id)
		);
		CREATE INDEX IF NOT EXISTS idx_artifact_master_persona_temp_task
			ON artifact_master_persona_temp(task_id);
	`); err != nil {
		return fmt.Errorf("create master persona artifact tables: %w", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)`, artifactSchemaVersion, time.Now().UTC()); err != nil {
		return fmt.Errorf("insert artifact schema version: %w", err)
	}
	return nil
}
