package infrastructure

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/database"
)

// InfrastructureSet provides all infrastructure-level components.
var InfrastructureSet = wire.NewSet(
	database.ProviderSet,
)
