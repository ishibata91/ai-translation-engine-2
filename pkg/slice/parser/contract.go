package parser

import (
	"context"
)

// Parser defines the interface for loading extracted data.
type Parser interface {
	// LoadExtractedJSON loads extracted data from a JSON file.
	// It supports automatic encoding detection and parallel processing.
	LoadExtractedJSON(ctx context.Context, path string) (*ParserOutput, error)
}
