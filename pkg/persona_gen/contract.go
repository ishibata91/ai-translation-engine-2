package persona_gen

import (
	"context"
)

// PersonaGenInput is the input data required for persona generation.
type PersonaGenInput struct {
	NPCs      map[string]PersonaNPC
	Dialogues []PersonaDialogue
}

type PersonaNPC struct {
	ID       string
	EditorID *string
	Type     string
	Name     string
}

type PersonaDialogue struct {
	ID               string
	EditorID         *string
	Type             string
	SpeakerID        *string
	Text             *string
	QuestID          *string
	IsServicesBranch bool
	Order            int
}

// NPCPersonaGenerator is the main entry point for NPC persona generation.
// It orchestrates dialogue collection, token estimation, LLM persona generation,
// and persistence to the persona database.
type NPCPersonaGenerator interface {
	GeneratePersonas(ctx context.Context, data PersonaGenInput) ([]PersonaResult, error)
}

// DialogueCollector collects per-NPC dialogue data from PersonaGenInput,
// applying importance scoring to select the top dialogues.
type DialogueCollector interface {
	CollectByNPC(ctx context.Context, data PersonaGenInput) ([]NPCDialogueData, error)
}

// ImportanceScorer scores dialogue pairs by importance for persona generation.
type ImportanceScorer interface {
	Score(ctx context.Context, englishText string, questID *string, isServicesBranch bool) int
}

// TokenEstimator estimates the token count of a given text.
type TokenEstimator interface {
	Estimate(ctx context.Context, text string) int
}

// ContextEvaluator evaluates token usage and trims dialogues to fit the context window.
type ContextEvaluator interface {
	Evaluate(ctx context.Context, dialogueData NPCDialogueData, config PersonaConfig) (TokenEstimation, []DialogueEntry)
}

// PersonaStore manages all operations on the npc_personas SQLite table,
// including schema creation, INSERT/UPSERT, and retrieval.
type PersonaStore interface {
	InitSchema(ctx context.Context) error
	SavePersona(ctx context.Context, result PersonaResult) error
	GetPersona(ctx context.Context, speakerID string) (string, error)
	Clear(ctx context.Context) error
}

// ProgressNotifier reports persona generation progress to the Process Manager.
type ProgressNotifier interface {
	OnProgress(completed int, total int)
}
