package persona_gen

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
		slog.String("slice", "PersonaGen"),
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
		persona_text TEXT NOT NULL,
		dialogue_count INTEGER NOT NULL,
		estimated_tokens INTEGER NOT NULL,
		source_plugin TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to initialize schema for npc_personas: %w", err)
	}

	slog.DebugContext(ctx, "EXIT InitSchema",
		slog.String("slice", "PersonaGen"),
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// SavePersona inserts or updates a persona generation result in the database.
func (s *sqlitePersonaStore) SavePersona(ctx context.Context, result PersonaResult) error {
	slog.DebugContext(ctx, "ENTER SavePersona",
		slog.String("slice", "PersonaGen"),
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
		persona_text, dialogue_count, estimated_tokens, source_plugin, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(speaker_id) DO UPDATE SET
		editor_id = excluded.editor_id,
		npc_name = excluded.npc_name,
		race = excluded.race,
		sex = excluded.sex,
		voice_type = excluded.voice_type,
		persona_text = excluded.persona_text,
		dialogue_count = excluded.dialogue_count,
		estimated_tokens = excluded.estimated_tokens,
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
		result.EstimatedTokens,
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
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", result.SpeakerID),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}

// GetPersona retrieves the persona text for a given speaker ID.
func (s *sqlitePersonaStore) GetPersona(ctx context.Context, speakerID string) (string, error) {
	slog.DebugContext(ctx, "ENTER GetPersona",
		slog.String("slice", "PersonaGen"),
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
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", speakerID),
		slog.Duration("elapsed", time.Since(start)),
		slog.Bool("found", true),
	)

	return personaText, nil
}

// Clear removes all personas from the database. Useful for resets.
func (s *sqlitePersonaStore) Clear(ctx context.Context) error {
	slog.DebugContext(ctx, "ENTER Clear",
		slog.String("slice", "PersonaGen"),
	)
	start := time.Now()

	_, err := s.db.ExecContext(ctx, `DELETE FROM npc_personas`)
	if err != nil {
		return fmt.Errorf("failed to clear personas: %w", err)
	}

	slog.DebugContext(ctx, "EXIT Clear",
		slog.String("slice", "PersonaGen"),
		slog.Duration("elapsed", time.Since(start)),
	)

	return nil
}
