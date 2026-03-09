package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"

	"github.com/google/wire"
	_ "modernc.org/sqlite"
)

var ProviderSet = wire.NewSet(NewConfigDB)

// NewConfigDB creates a new SQLite database connection for "config.db".
// This is the default database used for configuration and tasks.
func NewConfigDB() (*sql.DB, func(), error) {
	return NewSQLiteDB(context.Background(), "config.db")
}

// NewSQLiteDB creates a new SQLite database connection for the given filename.
// It resolves the data directory based on the OS.
func NewSQLiteDB(ctx context.Context, filename string) (*sql.DB, func(), error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	slog.DebugContext(ctx, "opening database", slog.String("filename", filename))

	dbPath, err := resolveDBPath(ctx, filename)
	if err != nil {
		slog.ErrorContext(ctx, "failed to resolve database path", telemetry.ErrorAttrs(err)...)
		return nil, nil, fmt.Errorf("resolve sqlite db path filename=%s: %w", filename, err)
	}

	return openAndConfigureDB(ctx, dbPath)
}

// resolveDBPath determines the database file path under "<current working directory>/db".
func resolveDBPath(ctx context.Context, filename string) (string, error) {
	dbDir, err := getWorkspaceDBDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace db dir: %w", err)
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create db dir: %w", err)
	}

	path := filepath.Join(dbDir, filename)
	slog.DebugContext(ctx, "resolved database path", slog.String("path", path))
	return path, nil
}

// openAndConfigureDB opens the SQLite database and configures the connection pool.
func openAndConfigureDB(ctx context.Context, dbPath string) (*sql.DB, func(), error) {
	// Enable WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.WarnContext(ctx, "failed to close database after ping failure",
				slog.String("path", dbPath),
				slog.String("error", closeErr.Error()),
			)
		}
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure pool
	db.SetMaxOpenConns(1) // SQLite supports single writer
	slog.InfoContext(ctx, "database connection opened and configured", slog.String("path", dbPath))

	cleanup := func() {
		slog.DebugContext(ctx, "closing database connection", slog.String("path", dbPath))
		if err := db.Close(); err != nil {
			slog.WarnContext(ctx, "failed to close database connection",
				slog.String("path", dbPath),
				slog.String("error", err.Error()),
			)
		}
	}

	return db, cleanup, nil
}

func getWorkspaceDBDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return filepath.Join(wd, "db"), nil
}
