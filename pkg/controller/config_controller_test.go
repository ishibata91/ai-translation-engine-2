package controller

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	_ "modernc.org/sqlite"
)

type configTestContextKey string

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

func (s *recordingConfigStore) Watch(namespace string, key string, callback config.ChangeCallback) config.UnsubscribeFunc {
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

func setupConfigControllerTest(t *testing.T) (*sql.DB, *ConfigController) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	logger := slog.Default()
	store, err := config.NewSQLiteStore(context.Background(), db, logger)
	if err != nil {
		t.Fatalf("failed to init sqlite store: %v", err)
	}
	return db, NewConfigController(store, logger)
}

func TestConfigController_ConfigSetManyAndGetAll(t *testing.T) {
	db, controller := setupConfigControllerTest(t)
	defer db.Close()

	values := map[string]string{
		"provider":    "lmstudio",
		"model":       "llama-3",
		"endpoint":    "http://localhost:1234",
		"api_key":     "plain-text-key",
		"temperature": "0.3",
		"max_tokens":  "500",
	}
	if err := controller.ConfigSetMany("master_persona.llm", values); err != nil {
		t.Fatalf("ConfigSetMany failed: %v", err)
	}

	got, err := controller.ConfigGetAll("master_persona.llm")
	if err != nil {
		t.Fatalf("ConfigGetAll failed: %v", err)
	}
	if len(got) != len(values) {
		t.Fatalf("unexpected key count: got=%d want=%d", len(got), len(values))
	}
	for key, want := range values {
		if got[key] != want {
			t.Fatalf("unexpected value for %s: got=%s want=%s", key, got[key], want)
		}
	}
}

func TestConfigController_ConfigGet_MissingReturnsEmpty(t *testing.T) {
	db, controller := setupConfigControllerTest(t)
	defer db.Close()

	got, err := controller.ConfigGet("master_persona.llm", "provider")
	if err != nil {
		t.Fatalf("ConfigGet failed: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty for missing key, got=%s", got)
	}
}

func TestConfigController_ConfigGetAll_MasterPersonaPromptDefaults(t *testing.T) {
	db, controller := setupConfigControllerTest(t)
	defer db.Close()

	got, err := controller.ConfigGetAll(config.MasterPersonaPromptNamespace)
	if err != nil {
		t.Fatalf("ConfigGetAll failed: %v", err)
	}
	if got[config.MasterPersonaUserPromptKey] != config.DefaultMasterPersonaUserPrompt {
		t.Fatalf("unexpected default user prompt: %q", got[config.MasterPersonaUserPromptKey])
	}
	if got[config.MasterPersonaSystemPromptKey] != config.DefaultMasterPersonaSystemPrompt {
		t.Fatalf("unexpected default system prompt: %q", got[config.MasterPersonaSystemPromptKey])
	}
}

func TestConfigController_ConfigGet_MasterPersonaPromptDefault(t *testing.T) {
	db, controller := setupConfigControllerTest(t)
	defer db.Close()

	got, err := controller.ConfigGet(config.MasterPersonaPromptNamespace, config.MasterPersonaSystemPromptKey)
	if err != nil {
		t.Fatalf("ConfigGet failed: %v", err)
	}
	if got != config.DefaultMasterPersonaSystemPrompt {
		t.Fatalf("unexpected default system prompt: %q", got)
	}
}

func TestConfigController_ContextPropagationAndErrorWrap(t *testing.T) {
	baseCtx := context.WithValue(context.Background(), configTestContextKey("trace_id"), "trace-123")

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
				logger:       slog.Default(),
				uiStateStore: store,
				configStore:  store,
			}

			err := tc.run(controller)
			if err == nil {
				t.Fatalf("expected error")
			}
			if got := store.lastCtx.Value(configTestContextKey("trace_id")); got != "trace-123" {
				t.Fatalf("expected propagated context value, got=%v", got)
			}
			for _, substr := range tc.wantErrSubstrs {
				if !strings.Contains(err.Error(), substr) {
					t.Fatalf("expected error to contain %q, got=%q", substr, err.Error())
				}
			}
		})
	}
}
