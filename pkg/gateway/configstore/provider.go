package configstore

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/wire"
)

// ProviderSet exposes the gateway config providers with gateway-local bindings.
var ProviderSet = wire.NewSet(
	NewSQLiteStore,
	wire.Bind(new(Config), new(*SQLiteStore)),
	wire.Bind(new(UIStateStore), new(*SQLiteStore)),
	wire.Bind(new(SecretStore), new(*SQLiteStore)),
)

// NewSQLiteStore creates a new SQLiteStore instance and runs migrations.
func NewSQLiteStore(ctx context.Context, db *sql.DB, logger *slog.Logger) (*SQLiteStore, error) {
	if err := Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	store := &SQLiteStore{
		db:       db,
		logger:   logger.With("component", "SQLiteStore"),
		watchers: make(map[string]map[int]ChangeCallback),
	}
	return store, nil
}
