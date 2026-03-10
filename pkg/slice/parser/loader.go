package parser

import (
	"context"
	"fmt"
	"log/slog"

	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

// jsonLoader implements contract.Parser interface.
type jsonLoader struct {
	config config.Config
}

// newJSONLoader creates a new instance of jsonLoader.
func newJSONLoader(config config.Config) Parser {
	return &jsonLoader{config: config}
}

// LoadExtractedJSON loads extracted data from a JSON file.
// It follows the Two-Phase Load strategy:
// 1. Decode file into map[string]json.RawMessage (Serial)
// 2. Unmarshal and normalize each section in parallel (Parallel)
func (l *jsonLoader) LoadExtractedJSON(ctx context.Context, path string) (*ParserOutput, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionParser)()
	slog.DebugContext(ctx, "starting JSON load", slog.String("path", path))

	// Phase 1: Serial Decode
	rawMap, err := DecodeFile(path)
	if err != nil {
		slog.ErrorContext(ctx, "JSON decode phase failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("phase 1 (decode) failed: %w", err)
	}

	slog.DebugContext(ctx, "JSON decoded", slog.Int("section_count", len(rawMap)))

	// Phase 2: Parallel Process
	processor := NewParallelProcessor(rawMap)
	data, err := processor.Process(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "parallel processing phase failed", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("phase 2 (process) failed: %w", err)
	}

	data.SourceJSON = path
	slog.InfoContext(ctx, "JSON load completed",
		slog.String("path", path),
		slog.Int("dialogue_group_count", len(data.DialogueGroups)),
		slog.Int("npc_count", len(data.NPCs)),
	)
	return data, nil
}
