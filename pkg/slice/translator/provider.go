package translator

import (
	"github.com/google/wire"
)

// ProviderSet represents the dependency injection providers for the Translator Slice.
var ProviderSet = wire.NewSet(
	NewDefaultPromptBuilder,
	NewTagProcessor,
	NewBookChunker,
	NewContextEngine,
	NewTranslatorSlice,
	newPersistenceProvider,
	NewDefaultToneResolver,
	NewPersonaLookupAdapter,
	NewSummaryLookupAdapter,
	NewTermLookupAdapter,
	wire.Bind(new(ResultWriter), new(*sqlitePersistence)),
	wire.Bind(new(ResumeLoader), new(*sqlitePersistence)),
)

// newPersistenceProvider is a helper to inject persistence with a default or configured path.
func newPersistenceProvider() *sqlitePersistence {
	// In a real app, this might come from global config
	return newSqlitePersistence("output/translations")
}
