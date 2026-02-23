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
// サポートプロバイダー: "gemini", "local", "xai"
func (m *Manager) GetClient(ctx context.Context, config LLMConfig) (LLMClient, error) {
	m.logger.DebugContext(ctx, "ENTER GetClient", "provider", config.Provider, "model", config.Model)

	var c LLMClient
	switch config.Provider {
	case "gemini":
		c = NewGeminiClient(m.logger, config)
	case "local":
		c = NewLocalClient(m.logger, config)
	case "xai":
		c = NewXAIClient(m.logger, config)
	default:
		return nil, fmt.Errorf("llm_manager: unknown provider %q (supported: gemini, local, xai)", config.Provider)
	}

	m.logger.DebugContext(ctx, "EXIT GetClient", "provider", config.Provider)
	return c, nil
}

// GetBatchClient は LLMConfig に基づいて BatchClient を返す。
// Batch サポート: "xai"
func (m *Manager) GetBatchClient(ctx context.Context, config LLMConfig) (BatchClient, error) {
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
		return nil, fmt.Errorf("llm_manager: gemini BatchClient not yet implemented")
	case "local":
		return nil, fmt.Errorf("llm_manager: provider %q does not support Batch API", config.Provider)
	default:
		return nil, fmt.Errorf("llm_manager: unknown provider %q", config.Provider)
	}
}

// ResolveBulkStrategy は BulkStrategy を受け取り、プロバイダーに応じた有効なバルク戦略を返す。
func (m *Manager) ResolveBulkStrategy(ctx context.Context, strategy BulkStrategy, provider string) BulkStrategy {
	m.logger.DebugContext(ctx, "ENTER ResolveBulkStrategy", "strategy", strategy, "provider", provider)

	if provider == "local" && strategy == BulkStrategyBatch {
		m.logger.WarnContext(ctx, "ResolveBulkStrategy: provider 'local' does not support batch strategy; falling back to sync",
			"provider", provider,
		)
		return BulkStrategySync
	}
	if strategy == "" {
		return BulkStrategySync
	}
	return strategy
}
