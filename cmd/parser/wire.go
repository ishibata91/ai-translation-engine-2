//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
)

// InitializeParser creates a new Parser instance with all dependencies wired.
func InitializeParser(ctx context.Context) (parser.Parser, func(), error) {
	wire.Build(
		infrastructure.InfrastructureSet,
		config.ConfigSet,
		parser.ParserSet,
	)
	return nil, nil, nil
}
