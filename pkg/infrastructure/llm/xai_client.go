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
	xaiBaseURL           = "https://api.x.ai/v1"
	xaiChatEndpoint      = "/chat/completions"
	xaiBatchesEndpoint   = "/batches"
	xaiModelsEndpoint    = "/models"
	xaiDefaultTimeout    = 300 * time.Second
	xaiPollInterval      = 30 * time.Second
	xaiMaxBatchChunkSize = 100
)

// XAIBatchSupportedModels は xAI Batch API がサポートするモデル一覧。
var XAIBatchSupportedModels = map[string]bool{
	"grok-3":                      true,
	"grok-4-0709":                 true,
	"grok-4-fast-non-reasoning":   true,
	"grok-4-fast-reasoning":       true,
	"grok-4-1-fast-non-reasoning": true,
	"grok-4-1-fast-reasoning":     true,
}

// ─────────────────────────────────────────────
// 同期クライアント（LLMClient 実装）
// ─────────────────────────────────────────────

// xaiClient は xAI Grok API の LLMClient 実装（同期）。
type xaiClient struct {
	config     LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
	retryCfg   RetryConfig
}

// NewXAIClient は xAI 同期 client を返す。
func NewXAIClient(logger *slog.Logger, config LLMConfig) LLMClient {
	return &xaiClient{
		config:     config,
		httpClient: &http.Client{Timeout: xaiDefaultTimeout},
		logger:     logger.With("component", "xai_client", "model", config.Model),
		retryCfg:   DefaultRetryConfig(),
	}
}

// Complete はテキスト生成リクエストを実行し、結果を返す。
func (c *xaiClient) Complete(ctx context.Context, req Request) (Response, error) {
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
		"total_tokens", resp.Usage.TotalTokens,
	)
	return resp, nil
}

// StreamComplete はストリーミングレスポンスを返す（フォールバック）。
func (c *xaiClient) StreamComplete(ctx context.Context, req Request) (StreamResponse, error) {
	c.logger.DebugContext(ctx, "ENTER StreamComplete (non-streaming fallback)")
	resp, err := c.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	c.logger.DebugContext(ctx, "EXIT StreamComplete")
	return &xaiStreamResponse{resp: resp}, nil
}

// GetEmbedding はスタブ実装（xAI は現在 embedding 非対応）。
func (c *xaiClient) GetEmbedding(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("xai: GetEmbedding not supported by xAI API")
}

// HealthCheck は xAI API への疎通確認を行う。
func (c *xaiClient) HealthCheck(ctx context.Context) error {
	c.logger.DebugContext(ctx, "ENTER HealthCheck")
	url := xaiBaseURL + xaiModelsEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("xai: HealthCheck request creation failed: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("xai: HealthCheck request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("xai: HealthCheck returned status %d", httpResp.StatusCode)
	}
	c.logger.DebugContext(ctx, "EXIT HealthCheck", "status", httpResp.StatusCode)
	return nil
}

// doComplete は1回分の HTTPリクエストを実行する。
func (c *xaiClient) doComplete(ctx context.Context, req Request) (Response, error) {
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return Response{}, err
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("xai: HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()
	elapsed := time.Since(start)

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("xai: failed to read response body: %w", err)
	}

	if IsRetryableStatusCode(httpResp.StatusCode) {
		return Response{}, &RetryableError{
			StatusCode: httpResp.StatusCode,
			Message:    string(body),
		}
	}
	if httpResp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("xai: API error %d: %s", httpResp.StatusCode, string(body))
	}

	c.logger.DebugContext(ctx, "xai HTTP response",
		"status", httpResp.StatusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
	return c.parseResponse(ctx, body)
}

