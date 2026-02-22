package llm_manager

import (
	"context"
	"fmt"
	"log/slog"

	llm "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client/gemini"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client/local"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client/xai"
)

// Manager は llm.LLMManager インターフェースの実装。
type Manager struct {
	logger *slog.Logger
}

// NewLLMManager は LLMManager のインスタンスを返す。
// Wire Provider として使用する。
func NewLLMManager(logger *slog.Logger) llm.LLMManager {
	return &Manager{
		logger: logger.With("component", "llm_manager"),
	}
}

// GetClient は config.Provider に基づいて LLMClient を返す。
// サポートプロバイダー: "gemini", "local", "xai"
func (m *Manager) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	m.logger.DebugContext(ctx, "ENTER GetClient", "provider", config.Provider, "model", config.Model)

	var c llm.LLMClient
	switch config.Provider {
	case "gemini":
		c = gemini.New(m.logger, config)
	case "local":
		c = local.New(m.logger, config)
	case "xai":
		c = xai.New(m.logger, config)
	default:
		return nil, fmt.Errorf("llm_manager: unknown provider %q (supported: gemini, local, xai)", config.Provider)
	}

	m.logger.DebugContext(ctx, "EXIT GetClient", "provider", config.Provider)
	return c, nil
}

// GetBatchClient は config.Provider に基づいて BatchClient を返す。
// Batch サポート: "xai"（"gemini" は将来実装、"local" は非対応）
func (m *Manager) GetBatchClient(ctx context.Context, config llm.LLMConfig) (llm.BatchClient, error) {
	m.logger.DebugContext(ctx, "ENTER GetBatchClient", "provider", config.Provider, "model", config.Model)

	switch config.Provider {
	case "xai":
		bc, err := xai.NewBatchClient(m.logger, config)
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
