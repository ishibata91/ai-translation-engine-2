package translationinput

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation"
)

const defaultPreviewPageSize = 50

type sqliteRepository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed translation input repository.
func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

// EnsureTask creates task parent row when it does not exist.
func (r *sqliteRepository) EnsureTask(ctx context.Context, taskID string) error {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return fmt.Errorf("task_id is required")
	}

	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO translation_input_tasks (task_id, status, created_at, updated_at)
		VALUES (?, 'pending', ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET updated_at = excluded.updated_at
	`, trimmedTaskID, now, now)
	if err != nil {
		return fmt.Errorf("ensure translation input task task_id=%s: %w", trimmedTaskID, err)
	}
	return nil
}

// SaveParsedOutput replaces one source file payload under the specified task.
func (r *sqliteRepository) SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (InputFile, error) {
	if output == nil {
		return InputFile{}, fmt.Errorf("parser output is required")
	}

	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return InputFile{}, fmt.Errorf("task_id is required")
	}

	normalizedPath := strings.TrimSpace(filepath.Clean(sourceFilePath))
	if normalizedPath == "" {
		return InputFile{}, fmt.Errorf("source_file_path is required")
	}

	if err := r.EnsureTask(ctx, trimmedTaskID); err != nil {
		return InputFile{}, fmt.Errorf("ensure task before save parsed output task_id=%s: %w", trimmedTaskID, err)
	}

	fileName := filepath.Base(normalizedPath)
	fileHash := hashFilePath(normalizedPath)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return InputFile{}, fmt.Errorf("begin transaction task_id=%s file=%s: %w", trimmedTaskID, normalizedPath, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	existingFileID, hasExisting, err := r.findExistingFile(ctx, tx, trimmedTaskID, fileHash)
	if err != nil {
		return InputFile{}, fmt.Errorf("find existing file before save task_id=%s file_hash=%s: %w", trimmedTaskID, fileHash, err)
	}
	if hasExisting {
		if _, err := tx.ExecContext(ctx, `DELETE FROM translation_input_files WHERE id = ?`, existingFileID); err != nil {
			return InputFile{}, fmt.Errorf("delete existing translation file id=%d: %w", existingFileID, err)
		}
	}

	insertResult, err := tx.ExecContext(ctx, `
		INSERT INTO translation_input_files (
			task_id, source_file_path, source_file_name, source_file_hash,
			parse_status, preview_row_count, parsed_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, 'loaded', 0, ?, ?, ?)
	`, trimmedTaskID, normalizedPath, fileName, fileHash, now, now, now)
	if err != nil {
		return InputFile{}, fmt.Errorf("insert translation file task_id=%s file=%s: %w", trimmedTaskID, normalizedPath, err)
	}

	fileID, err := insertResult.LastInsertId()
	if err != nil {
		return InputFile{}, fmt.Errorf("resolve inserted translation file id task_id=%s file=%s: %w", trimmedTaskID, normalizedPath, err)
	}

	if err := r.insertDialogue(ctx, tx, fileID, output.DialogueGroups); err != nil {
		return InputFile{}, fmt.Errorf("insert dialogue file_id=%d: %w", fileID, err)
	}
	if err := r.insertQuests(ctx, tx, fileID, output.Quests); err != nil {
		return InputFile{}, fmt.Errorf("insert quests file_id=%d: %w", fileID, err)
	}
	if err := r.insertItems(ctx, tx, fileID, output.Items); err != nil {
		return InputFile{}, fmt.Errorf("insert items file_id=%d: %w", fileID, err)
	}
	if err := r.insertMagic(ctx, tx, fileID, output.Magic); err != nil {
		return InputFile{}, fmt.Errorf("insert magic file_id=%d: %w", fileID, err)
	}
	if err := r.insertLocations(ctx, tx, fileID, output.Locations); err != nil {
		return InputFile{}, fmt.Errorf("insert locations file_id=%d: %w", fileID, err)
	}
	if err := r.insertCells(ctx, tx, fileID, output.Cells); err != nil {
		return InputFile{}, fmt.Errorf("insert cells file_id=%d: %w", fileID, err)
	}
	if err := r.insertSystemRecords(ctx, tx, fileID, output.System); err != nil {
		return InputFile{}, fmt.Errorf("insert system records file_id=%d: %w", fileID, err)
	}
	if err := r.insertMessages(ctx, tx, fileID, output.Messages); err != nil {
		return InputFile{}, fmt.Errorf("insert messages file_id=%d: %w", fileID, err)
	}
	if err := r.insertLoadScreens(ctx, tx, fileID, output.LoadScreens); err != nil {
		return InputFile{}, fmt.Errorf("insert load screens file_id=%d: %w", fileID, err)
	}
	if err := r.insertNPCs(ctx, tx, fileID, output.NPCs); err != nil {
		return InputFile{}, fmt.Errorf("insert npcs file_id=%d: %w", fileID, err)
	}
	if err := r.insertTerminologyEntries(ctx, tx, fileID, fileName, output); err != nil {
		return InputFile{}, fmt.Errorf("insert terminology entries file_id=%d: %w", fileID, err)
	}

	previewCount, err := r.countPreviewRows(ctx, tx, fileID)
	if err != nil {
		return InputFile{}, fmt.Errorf("count preview rows after save file_id=%d: %w", fileID, err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE translation_input_files
		SET preview_row_count = ?, updated_at = ?
		WHERE id = ?
	`, previewCount, now, fileID); err != nil {
		return InputFile{}, fmt.Errorf("update preview_row_count file_id=%d: %w", fileID, err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE translation_input_tasks
		SET status = 'loaded', updated_at = ?
		WHERE task_id = ?
	`, now, trimmedTaskID); err != nil {
		return InputFile{}, fmt.Errorf("update translation input task status task_id=%s: %w", trimmedTaskID, err)
	}

	if err := tx.Commit(); err != nil {
		return InputFile{}, fmt.Errorf("commit translation file save task_id=%s file=%s: %w", trimmedTaskID, normalizedPath, err)
	}

	return InputFile{
		ID:              fileID,
		TaskID:          trimmedTaskID,
		SourceFilePath:  normalizedPath,
		SourceFileName:  fileName,
		SourceFileHash:  fileHash,
		ParseStatus:     "loaded",
		PreviewRowCount: previewCount,
	}, nil
}

// ListFiles returns parsed file rows for one task.
func (r *sqliteRepository) ListFiles(ctx context.Context, taskID string) ([]InputFile, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, source_file_path, source_file_name, source_file_hash, parse_status, preview_row_count
		FROM translation_input_files
		WHERE task_id = ?
		ORDER BY parsed_at ASC, id ASC
	`, trimmedTaskID)
	if err != nil {
		return nil, fmt.Errorf("list translation files task_id=%s: %w", trimmedTaskID, err)
	}
	defer rows.Close()

	files := make([]InputFile, 0)
	for rows.Next() {
		var file InputFile
		if err := rows.Scan(
			&file.ID,
			&file.TaskID,
			&file.SourceFilePath,
			&file.SourceFileName,
			&file.SourceFileHash,
			&file.ParseStatus,
			&file.PreviewRowCount,
		); err != nil {
			return nil, fmt.Errorf("scan translation file row task_id=%s: %w", trimmedTaskID, err)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate translation file rows task_id=%s: %w", trimmedTaskID, err)
	}
	return files, nil
}

// ListPreviewRows returns paginated preview rows projected from all parser sections.
func (r *sqliteRepository) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error) {
	if fileID <= 0 {
		return PreviewPage{}, fmt.Errorf("file_id must be positive")
	}

	safePage := page
	if safePage <= 0 {
		safePage = 1
	}
	safePageSize := pageSize
	if safePageSize <= 0 {
		safePageSize = defaultPreviewPageSize
	}

	totalRows, err := r.countPreviewRows(ctx, r.db, fileID)
	if err != nil {
		return PreviewPage{}, fmt.Errorf("count preview rows before list file_id=%d: %w", fileID, err)
	}

	offset := (safePage - 1) * safePageSize
	query := previewUnionSQL + ` ORDER BY section ASC, row_id ASC LIMIT ? OFFSET ?`
	args := append(previewUnionArgs(fileID), safePageSize, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return PreviewPage{}, fmt.Errorf("list preview rows file_id=%d page=%d size=%d: %w", fileID, safePage, safePageSize, err)
	}
	defer rows.Close()

	previewRows := make([]PreviewRow, 0)
	for rows.Next() {
		var row PreviewRow
		if err := rows.Scan(&row.ID, &row.Section, &row.RecordType, &row.EditorID, &row.SourceText); err != nil {
			return PreviewPage{}, fmt.Errorf("scan preview row file_id=%d: %w", fileID, err)
		}
		previewRows = append(previewRows, row)
	}
	if err := rows.Err(); err != nil {
		return PreviewPage{}, fmt.Errorf("iterate preview rows file_id=%d: %w", fileID, err)
	}

	return PreviewPage{
		FileID:    fileID,
		Page:      safePage,
		PageSize:  safePageSize,
		TotalRows: totalRows,
		Rows:      previewRows,
	}, nil
}

