package datastore

import (
	"context"
	"database/sql"

	base "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/datastore"
)

// NewSQLiteDB exposes SQLite connection creation at the gateway boundary.
func NewSQLiteDB(ctx context.Context, dsn string) (*sql.DB, func(), error) {
	return base.NewSQLiteDB(ctx, dsn)
}
