package terminology

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
)

// TerminologyInput is the input data required for term translation.
type TerminologyInput struct {
	NPCs      map[string]TermNPC
	Items     []TermItem
	Magic     []TermMagic
	Locations []TermLocation
	Messages  []TermMessage
	Quests    []TermQuest
}

type TermNPC struct {
	ID         string
	EditorID   *string
	Type       string
	Name       string
	SourceFile string
}

type TermItem struct {
	ID         string
	EditorID   *string
	Type       string
	Name       *string
	Text       *string
	SourceFile string
}

type TermMagic struct {
	ID         string
	EditorID   *string
	Type       string
	Name       *string
	SourceFile string
}

type TermLocation struct {
	ID         string
	EditorID   *string
	Type       string
	Name       *string
	SourceFile string
}

type TermMessage struct {
	ID         string
	EditorID   *string
	Type       string
	Title      *string
	SourceFile string
}

type TermQuest struct {
	ID         string
	EditorID   *string
	Type       string
	Name       *string
	SourceFile string
}

// RequestConfig stores runtime request settings passed from workflow/UI.
type RequestConfig struct {
	Provider        string
	Model           string
	Endpoint        string
	APIKey          string
	Temperature     float32
	ContextLength   int
	SyncConcurrency int
	BulkStrategy    string
}

// PromptConfig stores runtime prompt settings passed from workflow/UI.
type PromptConfig struct {
	UserPrompt   string
	SystemPrompt string
}

// PhaseOptions contains the DTO boundary passed from workflow.
type PhaseOptions struct {
	Request RequestConfig
	Prompt  PromptConfig
}

// PhaseSummary reports the persisted terminology phase state.
type PhaseSummary struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	TargetCount int    `json:"target_count"`
	SavedCount  int    `json:"saved_count"`
	FailedCount int    `json:"failed_count"`
}

// Terminology is the main entry point for term translation (Pass 1).
type Terminology interface {
	// ID returns the unique identifier of the slice.
	ID() string

	// PreparePrompts (Phase 1) generates LLM requests from task-scoped artifact input.
	PreparePrompts(ctx context.Context, taskID string, options PhaseOptions) ([]llmio.Request, error)

	// SaveResults (Phase 2) persists LLM responses for one task.
	SaveResults(ctx context.Context, taskID string, responses []llmio.Response) error

	// GetPhaseSummary returns persisted counts/status for one task.
	GetPhaseSummary(ctx context.Context, taskID string) (PhaseSummary, error)
}

// TermRequestBuilder extracts term translation targets from TerminologyInput
// and constructs translation requests, including NPC FULL+SHRT pairing.
type TermRequestBuilder interface {
	BuildRequests(ctx context.Context, data TerminologyInput) ([]TermTranslationRequest, error)
}

// TermDictionarySearcher searches the dictionary DB (built by Dictionary Builder Slice)
// for reference terms to provide as LLM context.
type TermDictionarySearcher interface {
	SearchExact(ctx context.Context, text string) ([]ReferenceTerm, error)
	SearchKeywords(ctx context.Context, keywords []string) ([]ReferenceTerm, error)
	SearchNPCPartial(ctx context.Context, keywords []string, consumedKeywords []string, isNPC bool) ([]ReferenceTerm, error)
	SearchBatch(ctx context.Context, texts []string) (map[string][]ReferenceTerm, error)
}

// ModTermStore manages all operations on the Mod term SQLite database,
// including schema creation, INSERT/UPSERT, and FTS5 index management.
type ModTermStore interface {
	InitSchema(ctx context.Context) error
	SaveTerms(ctx context.Context, results []TermTranslationResult) error
	GetTerm(ctx context.Context, originalEN string) (string, error)
	Clear(ctx context.Context) error
	UpdatePhaseSummary(ctx context.Context, summary PhaseSummary) error
	GetPhaseSummary(ctx context.Context, taskID string) (PhaseSummary, error)
}

// ProgressNotifier reports translation progress to the Process Manager.
type ProgressNotifier interface {
	OnProgress(completed int, total int)
}

// TranslationInputRepository loads terminology targets from shared artifact storage.
type TranslationInputRepository interface {
	LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error)
}
