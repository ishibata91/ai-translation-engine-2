package modelcatalog

import "context"

// Service provides model catalog listing for UI without exposing LLM internals.
type Service interface {
	ListModels(ctx context.Context, input ListModelsInput) ([]ModelOption, error)
}
