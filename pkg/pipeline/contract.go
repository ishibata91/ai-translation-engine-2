package pipeline

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

// Slice represents a vertical slice that can be orchestrated by the Pipeline.
// Each slice is responsible for generating LLM prompts (Phase 1) and
// persisting the results (Phase 2).
type Slice interface {
	// ID returns the unique identifier of the slice (e.g., "Terminology").
	ID() string

	// PreparePrompts (Phase 1) generates LLM requests based on slice-specific input.
	// The input should be the DTO expected by the slice.
	PreparePrompts(ctx context.Context, input any) ([]llm_client.Request, error)

	// SaveResults (Phase 2) persists the responses received from the JobQueue.
	SaveResults(ctx context.Context, results []llm_client.Response) error
}
