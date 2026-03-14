package controller

import (
	"errors"
	"testing"

	taskcontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/taskcontroller"
	task "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskController_API_TableDriven(t *testing.T) {
	resumeErr := errors.New("resume failed")
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
