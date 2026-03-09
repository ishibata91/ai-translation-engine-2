package task

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migrate ensures task slice schema exists in task.db.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create schema_version: %w", err)
	}

	var currentVersion int
	if err := db.QueryRowContext(ctx, "SELECT IFNULL(MAX(version), 0) FROM schema_version").Scan(&currentVersion); err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}

	if currentVersion >= 1 {
		return nil
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS frontend_tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			phase TEXT NOT NULL,
			progress REAL NOT NULL DEFAULT 0.0,
			error_msg TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("failed to create frontend_tasks: %w", err)
	}

	if _, err := db.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (1, ?)", time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to insert task schema version: %w", err)
	}

	return nil
}
