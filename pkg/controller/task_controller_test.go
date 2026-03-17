package controller

import (
	"context"
	"errors"
	"testing"

	taskcontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/taskcontroller"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	task "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskController_API_TableDriven(t *testing.T) {
	resumeErr := errors.New("resume failed")
	deleteErr := errors.New("delete failed")
	storeErr := errors.New("store failed")

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env)
	}{
		{
			name: "GetActiveTasks returns manager tasks",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				env.Manager.ActiveTasks = []task.Task{{ID: "a1"}}
				got := controller.GetActiveTasks()
				assert.Equal(t, env.Manager.ActiveTasks, got)
			},
		},
		{
			name: "GetAllTasks returns store tasks with context",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				env.Store.Tasks = []task.Task{{ID: "all1"}}
				got, err := controller.GetAllTasks()
				require.NoError(t, err)
				assert.Equal(t, env.Store.Tasks, got)
				assert.Equal(t, env.TestEnv.Ctx, env.Store.LastCtx)
			},
		},
		{
			name: "GetAllTasks returns store error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				env.Store.Err = storeErr
				_, err := controller.GetAllTasks()
				require.Error(t, err)
				assert.ErrorIs(t, err, storeErr)
			},
		},
		{
			name: "ResumeTask delegates task id",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				err := controller.ResumeTask("r1")
				require.NoError(t, err)
				assert.Equal(t, "r1", env.Manager.ResumeTaskID)
			},
		},
		{
			name: "ResumeTask returns manager error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				env.Manager.ResumeErr = resumeErr
				err := controller.ResumeTask("r2")
				require.Error(t, err)
				assert.ErrorIs(t, err, resumeErr)
			},
		},
		{
			name: "DeleteTask delegates task id with context",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				err := controller.DeleteTask("d1")
				require.NoError(t, err)
				assert.Equal(t, "d1", env.Manager.DeleteTaskID)
				assert.Equal(t, env.TestEnv.Ctx, env.Manager.DeleteTaskCtx)
			},
		},
		{
			name: "DeleteTask returns manager error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				env.Manager.DeleteErr = deleteErr
				err := controller.DeleteTask("d2")
				require.Error(t, err)
				assert.ErrorIs(t, err, deleteErr)
			},
		},
		{
			name: "CancelTask delegates task id",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env) {
				controller.CancelTask("c1")
				assert.Equal(t, "c1", env.Manager.CancelTaskID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := taskcontrollertest.Build(t, tc.name)
			controller := NewTaskController(env.Manager)
			controller.SetContext(env.TestEnv.Ctx)
			tc.run(t, controller, env)
		})
	}
}

