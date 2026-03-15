package masterpersonaartifact

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type sqliteRepository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed master persona artifact repository.
func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) SaveOrUpdateTempBase(ctx context.Context, temp TempPersona, overwriteExisting bool) (int64, string, error) {
	key := normalizeLookupKey(LookupKey{SourcePlugin: temp.SourcePlugin, SpeakerID: temp.SpeakerID})
	if key.SpeakerID == "" {
		return 0, "", errors.New("speaker_id is required")
	}
	taskID := strings.TrimSpace(temp.TaskID)
	if taskID == "" {
		return 0, "", errors.New("task_id is required")
	}

	finalPersonaID, finalPersonaText, err := r.getFinalLookupState(ctx, key)
	if err != nil {
		return 0, "", fmt.Errorf("query final persona by lookup: %w", err)
	}
	if finalPersonaID > 0 && strings.TrimSpace(finalPersonaText) != "" && !overwriteExisting {
		return finalPersonaID, finalPersonaText, nil
	}

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO artifact_master_persona_temp (
			task_id, source_plugin, speaker_id, editor_id, npc_name, race, sex, voice_type,
			generation_request, dialogues_json, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, '', '[]', ?)
		ON CONFLICT(task_id, source_plugin, speaker_id) DO UPDATE SET
			editor_id = excluded.editor_id,
			npc_name = excluded.npc_name,
			race = excluded.race,
			sex = excluded.sex,
			voice_type = excluded.voice_type,
			updated_at = excluded.updated_at
	`, taskID, key.SourcePlugin, key.SpeakerID, temp.EditorID, temp.NPCName, temp.Race, temp.Sex, temp.VoiceType, now)
	if err != nil {
		return 0, "", fmt.Errorf("save temp persona base task_id=%s source_plugin=%s speaker_id=%s: %w", taskID, key.SourcePlugin, key.SpeakerID, err)
	}

	if id, err := result.LastInsertId(); err == nil && id > 0 {
		return id, finalPersonaText, nil
	}
	var tempID int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM artifact_master_persona_temp
		WHERE task_id = ? AND source_plugin = ? AND speaker_id = ?
	`, taskID, key.SourcePlugin, key.SpeakerID).Scan(&tempID); err != nil {
		return 0, "", fmt.Errorf("resolve temp persona id task_id=%s source_plugin=%s speaker_id=%s: %w", taskID, key.SourcePlugin, key.SpeakerID, err)
	}
	return tempID, finalPersonaText, nil
}

func (r *sqliteRepository) SaveTempGenerationRequest(ctx context.Context, taskID string, key LookupKey, generationRequest string) error {
	normalizedTaskID := strings.TrimSpace(taskID)
	if normalizedTaskID == "" {
		return errors.New("task_id is required")
	}
	normalizedKey := normalizeLookupKey(key)
	if normalizedKey.SpeakerID == "" {
		return errors.New("speaker_id is required")
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE artifact_master_persona_temp
		SET generation_request = ?, updated_at = ?
		WHERE task_id = ? AND source_plugin = ? AND speaker_id = ?
	`, generationRequest, time.Now().UTC(), normalizedTaskID, normalizedKey.SourcePlugin, normalizedKey.SpeakerID); err != nil {
		return fmt.Errorf("save temp generation request task_id=%s source_plugin=%s speaker_id=%s: %w", normalizedTaskID, normalizedKey.SourcePlugin, normalizedKey.SpeakerID, err)
	}
	return nil
}

func (r *sqliteRepository) ReplaceTempDialogues(ctx context.Context, taskID string, key LookupKey, dialogues []Dialogue) error {
	normalizedTaskID := strings.TrimSpace(taskID)
	if normalizedTaskID == "" {
		return errors.New("task_id is required")
	}
	normalizedKey := normalizeLookupKey(key)
	if normalizedKey.SpeakerID == "" {
		return errors.New("speaker_id is required")
	}
	dialoguesJSON, err := marshalDialogues(dialogues)
	if err != nil {
		return fmt.Errorf("marshal temp dialogues: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE artifact_master_persona_temp
		SET dialogues_json = ?, updated_at = ?
		WHERE task_id = ? AND source_plugin = ? AND speaker_id = ?
	`, dialoguesJSON, time.Now().UTC(), normalizedTaskID, normalizedKey.SourcePlugin, normalizedKey.SpeakerID); err != nil {
		return fmt.Errorf("replace temp dialogues task_id=%s source_plugin=%s speaker_id=%s: %w", normalizedTaskID, normalizedKey.SourcePlugin, normalizedKey.SpeakerID, err)
	}
	return nil
}

