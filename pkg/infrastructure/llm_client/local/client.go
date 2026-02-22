package local

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	llm "github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

const (
	defaultEndpoint = "http://localhost:11434"
	defaultTimeout  = 120 * time.Second
)

// client は Ollama 互換 Local LLM の LLMClient 実装。
type client struct {
	config     llm.LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
	retryCfg   llm.RetryConfig
}

// New は Local (Ollama互換) client を返す。
func New(logger *slog.Logger, config llm.LLMConfig) llm.LLMClient {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	config.Endpoint = endpoint
	return &client{
		config:     config,
		httpClient: &http.Client{Timeout: defaultTimeout},
		logger:     logger.With("component", "local_client", "endpoint", endpoint),
		retryCfg:   llm.DefaultRetryConfig(),
	}
}

// ─────────────────────────────────────────────
// LLMClient 実装
// ─────────────────────────────────────────────

// Complete はテキスト生成リクエストを実行し、結果を返す。
func (c *client) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	c.logger.DebugContext(ctx, "ENTER Complete",
		"system_prompt_len", len(req.SystemPrompt),
		"user_prompt_len", len(req.UserPrompt),
	)

	var resp llm.Response
	err := llm.RetryWithBackoff(ctx, c.retryCfg, func() error {
		var innerErr error
		resp, innerErr = c.doComplete(ctx, req)
		return innerErr
	})
	if err != nil {
		c.logger.DebugContext(ctx, "EXIT Complete (error)", "error", err)
		return llm.Response{}, err
	}

	c.logger.DebugContext(ctx, "EXIT Complete", "content_len", len(resp.Content))
	return resp, nil
}

// StreamComplete はストリーミングレスポンスを返す（現在は非ストリーミングフォールバック）。
func (c *client) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	c.logger.DebugContext(ctx, "ENTER StreamComplete (non-streaming fallback)")
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	c.logger.DebugContext(ctx, "EXIT StreamComplete")
	return &singleStreamResponse{resp: resp}, nil
}

// GetEmbedding はテキストの埋め込みベクトルを返す（スタブ）。
func (c *client) GetEmbedding(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("local: GetEmbedding not implemented")
}

// HealthCheck は Local LLM サーバーへの疎通確認を行う。
func (c *client) HealthCheck(ctx context.Context) error {
	c.logger.DebugContext(ctx, "ENTER HealthCheck")
	url := fmt.Sprintf("%s/api/tags", c.config.Endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("local: HealthCheck request creation failed: %w", err)
	}
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("local: HealthCheck request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("local: HealthCheck returned status %d", httpResp.StatusCode)
	}
	c.logger.DebugContext(ctx, "EXIT HealthCheck", "status", httpResp.StatusCode)
	return nil
}

// ─────────────────────────────────────────────
// プライベートメソッド（同一ファイル内 SRP 分割）
// ─────────────────────────────────────────────

// doComplete は1回分の HTTPリクエストを実行する。
func (c *client) doComplete(ctx context.Context, req llm.Request) (llm.Response, error) {
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return llm.Response{}, err
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("local: HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()
	elapsed := time.Since(start)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return llm.Response{}, fmt.Errorf("local: failed to read response body: %w", err)
	}

	if llm.IsRetryableStatusCode(httpResp.StatusCode) {
		return llm.Response{}, &llm.RetryableError{
			StatusCode: httpResp.StatusCode,
			Message:    string(body),
		}
	}
	if httpResp.StatusCode != http.StatusOK {
		return llm.Response{}, fmt.Errorf("local: API error %d: %s", httpResp.StatusCode, string(body))
	}

	c.logger.DebugContext(ctx, "local HTTP response",
		"status", httpResp.StatusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
	return c.parseResponse(ctx, body)
}

// buildRequest は Ollama API 形式の *http.Request を構築する。
func (c *client) buildRequest(ctx context.Context, req llm.Request) (*http.Request, error) {
	c.logger.DebugContext(ctx, "ENTER buildRequest", "model", c.config.Model)

	url := fmt.Sprintf("%s/api/generate", c.config.Endpoint)

	// system_promptとuser_promptを結合してpromptとして送信
	prompt := req.UserPrompt
	if req.SystemPrompt != "" {
		prompt = req.SystemPrompt + "\n\n" + req.UserPrompt
	}

	type requestBody struct {
		Model       string  `json:"model"`
		Prompt      string  `json:"prompt"`
		Stream      bool    `json:"stream"`
		Temperature float32 `json:"temperature,omitempty"`
	}

	body := requestBody{
		Model:       c.config.Model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: req.Temperature,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("local: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("local: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.DebugContext(ctx, "EXIT buildRequest", "url", url)
	return httpReq, nil
}

// parseResponse は Ollama API のレスポンスボディをパースして llm.Response を返す。
func (c *client) parseResponse(ctx context.Context, body []byte) (llm.Response, error) {
	c.logger.DebugContext(ctx, "ENTER parseResponse", "body_len", len(body))

	var raw struct {
		Response        string `json:"response"`
		Done            bool   `json:"done"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return llm.Response{}, fmt.Errorf("local: failed to unmarshal response: %w", err)
	}

	resp := llm.Response{
		Content: raw.Response,
		Success: true,
		Usage: llm.TokenUsage{
			PromptTokens:     raw.PromptEvalCount,
			CompletionTokens: raw.EvalCount,
			TotalTokens:      raw.PromptEvalCount + raw.EvalCount,
		},
	}

	c.logger.DebugContext(ctx, "EXIT parseResponse",
		"content_len", len(raw.Response),
		"total_tokens", resp.Usage.TotalTokens,
	)
	return resp, nil
}

// ─────────────────────────────────────────────
// singleStreamResponse — フォールバック用
// ─────────────────────────────────────────────

type singleStreamResponse struct {
	resp llm.Response
	done bool
}

func (s *singleStreamResponse) Next() (llm.Response, bool) {
	if s.done {
		return llm.Response{}, false
	}
	s.done = true
	return s.resp, true
}

func (s *singleStreamResponse) Close() error { return nil }
