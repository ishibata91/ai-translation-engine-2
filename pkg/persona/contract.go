package persona

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

// PersonaGenInput is the input data required for persona generation.
type PersonaGenInput struct {
	NPCs      map[string]PersonaNPC
	Dialogues []PersonaDialogue
}

type PersonaNPC struct {
	ID           string
	EditorID     *string
	Type         string
	Name         string
	Race         string
	Sex          string
	VoiceType    string
	SourcePlugin string
}

type PersonaDialogue struct {
	ID               string
	EditorID         *string
	GroupEditorID    *string
	Type             string
	SpeakerID        *string
	Text             *string
	QuestID          *string
	SourcePlugin     *string
	IsServicesBranch bool
	Order            int
}

// NPCPersonaGenerator is the main entry point for NPC persona generation.
type NPCPersonaGenerator interface {
	// ID returns the unique identifier of the slice.
	ID() string

	// PreparePrompts (Phase 1) generates LLM requests.
	PreparePrompts(ctx context.Context, input any) ([]llm.Request, error)

	// SaveResults (Phase 2) persists LLM responses.
	SaveResults(ctx context.Context, responses []llm.Response) error
}

// SaveResultsSummary reports phase-2 persistence outcomes.
type SaveResultsSummary struct {
	SuccessCount int `json:"success_count"`
	FailCount    int `json:"fail_count"`
}

// SaveResultsReporter optionally exposes detailed save summary for orchestration.
type SaveResultsReporter interface {
	SaveResultsWithSummary(ctx context.Context, responses []llm.Response) (SaveResultsSummary, error)
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
	SavePersonaBase(ctx context.Context, data NPCDialogueData) error
	ReplaceDialogues(ctx context.Context, speakerID string, dialogues []DialogueEntry) error
	GetPersona(ctx context.Context, speakerID string) (string, error)
	ListNPCs(ctx context.Context) ([]PersonaNPCView, error)
	ListDialoguesBySpeaker(ctx context.Context, speakerID string) ([]PersonaDialogueView, error)
	Clear(ctx context.Context) error
}

// ProgressNotifier reports persona generation progress to the Process Manager.
type ProgressNotifier interface {
	OnProgress(completed int, total int)
}
