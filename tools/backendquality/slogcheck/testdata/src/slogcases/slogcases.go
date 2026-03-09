package slogcases

import (
	"context"
	"log/slog"
)

type wrappedLogger struct{}

func (wrappedLogger) Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func directPackageCall() {
	slog.Info("started") // want "slog usage: use slog.\\*Context when calling log/slog directly"
}

func directLoggerMethod(logger *slog.Logger) {
	logger.Warn("queued") // want "slog usage: use slog.\\*Context when calling log/slog directly"
}

func directLoggerMethodWithBadKey(logger *slog.Logger, recordCount int) {
	logger.Error("failed", "recordCount", recordCount) // want "slog usage: use slog.\\*Context when calling log/slog directly" "slog usage: structured log keys must be lower_snake_case"
}

func attrBuilderBadKey() slog.Attr {
	return slog.String("taskId", "1") // want "slog usage: structured log keys must be lower_snake_case"
}

func goodContextCall(ctx context.Context, taskID string, recordCount int) {
	slog.InfoContext(ctx, "done", "task_id", taskID, slog.Int("record_count", recordCount))
}

func goodLoggerContextCall(ctx context.Context, logger *slog.Logger, taskID string) {
	logger.DebugContext(ctx, "done", "task_id", taskID)
}

func goodWrapper(ctx context.Context, logger wrappedLogger, taskID string) {
	logger.Info(ctx, "done", "task_id", taskID)
}
