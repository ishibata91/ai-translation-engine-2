package progress

import "github.com/google/wire"

// NewNoopNotifier は NoopNotifier を ProgressNotifier として返す Wire Provider。
func NewNoopNotifier() ProgressNotifier {
	return &NoopNotifier{}
}

// ProviderSet は progress パッケージの Wire ProviderSet。
var ProviderSet = wire.NewSet(NewNoopNotifier)
