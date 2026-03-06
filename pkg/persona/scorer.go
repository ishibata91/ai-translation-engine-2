package persona

import (
	"context"
	"log/slog"
	"strings"
	"time"
	"unicode"
)

// DefaultScorer implements ImportanceScorer using lightweight heuristics.
type DefaultScorer struct {
	WeightUppercasePhrase int
	WeightEmotion         int
	BasePriority          int
}

// NewDefaultScorer creates a new DefaultScorer with standard weights.
func NewDefaultScorer() *DefaultScorer {
	return &DefaultScorer{
		WeightUppercasePhrase: 4,
		WeightEmotion:         1,
		BasePriority:          1, // Default base priority if not quest event
	}
}

// Score computes the importance of the dialogue text.
func (s *DefaultScorer) Score(ctx context.Context, englishText string, questID *string, isServicesBranch bool) int {
	slog.DebugContext(ctx, "ENTER Score",
		slog.String("slice", "Persona"),
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

	isJapanese := isLikelyJapanese(englishText)
	uppercasePhraseRatio := 0.0
	if !isJapanese {
		uppercasePhraseRatio = countUppercasePhraseRatio(englishText)
		score += int(uppercasePhraseRatio * 100.0 * float64(s.WeightUppercasePhrase))
	}

	slog.DebugContext(ctx, "EXIT Score",
		slog.String("slice", "Persona"),
		slog.Int("score", score),
		slog.Bool("is_japanese", isJapanese),
		slog.Float64("uppercase_phrase_ratio", uppercasePhraseRatio),
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

func isLikelyJapanese(text string) bool {
	for _, r := range text {
		if isJapaneseRune(r) {
			return true
		}
	}
	return false
}

func isJapaneseRune(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0x4E00 && r <= 0x9FFF) // CJK Unified Ideographs
}

func countUppercasePhraseRatio(text string) float64 {
	tokens := strings.Fields(text)
	if len(tokens) == 0 {
		return 0
	}

	totalAlphaTokens := 0
	uppercaseTokens := 0
	for _, raw := range tokens {
		token := strings.Trim(raw, ".,!?;:\"'()[]{}")
		if token == "" {
			continue
		}

		hasLetter := false
		isUpper := true
		for _, r := range token {
			if !unicode.IsLetter(r) {
				continue
			}
			hasLetter = true
			if !unicode.IsUpper(r) {
				isUpper = false
				break
			}
		}

		if !hasLetter {
			continue
		}
		totalAlphaTokens++
		if isUpper {
			uppercaseTokens++
		}
	}

	if totalAlphaTokens == 0 {
		return 0
	}

	return float64(uppercaseTokens) / float64(totalAlphaTokens)
}
