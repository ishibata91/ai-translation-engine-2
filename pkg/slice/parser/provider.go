package parser

import (
	"github.com/google/wire"
)

// ProvideParser returns an implementation of Parser.
func ProvideParser() Parser {
	return newJSONLoader()
}

// ParserSet provides the loader components for dependency injection.
var ParserSet = wire.NewSet(ProvideParser)