// LoadTerminologyInput projects terminology-phase targets from saved translation input rows.
func (r *sqliteRepository) LoadTerminologyInput(ctx context.Context, taskID string) (TerminologyInput, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return TerminologyInput{}, fmt.Errorf("task_id is required")
	}

	files, err := r.ListFiles(ctx, trimmedTaskID)
	if err != nil {
		return TerminologyInput{}, fmt.Errorf("list terminology source files task_id=%s: %w", trimmedTaskID, err)
	}
	input := TerminologyInput{
		TaskID:    trimmedTaskID,
		FileNames: make([]string, 0, len(files)),
		Entries:   make([]TerminologyEntry, 0),
	}
	for _, file := range files {
		input.FileNames = append(input.FileNames, file.SourceFileName)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT e.source_record_id, COALESCE(e.editor_id, ''), e.record_type, e.source_text, f.source_file_name, COALESCE(e.pair_key, ''), e.variant
		FROM translation_input_terminology_entries e
		JOIN translation_input_files f ON f.id = e.file_id
		WHERE f.task_id = ?
		ORDER BY f.id, e.id
	`, trimmedTaskID)
	if err != nil {
		return TerminologyInput{}, fmt.Errorf("load terminology entries task_id=%s: %w", trimmedTaskID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry TerminologyEntry
		if err := rows.Scan(&entry.ID, &entry.EditorID, &entry.RecordType, &entry.SourceText, &entry.SourceFile, &entry.PairKey, &entry.Variant); err != nil {
			return TerminologyInput{}, fmt.Errorf("scan terminology entry task_id=%s: %w", trimmedTaskID, err)
		}
		entry.RecordType = normalizeTerminologyRecordType(entry.RecordType)
		input.Entries = append(input.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return TerminologyInput{}, fmt.Errorf("iterate terminology entries task_id=%s: %w", trimmedTaskID, err)
	}

	return input, nil
}

// LoadPersonaInput projects raw persona candidates from saved translation input rows.
func (r *sqliteRepository) LoadPersonaInput(ctx context.Context, taskID string) (PersonaInput, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return PersonaInput{}, fmt.Errorf("task_id is required")
	}

	files, err := r.ListFiles(ctx, trimmedTaskID)
	if err != nil {
		return PersonaInput{}, fmt.Errorf("list persona source files task_id=%s: %w", trimmedTaskID, err)
	}

	input := PersonaInput{
		TaskID:    trimmedTaskID,
		NPCs:      make(map[string]PersonaNPC, len(files)),
		Dialogues: make([]PersonaDialogue, 0),
	}

	npcRows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(n.source_record_id, ''),
			COALESCE(n.npc_key, ''),
			COALESCE(n.editor_id, ''),
			COALESCE(n.record_type, ''),
			COALESCE(n.name, ''),
			COALESCE(n.race, ''),
			COALESCE(n.sex, ''),
			COALESCE(n.voice, ''),
			COALESCE(f.source_file_name, ''),
			COALESCE(n.source_json_path, ''),
			COALESCE(n.source, '')
		FROM translation_input_npcs n
		JOIN translation_input_files f ON f.id = n.file_id
		WHERE f.task_id = ?
		ORDER BY f.id, n.id
	`, trimmedTaskID)
	if err != nil {
		return PersonaInput{}, fmt.Errorf("load persona npcs task_id=%s: %w", trimmedTaskID, err)
	}
	defer npcRows.Close()

	for npcRows.Next() {
		var npc PersonaNPC
		var sourceFileName string
		var sourceJSONPath string
		var source string
		if err := npcRows.Scan(
			&npc.SourceRecordID,
			&npc.NPCKey,
			&npc.EditorID,
			&npc.RecordType,
			&npc.NPCName,
			&npc.Race,
			&npc.Sex,
			&npc.VoiceType,
			&sourceFileName,
			&sourceJSONPath,
			&source,
		); err != nil {
			return PersonaInput{}, fmt.Errorf("scan persona npc task_id=%s: %w", trimmedTaskID, err)
		}

		npc.SpeakerID = strings.TrimSpace(npc.SourceRecordID)
		if npc.SpeakerID == "" {
			continue
		}
		npc.SourcePlugin = resolvePersonaSourcePlugin(sourceFileName, sourceJSONPath, source)
		npc.SourceHint = resolvePersonaSourceHint(sourceFileName, sourceJSONPath, source)
		input.NPCs[npc.SpeakerID] = npc
	}
	if err := npcRows.Err(); err != nil {
		return PersonaInput{}, fmt.Errorf("iterate persona npcs task_id=%s: %w", trimmedTaskID, err)
	}

	dialogueRows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(r.source_record_id, ''),
			COALESCE(r.speaker_id, ''),
			COALESCE(r.editor_id, ''),
			COALESCE(g.editor_id, ''),
			COALESCE(r.record_type, ''),
			COALESCE(r.text, ''),
			COALESCE(g.quest_id, ''),
			g.is_services_branch,
			r.response_order,
			COALESCE(f.source_file_name, ''),
			COALESCE(r.source_json_path, ''),
			COALESCE(g.source_json_path, ''),
			COALESCE(r.source, ''),
			COALESCE(g.source, '')
		FROM translation_input_dialogue_responses r
		JOIN translation_input_dialogue_groups g ON g.id = r.dialogue_group_id
		JOIN translation_input_files f ON f.id = g.file_id
		WHERE f.task_id = ?
		  AND TRIM(COALESCE(r.text, '')) <> ''
		  AND TRIM(COALESCE(r.speaker_id, '')) <> ''
		ORDER BY f.id, g.id, r.response_order, r.id
	`, trimmedTaskID)
	if err != nil {
		return PersonaInput{}, fmt.Errorf("load persona dialogues task_id=%s: %w", trimmedTaskID, err)
	}
	defer dialogueRows.Close()

	for dialogueRows.Next() {
		var dialogue PersonaDialogue
		var isServicesBranch int
		var sourceFileName string
		var responseSourceJSONPath string
		var groupSourceJSONPath string
		var responseSource string
		var groupSource string
		if err := dialogueRows.Scan(
			&dialogue.ID,
			&dialogue.SpeakerID,
			&dialogue.EditorID,
			&dialogue.GroupEditorID,
			&dialogue.RecordType,
			&dialogue.Text,
			&dialogue.QuestID,
			&isServicesBranch,
			&dialogue.Order,
			&sourceFileName,
			&responseSourceJSONPath,
			&groupSourceJSONPath,
			&responseSource,
			&groupSource,
		); err != nil {
			return PersonaInput{}, fmt.Errorf("scan persona dialogue task_id=%s: %w", trimmedTaskID, err)
		}

		dialogue.SpeakerID = strings.TrimSpace(dialogue.SpeakerID)
		if dialogue.SpeakerID == "" {
			continue
		}
		dialogue.IsServicesBranch = isServicesBranch != 0
		dialogue.SourcePlugin = resolvePersonaSourcePlugin(sourceFileName, responseSourceJSONPath, groupSourceJSONPath, responseSource, groupSource)
		dialogue.SourceHint = resolvePersonaSourceHint(sourceFileName, responseSourceJSONPath, groupSourceJSONPath, responseSource, groupSource)
		input.Dialogues = append(input.Dialogues, dialogue)
	}
	if err := dialogueRows.Err(); err != nil {
		return PersonaInput{}, fmt.Errorf("iterate persona dialogues task_id=%s: %w", trimmedTaskID, err)
	}

	return input, nil
}

func (r *sqliteRepository) insertTerminologyEntries(ctx context.Context, tx *sql.Tx, fileID int64, sourceFileName string, output *skyrim.ParserOutput) error {
	entries := terminologyEntriesFromOutput(output, sourceFileName)
	now := time.Now().UTC()
	for _, entry := range entries {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_terminology_entries (
				file_id, source_record_id, editor_id, record_type, source_text, source_file_name, pair_key, variant, created_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			entry.ID,
			nullableStringValue(entry.EditorID),
			entry.RecordType,
			entry.SourceText,
			entry.SourceFile,
			nullableStringValue(entry.PairKey),
			entry.Variant,
			now,
		); err != nil {
			return fmt.Errorf("insert terminology entry file_id=%d source_record_id=%s: %w", fileID, entry.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) findExistingFile(ctx context.Context, tx *sql.Tx, taskID string, sourceHash string) (int64, bool, error) {
	var fileID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM translation_input_files
		WHERE task_id = ? AND source_file_hash = ?
	`, taskID, sourceHash).Scan(&fileID)
	if err == nil {
		return fileID, true, nil
	}
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	return 0, false, fmt.Errorf("find existing translation file task_id=%s source_hash=%s: %w", taskID, sourceHash, err)
}