func (r *sqliteRepository) ReplaceTempDialoguesByID(ctx context.Context, tempID int64, dialogues []Dialogue) error {
	if tempID <= 0 {
		return errors.New("temp_id is required")
	}
	dialoguesJSON, err := marshalDialogues(dialogues)
	if err != nil {
		return fmt.Errorf("marshal temp dialogues: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE artifact_master_persona_temp
		SET dialogues_json = ?, updated_at = ?
		WHERE id = ?
	`, dialoguesJSON, time.Now().UTC(), tempID); err != nil {
		return fmt.Errorf("replace temp dialogues temp_id=%d: %w", tempID, err)
	}
	return nil
}

func (r *sqliteRepository) SaveOrUpdateFinal(ctx context.Context, taskID string, persona FinalPersona, overwriteExisting bool) error {
	key := normalizeLookupKey(LookupKey{SourcePlugin: persona.SourcePlugin, SpeakerID: persona.SpeakerID})
	if key.SpeakerID == "" {
		return errors.New("speaker_id is required")
	}

	formID := strings.TrimSpace(persona.FormID)
	if formID == "" {
		formID = key.SpeakerID
	}

	finalPersonaID, existingPersonaText, err := r.getFinalLookupState(ctx, key)
	if err != nil {
		return fmt.Errorf("query final persona before save: %w", err)
	}
	if finalPersonaID > 0 && strings.TrimSpace(existingPersonaText) != "" && !overwriteExisting {
		return nil
	}

	temp, err := r.loadTempForFinalization(ctx, strings.TrimSpace(taskID), key)
	if err != nil {
		return fmt.Errorf("load temp for finalization source_plugin=%s speaker_id=%s: %w", key.SourcePlugin, key.SpeakerID, err)
	}

	dialogues := persona.Dialogues
	if len(dialogues) == 0 {
		dialogues = temp.Dialogues
	}
	generationRequest := strings.TrimSpace(persona.GenerationRequest)
	if generationRequest == "" {
		generationRequest = temp.GenerationRequest
	}
	dialoguesJSON, err := marshalDialogues(dialogues)
	if err != nil {
		return fmt.Errorf("marshal final dialogues: %w", err)
	}

	now := time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO artifact_master_persona_final (
			form_id, source_plugin, speaker_id, npc_name, editor_id, race, sex, voice_type,
			updated_at, persona_text, generation_request, dialogues_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_plugin, speaker_id) DO UPDATE SET
			form_id = excluded.form_id,
			npc_name = excluded.npc_name,
			editor_id = excluded.editor_id,
			race = excluded.race,
			sex = excluded.sex,
			voice_type = excluded.voice_type,
			updated_at = excluded.updated_at,
			persona_text = excluded.persona_text,
			generation_request = excluded.generation_request,
			dialogues_json = excluded.dialogues_json
	`, formID, key.SourcePlugin, key.SpeakerID, persona.NPCName, persona.EditorID, persona.Race, persona.Sex, persona.VoiceType, now, persona.PersonaText, generationRequest, dialoguesJSON); err != nil {
		return fmt.Errorf("save final persona source_plugin=%s speaker_id=%s: %w", key.SourcePlugin, key.SpeakerID, err)
	}
	return nil
}

func (r *sqliteRepository) GetFinalPersonaText(ctx context.Context, key LookupKey) (string, error) {
	normalized := normalizeLookupKey(key)
	if normalized.SpeakerID == "" {
		return "", errors.New("speaker_id is required")
	}
	var personaText string
	err := r.db.QueryRowContext(ctx, `
		SELECT persona_text
		FROM artifact_master_persona_final
		WHERE source_plugin = ? AND speaker_id = ?
	`, normalized.SourcePlugin, normalized.SpeakerID).Scan(&personaText)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get final persona text source_plugin=%s speaker_id=%s: %w", normalized.SourcePlugin, normalized.SpeakerID, err)
	}
	return personaText, nil
}

