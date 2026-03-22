package controller

import (
	"context"
	"errors"
	"testing"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	personataskcontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/personataskcontroller"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	task "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonaTaskController_API_TableDriven(t *testing.T) {
	workflowErr := errors.New("workflow failed")

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env)
	}{
		{
			name: "GetAllTasks returns manager tasks",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				expected := []task.Task{{ID: "t1"}}
				env.Manager.StoreRef.Tasks = expected

				got, err := controller.GetAllTasks()
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.Equal(t, env.TestEnv.Ctx, env.Manager.StoreRef.LastCtx)
			},
		},
		{
			name: "StartMasterPersonTask maps input and returns task id",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.StartTaskID = "task-123"
				id, err := controller.StartMasterPersonTask(task.StartMasterPersonTaskInput{
					SourceJSONPath:    "input.json",
					OverwriteExisting: true,
				})
				require.NoError(t, err)
				assert.Equal(t, "task-123", id)
				assert.Equal(t, "input.json", env.Workflow.StartInput.SourceJSONPath)
				assert.True(t, env.Workflow.StartInput.OverwriteExisting)
				assert.Equal(t, env.TestEnv.Ctx, env.Workflow.LastCtx)
			},
		},
		{
			name: "StartMasterPersonTask returns workflow error",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.StartErr = workflowErr
				_, err := controller.StartMasterPersonTask(task.StartMasterPersonTaskInput{})
				require.Error(t, err)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "ResumeTask delegates task id",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				err := controller.ResumeTask("task-9")
				require.NoError(t, err)
				assert.Equal(t, "task-9", env.Workflow.ResumeTaskID)
				assert.Equal(t, env.TestEnv.Ctx, env.Workflow.LastCtx)
			},
		},
		{
			name: "ResumeTask returns workflow error",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.ResumeErr = workflowErr
				err := controller.ResumeTask("task-9")
				require.Error(t, err)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "ResumeMasterPersonaTask delegates to resume",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				err := controller.ResumeMasterPersonaTask("task-10")
				require.NoError(t, err)
				assert.Equal(t, "task-10", env.Workflow.ResumeTaskID)
			},
		},
		{
			name: "GetTaskRequestState returns workflow state",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.State = runtimequeue.TaskRequestState{TaskID: "task-2", Total: 4}
				got, err := controller.GetTaskRequestState("task-2")
				require.NoError(t, err)
				assert.Equal(t, env.Workflow.State, got)
				assert.Equal(t, "task-2", env.Workflow.StateTaskID)
			},
		},
		{
			name: "GetTaskRequestState returns workflow error",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.StateErr = workflowErr
				_, err := controller.GetTaskRequestState("task-2")
				require.Error(t, err)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "GetTaskRequests returns workflow requests",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.Requests = []runtimequeue.JobRequest{{ID: "r1"}}
				got, err := controller.GetTaskRequests("task-3")
				require.NoError(t, err)
				assert.Equal(t, env.Workflow.Requests, got)
				assert.Equal(t, "task-3", env.Workflow.RequestsTaskID)
			},
		},
		{
			name: "GetTaskRequests returns workflow error",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				env.Workflow.RequestsErr = workflowErr
				_, err := controller.GetTaskRequests("task-3")
				require.Error(t, err)
				assert.ErrorIs(t, err, workflowErr)
			},
		},
		{
			name: "CancelTask delegates when workflow configured",
			run: func(t *testing.T, controller *PersonaTaskController, env *personataskcontrollertest.Env) {
				controller.CancelTask("task-4")
				assert.Equal(t, "task-4", env.Workflow.CancelTaskID)
				assert.Equal(t, env.TestEnv.Ctx, env.Workflow.LastCtx)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := personataskcontrollertest.Build(t, tc.name)
			controller := NewPersonaTaskController(env.Manager, &masterPersonaWorkflowAdapter{FakeWorkflow: env.Workflow})
			controller.SetContext(env.TestEnv.Ctx)
			tc.run(t, controller, env)
		})
	}
}

func TestPersonaTaskController_NilWorkflowGuards(t *testing.T) {
	env := personataskcontrollertest.Build(t, "nil-workflow")
	controller := NewPersonaTaskController(env.Manager, nil)
	controller.SetContext(env.TestEnv.Ctx)

	_, err := controller.StartMasterPersonTask(task.StartMasterPersonTaskInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	err = controller.ResumeTask("task-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = controller.GetTaskRequestState("task-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = controller.GetTaskRequests("task-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	assert.NotPanics(t, func() { controller.CancelTask("task-1") })
}

type masterPersonaWorkflowAdapter struct {
	*personataskcontrollertest.FakeWorkflow
}

func (a *masterPersonaWorkflowAdapter) RunPersonaPhase(context.Context, workflow.PersonaExecutionInput) error {
	return nil
}

func (a *masterPersonaWorkflowAdapter) ListPersonaRuntime(context.Context, string) ([]workflow.PersonaRuntimeEntry, error) {
	return nil, nil
}
