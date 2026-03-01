package dictionary

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// ProviderSet は辞書パッケージの Wire プロバイダセット。
// Store・Importer・Service の実装をまとめる。
var ProviderSet = wire.NewSet(
	NewDictionaryStore,
	NewImporter,
	NewDictionaryService,
)

// ConfigProvider は DefaultConfig を提供する。
func ConfigProvider() Config {
	return DefaultConfig()
}

// DefaultProviderSet はデフォルト設定を含む完全なプロバイダセット。
var DefaultProviderSet = wire.NewSet(
	ProviderSet,
	ConfigProvider,
)

// NewDictionaryServiceWithDefaults は依存関係を直接受け取って DictionaryService を構築する。
// Wire を使わない場合や main.go での手動ワイヤリングに使用する。
func NewDictionaryServiceWithDefaults(store DictionaryStore, notifier progress.ProgressNotifier, logger *slog.Logger) *DictionaryService {
	config := DefaultConfig()
	importer := NewImporter(config, store, notifier, logger)
	svc := NewDictionaryService(store, importer, logger)
	return svc
}
