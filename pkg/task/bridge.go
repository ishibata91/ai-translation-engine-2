package task

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
)

type Bridge struct {
	manager          *Manager
	logger           *slog.Logger
	parser           parser.Parser
	personaGenerator persona.NPCPersonaGenerator
	notifier         progress.ProgressNotifier
	queue            *queue.Queue
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
	parser parser.Parser,
	personaGenerator persona.NPCPersonaGenerator,
	notifier progress.ProgressNotifier,
	requestQueue *queue.Queue,
) *Bridge {
	bridge := NewBridge(manager, logger)
	bridge.parser = parser
	bridge.personaGenerator = personaGenerator
	bridge.notifier = notifier
	bridge.queue = requestQueue
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

func (b *Bridge) CancelTask(taskID string) {
	b.manager.CancelTask(taskID)
}

func (b *Bridge) GetTaskRequestState(taskID string) (queue.TaskRequestState, error) {
	if b.queue == nil {
		return queue.TaskRequestState{}, errors.New("request queue is not configured")
	}
	return b.queue.GetTaskRequestState(context.Background(), taskID)
}

func (b *Bridge) GetTaskRequests(taskID string) ([]queue.JobRequest, error) {
	if b.queue == nil {
		return nil, errors.New("request queue is not configured")
	}
	return b.queue.GetTaskRequests(context.Background(), taskID)
}

func (b *Bridge) reportProgress(ctx context.Context, correlationID string, completed int, status string, message string) {
	if b.notifier == nil {
		return
	}
	b.notifier.OnProgress(ctx, progress.ProgressEvent{
		CorrelationID: correlationID,
		Total:         100,
		Completed:     completed,
		Status:        status,
		Message:       message,
	})
}
