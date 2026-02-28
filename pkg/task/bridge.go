package task

import (
	"context"
)

type Bridge struct {
	manager *Manager
}

func NewBridge(manager *Manager) *Bridge {
	return &Bridge{manager: manager}
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
