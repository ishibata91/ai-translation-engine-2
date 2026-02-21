//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure"
	"github.com/ishibata91/ai-translation-engine-2/pkg/loader_slice"
)

// InitializeLoader creates a new Loader instance with all dependencies wired.
func InitializeLoader(ctx context.Context) (loader_slice.Loader, func(), error) {
	wire.Build(
		infrastructure.InfrastructureSet,
		config_store.ConfigStoreSet,
		loader_slice.LoaderSet,
	)
	return nil, nil, nil
}
