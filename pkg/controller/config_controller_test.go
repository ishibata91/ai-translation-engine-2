package controller

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/configstore"
	configcontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/configcontroller"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
	workflowpersona "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/persona"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingConfigStore struct {
	lastCtx       context.Context
	getErr        error
	setErr        error
	deleteErr     error
	getAllErr     error
	uiStateGetErr error
	uiStateSetErr error
	uiStateDelErr error
}

func (s *recordingConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	s.lastCtx = ctx
	if s.uiStateGetErr != nil {
		return "", s.uiStateGetErr
	}
	if s.getErr != nil {
		return "", s.getErr
	}
	return "", nil
}

func (s *recordingConfigStore) Set(ctx context.Context, namespace string, key string, value string) error {
	s.lastCtx = ctx
	return s.setErr
}

func (s *recordingConfigStore) Delete(ctx context.Context, namespace string, key string) error {
	s.lastCtx = ctx
	if s.uiStateDelErr != nil {
		return s.uiStateDelErr
	}
	return s.deleteErr
}

func (s *recordingConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	s.lastCtx = ctx
	if s.getAllErr != nil {
		return nil, s.getAllErr
	}
	return map[string]string{}, nil
}

func (s *recordingConfigStore) Watch(namespace string, key string, callback configstore.ChangeCallback) configstore.UnsubscribeFunc {
	return func() {}
}

func (s *recordingConfigStore) SetJSON(ctx context.Context, namespace string, key string, value any) error {
	s.lastCtx = ctx
	return s.uiStateSetErr
}

func (s *recordingConfigStore) GetJSON(ctx context.Context, namespace string, key string, target any) error {
	s.lastCtx = ctx
	return nil
}

func TestConfigController_API_TableDriven(t *testing.T) {
	testCases := []struct {
		name string
		run  func(t *testing.T, controller *ConfigController)
	}{
		{
			name: "ConfigSetMany and ConfigGetAll",
			run: func(t *testing.T, controller *ConfigController) {
				values := map[string]string{
					"provider":    "lmstudio",
					"model":       "llama-3",
					"endpoint":    "http://localhost:1234",
					"api_key":     "plain-text-key",
					"temperature": "0.3",
					"max_tokens":  "500",
				}

				err := controller.ConfigSetMany("master_persona.llm", values)
				require.NoError(t, err)

				got, err := controller.ConfigGetAll("master_persona.llm")
				require.NoError(t, err)
				require.Len(t, got, len(values))
				assert.Equal(t, values, got)
			},
		},
		{
			name: "ConfigGet missing key returns empty",
			run: func(t *testing.T, controller *ConfigController) {
				got, err := controller.ConfigGet("master_persona.llm", "provider")
				require.NoError(t, err)
				assert.Empty(t, got)
			},
		},
		{
			name: "ConfigGetAll returns master persona prompt defaults",
			run: func(t *testing.T, controller *ConfigController) {
				got, err := controller.ConfigGetAll(workflowpersona.MasterPersonaPromptNamespace)
				require.NoError(t, err)
				assert.Equal(t, workflowpersona.DefaultMasterPersonaUserPrompt, got[workflowpersona.MasterPersonaUserPromptKey])
				assert.Equal(t, workflowpersona.DefaultMasterPersonaSystemPrompt, got[workflowpersona.MasterPersonaSystemPromptKey])
			},
		},
		{
			name: "ConfigGet returns master persona system prompt default",
			run: func(t *testing.T, controller *ConfigController) {
				got, err := controller.ConfigGet(
					workflowpersona.MasterPersonaPromptNamespace,
					workflowpersona.MasterPersonaSystemPromptKey,
				)
				require.NoError(t, err)
				assert.Equal(t, workflowpersona.DefaultMasterPersonaSystemPrompt, got)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := configcontrollertest.Build(t, tc.name)
			controller := NewConfigController(env.Store, env.TestEnv.Logger)
			controller.SetContext(env.TestEnv.Ctx)
			tc.run(t, controller)
		})
	}
}

func TestConfigController_APITestEnv_UsesTmpDBPath(t *testing.T) {
	env := configcontrollertest.Build(t, "db-path")
	require.NotEmpty(t, env.TestEnv.DBPath)
	assert.Contains(t, filepath.ToSlash(env.TestEnv.DBPath), "tmp/api_test_db/")
}

func TestConfigController_ContextPropagationAndErrorWrap(t *testing.T) {
	baseCtx := testenv.NewTraceContext("trace-123")

	testCases := []struct {
		name           string
		run            func(controller *ConfigController) error
		setupStore     func(store *recordingConfigStore)
		wantErrSubstrs []string
	}{
		{
			name: "UIStateGetJSON propagates context and wraps error",
			setupStore: func(store *recordingConfigStore) {
				store.uiStateGetErr = errors.New("db unavailable")
			},
			run: func(controller *ConfigController) error {
				_, err := controller.UIStateGetJSON("ui", "layout")
				return err
			},
			wantErrSubstrs: []string{"namespace=ui", "key=layout", "db unavailable"},
		},
		{
			name: "ConfigSetMany propagates context and wraps key",
			setupStore: func(store *recordingConfigStore) {
				store.setErr = errors.New("write failed")
			},
			run: func(controller *ConfigController) error {
				return controller.ConfigSetMany("master_persona.llm", map[string]string{"model": "test"})
			},
			wantErrSubstrs: []string{"namespace=master_persona.llm", "key=model", "write failed"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingConfigStore{}
			if tc.setupStore != nil {
				tc.setupStore(store)
			}

			controller := &ConfigController{
				ctx:          baseCtx,
				logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
				uiStateStore: store,
				configStore:  store,
			}

			err := tc.run(controller)
			require.Error(t, err)
			assert.Equal(t, "trace-123", testenv.TraceIDValue(store.lastCtx))
			for _, substr := range tc.wantErrSubstrs {
				assert.True(t, strings.Contains(err.Error(), substr), "expected wrapped error to contain %q: %s", substr, err.Error())
			}
		})
	}
}
