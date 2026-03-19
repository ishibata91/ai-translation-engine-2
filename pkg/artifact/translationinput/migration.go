package translationinput

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const artifactSchemaVersion = 1

// Migrate ensures translation-input artifact tables exist in artifact.db.
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
		CREATE TABLE IF NOT EXISTS translation_input_tasks (
			task_id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE IF NOT EXISTS translation_input_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			source_file_path TEXT NOT NULL,
			source_file_name TEXT NOT NULL,
			source_file_hash TEXT NOT NULL,
			parse_status TEXT NOT NULL,
			preview_row_count INTEGER NOT NULL DEFAULT 0,
			parsed_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (task_id) REFERENCES translation_input_tasks(task_id) ON DELETE CASCADE,
			UNIQUE (task_id, source_file_hash)
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_files_task_id ON translation_input_files(task_id);

		CREATE TABLE IF NOT EXISTS translation_input_dialogue_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			player_text TEXT,
			quest_id TEXT,
			is_services_branch INTEGER NOT NULL DEFAULT 0,
			services_type TEXT,
			nam1 TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_dialogue_groups_file_id ON translation_input_dialogue_groups(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_dialogue_responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dialogue_group_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			text TEXT,
			prompt TEXT,
			topic_text TEXT,
			menu_display_text TEXT,
			speaker_id TEXT,
			voice_type TEXT,
			response_order INTEGER NOT NULL,
			previous_id TEXT,
			source TEXT,
			response_index INTEGER,
			FOREIGN KEY (dialogue_group_id) REFERENCES translation_input_dialogue_groups(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_dialogue_responses_group_id ON translation_input_dialogue_responses(dialogue_group_id);

		CREATE TABLE IF NOT EXISTS translation_input_quests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_quests_file_id ON translation_input_quests(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_quest_stages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quest_id INTEGER NOT NULL,
			stage_index INTEGER NOT NULL,
			log_index INTEGER NOT NULL,
			stage_type TEXT NOT NULL,
			text TEXT,
			parent_id TEXT,
			parent_editor_id TEXT,
			source TEXT,
			FOREIGN KEY (quest_id) REFERENCES translation_input_quests(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_quest_stages_quest_id ON translation_input_quest_stages(quest_id);

		CREATE TABLE IF NOT EXISTS translation_input_quest_objectives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quest_id INTEGER NOT NULL,
			objective_index TEXT NOT NULL,
			objective_type TEXT NOT NULL,
			text TEXT,
			parent_id TEXT,
			parent_editor_id TEXT,
			source TEXT,
			FOREIGN KEY (quest_id) REFERENCES translation_input_quests(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_quest_objectives_quest_id ON translation_input_quest_objectives(quest_id);

		CREATE TABLE IF NOT EXISTS translation_input_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			description TEXT,
			text TEXT,
			type_hint TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_items_file_id ON translation_input_items(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_magic (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			description TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_magic_file_id ON translation_input_magic(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_locations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			parent_id TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_locations_file_id ON translation_input_locations(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_cells (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			parent_id TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_cells_file_id ON translation_input_cells(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_system_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			description TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_system_records_file_id ON translation_input_system_records(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			text TEXT,
			title TEXT,
			quest_id TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_messages_file_id ON translation_input_messages(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_load_screens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			text TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_load_screens_file_id ON translation_input_load_screens(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_npcs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			npc_key TEXT NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_json_path TEXT,
			name TEXT,
			race TEXT,
			voice TEXT,
			sex TEXT,
			class_name TEXT,
			source TEXT,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_npcs_file_id ON translation_input_npcs(file_id);

		CREATE TABLE IF NOT EXISTS translation_input_terminology_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			source_record_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT NOT NULL,
			source_text TEXT NOT NULL,
			source_file_name TEXT NOT NULL,
			pair_key TEXT,
			variant TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (file_id) REFERENCES translation_input_files(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_translation_input_terminology_entries_file_id ON translation_input_terminology_entries(file_id);
	`); err != nil {
		return fmt.Errorf("create translation input artifact tables: %w", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT OR IGNORE INTO schema_version (version, applied_at) VALUES (?, ?)`, artifactSchemaVersion, time.Now().UTC()); err != nil {
		return fmt.Errorf("insert artifact schema version: %w", err)
	}
	return nil
}
