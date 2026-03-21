package llmexec

import (
	"context"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
)

func TestSyncExecutorExecuteWithProgress_UsesQueueJobIDContractAndBatchClient(t *testing.T) {
	batchClient := &stubBatchClient{
		statuses: []gatewayllm.BatchStatus{
			{State: gatewayllm.BatchStateCompleted, Progress: 1.0},
		},
		results: []gatewayllm.Response{
			{
				Success:  true,
				Content:  "res-b",
				Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-1"},
			},
			{
				Success:  true,
				Content:  "res-a",
				Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-0"},
			},
		},
	}
	manager := &stubLLMManager{
		bulkStrategy: gatewayllm.BulkStrategyBatch,
		batchClient:  batchClient,
	}
	executor := NewSyncExecutor(manager)

	requests := []llmio.Request{
		{Metadata: map[string]interface{}{"source_text": "A"}},
		{Metadata: map[string]interface{}{"source_text": "B"}},
	}
	progressLog := make([]int, 0)
	responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
		Provider:     "xai",
		Model:        "grok-3",
		BulkStrategy: "batch",
	}, requests, func(completed, total int) {
		_ = total
		progressLog = append(progressLog, completed)
	})
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	if manager.getBatchClientCalls != 1 {
		t.Fatalf("expected GetBatchClient to be called once, got=%d", manager.getBatchClientCalls)
	}
	if manager.getClientCalls != 0 {
		t.Fatalf("expected GetClient not to be called, got=%d", manager.getClientCalls)
	}
	if batchClient.submitCalls != 1 {
		t.Fatalf("expected SubmitBatch to be called once, got=%d", batchClient.submitCalls)
	}
	if len(batchClient.submittedRequests) != 2 {
		t.Fatalf("submitted requests = %d, want 2", len(batchClient.submittedRequests))
	}
	if got := batchClient.submittedRequests[0].Metadata[gatewayllm.BatchMetadataQueueJobIDKey]; got != "terminology-0" {
		t.Fatalf("request[0] queue_job_id = %v, want terminology-0", got)
	}
	if got := batchClient.submittedRequests[1].Metadata[gatewayllm.BatchMetadataQueueJobIDKey]; got != "terminology-1" {
		t.Fatalf("request[1] queue_job_id = %v, want terminology-1", got)
	}
	if got := batchClient.submittedRequests[0].Metadata[gatewayllm.BatchMetadataQueueRequestSeqKey]; got != 0 {
		t.Fatalf("request[0] queue_request_seq = %v, want 0", got)
	}
	if got := batchClient.submittedRequests[1].Metadata[gatewayllm.BatchMetadataQueueRequestSeqKey]; got != 1 {
		t.Fatalf("request[1] queue_request_seq = %v, want 1", got)
	}
	if len(responses) != 2 {
		t.Fatalf("response count = %d, want 2", len(responses))
	}
	if responses[0].Content != "res-a" || responses[1].Content != "res-b" {
		t.Fatalf("unexpected response order: [%q, %q]", responses[0].Content, responses[1].Content)
	}
	if got := responses[0].Metadata["source_text"]; got != "A" {
		t.Fatalf("response[0] source_text = %v, want A", got)
	}
	if got := responses[1].Metadata["source_text"]; got != "B" {
		t.Fatalf("response[1] source_text = %v, want B", got)
	}
	assertProgressEndsAt(t, progressLog, 2)
}

func TestSyncExecutorExecuteWithProgress_BatchTerminalStatesAlwaysCompleteProgress(t *testing.T) {
	terminalStates := []gatewayllm.BatchState{
		gatewayllm.BatchStateCompleted,
		gatewayllm.BatchStatePartialFailed,
		gatewayllm.BatchStateFailed,
		gatewayllm.BatchStateCancelled,
	}
	for _, state := range terminalStates {
		t.Run(string(state), func(t *testing.T) {
			batchClient := &stubBatchClient{
				statuses: []gatewayllm.BatchStatus{
					{State: state, Progress: 0.6},
				},
				results: []gatewayllm.Response{
					{
						Success:  true,
						Content:  "ok",
						Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-0"},
					},
				},
			}
			manager := &stubLLMManager{
				bulkStrategy: gatewayllm.BulkStrategyBatch,
				batchClient:  batchClient,
			}
			executor := NewSyncExecutor(manager)
			progressLog := make([]int, 0)
			responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
				Provider:     "xai",
				Model:        "grok-3",
				BulkStrategy: "batch",
			}, []llmio.Request{{Metadata: map[string]interface{}{"source_text": "A"}}}, func(completed, total int) {
				_ = total
				progressLog = append(progressLog, completed)
			})
			if err != nil {
				t.Fatalf("ExecuteWithProgress failed: %v", err)
			}
			if len(responses) != 1 || !responses[0].Success {
				t.Fatalf("unexpected responses: %+v", responses)
			}
			assertProgressEndsAt(t, progressLog, 1)
		})
	}
}

