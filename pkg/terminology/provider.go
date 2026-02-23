package terminology

import (
	"github.com/google/wire"
)

// TerminologySet provides all components for the Term Translator slice.
var TerminologySet = wire.NewSet(
	NewTermTranslator,
	wire.Bind(new(Terminology), new(*TermTranslatorImpl)),
	// Add other internal providers if necessary (store, searcher, etc.)
)
