package persona

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var pluginNamePattern = regexp.MustCompile(`(?i)[^\\/:*?"<>|\s]+\.(esm|esl|esp)`)

const (
	personaStatusDraft     = "draft"
	personaStatusGenerated = "generated"
)

// sqlitePersonaStore implements PersonaStore using SQLite.
type sqlitePersonaStore struct {
	db *sql.DB
}

// NewPersonaStore creates a new PersonaStore instance.
func NewPersonaStore(db *sql.DB) PersonaStore {
	return &sqlitePersonaStore{db: db}
}

// InitSchema creates the persona tables and recreates incompatible dev schemas.
func (s *sqlitePersonaStore) InitSchema(ctx context.Context) error {
	slog.DebugContext(ctx, "ENTER InitSchema", slog.String("slice", "Persona"))
	start := time.Now()

	needsReset, err := s.schemaNeedsReset(ctx)
	if err != nil {
		return fmt.Errorf("failed to inspect persona schema: %w", err)
	}
	if needsReset {
		if err := s.resetSchema(ctx); err != nil {
			return err
		}
	}

	if _, err := s.db.ExecContext(ctx, `
		PRAGMA foreign_keys = ON;
		CREATE TABLE IF NOT EXISTS npc_personas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			speaker_id TEXT NOT NULL,
			editor_id TEXT,
			npc_name TEXT,
			race TEXT,
			sex TEXT,
			voice_type TEXT,
			persona_text TEXT NOT NULL DEFAULT '',
			generation_request TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			source_plugin TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_plugin, speaker_id)
		);
		CREATE TABLE IF NOT EXISTS npc_dialogues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			persona_id INTEGER NOT NULL,
			source_plugin TEXT NOT NULL,
			speaker_id TEXT NOT NULL,
			editor_id TEXT,
			record_type TEXT,
			source_text TEXT NOT NULL,
			quest_id TEXT,
			is_services_branch INTEGER NOT NULL DEFAULT 0,
			dialogue_order INTEGER NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (persona_id) REFERENCES npc_personas(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_npc_dialogues_persona_id ON npc_dialogues(persona_id);
		CREATE INDEX IF NOT EXISTS idx_npc_dialogues_plugin_speaker ON npc_dialogues(source_plugin, speaker_id);
	`); err != nil {
		return fmt.Errorf("failed to initialize persona schema: %w", err)
	}
	if err := s.migratePersonaSchema(ctx); err != nil {
		return fmt.Errorf("failed to migrate persona schema: %w", err)
	}

	slog.DebugContext(ctx, "EXIT InitSchema",
		slog.String("slice", "Persona"),
		slog.Bool("schema_reset", needsReset),
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// SavePersona inserts or updates a persona generation result in the datastore.
func (s *sqlitePersonaStore) SavePersona(ctx context.Context, result PersonaResult, overwriteExisting bool) error {
	slog.DebugContext(ctx, "ENTER SavePersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", result.SpeakerID),
		slog.String("source_plugin", result.SourcePlugin),
		slog.Bool("overwrite_existing", overwriteExisting),
	)
	start := time.Now()

	pluginName := normalizeSourcePlugin(result.SourcePlugin, "")
	if strings.TrimSpace(result.SpeakerID) == "" {
		return errors.New("speaker_id is required for saving persona")
	}

	var personaID int64
	var existingPersona string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, persona_text
		FROM npc_personas
		WHERE source_plugin = ? AND speaker_id = ?
	`, pluginName, result.SpeakerID).Scan(&personaID, &existingPersona)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO npc_personas (
				speaker_id, editor_id, npc_name, race, sex, voice_type,
				persona_text, status, source_plugin, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			result.SpeakerID,
			result.EditorID,
			result.NPCName,
			result.Race,
			result.Sex,
			result.VoiceType,
			result.PersonaText,
			personaStatusGenerated,
			pluginName,
			time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert persona for speaker %s: %w", result.SpeakerID, err)
		}
	case err != nil:
		return fmt.Errorf("failed to query persona row for speaker %s: %w", result.SpeakerID, err)
	default:
		if strings.TrimSpace(existingPersona) != "" && !overwriteExisting {
			return nil
		}
		_, err = s.db.ExecContext(ctx, `
			UPDATE npc_personas
			SET editor_id = ?,
				npc_name = ?,
				race = ?,
				sex = ?,
				voice_type = ?,
				persona_text = ?,
				status = ?,
				source_plugin = ?,
				updated_at = ?
			WHERE id = ?
		`,
			result.EditorID,
			result.NPCName,
			result.Race,
			result.Sex,
			result.VoiceType,
			result.PersonaText,
			personaStatusGenerated,
			pluginName,
			time.Now().UTC(),
			personaID,
		)
		if err != nil {
			return fmt.Errorf("failed to update persona for speaker %s: %w", result.SpeakerID, err)
		}
	}

	slog.DebugContext(ctx, "EXIT SavePersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", result.SpeakerID),
		slog.String("source_plugin", pluginName),
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

func (s *sqlitePersonaStore) SavePersonaBase(ctx context.Context, data NPCDialogueData, overwriteExisting bool) (PersonaSaveState, error) {
	if strings.TrimSpace(data.SpeakerID) == "" {
		return PersonaSaveState{}, errors.New("speaker_id is required")
	}

	pluginName := normalizeSourcePlugin(data.SourcePlugin, data.SourceHint)
	now := time.Now().UTC()

	var state PersonaSaveState
	err := s.db.QueryRowContext(ctx, `
		SELECT id, persona_text
		FROM npc_personas
		WHERE source_plugin = ? AND speaker_id = ?
	`, pluginName, data.SpeakerID).Scan(&state.PersonaID, &state.PersonaText)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		result, insertErr := s.db.ExecContext(ctx, `
			INSERT INTO npc_personas (
				speaker_id, editor_id, npc_name, race, sex, voice_type,
				persona_text, status, source_plugin, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			data.SpeakerID,
			data.EditorID,
			data.NPCName,
			data.Race,
			data.Sex,
			data.VoiceType,
			"",
			personaStatusDraft,
			pluginName,
			now,
		)
		if insertErr != nil {
			return PersonaSaveState{}, fmt.Errorf("failed to save persona base for %s: %w", data.SpeakerID, insertErr)
		}
		state.PersonaID, insertErr = result.LastInsertId()
		if insertErr != nil {
			return PersonaSaveState{}, fmt.Errorf("failed to resolve persona id for %s: %w", data.SpeakerID, insertErr)
		}
	case err != nil:
		return PersonaSaveState{}, fmt.Errorf("failed to query persona base for %s: %w", data.SpeakerID, err)
	default:
		if strings.TrimSpace(state.PersonaText) != "" && !overwriteExisting {
			return state, nil
		}
		if _, err := s.db.ExecContext(ctx, `
			UPDATE npc_personas
			SET editor_id = ?,
				npc_name = ?,
				race = ?,
				sex = ?,
				voice_type = ?,
				status = ?,
				source_plugin = ?,
				updated_at = ?
			WHERE id = ?
		`,
			data.EditorID,
			data.NPCName,
			data.Race,
			data.Sex,
			data.VoiceType,
			personaStatusDraft,
			pluginName,
			now,
			state.PersonaID,
		); err != nil {
			return PersonaSaveState{}, fmt.Errorf("failed to update persona base for %s: %w", data.SpeakerID, err)
		}
	}

	return state, nil
}

