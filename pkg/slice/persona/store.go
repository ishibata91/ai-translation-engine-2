package persona

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	master_persona_artifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/master_persona_artifact"
)

var pluginNamePattern = regexp.MustCompile(`(?i)[^\\/:*?"<>|]+\.(esm|esl|esp)`)

type artifactPersonaStore struct {
	repo master_persona_artifact.Repository
}

// NewPersonaStore creates a PersonaStore backed by master_persona_artifact repository.
func NewPersonaStore(repo master_persona_artifact.Repository) PersonaStore {
	return &artifactPersonaStore{repo: repo}
}

func (s *artifactPersonaStore) InitSchema(ctx context.Context) error {
	// Schema migration is owned by artifact repository migration at composition root.
	slog.DebugContext(ctx, "persona store init schema is no-op for artifact-backed store")
	return nil
}

func (s *artifactPersonaStore) SavePersona(ctx context.Context, result PersonaResult, overwriteExisting bool) error {
	taskID := taskIDFromContext(ctx)
	if err := s.repo.SaveOrUpdateFinal(ctx, taskID, master_persona_artifact.FinalPersona{
		FormID:       result.SpeakerID,
		SourcePlugin: normalizeSourcePlugin(result.SourcePlugin, ""),
		SpeakerID:    strings.TrimSpace(result.SpeakerID),
		NPCName:      result.NPCName,
		EditorID:     result.EditorID,
		Race:         result.Race,
		Sex:          result.Sex,
		VoiceType:    result.VoiceType,
		PersonaText:  result.PersonaText,
	}, overwriteExisting); err != nil {
		return fmt.Errorf("save persona final artifact speaker_id=%s: %w", result.SpeakerID, err)
	}
	return nil
}

func (s *artifactPersonaStore) SavePersonaBase(ctx context.Context, data NPCDialogueData, overwriteExisting bool) (PersonaSaveState, error) {
	taskID := taskIDFromContext(ctx)
	if taskID == "" {
		taskID = "adhoc"
	}
	sourcePlugin := normalizeSourcePlugin(data.SourcePlugin, data.SourceHint)
	tempID, existingPersonaText, err := s.repo.SaveOrUpdateTempBase(ctx, master_persona_artifact.TempPersona{
		TaskID:       taskID,
		SourcePlugin: sourcePlugin,
		SpeakerID:    strings.TrimSpace(data.SpeakerID),
		EditorID:     data.EditorID,
		NPCName:      data.NPCName,
		Race:         data.Race,
		Sex:          data.Sex,
		VoiceType:    data.VoiceType,
	}, overwriteExisting)
	if err != nil {
		return PersonaSaveState{}, fmt.Errorf("save persona temp base task_id=%s speaker_id=%s: %w", taskID, data.SpeakerID, err)
	}
	return PersonaSaveState{
		PersonaID:   tempID,
		PersonaText: existingPersonaText,
	}, nil
}

func (s *artifactPersonaStore) SaveGenerationRequest(ctx context.Context, sourcePlugin string, speakerID string, generationRequest string) error {
	taskID := taskIDFromContext(ctx)
	if taskID == "" {
		taskID = "adhoc"
	}
	if err := s.repo.SaveTempGenerationRequest(ctx, taskID, master_persona_artifact.LookupKey{
		SourcePlugin: normalizeSourcePlugin(sourcePlugin, ""),
		SpeakerID:    strings.TrimSpace(speakerID),
	}, generationRequest); err != nil {
		return fmt.Errorf("save generation request task_id=%s speaker_id=%s: %w", taskID, speakerID, err)
	}
	return nil
}

