package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	geminiBaseURL        = "https://generativelanguage.googleapis.com"
	geminiAPIVersion     = "v1beta"
	geminiDefaultModel   = "gemini-1.5-flash"
	geminiDefaultTimeout = 120 * time.Second
)

// geminiClient は Gemini API の LLMClient 実装。
type geminiClient struct {
	config     LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
	retryCfg   RetryConfig
}

// NewGeminiClient は Gemini client を返す。
func NewGeminiClient(logger *slog.Logger, config LLMConfig) LLMClient {
	model := config.Model
	if model == "" {
		model = geminiDefaultModel
	}
	config.Model = model
	return &geminiClient{
		config:     config,
		httpClient: &http.Client{Timeout: geminiDefaultTimeout},
		logger:     logger.With("component", "gemini_client", "model", model),
		retryCfg:   DefaultRetryConfig(),
	}
}

// Complete はテキスト生成リクエストを実行し、結果を返す。
func (c *geminiClient) Complete(ctx context.Context, req Request) (Response, error) {
	c.logger.DebugContext(ctx, "ENTER Complete",
		"system_prompt_len", len(req.SystemPrompt),
		"user_prompt_len", len(req.UserPrompt),
	)

	var resp Response
	err := RetryWithBackoff(ctx, c.retryCfg, func() error {
		var innerErr error
		resp, innerErr = c.doComplete(ctx, req)
		return innerErr
	})
	if err != nil {
		c.logger.DebugContext(ctx, "EXIT Complete (error)", "error", err)
		return Response{}, err
	}

	resp.Metadata = req.Metadata

	c.logger.DebugContext(ctx, "EXIT Complete",
		"content_len", len(resp.Content),
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens,
	)
	return resp, nil
}

// StreamComplete はストリーミングレスポンスを返す（現在は非ストリーミングフォールバック）。
func (c *geminiClient) StreamComplete(ctx context.Context, req Request) (StreamResponse, error) {
	c.logger.DebugContext(ctx, "ENTER StreamComplete (non-streaming fallback)")
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	c.logger.DebugContext(ctx, "EXIT StreamComplete")
	return &geminiStreamResponse{resp: resp}, nil
}

// GetEmbedding はテキストの埋め込みベクトルを返す（スタブ）。
func (c *geminiClient) GetEmbedding(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("gemini: GetEmbedding not implemented")
}

// HealthCheck は Gemini API への疎通確認を行う。
func (c *geminiClient) HealthCheck(ctx context.Context) error {
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

// doComplete は1回分の HTTPリクエストを実行する（リトライは呼び出し元が制御）。
func (c *geminiClient) doComplete(ctx context.Context, req Request) (Response, error) {
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return Response{}, err
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("gemini: HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()
	elapsed := time.Since(start)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("gemini: failed to read response body: %w", err)
	}

	if IsRetryableStatusCode(httpResp.StatusCode) {
		return Response{}, &RetryableError{
			StatusCode: httpResp.StatusCode,
			Message:    string(body),
		}
	}
	if httpResp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("gemini: API error %d: %s", httpResp.StatusCode, string(body))
	}

	c.logger.DebugContext(ctx, "gemini HTTP response",
		"status", httpResp.StatusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
	return c.parseResponse(ctx, body)
}

// buildRequest は Gemini API 形式の *http.Request を構築する。
func (c *geminiClient) buildRequest(ctx context.Context, req Request) (*http.Request, error) {
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

// parseResponse は Gemini API のレスポンスボディをパースして Response を返す。
func (c *geminiClient) parseResponse(ctx context.Context, body []byte) (Response, error) {
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
		return Response{}, fmt.Errorf("gemini: failed to unmarshal response: %w", err)
	}

	if len(raw.Candidates) == 0 || len(raw.Candidates[0].Content.Parts) == 0 {
		return Response{}, fmt.Errorf("gemini: empty candidates in response")
	}

	content := raw.Candidates[0].Content.Parts[0].Text
	resp := Response{
		Content: content,
		Success: true,
		Usage: TokenUsage{
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

// geminiStreamResponse — Complete のフォールバック用
type geminiStreamResponse struct {
	resp Response
	done bool
}

func (s *geminiStreamResponse) Next() (Response, bool) {
	if s.done {
		return Response{}, false
	}
	s.done = true
	return s.resp, true
}

func (s *geminiStreamResponse) Close() error { return nil }
