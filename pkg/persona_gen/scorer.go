package persona_gen

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/jdkato/prose/v2"
)

// DefaultScorer implements ImportanceScorer based on nouns and emotion heuristics.
type DefaultScorer struct {
	WeightNoun    int
	WeightEmotion int
	BasePriority  int
}

// NewDefaultScorer creates a new DefaultScorer with standard weights.
func NewDefaultScorer() *DefaultScorer {
	return &DefaultScorer{
		WeightNoun:    2,
		WeightEmotion: 1,
		BasePriority:  1, // Default base priority if not quest event
	}
}

// Score computes the importance of the dialogue text.
func (s *DefaultScorer) Score(ctx context.Context, englishText string, questID *string, isServicesBranch bool) int {
	slog.DebugContext(ctx, "ENTER Score",
		slog.String("slice", "PersonaGen"),
		slog.Int("text_length", len(englishText)),
		slog.Bool("has_quest", questID != nil),
		slog.Bool("is_services", isServicesBranch),
	)
	start := time.Now()

	score := 0

	// Add base priority if it's a quest event
	if questID != nil {
		score += s.BasePriority * 2 // Give quest events higher base priority
	} else if isServicesBranch {
		score += s.BasePriority // Give service branches a baseline priority too
	} else {
		score += 0 // Generic dialogue
	}

	// 1. Emotion heuristic
	if strings.Contains(englishText, "!") || strings.Contains(englishText, "?") {
		score += s.WeightEmotion
	}
	// Check for ALL CAPS (proxy for strong emotion/shouting) - simplified check
	if isAllCaps(englishText) {
		score += s.WeightEmotion
	}

	// 2. Noun heuristic using full NLP library
	properNounCount := countProperNounsProse(ctx, englishText)
	score += properNounCount * s.WeightNoun

	slog.DebugContext(ctx, "EXIT Score",
		slog.String("slice", "PersonaGen"),
		slog.Int("score", score),
		slog.Int("noun_count", properNounCount),
		slog.Duration("elapsed", time.Since(start)),
	)

	return score
}

// isAllCaps checks if the text has letters and they are all uppercase.
func isAllCaps(text string) bool {
	hasLetter := false
	for _, r := range text {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// countProperNounsProse uses the prose NLP library to count proper nouns.
func countProperNounsProse(ctx context.Context, text string) int {
	// prose is an English NLP library.
	if text == "" {
		return 0
	}

	doc, err := prose.NewDocument(text)
	if err != nil {
		slog.WarnContext(ctx, "prose NLP failed to parse document, falling back to 0 proper nouns",
			slog.String("slice", "PersonaGen"),
			slog.String("error", err.Error()),
		)
		return 0
	}

	count := 0
	for _, tok := range doc.Tokens() {
		// NNP = Proper noun, singular
		// NNPS = Proper noun, plural
		if tok.Tag == "NNP" || tok.Tag == "NNPS" {
			count++
		}
	}

	return count
}
