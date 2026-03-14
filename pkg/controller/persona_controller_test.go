package controller

import (
	"errors"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	personacontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/personacontroller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonaController_API_TableDriven(t *testing.T) {
	serviceErr := errors.New("service failed")

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *PersonaController, env *personacontrollertest.Env)
	}{
		{
			name: "ListNPCs returns service result",
			run: func(t *testing.T, controller *PersonaController, env *personacontrollertest.Env) {
				env.Service.NPCs = []persona.PersonaNPCView{{PersonaID: 1, NPCName: "Lydia"}}
				got, err := controller.ListNPCs()
				require.NoError(t, err)
				assert.Equal(t, env.Service.NPCs, got)
			},
		},
		{
			name: "ListNPCs returns service error",
			run: func(t *testing.T, controller *PersonaController, env *personacontrollertest.Env) {
				env.Service.NPCsErr = serviceErr
				_, err := controller.ListNPCs()
				require.Error(t, err)
				assert.ErrorIs(t, err, serviceErr)
			},
		},
		{
			name: "ListDialoguesByPersonaID returns service result",
			run: func(t *testing.T, controller *PersonaController, env *personacontrollertest.Env) {
				env.Service.Dialogues = []persona.PersonaDialogueView{{ID: 11, PersonaID: 3}}
				got, err := controller.ListDialoguesByPersonaID(3)
				require.NoError(t, err)
				assert.Equal(t, env.Service.Dialogues, got)
				assert.Equal(t, int64(3), env.Service.LastDialogueID)
			},
		},
		{
			name: "ListDialoguesByPersonaID returns service error",
			run: func(t *testing.T, controller *PersonaController, env *personacontrollertest.Env) {
				env.Service.DialoguesErr = serviceErr
				_, err := controller.ListDialoguesByPersonaID(3)
				require.Error(t, err)
				assert.ErrorIs(t, err, serviceErr)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := personacontrollertest.Build(t, tc.name)
			controller := NewPersonaController(env.Service)
			controller.SetContext(env.TestEnv.Ctx)
			tc.run(t, controller, env)
		})
	}
}

func TestPersonaController_ListDialoguesByPersonaID_ValidateInput(t *testing.T) {
	env := personacontrollertest.Build(t, "validation")
	controller := NewPersonaController(env.Service)

	_, err := controller.ListDialoguesByPersonaID(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persona_id is required")
}
