//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/dictionary_builder"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/logger"
)

func initializeImporter(db *sql.DB) dictionary_builder.DictionaryImporter {
	wire.Build(
		dictionary_builder.DefaultProviderSet,
		logger.ProviderSet,
	)
	return nil
}
