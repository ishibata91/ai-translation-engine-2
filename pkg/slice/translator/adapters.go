package translator

import "context"

// NewPersonaLookupAdapter returns a boundary-safe default implementation.
func NewPersonaLookupAdapter() PersonaLookup {
	return &noopPersonaLookup{}
}

type noopPersonaLookup struct{}

func (a *noopPersonaLookup) FindBySpeakerID(ctx context.Context, speakerID string) (*string, error) {
	return nil, nil
}

// NewSummaryLookupAdapter returns a boundary-safe default implementation.
func NewSummaryLookupAdapter() SummaryLookup {
	return &noopSummaryLookup{}
}

type noopSummaryLookup struct{}

func (a *noopSummaryLookup) FindDialogueSummary(ctx context.Context, dialogueGroupID string) (*string, error) {
	return nil, nil
}

func (a *noopSummaryLookup) FindQuestSummary(ctx context.Context, questID string) (*string, error) {
	return nil, nil
}

// NewTermLookupAdapter returns a boundary-safe default implementation.
func NewTermLookupAdapter() TermLookup {
	return &noopTermLookup{}
}

type noopTermLookup struct{}

func (a *noopTermLookup) Search(ctx context.Context, sourceText string) ([]Pass2ReferenceTerm, *string, error) {
	return nil, nil, nil
}

// defaultToneResolver implements ToneResolver with basic rules.
type defaultToneResolver struct{}

func NewDefaultToneResolver() ToneResolver {
	return &defaultToneResolver{}
}

func (r *defaultToneResolver) Resolve(race string, voiceType string, sex string) string {
	tone := "丁寧な話し方"
	if sex == "Male" {
		tone = "標準的な男性の話し方"
	} else if sex == "Female" {
		tone = "標準的な女性の話し方"
	}

	if race == "Nord" {
		tone += " (力強く)"
	}
	return tone
}
