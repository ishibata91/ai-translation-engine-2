package queue

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
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
	Provider     string
	Model        string
	RequestFingerprint            string
	StructuredOutputSchemaVersion string
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
	if version < 2 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		ddl := []string{
			`ALTER TABLE llm_jobs ADD COLUMN provider TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN model TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN request_fingerprint TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN structured_output_schema_version TEXT;`,
		}
		for _, q := range ddl {
			if _, err := tx.ExecContext(ctx, q); err != nil && !isDuplicateColumnError(err) {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (2, ?)", time.Now().UTC()); err != nil {
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
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	q.logger.DebugContext(ctx, "submitting jobs", slog.String("process_id", processID), slog.Int("job_count", len(reqs)))

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to begin transaction for submitting jobs", telemetry.ErrorAttrs(err)...)
		return fmt.Errorf("SubmitJobs begin tx failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO llm_jobs (
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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

		fingerprint := hashRequest(string(data))
		schemaVersion := extractSchemaVersion(req)
		jobID := uuid.New().String()
		if _, err := stmt.ExecContext(
			ctx,
			jobID,
			processID,
			string(data),
			StatusPending,
			"",
			"",
			fingerprint,
			schemaVersion,
			now,
			now,
		); err != nil {
			return fmt.Errorf("failed to insert job %s: %w", jobID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		q.logger.ErrorContext(ctx, "failed to commit jobs", telemetry.ErrorAttrs(err)...)
		return fmt.Errorf("SubmitJobs commit failed: %w", err)
	}

	q.logger.InfoContext(ctx, "jobs submitted successfully",
		slog.String("process_id", processID),
		slog.Int("inserted_count", len(reqs)),
	)
	return nil
}

// GetResults retrieves completed job responses for a given processID.
func (q *Queue) GetResults(ctx context.Context, processID string) ([]JobRequest, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	q.logger.DebugContext(ctx, "fetching job results", slog.String("process_id", processID))

	rows, err := q.db.QueryContext(ctx, `
		SELECT
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version,
			batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ?
	`, processID)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to query job results", telemetry.ErrorAttrs(err)...)
		return nil, fmt.Errorf("GetResults query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var provider, model, requestFingerprint, schemaVersion sql.NullString
		var batchJobID, responseJSON, errorMsg sql.NullString

		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&provider, &model, &requestFingerprint, &schemaVersion,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		job.Provider = nullableString(provider)
		job.Model = nullableString(model)
		job.RequestFingerprint = nullableString(requestFingerprint)
		job.StructuredOutputSchemaVersion = nullableString(schemaVersion)

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

	q.logger.DebugContext(ctx, "job results fetched", slog.String("process_id", processID), slog.Int("fetched_count", len(jobs)))
	return jobs, nil
}

// DeleteJobs performs a Hard Delete of jobs associated with the processID.
func (q *Queue) DeleteJobs(ctx context.Context, processID string) error {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	q.logger.InfoContext(ctx, "deleting jobs", slog.String("process_id", processID))

	res, err := q.db.ExecContext(ctx, "DELETE FROM llm_jobs WHERE process_id = ?", processID)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to delete jobs", telemetry.ErrorAttrs(err)...)
		return fmt.Errorf("DeleteJobs failed: %w", err)
	}

	deleted, _ := res.RowsAffected()
	q.logger.InfoContext(ctx, "jobs deleted", slog.String("process_id", processID), slog.Int64("deleted_count", deleted))
	return nil
}

// GetJobsByStatus retrieves jobs for a given processID that match a specific status.
func (q *Queue) GetJobsByStatus(ctx context.Context, processID string, status string) ([]JobRequest, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	q.logger.DebugContext(ctx, "fetching jobs by status", slog.String("process_id", processID), slog.String("status", status))

	rows, err := q.db.QueryContext(ctx, `
		SELECT
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version,
			batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ? AND status = ?
	`, processID, status)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to query jobs by status", telemetry.ErrorAttrs(err)...)
		return nil, fmt.Errorf("GetJobsByStatus query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var provider, model, requestFingerprint, schemaVersion sql.NullString
		var batchJobID, responseJSON, errorMsg sql.NullString

		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&provider, &model, &requestFingerprint, &schemaVersion,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		job.Provider = nullableString(provider)
		job.Model = nullableString(model)
		job.RequestFingerprint = nullableString(requestFingerprint)
		job.StructuredOutputSchemaVersion = nullableString(schemaVersion)

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

	q.logger.DebugContext(ctx, "jobs by status fetched", slog.String("process_id", processID), slog.String("status", status), slog.Int("fetched_count", len(jobs)))
	return jobs, nil
}

// UpdateJob updates the status, response, error message, and batch job ID for a specific job.
func (q *Queue) UpdateJob(ctx context.Context, jobID string, status string, responseJSON *string, errorMsg *string, batchJobID *string) error {
	defer telemetry.StartSpan(ctx, telemetry.ActionDBQuery)()
	q.logger.DebugContext(ctx, "updating job", slog.String("job_id", jobID), slog.String("status", status))

	res, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs 
		SET status = ?, response_json = ?, error_message = ?, batch_job_id = ?, updated_at = ?
		WHERE id = ?
	`, status, responseJSON, errorMsg, batchJobID, time.Now().UTC(), jobID)

	if err != nil {
		attrs := append(telemetry.ErrorAttrs(err), slog.String("job_id", jobID))
		q.logger.ErrorContext(ctx, "failed to update job", attrs...)
		return fmt.Errorf("UpdateJob failed: %w", err)
	}

	affected, _ := res.RowsAffected()
	q.logger.InfoContext(ctx, "job updated",
		slog.String("job_id", jobID),
		slog.String("new_status", status),
		slog.Int64("affected_rows", affected),
	)
	return nil
}

// UpdateProcessMetadata stores provider/model for all jobs in the process.
func (q *Queue) UpdateProcessMetadata(ctx context.Context, processID, provider, model string) error {
	_, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs
		SET provider = ?, model = ?, updated_at = ?
		WHERE process_id = ?
	`, provider, model, time.Now().UTC(), processID)
	if err != nil {
		return fmt.Errorf("UpdateProcessMetadata failed: %w", err)
	}
	return nil
}

func hashRequest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum[:])
}

func extractSchemaVersion(req any) string {
	if typed, ok := req.(llm.Request); ok {
		if len(typed.ResponseSchema) == 0 {
			return "none"
		}
		if v, ok := typed.Metadata["structured_output_schema_version"].(string); ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	// Schema version is optional at request stage; "none" means non-structured request.
	return "none"
}

func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate column name")
}

func nullableString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}
