package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Supported statuses
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

// JobRequest represents a single request inside the queue
type JobRequest struct {
	ID           string
	ProcessID    string
	RequestJSON  string
	Status       string
	BatchJobID   *string
	ResponseJSON *string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Queue manages the llm_jobs sqlite database connection and operations.
type Queue struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewQueue opens the SQLite connection to llm_jobs, sets pragmas, and runs DDL.
func NewQueue(ctx context.Context, defaultDSN string, logger *slog.Logger) (*Queue, error) {
	db, err := sql.Open("sqlite3", defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open llm_jobs db: %w", err)
	}

	// PRAGMA settings for performance and concurrency
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", p, err)
		}
	}

	if err := runMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("migration failed for llm_jobs db: %w", err)
	}

	q := &Queue{
		db:     db,
		logger: logger.With("slice", "JobQueue"),
	}

	return q, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	// Simple schema version management
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	var version int
	err = db.QueryRowContext(ctx, "SELECT version FROM schema_version ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if version < 1 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		ddl := `
			CREATE TABLE IF NOT EXISTS llm_jobs (
				id TEXT PRIMARY KEY,
				process_id TEXT NOT NULL,
				request_json TEXT NOT NULL,
				status TEXT NOT NULL,
				batch_job_id TEXT,
				response_json TEXT,
				error_message TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_llm_jobs_process_id ON llm_jobs(process_id);
		`
		if _, err := tx.ExecContext(ctx, ddl); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (1, ?)", time.Now().UTC()); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (q *Queue) Close() error {
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

// SubmitJobs saves incoming requests to the queue.
func (q *Queue) SubmitJobs(ctx context.Context, processID string, reqs []any) error {
	q.logger.DebugContext(ctx, "ENTER SubmitJobs", slog.String("process_id", processID), slog.Int("job_count", len(reqs)))

	// Check wait limit/deadline
	if _, ok := ctx.Deadline(); !ok {
		// Just a fallback default context wrap if not present
	}

	start := time.Now()
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		q.logger.ErrorContext(ctx, "SubmitJobs begin tx failed", slog.String("error", err.Error()))
		return fmt.Errorf("SubmitJobs begin tx failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO llm_jobs (id, process_id, request_json, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("SubmitJobs prepare failed: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for i, req := range reqs {
		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request at index %d: %w", i, err)
		}

		// generate simple id for job, appending index to process_id or generating uuid.
		// Since we need UUID, we can use a basic helper if UUID lib is used, or simulate it.
		// Let's use formatting for now. A standard approach is using github.com/google/uuid,
		// but since we only need a unique string, processID + suffix works for uniqueness.
		jobID := uuid.New().String()

		if _, err := stmt.ExecContext(ctx, jobID, processID, string(data), StatusPending, now, now); err != nil {
			return fmt.Errorf("failed to insert job %s: %w", jobID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SubmitJobs commit failed: %w", err)
	}

	q.logger.DebugContext(ctx, "EXIT SubmitJobs",
		slog.Any("result", map[string]any{"inserted": len(reqs), "elapsed": time.Since(start).String()}))
	return nil
}

// GetResults retrieves completed job responses for a given processID.
func (q *Queue) GetResults(ctx context.Context, processID string) ([]JobRequest, error) {
	q.logger.DebugContext(ctx, "ENTER GetResults", slog.String("process_id", processID))
	start := time.Now()

	rows, err := q.db.QueryContext(ctx, `
		SELECT id, process_id, request_json, status, batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ?
	`, processID)
	if err != nil {
		return nil, fmt.Errorf("GetResults query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var batchJobID, responseJSON, errorMsg sql.NullString

		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if batchJobID.Valid {
			job.BatchJobID = &batchJobID.String
		}
		if responseJSON.Valid {
			job.ResponseJSON = &responseJSON.String
		}
		if errorMsg.Valid {
			job.ErrorMessage = &errorMsg.String
		}
		jobs = append(jobs, job)
	}

	q.logger.DebugContext(ctx, "EXIT GetResults",
		slog.Any("result", map[string]any{"fetched": len(jobs), "elapsed": time.Since(start).String()}))
	return jobs, nil
}

// DeleteJobs performs a Hard Delete of jobs associated with the processID.
func (q *Queue) DeleteJobs(ctx context.Context, processID string) error {
	q.logger.DebugContext(ctx, "ENTER DeleteJobs", slog.String("process_id", processID))
	start := time.Now()

	res, err := q.db.ExecContext(ctx, "DELETE FROM llm_jobs WHERE process_id = ?", processID)
	if err != nil {
		return fmt.Errorf("DeleteJobs failed: %w", err)
	}

	deleted, _ := res.RowsAffected()

	q.logger.DebugContext(ctx, "EXIT DeleteJobs",
		slog.Any("result", map[string]any{"deleted": deleted, "elapsed": time.Since(start).String()}))
	return nil
}

// GetJobsByStatus retrieves jobs for a given processID that match a specific status.
func (q *Queue) GetJobsByStatus(ctx context.Context, processID string, status string) ([]JobRequest, error) {
	q.logger.DebugContext(ctx, "ENTER GetJobsByStatus", slog.String("process_id", processID), slog.String("status", status))
	start := time.Now()

	rows, err := q.db.QueryContext(ctx, `
		SELECT id, process_id, request_json, status, batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ? AND status = ?
	`, processID, status)
	if err != nil {
		return nil, fmt.Errorf("GetJobsByStatus query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var batchJobID, responseJSON, errorMsg sql.NullString

		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		if batchJobID.Valid {
			job.BatchJobID = &batchJobID.String
		}
		if responseJSON.Valid {
			job.ResponseJSON = &responseJSON.String
		}
		if errorMsg.Valid {
			job.ErrorMessage = &errorMsg.String
		}
		jobs = append(jobs, job)
	}

	q.logger.DebugContext(ctx, "EXIT GetJobsByStatus",
		slog.Any("result", map[string]any{"fetched": len(jobs), "elapsed": time.Since(start).String()}))
	return jobs, nil
}

// UpdateJob updates the status, response, error message, and batch job ID for a specific job.
func (q *Queue) UpdateJob(ctx context.Context, jobID string, status string, responseJSON *string, errorMsg *string, batchJobID *string) error {
	q.logger.DebugContext(ctx, "ENTER UpdateJob", slog.String("job_id", jobID), slog.String("status", status))
	start := time.Now()

	res, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs 
		SET status = ?, response_json = ?, error_message = ?, batch_job_id = ?, updated_at = ?
		WHERE id = ?
	`, status, responseJSON, errorMsg, batchJobID, time.Now().UTC(), jobID)

	if err != nil {
		return fmt.Errorf("UpdateJob failed: %w", err)
	}

	affected, _ := res.RowsAffected()

	q.logger.DebugContext(ctx, "EXIT UpdateJob",
		slog.Any("result", map[string]any{"affected": affected, "elapsed": time.Since(start).String()}))
	return nil
}
