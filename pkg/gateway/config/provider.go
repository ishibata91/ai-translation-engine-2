package config

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/wire"
	config2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

// ConfigSet re-exports the existing config providers at the gateway boundary.
var ConfigSet = config2.ConfigSet

// SQLiteStore aliases the concrete gateway implementation used at composition root.
type SQLiteStore = config2.SQLiteStore

// TypedAccessor aliases the typed config accessor at the gateway boundary.
type TypedAccessor = config2.TypedAccessor

// NewSQLiteStore creates the gateway-backed config store.
func NewSQLiteStore(ctx context.Context, db *sql.DB, logger *slog.Logger) (*SQLiteStore, error) {
	return config2.NewSQLiteStore(ctx, db, logger)
}

// NewTypedAccessor creates a typed accessor for configuration reads.
func NewTypedAccessor(store Config) *TypedAccessor {
	return config2.NewTypedAccessor(store)
}

// ProviderSet exposes the gateway config providers with gateway-local bindings.
var ProviderSet = wire.NewSet(
	NewSQLiteStore,
	NewTypedAccessor,
	wire.Bind(new(Config), new(*SQLiteStore)),
	wire.Bind(new(UIStateStore), new(*SQLiteStore)),
	wire.Bind(new(SecretStore), new(*SQLiteStore)),
)