// buildRequest は xAI OpenAI互換形式の *http.Request を構築する。
func (c *xaiClient) buildRequest(ctx context.Context, req Request) (*http.Request, error) {
	c.logger.DebugContext(ctx, "ENTER buildRequest", "model", c.config.Model)

	url := xaiBaseURL + xaiChatEndpoint

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type requestBody struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float32   `json:"temperature,omitempty"`
		MaxTokens   int       `json:"max_tokens,omitempty"`
		Stream      bool      `json:"stream"`
	}

	messages := []message{}
	if req.SystemPrompt != "" {
		messages = append(messages, message{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, message{Role: "user", Content: req.UserPrompt})

	body := requestBody{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("xai: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("xai: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	c.logger.DebugContext(ctx, "EXIT buildRequest", "url", url)
	return httpReq, nil
}

// parseResponse は xAI OpenAI互換レスポンスをパースして Response を返す。
func (c *xaiClient) parseResponse(ctx context.Context, body []byte) (Response, error) {
	c.logger.DebugContext(ctx, "ENTER parseResponse", "body_len", len(body))

	var raw struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return Response{}, fmt.Errorf("xai: failed to unmarshal response: %w", err)
	}
	if len(raw.Choices) == 0 {
		return Response{}, fmt.Errorf("xai: no choices in response")
	}

	content := raw.Choices[0].Message.Content
	resp := Response{
		Content: content,
		Success: true,
		Usage: TokenUsage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		},
	}

	c.logger.DebugContext(ctx, "EXIT parseResponse",
		"content_len", len(content),
		"total_tokens", resp.Usage.TotalTokens,
	)
	return resp, nil
}

// ─────────────────────────────────────────────
// バッチクライアント（BatchClient 実装）
// xAI 専用フォーマット — OpenAI Batch API と非互換
// ─────────────────────────────────────────────

// xaiBatchClient は xAI Batch API の BatchClient 実装。
type xaiBatchClient struct {
	config     LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
}

// NewXAIBatchClient は xAI BatchClient を返す。
func NewXAIBatchClient(logger *slog.Logger, config LLMConfig) (BatchClient, error) {
	if !XAIBatchSupportedModels[config.Model] {
		return nil, fmt.Errorf("xai: model %q does not support Batch API (supported: grok-3, grok-4-*)", config.Model)
	}
	return &xaiBatchClient{
		config:     config,
		httpClient: &http.Client{Timeout: xaiDefaultTimeout},
		logger:     logger.With("component", "xai_batch_client", "model", config.Model),
	}, nil
}

// SubmitBatch はリクエストリストをチャンク単位でバッチジョブに送信し、BatchJobID を返す。
func (b *xaiBatchClient) SubmitBatch(ctx context.Context, reqs []Request) (BatchJobID, error) {
	b.logger.DebugContext(ctx, "ENTER SubmitBatch", "request_count", len(reqs))

	if len(reqs) == 0 {
		return BatchJobID{}, fmt.Errorf("xai: no requests to submit")
	}

	// 1. バッチジョブを作成
	batchID, err := b.createBatch(ctx, fmt.Sprintf("translate-%d", len(reqs)))
	if err != nil {
		return BatchJobID{}, err
	}
	b.logger.DebugContext(ctx, "batch created", "batch_id", batchID)

	// 2. チャンク単位でリクエストを追加 → ポーリング
	for chunkStart := 0; chunkStart < len(reqs); chunkStart += xaiMaxBatchChunkSize {
		chunkEnd := chunkStart + xaiMaxBatchChunkSize
		if chunkEnd > len(reqs) {
			chunkEnd = len(reqs)
		}
		chunk := reqs[chunkStart:chunkEnd]

		if err := b.addRequests(ctx, batchID, chunk, chunkStart); err != nil {
			return BatchJobID{}, fmt.Errorf("xai: failed to add requests chunk [%d:%d]: %w", chunkStart, chunkEnd, err)
		}

		if err := b.pollUntilCompleted(ctx, batchID); err != nil {
			return BatchJobID{}, fmt.Errorf("xai: polling failed for chunk [%d:%d]: %w", chunkStart, chunkEnd, err)
		}

		b.logger.DebugContext(ctx, "chunk completed",
			"batch_id", batchID,
			"chunk_start", chunkStart,
			"chunk_end", chunkEnd,
		)
	}

	jobID := BatchJobID{ID: batchID, Provider: "xai"}
	b.logger.DebugContext(ctx, "EXIT SubmitBatch", "batch_id", batchID)
	return jobID, nil
}

// GetBatchStatus はバッチジョブのステータスを返す。
func (b *xaiBatchClient) GetBatchStatus(ctx context.Context, id BatchJobID) (BatchStatus, error) {
	b.logger.DebugContext(ctx, "ENTER GetBatchStatus", "batch_id", id.ID)

	url := xaiBaseURL + xaiBatchesEndpoint + "/" + id.ID
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("xai: GetBatchStatus request creation failed: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("xai: GetBatchStatus request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return BatchStatus{}, fmt.Errorf("xai: GetBatchStatus error %d: %s", httpResp.StatusCode, string(body))
	}

	status, err := b.parseBatchStatus(ctx, body, id.ID)
	if err != nil {
		return BatchStatus{}, err
	}

	b.logger.DebugContext(ctx, "EXIT GetBatchStatus",
		"batch_id", id.ID,
		"state", status.State,
		"progress", status.Progress,
	)
	return status, nil
}

// GetBatchResults はバッチジョブの全結果をページネーションで取得して返す。
func (b *xaiBatchClient) GetBatchResults(ctx context.Context, id BatchJobID) ([]Response, error) {
	b.logger.DebugContext(ctx, "ENTER GetBatchResults", "batch_id", id.ID)

	var allResults []Response
	paginationToken := ""

	for {
		results, nextToken, err := b.fetchResultPage(ctx, id.ID, paginationToken)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)

		if nextToken == "" {
			break
		}
		paginationToken = nextToken
	}

	b.logger.DebugContext(ctx, "EXIT GetBatchResults",
		"batch_id", id.ID,
		"total_results", len(allResults),
	)
	return allResults, nil
}

