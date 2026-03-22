package llmexec

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
)

const batchPollingInterval = 1 * time.Second

// SyncExecutor runs workflow-level synchronous LLM requests through the gateway layer.
type SyncExecutor struct {
	llmManager gatewayllm.LLMManager
}

// NewSyncExecutor creates a runtime adapter for synchronous LLM execution.
func NewSyncExecutor(llmManager gatewayllm.LLMManager) *SyncExecutor {
	return &SyncExecutor{llmManager: llmManager}
}

// Execute resolves a client and executes requests in input order.
func (e *SyncExecutor) Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error) {
	return e.ExecuteWithProgress(ctx, config, requests, nil)
}

// ExecuteWithProgress resolves a client and executes requests while reporting completed count.
func (e *SyncExecutor) ExecuteWithProgress(
	ctx context.Context,
	config llmio.ExecutionConfig,
	requests []llmio.Request,
	progress func(completed, total int),
) ([]llmio.Response, error) {
	llmConfig := gatewayllm.LLMConfig{
		Provider: config.Provider,
		APIKey:   config.APIKey,
		Endpoint: config.Endpoint,
		Model:    config.Model,
		Parameters: map[string]interface{}{
			"context_length": config.ContextLength,
		},
		Concurrency: config.SyncConcurrency,
	}
	strategy := resolveBulkStrategy(config.BulkStrategy)
	resolvedStrategy := e.llmManager.ResolveBulkStrategy(ctx, strategy, config.Provider)
	if resolvedStrategy == gatewayllm.BulkStrategyBatch {
		return e.executeBatch(ctx, llmConfig, requests, progress)
	}
	return e.executeSync(ctx, llmConfig, requests, progress)
}

func (e *SyncExecutor) executeSync(
	ctx context.Context,
	llmConfig gatewayllm.LLMConfig,
	requests []llmio.Request,
	progress func(completed, total int),
) ([]llmio.Response, error) {
	client, err := e.llmManager.GetClient(ctx, llmConfig)
	if err != nil {
		return nil, fmt.Errorf("create llm client: %w", err)
	}

	var lifecycleClient gatewayllm.ModelLifecycleClient
	instanceID := ""
	if lifecycle, ok := client.(gatewayllm.ModelLifecycleClient); ok {
		lifecycleClient = lifecycle
		ctxLen := extractContextLength(llmConfig.Parameters)
		var loadErr error
		instanceID, loadErr = lifecycleClient.LoadModel(ctx, llmConfig.Model, ctxLen)
		if loadErr != nil {
			return nil, fmt.Errorf("load model: %w", loadErr)
		}
	}

	gatewayReqs := toGatewayRequests(requests, false)

	responses, err := gatewayllm.ExecuteBulkSyncWithProgress(ctx, client, gatewayReqs, llmConfig.Concurrency, progress)
	if lifecycleClient != nil {
		unloadCtx := context.WithoutCancel(ctx)
		if unloadErr := lifecycleClient.UnloadModel(unloadCtx, instanceID); unloadErr != nil {
			unloadWrapped := fmt.Errorf("unload model: %w", unloadErr)
			if err != nil {
				return nil, errors.Join(err, unloadWrapped)
			}
			return nil, unloadWrapped
		}
	}
	if err != nil {
		return nil, fmt.Errorf("execute bulk sync: %w", err)
	}

	return toExecutionResponses(responses), nil
}

