package translator

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/summary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/terminology"
)

// personaLookupAdapter adapts persona.PersonaStore to PersonaLookup interface.
type personaLookupAdapter struct {
	store persona.PersonaStore
}

func NewPersonaLookupAdapter(store persona.PersonaStore) PersonaLookup {
	return &personaLookupAdapter{store: store}
}

func (a *personaLookupAdapter) FindBySpeakerID(ctx context.Context, speakerID string) (*string, error) {
	text, err := a.store.GetPersona(ctx, speakerID)
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, nil
	}
	return &text, nil
}

// summaryLookupAdapter adapts summary.SummaryStore to SummaryLookup interface.
type summaryLookupAdapter struct {
	store summary.SummaryStore
}

func NewSummaryLookupAdapter(store summary.SummaryStore) SummaryLookup {
	return &summaryLookupAdapter{store: store}
}

func (a *summaryLookupAdapter) FindDialogueSummary(ctx context.Context, dialogueGroupID string) (*string, error) {
	s, err := a.store.GetByRecordID(ctx, dialogueGroupID, "dialogue")
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return &s.SummaryText, nil
}

func (a *summaryLookupAdapter) FindQuestSummary(ctx context.Context, questID string) (*string, error) {
	s, err := a.store.GetByRecordID(ctx, questID, "quest")
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return &s.SummaryText, nil
}

// termLookupAdapter adapts terminology.ModTermStore to TermLookup interface.
type termLookupAdapter struct {
	store terminology.ModTermStore
}

func NewTermLookupAdapter(store terminology.ModTermStore) TermLookup {
	return &termLookupAdapter{store: store}
}

func (a *termLookupAdapter) Search(ctx context.Context, sourceText string) ([]Pass2ReferenceTerm, *string, error) {
	// Simple implementation: check if the whole text is a term (e.g. for Item Names)
	// For actual dialogue, we would need a more complex scanner (Pass 1 Terminology logic).
	// For now, let's satisfy the critical requirement with a store-based lookup.
	translated, err := a.store.GetTerm(ctx, sourceText)
	if err == nil && translated != "" {
		return []Pass2ReferenceTerm{{OriginalEN: sourceText, OriginalJA: translated}}, &translated, nil
	}
	return nil, nil, nil
}

// defaultToneResolver implements ToneResolver with basic rules.
type defaultToneResolver struct{}

func NewDefaultToneResolver() ToneResolver {
	return &defaultToneResolver{}
}

func (r *defaultToneResolver) Resolve(race string, voiceType string, sex string) string {
	// Basic tone rules
	tone := "丁寧な話し方"
	if sex == "Male" {
		tone = "標準的な男性の話し方"
	} else if sex == "Female" {
		tone = "標準的な女性の話し方"
	}

	if race == "Nord" {
		tone += " (力強く)"
	}
	// ... more rules can be added
	return tone
}
