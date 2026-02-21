package loader_slice

import (
	"github.com/google/wire"
)

// ProvideLoader returns an implementation of Loader.
func ProvideLoader() Loader {
	return newJSONLoader()
}

// LoaderSet provides the loader components for dependency injection.
var LoaderSet = wire.NewSet(ProvideLoader)