func (r *sqliteRepository) insertDialogue(ctx context.Context, tx *sql.Tx, fileID int64, groups []skyrim.DialogueGroup) error {
	for _, group := range groups {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_dialogue_groups (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				player_text, quest_id, is_services_branch, services_type, nam1, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			group.ID,
			nullableString(group.EditorID),
			group.Type,
			nullableStringValue(group.SourceJSON),
			nullableString(group.PlayerText),
			nullableString(group.QuestID),
			boolToInt(group.IsServicesBranch),
			nullableString(group.ServicesType),
			nullableString(group.NAM1),
			nullableString(group.Source),
		)
		if err != nil {
			return fmt.Errorf("insert dialogue group file_id=%d source_record_id=%s: %w", fileID, group.ID, err)
		}
		dialogueGroupID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("resolve dialogue group id file_id=%d source_record_id=%s: %w", fileID, group.ID, err)
		}
		for _, response := range group.Responses {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO translation_input_dialogue_responses (
					dialogue_group_id, source_record_id, editor_id, record_type, source_json_path,
					text, prompt, topic_text, menu_display_text, speaker_id, voice_type,
					response_order, previous_id, source, response_index
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				dialogueGroupID,
				response.ID,
				nullableString(response.EditorID),
				response.Type,
				nullableStringValue(response.SourceJSON),
				response.Text,
				nullableString(response.Prompt),
				nullableString(response.TopicText),
				nullableString(response.MenuDisplayText),
				nullableString(response.SpeakerID),
				nullableString(response.VoiceType),
				response.Order,
				nullableString(response.PreviousID),
				nullableString(response.Source),
				nullableInt(response.Index),
			); err != nil {
				return fmt.Errorf("insert dialogue response file_id=%d response_record_id=%s: %w", fileID, response.ID, err)
			}
		}
	}
	return nil
}

