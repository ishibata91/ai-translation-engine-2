package loader_slice

import (
	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
)

// ProvideLoader returns an implementation of Loader.
func ProvideLoader(config config_store.ConfigStore) Loader {
	return newJSONLoader(config)
}

// LoaderSet provides the loader components for dependency injection.
var LoaderSet = wire.NewSet(ProvideLoader)
