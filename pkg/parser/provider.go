package parser

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
)

// ProvideParser returns an implementation of Parser.
func ProvideParser(config config.Config) Parser {
	return newJSONLoader(config)
}

// ParserSet provides the loader components for dependency injection.
var ParserSet = wire.NewSet(ProvideParser)
