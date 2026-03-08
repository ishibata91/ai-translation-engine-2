package controller

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/task"
)

// TaskController exposes generic Wails-facing task operations.
type TaskController struct {
	ctx     context.Context
	manager *task.Manager
}

// NewTaskController constructs the task controller adapter.
func NewTaskController(manager *task.Manager) *TaskController {
	return &TaskController{
		ctx:     context.Background(),
		manager: manager,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *TaskController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// GetActiveTasks returns in-memory active tasks for dashboard polling.
func (c *TaskController) GetActiveTasks() []task.Task {
	return c.manager.GetActiveTasks()
}

// GetAllTasks loads all persisted tasks.
func (c *TaskController) GetAllTasks() ([]task.Task, error) {
	return c.manager.Store().GetAllTasks(c.ctx)
}

// ResumeTask resumes a generic task through task manager.
func (c *TaskController) ResumeTask(taskID string) error {
	return c.manager.ResumeTask(taskID)
}

// CancelTask cancels a generic task through task manager.
func (c *TaskController) CancelTask(taskID string) {
	c.manager.CancelTask(taskID)
}
