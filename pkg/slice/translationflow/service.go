package translationflow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	masterpersonaartifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/master_persona_artifact"
	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
)

type service struct {
	inputRepo         translationinput.Repository
	masterPersonaRepo masterpersonaartifact.Repository
}

// NewService creates a translation-flow slice service backed by artifact repository.
func NewService(inputRepo translationinput.Repository, masterPersonaRepo masterpersonaartifact.Repository) Service {
	return &service{
		inputRepo:         inputRepo,
		masterPersonaRepo: masterPersonaRepo,
	}
}

// EnsureTask creates or updates task parent row in artifact storage.
func (s *service) EnsureTask(ctx context.Context, taskID string) error {
	if err := s.inputRepo.EnsureTask(ctx, taskID); err != nil {
		return fmt.Errorf("ensure translation-flow task task_id=%s: %w", taskID, err)
	}
	return nil
}

// SaveParsedOutput persists one parser output under the task/file boundary.
func (s *service) SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (LoadedFile, error) {
	stored, err := s.inputRepo.SaveParsedOutput(ctx, taskID, sourceFilePath, output)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("save parsed translation input task_id=%s file=%s: %w", taskID, sourceFilePath, err)
	}
	return LoadedFile{
		ID:              stored.ID,
		TaskID:          stored.TaskID,
		SourceFilePath:  stored.SourceFilePath,
		SourceFileName:  stored.SourceFileName,
		SourceFileHash:  stored.SourceFileHash,
		ParseStatus:     stored.ParseStatus,
		PreviewRowCount: stored.PreviewRowCount,
	}, nil
}

// ListFiles loads all saved files for one translation-flow task.
func (s *service) ListFiles(ctx context.Context, taskID string) ([]LoadedFile, error) {
	storedFiles, err := s.inputRepo.ListFiles(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list parsed files task_id=%s: %w", taskID, err)
	}
	files := make([]LoadedFile, 0, len(storedFiles))
	for _, stored := range storedFiles {
		files = append(files, LoadedFile{
			ID:              stored.ID,
			TaskID:          stored.TaskID,
			SourceFilePath:  stored.SourceFilePath,
			SourceFileName:  stored.SourceFileName,
			SourceFileHash:  stored.SourceFileHash,
			ParseStatus:     stored.ParseStatus,
			PreviewRowCount: stored.PreviewRowCount,
		})
	}
	return files, nil
}

// ListPreviewRows loads paginated preview rows for one file.
func (s *service) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error) {
	preview, err := s.inputRepo.ListPreviewRows(ctx, fileID, page, pageSize)
	if err != nil {
		return PreviewPage{}, fmt.Errorf("list preview rows file_id=%d page=%d size=%d: %w", fileID, page, pageSize, err)
	}
	rows := make([]PreviewRow, 0, len(preview.Rows))
	for _, row := range preview.Rows {
		rows = append(rows, PreviewRow{
			ID:         row.ID,
			Section:    row.Section,
			RecordType: row.RecordType,
			EditorID:   row.EditorID,
			SourceText: row.SourceText,
		})
	}
	return PreviewPage{
		FileID:    preview.FileID,
		Page:      preview.Page,
		PageSize:  preview.PageSize,
		TotalRows: preview.TotalRows,
		Rows:      rows,
	}, nil
}

// LoadTerminologyInput exposes persisted terminology rows for workflow preview projection.
func (s *service) LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error) {
	input, err := s.inputRepo.LoadTerminologyInput(ctx, taskID)
	if err != nil {
		return translationinput.TerminologyInput{}, fmt.Errorf("load terminology input task_id=%s: %w", taskID, err)
	}
	return input, nil
}

// LoadPersonaCandidates exposes persisted persona candidates via translation-flow local contract.
func (s *service) LoadPersonaCandidates(ctx context.Context, taskID string) (PersonaCandidateInput, error) {
	input, err := s.inputRepo.LoadPersonaInput(ctx, taskID)
	if err != nil {
		return PersonaCandidateInput{}, fmt.Errorf("load persona input task_id=%s: %w", taskID, err)
	}
	candidates := make(map[string]PersonaCandidate, len(input.NPCs))
	for key, candidate := range input.NPCs {
		candidates[key] = PersonaCandidate{
			SpeakerID:      candidate.SpeakerID,
			SourceRecordID: candidate.SourceRecordID,
			NPCKey:         candidate.NPCKey,
			EditorID:       candidate.EditorID,
			RecordType:     candidate.RecordType,
			NPCName:        candidate.NPCName,
			Race:           candidate.Race,
			Sex:            candidate.Sex,
			VoiceType:      candidate.VoiceType,
			SourcePlugin:   candidate.SourcePlugin,
			SourceHint:     candidate.SourceHint,
		}
	}
	dialogues := make([]PersonaDialogueExcerpt, 0, len(input.Dialogues))
	for _, dialogue := range input.Dialogues {
		dialogues = append(dialogues, PersonaDialogueExcerpt{
			ID:               dialogue.ID,
			SpeakerID:        dialogue.SpeakerID,
			EditorID:         dialogue.EditorID,
			GroupEditorID:    dialogue.GroupEditorID,
			RecordType:       dialogue.RecordType,
			Text:             dialogue.Text,
			QuestID:          dialogue.QuestID,
			SourcePlugin:     dialogue.SourcePlugin,
			SourceHint:       dialogue.SourceHint,
			IsServicesBranch: dialogue.IsServicesBranch,
			Order:            dialogue.Order,
		})
	}
	return PersonaCandidateInput{
		TaskID:     input.TaskID,
		Candidates: candidates,
		Dialogues:  dialogues,
	}, nil
}

// FindPersonaFinal exposes final master persona lookup via translation-flow local contract.
func (s *service) FindPersonaFinal(ctx context.Context, key PersonaLookupKey) (PersonaFinalSummary, bool, error) {
	lookupKey := masterpersonaartifact.LookupKey{
		SourcePlugin: key.SourcePlugin,
		SpeakerID:    key.SpeakerID,
	}
	persona, err := s.masterPersonaRepo.FindFinalByLookup(ctx, lookupKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PersonaFinalSummary{}, false, nil
		}
		return PersonaFinalSummary{}, false, fmt.Errorf("find final persona by lookup source_plugin=%s speaker_id=%s: %w", key.SourcePlugin, key.SpeakerID, err)
	}
	return PersonaFinalSummary{
		PersonaID:    persona.PersonaID,
		SourcePlugin: persona.SourcePlugin,
		SpeakerID:    persona.SpeakerID,
		PersonaText:  persona.PersonaText,
	}, true, nil
}
