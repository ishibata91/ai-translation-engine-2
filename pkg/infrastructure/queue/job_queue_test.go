package queue

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// mockLLMClient simulates LLM completion
type mockLLMClient struct {
	delay     time.Duration
	failIdx   int
	count     int
	loadErr   error
	loadCnt   int
	unloadCnt int
}

func (m *mockLLMClient) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{{ID: "mock-model", DisplayName: "mock-model"}}, nil
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
func (m *mockLLMClient) GenerateStructured(ctx context.Context, req llm.Request) (llm.Response, error) {
	return m.Complete(ctx, req)
}
func (m *mockLLMClient) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	return nil, nil
}
func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *mockLLMClient) HealthCheck(ctx context.Context) error { return nil }
func (m *mockLLMClient) LoadModel(ctx context.Context, model string, contextLength int) (string, error) {
	m.loadCnt++
	if m.loadErr != nil {
		return "", m.loadErr
	}
	return "instance-1", nil
}
func (m *mockLLMClient) UnloadModel(ctx context.Context, instanceID string) error {
	m.unloadCnt++
	return nil
}

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
	lastConfig  llm.LLMConfig
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	m.lastConfig = config
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
	switch key {
	case llm.LLMDefaultProviderKey:
		return "lmstudio", nil
	case "lmstudio_model_id":
		return "mock-model", nil
	case "lmstudio_endpoint":
		return "http://localhost:1234", nil
	default:
		return "mock", nil
	}
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

func TestJobQueue_LoadFailNoRetryAndUnloadOnce(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, _ := NewQueue(ctx, ":memory:", logger)
	defer q.Close()

	processID := "load-fail"
	_ = q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}})

	client := &mockLLMClient{loadErr: context.DeadlineExceeded}
	worker := NewWorker(q, &mockLLMManager{client: client}, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	err := worker.ProcessProcessID(ctx, processID)
	if err == nil {
		t.Fatalf("expected error on load failure")
	}
	if client.loadCnt != 1 {
		t.Fatalf("expected load called once, got %d", client.loadCnt)
	}
	if client.unloadCnt != 0 {
		t.Fatalf("expected unload not called on failed load, got %d", client.unloadCnt)
	}
}

func TestJobQueue_ResumeUsesStoredProviderModel(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, _ := NewQueue(ctx, ":memory:", logger)
	defer q.Close()

	processID := "resume-ok"
	_ = q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}})
	_ = q.UpdateProcessMetadata(ctx, processID, "lmstudio", "stored-model")

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	worker := NewWorker(q, manager, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	if err := worker.ProcessProcessID(ctx, processID); err != nil {
		t.Fatalf("ProcessProcessID failed: %v", err)
	}
	if manager.lastConfig.Provider != "lmstudio" || manager.lastConfig.Model != "stored-model" {
		t.Fatalf("expected stored provider/model, got %s/%s", manager.lastConfig.Provider, manager.lastConfig.Model)
	}
}

func TestJobQueue_ResumeFailsWhenMetadataMissing(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, _ := NewQueue(ctx, ":memory:", logger)
	defer q.Close()

	processID := "resume-missing"
	_ = q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}})
	_, _ = q.db.ExecContext(ctx, "UPDATE llm_jobs SET provider = 'lmstudio', model = '' WHERE process_id = ?", processID)

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	worker := NewWorker(q, manager, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	err := worker.ProcessProcessID(ctx, processID)
	if err == nil {
		t.Fatalf("expected metadata missing error")
	}
}

func TestQueue_NullMetadataColumnsCanBeScanned(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, _ := NewQueue(ctx, ":memory:", logger)
	defer q.Close()

	processID := "legacy-null"
	if _, err := q.db.ExecContext(ctx, `
		INSERT INTO llm_jobs (
			id, process_id, request_json, status,
			provider, model, request_fingerprint, structured_output_schema_version,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, NULL, NULL, NULL, NULL, datetime('now'), datetime('now'))
	`, "job-1", processID, `{"user_prompt":"x"}`, StatusPending); err != nil {
		t.Fatalf("insert legacy row failed: %v", err)
	}

	jobs, err := q.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil {
		t.Fatalf("GetJobsByStatus failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 row, got %d", len(jobs))
	}
	if jobs[0].Provider != "" || jobs[0].Model != "" || jobs[0].RequestFingerprint != "" || jobs[0].StructuredOutputSchemaVersion != "" {
		t.Fatalf("expected nullable metadata to be normalized to empty strings, got %+v", jobs[0])
	}
}

func TestQueue_SubmitTaskRequests_StoresTaskStateAndCursor(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	taskID := "task-001"
	reqs := []llm.Request{
		{UserPrompt: "p1"},
		{UserPrompt: "p2"},
	}
	if err := q.SubmitTaskRequests(ctx, taskID, "persona_extraction", reqs); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	state, err := q.GetTaskRequestState(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequestState failed: %v", err)
	}
	if state.TaskType != "persona_extraction" || state.Total != 2 || state.Pending != 2 {
		t.Fatalf("unexpected state: %+v", state)
	}

	items, err := q.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(items))
	}
	if items[0].ResumeCursor != 0 || items[1].ResumeCursor != 0 {
		t.Fatalf("unexpected resume cursors: %d, %d", items[0].ResumeCursor, items[1].ResumeCursor)
	}
	if items[0].RequestState != RequestStatePending {
		t.Fatalf("unexpected request state: %s", items[0].RequestState)
	}
}

func TestQueue_TaskRequests_SurviveReopen(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dbPath := filepath.Join(t.TempDir(), "llm_jobs_test.db")

	q1, err := NewQueue(ctx, dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	taskID := "task-reopen"
	if err := q1.SubmitTaskRequests(ctx, taskID, "persona_extraction", []llm.Request{{UserPrompt: "persist"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}
	_ = q1.Close()

	q2, err := NewQueue(ctx, dbPath, logger)
	if err != nil {
		t.Fatalf("failed to reopen queue: %v", err)
	}
	defer q2.Close()

	items, err := q2.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 persisted request, got %d", len(items))
	}
	if items[0].RequestState != RequestStatePending {
		t.Fatalf("unexpected persisted request state: %s", items[0].RequestState)
	}
}
