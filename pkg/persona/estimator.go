package persona

import (
	"context"
	"log/slog"
	"time"
	"unicode/utf8"
)

// SimpleTokenEstimator implements TokenEstimator using a rough text length heuristic.
type SimpleTokenEstimator struct {
	CharsPerToken int // Typically ~4 chars per token for English
}

// NewSimpleTokenEstimator creates a new TokenEstimator.
func NewSimpleTokenEstimator() *SimpleTokenEstimator {
	return &SimpleTokenEstimator{
		CharsPerToken: 4,
	}
}

// Estimate returns the estimated token count of the text.
func (e *SimpleTokenEstimator) Estimate(ctx context.Context, text string) int {
	slog.DebugContext(ctx, "ENTER Estimate",
		slog.String("slice", "Persona"),
		slog.Int("text_length", len(text)),
	)
	start := time.Now()

	length := utf8.RuneCountInString(text)
	if length == 0 {
		slog.DebugContext(ctx, "EXIT Estimate",
			slog.String("slice", "Persona"),
			slog.Int("estimated_tokens", 0),
			slog.Duration("elapsed", time.Since(start)),
		)
		return 0
	}
	// length / 4 + some buffer just in case
	estimated := (length / e.CharsPerToken) + 1

	slog.DebugContext(ctx, "EXIT Estimate",
		slog.String("slice", "Persona"),
		slog.Int("estimated_tokens", estimated),
		slog.Duration("elapsed", time.Since(start)),
	)

	return estimated
}
