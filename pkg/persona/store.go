package persona

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// sqlitePersonaStore implements PersonaStore using SQLite.
type sqlitePersonaStore struct {
	db *sql.DB
}

// NewPersonaStore creates a new PersonaStore instance.
func NewPersonaStore(db *sql.DB) PersonaStore {
	return &sqlitePersonaStore{
		db: db,
	}
}

// InitSchema creates the npc_personas table if it doesn't exist.
func (s *sqlitePersonaStore) InitSchema(ctx context.Context) error {
	slog.DebugContext(ctx, "ENTER InitSchema",
		slog.String("slice", "Persona"),
	)
	start := time.Now()

	query := `
	CREATE TABLE IF NOT EXISTS npc_personas (
		speaker_id TEXT PRIMARY KEY,
		editor_id TEXT,
		npc_name TEXT,
		race TEXT,
		sex TEXT,
		voice_type TEXT,
		persona_text TEXT NOT NULL DEFAULT '',
		dialogue_count INTEGER NOT NULL,
		source_plugin TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS npc_dialogues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		speaker_id TEXT NOT NULL,
		editor_id TEXT,
		record_type TEXT,
		source_text TEXT NOT NULL,
		translated_text TEXT,
		quest_id TEXT,
		is_services_branch INTEGER NOT NULL DEFAULT 0,
		dialogue_order INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (speaker_id) REFERENCES npc_personas(speaker_id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_npc_dialogues_speaker ON npc_dialogues(speaker_id);
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to initialize schema for npc_personas: %w", err)
	}

	slog.DebugContext(ctx, "EXIT InitSchema",
		slog.String("slice", "Persona"),
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// SavePersona inserts or updates a persona generation result in the datastore.
func (s *sqlitePersonaStore) SavePersona(ctx context.Context, result PersonaResult) error {
	slog.DebugContext(ctx, "ENTER SavePersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", result.SpeakerID),
		slog.Int("dialogue_count", result.DialogueCount),
	)
	start := time.Now()

	if result.SpeakerID == "" {
		return errors.New("speaker_id is required for saving persona")
	}

	query := `
	INSERT INTO npc_personas (
		speaker_id, editor_id, npc_name, race, sex, voice_type,
		persona_text, dialogue_count, source_plugin, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(speaker_id) DO UPDATE SET
		editor_id = excluded.editor_id,
		npc_name = excluded.npc_name,
		race = excluded.race,
		sex = excluded.sex,
		voice_type = excluded.voice_type,
		persona_text = excluded.persona_text,
		dialogue_count = excluded.dialogue_count,
		source_plugin = excluded.source_plugin,
		updated_at = excluded.updated_at;
	`

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, query,
		result.SpeakerID,
		result.EditorID,
		result.NPCName,
		result.Race,
		result.Sex,
		result.VoiceType,
		result.PersonaText,
		result.DialogueCount,
		result.SourcePlugin,
		now,
	)

	if err != nil {
		// Log the error concisely
		slog.ErrorContext(ctx, "failed to upsert persona",
			slog.String("speaker_id", result.SpeakerID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to save persona for speaker %s: %w", result.SpeakerID, err)
	}

	slog.DebugContext(ctx, "EXIT SavePersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", result.SpeakerID),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}

func (s *sqlitePersonaStore) SavePersonaBase(ctx context.Context, data NPCDialogueData) error {
	if strings.TrimSpace(data.SpeakerID) == "" {
		return errors.New("speaker_id is required")
	}
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO npc_personas (
			speaker_id, editor_id, npc_name, race, sex, voice_type,
			persona_text, dialogue_count, source_plugin, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(speaker_id) DO UPDATE SET
			editor_id = excluded.editor_id,
			npc_name = excluded.npc_name,
			race = excluded.race,
			sex = excluded.sex,
			voice_type = excluded.voice_type,
			dialogue_count = excluded.dialogue_count,
			updated_at = excluded.updated_at
	`,
		data.SpeakerID,
		data.EditorID,
		data.NPCName,
		data.Race,
		data.Sex,
		data.VoiceType,
		"",
		len(data.Dialogues),
		data.SourcePlugin,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to save persona base for %s: %w", data.SpeakerID, err)
	}
	return nil
}

