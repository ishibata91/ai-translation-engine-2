package config

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// Migrate initializes the database schema for the Config.
func Migrate(ctx context.Context, db *sql.DB) error {
	slog.DebugContext(ctx, "ENTER Migrate")
	defer slog.DebugContext(ctx, "EXIT Migrate")

	if err := ensureSchemaVersionTable(ctx, db); err != nil {
		return err
	}

	currentVersion, err := getCurrentVersion(ctx, db)
	if err != nil {
		return err
	}

	return applyPendingMigrations(ctx, db, currentVersion)
}

// ensureSchemaVersionTable creates the schema_version table if it doesn't exist.
func ensureSchemaVersionTable(ctx context.Context, db *sql.DB) error {
	slog.DebugContext(ctx, "ENTER ensureSchemaVersionTable")

	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}
	return nil
}

// getCurrentVersion retrieves the current schema version from the database.
func getCurrentVersion(ctx context.Context, db *sql.DB) (int, error) {
	slog.DebugContext(ctx, "ENTER getCurrentVersion")

	var currentVersion int
	err := db.QueryRowContext(ctx, "SELECT IFNULL(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query schema version: %w", err)
	}
	return currentVersion, nil
}

// applyPendingMigrations runs all migrations that haven't been applied yet.
func applyPendingMigrations(ctx context.Context, db *sql.DB, currentVersion int) error {
	slog.DebugContext(ctx, "ENTER applyPendingMigrations", slog.Int("currentVersion", currentVersion))

	targetVersion := 1
	if currentVersion < targetVersion {
		if err := runMigrationV1(ctx, db); err != nil {
			return err
		}
	}
	return nil
}

func runMigrationV1(ctx context.Context, db *sql.DB) error {
	slog.InfoContext(ctx, "ENTER runMigrationV1")

	queries := buildV1MigrationQueries()
	return executeMigrationQueries(ctx, db, queries)
}

// buildV1MigrationQueries returns the SQL statements for schema version 1.
func buildV1MigrationQueries() []string {
	return []string{
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
}

// executeMigrationQueries executes a list of migration SQL statements sequentially.
func executeMigrationQueries(ctx context.Context, db *sql.DB, queries []string) error {
	slog.DebugContext(ctx, "ENTER executeMigrationQueries", slog.Int("queryCount", len(queries)))

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
