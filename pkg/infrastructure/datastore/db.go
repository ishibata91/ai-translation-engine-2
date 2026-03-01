package datastore

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/wire"
	_ "modernc.org/sqlite"
)

var ProviderSet = wire.NewSet(NewConfigDB)

// NewConfigDB creates a new SQLite database connection for "config.db".
// This is the default database used for configuration and tasks.
func NewConfigDB() (*sql.DB, func(), error) {
	return NewSQLiteDB("config.db")
}

// NewSQLiteDB creates a new SQLite database connection for the given filename.
// It resolves the data directory based on the OS.
func NewSQLiteDB(filename string) (*sql.DB, func(), error) {
	slog.Debug("ENTER NewSQLiteDB", slog.String("filename", filename))
	defer slog.Debug("EXIT NewSQLiteDB")

	dbPath, err := resolveDBPath(filename)
	if err != nil {
		return nil, nil, err
	}

	return openAndConfigureDB(dbPath)
}

// resolveDBPath determines the database file path based on the OS.
func resolveDBPath(filename string) (string, error) {
	slog.Debug("ENTER resolveDBPath", slog.String("filename", filename))

	dbDir, err := getAppDataDir()
	if err != nil {
		return "", fmt.Errorf("failed to get app data dir: %w", err)
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data dir: %w", err)
	}

	return filepath.Join(dbDir, filename), nil
}

// openAndConfigureDB opens the SQLite database and configures the connection pool.
func openAndConfigureDB(dbPath string) (*sql.DB, func(), error) {
	slog.Debug("ENTER openAndConfigureDB", slog.String("dbPath", dbPath))

	// Enable WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure pool
	db.SetMaxOpenConns(1) // SQLite supports single writer

	cleanup := func() {
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
