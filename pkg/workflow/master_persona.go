package workflow

import (
	"context"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
)

// StartMasterPersonaInput is the workflow entry DTO for master persona generation.
type StartMasterPersonaInput struct {
	SourceJSONPath    string `json:"source_json_path"`
	OverwriteExisting bool   `json:"overwrite_existing"`
}

// MasterPersona defines the controller-facing workflow contract for persona generation.
type MasterPersona interface {
	StartMasterPersona(ctx context.Context, input StartMasterPersonaInput) (string, error)
	ResumeMasterPersona(ctx context.Context, taskID string) error
	CancelMasterPersona(ctx context.Context, taskID string) error
	GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error)
	GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error)
}
