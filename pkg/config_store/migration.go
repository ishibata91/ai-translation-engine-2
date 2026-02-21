package config_store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// Migrate initializes the database schema for the ConfigStore.
func Migrate(ctx context.Context, db *sql.DB) error {
	slog.DebugContext(ctx, "ENTER Migrate")
	defer slog.DebugContext(ctx, "EXIT Migrate")

	// Create schema_version table if it doesn't exist
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Check current version
	var currentVersion int
	err = db.QueryRowContext(ctx, "SELECT IFNULL(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to query schema version: %w", err)
	}

	targetVersion := 1
	if currentVersion < targetVersion {
		err = runMigrationV1(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func runMigrationV1(ctx context.Context, db *sql.DB) error {
	slog.InfoContext(ctx, "Running migration to version 1")

	queries := []string{
		`CREATE TABLE IF NOT EXISTS config (
			namespace TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (namespace, key)
		);`,
		`CREATE TABLE IF NOT EXISTS ui_state (
			namespace TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (namespace, key)
		);`,
		`CREATE TABLE IF NOT EXISTS secrets (
			namespace TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (namespace, key)
		);`,
		`INSERT INTO schema_version (version, applied_at) VALUES (1, ?);`,
	}

	now := time.Now()
	for _, q := range queries {
		var err error
		if q == queries[len(queries)-1] {
			_, err = db.ExecContext(ctx, q, now)
		} else {
			_, err = db.ExecContext(ctx, q)
		}
		if err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}

	return nil
}
