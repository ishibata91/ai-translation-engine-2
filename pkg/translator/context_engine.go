package translator

import (
	"context"
)

// ContextEngineInput is the input data required for building translation context.
type ContextEngineInput struct {
	NPCs      map[string]ContextNPC
	Dialogues []ContextDialogue
	Quests    []ContextQuest
	Items     []ContextItem
	Magic     []ContextMagic
	Locations []ContextLocation
}

type ContextNPC struct {
	ID       string
	EditorID *string
	Type     string
	Name     string
}

type ContextDialogue struct {
	ID               string
	EditorID         *string
	Type             string
	SpeakerID        *string
	Text             *string
	QuestID          *string
	IsServicesBranch bool
	Order            int
}

type ContextQuest struct {
	ID         string
	EditorID   *string
	Type       string
	Name       *string
	Stages     []ContextQuestStage
	Objectives []ContextQuestObjective
}

type ContextQuestStage struct {
	StageIndex     int
	LogIndex       int
	Type           string
	Text           string
	ParentID       string
	ParentEditorID string
}

type ContextQuestObjective struct {
	Index          string
	Type           string
	Text           string
	ParentID       string
	ParentEditorID string
}

type ContextItem struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
	Text     *string
}

type ContextMagic struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
}

type ContextLocation struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
}

// ContextEngine (Internal) handles context building logic moved from Lore slice.
// It integrates dialogue tree analysis, speaker profiling, reference term lookup,
// and summary lookup.
type ContextEngine interface {
	BuildTranslationContext(ctx context.Context, record interface{}, input *ContextEngineInput) (*Pass2Context, []Pass2ReferenceTerm, *string, error)
}

// ToneResolver generates a tone instruction string from NPC attributes.
type ToneResolver interface {
	Resolve(race string, voiceType string, sex string) string
}

// PersonaLookup searches for an NPC's persona text by speaker ID.
type PersonaLookup interface {
	FindBySpeakerID(ctx context.Context, speakerID string) (*string, error)
}

// TermLookup searches dictionary and Mod term databases for reference terms.
type TermLookup interface {
	Search(ctx context.Context, sourceText string) ([]Pass2ReferenceTerm, *string, error)
}

// SummaryLookup retrieves cached summaries.
type SummaryLookup interface {
	FindDialogueSummary(ctx context.Context, dialogueGroupID string) (*string, error)
	FindQuestSummary(ctx context.Context, questID string) (*string, error)
}