func extractContextLength(parameters map[string]interface{}) int {
	if len(parameters) == 0 {
		return 0
	}
	raw, ok := parameters["context_length"]
	if !ok {
		return 0
	}
	switch value := raw.(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func (e *SyncExecutor) executeBatch(
	ctx context.Context,
	llmConfig gatewayllm.LLMConfig,
	requests []llmio.Request,
	progress func(completed, total int),
) ([]llmio.Response, error) {
	batchClient, err := e.llmManager.GetBatchClient(ctx, llmConfig)
	if err != nil {
		return nil, fmt.Errorf("create terminology batch client: %w", err)
	}

	gatewayReqs, requestOrder := toBatchGatewayRequests(requests)
	jobID, err := batchClient.SubmitBatch(ctx, gatewayReqs)
	if err != nil {
		return nil, fmt.Errorf("submit batch: %w", err)
	}
	if progress != nil {
		progress(0, len(requests))
	}

	for {
		status, statusErr := batchClient.GetBatchStatus(ctx, jobID)
		if statusErr != nil {
			return nil, fmt.Errorf("poll batch status: %w", statusErr)
		}
		if progress != nil {
			progress(batchProgressToCompleted(status.Progress, len(requests)), len(requests))
		}
		if isTerminalBatchState(status.State) {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(batchPollingInterval):
		}
	}

	responses, err := batchClient.GetBatchResults(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("fetch batch results: %w", err)
	}
	if progress != nil {
		progress(len(requests), len(requests))
	}

	return toExecutionResponses(reorderBatchResponses(responses, requests, requestOrder)), nil
}

func toGatewayRequests(requests []llmio.Request, attachSequence bool) []gatewayllm.Request {
	gatewayReqs := make([]gatewayllm.Request, 0, len(requests))
	for idx, request := range requests {
		metadata := copyMetadata(request.Metadata)
		if attachSequence {
			metadata[gatewayllm.BatchMetadataQueueRequestSeqKey] = idx
		}
		gatewayReqs = append(gatewayReqs, gatewayllm.Request{
			SystemPrompt:   request.SystemPrompt,
			UserPrompt:     request.UserPrompt,
			Temperature:    request.Temperature,
			ResponseSchema: request.ResponseSchema,
			StopSequences:  request.StopSequences,
			Metadata:       metadata,
		})
	}
	return gatewayReqs
}

func toBatchGatewayRequests(requests []llmio.Request) ([]gatewayllm.Request, []string) {
	gatewayReqs := make([]gatewayllm.Request, 0, len(requests))
	requestOrder := make([]string, 0, len(requests))
	for idx, request := range requests {
		metadata := copyMetadata(request.Metadata)
		queueJobID := readQueueJobID(metadata)
		if queueJobID == "" {
			queueJobID = buildTerminologyQueueJobID(idx)
		}
		metadata[gatewayllm.BatchMetadataQueueJobIDKey] = queueJobID
		metadata[gatewayllm.BatchMetadataQueueRequestSeqKey] = idx
		requestOrder = append(requestOrder, queueJobID)
		gatewayReqs = append(gatewayReqs, gatewayllm.Request{
			SystemPrompt:   request.SystemPrompt,
			UserPrompt:     request.UserPrompt,
			Temperature:    request.Temperature,
			ResponseSchema: request.ResponseSchema,
			StopSequences:  request.StopSequences,
			Metadata:       metadata,
		})
	}
	return gatewayReqs, requestOrder
}

func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func buildTerminologyQueueJobID(idx int) string {
	return fmt.Sprintf("terminology-%d", idx)
}

func readQueueJobID(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return ""
	}
	raw, exists := metadata[gatewayllm.BatchMetadataQueueJobIDKey]
	if !exists {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", raw))
}

func toExecutionResponses(responses []gatewayllm.Response) []llmio.Response {
	result := make([]llmio.Response, 0, len(responses))
	for _, response := range responses {
		result = append(result, llmio.Response{
			Content:  response.Content,
			Success:  response.Success,
			Error:    response.Error,
			Usage:    llmio.TokenUsage(response.Usage),
			Metadata: response.Metadata,
		})
	}
	return result
}

func resolveBulkStrategy(raw string) gatewayllm.BulkStrategy {
	if strings.EqualFold(strings.TrimSpace(raw), string(gatewayllm.BulkStrategyBatch)) {
		return gatewayllm.BulkStrategyBatch
	}
	return gatewayllm.BulkStrategySync
}

func isTerminalBatchState(state gatewayllm.BatchState) bool {
	switch state {
	case gatewayllm.BatchStateCompleted,
		gatewayllm.BatchStatePartialFailed,
		gatewayllm.BatchStateFailed,
		gatewayllm.BatchStateCancelled:
		return true
	default:
		return false
	}
}

func batchProgressToCompleted(progress float32, total int) int {
	if total <= 0 {
		return 0
	}
	completed := int(progress * float32(total))
	if completed < 0 {
		return 0
	}
	if completed > total {
		return total
	}
	return completed
}

func reorderBatchResponses(responses []gatewayllm.Response, requests []llmio.Request, requestOrder []string) []gatewayllm.Response {
	if len(requests) == 0 {
		return nil
	}
	ordered := make([]gatewayllm.Response, len(requests))
	filled := make([]bool, len(requests))
	fallback := make([]gatewayllm.Response, 0, len(responses))
	idToIndex := make(map[string]int, len(requestOrder))
	for idx, id := range requestOrder {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := idToIndex[trimmed]; !exists {
			idToIndex[trimmed] = idx
		}
	}

	for _, response := range responses {
		if idx, ok := findResponseIndex(response.Metadata, idToIndex); ok && !filled[idx] {
			ordered[idx] = response
			filled[idx] = true
			continue
		}
		seq, ok := extractRequestSeq(response.Metadata)
		if !ok || seq < 0 || seq >= len(requests) || filled[seq] {
			fallback = append(fallback, response)
			continue
		}
		ordered[seq] = response
		filled[seq] = true
	}

	fallbackIndex := 0
	for idx := range ordered {
		if filled[idx] {
			ordered[idx].Metadata = mergeMetadata(requests[idx].Metadata, ordered[idx].Metadata)
			continue
		}
		if fallbackIndex < len(fallback) {
			ordered[idx] = fallback[fallbackIndex]
			ordered[idx].Metadata = mergeMetadata(requests[idx].Metadata, ordered[idx].Metadata)
			fallbackIndex++
			continue
		}
		ordered[idx] = gatewayllm.Response{
			Success:  false,
			Error:    "batch result missing for request",
			Metadata: copyMetadata(requests[idx].Metadata),
		}
	}
	return ordered
}

func findResponseIndex(metadata map[string]interface{}, idToIndex map[string]int) (int, bool) {
	id := readQueueJobID(metadata)
	if id == "" {
		return 0, false
	}
	idx, ok := idToIndex[id]
	return idx, ok
}

func mergeMetadata(base map[string]interface{}, overlay map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(overlay) == 0 {
		return map[string]interface{}{}
	}
	merged := copyMetadata(base)
	for key, value := range overlay {
		merged[key] = value
	}
	return merged
}

func extractRequestSeq(metadata map[string]interface{}) (int, bool) {
	if len(metadata) == 0 {
		return 0, false
	}
	raw, exists := metadata[gatewayllm.BatchMetadataQueueRequestSeqKey]
	if !exists {
		return 0, false
	}
	switch value := raw.(type) {
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case float64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
