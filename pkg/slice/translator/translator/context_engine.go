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
	ID        string
	EditorID  *string
	Type      string
	Name      string
	Race      string
	Gender    string
	VoiceType string
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

// contextEngine implements ContextEngine interface.
type contextEngine struct {
	toneResolver  ToneResolver
	personaLookup PersonaLookup
	termLookup    TermLookup
	summaryLookup SummaryLookup
}

// NewContextEngine creates a new ContextEngine instance.
func NewContextEngine(
	tr ToneResolver,
	pl PersonaLookup,
	tl TermLookup,
	sl SummaryLookup,
) ContextEngine {
	return &contextEngine{
		toneResolver:  tr,
		personaLookup: pl,
		termLookup:    tl,
		summaryLookup: sl,
	}
}

func (e *contextEngine) BuildTranslationContext(ctx context.Context, record interface{}, input *ContextEngineInput) (*Pass2Context, []Pass2ReferenceTerm, *string, error) {
	pass2Ctx := &Pass2Context{}
	var terms []Pass2ReferenceTerm
	var forcedTranslation *string

	switch r := record.(type) {
	case ContextDialogue:
		// 1. Speaker Info
		if r.SpeakerID != nil {
			speaker, ok := input.NPCs[*r.SpeakerID]
			if ok {
				profile := &Pass2SpeakerProfile{
					Name:      speaker.Name,
					Race:      speaker.Race,
					Gender:    speaker.Gender,
					VoiceType: speaker.VoiceType,
				}
				// Solve tone instruction
				profile.ToneInstruction = e.toneResolver.Resolve(speaker.Race, speaker.VoiceType, speaker.Gender)

				// Fetch persona if available
				persona, err := e.personaLookup.FindBySpeakerID(ctx, *r.SpeakerID)
				if err == nil && persona != nil {
					profile.PersonaText = persona
				}
				pass2Ctx.Speaker = profile
			}
		}

		// 2. Summary Lookup
		if r.QuestID != nil {
			summary, err := e.summaryLookup.FindQuestSummary(ctx, *r.QuestID)
			if err == nil && summary != nil {
				pass2Ctx.QuestSummary = summary
			}
		}

		// 3. Term Lookup for the source text
		if r.Text != nil {
			t, forced, err := e.termLookup.Search(ctx, *r.Text)
			if err == nil {
				terms = t
				forcedTranslation = forced
			}
		}

	case ContextQuestStage:
		// Quest stages might not have speaker but have quest summary
		summary, err := e.summaryLookup.FindQuestSummary(ctx, r.ParentID)
		if err == nil && summary != nil {
			pass2Ctx.QuestSummary = summary
		}
		t, forced, err := e.termLookup.Search(ctx, r.Text)
		if err == nil {
			terms = t
			forcedTranslation = forced
		}

	case ContextItem:
		if r.Name != nil {
			t, forced, err := e.termLookup.Search(ctx, *r.Name)
			if err == nil {
				terms = t
				forcedTranslation = forced
			}
		}
	}

	return pass2Ctx, terms, forcedTranslation, nil
}