func (s *artifactPersonaStore) ReplaceDialogues(ctx context.Context, personaID int64, sourcePlugin string, speakerID string, dialogues []DialogueEntry) error {
	_ = personaID
	taskID := taskIDFromContext(ctx)
	if taskID == "" {
		taskID = "adhoc"
	}
	mapped := make([]master_persona_artifact.Dialogue, 0, len(dialogues))
	for _, dialogue := range dialogues {
		sourceText := strings.TrimSpace(dialogue.EnglishText)
		if sourceText == "" {
			sourceText = strings.TrimSpace(dialogue.Text)
		}
		if sourceText == "" {
			continue
		}
		mapped = append(mapped, master_persona_artifact.Dialogue{
			RecordType:       dialogue.RecordType,
			EditorID:         dialogue.EditorID,
			SourceText:       sourceText,
			QuestID:          dialogue.QuestID,
			IsServicesBranch: dialogue.IsServicesBranch,
			Order:            dialogue.Order,
		})
	}
	var err error
	if strings.TrimSpace(sourcePlugin) == "" && personaID > 0 {
		err = s.repo.ReplaceTempDialoguesByID(ctx, personaID, mapped)
	} else {
		err = s.repo.ReplaceTempDialogues(ctx, taskID, master_persona_artifact.LookupKey{
			SourcePlugin: normalizeSourcePlugin(sourcePlugin, ""),
			SpeakerID:    strings.TrimSpace(speakerID),
		}, mapped)
	}
	if err != nil {
		return fmt.Errorf("replace temp dialogues task_id=%s speaker_id=%s: %w", taskID, speakerID, err)
	}
	return nil
}

func (s *artifactPersonaStore) GetPersona(ctx context.Context, sourcePlugin string, speakerID string) (string, error) {
	personaText, err := s.repo.GetFinalPersonaText(ctx, master_persona_artifact.LookupKey{
		SourcePlugin: normalizeSourcePlugin(sourcePlugin, ""),
		SpeakerID:    strings.TrimSpace(speakerID),
	})
	if err != nil {
		return "", fmt.Errorf("get persona text source_plugin=%s speaker_id=%s: %w", sourcePlugin, speakerID, err)
	}
	return personaText, nil
}

func (s *artifactPersonaStore) ListNPCs(ctx context.Context) ([]PersonaNPCView, error) {
	finalRows, err := s.repo.ListFinalPersonas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list final personas: %w", err)
	}
	result := make([]PersonaNPCView, 0, len(finalRows))
	for _, row := range finalRows {
		result = append(result, PersonaNPCView{
			PersonaID:         row.PersonaID,
			SpeakerID:         row.FormID,
			EditorID:          row.EditorID,
			NPCName:           row.NPCName,
			Race:              row.Race,
			Sex:               row.Sex,
			VoiceType:         row.VoiceType,
			PersonaText:       row.PersonaText,
			GenerationRequest: row.GenerationRequest,
			SourcePlugin:      row.SourcePlugin,
			UpdatedAt:         row.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	return result, nil
}

func (s *artifactPersonaStore) ListDialoguesByPersonaID(ctx context.Context, personaID int64) ([]PersonaDialogueView, error) {
	if personaID <= 0 {
		return nil, errors.New("persona_id is required")
	}
	finalRow, err := s.repo.GetFinalByPersonaID(ctx, personaID)
	if err != nil {
		return nil, fmt.Errorf("get final persona dialogues persona_id=%d: %w", personaID, err)
	}
	result := make([]PersonaDialogueView, 0, len(finalRow.Dialogues))
	for idx, dialogue := range finalRow.Dialogues {
		result = append(result, PersonaDialogueView{
			ID:               int64(idx + 1),
			PersonaID:        finalRow.PersonaID,
			SpeakerID:        finalRow.SpeakerID,
			SourcePlugin:     finalRow.SourcePlugin,
			EditorID:         dialogue.EditorID,
			RecordType:       dialogue.RecordType,
			SourceText:       dialogue.SourceText,
			QuestID:          dialogue.QuestID,
			IsServicesBranch: dialogue.IsServicesBranch,
			DialogueOrder:    dialogue.Order,
		})
	}
	return result, nil
}

func (s *artifactPersonaStore) CleanupTaskArtifacts(ctx context.Context, taskID string) error {
	if err := s.repo.CleanupTaskTemp(ctx, taskID); err != nil {
		return fmt.Errorf("cleanup persona temp artifacts task_id=%s: %w", taskID, err)
	}
	return nil
}

func (s *artifactPersonaStore) Clear(ctx context.Context) error {
	if err := s.repo.ClearAll(ctx); err != nil {
		return fmt.Errorf("clear persona artifacts: %w", err)
	}
	return nil
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
