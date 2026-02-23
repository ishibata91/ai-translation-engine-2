package job_queue

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// mockLLMClient simulates LLM completion
type mockLLMClient struct {
	delay   time.Duration
	failIdx int
	count   int
}

func (m *mockLLMClient) Complete(ctx context.Context, req llm_client.Request) (llm_client.Response, error) {
	m.count++
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.failIdx > 0 && m.count == m.failIdx {
		return llm_client.Response{Success: false, Error: "mock failure"}, nil
	}
	return llm_client.Response{Success: true, Content: "mock response"}, nil
}
func (m *mockLLMClient) StreamComplete(ctx context.Context, req llm_client.Request) (llm_client.StreamResponse, error) {
	return nil, nil
}
func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *mockLLMClient) HealthCheck(ctx context.Context) error { return nil }

type mockBatchClient struct {
	status string
}

func (m *mockBatchClient) SubmitBatch(ctx context.Context, reqs []llm_client.Request) (llm_client.BatchJobID, error) {
	return llm_client.BatchJobID{ID: "batch-123", Provider: "mock"}, nil
}
func (m *mockBatchClient) GetBatchStatus(ctx context.Context, id llm_client.BatchJobID) (llm_client.BatchStatus, error) {
	return llm_client.BatchStatus{ID: id.ID, State: m.status, Progress: 0.5}, nil
}
func (m *mockBatchClient) GetBatchResults(ctx context.Context, id llm_client.BatchJobID) ([]llm_client.Response, error) {
	return []llm_client.Response{{Success: true, Content: "batch results"}}, nil
}

type mockLLMManager struct {
	client      *mockLLMClient
	batchClient *mockBatchClient
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm_client.LLMConfig) (llm_client.LLMClient, error) {
	return m.client, nil
}
func (m *mockLLMManager) GetBatchClient(ctx context.Context, config llm_client.LLMConfig) (llm_client.BatchClient, error) {
	return m.batchClient, nil
}
func (m *mockLLMManager) ResolveBulkStrategy(ctx context.Context, strategy llm_client.BulkStrategy, provider string) llm_client.BulkStrategy {
	return strategy
}

type mockConfigStore struct{}

func (m *mockConfigStore) Get(ctx context.Context, ns, key string) (string, error) {
	if key == llm_client.LLMBulkStrategyKey {
		return "sync", nil
	}
	return "mock", nil
}
func (m *mockConfigStore) Set(ctx context.Context, ns, key, val string) error { return nil }
func (m *mockConfigStore) Delete(ctx context.Context, ns, key string) error   { return nil }
func (m *mockConfigStore) GetAll(ctx context.Context, ns string) (map[string]string, error) {
	return nil, nil
}
func (m *mockConfigStore) Watch(ns, key string, cb config.ChangeCallback) config.UnsubscribeFunc {
	return func() {}
}

type mockSecretStore struct{}

func (m *mockSecretStore) GetSecret(ctx context.Context, ns, key string) (string, error) {
	return "key", nil
}
func (m *mockSecretStore) SetSecret(ctx context.Context, ns, key, val string) error { return nil }
func (m *mockSecretStore) DeleteSecret(ctx context.Context, ns, key string) error   { return nil }
func (m *mockSecretStore) ListSecretKeys(ctx context.Context, ns string) ([]string, error) {
	return nil, nil
}

func TestJobQueue_Lifecycle(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dbPath := ":memory:"

	q, err := NewQueue(ctx, dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "test-process-uuid"
	reqs := []any{
		llm_client.Request{UserPrompt: "prompt 1"},
		llm_client.Request{UserPrompt: "prompt 2"},
	}

	// 1. Submit
	err = q.SubmitJobs(ctx, processID, reqs)
	if err != nil {
		t.Errorf("SubmitJobs failed: %v", err)
	}

	// 2. Fetch Pending
	jobs, err := q.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil || len(jobs) != 2 {
		t.Errorf("Expected 2 pending jobs, got %d (err: %v)", len(jobs), err)
	}

	// 3. Worker Execution (Partial Mock)
	worker := NewWorker(q, &mockLLMManager{client: &mockLLMClient{}}, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	err = worker.ProcessProcessID(ctx, processID)
	if err != nil {
		t.Errorf("Worker execution failed: %v", err)
	}

	// 4. Get Completed
	completed, err := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if err != nil || len(completed) != 2 {
		t.Errorf("Expected 2 completed jobs, got %d (err: %v)", len(completed), err)
	}

	// 5. Delete
	err = q.DeleteJobs(ctx, processID)
	if err != nil {
		t.Errorf("DeleteJobs failed: %v", err)
	}

	// 6. Verify Empty
	remaining, _ := q.GetResults(ctx, processID)
	if len(remaining) != 0 {
		t.Errorf("Expected 0 jobs after deletion, got %d", len(remaining))
	}
}

func TestJobQueue_BatchPolling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dbPath := ":memory:"

	q, _ := NewQueue(ctx, dbPath, logger)
	defer q.Close()

	processID := "batch-test"
	q.SubmitJobs(ctx, processID, []any{llm_client.Request{UserPrompt: "q"}})

	batchMock := &mockBatchClient{status: "PENDING"}
	manager := &mockLLMManager{batchClient: batchMock}

	// Fast-forward status in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		batchMock.status = "COMPLETED"
	}()

	// Mock config to force batch
	cfg := &mockConfigStore{} // Note: implementation above returns "sync", but we can use a custom one if needed.
	// For simplicity in test, we just call processBatch directly or update mock.

	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(50 * time.Millisecond)

	// Inject strategy override in worker isn't easy without modifying worker,
	// but we can trust the flow if we test the sub-method.
	err := worker.processBatch(ctx, processID, llm_client.LLMConfig{})
	if err != nil {
		t.Errorf("processBatch failed: %v", err)
	}

	results, _ := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if len(results) != 1 {
		t.Errorf("Expected 1 completed job from batch, got %d", len(results))
	}
}
