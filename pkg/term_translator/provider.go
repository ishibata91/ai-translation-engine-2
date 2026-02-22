package term_translator

import (
	"github.com/google/wire"
)

// TermTranslatorSet provides all components for the TermTranslator vertical slice.
// Dependencies provided externally (from InfrastructureSet or other slices):
//   - *sql.DB (dictionary DB connection)
//   - *sql.DB (mod term DB — providers must distinguish via value types or named injection in call site)
//   - llm_client.LLMClient
//   - *slog.Logger
//   - ProgressNotifier (optional, may be nil)
//
// Internal providers:
//   - *SnowballStemmer  — created via NewSnowballStemmer (uses default "english" language)
//   - *TermPromptBuilderImpl — created via NewTermPromptBuilderDefault
//   - *TermRequestBuilderImpl — created via NewTermRequestBuilder
//   - *SQLiteTermDictionarySearcher — created via NewSQLiteTermDictionarySearcher
//   - *SQLiteModTermStore — created via NewSQLiteModTermStore
//   - *TermTranslatorImpl — created via NewTermTranslator
var TermTranslatorSet = wire.NewSet(
	// Utility
	NewSnowballStemmerDefault,
	NewTermPromptBuilderDefault,

	// Request builder
	NewTermRequestBuilder,
	wire.Bind(new(TermRequestBuilder), new(*TermRequestBuilderImpl)),

	// Dictionary searcher
	NewSQLiteTermDictionarySearcher,
	wire.Bind(new(TermDictionarySearcher), new(*SQLiteTermDictionarySearcher)),

	// Mod term store
	NewSQLiteModTermStore,
	wire.Bind(new(ModTermStore), new(*SQLiteModTermStore)),

	// Orchestrator
	NewTermTranslator,
	wire.Bind(new(TermTranslator), new(*TermTranslatorImpl)),
)

// NewSnowballStemmerDefault creates the default Snowball stemmer for English.
func NewSnowballStemmerDefault() *SnowballStemmer {
	return NewSnowballStemmer("english")
}

// NewTermPromptBuilderDefault creates a TermPromptBuilder with the default template.
// Returns TermPromptBuilder interface for DI.
func NewTermPromptBuilderDefault() (TermPromptBuilder, error) {
	return NewTermPromptBuilder("") // uses defaultTermPromptSystem
}
