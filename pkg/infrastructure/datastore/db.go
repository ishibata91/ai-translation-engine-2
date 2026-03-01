package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

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
		return nil, nil, err
	}

	return openAndConfigureDB(ctx, dbPath)
}

// resolveDBPath determines the database file path based on the OS.
func resolveDBPath(ctx context.Context, filename string) (string, error) {
	dbDir, err := getAppDataDir()
	if err != nil {
		return "", fmt.Errorf("failed to get app data dir: %w", err)
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data dir: %w", err)
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
		db.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure pool
	db.SetMaxOpenConns(1) // SQLite supports single writer
	slog.InfoContext(ctx, "database connection opened and configured", slog.String("path", dbPath))

	cleanup := func() {
		slog.DebugContext(ctx, "closing database connection", slog.String("path", dbPath))
		db.Close()
	}

	return db, cleanup, nil
}

func getAppDataDir() (string, error) {
	appName := "ai-translation-engine-2"

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, appName), nil
	case "darwin":
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable not set")
		}
		return filepath.Join(home, "Library", "Application Support", appName), nil
	default: // Linux and others
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable not set")
		}
		return filepath.Join(home, ".config", appName), nil
	}
}
