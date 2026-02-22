package gemini

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
	geminiBaseURL    = "https://generativelanguage.googleapis.com"
	geminiAPIVersion = "v1beta"
	defaultModel     = "gemini-1.5-flash"
	defaultTimeout   = 120 * time.Second
)

// client は Gemini API の LLMClient 実装。
type client struct {
	config     llm.LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
	retryCfg   llm.RetryConfig
}

// New は Gemini client を返す。
func New(logger *slog.Logger, config llm.LLMConfig) llm.LLMClient {
	model := config.Model
	if model == "" {
		model = defaultModel
	}
	config.Model = model
	return &client{
		config:     config,
		httpClient: &http.Client{Timeout: defaultTimeout},
		logger:     logger.With("component", "gemini_client", "model", model),
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

	c.logger.DebugContext(ctx, "EXIT Complete",
		"content_len", len(resp.Content),
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens,
	)
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
	return nil, fmt.Errorf("gemini: GetEmbedding not implemented")
}

// HealthCheck は Gemini API への疎通確認を行う。
func (c *client) HealthCheck(ctx context.Context) error {
	c.logger.DebugContext(ctx, "ENTER HealthCheck")
	url := fmt.Sprintf("%s/%s/models?key=%s", geminiBaseURL, geminiAPIVersion, c.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("gemini: HealthCheck request creation failed: %w", err)
	}
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("gemini: HealthCheck request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("gemini: HealthCheck returned status %d", httpResp.StatusCode)
	}
	c.logger.DebugContext(ctx, "EXIT HealthCheck", "status", httpResp.StatusCode)
	return nil
}

// ─────────────────────────────────────────────
// プライベートメソッド（同一ファイル内 SRP 分割）
// ─────────────────────────────────────────────

// doComplete は1回分の HTTPリクエストを実行する（リトライは呼び出し元が制御）。
func (c *client) doComplete(ctx context.Context, req llm.Request) (llm.Response, error) {
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return llm.Response{}, err
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("gemini: HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()
	elapsed := time.Since(start)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return llm.Response{}, fmt.Errorf("gemini: failed to read response body: %w", err)
	}

	if llm.IsRetryableStatusCode(httpResp.StatusCode) {
		return llm.Response{}, &llm.RetryableError{
			StatusCode: httpResp.StatusCode,
			Message:    string(body),
		}
	}
	if httpResp.StatusCode != http.StatusOK {
		return llm.Response{}, fmt.Errorf("gemini: API error %d: %s", httpResp.StatusCode, string(body))
	}

	c.logger.DebugContext(ctx, "gemini HTTP response",
		"status", httpResp.StatusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
	return c.parseResponse(ctx, body)
}

// buildRequest は Gemini API 形式の *http.Request を構築する。
func (c *client) buildRequest(ctx context.Context, req llm.Request) (*http.Request, error) {
	c.logger.DebugContext(ctx, "ENTER buildRequest", "model", c.config.Model)

	url := fmt.Sprintf("%s/%s/models/%s:generateContent?key=%s",
		geminiBaseURL, geminiAPIVersion, c.config.Model, c.config.APIKey)

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}
	type generationConfig struct {
		Temperature     float32 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	}
	type requestBody struct {
		Contents          []content        `json:"contents"`
		SystemInstruction *content         `json:"systemInstruction,omitempty"`
		GenerationConfig  generationConfig `json:"generationConfig,omitempty"`
	}

	body := requestBody{
		Contents: []content{
			{Role: "user", Parts: []part{{Text: req.UserPrompt}}},
		},
		GenerationConfig: generationConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}
	if req.SystemPrompt != "" {
		body.SystemInstruction = &content{
			Role:  "user",
			Parts: []part{{Text: req.SystemPrompt}},
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.DebugContext(ctx, "EXIT buildRequest", "url_path", fmt.Sprintf("/models/%s:generateContent", c.config.Model))
	return httpReq, nil
}

// parseResponse は Gemini API のレスポンスボディをパースして llm.Response を返す。
func (c *client) parseResponse(ctx context.Context, body []byte) (llm.Response, error) {
	c.logger.DebugContext(ctx, "ENTER parseResponse", "body_len", len(body))

	var raw struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return llm.Response{}, fmt.Errorf("gemini: failed to unmarshal response: %w", err)
	}

	if len(raw.Candidates) == 0 || len(raw.Candidates[0].Content.Parts) == 0 {
		return llm.Response{}, fmt.Errorf("gemini: empty candidates in response")
	}

	content := raw.Candidates[0].Content.Parts[0].Text
	resp := llm.Response{
		Content: content,
		Success: true,
		Usage: llm.TokenUsage{
			PromptTokens:     raw.UsageMetadata.PromptTokenCount,
			CompletionTokens: raw.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      raw.UsageMetadata.TotalTokenCount,
		},
	}

	c.logger.DebugContext(ctx, "EXIT parseResponse",
		"content_len", len(content),
		"total_tokens", resp.Usage.TotalTokens,
	)
	return resp, nil
}

// ─────────────────────────────────────────────
// singleStreamResponse — Complete のフォールバック用
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
