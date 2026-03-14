package llm

import (
	"context"
	"fmt"
	"log/slog"
)

// Manager は LLMManager インターフェースの実装。
type Manager struct {
	logger *slog.Logger
}

// NewLLMManager は LLMManager のインスタンスを返す。
// Wire Provider として使用する。
func NewLLMManager(logger *slog.Logger) LLMManager {
	return &Manager{
		logger: logger.With("component", "llm_manager"),
	}
}

// GetClient は LLMConfig に基づいて LLMClient を返す。
// サポートプロバイダー: "gemini", "lmstudio"(互換: "local", "local-llm"), "xai"
func (m *Manager) GetClient(ctx context.Context, config LLMConfig) (LLMClient, error) {
	config.Provider = NormalizeProvider(config.Provider)
	m.logger.DebugContext(ctx, "ENTER GetClient", "provider", config.Provider, "model", config.Model)
	if config.Model == "" {
		return nil, ErrModelRequired
	}

	var c LLMClient
	switch config.Provider {
	case "gemini":
		c = NewGeminiClient(m.logger, config)
	case "lmstudio":
		c = NewLMStudioClient(m.logger, config)
	case "xai":
		c = NewXAIClient(m.logger, config)
	default:
		return nil, fmt.Errorf("llm_manager: unknown provider %q (supported: gemini, lmstudio, xai)", config.Provider)
	}

	m.logger.DebugContext(ctx, "EXIT GetClient", "provider", config.Provider)
	return c, nil
}

// GetBatchClient は LLMConfig に基づいて BatchClient を返す。
// Batch サポート: "gemini", "xai"
func (m *Manager) GetBatchClient(ctx context.Context, config LLMConfig) (BatchClient, error) {
	config.Provider = NormalizeProvider(config.Provider)
	m.logger.DebugContext(ctx, "ENTER GetBatchClient", "provider", config.Provider, "model", config.Model)

	switch config.Provider {
	case "xai":
		bc, err := NewXAIBatchClient(m.logger, config)
		if err != nil {
			return nil, fmt.Errorf("llm_manager: xai BatchClient creation failed: %w", err)
		}
		m.logger.DebugContext(ctx, "EXIT GetBatchClient", "provider", "xai")
		return bc, nil
	case "gemini":
		bc, err := NewGeminiBatchClient(m.logger, config)
		if err != nil {
			return nil, fmt.Errorf("llm_manager: gemini BatchClient creation failed: %w", err)
		}
		m.logger.DebugContext(ctx, "EXIT GetBatchClient", "provider", "gemini")
		return bc, nil
	case "lmstudio":
		return nil, fmt.Errorf("llm_manager: provider %q does not support Batch API", config.Provider)
	default:
		return nil, fmt.Errorf("llm_manager: unknown provider %q", config.Provider)
	}
}

// ResolveBulkStrategy は BulkStrategy を受け取り、プロバイダーに応じた有効なバルク戦略を返す。
func (m *Manager) ResolveBulkStrategy(ctx context.Context, strategy BulkStrategy, provider string) BulkStrategy {
	provider = NormalizeProvider(provider)
	m.logger.DebugContext(ctx, "ENTER ResolveBulkStrategy", "strategy", strategy, "provider", provider)

	if strategy == BulkStrategyBatch && !ProviderSupportsBatch(provider) {
		m.logger.WarnContext(ctx, "ResolveBulkStrategy: provider does not support batch strategy; falling back to sync",
			"provider", provider,
		)
		return BulkStrategySync
	}
	if strategy == "" {
		return BulkStrategySync
	}
	return strategy
}
