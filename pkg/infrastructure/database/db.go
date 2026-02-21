package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/wire"
	_ "modernc.org/sqlite"
)

var ProviderSet = wire.NewSet(NewSQLiteDB)

// NewSQLiteDB creates a new SQLite database connection.
// It resolves the data directory based on the OS.
func NewSQLiteDB() (*sql.DB, func(), error) {
	dbDir, err := getAppDataDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get app data dir: %w", err)
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create app data dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "config.db")
	db, err := sql.Open("sqlite", dbPath)
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