func TestSyncExecutorExecuteWithProgress_BatchMissingMetadataUsesFallbackOrder(t *testing.T) {
	batchClient := &stubBatchClient{
		statuses: []gatewayllm.BatchStatus{
			{State: gatewayllm.BatchStateCompleted, Progress: 1.0},
		},
		results: []gatewayllm.Response{
			{Success: true, Content: "first", Metadata: map[string]interface{}{}},
			{Success: true, Content: "second", Metadata: map[string]interface{}{}},
		},
	}
	manager := &stubLLMManager{
		bulkStrategy: gatewayllm.BulkStrategyBatch,
		batchClient:  batchClient,
	}
	executor := NewSyncExecutor(manager)
	progressLog := make([]int, 0)
	responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
		Provider:     "xai",
		Model:        "grok-3",
		BulkStrategy: "batch",
	}, []llmio.Request{
		{Metadata: map[string]interface{}{"source_text": "A"}},
		{Metadata: map[string]interface{}{"source_text": "B"}},
	}, func(completed, total int) {
		_ = total
		progressLog = append(progressLog, completed)
	})
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("response count = %d, want 2", len(responses))
	}
	if responses[0].Content != "first" || responses[1].Content != "second" {
		t.Fatalf("unexpected fallback order: [%q, %q]", responses[0].Content, responses[1].Content)
	}
	if got := responses[0].Metadata["source_text"]; got != "A" {
		t.Fatalf("response[0] source_text = %v, want A", got)
	}
	if got := responses[1].Metadata["source_text"]; got != "B" {
		t.Fatalf("response[1] source_text = %v, want B", got)
	}
	assertProgressEndsAt(t, progressLog, 2)
}

func TestSyncExecutorExecuteWithProgress_BatchDuplicateQueueJobIDFallsBackToRemainingOrder(t *testing.T) {
	batchClient := &stubBatchClient{
		statuses: []gatewayllm.BatchStatus{
			{State: gatewayllm.BatchStateCompleted, Progress: 1.0},
		},
		results: []gatewayllm.Response{
			{
				Success:  true,
				Content:  "first-for-0",
				Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-0"},
			},
			{
				Success:  true,
				Content:  "dup-for-0",
				Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-0"},
			},
		},
	}
	manager := &stubLLMManager{
		bulkStrategy: gatewayllm.BulkStrategyBatch,
		batchClient:  batchClient,
	}
	executor := NewSyncExecutor(manager)
	progressLog := make([]int, 0)
	responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
		Provider:     "xai",
		Model:        "grok-3",
		BulkStrategy: "batch",
	}, []llmio.Request{
		{Metadata: map[string]interface{}{"source_text": "A"}},
		{Metadata: map[string]interface{}{"source_text": "B"}},
	}, func(completed, total int) {
		_ = total
		progressLog = append(progressLog, completed)
	})
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("response count = %d, want 2", len(responses))
	}
	if responses[0].Content != "first-for-0" {
		t.Fatalf("response[0].Content = %q, want first-for-0", responses[0].Content)
	}
	if responses[1].Content != "dup-for-0" {
		t.Fatalf("response[1].Content = %q, want dup-for-0", responses[1].Content)
	}
	if got := responses[1].Metadata["source_text"]; got != "B" {
		t.Fatalf("response[1] source_text = %v, want B", got)
	}
	assertProgressEndsAt(t, progressLog, 2)
}

func TestSyncExecutorExecuteWithProgress_BatchMissingResultCreatesFailureStub(t *testing.T) {
	batchClient := &stubBatchClient{
		statuses: []gatewayllm.BatchStatus{
			{State: gatewayllm.BatchStateCompleted, Progress: 1.0},
		},
		results: []gatewayllm.Response{
			{
				Success:  true,
				Content:  "only-first",
				Metadata: map[string]interface{}{gatewayllm.BatchMetadataQueueJobIDKey: "terminology-0"},
			},
		},
	}
	manager := &stubLLMManager{
		bulkStrategy: gatewayllm.BulkStrategyBatch,
		batchClient:  batchClient,
	}
	executor := NewSyncExecutor(manager)
	progressLog := make([]int, 0)
	responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
		Provider:     "xai",
		Model:        "grok-3",
		BulkStrategy: "batch",
	}, []llmio.Request{
		{Metadata: map[string]interface{}{"source_text": "A"}},
		{Metadata: map[string]interface{}{"source_text": "B"}},
	}, func(completed, total int) {
		_ = total
		progressLog = append(progressLog, completed)
	})
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("response count = %d, want 2", len(responses))
	}
	if responses[0].Content != "only-first" || !responses[0].Success {
		t.Fatalf("unexpected response[0]: %+v", responses[0])
	}
	if responses[1].Success {
		t.Fatalf("response[1] should be failure stub: %+v", responses[1])
	}
	if responses[1].Error != "batch result missing for request" {
		t.Fatalf("response[1].Error = %q, want batch result missing for request", responses[1].Error)
	}
	if got := responses[1].Metadata["source_text"]; got != "B" {
		t.Fatalf("response[1] source_text = %v, want B", got)
	}
	assertProgressEndsAt(t, progressLog, 2)
}

