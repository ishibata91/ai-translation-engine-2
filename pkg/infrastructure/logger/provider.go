package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/wire"
	slogcommon "github.com/samber/slog-otel"
)

// ProvideLogger returns a *slog.Logger configured with slog-otel handler.
func ProvideLogger() *slog.Logger {
	// Base JSON handler
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Wrap with a custom handler to include trace/span IDs from context
	otelHandler := &otelHandler{next: baseHandler}

	logger := slog.New(otelHandler)

	// Set as default to catch any direct slog calls
	slog.SetDefault(logger)

	return logger
}

// ProviderSet provides the logger for dependency injection.
var ProviderSet = wire.NewSet(ProvideLogger)

type otelHandler struct {
	next slog.Handler
}

func (h *otelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := slogcommon.ExtractOtelAttrFromContext([]string{}, "trace_id", "span_id")(ctx)
	r.AddAttrs(attrs...)
	return h.next.Handle(ctx, r)
}

func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{next: h.next.WithAttrs(attrs)}
}

func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{next: h.next.WithGroup(name)}
}
