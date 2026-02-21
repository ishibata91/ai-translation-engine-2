package persona_gen

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
)

// NPCPersonaGenerator is the main entry point for NPC persona generation.
// It orchestrates dialogue collection, token estimation, LLM persona generation,
// and persistence to the persona database.
type NPCPersonaGenerator interface {
	GeneratePersonas(ctx context.Context, data models.ExtractedData) ([]PersonaResult, error)
}

// DialogueCollector collects per-NPC dialogue data from ExtractedData,
// applying importance scoring to select the top dialogues.
type DialogueCollector interface {
	CollectByNPC(ctx context.Context, data models.ExtractedData) ([]NPCDialogueData, error)
}

// ImportanceScorer scores dialogue pairs by importance for persona generation.
type ImportanceScorer interface {
	Score(englishText string, questID *string, isServicesBranch bool) int
}

// TokenEstimator estimates the token count of a given text.
type TokenEstimator interface {
	Estimate(text string) int
}

// ContextEvaluator evaluates token usage and trims dialogues to fit the context window.
type ContextEvaluator interface {
	Evaluate(dialogueData NPCDialogueData, config PersonaConfig) (TokenEstimation, []DialogueEntry)
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
