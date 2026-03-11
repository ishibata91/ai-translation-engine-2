//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	parser2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/parser"
)

// InitializeParser creates a new Parser instance with all dependencies wired.
func InitializeParser(ctx context.Context) (parser2.Parser, func(), error) {
	wire.Build(
		parser2.ParserSet,
	)
	return nil, nil, nil
}
