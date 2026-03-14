package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

// geminiBatchClient は Gemini Batch API の BatchClient 実装。
type geminiBatchClient struct {
	config     LLMConfig
	httpClient *http.Client
	logger     *slog.Logger
}

// NewGeminiBatchClient は Gemini BatchClient を返す。
func NewGeminiBatchClient(logger *slog.Logger, config LLMConfig) (BatchClient, error) {
	if strings.TrimSpace(config.Model) == "" {
		return nil, ErrModelRequired
	}

	return &geminiBatchClient{
		config:     config,
		httpClient: &http.Client{Timeout: geminiDefaultTimeout},
		logger:     logger.With("component", "gemini_batch_client", "model", config.Model),
	}, nil
}

// SubmitBatch は inlined requests を Gemini Batch API へ送信し BatchJobID を返す。
func (b *geminiBatchClient) SubmitBatch(ctx context.Context, reqs []Request) (BatchJobID, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionLLMRequest)()
	if len(reqs) == 0 {
		return BatchJobID{}, fmt.Errorf("gemini: no requests to submit")
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Role  string `json:"role,omitempty"`
		Parts []part `json:"parts"`
	}
	type generationConfig struct {
		Temperature float32 `json:"temperature,omitempty"`
	}
	type generateContentRequest struct {
		Contents          []content        `json:"contents"`
		SystemInstruction *content         `json:"systemInstruction,omitempty"`
		GenerationConfig  generationConfig `json:"generationConfig,omitempty"`
	}
	type inlinedRequest struct {
		Request  generateContentRequest `json:"request"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	type requestBody struct {
		Batch struct {
			DisplayName string `json:"displayName"`
			InputConfig struct {
				Requests struct {
					Requests []inlinedRequest `json:"requests"`
				} `json:"requests"`
			} `json:"inputConfig"`
		} `json:"batch"`
	}

	body := requestBody{}
	body.Batch.DisplayName = fmt.Sprintf("batch-%d", time.Now().UTC().Unix())

	for _, req := range reqs {
		inlineReq := inlinedRequest{}
		inlineReq.Request.Contents = []content{{
			Role:  "user",
			Parts: []part{{Text: req.UserPrompt}},
		}}
		if req.SystemPrompt != "" {
			inlineReq.Request.SystemInstruction = &content{
				Role:  "user",
				Parts: []part{{Text: req.SystemPrompt}},
			}
		}
		inlineReq.Request.GenerationConfig = generationConfig{Temperature: req.Temperature}
		if len(req.Metadata) > 0 {
			inlineReq.Metadata = req.Metadata
		}
		body.Batch.InputConfig.Requests.Requests = append(body.Batch.InputConfig.Requests.Requests, inlineReq)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return BatchJobID{}, fmt.Errorf("gemini: failed to marshal submit batch request: %w", err)
	}

	modelPath := normalizeGeminiModelResource(b.config.Model)
	url := fmt.Sprintf("%s/%s/%s:batchGenerateContent?key=%s", geminiBaseURL, geminiAPIVersion, modelPath, b.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return BatchJobID{}, fmt.Errorf("gemini: submit batch request creation failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return BatchJobID{}, fmt.Errorf("gemini: submit batch request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return BatchJobID{}, fmt.Errorf("gemini: submit batch response read failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		return BatchJobID{}, fmt.Errorf("gemini: submit batch status %d: %s", httpResp.StatusCode, string(respBody))
	}

	batchName, err := extractGeminiBatchName(respBody)
	if err != nil {
		return BatchJobID{}, fmt.Errorf("gemini: failed to parse batch id: %w", err)
	}

	return BatchJobID{ID: batchName, Provider: "gemini"}, nil
}

// GetBatchStatus は batch state と batchStats から共通 BatchStatus を返す。
func (b *geminiBatchClient) GetBatchStatus(ctx context.Context, id BatchJobID) (BatchStatus, error) {
	batchName := normalizeGeminiBatchResource(id.ID)
	url := fmt.Sprintf("%s/%s/%s?key=%s", geminiBaseURL, geminiAPIVersion, batchName, b.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("gemini: get batch status request creation failed: %w", err)
	}

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("gemini: get batch status request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("gemini: get batch status response read failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		return BatchStatus{}, fmt.Errorf("gemini: get batch status status %d: %s", httpResp.StatusCode, string(respBody))
	}

	status, err := parseGeminiBatchStatus(respBody, batchName)
	if err != nil {
		return BatchStatus{}, fmt.Errorf("gemini: parse batch status failed: %w", err)
	}
	return status, nil
}

// GetBatchResults は inlined responses を共通 Response へ変換して返す。
func (b *geminiBatchClient) GetBatchResults(ctx context.Context, id BatchJobID) ([]Response, error) {
	batchName := normalizeGeminiBatchResource(id.ID)
	url := fmt.Sprintf("%s/%s/%s?key=%s", geminiBaseURL, geminiAPIVersion, batchName, b.config.APIKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini: get batch results request creation failed: %w", err)
	}

	httpResp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: get batch results request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: get batch results response read failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini: get batch results status %d: %s", httpResp.StatusCode, string(respBody))
	}

	results, err := parseGeminiBatchResults(respBody)
	if err != nil {
		return nil, fmt.Errorf("gemini: parse batch results failed: %w", err)
	}
	return results, nil
}

type geminiBatchStats struct {
	RequestCount           string `json:"requestCount"`
	SuccessfulRequestCount string `json:"successfulRequestCount"`
	FailedRequestCount     string `json:"failedRequestCount"`
	PendingRequestCount    string `json:"pendingRequestCount"`
}

type geminiBatchOutput struct {
	ResponsesFile    string `json:"responsesFile"`
	InlinedResponses struct {
		InlinedResponses []geminiInlinedResponse `json:"inlinedResponses"`
	} `json:"inlinedResponses"`
}

type geminiBatchResource struct {
	Name       string            `json:"name"`
	State      string            `json:"state"`
	BatchStats geminiBatchStats  `json:"batchStats"`
	Output     geminiBatchOutput `json:"output"`
	Batch      struct {
		Name       string            `json:"name"`
		State      string            `json:"state"`
		BatchStats geminiBatchStats  `json:"batchStats"`
		Output     geminiBatchOutput `json:"output"`
	} `json:"batch"`
}

type geminiBatchOperation struct {
	Name       string            `json:"name"`
	Done       bool              `json:"done"`
	State      string            `json:"state"`
	BatchStats geminiBatchStats  `json:"batchStats"`
	Output     geminiBatchOutput `json:"output"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Batch    geminiBatchResource `json:"batch"`
	Metadata geminiBatchResource `json:"metadata"`
	Response geminiBatchResource `json:"response"`
}

