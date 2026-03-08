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

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
)

const (
	lmStudioDefaultEndpoint = "http://localhost:1234"
	lmStudioDefaultTimeout  = 120 * time.Second
)

// ModelLifecycleClient provides LM Studio model load/unload lifecycle hooks.
type ModelLifecycleClient interface {
	LoadModel(ctx context.Context, model string, contextLength int) (string, error)
	UnloadModel(ctx context.Context, instanceID string) error
}

// lmStudioClient is LM Studio/OpenAI-compatible LLM client implementation.
type lmStudioClient struct {
	config     LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
	retryCfg   RetryConfig
}

// NewLMStudioClient returns LM Studio client.
func NewLMStudioClient(logger *slog.Logger, config LLMConfig) LLMClient {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = lmStudioDefaultEndpoint
	}
	config.Endpoint = endpoint
	return &lmStudioClient{
		config:     config,
		httpClient: &http.Client{Timeout: lmStudioDefaultTimeout},
		logger:     logger.With("component", "lmstudio_client", "endpoint", endpoint),
		retryCfg:   DefaultRetryConfig(),
	}
}

// NewLocalClient is a legacy alias kept for compatibility.
func NewLocalClient(logger *slog.Logger, config LLMConfig) LLMClient {
	return NewLMStudioClient(logger, config)
}

func (c *lmStudioClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/api/v1/models", c.config.Endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: list models request creation failed: %w", err)
	}
	c.setAuthHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: list models request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("lmstudio: list models read failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lmstudio: list models error %d: %s", httpResp.StatusCode, string(body))
	}

	var raw struct {
		Models []struct {
			Type             string `json:"type"`
			Key              string `json:"key"`
			DisplayName      string `json:"display_name"`
			MaxContextLength int    `json:"max_context_length"`
			LoadedInstances  []struct {
				ID string `json:"id"`
			} `json:"loaded_instances"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("lmstudio: list models unmarshal failed: %w", err)
	}

	models := make([]ModelInfo, 0, len(raw.Models))
	for _, m := range raw.Models {
		if m.Type != "llm" {
			continue
		}
		models = append(models, ModelInfo{
			ID:               m.Key,
			DisplayName:      m.DisplayName,
			MaxContextLength: m.MaxContextLength,
			Loaded:           len(m.LoadedInstances) > 0,
		})
	}
	return models, nil
}

// Complete performs plain text completion.
func (c *lmStudioClient) Complete(ctx context.Context, req Request) (Response, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionLLMRequest)()
	if c.config.Model == "" {
		return Response{}, ErrModelRequired
	}

	var resp Response
	err := RetryWithBackoff(ctx, c.retryCfg, func() error {
		var innerErr error
		resp, innerErr = c.doChatCompletion(ctx, req, false)
		return innerErr
	})
	if err != nil {
		return Response{}, err
	}
	resp.Metadata = req.Metadata
	return resp, nil
}

// GenerateStructured performs OpenAI-compatible json_schema completion.
func (c *lmStudioClient) GenerateStructured(ctx context.Context, req Request) (Response, error) {
	defer telemetry.StartSpan(ctx, telemetry.ActionLLMRequest)()
	if c.config.Model == "" {
		return Response{}, ErrModelRequired
	}
	if len(req.ResponseSchema) == 0 {
		return Response{}, fmt.Errorf("lmstudio: response schema is required for structured generation")
	}
	var resp Response
	err := RetryWithBackoff(ctx, c.retryCfg, func() error {
		var innerErr error
		resp, innerErr = c.doChatCompletion(ctx, req, true)
		return innerErr
	})
	if err != nil {
		return Response{}, err
	}
	resp.Metadata = req.Metadata
	return resp, nil
}

func (c *lmStudioClient) StreamComplete(ctx context.Context, req Request) (StreamResponse, error) {
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	return &localStreamResponse{resp: resp}, nil
}

func (c *lmStudioClient) GetEmbedding(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("lmstudio: GetEmbedding not implemented")
}

func (c *lmStudioClient) HealthCheck(ctx context.Context) error {
	_, err := c.ListModels(ctx)
	return err
}

func (c *lmStudioClient) LoadModel(ctx context.Context, model string, contextLength int) (string, error) {
	if model == "" {
		return "", ErrModelRequired
	}
	payload := map[string]interface{}{
		"model": model,
	}
	if contextLength > 0 {
		payload["context_length"] = contextLength
	}
	bodyBytes, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/models/load", c.config.Endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("lmstudio: load request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("lmstudio: load request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lmstudio: load error %d: %s", httpResp.StatusCode, string(respBody))
	}

	var raw struct {
		InstanceID string `json:"instance_id"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "", fmt.Errorf("lmstudio: load unmarshal failed: %w", err)
	}
	if raw.InstanceID == "" {
		return "", fmt.Errorf("lmstudio: load response missing instance_id")
	}
	return raw.InstanceID, nil
}

