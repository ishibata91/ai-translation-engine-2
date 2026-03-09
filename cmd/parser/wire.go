//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/datastore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/telemetry"
	parser2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

// InitializeParser creates a new Parser instance with all dependencies wired.
func InitializeParser(ctx context.Context) (parser2.Parser, func(), error) {
	wire.Build(
		datastore.ProviderSet,
		telemetry.ProviderSet,
		config.ConfigSet,
		parser2.ParserSet,
	)
	return nil, nil, nil
}