type geminiInlinedResponse struct {
	Metadata map[string]interface{} `json:"metadata"`
	Error    *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Response *struct {
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
	} `json:"response"`
}

func extractGeminiBatchName(body []byte) (string, error) {
	var raw geminiBatchOperation
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("unmarshal failed: %w", err)
	}

	candidates := []string{
		raw.Batch.Name,
		raw.Response.Batch.Name,
		raw.Response.Name,
		raw.Metadata.Batch.Name,
		raw.Metadata.Name,
		raw.Name,
	}
	for _, candidate := range candidates {
		normalized := normalizeGeminiBatchResource(candidate)
		if strings.HasPrefix(normalized, "batches/") {
			return normalized, nil
		}
	}

	return "", fmt.Errorf("batch name not found in submit response")
}

func parseGeminiBatchStatus(body []byte, batchName string) (BatchStatus, error) {
	var raw geminiBatchOperation
	if err := json.Unmarshal(body, &raw); err != nil {
		return BatchStatus{}, fmt.Errorf("unmarshal failed: %w", err)
	}

	state := firstNonEmpty(
		raw.Batch.State,
		raw.Response.Batch.State,
		raw.Response.State,
		raw.Metadata.Batch.State,
		raw.Metadata.State,
		raw.State,
	)
	stats := firstNonZeroStats(
		raw.Batch.BatchStats,
		raw.Response.Batch.BatchStats,
		raw.Response.BatchStats,
		raw.Metadata.Batch.BatchStats,
		raw.Metadata.BatchStats,
		raw.BatchStats,
	)

	requestCount := parseGeminiInt(stats.RequestCount)
	successCount := parseGeminiInt(stats.SuccessfulRequestCount)
	failedCount := parseGeminiInt(stats.FailedRequestCount)
	pendingCount := parseGeminiInt(stats.PendingRequestCount)

	normalizedState := normalizeGeminiBatchState(state, failedCount, raw.Done, raw.Error != nil)
	progress := computeGeminiBatchProgress(normalizedState, requestCount, successCount, failedCount, pendingCount)

	return BatchStatus{
		ID:       batchName,
		State:    normalizedState,
		Progress: progress,
	}, nil
}

