package personataskcontroller

import (
	"context"
	"fmt"
	"testing"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	task "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

// Env bundles persona task controller test dependencies.
type Env struct {
	Manager  *FakeManager
	Workflow *FakeWorkflow
	TestEnv  *testenv.Env
}

// FakeManagerStore stubs task store APIs required by controller.
type FakeManagerStore struct {
	LastCtx context.Context
	Tasks   []task.Task
	Err     error
}

func (s *FakeManagerStore) GetAllTasks(ctx context.Context) ([]task.Task, error) {
	s.LastCtx = ctx
	return s.Tasks, s.Err
}

// FakeManager stubs task manager APIs required by controller.
type FakeManager struct {
	StoreRef *FakeManagerStore
}

func (m *FakeManager) Store() *FakeManagerStore {
	return m.StoreRef
}

func (m *FakeManager) GetAllTasks(ctx context.Context) ([]task.Task, error) {
	return m.StoreRef.GetAllTasks(ctx)
}

// FakeWorkflow stubs workflow.MasterPersona APIs required by controller.
type FakeWorkflow struct {
	LastCtx context.Context

	StartTaskID string
	StartErr    error
	StartInput  workflow.StartMasterPersonaInput

	ResumeTaskID string
	ResumeErr    error

	CancelTaskID string
	CancelErr    error

	StateTaskID string
	State       runtimequeue.TaskRequestState
	StateErr    error

	RequestsTaskID string
	Requests       []runtimequeue.JobRequest
	RequestsErr    error
}

func (w *FakeWorkflow) StartMasterPersona(ctx context.Context, input workflow.StartMasterPersonaInput) (string, error) {
	w.LastCtx = ctx
	w.StartInput = input
	if w.StartTaskID == "" {
		w.StartTaskID = "task-1"
	}
	return w.StartTaskID, w.StartErr
}

func (w *FakeWorkflow) ResumeMasterPersona(ctx context.Context, taskID string) error {
	w.LastCtx = ctx
	w.ResumeTaskID = taskID
	return w.ResumeErr
}

func (w *FakeWorkflow) CancelMasterPersona(ctx context.Context, taskID string) error {
	w.LastCtx = ctx
	w.CancelTaskID = taskID
	return w.CancelErr
}

func (w *FakeWorkflow) GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error) {
	w.LastCtx = ctx
	w.StateTaskID = taskID
	return w.State, w.StateErr
}

func (w *FakeWorkflow) GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error) {
	w.LastCtx = ctx
	w.RequestsTaskID = taskID
	return w.Requests, w.RequestsErr
}

// Build creates persona task controller dependencies on shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()
	base := testenv.NewFileSQLiteEnv(t, name)

	store := &FakeManagerStore{}
	return &Env{
		Manager:  &FakeManager{StoreRef: store},
		Workflow: &FakeWorkflow{},
		TestEnv:  base,
	}
}

// String returns a short summary useful in failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil personataskcontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, testenv.TraceIDValue(e.TestEnv.Ctx))
}
