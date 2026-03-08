package controller

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/task"
)

// TaskController exposes generic Wails-facing task operations.
type TaskController struct {
	manager *task.Manager
}

// NewTaskController constructs the task controller adapter.
func NewTaskController(manager *task.Manager) *TaskController {
	return &TaskController{
		manager: manager,
	}
}

// GetActiveTasks returns in-memory active tasks for dashboard polling.
func (c *TaskController) GetActiveTasks() []task.Task {
	return c.manager.GetActiveTasks()
}

// GetAllTasks loads all persisted tasks.
func (c *TaskController) GetAllTasks() ([]task.Task, error) {
	return c.manager.Store().GetAllTasks(context.Background())
}

// ResumeTask resumes a generic task through task manager.
func (c *TaskController) ResumeTask(taskID string) error {
	return c.manager.ResumeTask(taskID)
}

// CancelTask cancels a generic task through task manager.
func (c *TaskController) CancelTask(taskID string) {
	c.manager.CancelTask(taskID)
}
