package loader_slice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
)

// jsonLoader implements contract.Loader interface.
type jsonLoader struct {
	config config_store.ConfigStore
}

// newJSONLoader creates a new instance of jsonLoader.
func newJSONLoader(config config_store.ConfigStore) Loader {
	return &jsonLoader{config: config}
}

// LoadExtractedJSON loads extracted data from a JSON file.
// It follows the Two-Phase Load strategy:
// 1. Decode file into map[string]json.RawMessage (Serial)
// 2. Unmarshal and normalize each section in parallel (Parallel)
func (l *jsonLoader) LoadExtractedJSON(ctx context.Context, path string) (*LoaderOutput, error) {
	slog.DebugContext(ctx, "ENTER jsonLoader.LoadExtractedJSON", slog.String("path", path))
	defer slog.DebugContext(ctx, "EXIT jsonLoader.LoadExtractedJSON")

	// Phase 1: Serial Decode
	rawMap, err := DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("phase 1 (decode) failed: %w", err)
	}

	// Phase 2: Parallel Process
	processor := NewParallelProcessor(rawMap)
	data, err := processor.Process(ctx)
	if err != nil {
		return nil, fmt.Errorf("phase 2 (process) failed: %w", err)
	}

	data.SourceJSON = path
	return data, nil
}
