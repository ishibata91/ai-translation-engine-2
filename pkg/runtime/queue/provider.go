package queue

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/wire"
	gatewayconfig "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/config"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	base "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/progress"
)

// NewQueue creates the default runtime queue implementation.
func NewQueue(ctx context.Context, defaultDSN string, logger *slog.Logger) (*Queue, error) {
	return base.NewQueue(ctx, defaultDSN, logger)
}

// NewQueueProvider creates the default queue database at the runtime boundary.
func NewQueueProvider(ctx context.Context, logger *slog.Logger) (*Queue, error) {
	return base.NewQueueProvider(ctx, logger)
}

// NewWorker creates the runtime queue worker that delegates LLM calls to a gateway contract.
func NewWorker(
	queue *Queue,
	llmManager gatewayllm.LLMManager,
	configStore gatewayconfig.Config,
	secretStore gatewayconfig.SecretStore,
	notifier runtimeprogress.ProgressNotifier,
	logger *slog.Logger,
) *Worker {
	return base.NewWorker(queue, llmManager, configStore, secretStore, notifier, logger)
}

// ProviderSet exposes runtime queue providers.
var ProviderSet = wire.NewSet(
	NewQueueProvider,
	NewWorker,
)

// SetPollingInterval overrides polling cadence for runtime tests.
func SetPollingInterval(worker *Worker, d time.Duration) {
	worker.SetPollingInterval(d)
}
