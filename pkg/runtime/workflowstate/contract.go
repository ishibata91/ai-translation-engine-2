package workflowstate

import "context"

// Snapshot captures workflow-owned runtime state without embedding slice DTOs.
type Snapshot struct {
	WorkflowID   string
	TaskType     string
	Phase        string
	Progress     float64
	ResumeCursor int
	Status       string
}

// Store persists workflow runtime state snapshots.
type Store interface {
	Save(ctx context.Context, snapshot Snapshot) error
	Load(ctx context.Context, workflowID string) (Snapshot, error)
	ListActive(ctx context.Context) ([]Snapshot, error)
}
