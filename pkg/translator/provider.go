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
	NewPersistenceProvider,
	NewDefaultToneResolver,
	NewPersonaLookupAdapter,
	NewSummaryLookupAdapter,
	NewTermLookupAdapter,
	wire.Bind(new(ResultWriter), new(*sqlitePersistence)),
	wire.Bind(new(ResumeLoader), new(*sqlitePersistence)),
)

// NewPersistenceProvider is a helper to inject persistence with a default or configured path.
func NewPersistenceProvider() *sqlitePersistence {
	// In a real app, this might come from global config
	return NewSqlitePersistence("output/translations")
}