func (s *sqlitePersonaStore) SaveGenerationRequest(ctx context.Context, sourcePlugin string, speakerID string, generationRequest string) error {
	if strings.TrimSpace(speakerID) == "" {
		return errors.New("speaker_id is required")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE npc_personas
		SET generation_request = ?, updated_at = ?
		WHERE source_plugin = ? AND speaker_id = ?
	`, generationRequest, time.Now().UTC(), normalizeSourcePlugin(sourcePlugin, ""), speakerID)
	if err != nil {
		return fmt.Errorf("failed to save generation_request for %s: %w", speakerID, err)
	}
	return nil
}

func (s *sqlitePersonaStore) ReplaceDialogues(ctx context.Context, personaID int64, sourcePlugin string, speakerID string, dialogues []DialogueEntry) error {
	if personaID <= 0 {
		return errors.New("persona_id is required")
	}
	if strings.TrimSpace(speakerID) == "" {
		return errors.New("speaker_id is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin dialogue transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM npc_dialogues WHERE persona_id = ?`, personaID); err != nil {
		return fmt.Errorf("failed to delete existing dialogues: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO npc_dialogues (
			persona_id, source_plugin, speaker_id, editor_id, record_type, source_text, quest_id, is_services_branch, dialogue_order, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare dialogue insert: %w", err)
	}
	defer stmt.Close()

	pluginName := normalizeSourcePlugin(sourcePlugin, "")
	now := time.Now().UTC()
	for _, d := range dialogues {
		src := strings.TrimSpace(d.EnglishText)
		if src == "" {
			src = strings.TrimSpace(d.Text)
		}
		if src == "" {
			continue
		}

		dialoguePlugin := pluginName
		if strings.TrimSpace(d.SourcePlugin) != "" {
			dialoguePlugin = normalizeSourcePlugin(d.SourcePlugin, "")
		}

		if _, err := stmt.ExecContext(
			ctx,
			personaID,
			dialoguePlugin,
			speakerID,
			d.EditorID,
			d.RecordType,
			src,
			d.QuestID,
			boolToInt(d.IsServicesBranch),
			d.Order,
			now,
		); err != nil {
			return fmt.Errorf("failed to insert dialogue for persona %d: %w", personaID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit dialogues: %w", err)
	}
	return nil
}

// GetPersona retrieves the persona text for a given source plugin and speaker ID.
func (s *sqlitePersonaStore) GetPersona(ctx context.Context, sourcePlugin string, speakerID string) (string, error) {
	slog.DebugContext(ctx, "ENTER GetPersona",
		slog.String("slice", "Persona"),
		slog.String("speaker_id", speakerID),
		slog.String("source_plugin", sourcePlugin),
	)
	start := time.Now()

	if strings.TrimSpace(speakerID) == "" {
		return "", errors.New("speaker_id must not be empty")
	}

	query := `SELECT persona_text FROM npc_personas WHERE source_plugin = ? AND speaker_id = ?`
	var personaText string
	err := s.db.QueryRowContext(ctx, query, normalizeSourcePlugin(sourcePlugin, ""), speakerID).Scan(&personaText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
		SELECT
			p.id,
			p.speaker_id,
			p.editor_id,
			p.npc_name,
			p.race,
			p.sex,
			p.voice_type,
			p.persona_text,
			p.generation_request,
			p.status,
			COUNT(d.id) AS dialogue_count,
			p.source_plugin,
			p.updated_at
		FROM npc_personas p
		LEFT JOIN npc_dialogues d ON d.persona_id = p.id
		GROUP BY p.id, p.speaker_id, p.editor_id, p.npc_name, p.race, p.sex, p.voice_type, p.persona_text, p.generation_request, p.status, p.source_plugin, p.updated_at
		ORDER BY p.updated_at DESC, p.id DESC
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
			&row.PersonaID,
			&row.SpeakerID,
			&row.EditorID,
			&row.NPCName,
			&row.Race,
			&row.Sex,
			&row.VoiceType,
			&row.PersonaText,
			&row.GenerationRequest,
			&row.Status,
			&row.DialogueCount,
			&row.SourcePlugin,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan npc row: %w", err)
		}
		row.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		result = append(result, row)
	}
	return result, nil
}

func (s *sqlitePersonaStore) ListDialoguesByPersonaID(ctx context.Context, personaID int64) ([]PersonaDialogueView, error) {
	if personaID <= 0 {
		return nil, errors.New("persona_id is required")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, persona_id, speaker_id, source_plugin, editor_id, record_type, source_text, quest_id, is_services_branch, dialogue_order
		FROM npc_dialogues
		WHERE persona_id = ?
		ORDER BY dialogue_order ASC, id ASC
	`, personaID)
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
			&row.PersonaID,
			&row.SpeakerID,
			&row.SourcePlugin,
			&row.EditorID,
			&row.RecordType,
			&row.SourceText,
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
	slog.DebugContext(ctx, "ENTER Clear", slog.String("slice", "Persona"))
	start := time.Now()

	if _, err := s.db.ExecContext(ctx, `DELETE FROM npc_dialogues; DELETE FROM npc_personas;`); err != nil {
		return fmt.Errorf("failed to clear personas: %w", err)
	}

	slog.DebugContext(ctx, "EXIT Clear",
		slog.String("slice", "Persona"),
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

func (s *sqlitePersonaStore) schemaNeedsReset(ctx context.Context) (bool, error) {
	personaColumns, err := s.tableColumns(ctx, "npc_personas")
	if err != nil {
		return false, err
	}
	dialogueColumns, err := s.tableColumns(ctx, "npc_dialogues")
	if err != nil {
		return false, err
	}
	if len(personaColumns) == 0 || len(dialogueColumns) == 0 {
		return false, nil
	}
	if !personaColumns["id"] || !dialogueColumns["persona_id"] {
		return true, nil
	}
	if dialogueColumns["translated_text"] {
		return true, nil
	}
	return false, nil
}

func (s *sqlitePersonaStore) migratePersonaSchema(ctx context.Context) error {
	personaColumns, err := s.tableColumns(ctx, "npc_personas")
	if err != nil {
		return err
	}
	if len(personaColumns) == 0 {
		return nil
	}
	if personaColumns["generation_request"] && personaColumns["status"] && !personaColumns["dialogue_count"] {
		return nil
	}

	generationRequestExpr := `''`
	if personaColumns["generation_request"] {
		generationRequestExpr = `COALESCE(generation_request, '')`
	}
	statusExpr := fmt.Sprintf(`CASE WHEN TRIM(COALESCE(persona_text, '')) <> '' THEN '%s' ELSE '%s' END`, personaStatusGenerated, personaStatusDraft)
	if personaColumns["status"] {
		statusExpr = fmt.Sprintf(`COALESCE(NULLIF(TRIM(status), ''), CASE WHEN TRIM(COALESCE(persona_text, '')) <> '' THEN '%s' ELSE '%s' END)`, personaStatusGenerated, personaStatusDraft)
	}

	if _, err := s.db.ExecContext(ctx, `PRAGMA foreign_keys = OFF;`); err != nil {
		return err
	}
	defer s.db.ExecContext(context.Background(), `PRAGMA foreign_keys = ON;`)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	statements := []string{
		`CREATE TABLE npc_personas_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			speaker_id TEXT NOT NULL,
			editor_id TEXT,
			npc_name TEXT,
			race TEXT,
			sex TEXT,
			voice_type TEXT,
			persona_text TEXT NOT NULL DEFAULT '',
			generation_request TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'draft',
			source_plugin TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source_plugin, speaker_id)
		);`,
		fmt.Sprintf(`
			INSERT INTO npc_personas_new (
				id, speaker_id, editor_id, npc_name, race, sex, voice_type,
				persona_text, generation_request, status, source_plugin, updated_at
			)
			SELECT
				id, speaker_id, editor_id, npc_name, race, sex, voice_type,
				persona_text, %s, %s, source_plugin, updated_at
			FROM npc_personas;
		`, generationRequestExpr, statusExpr),
		`DROP TABLE npc_personas;`,
		`ALTER TABLE npc_personas_new RENAME TO npc_personas;`,
	}
	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *sqlitePersonaStore) resetSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
		DROP TABLE IF EXISTS npc_dialogues;
		DROP TABLE IF EXISTS npc_personas;
	`); err != nil {
		return fmt.Errorf("failed to recreate persona tables: %w", err)
	}
	return nil
}

func (s *sqlitePersonaStore) tableColumns(ctx context.Context, tableName string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, nil
}

func normalizeSourcePlugin(sourcePlugin string, sourceHint string) string {
	candidate := strings.TrimSpace(sourcePlugin)
	if candidate == "" {
		candidate = strings.TrimSpace(sourceHint)
	}
	match := pluginNamePattern.FindString(candidate)
	if match != "" {
		return match
	}
	if strings.TrimSpace(sourcePlugin) != "" {
		return strings.TrimSpace(sourcePlugin)
	}
	return "UNKNOWN"
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
