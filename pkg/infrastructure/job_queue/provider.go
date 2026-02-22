package job_queue

import (
	"context"
	"log/slog"

	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for the job_queue package.
var ProviderSet = wire.NewSet(
	NewQueueProvider,
	NewWorker,
)

// NewQueueProvider is a wire-compatible wrapper for creating the Queue.
func NewQueueProvider(ctx context.Context, logger *slog.Logger) (*Queue, error) {
	return NewQueue(ctx, "llm_jobs.db", logger)
}
