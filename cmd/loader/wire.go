//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/loader_slice"
)

// InitializeLoader creates a new Loader instance with all dependencies wired.
func InitializeLoader() loader_slice.Loader {
	wire.Build(loader_slice.LoaderSet)
	return nil
}