func (r *sqliteRepository) ListFinalPersonas(ctx context.Context) ([]FinalPersona, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT persona_id, form_id, source_plugin, speaker_id, npc_name, editor_id,
		       race, sex, voice_type, updated_at, persona_text, generation_request, dialogues_json
		FROM artifact_master_persona_final
		ORDER BY updated_at DESC, persona_id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list final personas: %w", err)
	}
	defer rows.Close()

	result := make([]FinalPersona, 0)
	for rows.Next() {
		final, err := scanFinalPersona(rows)
		if err != nil {
			return nil, fmt.Errorf("scan final persona row: %w", err)
		}
		result = append(result, final)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate final persona rows: %w", err)
	}
	return result, nil
}

func (r *sqliteRepository) GetFinalByPersonaID(ctx context.Context, personaID int64) (FinalPersona, error) {
	var dialoguesJSON string
	final := FinalPersona{}
	err := r.db.QueryRowContext(ctx, `
		SELECT persona_id, form_id, source_plugin, speaker_id, npc_name, editor_id,
		       race, sex, voice_type, updated_at, persona_text, generation_request, dialogues_json
		FROM artifact_master_persona_final
		WHERE persona_id = ?
	`, personaID).Scan(
		&final.PersonaID,
		&final.FormID,
		&final.SourcePlugin,
		&final.SpeakerID,
		&final.NPCName,
		&final.EditorID,
		&final.Race,
		&final.Sex,
		&final.VoiceType,
		&final.UpdatedAt,
		&final.PersonaText,
		&final.GenerationRequest,
		&dialoguesJSON,
	)
	if err != nil {
		return FinalPersona{}, fmt.Errorf("get final persona by id=%d: %w", personaID, err)
	}
	dialogues, err := unmarshalDialogues(dialoguesJSON)
	if err != nil {
		return FinalPersona{}, fmt.Errorf("decode final dialogues persona_id=%d: %w", personaID, err)
	}
	final.Dialogues = dialogues
	return final, nil
}

func (r *sqliteRepository) FindFinalByLookup(ctx context.Context, key LookupKey) (FinalPersona, error) {
	normalized := normalizeLookupKey(key)
	if normalized.SpeakerID == "" {
		return FinalPersona{}, errors.New("speaker_id is required")
	}
	var dialoguesJSON string
	final := FinalPersona{}
	err := r.db.QueryRowContext(ctx, `
		SELECT persona_id, form_id, source_plugin, speaker_id, npc_name, editor_id,
		       race, sex, voice_type, updated_at, persona_text, generation_request, dialogues_json
		FROM artifact_master_persona_final
		WHERE source_plugin = ? AND speaker_id = ?
	`, normalized.SourcePlugin, normalized.SpeakerID).Scan(
		&final.PersonaID,
		&final.FormID,
		&final.SourcePlugin,
		&final.SpeakerID,
		&final.NPCName,
		&final.EditorID,
		&final.Race,
		&final.Sex,
		&final.VoiceType,
		&final.UpdatedAt,
		&final.PersonaText,
		&final.GenerationRequest,
		&dialoguesJSON,
	)
	if err != nil {
		return FinalPersona{}, fmt.Errorf("find final persona by lookup source_plugin=%s speaker_id=%s: %w", normalized.SourcePlugin, normalized.SpeakerID, err)
	}
	dialogues, err := unmarshalDialogues(dialoguesJSON)
	if err != nil {
		return FinalPersona{}, fmt.Errorf("decode final dialogues source_plugin=%s speaker_id=%s: %w", normalized.SourcePlugin, normalized.SpeakerID, err)
	}
	final.Dialogues = dialogues
	return final, nil
}