func TestTaskController_TranslationFlowAPI_TableDriven(t *testing.T) {
	workflowErr := errors.New("workflow failed")
	ensureErr := errors.New("ensure failed")
	loadResult := workflow.TranslationLoadResult{TaskID: "task-1", Files: []workflow.TranslationLoadedFile{{FileID: 10}}}
	previewResult := workflow.TranslationPreviewPage{FileID: 10, Page: 2, PageSize: 50, TotalRows: 99}

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow)
	}{
		{
			name: "LoadTranslationFlowFiles resolves task id and returns result",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				env.Manager.EnsureTaskResolvedID = "task-resolved"
				wf.loadResult = loadResult
				got, err := controller.LoadTranslationFlowFiles("task-1", []string{"a.json", "b.json"})
				require.NoError(t, err)
				assert.Equal(t, loadResult, got)
				assert.Equal(t, "task-1", env.Manager.EnsureTaskInput)
				assert.Equal(t, "task-resolved", wf.lastLoadInput.TaskID)
				assert.Equal(t, []string{"a.json", "b.json"}, wf.lastLoadInput.FilePaths)
			},
		},
		{
			name: "LoadTranslationFlowFiles returns ensure error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				env.Manager.EnsureTaskErr = ensureErr
				_, err := controller.LoadTranslationFlowFiles("task-1", []string{"a.json"})
				require.Error(t, err)
				assert.ErrorIs(t, err, ensureErr)
				assert.Equal(t, workflow.LoadTranslationFlowInput{}, wf.lastLoadInput)
			},
		},
		{
			name: "LoadTranslationFlowFiles returns workflow error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				wf.loadErr = workflowErr
				_, err := controller.LoadTranslationFlowFiles("task-1", []string{"a.json"})
				require.Error(t, err)
				assert.Equal(t, "task-1", env.Manager.EnsureTaskInput)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "ListLoadedTranslationFlowFiles resolves task id",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				env.Manager.EnsureTaskResolvedID = "task-resolved"
				wf.listResult = loadResult
				got, err := controller.ListLoadedTranslationFlowFiles("task-2")
				require.NoError(t, err)
				assert.Equal(t, loadResult, got)
				assert.Equal(t, "task-2", env.Manager.EnsureTaskInput)
				assert.Equal(t, "task-resolved", wf.lastListTaskID)
			},
		},
		{
			name: "ListLoadedTranslationFlowFiles returns ensure error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				env.Manager.EnsureTaskErr = ensureErr
				_, err := controller.ListLoadedTranslationFlowFiles("task-3")
				require.Error(t, err)
				assert.ErrorIs(t, err, ensureErr)
				assert.Equal(t, "", wf.lastListTaskID)
			},
		},
		{
			name: "ListLoadedTranslationFlowFiles returns workflow error",
			run: func(t *testing.T, controller *TaskController, env *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				wf.listErr = workflowErr
				_, err := controller.ListLoadedTranslationFlowFiles("task-3")
				require.Error(t, err)
				assert.Equal(t, "task-3", env.Manager.EnsureTaskInput)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "ListTranslationFlowPreviewRows delegates arguments",
			run: func(t *testing.T, controller *TaskController, _ *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				wf.previewResult = previewResult
				got, err := controller.ListTranslationFlowPreviewRows(10, 2, 50)
				require.NoError(t, err)
				assert.Equal(t, previewResult, got)
				assert.Equal(t, int64(10), wf.lastPreviewFileID)
				assert.Equal(t, 2, wf.lastPreviewPage)
				assert.Equal(t, 50, wf.lastPreviewPageSize)
			},
		},
		{
			name: "ListTranslationFlowPreviewRows returns workflow error",
			run: func(t *testing.T, controller *TaskController, _ *taskcontrollertest.Env, wf *fakeTranslationFlowWorkflow) {
				wf.previewErr = workflowErr
				_, err := controller.ListTranslationFlowPreviewRows(99, 1, 50)
				require.Error(t, err)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := taskcontrollertest.Build(t, tc.name)
			wf := &fakeTranslationFlowWorkflow{}
			controller := NewTaskController(env.Manager)
			controller.SetContext(env.TestEnv.Ctx)
			controller.SetTranslationFlowWorkflow(wf)
			tc.run(t, controller, env, wf)
			if wf.lastCtx != nil {
				assert.Equal(t, env.TestEnv.Ctx, wf.lastCtx)
			}
		})
	}
}

func TestTaskController_TranslationFlowAPI_NilWorkflowGuard(t *testing.T) {
	env := taskcontrollertest.Build(t, "nil translation workflow")
	controller := NewTaskController(env.Manager)
	controller.SetContext(env.TestEnv.Ctx)

	_, err := controller.LoadTranslationFlowFiles("task-1", []string{"a.json"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = controller.ListLoadedTranslationFlowFiles("task-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = controller.ListTranslationFlowPreviewRows(1, 1, 50)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

type fakeTranslationFlowWorkflow struct {
	lastCtx              context.Context
	lastLoadInput        workflow.LoadTranslationFlowInput
	lastListTaskID       string
	lastPreviewFileID    int64
	lastPreviewPage      int
	lastPreviewPageSize  int
	lastTerminologyInput workflow.RunTerminologyPhaseInput

	loadResult        workflow.TranslationLoadResult
	loadErr           error
	listResult        workflow.TranslationLoadResult
	listErr           error
	previewResult     workflow.TranslationPreviewPage
	previewErr        error
	terminologyResult workflow.TerminologyPhaseResult
	terminologyErr    error
}

func (f *fakeTranslationFlowWorkflow) LoadFiles(ctx context.Context, input workflow.LoadTranslationFlowInput) (workflow.TranslationLoadResult, error) {
	f.lastCtx = ctx
	f.lastLoadInput = input
	return f.loadResult, f.loadErr
}

func (f *fakeTranslationFlowWorkflow) ListFiles(ctx context.Context, taskID string) (workflow.TranslationLoadResult, error) {
	f.lastCtx = ctx
	f.lastListTaskID = taskID
	return f.listResult, f.listErr
}

func (f *fakeTranslationFlowWorkflow) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (workflow.TranslationPreviewPage, error) {
	f.lastCtx = ctx
	f.lastPreviewFileID = fileID
	f.lastPreviewPage = page
	f.lastPreviewPageSize = pageSize
	return f.previewResult, f.previewErr
}

func (f *fakeTranslationFlowWorkflow) RunTerminologyPhase(ctx context.Context, input workflow.RunTerminologyPhaseInput) (workflow.TerminologyPhaseResult, error) {
	f.lastCtx = ctx
	f.lastTerminologyInput = input
	return f.terminologyResult, f.terminologyErr
}

func (f *fakeTranslationFlowWorkflow) GetTerminologyPhase(ctx context.Context, taskID string) (workflow.TerminologyPhaseResult, error) {
	f.lastCtx = ctx
	f.lastListTaskID = taskID
	return f.terminologyResult, f.terminologyErr
}
