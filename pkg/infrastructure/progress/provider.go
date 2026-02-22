package progress

import "github.com/google/wire"

// NewHubProvider provides the Hub as a singleton.
func NewHubProvider() *Hub {
	return NewHub()
}

// ProviderSet は progress パッケージの Wire ProviderSet。
var ProviderSet = wire.NewSet(
	NewHubProvider,
	wire.Bind(new(ProgressNotifier), new(*Hub)),
)
