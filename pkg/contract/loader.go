package contract

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
)

// Loader defines the interface for loading extracted data.
type Loader interface {
	// LoadExtractedJSON loads extracted data from a JSON file.
	// It supports automatic encoding detection and parallel processing.
	LoadExtractedJSON(ctx context.Context, path string) (*models.ExtractedData, error)
}
