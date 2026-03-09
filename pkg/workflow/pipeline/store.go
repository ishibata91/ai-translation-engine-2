package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ProcessState represents the orchestration state of a specific long-running job.
type ProcessState struct {
	ProcessID    string
	TargetSlice  string
	InputFile    string
	BatchJobID   *string
	CurrentPhase string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const (
	PhaseDispatched      = "DISPATCHED"
	PhasePendingCallback = "PENDING_CALLBACK"
	PhaseCompleted       = "COMPLETED"
	PhaseFailed          = "FAILED"
)

// Store manages the process_state.db.
type Store struct {
	db *sql.DB
}

// NewStore opens the SQLite connection and runs migrations.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open process_state db: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, pragma := range pragmas {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	if err := runMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("store migration failed: %w", err)
	}

	return &Store{db: db}, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS process_states (
		process_id TEXT PRIMARY KEY,
		target_slice TEXT NOT NULL,
		input_file TEXT NOT NULL,
		batch_job_id TEXT,
		current_phase TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("run process state migrations: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) SaveState(ctx context.Context, state ProcessState) error {
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO process_states (process_id, target_slice, input_file, batch_job_id, current_phase, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(process_id) DO UPDATE SET
			target_slice = excluded.target_slice,
			input_file = excluded.input_file,
			batch_job_id = excluded.batch_job_id,
			current_phase = excluded.current_phase,
			updated_at = excluded.updated_at
	`, state.ProcessID, state.TargetSlice, state.InputFile, state.BatchJobID, state.CurrentPhase, now, now); err != nil {
		return fmt.Errorf("save process state process_id=%s phase=%s: %w", state.ProcessID, state.CurrentPhase, err)
	}
	return nil
}

func (s *Store) GetState(ctx context.Context, processID string) (*ProcessState, error) {
	var state ProcessState
	var batchJobID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT process_id, target_slice, input_file, batch_job_id, current_phase, created_at, updated_at
		FROM process_states WHERE process_id = ?
	`, processID).Scan(&state.ProcessID, &state.TargetSlice, &state.InputFile, &batchJobID, &state.CurrentPhase, &state.CreatedAt, &state.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get process state process_id=%s: %w", processID, err)
	}
	if batchJobID.Valid {
		state.BatchJobID = &batchJobID.String
	}
	return &state, nil
}

func (s *Store) ListActiveStates(ctx context.Context) ([]ProcessState, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT process_id, target_slice, input_file, batch_job_id, current_phase, created_at, updated_at
		FROM process_states
		WHERE current_phase NOT IN (?, ?)
	`, PhaseCompleted, PhaseFailed)
	if err != nil {
		return nil, fmt.Errorf("list active process states: %w", err)
	}
	defer rows.Close()

	var states []ProcessState
	for rows.Next() {
		var state ProcessState
		var batchJobID sql.NullString
		if err := rows.Scan(&state.ProcessID, &state.TargetSlice, &state.InputFile, &batchJobID, &state.CurrentPhase, &state.CreatedAt, &state.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan active process state: %w", err)
		}
		if batchJobID.Valid {
			state.BatchJobID = &batchJobID.String
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active process states: %w", err)
	}
	return states, nil
}

func (s *Store) DeleteState(ctx context.Context, processID string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM process_states WHERE process_id = ?", processID); err != nil {
		return fmt.Errorf("delete process state process_id=%s: %w", processID, err)
	}
	return nil
}
