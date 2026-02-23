//go:build wireinject
// +build wireinject

package summary_generator

import (
	"database/sql"

	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for the summary_generator slice.
var ProviderSet = wire.NewSet(
	NewSummaryStore,
	NewSummaryGenerator,
)

// To be used by the main wire.go if needed, or by testing.
func InitializeGenerator(db *sql.DB) (SummaryGenerator, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
