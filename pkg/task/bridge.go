package task

import (
	"context"
	"errors"
	"log/slog"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
)

type masterPersonaWorkflow interface {
	StartMasterPersonTask(ctx context.Context, input StartMasterPersonTaskInput) (string, error)
	ResumeMasterPersonaTask(ctx context.Context, taskID string) error
	CancelTask(ctx context.Context, taskID string)
	GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error)
	GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error)
}

type Bridge struct {
	manager               *Manager
	logger                *slog.Logger
	masterPersonaWorkflow masterPersonaWorkflow
}

func NewBridge(
	manager *Manager,
	logger *slog.Logger,
) *Bridge {
	return &Bridge{
		manager: manager,
		logger:  logger.With("module", "task_bridge"),
	}
}

func NewMasterPersonaBridge(
	manager *Manager,
	logger *slog.Logger,
	masterPersona masterPersonaWorkflow,
) *Bridge {
	bridge := NewBridge(manager, logger)
	bridge.masterPersonaWorkflow = masterPersona
	return bridge
}

func (b *Bridge) GetActiveTasks() []Task {
	return b.manager.GetActiveTasks()
}

func (b *Bridge) GetAllTasks() ([]Task, error) {
	return b.manager.store.GetAllTasks(context.Background())
}

func (b *Bridge) ResumeTask(taskID string) error {
	return b.manager.ResumeTask(taskID)
}

func (b *Bridge) ResumeMasterPersonaTask(ctx context.Context, taskID string) error {
	if b.masterPersonaWorkflow == nil {
		return errors.New("master persona workflow is not configured")
	}
	return b.masterPersonaWorkflow.ResumeMasterPersonaTask(ctx, taskID)
}

func (b *Bridge) CancelTask(ctx context.Context, taskID string) {
	if b.masterPersonaWorkflow != nil {
		b.masterPersonaWorkflow.CancelTask(ctx, taskID)
		return
	}
	b.manager.CancelTask(taskID)
}

func (b *Bridge) GetTaskRequestState(taskID string) (runtimequeue.TaskRequestState, error) {
	if b.masterPersonaWorkflow == nil {
		return runtimequeue.TaskRequestState{}, errors.New("master persona workflow is not configured")
	}
	return b.masterPersonaWorkflow.GetTaskRequestState(context.Background(), taskID)
}

func (b *Bridge) GetTaskRequests(taskID string) ([]runtimequeue.JobRequest, error) {
	if b.masterPersonaWorkflow == nil {
		return nil, errors.New("master persona workflow is not configured")
	}
	return b.masterPersonaWorkflow.GetTaskRequests(context.Background(), taskID)
}
