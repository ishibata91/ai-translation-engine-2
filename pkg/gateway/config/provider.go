package config

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/wire"
	base "github.com/ishibata91/ai-translation-engine-2/pkg/config"
)

// ConfigSet re-exports the existing config providers at the gateway boundary.
var ConfigSet = base.ConfigSet

// SQLiteStore aliases the concrete gateway implementation used at composition root.
type SQLiteStore = base.SQLiteStore

// TypedAccessor aliases the typed config accessor at the gateway boundary.
type TypedAccessor = base.TypedAccessor

// NewSQLiteStore creates the gateway-backed config store.
func NewSQLiteStore(ctx context.Context, db *sql.DB, logger *slog.Logger) (*SQLiteStore, error) {
	return base.NewSQLiteStore(ctx, db, logger)
}

// NewTypedAccessor creates a typed accessor for configuration reads.
func NewTypedAccessor(store Config) *TypedAccessor {
	return base.NewTypedAccessor(store)
}

// ProviderSet exposes the gateway config providers with gateway-local bindings.
var ProviderSet = wire.NewSet(
	NewSQLiteStore,
	NewTypedAccessor,
	wire.Bind(new(Config), new(*SQLiteStore)),
	wire.Bind(new(UIStateStore), new(*SQLiteStore)),
	wire.Bind(new(SecretStore), new(*SQLiteStore)),
)