func TestSyncExecutorExecuteWithProgress_FallsBackToSyncWhenResolvedStrategyIsSync(t *testing.T) {
	manager := &stubLLMManager{
		bulkStrategy: gatewayllm.BulkStrategySync,
		client: &stubLLMClient{
			completeFn: func(req gatewayllm.Request) gatewayllm.Response {
				return gatewayllm.Response{
					Success:  true,
					Content:  "ok",
					Metadata: req.Metadata,
				}
			},
		},
	}
	executor := NewSyncExecutor(manager)

	responses, err := executor.ExecuteWithProgress(context.Background(), llmio.ExecutionConfig{
		Provider:        "lmstudio",
		Model:           "local-model",
		BulkStrategy:    "batch",
		SyncConcurrency: 1,
	}, []llmio.Request{{Metadata: map[string]interface{}{"source_text": "A"}}}, nil)
	if err != nil {
		t.Fatalf("ExecuteWithProgress failed: %v", err)
	}
	if manager.getClientCalls != 1 {
		t.Fatalf("expected GetClient to be called once, got=%d", manager.getClientCalls)
	}
	if manager.getBatchClientCalls != 0 {
		t.Fatalf("expected GetBatchClient not to be called, got=%d", manager.getBatchClientCalls)
	}
	if len(responses) != 1 || !responses[0].Success {
		t.Fatalf("unexpected sync response: %+v", responses)
	}
}

func assertProgressEndsAt(t *testing.T, progressLog []int, want int) {
	t.Helper()
	if len(progressLog) == 0 {
		t.Fatalf("progress callback was not called")
	}
	got := progressLog[len(progressLog)-1]
	if got != want {
		t.Fatalf("final progress = %d, want %d (all=%v)", got, want, progressLog)
	}
}

type stubLLMManager struct {
	client              gatewayllm.LLMClient
	batchClient         gatewayllm.BatchClient
	bulkStrategy        gatewayllm.BulkStrategy
	getClientCalls      int
	getBatchClientCalls int
}

func (s *stubLLMManager) GetClient(ctx context.Context, config gatewayllm.LLMConfig) (gatewayllm.LLMClient, error) {
	_ = ctx
	_ = config
	s.getClientCalls++
	return s.client, nil
}

func (s *stubLLMManager) GetBatchClient(ctx context.Context, config gatewayllm.LLMConfig) (gatewayllm.BatchClient, error) {
	_ = ctx
	_ = config
	s.getBatchClientCalls++
	return s.batchClient, nil
}

func (s *stubLLMManager) ResolveBulkStrategy(ctx context.Context, strategy gatewayllm.BulkStrategy, provider string) gatewayllm.BulkStrategy {
	_ = ctx
	_ = strategy
	_ = provider
	return s.bulkStrategy
}

type stubBatchClient struct {
	submittedRequests []gatewayllm.Request
	submitCalls       int
	statuses          []gatewayllm.BatchStatus
	statusCallIndex   int
	results           []gatewayllm.Response
}

func (s *stubBatchClient) SubmitBatch(ctx context.Context, reqs []gatewayllm.Request) (gatewayllm.BatchJobID, error) {
	_ = ctx
	s.submitCalls++
	s.submittedRequests = append([]gatewayllm.Request(nil), reqs...)
	return gatewayllm.BatchJobID{ID: "job-1", Provider: "xai"}, nil
}

func (s *stubBatchClient) GetBatchStatus(ctx context.Context, id gatewayllm.BatchJobID) (gatewayllm.BatchStatus, error) {
	_ = ctx
	_ = id
	if len(s.statuses) == 0 {
		return gatewayllm.BatchStatus{State: gatewayllm.BatchStateCompleted, Progress: 1.0}, nil
	}
	idx := s.statusCallIndex
	if idx >= len(s.statuses) {
		idx = len(s.statuses) - 1
	}
	s.statusCallIndex++
	return s.statuses[idx], nil
}

func (s *stubBatchClient) GetBatchResults(ctx context.Context, id gatewayllm.BatchJobID) ([]gatewayllm.Response, error) {
	_ = ctx
	_ = id
	return append([]gatewayllm.Response(nil), s.results...), nil
}

type stubLLMClient struct {
	completeFn func(req gatewayllm.Request) gatewayllm.Response
}

func (s *stubLLMClient) ListModels(ctx context.Context) ([]gatewayllm.ModelInfo, error) {
	_ = ctx
	return nil, nil
}

func (s *stubLLMClient) Complete(ctx context.Context, req gatewayllm.Request) (gatewayllm.Response, error) {
	_ = ctx
	if s.completeFn == nil {
		return gatewayllm.Response{Success: true}, nil
	}
	return s.completeFn(req), nil
}

func (s *stubLLMClient) GenerateStructured(ctx context.Context, req gatewayllm.Request) (gatewayllm.Response, error) {
	return s.Complete(ctx, req)
}

func (s *stubLLMClient) StreamComplete(ctx context.Context, req gatewayllm.Request) (gatewayllm.StreamResponse, error) {
	_ = ctx
	_ = req
	return nil, nil
}

func (s *stubLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	_ = text
	return nil, nil
}

func (s *stubLLMClient) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}
