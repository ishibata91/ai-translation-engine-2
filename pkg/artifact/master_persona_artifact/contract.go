package masterpersonaartifact

import (
	"context"
	"time"
)

// LookupKey identifies one final/master-persona record by source plugin and speaker.
type LookupKey struct {
	SourcePlugin string
	SpeakerID    string
}

// Dialogue represents one dialogue line persisted in final or temp artifacts.
type Dialogue struct {
	RecordType       string `json:"record_type"`
	EditorID         string `json:"editor_id"`
	SourceText       string `json:"source_text"`
	QuestID          string `json:"quest_id,omitempty"`
	IsServicesBranch bool   `json:"is_services_branch"`
	Order            int    `json:"order"`
}

// FinalPersona represents one generated persona final artifact.
type FinalPersona struct {
	PersonaID         int64
	FormID            string
	SourcePlugin      string
	SpeakerID         string
	NPCName           string
	EditorID          string
	Race              string
	Sex               string
	VoiceType         string
	UpdatedAt         time.Time
	PersonaText       string
	GenerationRequest string
	Dialogues         []Dialogue
}

// TempPersona represents one task-scoped intermediate artifact.
type TempPersona struct {
	ID                int64
	TaskID            string
	SourcePlugin      string
	SpeakerID         string
	EditorID          string
	NPCName           string
	Race              string
	Sex               string
	VoiceType         string
	GenerationRequest string
	Dialogues         []Dialogue
	UpdatedAt         time.Time
}

// Repository defines final/temp artifact operations for master persona.
type Repository interface {
	SaveOrUpdateTempBase(ctx context.Context, temp TempPersona, overwriteExisting bool) (int64, string, error)
	SaveTempGenerationRequest(ctx context.Context, taskID string, key LookupKey, generationRequest string) error
	ReplaceTempDialogues(ctx context.Context, taskID string, key LookupKey, dialogues []Dialogue) error
	ReplaceTempDialoguesByID(ctx context.Context, tempID int64, dialogues []Dialogue) error
	SaveOrUpdateFinal(ctx context.Context, taskID string, persona FinalPersona, overwriteExisting bool) error
	GetFinalPersonaText(ctx context.Context, key LookupKey) (string, error)
	ListFinalPersonas(ctx context.Context) ([]FinalPersona, error)
	GetFinalByPersonaID(ctx context.Context, personaID int64) (FinalPersona, error)
	FindFinalByLookup(ctx context.Context, key LookupKey) (FinalPersona, error)
	CleanupTaskTemp(ctx context.Context, taskID string) error
	ClearAll(ctx context.Context) error
}
