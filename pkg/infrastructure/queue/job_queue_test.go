package queue

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// mockLLMClient simulates LLM completion
type mockLLMClient struct {
	delay   time.Duration
	failIdx int
	count   int
}

func (m *mockLLMClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	m.count++
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.failIdx > 0 && m.count == m.failIdx {
		return llm.Response{Success: false, Error: "mock failure"}, nil
	}
	return llm.Response{Success: true, Content: "mock response"}, nil
}
func (m *mockLLMClient) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	return nil, nil
}
func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *mockLLMClient) HealthCheck(ctx context.Context) error { return nil }

type mockBatchClient struct {
	status string
}

func (m *mockBatchClient) SubmitBatch(ctx context.Context, reqs []llm.Request) (llm.BatchJobID, error) {
	return llm.BatchJobID{ID: "batch-123", Provider: "mock"}, nil
}
func (m *mockBatchClient) GetBatchStatus(ctx context.Context, id llm.BatchJobID) (llm.BatchStatus, error) {
	return llm.BatchStatus{ID: id.ID, State: m.status, Progress: 0.5}, nil
}
func (m *mockBatchClient) GetBatchResults(ctx context.Context, id llm.BatchJobID) ([]llm.Response, error) {
	return []llm.Response{{Success: true, Content: "batch results"}}, nil
}

type mockLLMManager struct {
	client      *mockLLMClient
	batchClient *mockBatchClient
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	return m.client, nil
}
func (m *mockLLMManager) GetBatchClient(ctx context.Context, config llm.LLMConfig) (llm.BatchClient, error) {
	return m.batchClient, nil
}
func (m *mockLLMManager) ResolveBulkStrategy(ctx context.Context, strategy llm.BulkStrategy, provider string) llm.BulkStrategy {
	return strategy
}

type mockConfigStore struct{}

func (m *mockConfigStore) Get(ctx context.Context, ns, key string) (string, error) {
	if key == llm.LLMBulkStrategyKey {
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
		llm.Request{UserPrompt: "prompt 1"},
		llm.Request{UserPrompt: "prompt 2"},
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
	q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}})

	batchMock := &mockBatchClient{status: "PENDING"}
	manager := &mockLLMManager{batchClient: batchMock}

	// Fast-forward status in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		batchMock.status = "COMPLETED"
	}()

	// Mock config to force batch
	cfg := &mockConfigStore{}

	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(50 * time.Millisecond)

	err := worker.processBatch(ctx, processID, llm.LLMConfig{})
	if err != nil {
		t.Errorf("processBatch failed: %v", err)
	}

	results, _ := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if len(results) != 1 {
		t.Errorf("Expected 1 completed job from batch, got %d", len(results))
	}
}
