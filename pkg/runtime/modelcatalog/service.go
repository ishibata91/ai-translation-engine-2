package modelcatalog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

// ModelCatalogService bridges UI model listing with config and llm infrastructure.
type ModelCatalogService struct {
	configStore config.Config
	secretStore config.SecretStore
	llmManager  llm.LLMManager
	logger      *slog.Logger
}

func NewModelCatalogService(
	configStore config.Config,
	secretStore config.SecretStore,
	llmManager llm.LLMManager,
	logger *slog.Logger,
) *ModelCatalogService {
	return &ModelCatalogService{
		configStore: configStore,
		secretStore: secretStore,
		llmManager:  llmManager,
		logger:      logger.With("component", "model_catalog_service"),
	}
}

// ListModels returns model options based on namespace config + optional UI overrides.
func (s *ModelCatalogService) ListModels(input ListModelsInput) ([]ModelOption, error) {
	ctx := context.Background()
	ns := strings.TrimSpace(input.Namespace)
	if ns == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	provider := s.resolveProvider(ctx, ns, input.Provider)
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	endpoint := firstNonEmpty(strings.TrimSpace(input.Endpoint), s.getConfig(ctx, ns, "endpoint"))
	apiKey := firstNonEmpty(input.APIKey, s.getConfig(ctx, ns, "api_key"), s.getSecret(ctx, ns, "api_key"))
	model := firstNonEmpty(
		s.getConfig(ctx, ns, "model"),
		s.getConfig(ctx, ns, provider+"_"+llm.LLMModelIDKeySuffix),
		"catalog-probe",
	)

	client, err := s.llmManager.GetClient(ctx, llm.LLMConfig{
		Provider: provider,
		Endpoint: endpoint,
		APIKey:   apiKey,
		Model:    model,
	})
	if err != nil {
		return nil, fmt.Errorf("get llm client provider=%s namespace=%s: %w", provider, ns, err)
	}

	models, err := client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("list models provider=%s namespace=%s: %w", provider, ns, err)
	}

	out := make([]ModelOption, 0, len(models))
	for _, m := range models {
		out = append(out, ModelOption{
			ID:               m.ID,
			DisplayName:      m.DisplayName,
			MaxContextLength: m.MaxContextLength,
			Loaded:           m.Loaded,
		})
	}

	s.logger.InfoContext(ctx, "listed models",
		slog.String("namespace", ns),
		slog.String("provider", provider),
		slog.Int("count", len(out)),
	)
	return out, nil
}

func (s *ModelCatalogService) resolveProvider(ctx context.Context, namespace string, override string) string {
	if normalized := llm.NormalizeProvider(override); normalized != "" {
		return normalized
	}
	if v := llm.NormalizeProvider(s.getConfig(ctx, namespace, "provider")); v != "" {
		return v
	}
	// Legacy "llm" namespace compatibility.
	if namespace == llm.LLMConfigNamespace {
		if v := llm.NormalizeProvider(s.getConfig(ctx, namespace, llm.LLMDefaultProviderKey)); v != "" {
			return v
		}
	}
	return ""
}

func (s *ModelCatalogService) getConfig(ctx context.Context, namespace, key string) string {
	val, err := s.configStore.Get(ctx, namespace, key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(val)
}

func (s *ModelCatalogService) getSecret(ctx context.Context, namespace, key string) string {
	val, err := s.secretStore.GetSecret(ctx, namespace, key)
	if err != nil {
		return ""
	}
	return val
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
