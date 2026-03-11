package queue

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	_ "github.com/mattn/go-sqlite3"
)

// Supported statuses
const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
	StatusCancelled  = "CANCELLED"
)

// Request states for slice-agnostic resume control.
const (
	RequestStatePending   = "pending"
	RequestStateRunning   = "running"
	RequestStateCompleted = "completed"
	RequestStateFailed    = "failed"
	RequestStateCanceled  = "canceled"
)

// JobRequest represents a single request inside the queue
type JobRequest struct {
	ID                            string
	ProcessID                     string
	TaskID                        string `json:"task_id"`
	TaskType                      string `json:"task_type"`
	RequestState                  string `json:"request_state"`
	ResumeCursor                  int    `json:"resume_cursor"`
	RequestJSON                   string
	Status                        string
	Provider                      string
	Model                         string
	RequestFingerprint            string
	StructuredOutputSchemaVersion string
	BatchJobID                    *string
	ResponseJSON                  *string
	ErrorMessage                  *string
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

// TaskRequestState aggregates request states for one task.
type TaskRequestState struct {
	TaskID       string `json:"task_id"`
	TaskType     string `json:"task_type"`
	Total        int    `json:"total"`
	Pending      int    `json:"pending"`
	Running      int    `json:"running"`
	Completed    int    `json:"completed"`
	Failed       int    `json:"failed"`
	Canceled     int    `json:"canceled"`
	ResumeCursor int    `json:"resume_cursor"`
}

// Queue manages the llm_jobs sqlite database connection and operations.
type Queue struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewQueue opens the SQLite connection to llm_jobs, sets pragmas, and runs DDL.
func NewQueue(ctx context.Context, defaultDSN string, logger *slog.Logger) (*Queue, error) {
	resolvedDSN, err := resolveDSN(defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve queue dsn: %w", err)
	}

	db, err := sql.Open("sqlite3", resolvedDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open llm_queue db: %w", err)
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
		return nil, fmt.Errorf("migration failed for llm_queue db: %w", err)
	}

	q := &Queue{
		db:     db,
		logger: logger.With("slice", "JobQueue"),
	}

	return q, nil
}

func resolveDSN(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("dsn is required")
	}
	// Keep in-memory and URI style DSN unchanged.
	if strings.Contains(trimmed, ":memory:") || strings.HasPrefix(trimmed, "file:") {
		return trimmed, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve queue working directory: %w", err)
	}

	// Relative path is placed under workspace db/.
	if !filepath.IsAbs(trimmed) {
		path := filepath.Join(wd, "db", trimmed)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", fmt.Errorf("create queue db directory for %s: %w", path, err)
		}
		return path, nil
	}

	if err := os.MkdirAll(filepath.Dir(trimmed), 0755); err != nil {
		return "", fmt.Errorf("create queue db directory for %s: %w", trimmed, err)
	}
	return trimmed, nil
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
		defer func() {
			_ = tx.Rollback()
		}()

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
		defer func() {
			_ = tx.Rollback()
		}()

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
	if version < 3 {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		ddl := []string{
			`ALTER TABLE llm_jobs ADD COLUMN task_id TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN task_type TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN request_state TEXT;`,
			`ALTER TABLE llm_jobs ADD COLUMN resume_cursor INTEGER NOT NULL DEFAULT 0;`,
		}
		for _, q := range ddl {
			if _, err := tx.ExecContext(ctx, q); err != nil && !isDuplicateColumnError(err) {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE llm_jobs
			SET task_id = COALESCE(NULLIF(task_id, ''), process_id),
				request_state = CASE
					WHEN request_state IS NOT NULL AND request_state != '' THEN request_state
					WHEN status = 'PENDING' THEN 'pending'
					WHEN status = 'IN_PROGRESS' THEN 'running'
					WHEN status = 'COMPLETED' THEN 'completed'
					WHEN status = 'FAILED' THEN 'failed'
					WHEN status = 'CANCELLED' THEN 'canceled'
					ELSE 'pending'
				END
		`); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (3, ?)", time.Now().UTC()); err != nil {
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
		if err := q.db.Close(); err != nil {
			return fmt.Errorf("close queue db: %w", err)
		}
	}
	return nil
}

// SubmitJobs saves incoming requests to the queue.
func (q *Queue) SubmitJobs(ctx context.Context, processID string, reqs []any) error {
	return q.submitJobsInternal(ctx, processID, processID, "", reqs)
}

// SubmitTaskRequests saves incoming task requests with task metadata for resume/state queries.
func (q *Queue) SubmitTaskRequests(ctx context.Context, taskID string, taskType string, reqs []llm.Request) error {
	requests := make([]any, 0, len(reqs))
	for _, req := range reqs {
		requests = append(requests, req)
	}
	return q.submitJobsInternal(ctx, taskID, taskID, taskType, requests)
}

// SubmitTaskSharedRequests saves llmio task requests with task metadata for resume/state queries.
func (q *Queue) SubmitTaskSharedRequests(ctx context.Context, taskID string, taskType string, reqs []llmio.Request) error {
	requests := make([]any, 0, len(reqs))
	for _, req := range reqs {
		requests = append(requests, req)
	}
	return q.submitJobsInternal(ctx, taskID, taskID, taskType, requests)
}

func (q *Queue) submitJobsInternal(ctx context.Context, processID string, taskID string, taskType string, reqs []any) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.DebugContext(ctx, "submitting jobs", slog.String("process_id", processID), slog.Int("job_count", len(reqs)))

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to begin transaction for submitting jobs", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("SubmitJobs begin tx failed: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO llm_jobs (
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version, task_id, task_type, request_state, resume_cursor,
			created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("SubmitJobs prepare failed: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, req := range reqs {
		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
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
			taskID,
			taskType,
			RequestStatePending,
			0,
			now,
			now,
		); err != nil {
			return fmt.Errorf("failed to insert job %s: %w", jobID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		q.logger.ErrorContext(ctx, "failed to commit jobs", telemetry2.ErrorAttrs(err)...)
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
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.DebugContext(ctx, "fetching job results", slog.String("process_id", processID))

	rows, err := q.db.QueryContext(ctx, `
		SELECT
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version, task_id, task_type, request_state, resume_cursor,
			batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ?
	`, processID)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to query job results", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("GetResults query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var provider, model, requestFingerprint, schemaVersion sql.NullString
		var batchJobID, responseJSON, errorMsg sql.NullString

		var taskID, taskType, requestState sql.NullString
		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&provider, &model, &requestFingerprint, &schemaVersion, &taskID, &taskType, &requestState, &job.ResumeCursor,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		job.Provider = nullableString(provider)
		job.Model = nullableString(model)
		job.RequestFingerprint = nullableString(requestFingerprint)
		job.StructuredOutputSchemaVersion = nullableString(schemaVersion)
		job.TaskID = nullableString(taskID)
		job.TaskType = nullableString(taskType)
		job.RequestState = nullableString(requestState)

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
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.InfoContext(ctx, "deleting jobs", slog.String("process_id", processID))

	res, err := q.db.ExecContext(ctx, "DELETE FROM llm_jobs WHERE process_id = ?", processID)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to delete jobs", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("DeleteJobs failed: %w", err)
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteJobs rows affected failed: %w", err)
	}
	q.logger.InfoContext(ctx, "jobs deleted", slog.String("process_id", processID), slog.Int64("deleted_count", deleted))
	return nil
}

// DeleteTaskRequests performs a hard delete of jobs associated with the task_id.
func (q *Queue) DeleteTaskRequests(ctx context.Context, taskID string) error {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.InfoContext(ctx, "deleting task jobs", slog.String("task_id", taskID))

	res, err := q.db.ExecContext(ctx, "DELETE FROM llm_jobs WHERE task_id = ?", taskID)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to delete task jobs", telemetry2.ErrorAttrs(err)...)
		return fmt.Errorf("DeleteTaskRequests failed: %w", err)
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteTaskRequests rows affected failed: %w", err)
	}
	q.logger.InfoContext(ctx, "task jobs deleted", slog.String("task_id", taskID), slog.Int64("deleted_count", deleted))
	return nil
}

// GetJobsByStatus retrieves jobs for a given processID that match a specific status.
func (q *Queue) GetJobsByStatus(ctx context.Context, processID string, status string) ([]JobRequest, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.DebugContext(ctx, "fetching jobs by status", slog.String("process_id", processID), slog.String("status", status))

	rows, err := q.db.QueryContext(ctx, `
		SELECT
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version, task_id, task_type, request_state, resume_cursor,
			batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE process_id = ? AND status = ?
	`, processID, status)
	if err != nil {
		q.logger.ErrorContext(ctx, "failed to query jobs by status", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("GetJobsByStatus query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var provider, model, requestFingerprint, schemaVersion sql.NullString
		var batchJobID, responseJSON, errorMsg sql.NullString

		var taskID, taskType, requestState sql.NullString
		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&provider, &model, &requestFingerprint, &schemaVersion, &taskID, &taskType, &requestState, &job.ResumeCursor,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		job.Provider = nullableString(provider)
		job.Model = nullableString(model)
		job.RequestFingerprint = nullableString(requestFingerprint)
		job.StructuredOutputSchemaVersion = nullableString(schemaVersion)
		job.TaskID = nullableString(taskID)
		job.TaskType = nullableString(taskType)
		job.RequestState = nullableString(requestState)

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
	defer telemetry2.StartSpan(ctx, telemetry2.ActionDBQuery)()
	q.logger.DebugContext(ctx, "updating job", slog.String("job_id", jobID), slog.String("status", status))

	res, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs 
		SET status = ?, request_state = ?, response_json = ?, error_message = ?, batch_job_id = ?, updated_at = ?
		WHERE id = ?
	`, status, requestStateFromStatus(status), responseJSON, errorMsg, batchJobID, time.Now().UTC(), jobID)

	if err != nil {
		attrs := append(telemetry2.ErrorAttrs(err), slog.String("job_id", jobID))
		q.logger.ErrorContext(ctx, "failed to update job", attrs...)
		return fmt.Errorf("UpdateJob failed: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateJob rows affected failed for job_id=%s: %w", jobID, err)
	}
	q.logger.InfoContext(ctx, "job updated",
		slog.String("job_id", jobID),
		slog.String("new_status", status),
		slog.Int64("affected_rows", affected),
	)
	return nil
}

// GetTaskRequests returns all requests for one task ordered by resume cursor.
func (q *Queue) GetTaskRequests(ctx context.Context, taskID string) ([]JobRequest, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version, task_id, task_type, request_state, resume_cursor,
			batch_job_id, response_json, error_message, created_at, updated_at
		FROM llm_jobs
		WHERE task_id = ?
		ORDER BY resume_cursor ASC, created_at ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("GetTaskRequests query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobRequest
	for rows.Next() {
		var job JobRequest
		var provider, model, requestFingerprint, schemaVersion sql.NullString
		var dbTaskID, taskType, requestState sql.NullString
		var batchJobID, responseJSON, errorMsg sql.NullString
		if err := rows.Scan(
			&job.ID, &job.ProcessID, &job.RequestJSON, &job.Status,
			&provider, &model, &requestFingerprint, &schemaVersion, &dbTaskID, &taskType, &requestState, &job.ResumeCursor,
			&batchJobID, &responseJSON, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetTaskRequests scan failed: %w", err)
		}
		job.Provider = nullableString(provider)
		job.Model = nullableString(model)
		job.RequestFingerprint = nullableString(requestFingerprint)
		job.StructuredOutputSchemaVersion = nullableString(schemaVersion)
		job.TaskID = nullableString(dbTaskID)
		job.TaskType = nullableString(taskType)
		job.RequestState = nullableString(requestState)
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
	return jobs, nil
}

// GetTaskRequestState returns aggregate state for one task.
func (q *Queue) GetTaskRequestState(ctx context.Context, taskID string) (TaskRequestState, error) {
	state := TaskRequestState{TaskID: taskID}
	var maxCursor sql.NullInt64
	err := q.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(MAX(task_type), ''),
			COUNT(*),
			SUM(CASE WHEN request_state = 'pending' THEN 1 ELSE 0 END),
			SUM(CASE WHEN request_state = 'running' THEN 1 ELSE 0 END),
			SUM(CASE WHEN request_state = 'completed' THEN 1 ELSE 0 END),
			SUM(CASE WHEN request_state = 'failed' THEN 1 ELSE 0 END),
			SUM(CASE WHEN request_state = 'canceled' THEN 1 ELSE 0 END),
			MAX(resume_cursor)
		FROM llm_jobs
		WHERE task_id = ?
	`, taskID).Scan(
		&state.TaskType,
		&state.Total,
		&state.Pending,
		&state.Running,
		&state.Completed,
		&state.Failed,
		&state.Canceled,
		&maxCursor,
	)
	if err != nil {
		return TaskRequestState{}, fmt.Errorf("GetTaskRequestState failed: %w", err)
	}
	if maxCursor.Valid {
		state.ResumeCursor = int(maxCursor.Int64)
	}
	return state, nil
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

// PrepareTaskResume moves non-completed task requests back to pending so only unfinished requests are retried.
func (q *Queue) PrepareTaskResume(ctx context.Context, taskID string) error {
	_, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs
		SET status = ?, request_state = ?, error_message = NULL, updated_at = ?
		WHERE task_id = ?
		  AND request_state IN (?, ?, ?, ?)
	`, StatusPending, RequestStatePending, time.Now().UTC(), taskID, RequestStatePending, RequestStateFailed, RequestStateCanceled, RequestStateRunning)
	if err != nil {
		return fmt.Errorf("PrepareTaskResume failed: %w", err)
	}
	return nil
}

// MarkTaskRequestsCanceled marks unfinished requests as canceled for a task.
func (q *Queue) MarkTaskRequestsCanceled(ctx context.Context, taskID string) error {
	message := "canceled by user"
	_, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs
		SET status = ?, request_state = ?, error_message = ?, updated_at = ?
		WHERE task_id = ?
		  AND request_state IN (?, ?)
	`, StatusCancelled, RequestStateCanceled, message, time.Now().UTC(), taskID, RequestStatePending, RequestStateRunning)
	if err != nil {
		return fmt.Errorf("MarkTaskRequestsCanceled failed: %w", err)
	}
	return nil
}

func hashRequest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum[:])
}

func extractSchemaVersion(req any) string {
	switch typed := req.(type) {
	case llm.Request:
		if len(typed.ResponseSchema) == 0 {
			return "none"
		}
		if v, ok := typed.Metadata["structured_output_schema_version"].(string); ok && strings.TrimSpace(v) != "" {
			return v
		}
	case llmio.Request:
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

func requestStateFromStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case StatusPending:
		return RequestStatePending
	case StatusInProgress:
		return RequestStateRunning
	case StatusCompleted:
		return RequestStateCompleted
	case StatusFailed:
		return RequestStateFailed
	case StatusCancelled:
		return RequestStateCanceled
	default:
		return RequestStatePending
	}
}