func (c *lmStudioClient) UnloadModel(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return nil
	}
	payload := map[string]string{"instance_id": instanceID}
	bodyBytes, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/api/v1/models/unload", c.config.Endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("lmstudio: unload request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("lmstudio: unload request failed: %w", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("lmstudio: unload error %d: %s", httpResp.StatusCode, string(body))
	}
	return nil
}

func (c *lmStudioClient) doChatCompletion(ctx context.Context, req Request, structured bool) (Response, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type responseFormat struct {
		Type       string                 `json:"type"`
		JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
	}
	type requestBody struct {
		Model          string          `json:"model"`
		Messages       []message       `json:"messages"`
		Temperature    float32         `json:"temperature,omitempty"`
		Stream         bool            `json:"stream"`
		ResponseFormat *responseFormat `json:"response_format,omitempty"`
	}

	msgs := make([]message, 0, 2)
	if req.SystemPrompt != "" {
		msgs = append(msgs, message{Role: "system", Content: req.SystemPrompt})
	}
	msgs = append(msgs, message{Role: "user", Content: req.UserPrompt})

	body := requestBody{
		Model:       c.config.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		Stream:      false,
	}
	if structured {
		body.ResponseFormat = &responseFormat{
			Type: "json_schema",
			JSONSchema: map[string]interface{}{
				"name":   "structured_output",
				"strict": true,
				"schema": req.ResponseSchema,
			},
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return Response{}, fmt.Errorf("lmstudio: marshal request failed: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", c.config.Endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return Response{}, fmt.Errorf("lmstudio: request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("lmstudio: request failed: %w", err)
	}
	defer httpResp.Body.Close()
	respBody, _ := io.ReadAll(httpResp.Body)

	if IsRetryableStatusCode(httpResp.StatusCode) {
		return Response{}, &RetryableError{StatusCode: httpResp.StatusCode, Message: string(respBody)}
	}
	if httpResp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("lmstudio: chat completions error %d: %s", httpResp.StatusCode, string(respBody))
	}

	var raw struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return Response{}, fmt.Errorf("lmstudio: response unmarshal failed: %w", err)
	}
	if len(raw.Choices) == 0 {
		return Response{}, fmt.Errorf("lmstudio: empty choices in response")
	}
	var content string
	if err := json.Unmarshal(raw.Choices[0].Message.Content, &content); err != nil {
		return Response{}, fmt.Errorf("lmstudio: response content decode failed: %w", err)
	}
	return Response{
		Content: content,
		Success: true,
		Usage: TokenUsage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		},
	}, nil
}

func (c *lmStudioClient) setAuthHeader(req *http.Request) {
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

// localStreamResponse is a fallback stream wrapper.
type localStreamResponse struct {
	resp Response
	done bool
}

func (s *localStreamResponse) Next() (Response, bool) {
	if s.done {
		return Response{}, false
	}
	s.done = true
	return s.resp, true
}

func (s *localStreamResponse) Close() error { return nil }
