package llm

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestGeminiBuildRequestNormalizesModelResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		model    string
		wantPath string
	}{
		{
			name:     "raw model ID",
			model:    "gemini-3.1-flash-lite-preview",
			wantPath: "/v1beta/models/gemini-3.1-flash-lite-preview:generateContent",
		},
		{
			name:     "resource model",
			model:    "models/gemini-3.1-flash-lite-preview",
			wantPath: "/v1beta/models/gemini-3.1-flash-lite-preview:generateContent",
		},
		{
			name:     "resource model with spaces",
			model:    "  models/gemini-3.1-flash-lite-preview  ",
			wantPath: "/v1beta/models/gemini-3.1-flash-lite-preview:generateContent",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &geminiClient{
				config: LLMConfig{
					APIKey: "test-key",
					Model:  tt.model,
				},
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			httpReq, err := client.buildRequest(context.Background(), Request{
				UserPrompt: "translate this",
			})
			if err != nil {
				t.Fatalf("buildRequest error: %v", err)
			}

			if got := httpReq.URL.Path; got != tt.wantPath {
				t.Fatalf("path = %q, want %q", got, tt.wantPath)
			}
			if got := httpReq.URL.Query().Get("key"); got != "test-key" {
				t.Fatalf("query key = %q, want %q", got, "test-key")
			}
		})
	}
}

func TestGeminiSyncBulkWithPrefixedModelUsesNormalizedPath(t *testing.T) {
	t.Parallel()

	const expectedPath = "/v1beta/models/gemini-3.1-flash-lite-preview:generateContent"

	var (
		mu            sync.Mutex
		receivedPaths []string
	)
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		receivedPaths = append(receivedPaths, req.URL.Path)
		mu.Unlock()

		if req.URL.Path != expectedPath {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
				Request:    req,
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(
				`{"candidates":[{"content":{"parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`,
			)),
			Request: req,
		}, nil
	})

	rawClient := NewGeminiClient(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		LLMConfig{
			APIKey: "test-key",
			Model:  "models/gemini-3.1-flash-lite-preview",
		},
	)
	client, ok := rawClient.(*geminiClient)
	if !ok {
		t.Fatalf("NewGeminiClient returned unexpected type: %T", rawClient)
	}
	client.httpClient = &http.Client{Transport: transport}

	reqs := make([]Request, 12)
	for i := range reqs {
		reqs[i] = Request{UserPrompt: fmt.Sprintf("prompt-%d", i)}
	}

	results, err := ExecuteBulkSync(context.Background(), client, reqs, 4)
	if err != nil {
		t.Fatalf("ExecuteBulkSync error: %v", err)
	}
	if len(results) != len(reqs) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(reqs))
	}
	for idx, res := range results {
		if !res.Success {
			t.Fatalf("result[%d] failed: %+v", idx, res)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(receivedPaths) != len(reqs) {
		t.Fatalf("len(receivedPaths) = %d, want %d", len(receivedPaths), len(reqs))
	}
	for idx, path := range receivedPaths {
		if path != expectedPath {
			t.Fatalf("path[%d] = %q, want %q", idx, path, expectedPath)
		}
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
