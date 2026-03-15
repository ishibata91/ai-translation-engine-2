package taskcontroller

import (
	"context"
	"fmt"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
	task "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

// Env bundles task controller test dependencies.
type Env struct {
	Manager *FakeManager
	Store   *FakeStore
	TestEnv *testenv.Env
}

// FakeStore stubs task store behavior.
type FakeStore struct {
	LastCtx context.Context
	Tasks   []task.Task
	Err     error
}

func (s *FakeStore) GetAllTasks(ctx context.Context) ([]task.Task, error) {
	s.LastCtx = ctx
	return s.Tasks, s.Err
}

// FakeManager stubs task manager behavior.
type FakeManager struct {
	StoreRef             *FakeStore
	ActiveTasks          []task.Task
	ResumeTaskID         string
	ResumeErr            error
	DeleteTaskID         string
	DeleteTaskCtx        context.Context
	DeleteErr            error
	CancelTaskID         string
	EnsureTaskInput      string
	EnsureTaskResolvedID string
	EnsureTaskErr        error
}

func (m *FakeManager) GetActiveTasks() []task.Task {
	return m.ActiveTasks
}

func (m *FakeManager) Store() *FakeStore {
	return m.StoreRef
}

func (m *FakeManager) GetAllTasks(ctx context.Context) ([]task.Task, error) {
	return m.StoreRef.GetAllTasks(ctx)
}

func (m *FakeManager) ResumeTask(taskID string) error {
	m.ResumeTaskID = taskID
	return m.ResumeErr
}

func (m *FakeManager) DeleteTask(ctx context.Context, taskID string) error {
	m.DeleteTaskCtx = ctx
	m.DeleteTaskID = taskID
	return m.DeleteErr
}

func (m *FakeManager) CancelTask(taskID string) {
	m.CancelTaskID = taskID
}

func (m *FakeManager) EnsureTranslationProjectTask(_ context.Context, taskID string) (string, error) {
	m.EnsureTaskInput = taskID
	if m.EnsureTaskErr != nil {
		return "", m.EnsureTaskErr
	}
	if m.EnsureTaskResolvedID != "" {
		return m.EnsureTaskResolvedID, nil
	}
	return taskID, nil
}

// Build creates task controller dependencies on shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()
	base := testenv.NewFileSQLiteEnv(t, name)
	store := &FakeStore{}
	return &Env{Manager: &FakeManager{StoreRef: store}, Store: store, TestEnv: base}
}

// String returns a short summary useful in failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil taskcontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, testenv.TraceIDValue(e.TestEnv.Ctx))
}