func (r *sqliteRepository) CleanupTaskTemp(ctx context.Context, taskID string) error {
	normalizedTaskID := strings.TrimSpace(taskID)
	if normalizedTaskID == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM artifact_master_persona_temp WHERE task_id = ?`, normalizedTaskID); err != nil {
		return fmt.Errorf("cleanup temp artifacts task_id=%s: %w", normalizedTaskID, err)
	}
	return nil
}

func (r *sqliteRepository) ClearAll(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `
		DELETE FROM artifact_master_persona_temp;
		DELETE FROM artifact_master_persona_final;
	`); err != nil {
		return fmt.Errorf("clear master persona artifacts: %w", err)
	}
	return nil
}

func (r *sqliteRepository) getFinalLookupState(ctx context.Context, key LookupKey) (int64, string, error) {
	var personaID int64
	var personaText string
	err := r.db.QueryRowContext(ctx, `
		SELECT persona_id, persona_text
		FROM artifact_master_persona_final
		WHERE source_plugin = ? AND speaker_id = ?
	`, key.SourcePlugin, key.SpeakerID).Scan(&personaID, &personaText)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return personaID, personaText, nil
}

func (r *sqliteRepository) loadTempForFinalization(ctx context.Context, taskID string, key LookupKey) (TempPersona, error) {
	candidates := make([]any, 0, 3)
	query := `
		SELECT id, task_id, source_plugin, speaker_id, editor_id, npc_name, race, sex, voice_type,
		       generation_request, dialogues_json, updated_at
		FROM artifact_master_persona_temp
		WHERE source_plugin = ? AND speaker_id = ?
	`
	candidates = append(candidates, key.SourcePlugin, key.SpeakerID)
	if taskID != "" {
		query += ` AND task_id = ?`
		candidates = append(candidates, taskID)
	}
	query += ` ORDER BY updated_at DESC, id DESC LIMIT 1`

	temp := TempPersona{}
	var dialoguesJSON string
	err := r.db.QueryRowContext(ctx, query, candidates...).Scan(
		&temp.ID,
		&temp.TaskID,
		&temp.SourcePlugin,
		&temp.SpeakerID,
		&temp.EditorID,
		&temp.NPCName,
		&temp.Race,
		&temp.Sex,
		&temp.VoiceType,
		&temp.GenerationRequest,
		&dialoguesJSON,
		&temp.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return TempPersona{}, nil
	}
	if err != nil {
		return TempPersona{}, err
	}
	dialogues, err := unmarshalDialogues(dialoguesJSON)
	if err != nil {
		return TempPersona{}, err
	}
	temp.Dialogues = dialogues
	return temp, nil
}

type finalPersonaScanner interface {
	Scan(dest ...any) error
}

func scanFinalPersona(scanner finalPersonaScanner) (FinalPersona, error) {
	var dialoguesJSON string
	final := FinalPersona{}
	if err := scanner.Scan(
		&final.PersonaID,
		&final.FormID,
		&final.SourcePlugin,
		&final.SpeakerID,
		&final.NPCName,
		&final.EditorID,
		&final.Race,
		&final.Sex,
		&final.VoiceType,
		&final.UpdatedAt,
		&final.PersonaText,
		&final.GenerationRequest,
		&dialoguesJSON,
	); err != nil {
		return FinalPersona{}, fmt.Errorf("scan final persona row: %w", err)
	}
	dialogues, err := unmarshalDialogues(dialoguesJSON)
	if err != nil {
		return FinalPersona{}, fmt.Errorf("decode final dialogues: %w", err)
	}
	final.Dialogues = dialogues
	return final, nil
}

func marshalDialogues(dialogues []Dialogue) (string, error) {
	if dialogues == nil {
		dialogues = []Dialogue{}
	}
	raw, err := json.Marshal(dialogues)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func unmarshalDialogues(raw string) ([]Dialogue, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []Dialogue{}, nil
	}
	dialogues := make([]Dialogue, 0)
	if err := json.Unmarshal([]byte(trimmed), &dialogues); err != nil {
		return nil, err
	}
	return dialogues, nil
}

func normalizeLookupKey(key LookupKey) LookupKey {
	normalized := LookupKey{
		SourcePlugin: strings.TrimSpace(key.SourcePlugin),
		SpeakerID:    strings.TrimSpace(key.SpeakerID),
	}
	if normalized.SourcePlugin == "" {
		normalized.SourcePlugin = "UNKNOWN"
	}
	return normalized
}
