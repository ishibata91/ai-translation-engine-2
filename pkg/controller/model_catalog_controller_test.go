package controller

import (
	"errors"
	"testing"

	modelcatalog "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/modelcatalog"
	modelcatalogcontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/modelcatalogcontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelCatalogController_API_TableDriven(t *testing.T) {
	serviceErr := errors.New("list failed")
	input := modelcatalog.ListModelsInput{Namespace: "master_persona.llm", Provider: "lmstudio", Endpoint: "http://localhost:1234"}

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *ModelCatalogController, env *modelcatalogcontrollertest.Env)
	}{
		{
			name: "ListModels returns service result",
			run: func(t *testing.T, controller *ModelCatalogController, env *modelcatalogcontrollertest.Env) {
				env.Service.Models = []modelcatalog.ModelOption{{ID: "m1", DisplayName: "Model-1"}}
				got, err := controller.ListModels(input)
				require.NoError(t, err)
				assert.Equal(t, env.Service.Models, got)
				assert.Equal(t, input, env.Service.LastInput)
				assert.Equal(t, env.TestEnv.Ctx, env.Service.LastCtx)
			},
		},
		{
			name: "ListModels returns service error",
			run: func(t *testing.T, controller *ModelCatalogController, env *modelcatalogcontrollertest.Env) {
				env.Service.Err = serviceErr
				_, err := controller.ListModels(input)
				require.Error(t, err)
				assert.ErrorIs(t, err, serviceErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := modelcatalogcontrollertest.Build(t, tc.name)
			controller := NewModelCatalogController(env.Service)
			controller.SetContext(env.TestEnv.Ctx)
			tc.run(t, controller, env)
		})
	}
}
