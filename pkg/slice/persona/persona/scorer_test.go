package persona

import (
	"context"
	"testing"
)

func TestDefaultScorer_TableDrivenLanguageBranch(t *testing.T) {
	ctx := context.Background()
	scorer := NewDefaultScorer()

	tests := []struct {
		name     string
		text     string
		questID  *string
		service  bool
		minScore int
		maxScore int
	}{
		{
			name:     "english uppercase phrase contributes",
			text:     "We must MOVE NOW before dawn.",
			questID:  nil,
			service:  false,
			minScore: 50,
			maxScore: 200,
		},
		{
			name:     "japanese skips uppercase phrase scoring",
			text:     "今すぐ移動しろ！",
			questID:  nil,
			service:  false,
			minScore: 0,
			maxScore: 5,
		},
		{
			name:     "japanese with uppercase token still skips uppercase phrase scoring",
			text:     "今すぐ HELLO へ向かえ",
			questID:  nil,
			service:  false,
			minScore: 0,
			maxScore: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := scorer.Score(ctx, tc.text, tc.questID, tc.service)
			if got < tc.minScore || got > tc.maxScore {
				t.Fatalf("score out of expected range: got=%d want_range=[%d,%d]", got, tc.minScore, tc.maxScore)
			}
		})
	}
}

func TestDefaultScorer_EnglishUppercaseHigherThanLowercase(t *testing.T) {
	ctx := context.Background()
	scorer := NewDefaultScorer()

	upper := scorer.Score(ctx, "The enemy is RIGHT THERE", nil, false)
	lower := scorer.Score(ctx, "The enemy is right there", nil, false)

	if upper <= lower {
		t.Fatalf("expected uppercase phrase score to be higher: upper=%d lower=%d", upper, lower)
	}
}
