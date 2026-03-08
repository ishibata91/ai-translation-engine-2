package queue

import base "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"

const (
	// StatusPending indicates queued work not yet dispatched.
	StatusPending = base.StatusPending
	// StatusInProgress indicates queued work currently being dispatched or saved.
	StatusInProgress = base.StatusInProgress
	// StatusCompleted indicates queued work completed successfully.
	StatusCompleted = base.StatusCompleted
	// StatusFailed indicates queued work failed.
	StatusFailed = base.StatusFailed
	// StatusCancelled indicates queued work was cancelled.
	StatusCancelled = base.StatusCancelled
)

const (
	// RequestStatePending indicates a request is queued for dispatch.
	RequestStatePending = base.RequestStatePending
	// RequestStateRunning indicates a request is being processed.
	RequestStateRunning = base.RequestStateRunning
	// RequestStateCompleted indicates a request finished successfully.
	RequestStateCompleted = base.RequestStateCompleted
	// RequestStateFailed indicates a request failed.
	RequestStateFailed = base.RequestStateFailed
	// RequestStateCanceled indicates a request was cancelled.
	RequestStateCanceled = base.RequestStateCanceled
)

// Queue aliases the concrete runtime queue implementation.
type Queue = base.Queue

// Worker aliases the concrete runtime queue worker implementation.
type Worker = base.Worker

// JobRequest aliases the persisted queue request DTO.
type JobRequest = base.JobRequest

// TaskRequestState aliases the persisted queue aggregate DTO.
type TaskRequestState = base.TaskRequestState

// ProcessHooks aliases worker lifecycle hooks used by workflow orchestration.
type ProcessHooks = base.ProcessHooks

// ProcessOptions aliases worker execution options used by workflow orchestration.
type ProcessOptions = base.ProcessOptions

// ConfigReadOptions aliases config resolution options for queue workers.
type ConfigReadOptions = base.ConfigReadOptions