// ─────────────────────────────────────────────
// プライベートメソッド（バッチ同一ファイル内 SRP 分割）
// ─────────────────────────────────────────────

// createBatch は xAI Batch API でバッチジョブを作成し、batch_id を返す。
func (b *xaiBatchClient) createBatch(ctx context.Context, name string) (string, error) {
	b.logger.DebugContext(ctx, "ENTER createBatch", "name", name)

	url := xaiBaseURL + xaiBatchesEndpoint
	type createReq struct {
		Endpoint         string `json:"endpoint"`
		CompletionWindow string `json:"completion_window"`
		Name             string `json:"name"`
	}
	body := createReq{
		Endpoint:         xaiChatEndpoint,
		CompletionWindow: "24h",
		Name:             name,
	}

	bodyBytes, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("xai: createBatch request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("xai: createBatch request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("xai: createBatch error %d: %s", httpResp.StatusCode, string(respBody))
	}

	var result struct {
		BatchID string `json:"batch_id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("xai: createBatch unmarshal failed: %w", err)
	}
	if result.BatchID == "" {
		return "", fmt.Errorf("xai: createBatch: no batch_id in response: %s", string(respBody))
	}

	b.logger.DebugContext(ctx, "EXIT createBatch", "batch_id", result.BatchID)
	return result.BatchID, nil
}

// addRequests は batchID にリクエストを xAI 独自形式で追加する。
func (b *xaiBatchClient) addRequests(ctx context.Context, batchID string, reqs []Request, startIdx int) error {
	b.logger.DebugContext(ctx, "ENTER addRequests",
		"batch_id", batchID,
		"count", len(reqs),
		"start_idx", startIdx,
	)

	url := fmt.Sprintf("%s%s/%s/requests", xaiBaseURL, xaiBatchesEndpoint, batchID)

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type chatCompletion struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float32   `json:"temperature,omitempty"`
		MaxTokens   int       `json:"max_tokens,omitempty"`
	}
	type batchRequest struct {
		BatchRequestID string `json:"batch_request_id"`
		BatchRequest   struct {
			ChatGetCompletion chatCompletion `json:"chat_get_completion"`
		} `json:"batch_request"`
	}

	batchReqs := make([]batchRequest, 0, len(reqs))
	for i, req := range reqs {
		msgs := []message{}
		if req.SystemPrompt != "" {
			msgs = append(msgs, message{Role: "system", Content: req.SystemPrompt})
		}
		msgs = append(msgs, message{Role: "user", Content: req.UserPrompt})

		br := batchRequest{
			BatchRequestID: fmt.Sprintf("req-%d", startIdx+i),
		}
		br.BatchRequest.ChatGetCompletion = chatCompletion{
			Model:       b.config.Model,
			Messages:    msgs,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		}
		batchReqs = append(batchReqs, br)
	}

	payload := map[string]interface{}{"batch_requests": batchReqs}
	bodyBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("xai: addRequests request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("xai: addRequests request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("xai: addRequests error %d: %s", httpResp.StatusCode, string(respBody))
	}

	b.logger.DebugContext(ctx, "EXIT addRequests", "batch_id", batchID, "added", len(reqs))
	return nil
}

// pollUntilCompleted はバッチジョブが completed になるまで polling する。
func (b *xaiBatchClient) pollUntilCompleted(ctx context.Context, batchID string) error {
	b.logger.DebugContext(ctx, "ENTER pollUntilCompleted", "batch_id", batchID)

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("xai: polling cancelled: %w", err)
		}

		status, err := b.GetBatchStatus(ctx, BatchJobID{ID: batchID, Provider: "xai"})
		if err != nil {
			return fmt.Errorf("xai: polling status check failed: %w", err)
		}

		switch status.State {
		case "completed":
			b.logger.DebugContext(ctx, "EXIT pollUntilCompleted", "batch_id", batchID, "state", "completed")
			return nil
		case "failed", "cancelled":
			return fmt.Errorf("xai: batch %s ended with state %q", batchID, status.State)
		}

		b.logger.DebugContext(ctx, "polling...",
			"batch_id", batchID,
			"state", status.State,
			"progress", status.Progress,
		)

		select {
		case <-ctx.Done():
			return fmt.Errorf("xai: polling cancelled: %w", ctx.Err())
		case <-time.After(xaiPollInterval):
		}
	}
}

// parseBatchStatus は GET /v1/batches/{id} のレスポンスを BatchStatus に変換する。
func (b *xaiBatchClient) parseBatchStatus(ctx context.Context, body []byte, batchID string) (BatchStatus, error) {
	b.logger.DebugContext(ctx, "ENTER parseBatchStatus", "batch_id", batchID)

	var raw struct {
		CancelTime *string `json:"cancel_time"`
		State      struct {
			NumRequests  int `json:"num_requests"`
			NumPending   int `json:"num_pending"`
			NumSuccess   int `json:"num_success"`
			NumError     int `json:"num_error"`
			NumCancelled int `json:"num_cancelled"`
		} `json:"state"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return BatchStatus{}, fmt.Errorf("xai: parseBatchStatus unmarshal failed: %w", err)
	}

	var state string
	switch {
	case raw.CancelTime != nil:
		state = "cancelled"
	case raw.State.NumRequests == 0:
		state = "validating"
	case raw.State.NumPending > 0:
		state = "in_progress"
	case raw.State.NumError > 0 && raw.State.NumSuccess == 0:
		state = "failed"
	default:
		state = "completed"
	}

	total := raw.State.NumRequests
	var progress float32
	if total > 0 {
		progress = float32(raw.State.NumSuccess+raw.State.NumError) / float32(total)
	}

	status := BatchStatus{
		ID:       batchID,
		State:    state,
		Progress: progress,
	}

	b.logger.DebugContext(ctx, "EXIT parseBatchStatus",
		"batch_id", batchID,
		"state", state,
		"progress", progress,
	)
	return status, nil
}

