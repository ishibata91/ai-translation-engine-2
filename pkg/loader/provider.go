package loader

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/contract"
)

// ProvideLoader returns an implementation of contract.Loader.
func ProvideLoader() contract.Loader {
	return newJSONLoader()
}

// LoaderSet provides the loader components for dependency injection.
var LoaderSet = wire.NewSet(ProvideLoader)
