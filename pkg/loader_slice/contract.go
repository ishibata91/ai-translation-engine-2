package loader_slice

import (
	"context"
)

// Loader defines the interface for loading extracted data.
type Loader interface {
	// LoadExtractedJSON loads extracted data from a JSON file.
	// It supports automatic encoding detection and parallel processing.
	LoadExtractedJSON(ctx context.Context, path string) (*LoaderOutput, error)
}
