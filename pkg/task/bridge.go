package task

import (
	"context"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
)

type Bridge struct {
	manager          *Manager
	logger           *slog.Logger
	parser           parser.Parser
	personaGenerator persona.NPCPersonaGenerator
	notifier         progress.ProgressNotifier
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
) *Bridge {
	bridge := NewBridge(manager, logger)
	bridge.parser = parser
	bridge.personaGenerator = personaGenerator
	bridge.notifier = notifier
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
