package term_translator

import (
	"context"
)

// TermTranslatorInput is the input data required for term translation.
type TermTranslatorInput struct {
	NPCs      map[string]TermNPC
	Items     []TermItem
	Magic     []TermMagic
	Locations []TermLocation
}

type TermNPC struct {
	ID       string
	EditorID *string
	Type     string
	Name     string
}

type TermItem struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
	Text     *string
}

type TermMagic struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
}

type TermLocation struct {
	ID       string
	EditorID *string
	Type     string
	Name     *string
}

// TermTranslator is the main entry point for term translation (Pass 1).
// It orchestrates request building, dictionary search, LLM translation,
// and persistence to the Mod term database.
type TermTranslator interface {
	TranslateTerms(ctx context.Context, data TermTranslatorInput) ([]TermTranslationResult, error)
}

// TermRequestBuilder extracts term translation targets from TermTranslatorInput
// and constructs translation requests, including NPC FULL+SHRT pairing.
type TermRequestBuilder interface {
	BuildRequests(ctx context.Context, data TermTranslatorInput) ([]TermTranslationRequest, error)
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
}

// ProgressNotifier reports translation progress to the Process Manager.
type ProgressNotifier interface {
	OnProgress(completed int, total int)
}