func (s *sqlitePersonaStore) ReplaceDialogues(ctx context.Context, speakerID string, dialogues []DialogueEntry) error {
	if strings.TrimSpace(speakerID) == "" {
		return errors.New("speaker_id is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin dialogue transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM npc_dialogues WHERE speaker_id = ?`, speakerID); err != nil {
		return fmt.Errorf("failed to delete existing dialogues: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO npc_dialogues (
			speaker_id, editor_id, record_type, source_text, translated_text, quest_id, is_services_branch, dialogue_order, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare dialogue insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, d := range dialogues {
		if strings.TrimSpace(d.Text) == "" && strings.TrimSpace(d.EnglishText) == "" {
			continue
		}
		src := d.EnglishText
		if strings.TrimSpace(src) == "" {
			src = d.Text
		}
		if _, err := stmt.ExecContext(
			ctx,
			speakerID,
			d.EditorID,
			d.RecordType,
			src,
			d.Text,
			d.QuestID,
			boolToInt(d.IsServicesBranch),
			d.Order,
			now,
		); err != nil {
			return fmt.Errorf("failed to insert dialogue for %s: %w", speakerID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit dialogues: %w", err)
	}
	return nil
}

// GetPersona retrieves the persona text for a given speaker ID.
func (s *sqlitePersonaStore) GetPersona(ctx context.Context, speakerID string) (string, error) {
	slog.DebugContext(ctx, "ENTER GetPersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", speakerID),
	)
	start := time.Now()

	if speakerID == "" {
		return "", errors.New("speaker_id must not be empty")
	}

	query := `SELECT persona_text FROM npc_personas WHERE speaker_id = ?`
	var personaText string
	err := s.db.QueryRowContext(ctx, query, speakerID).Scan(&personaText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Not finding a persona is a valid case (we might generate it dynamically in some designs,
			// or just return empty so the consumer knows it doesn't exist).
			return "", nil
		}
		return "", fmt.Errorf("failed to retrieve persona for speaker %s: %w", speakerID, err)
	}

	slog.DebugContext(ctx, "EXIT GetPersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", speakerID),
		slog.Duration("elapsed", time.Since(start)),
		slog.Bool("found", true),
	)

	return personaText, nil
}

func (s *sqlitePersonaStore) ListNPCs(ctx context.Context) ([]PersonaNPCView, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT speaker_id, editor_id, npc_name, race, sex, voice_type, persona_text, dialogue_count, updated_at
		FROM npc_personas
		ORDER BY updated_at DESC, speaker_id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list npcs: %w", err)
	}
	defer rows.Close()

	result := make([]PersonaNPCView, 0)
	for rows.Next() {
		var row PersonaNPCView
		var updatedAt time.Time
		if err := rows.Scan(
			&row.SpeakerID,
			&row.EditorID,
			&row.NPCName,
			&row.Race,
			&row.Sex,
			&row.VoiceType,
			&row.PersonaText,
			&row.DialogueCount,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan npc row: %w", err)
		}
		row.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		result = append(result, row)
	}
	return result, nil
}

func (s *sqlitePersonaStore) ListDialoguesBySpeaker(ctx context.Context, speakerID string) ([]PersonaDialogueView, error) {
	if strings.TrimSpace(speakerID) == "" {
		return nil, errors.New("speaker_id is required")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, speaker_id, editor_id, record_type, source_text, translated_text, quest_id, is_services_branch, dialogue_order
		FROM npc_dialogues
		WHERE speaker_id = ?
		ORDER BY dialogue_order ASC, id ASC
	`, speakerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list dialogues: %w", err)
	}
	defer rows.Close()

	result := make([]PersonaDialogueView, 0)
	for rows.Next() {
		var row PersonaDialogueView
		var services int
		if err := rows.Scan(
			&row.ID,
			&row.SpeakerID,
			&row.EditorID,
			&row.RecordType,
			&row.SourceText,
			&row.TranslatedText,
			&row.QuestID,
			&services,
			&row.DialogueOrder,
		); err != nil {
			return nil, fmt.Errorf("failed to scan dialogue row: %w", err)
		}
		row.IsServicesBranch = services == 1
		result = append(result, row)
	}
	return result, nil
}

// Clear removes all personas from the datastore. Useful for resets.
func (s *sqlitePersonaStore) Clear(ctx context.Context) error {
	slog.DebugContext(ctx, "ENTER Clear",
		slog.String("slice", "Persona"),
	)
	start := time.Now()

	_, err := s.db.ExecContext(ctx, `DELETE FROM npc_personas`)
	if err != nil {
		return fmt.Errorf("failed to clear personas: %w", err)
	}

	slog.DebugContext(ctx, "EXIT Clear",
		slog.String("slice", "Persona"),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
