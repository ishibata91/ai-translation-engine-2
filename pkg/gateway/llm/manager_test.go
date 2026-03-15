package llm

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestLLMManager_GetClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewLLMManager(logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		config  LLMConfig
		wantErr bool
	}{
		{
			name:    "正常系: Gemini クライアントが返る",
			config:  LLMConfig{Provider: "gemini", APIKey: "test-key", Model: "gemini-1.5-flash"},
			wantErr: false,
		},
		{
			name:    "正常系: LM Studio クライアントが返る",
			config:  LLMConfig{Provider: "lmstudio", Endpoint: "http://localhost:1234", Model: "llama3"},
			wantErr: false,
		},
		{
			name:    "正常系: xAI クライアントが返る",
			config:  LLMConfig{Provider: "xai", APIKey: "test-key", Model: "grok-3"},
			wantErr: false,
		},
		{
			name:    "異常系: 不明なプロバイダーはエラー",
			config:  LLMConfig{Provider: "unknown"},
			wantErr: true,
		},
		{
			name:    "異常系: モデル未指定はエラー",
			config:  LLMConfig{Provider: "gemini"},
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

func TestLLMManager_ResolveBulkStrategy(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewLLMManager(logger)
	ctx := context.Background()

	if got := manager.ResolveBulkStrategy(ctx, BulkStrategyBatch, "lmstudio"); got != BulkStrategySync {
		t.Fatalf("lmstudio batch must fallback to sync: got=%s", got)
	}
	if got := manager.ResolveBulkStrategy(ctx, BulkStrategyBatch, "gemini"); got != BulkStrategyBatch {
		t.Fatalf("gemini batch should stay batch: got=%s", got)
	}
	if got := manager.ResolveBulkStrategy(ctx, BulkStrategy(""), "xai"); got != BulkStrategySync {
		t.Fatalf("empty strategy must default to sync: got=%s", got)
	}
}

func TestLLMManager_GetBatchClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	manager := NewLLMManager(logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		config  LLMConfig
		wantErr bool
	}{
		{
			name:    "正常系: Gemini BatchClient が返る",
			config:  LLMConfig{Provider: "gemini", APIKey: "test-key", Model: "gemini-2.0-flash"},
			wantErr: false,
		},
		{
			name:    "正常系: xAI BatchClient (grok-3) が返る",
			config:  LLMConfig{Provider: "xai", APIKey: "test-key", Model: "grok-3"},
			wantErr: false,
		},
		{
			name:    "異常系: lmstudio は Batch 非対応",
			config:  LLMConfig{Provider: "lmstudio", Endpoint: "http://localhost:1234", Model: "llama3"},
			wantErr: true,
		},
		{
			name:    "異常系: 不明なプロバイダーはエラー",
			config:  LLMConfig{Provider: "unknown"},
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
