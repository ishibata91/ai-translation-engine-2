//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
)

func initializeImporter(db *sql.DB) dictionary.DictionaryImporter {
	wire.Build(
		dictionary.DefaultProviderSet,
		telemetry.ProviderSet,
	)
	return nil
}
