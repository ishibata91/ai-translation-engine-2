package persona_gen

import (
	"context"
	"log/slog"
	"sort"
	"time"
)

// DefaultContextEvaluator filters dialogue entries based on their importance score,
// ensuring the total token count does not exceed the allowed maximum.
type DefaultContextEvaluator struct {
	Scorer         ImportanceScorer
	TokenEstimator TokenEstimator
}

// NewDefaultContextEvaluator creates a new DefaultContextEvaluator.
func NewDefaultContextEvaluator(scorer ImportanceScorer, estimator TokenEstimator) *DefaultContextEvaluator {
	return &DefaultContextEvaluator{
		Scorer:         scorer,
		TokenEstimator: estimator,
	}
}

// Evaluate performs token usage estimation and trims dialogues to fit the context window.
func (e *DefaultContextEvaluator) Evaluate(ctx context.Context, dialogueData NPCDialogueData, config PersonaConfig) (TokenEstimation, []DialogueEntry) {
	slog.DebugContext(ctx, "ENTER Evaluate",
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", dialogueData.SpeakerID),
		slog.Int("dialogue_count", len(dialogueData.Dialogues)),
	)
	start := time.Now()

	availableTokens := config.ContextWindowLimit - config.SystemPromptOverhead

	estimation := TokenEstimation{
		ExceedsLimit: false,
		InputTokens:  0,
		OutputTokens: config.MaxOutputTokens, // Placeholder for output allocation
		TotalTokens:  0,
	}

	if availableTokens <= 0 {
		slog.DebugContext(ctx, "EXIT Evaluate",
			slog.String("slice", "PersonaGen"),
			slog.String("speaker_id", dialogueData.SpeakerID),
			slog.String("reason", "no_available_tokens"),
			slog.Duration("elapsed", time.Since(start)),
		)
		return estimation, []DialogueEntry{} // No room for anything
	}

	var scored []ScoredDialogueEntry
	for _, entry := range dialogueData.Dialogues {
		s := e.Scorer.Score(ctx, entry.EnglishText, &entry.QuestID, entry.IsServicesBranch)
		t := e.TokenEstimator.Estimate(ctx, entry.EnglishText)

		// Accumulate original tokens before dropping
		estimation.InputTokens += t

		scored = append(scored, ScoredDialogueEntry{
			Entry:           entry,
			ImportanceScore: s,
			// ProperNounHits and EmotionIndicators could be returned by a more advanced Scorer
		})
	}

	estimation.TotalTokens = estimation.InputTokens + estimation.OutputTokens
	if estimation.InputTokens > availableTokens {
		estimation.ExceedsLimit = true
	}

	// Sort by highest score first
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].ImportanceScore > scored[j].ImportanceScore
	})

	var selected []DialogueEntry
	currentTokens := 0

	for _, se := range scored {
		t := e.TokenEstimator.Estimate(ctx, se.Entry.EnglishText)
		if currentTokens+t <= availableTokens {
			selected = append(selected, se.Entry)
			currentTokens += t
		}
	}

	slog.DebugContext(ctx, "EXIT Evaluate",
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", dialogueData.SpeakerID),
		slog.Int("selected_count", len(selected)),
		slog.Bool("exceeds_limit", estimation.ExceedsLimit),
		slog.Int("total_tokens", estimation.TotalTokens),
		slog.Duration("elapsed", time.Since(start)),
	)

	return estimation, selected
}
