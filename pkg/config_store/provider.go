package config_store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/wire"
)

var ConfigStoreSet = wire.NewSet(
	NewSQLiteStore,
	NewTypedAccessor,
	wire.Bind(new(ConfigStore), new(*SQLiteStore)),
	wire.Bind(new(UIStateStore), new(*SQLiteStore)),
	wire.Bind(new(SecretStore), new(*SQLiteStore)),
)

// NewSQLiteStore creates a new SQLiteStore instance and runs migrations.
func NewSQLiteStore(ctx context.Context, db *sql.DB) (*SQLiteStore, error) {
	if err := Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	store := &SQLiteStore{
		db:       db,
		watchers: make(map[string][]ChangeCallback),
	}
	return store, nil
}