func (r *sqliteRepository) insertQuests(ctx context.Context, tx *sql.Tx, fileID int64, quests []skyrim.Quest) error {
	for _, quest := range quests {
		result, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_quests (
				file_id, source_record_id, editor_id, record_type, source_json_path, name, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			quest.ID,
			nullableString(quest.EditorID),
			quest.Type,
			nullableStringValue(quest.SourceJSON),
			nullableString(quest.Name),
			nullableString(quest.Source),
		)
		if err != nil {
			return fmt.Errorf("insert quest file_id=%d source_record_id=%s: %w", fileID, quest.ID, err)
		}
		questID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("resolve quest id file_id=%d source_record_id=%s: %w", fileID, quest.ID, err)
		}

		for _, stage := range quest.Stages {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO translation_input_quest_stages (
					quest_id, stage_index, log_index, stage_type, text, parent_id, parent_editor_id, source
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`,
				questID,
				stage.StageIndex,
				stage.LogIndex,
				stage.Type,
				stage.Text,
				nullableStringValue(stage.ParentID),
				nullableStringValue(stage.ParentEditorID),
				nullableString(stage.Source),
			); err != nil {
				return fmt.Errorf("insert quest stage file_id=%d quest_record_id=%s stage_index=%d: %w", fileID, quest.ID, stage.StageIndex, err)
			}
		}

		for _, objective := range quest.Objectives {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO translation_input_quest_objectives (
					quest_id, objective_index, objective_type, text, parent_id, parent_editor_id, source
				)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`,
				questID,
				objective.Index,
				objective.Type,
				objective.Text,
				nullableStringValue(objective.ParentID),
				nullableStringValue(objective.ParentEditorID),
				nullableString(objective.Source),
			); err != nil {
				return fmt.Errorf("insert quest objective file_id=%d quest_record_id=%s objective_index=%s: %w", fileID, quest.ID, objective.Index, err)
			}
		}
	}
	return nil
}

