package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InsertTask(ctx context.Context, task Task) error {
	metadataJSON, err := json.Marshal(task.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO frontend_tasks (id, name, type, status, phase, progress, error_msg, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.Name, task.Type, task.Status, task.Phase, task.Progress, task.ErrorMsg, string(metadataJSON), now, now)
	return err
}

func (s *Store) UpdateTask(ctx context.Context, id string, status TaskStatus, phase string, progress float64, errorMsg string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		UPDATE frontend_tasks SET status = ?, phase = ?, progress = ?, error_msg = ?, updated_at = ?
		WHERE id = ?
	`, status, phase, progress, errorMsg, now, id)
	return err
}

func (s *Store) SaveMetadata(ctx context.Context, id string, metadata TaskMetadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `
		UPDATE frontend_tasks SET metadata = ?, updated_at = ?
		WHERE id = ?
	`, string(metadataJSON), now, id)
	return err
}

func (s *Store) GetMetadata(ctx context.Context, id string) (TaskMetadata, error) {
	var metadataStr string
	err := s.db.QueryRowContext(ctx, "SELECT metadata FROM frontend_tasks WHERE id = ?", id).Scan(&metadataStr)
	if err != nil {
		return nil, err
	}

	var metadata TaskMetadata
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return metadata, nil
}

func (s *Store) GetAllTasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, status, phase, progress, error_msg, metadata, created_at, updated_at
		FROM frontend_tasks ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTasks(rows)
}

func (s *Store) GetActiveTasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, status, phase, progress, error_msg, metadata, created_at, updated_at
		FROM frontend_tasks
		WHERE status IN (?, ?)
		ORDER BY created_at DESC
	`, StatusRunning, StatusPaused)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTasks(rows)
}

func (s *Store) scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var t Task
		var metadataStr string
		err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Status, &t.Phase, &t.Progress, &t.ErrorMsg, &metadataStr, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(metadataStr), &t.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task metadata: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
