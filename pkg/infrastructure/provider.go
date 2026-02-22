package infrastructure

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/database"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_manager"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/logger"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// InfrastructureSet provides all infrastructure-level components.
var InfrastructureSet = wire.NewSet(
	database.ProviderSet,
	logger.ProviderSet,
	progress.ProviderSet,
	llm_manager.NewLLMManager,
)
