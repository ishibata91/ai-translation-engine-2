//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/contract"
	"github.com/ishibata91/ai-translation-engine-2/pkg/loader"
)

// InitializeLoader creates a new Loader instance with all dependencies wired.
func InitializeLoader() contract.Loader {
	wire.Build(loader.LoaderSet)
	return nil
}
