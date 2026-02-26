package translator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

func (s *translatorSlice) SaveResults(ctx context.Context, responses []llm.Response) error {
	slog.DebugContext(ctx, "ENTER SaveResults",
		slog.Int("responses_count", len(responses)),
	)
	start := time.Now()

	for _, resp := range responses {
		// Extract metadata
		id, ok := resp.Metadata["id"].(string)
		if !ok {
			slog.WarnContext(ctx, "response missing record id in metadata", "content", resp.Content)
			continue
		}

		recordType, _ := resp.Metadata["record_type"].(string)
		sourcePlugin, _ := resp.Metadata["source_plugin"].(string)
		tags, _ := resp.Metadata["tags"].(map[string]string)
		chunkIndex, _ := resp.Metadata["chunk_index"].(int)
		isChunked, _ := resp.Metadata["is_chunked"].(bool)

		// 1. Tag Restoration & Validation
		var restoredText string
		var status string = "completed"
		var errMsg *string

		if len(tags) > 0 {
			// Validate LLM output didn't lose or hallucinate tags
			if err := s.tagProcessor.Validate(resp.Content, tags); err != nil {
				slog.WarnContext(ctx, "tag validation failed", "id", id, "error", err)
				msg := err.Error()
				errMsg = &msg
				status = "failed"
				// We still might want to save the raw content or restored content?
				// Usually, we want to flag it for human review.
			}
			restoredText = s.tagProcessor.Postprocess(resp.Content, tags)
		} else {
			restoredText = resp.Content
		}

		// 2. Prepare Result DTO
		result := TranslationResult{
			ID:             id,
			RecordType:     recordType,
			TranslatedText: &restoredText,
			Status:         status,
			ErrorMessage:   errMsg,
			SourcePlugin:   sourcePlugin,
		}

		// Handle chunking (if chunked, we might need a more complex merging logic later,
		// but for now we store each chunk with its index)
		if isChunked {
			idx := chunkIndex
			result.Index = &idx
		}

		// 3. Write to persistent storage
		if err := s.resultWriter.Write(result); err != nil {
			slog.ErrorContext(ctx, "failed to write result", "id", id, "error", err)
			return fmt.Errorf("failed to write result for %s: %w", id, err)
		}
	}

	// 4. Finalize
	if err := s.resultWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush results: %w", err)
	}

	slog.DebugContext(ctx, "EXIT SaveResults",
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}