// fetchResultPage は GET /v1/batches/{id}/results の1ページ分を取得する。
func (b *xaiBatchClient) fetchResultPage(ctx context.Context, batchID, paginationToken string) ([]Response, string, error) {
	b.logger.DebugContext(ctx, "ENTER fetchResultPage",
		"batch_id", batchID,
		"has_token", paginationToken != "",
	)

	url := fmt.Sprintf("%s%s/%s/results", xaiBaseURL, xaiBatchesEndpoint, batchID)
	if paginationToken != "" {
		url += "?pagination_token=" + paginationToken
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("xai: fetchResultPage request creation failed: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+b.config.APIKey)

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("xai: fetchResultPage request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("xai: fetchResultPage error %d: %s", httpResp.StatusCode, string(body))
	}

	results, nextToken, err := b.parseResults(ctx, body)
	if err != nil {
		return nil, "", err
	}

	b.logger.DebugContext(ctx, "EXIT fetchResultPage",
		"results", len(results),
		"has_next_token", nextToken != "",
	)
	return results, nextToken, nil
}

// parseResults は GET /v1/batches/{id}/results のレスポンスをパースして []Response と次ページトークンを返す。
func (b *xaiBatchClient) parseResults(ctx context.Context, body []byte) ([]Response, string, error) {
	b.logger.DebugContext(ctx, "ENTER parseResults", "body_len", len(body))

	var raw struct {
		Results []struct {
			BatchRequestID string `json:"batch_request_id"`
			Error          *struct {
				Message string `json:"message"`
			} `json:"error"`
			BatchResult struct {
				Response struct {
					ChatGetCompletion struct {
						Choices []struct {
							Message struct {
								Content string `json:"content"`
							} `json:"message"`
						} `json:"choices"`
						Usage struct {
							PromptTokens     int `json:"prompt_tokens"`
							CompletionTokens int `json:"completion_tokens"`
							TotalTokens      int `json:"total_tokens"`
						} `json:"usage"`
					} `json:"chat_get_completion"`
				} `json:"response"`
			} `json:"batch_result"`
		} `json:"results"`
		PaginationToken string `json:"pagination_token"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, "", fmt.Errorf("xai: parseResults unmarshal failed: %w", err)
	}

	responses := make([]Response, 0, len(raw.Results))
	for _, r := range raw.Results {
		if r.Error != nil {
			responses = append(responses, Response{
				Content: "",
				Success: false,
				Error:   r.Error.Message,
			})
			continue
		}

		completion := r.BatchResult.Response.ChatGetCompletion
		if len(completion.Choices) == 0 {
			responses = append(responses, Response{
				Content: "",
				Success: false,
				Error:   "no choices in response",
			})
			continue
		}

		responses = append(responses, Response{
			Content: completion.Choices[0].Message.Content,
			Success: true,
			Usage: TokenUsage{
				PromptTokens:     completion.Usage.PromptTokens,
				CompletionTokens: completion.Usage.CompletionTokens,
				TotalTokens:      completion.Usage.TotalTokens,
			},
		})
	}

	b.logger.DebugContext(ctx, "EXIT parseResults",
		"parsed_count", len(responses),
		"next_token", raw.PaginationToken,
	)
	return responses, raw.PaginationToken, nil
}

// xaiStreamResponse — フォールバック用
type xaiStreamResponse struct {
	resp Response
	done bool
}

func (s *xaiStreamResponse) Next() (Response, bool) {
	if s.done {
		return Response{}, false
	}
	s.done = true
	return s.resp, true
}

func (s *xaiStreamResponse) Close() error { return nil }
