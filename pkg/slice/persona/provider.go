package persona

import (
	"github.com/google/wire"
)

// ProviderSet represents the dependency injection providers for the PersonaGenSlice.
var ProviderSet = wire.NewSet(
	NewDefaultScorer,
	wire.Bind(new(ImportanceScorer), new(*DefaultScorer)),
	NewSimpleTokenEstimator,
	wire.Bind(new(TokenEstimator), new(*SimpleTokenEstimator)),
	NewDefaultContextEvaluator,
	wire.Bind(new(ContextEvaluator), new(*DefaultContextEvaluator)),
	NewDefaultDialogueCollector,
	wire.Bind(new(DialogueCollector), new(*DefaultDialogueCollector)),
	NewPersonaStore,
	NewPersonaGenerator,
	wire.Bind(new(NPCPersonaGenerator), new(*DefaultPersonaGenerator)),
)
