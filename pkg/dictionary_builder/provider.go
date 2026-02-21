package dictionary_builder

import (
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the dictionary_builder.
// It bundles the Store and Importer implementations into a ready-to-use set.
var ProviderSet = wire.NewSet(
	NewDictionaryStore,
	NewImporter,
)

// ConfigProvider can be used by Wire to provide the DefaultConfig if
// a dynamic config is not provided from outside.
func ConfigProvider() Config {
	return DefaultConfig()
}

// DefaultProviderSet includes the default config provider for standalone usage.
var DefaultProviderSet = wire.NewSet(
	ProviderSet,
	ConfigProvider,
)
