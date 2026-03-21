package llmexec

import (
	"context"
	"fmt"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
)

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
	client, err := e.llmManager.GetClient(ctx, gatewayllm.LLMConfig{
		Provider: config.Provider,
		APIKey:   config.APIKey,
		Endpoint: config.Endpoint,
		Model:    config.Model,
		Parameters: map[string]interface{}{
			"context_length": config.ContextLength,
		},
		Concurrency: config.SyncConcurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("create terminology llm client: %w", err)
	}

	gatewayReqs := make([]gatewayllm.Request, 0, len(requests))
	for _, request := range requests {
		gatewayReqs = append(gatewayReqs, gatewayllm.Request{
			SystemPrompt:   request.SystemPrompt,
			UserPrompt:     request.UserPrompt,
			Temperature:    request.Temperature,
			ResponseSchema: request.ResponseSchema,
			StopSequences:  request.StopSequences,
			Metadata:       request.Metadata,
		})
	}

	responses, err := gatewayllm.ExecuteBulkSyncWithProgress(ctx, client, gatewayReqs, config.SyncConcurrency, progress)
	if err != nil {
		return nil, fmt.Errorf("execute bulk sync: %w", err)
	}

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
	return result, nil
}
