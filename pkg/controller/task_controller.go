package controller

import (
	"context"

	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

type taskManager interface {
	GetActiveTasks() []task2.Task
	GetAllTasks(ctx context.Context) ([]task2.Task, error)
	ResumeTask(taskID string) error
	CancelTask(taskID string)
}

// TaskController exposes generic Wails-facing task operations.
type TaskController struct {
	ctx     context.Context
	manager taskManager
}

// NewTaskController constructs the task controller adapter.
func NewTaskController(manager taskManager) *TaskController {
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
func (c *TaskController) GetActiveTasks() []task2.Task {
	return c.manager.GetActiveTasks()
}

// GetAllTasks loads all persisted tasks.
func (c *TaskController) GetAllTasks() ([]task2.Task, error) {
	return c.manager.GetAllTasks(c.ctx)
}

// ResumeTask resumes a generic task through task manager.
func (c *TaskController) ResumeTask(taskID string) error {
	return c.manager.ResumeTask(taskID)
}

// CancelTask cancels a generic task through task manager.
func (c *TaskController) CancelTask(taskID string) {
	c.manager.CancelTask(taskID)
}
