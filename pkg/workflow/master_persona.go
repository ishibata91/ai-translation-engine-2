package workflow

import (
	"context"
	"time"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
)

// StartMasterPersonaInput is the workflow entry DTO for master persona generation.
type StartMasterPersonaInput struct {
	SourceJSONPath    string `json:"source_json_path"`
	OverwriteExisting bool   `json:"overwrite_existing"`
}

// PersonaExecutionInput is the workflow-local contract for persona phase bootstrap/resume.
type PersonaExecutionInput struct {
	TaskID            string                   `json:"task_id"`
	SourceJSONPath    string                   `json:"source_json_path"`
	OverwriteExisting bool                     `json:"overwrite_existing"`
	Request           TranslationRequestConfig `json:"request"`
	Prompt            TranslationPromptConfig  `json:"prompt"`
}

// PersonaRuntimeEntry is one task-scoped runtime snapshot row for persona execution.
type PersonaRuntimeEntry struct {
	RequestID    string    `json:"request_id"`
	SourcePlugin string    `json:"source_plugin"`
	SpeakerID    string    `json:"speaker_id"`
	RequestState string    `json:"request_state"`
	ResumeCursor int       `json:"resume_cursor"`
	ErrorMessage string    `json:"error_message"`
	UpdatedAt    time.Time `json:"updated_at"`
	HasResponse  bool      `json:"has_response"`
	HasLookupKey bool      `json:"has_lookup_key"`
}

// MasterPersona defines the controller-facing workflow contract for persona generation.
type MasterPersona interface {
	StartMasterPersona(ctx context.Context, input StartMasterPersonaInput) (string, error)
	RunPersonaPhase(ctx context.Context, input PersonaExecutionInput) error
	ListPersonaRuntime(ctx context.Context, taskID string) ([]PersonaRuntimeEntry, error)
	ResumeMasterPersona(ctx context.Context, taskID string) error
	CancelMasterPersona(ctx context.Context, taskID string) error
	GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error)
	GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error)
}