func (r *sqliteRepository) insertItems(ctx context.Context, tx *sql.Tx, fileID int64, items []skyrim.Item) error {
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_items (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				name, description, text, type_hint, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			item.ID,
			nullableString(item.EditorID),
			item.Type,
			nullableStringValue(item.SourceJSON),
			nullableString(item.Name),
			nullableString(item.Description),
			nullableString(item.Text),
			nullableString(item.TypeHint),
			nullableString(item.Source),
		); err != nil {
			return fmt.Errorf("insert item file_id=%d source_record_id=%s: %w", fileID, item.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertMagic(ctx context.Context, tx *sql.Tx, fileID int64, magicRecords []skyrim.Magic) error {
	for _, record := range magicRecords {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_magic (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				name, description, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			record.ID,
			nullableString(record.EditorID),
			record.Type,
			nullableStringValue(record.SourceJSON),
			nullableString(record.Name),
			nullableString(record.Description),
			nullableString(record.Source),
		); err != nil {
			return fmt.Errorf("insert magic file_id=%d source_record_id=%s: %w", fileID, record.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertLocations(ctx context.Context, tx *sql.Tx, fileID int64, locations []skyrim.Location) error {
	for _, location := range locations {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_locations (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				name, parent_id, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			location.ID,
			nullableString(location.EditorID),
			location.Type,
			nullableStringValue(location.SourceJSON),
			nullableString(location.Name),
			nullableString(location.ParentID),
			nullableString(location.Source),
		); err != nil {
			return fmt.Errorf("insert location file_id=%d source_record_id=%s: %w", fileID, location.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertCells(ctx context.Context, tx *sql.Tx, fileID int64, cells []skyrim.Location) error {
	for _, cell := range cells {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_cells (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				name, parent_id, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			cell.ID,
			nullableString(cell.EditorID),
			cell.Type,
			nullableStringValue(cell.SourceJSON),
			nullableString(cell.Name),
			nullableString(cell.ParentID),
			nullableString(cell.Source),
		); err != nil {
			return fmt.Errorf("insert cell file_id=%d source_record_id=%s: %w", fileID, cell.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertSystemRecords(ctx context.Context, tx *sql.Tx, fileID int64, records []skyrim.SystemRecord) error {
	for _, record := range records {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_system_records (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				name, description, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			record.ID,
			nullableString(record.EditorID),
			record.Type,
			nullableStringValue(record.SourceJSON),
			nullableString(record.Name),
			nullableString(record.Description),
			nullableString(record.Source),
		); err != nil {
			return fmt.Errorf("insert system record file_id=%d source_record_id=%s: %w", fileID, record.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertMessages(ctx context.Context, tx *sql.Tx, fileID int64, messages []skyrim.Message) error {
	for _, message := range messages {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_messages (
				file_id, source_record_id, editor_id, record_type, source_json_path,
				text, title, quest_id, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			message.ID,
			nullableString(message.EditorID),
			message.Type,
			nullableStringValue(message.SourceJSON),
			message.Text,
			nullableString(message.Title),
			nullableString(message.QuestID),
			nullableString(message.Source),
		); err != nil {
			return fmt.Errorf("insert message file_id=%d source_record_id=%s: %w", fileID, message.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertLoadScreens(ctx context.Context, tx *sql.Tx, fileID int64, loadScreens []skyrim.LoadScreen) error {
	for _, loadScreen := range loadScreens {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_load_screens (
				file_id, source_record_id, editor_id, record_type, source_json_path, text, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			loadScreen.ID,
			nullableString(loadScreen.EditorID),
			loadScreen.Type,
			nullableStringValue(loadScreen.SourceJSON),
			loadScreen.Text,
			nullableString(loadScreen.Source),
		); err != nil {
			return fmt.Errorf("insert load screen file_id=%d source_record_id=%s: %w", fileID, loadScreen.ID, err)
		}
	}
	return nil
}

func (r *sqliteRepository) insertNPCs(ctx context.Context, tx *sql.Tx, fileID int64, npcs map[string]skyrim.NPC) error {
	keys := make([]string, 0, len(npcs))
	for key := range npcs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		npc := npcs[key]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO translation_input_npcs (
				file_id, npc_key, source_record_id, editor_id, record_type, source_json_path,
				name, race, voice, sex, class_name, source
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fileID,
			key,
			npc.ID,
			nullableString(npc.EditorID),
			npc.Type,
			nullableStringValue(npc.SourceJSON),
			nullableStringValue(npc.Name),
			nullableStringValue(npc.Race),
			nullableStringValue(npc.Voice),
			nullableStringValue(npc.Sex),
			nullableString(npc.ClassName),
			nullableString(npc.Source),
		); err != nil {
			return fmt.Errorf("insert npc file_id=%d npc_key=%s: %w", fileID, key, err)
		}
	}
	return nil
}

func (r *sqliteRepository) countPreviewRows(ctx context.Context, querier sqlQuerier, fileID int64) (int, error) {
	query := `SELECT COUNT(1) FROM (` + previewUnionSQL + `)`
	var count int
	if err := querier.QueryRowContext(ctx, query, previewUnionArgs(fileID)...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count preview rows file_id=%d: %w", fileID, err)
	}
	return count, nil
}

type sqlQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func hashFilePath(path string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(path)))
	return hex.EncodeToString(hash[:])
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableStringValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func previewUnionArgs(fileID int64) []any {
	args := make([]any, 0, 16)
	for i := 0; i < 16; i++ {
		args = append(args, fileID)
	}
	return args
}

func terminologyEntriesFromOutput(output *skyrim.ParserOutput, sourceFileName string) []TerminologyEntry {
	entries := make([]TerminologyEntry, 0)
	if output == nil {
		return entries
	}

	for _, item := range output.Items {
		entries = appendTerminologyItemEntries(entries, item.BaseExtractedRecord.ID, item.EditorID, item.Type, item.Name, item.Description, item.Text, sourceFileName)
	}
	for _, location := range output.Locations {
		entries = appendTerminologyTextEntry(entries, location.BaseExtractedRecord.ID, location.EditorID, location.Type, location.Name, sourceFileName)
	}
	for _, cell := range output.Cells {
		entries = appendTerminologyTextEntry(entries, cell.BaseExtractedRecord.ID, cell.EditorID, cell.Type, cell.Name, sourceFileName)
	}
	for _, npc := range output.NPCs {
		entries = appendTerminologyNPCEntry(entries, npc.BaseExtractedRecord.ID, npc.EditorID, npc.Type, npc.Name, sourceFileName)
	}

	return entries
}

func appendTerminologyItemEntries(entries []TerminologyEntry, recordID string, editorID *string, recordType string, name *string, description *string, text *string, sourceFileName string) []TerminologyEntry {
	normalizedRecordType := normalizeTerminologyRecordType(recordType)

	entries = appendTerminologyTextEntry(entries, recordID, editorID, normalizedRecordType, name, sourceFileName)
	entries = appendTerminologyTextEntry(entries, recordID, editorID, normalizedRecordType, description, sourceFileName)
	entries = appendTerminologyTextEntry(entries, recordID, editorID, normalizedRecordType, text, sourceFileName)
	return entries
}

func appendTerminologyNPCEntry(entries []TerminologyEntry, recordID string, editorID *string, recordType string, sourceText string, sourceFileName string) []TerminologyEntry {
	normalizedRecordType := normalizeTerminologyRecordType(recordType)
	return appendTerminologyEntry(entries, recordID, editorID, normalizedRecordType, sourceText, sourceFileName, normalizePairKey(editorID, recordID), "full")
}

func appendTerminologyTextEntry(entries []TerminologyEntry, recordID string, editorID *string, recordType string, sourceText *string, sourceFileName string) []TerminologyEntry {
	if sourceText == nil {
		return entries
	}
	return appendTerminologyEntry(entries, recordID, editorID, recordType, *sourceText, sourceFileName, "", "single")
}

func appendTerminologyEntry(entries []TerminologyEntry, recordID string, editorID *string, recordType string, sourceText string, sourceFileName string, pairKey string, variant string) []TerminologyEntry {
	trimmedText := strings.TrimSpace(sourceText)
	if trimmedText == "" {
		return entries
	}

	trimmedRecordType := normalizeTerminologyRecordType(recordType)
	if !foundation.IsDictionaryImportREC(trimmedRecordType) {
		return entries
	}
	entries = append(entries, TerminologyEntry{
		ID:         recordID,
		EditorID:   trimStringPtr(editorID),
		RecordType: trimmedRecordType,
		SourceText: trimmedText,
		SourceFile: sourceFileName,
		PairKey:    strings.TrimSpace(pairKey),
		Variant:    strings.TrimSpace(variant),
	})
	return entries
}

func normalizeTerminologyRecordType(recordType string) string {
	trimmed := strings.TrimSpace(recordType)
	if trimmed == "" {
		return trimmed
	}
	parts := strings.Fields(trimmed)
	if len(parts) == 2 {
		variant := strings.ToUpper(parts[1])
		if variant == "FULL" || variant == "SHRT" {
			return parts[0] + ":" + variant
		}
	}
	if strings.Contains(trimmed, ":") {
		pair := strings.SplitN(trimmed, ":", 2)
		left := strings.Join(strings.Fields(pair[0]), "")
		right := strings.ToUpper(strings.Join(strings.Fields(pair[1]), ""))
		if right == "" {
			right = "FULL"
		}
		return left + ":" + right
	}
	return strings.Join(parts, "") + ":FULL"
}

func normalizePairKey(editorID *string, recordID string) string {
	if trimmed := trimStringPtr(editorID); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(recordID)
}

func trimStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

var sourcePluginPattern = regexp.MustCompile(`(?i)[^\\/:*?"<>|]+\.(esm|esl|esp)`)

func resolvePersonaSourcePlugin(candidates ...string) string {
	for _, candidate := range candidates {
		match := sourcePluginPattern.FindString(strings.TrimSpace(candidate))
		if match != "" {
			return match
		}
	}
	return ""
}

func resolvePersonaSourceHint(candidates ...string) string {
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

const previewUnionSQL = `
SELECT row_id, section, record_type, editor_id, source_text
FROM (
	SELECT
		printf('dialogue_response:%d', r.id) AS row_id,
		'dialogue_response' AS section,
		COALESCE(r.record_type, '') AS record_type,
		COALESCE(NULLIF(r.editor_id, ''), r.source_record_id, '') AS editor_id,
		r.text AS source_text
	FROM translation_input_dialogue_responses r
	JOIN translation_input_dialogue_groups g ON g.id = r.dialogue_group_id
	WHERE g.file_id = ? AND TRIM(COALESCE(r.text, '')) <> ''

	UNION ALL

	SELECT
		printf('quest_stage:%d', qs.id) AS row_id,
		'quest_stage' AS section,
		COALESCE(q.record_type, '') AS record_type,
		COALESCE(NULLIF(q.editor_id, ''), NULLIF(qs.parent_editor_id, ''), q.source_record_id, '') AS editor_id,
		qs.text AS source_text
	FROM translation_input_quest_stages qs
	JOIN translation_input_quests q ON q.id = qs.quest_id
	WHERE q.file_id = ? AND TRIM(COALESCE(qs.text, '')) <> ''

	UNION ALL

	SELECT
		printf('quest_objective:%d', qo.id) AS row_id,
		'quest_objective' AS section,
		COALESCE(q.record_type, '') AS record_type,
		COALESCE(NULLIF(q.editor_id, ''), NULLIF(qo.parent_editor_id, ''), q.source_record_id, '') AS editor_id,
		qo.text AS source_text
	FROM translation_input_quest_objectives qo
	JOIN translation_input_quests q ON q.id = qo.quest_id
	WHERE q.file_id = ? AND TRIM(COALESCE(qo.text, '')) <> ''

	UNION ALL

	SELECT
		printf('item_name:%d', i.id) AS row_id,
		'item_name' AS section,
		COALESCE(i.record_type, '') AS record_type,
		COALESCE(NULLIF(i.editor_id, ''), i.source_record_id, '') AS editor_id,
		i.name AS source_text
	FROM translation_input_items i
	WHERE i.file_id = ? AND TRIM(COALESCE(i.name, '')) <> ''

	UNION ALL

	SELECT
		printf('item_description:%d', i.id) AS row_id,
		'item_description' AS section,
		COALESCE(i.record_type, '') AS record_type,
		COALESCE(NULLIF(i.editor_id, ''), i.source_record_id, '') AS editor_id,
		i.description AS source_text
	FROM translation_input_items i
	WHERE i.file_id = ? AND TRIM(COALESCE(i.description, '')) <> ''

	UNION ALL

	SELECT
		printf('item_text:%d', i.id) AS row_id,
		'item_text' AS section,
		COALESCE(i.record_type, '') AS record_type,
		COALESCE(NULLIF(i.editor_id, ''), i.source_record_id, '') AS editor_id,
		i.text AS source_text
	FROM translation_input_items i
	WHERE i.file_id = ? AND TRIM(COALESCE(i.text, '')) <> ''

	UNION ALL

	SELECT
		printf('magic_name:%d', m.id) AS row_id,
		'magic_name' AS section,
		COALESCE(m.record_type, '') AS record_type,
		COALESCE(NULLIF(m.editor_id, ''), m.source_record_id, '') AS editor_id,
		m.name AS source_text
	FROM translation_input_magic m
	WHERE m.file_id = ? AND TRIM(COALESCE(m.name, '')) <> ''

	UNION ALL

	SELECT
		printf('magic_description:%d', m.id) AS row_id,
		'magic_description' AS section,
		COALESCE(m.record_type, '') AS record_type,
		COALESCE(NULLIF(m.editor_id, ''), m.source_record_id, '') AS editor_id,
		m.description AS source_text
	FROM translation_input_magic m
	WHERE m.file_id = ? AND TRIM(COALESCE(m.description, '')) <> ''

	UNION ALL

	SELECT
		printf('location_name:%d', l.id) AS row_id,
		'location_name' AS section,
		COALESCE(l.record_type, '') AS record_type,
		COALESCE(NULLIF(l.editor_id, ''), l.source_record_id, '') AS editor_id,
		l.name AS source_text
	FROM translation_input_locations l
	WHERE l.file_id = ? AND TRIM(COALESCE(l.name, '')) <> ''

	UNION ALL

	SELECT
		printf('cell_name:%d', c.id) AS row_id,
		'cell_name' AS section,
		COALESCE(c.record_type, '') AS record_type,
		COALESCE(NULLIF(c.editor_id, ''), c.source_record_id, '') AS editor_id,
		c.name AS source_text
	FROM translation_input_cells c
	WHERE c.file_id = ? AND TRIM(COALESCE(c.name, '')) <> ''

	UNION ALL

	SELECT
		printf('system_name:%d', s.id) AS row_id,
		'system_name' AS section,
		COALESCE(s.record_type, '') AS record_type,
		COALESCE(NULLIF(s.editor_id, ''), s.source_record_id, '') AS editor_id,
		s.name AS source_text
	FROM translation_input_system_records s
	WHERE s.file_id = ? AND TRIM(COALESCE(s.name, '')) <> ''

	UNION ALL

	SELECT
		printf('system_description:%d', s.id) AS row_id,
		'system_description' AS section,
		COALESCE(s.record_type, '') AS record_type,
		COALESCE(NULLIF(s.editor_id, ''), s.source_record_id, '') AS editor_id,
		s.description AS source_text
	FROM translation_input_system_records s
	WHERE s.file_id = ? AND TRIM(COALESCE(s.description, '')) <> ''

	UNION ALL

	SELECT
		printf('message_text:%d', m.id) AS row_id,
		'message_text' AS section,
		COALESCE(m.record_type, '') AS record_type,
		COALESCE(NULLIF(m.editor_id, ''), m.source_record_id, '') AS editor_id,
		m.text AS source_text
	FROM translation_input_messages m
	WHERE m.file_id = ? AND TRIM(COALESCE(m.text, '')) <> ''

	UNION ALL

	SELECT
		printf('message_title:%d', m.id) AS row_id,
		'message_title' AS section,
		COALESCE(m.record_type, '') AS record_type,
		COALESCE(NULLIF(m.editor_id, ''), m.source_record_id, '') AS editor_id,
		m.title AS source_text
	FROM translation_input_messages m
	WHERE m.file_id = ? AND TRIM(COALESCE(m.title, '')) <> ''

	UNION ALL

	SELECT
		printf('load_screen_text:%d', ls.id) AS row_id,
		'load_screen_text' AS section,
		COALESCE(ls.record_type, '') AS record_type,
		COALESCE(NULLIF(ls.editor_id, ''), ls.source_record_id, '') AS editor_id,
		ls.text AS source_text
	FROM translation_input_load_screens ls
	WHERE ls.file_id = ? AND TRIM(COALESCE(ls.text, '')) <> ''

	UNION ALL

	SELECT
		printf('npc_name:%d', n.id) AS row_id,
		'npc_name' AS section,
		COALESCE(n.record_type, '') AS record_type,
		COALESCE(NULLIF(n.editor_id, ''), n.source_record_id, '') AS editor_id,
		n.name AS source_text
	FROM translation_input_npcs n
	WHERE n.file_id = ? AND TRIM(COALESCE(n.name, '')) <> ''
) preview_rows
`
