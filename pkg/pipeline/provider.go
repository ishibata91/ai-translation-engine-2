package pipeline

import (
	"context"
	"log/slog"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/job_queue"
)

// ProviderSet represents the dependencies for pipeline.
var ProviderSet = wire.NewSet(
	NewStoreProvider,
	ManagerProvider,
	NewHandler,
)

// NewStoreProvider is a helper for Wire to provide the Store with a specific DSN.
func NewStoreProvider(ctx context.Context, logger *slog.Logger) (*Store, error) {
	// For now using a local file, ideally path from config in future
	dsn := "process_state.db"
	return NewStore(ctx, dsn)
}

// ManagerProvider provides the Manager and ensures it's initialized.
func ManagerProvider(
	store *Store,
	jobQueue *job_queue.Queue,
	worker *job_queue.Worker,
	logger *slog.Logger,
) *Manager {
	return NewManager(store, jobQueue, worker, logger)
}
