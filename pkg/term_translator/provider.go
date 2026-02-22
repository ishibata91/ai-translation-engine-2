package term_translator

import (
	"github.com/google/wire"
)

// TermTranslatorSet provides all components for the Term Translator slice.
var TermTranslatorSet = wire.NewSet(
	NewTermTranslator,
	wire.Bind(new(TermTranslator), new(*TermTranslatorImpl)),
	// Add other internal providers if necessary (store, searcher, etc.)
)
