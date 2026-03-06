package llm

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLMStudioClient_ListModels_Normalize(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]any{
				{"type": "embedding", "key": "emb-1", "display_name": "Emb", "max_context_length": 0},
				{"type": "llm", "key": "m1", "display_name": "Model 1", "max_context_length": 4096, "loaded_instances": []map[string]any{{"id": "i1"}}},
			},
		})
	}))
	defer srv.Close()

	client := NewLMStudioClient(slog.New(slog.NewTextHandler(os.Stdout, nil)), LLMConfig{
		Provider: "lmstudio",
		Endpoint: srv.URL,
		Model:    "m1",
	})
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 llm model, got %d", len(models))
	}
	if models[0].ID != "m1" || !models[0].Loaded || models[0].MaxContextLength != 4096 {
		t.Fatalf("unexpected normalized model: %+v", models[0])
	}
}

func TestLMStudioClient_LoadUnloadAndStructured(t *testing.T) {
	t.Parallel()
	loadCount := 0
	unloadCount := 0
	structuredStrict := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/models/load":
			loadCount++
			_ = json.NewEncoder(w).Encode(map[string]any{"instance_id": "inst-1", "status": "loaded"})
		case "/api/v1/models/unload":
			unloadCount++
			_ = json.NewEncoder(w).Encode(map[string]any{"instance_id": "inst-1"})
		case "/v1/chat/completions":
			b, _ := io.ReadAll(r.Body)
			var payload map[string]any
			_ = json.Unmarshal(b, &payload)
			if rf, ok := payload["response_format"].(map[string]any); ok {
				if js, ok := rf["json_schema"].(map[string]any); ok {
					if v, ok := js["strict"].(bool); ok && v {
						structuredStrict = true
					}
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{"content": `{"ok":true}`}}},
				"usage":   map[string]any{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewLMStudioClient(slog.New(slog.NewTextHandler(os.Stdout, nil)), LLMConfig{
		Provider: "lmstudio",
		Endpoint: srv.URL,
		Model:    "m1",
	})
	lifecycle, ok := client.(ModelLifecycleClient)
	if !ok {
		t.Fatalf("expected ModelLifecycleClient")
	}
	instanceID, err := lifecycle.LoadModel(context.Background(), "m1", 8192)
	if err != nil {
		t.Fatalf("LoadModel failed: %v", err)
	}
	if err := lifecycle.UnloadModel(context.Background(), instanceID); err != nil {
		t.Fatalf("UnloadModel failed: %v", err)
	}
	_, err = client.GenerateStructured(context.Background(), Request{
		UserPrompt:     "test",
		ResponseSchema: map[string]interface{}{"type": "object"},
	})
	if err != nil {
		t.Fatalf("GenerateStructured failed: %v", err)
	}
	if loadCount != 1 || unloadCount != 1 {
		t.Fatalf("expected load/unload = 1/1, got %d/%d", loadCount, unloadCount)
	}
	if !structuredStrict {
		t.Fatalf("expected strict=true in response_format.json_schema")
	}
}

func TestStructuredOutputNotSupportedProviders(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tests := []LLMClient{
		NewGeminiClient(logger, LLMConfig{Provider: "gemini", Model: "gemini-1.5-flash"}),
		NewXAIClient(logger, LLMConfig{Provider: "xai", Model: "grok-3"}),
	}
	for _, c := range tests {
		_, err := c.GenerateStructured(context.Background(), Request{
			UserPrompt:     "x",
			ResponseSchema: map[string]interface{}{"type": "object"},
		})
		if err != ErrStructuredOutputNotSupported {
			t.Fatalf("expected ErrStructuredOutputNotSupported, got %v", err)
		}
	}
}