func parseGeminiBatchResults(body []byte) ([]Response, error) {
	var raw geminiBatchOperation
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	output := firstNonEmptyOutput(
		raw.Batch.Output,
		raw.Response.Batch.Output,
		raw.Response.Output,
		raw.Metadata.Batch.Output,
		raw.Metadata.Output,
		raw.Output,
	)

	if len(output.InlinedResponses.InlinedResponses) == 0 {
		if output.ResponsesFile != "" {
			return nil, fmt.Errorf("responsesFile output is not supported yet")
		}
		return []Response{}, nil
	}

	results := make([]Response, 0, len(output.InlinedResponses.InlinedResponses))
	for _, item := range output.InlinedResponses.InlinedResponses {
		if item.Error != nil {
			errMsg := strings.TrimSpace(item.Error.Message)
			if errMsg == "" {
				errMsg = "gemini: batch request failed"
			}
			results = append(results, Response{
				Success:  false,
				Error:    errMsg,
				Metadata: item.Metadata,
			})
			continue
		}

		if item.Response == nil || len(item.Response.Candidates) == 0 || len(item.Response.Candidates[0].Content.Parts) == 0 {
			results = append(results, Response{
				Success:  false,
				Error:    "gemini: empty inlined response",
				Metadata: item.Metadata,
			})
			continue
		}

		results = append(results, Response{
			Content: item.Response.Candidates[0].Content.Parts[0].Text,
			Success: true,
			Usage: TokenUsage{
				PromptTokens:     item.Response.UsageMetadata.PromptTokenCount,
				CompletionTokens: item.Response.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      item.Response.UsageMetadata.TotalTokenCount,
			},
			Metadata: item.Metadata,
		})
	}

	return results, nil
}

func normalizeGeminiBatchState(rawState string, failedCount int, done bool, hasError bool) BatchState {
	switch strings.ToUpper(strings.TrimSpace(rawState)) {
	case "BATCH_STATE_PENDING", "BATCH_STATE_UNSPECIFIED":
		return BatchStateQueued
	case "BATCH_STATE_RUNNING":
		return BatchStateRunning
	case "BATCH_STATE_SUCCEEDED":
		if failedCount > 0 {
			return BatchStatePartialFailed
		}
		return BatchStateCompleted
	case "BATCH_STATE_FAILED", "BATCH_STATE_EXPIRED":
		return BatchStateFailed
	case "BATCH_STATE_CANCELLED":
		return BatchStateCancelled
	}

	if hasError {
		return BatchStateFailed
	}
	if done {
		if failedCount > 0 {
			return BatchStatePartialFailed
		}
		return BatchStateCompleted
	}
	return BatchStateRunning
}

func computeGeminiBatchProgress(state BatchState, total, successful, failed, pending int) float32 {
	if total > 0 {
		completedLike := successful + failed

		// v1main payloads may omit successful/failed counts while pending decreases.
		if pending >= 0 && pending <= total {
			completedFromPending := total - pending
			if completedFromPending > completedLike {
				completedLike = completedFromPending
			}
		}
		if pending == 0 {
			switch state {
			case BatchStateCompleted, BatchStatePartialFailed, BatchStateFailed, BatchStateCancelled:
				completedLike = total
			}
		}

		if completedLike < 0 {
			completedLike = 0
		}
		if completedLike > total {
			completedLike = total
		}
		return float32(completedLike) / float32(total)
	}

	switch state {
	case BatchStateCompleted, BatchStatePartialFailed, BatchStateFailed, BatchStateCancelled:
		return 1
	default:
		return 0
	}
}

func parseGeminiInt(value string) int {
	v, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return v
}

func normalizeGeminiModelResource(model string) string {
	trimmed := strings.TrimSpace(model)
	if strings.HasPrefix(trimmed, "models/") {
		return trimmed
	}
	return "models/" + trimmed
}

func normalizeGeminiBatchResource(batchID string) string {
	trimmed := strings.TrimSpace(batchID)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "batches/") {
		return trimmed
	}
	if idx := strings.Index(trimmed, "batches/"); idx >= 0 {
		return trimmed[idx:]
	}
	return "batches/" + trimmed
}

func firstNonEmpty(candidates ...string) string {
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstNonZeroStats(candidates ...geminiBatchStats) geminiBatchStats {
	for _, candidate := range candidates {
		if candidate.RequestCount != "" || candidate.SuccessfulRequestCount != "" || candidate.FailedRequestCount != "" || candidate.PendingRequestCount != "" {
			return candidate
		}
	}
	return geminiBatchStats{}
}

func firstNonEmptyOutput(candidates ...geminiBatchOutput) geminiBatchOutput {
	for _, candidate := range candidates {
		if candidate.ResponsesFile != "" || len(candidate.InlinedResponses.InlinedResponses) > 0 {
			return candidate
		}
	}
	return geminiBatchOutput{}
}
