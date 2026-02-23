package infrastructure

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
)

// InfrastructureSet provides all infrastructure-level components.
var InfrastructureSet = wire.NewSet(
	datastore.ProviderSet,
	telemetry.ProviderSet,
	progress.ProviderSet,
	queue.ProviderSet,
	llm.NewLLMManager,
)
