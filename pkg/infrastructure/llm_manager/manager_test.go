package llm_manager_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	llm "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_manager"
)

func TestLLMManager_GetClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := llm_manager.NewLLMManager(logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		config  llm.LLMConfig
		wantErr bool
	}{
		{
			name:    "正常系: Gemini クライアントが返る",
			config:  llm.LLMConfig{Provider: "gemini", APIKey: "test-key", Model: "gemini-1.5-flash"},
			wantErr: false,
		},
		{
			name:    "正常系: Local クライアントが返る",
			config:  llm.LLMConfig{Provider: "local", Endpoint: "http://localhost:11434", Model: "llama3"},
			wantErr: false,
		},
		{
			name:    "正常系: xAI クライアントが返る",
			config:  llm.LLMConfig{Provider: "xai", APIKey: "test-key", Model: "grok-3"},
			wantErr: false,
		},
		{
			name:    "異常系: 不明なプロバイダーはエラー",
			config:  llm.LLMConfig{Provider: "unknown"},
			wantErr: true,
		},
		{
			name:    "異常系: 空のプロバイダーはエラー",
			config:  llm.LLMConfig{},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := manager.GetClient(ctx, tc.config)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if client == nil {
				t.Error("expected non-nil client")
			}
		})
	}
}

func TestLLMManager_GetBatchClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := llm_manager.NewLLMManager(logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		config  llm.LLMConfig
		wantErr bool
	}{
		{
			name:    "正常系: xAI BatchClient (grok-3) が返る",
			config:  llm.LLMConfig{Provider: "xai", APIKey: "test-key", Model: "grok-3"},
			wantErr: false,
		},
		{
			name:    "異常系: xAI grok-3-mini は Batch 非対応",
			config:  llm.LLMConfig{Provider: "xai", APIKey: "test-key", Model: "grok-3-mini"},
			wantErr: true,
		},
		{
			name:    "異常系: local は Batch 非対応",
			config:  llm.LLMConfig{Provider: "local", Endpoint: "http://localhost:11434", Model: "llama3"},
			wantErr: true,
		},
		{
			name:    "異常系: 不明なプロバイダーはエラー",
			config:  llm.LLMConfig{Provider: "unknown"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bc, err := manager.GetBatchClient(ctx, tc.config)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if bc == nil {
				t.Error("expected non-nil batch client")
			}
		})
	}
}
