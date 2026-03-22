package llm

import (
	"context"
	"log/slog"
)

func logFinalPrompt(ctx context.Context, logger *slog.Logger, provider, mode string, req Request, extraAttrs ...slog.Attr) {
	if logger == nil {
		return
	}

	attrs := []slog.Attr{
		slog.String("provider", provider),
		slog.String("mode", mode),
		slog.Int("system_prompt_len", len(req.SystemPrompt)),
		slog.Int("user_prompt_len", len(req.UserPrompt)),
		slog.String("system_prompt", req.SystemPrompt),
		slog.String("user_prompt", req.UserPrompt),
	}
	if queueJobID, ok := readBatchMetadataString(req.Metadata, BatchMetadataQueueJobIDKey); ok {
		attrs = append(attrs, slog.String("queue_job_id", queueJobID))
	}
	attrs = append(attrs, extraAttrs...)

	logger.InfoContext(ctx, "LLM final prompt", attrsToAny(attrs)...)
}

func attrsToAny(attrs []slog.Attr) []any {
	result := make([]any, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, attr.Key, attr.Value.Any())
	}
	return result
}

func requestIndexAttr(index int) slog.Attr {
	return slog.Int("request_index", index)
}

func requestModeAttr(structured bool) string {
	if structured {
		return "structured"
	}
	return "sync"
}
