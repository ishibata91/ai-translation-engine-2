//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/contract"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/loader"
)

// InitializeLoader creates a new Loader instance with all dependencies wired.
func InitializeLoader() contract.Loader {
	wire.Build(
		loader.NewJSONLoader,
		// In the future, other dependencies will be added here
	)
	return nil
}
