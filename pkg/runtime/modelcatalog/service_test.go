package modelcatalog

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/configstore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
)

type mockConfigStore struct {
	values map[string]map[string]string
}

func (m *mockConfigStore) Get(ctx context.Context, namespace, key string) (string, error) {
	if ns, ok := m.values[namespace]; ok {
		return ns[key], nil
	}
	return "", nil
}
func (m *mockConfigStore) Set(ctx context.Context, namespace, key, value string) error { return nil }
func (m *mockConfigStore) Delete(ctx context.Context, namespace, key string) error     { return nil }
func (m *mockConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	return nil, nil
}
func (m *mockConfigStore) Watch(namespace, key string, callback configstore.ChangeCallback) configstore.UnsubscribeFunc {
	return func() {}
}

type mockSecretStore struct {
	values map[string]map[string]string
}

func (m *mockSecretStore) GetSecret(ctx context.Context, namespace, key string) (string, error) {
	if ns, ok := m.values[namespace]; ok {
		return ns[key], nil
	}
	return "", nil
}
func (m *mockSecretStore) SetSecret(ctx context.Context, namespace, key, value string) error {
	return nil
}
func (m *mockSecretStore) DeleteSecret(ctx context.Context, namespace, key string) error { return nil }
func (m *mockSecretStore) ListSecretKeys(ctx context.Context, namespace string) ([]string, error) {
	return nil, nil
}

type mockLLMClient struct {
	models []llm.ModelInfo
}

func (m *mockLLMClient) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	return m.models, nil
}
func (m *mockLLMClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, nil
}
func (m *mockLLMClient) GenerateStructured(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, nil
}
func (m *mockLLMClient) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	return nil, nil
}
func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *mockLLMClient) HealthCheck(ctx context.Context) error { return nil }

type mockLLMManager struct {
	client     llm.LLMClient
	lastConfig llm.LLMConfig
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	m.lastConfig = config
	return m.client, nil
}
func (m *mockLLMManager) GetBatchClient(ctx context.Context, config llm.LLMConfig) (llm.BatchClient, error) {
	return nil, nil
}
func (m *mockLLMManager) ResolveBulkStrategy(ctx context.Context, strategy llm.BulkStrategy, provider string) llm.BulkStrategy {
	return strategy
}

func TestModelCatalogService_ListModels_UsesNamespaceConfig(t *testing.T) {
	manager := &mockLLMManager{
		client: &mockLLMClient{
			models: []llm.ModelInfo{{ID: "model-a", DisplayName: "Model A", Loaded: true, SupportsBatch: true}},
		},
	}
	service := NewModelCatalogService(
		&mockConfigStore{values: map[string]map[string]string{
			"master_persona.llm": {
				"provider": "lmstudio",
				"model":    "stored-model",
				"endpoint": "http://localhost:1234",
			},
		}},
		&mockSecretStore{},
		manager,
		slog.Default(),
	)

	got, err := service.ListModels(context.Background(), ListModelsInput{Namespace: "master_persona.llm"})
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(got) != 1 || got[0].ID != "model-a" {
		t.Fatalf("unexpected model list: %+v", got)
	}
	if !got[0].Capability.SupportsBatch {
		t.Fatalf("expected capability.supports_batch=true, got %+v", got[0])
	}
	if manager.lastConfig.Provider != "lmstudio" || manager.lastConfig.Model != "stored-model" {
		t.Fatalf("unexpected config used: %+v", manager.lastConfig)
	}
}

func TestModelCatalogService_ListModels_OverrideWins(t *testing.T) {
	manager := &mockLLMManager{client: &mockLLMClient{}}
	service := NewModelCatalogService(
		&mockConfigStore{values: map[string]map[string]string{
			"master_persona.llm": {"provider": "gemini"},
		}},
		&mockSecretStore{},
		manager,
		slog.Default(),
	)

	_, err := service.ListModels(context.Background(), ListModelsInput{
		Namespace: "master_persona.llm",
		Provider:  "xai",
		Endpoint:  "https://example.test",
		APIKey:    "k",
	})
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if manager.lastConfig.Provider != "xai" {
		t.Fatalf("expected override provider xai, got %s", manager.lastConfig.Provider)
	}
	if manager.lastConfig.Endpoint != "https://example.test" {
		t.Fatalf("expected endpoint override, got %s", manager.lastConfig.Endpoint)
	}
	if manager.lastConfig.APIKey != "k" {
		t.Fatalf("expected api key override")
	}
}
