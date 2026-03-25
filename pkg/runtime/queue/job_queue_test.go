package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	gatewayconfig "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
)

// mockLLMClient simulates LLM completion
type mockLLMClient struct {
	delay      time.Duration
	failIdx    int
	count      int
	loadErr    error
	loadCnt    int
	loadCtxLen int
	unloadCnt  int
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
	m.loadCtxLen = contextLength
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
	status        llm.BatchState
	results       []llm.Response
	submitCalls   int
	statusCalls   int
	resultsCalls  int
	submittedReqs []llm.Request
}

func (m *mockBatchClient) SubmitBatch(ctx context.Context, reqs []llm.Request) (llm.BatchJobID, error) {
	m.submitCalls++
	m.submittedReqs = append([]llm.Request(nil), reqs...)
	return llm.BatchJobID{ID: "batch-123", Provider: "mock"}, nil
}
func (m *mockBatchClient) GetBatchStatus(ctx context.Context, id llm.BatchJobID) (llm.BatchStatus, error) {
	m.statusCalls++
	return llm.BatchStatus{ID: id.ID, State: m.status, Progress: 0.5}, nil
}
func (m *mockBatchClient) GetBatchResults(ctx context.Context, id llm.BatchJobID) ([]llm.Response, error) {
	m.resultsCalls++
	if len(m.results) > 0 {
		return m.results, nil
	}
	return []llm.Response{{Success: true, Content: "batch results"}}, nil
}

type mockLLMManager struct {
	client      *mockLLMClient
	batchClient *mockBatchClient
	lastConfig  llm.LLMConfig
	clientCalls int
	batchCalls  int
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	m.clientCalls++
	m.lastConfig = config
	return m.client, nil
}
func (m *mockLLMManager) GetBatchClient(ctx context.Context, config llm.LLMConfig) (llm.BatchClient, error) {
	m.batchCalls++
	m.lastConfig = config
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
func (m *mockConfigStore) Watch(ns, key string, cb gatewayconfig.ChangeCallback) gatewayconfig.UnsubscribeFunc {
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

type mapConfigStore struct {
	values map[string]string
}

func (m *mapConfigStore) Get(ctx context.Context, ns, key string) (string, error) {
	if v, ok := m.values[ns+"::"+key]; ok {
		return v, nil
	}
	return "", nil
}
func (m *mapConfigStore) Set(ctx context.Context, ns, key, val string) error { return nil }
func (m *mapConfigStore) Delete(ctx context.Context, ns, key string) error   { return nil }
func (m *mapConfigStore) GetAll(ctx context.Context, ns string) (map[string]string, error) {
	return nil, nil
}
func (m *mapConfigStore) Watch(ns, key string, cb gatewayconfig.ChangeCallback) gatewayconfig.UnsubscribeFunc {
	return func() {}
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

	q, err := NewQueue(ctx, dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "batch-test"
	if err := q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitJobs failed: %v", err)
	}

	batchMock := &mockBatchClient{status: llm.BatchStateRunning}
	manager := &mockLLMManager{batchClient: batchMock}

	// Fast-forward status in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		batchMock.status = llm.BatchStateCompleted
	}()

	// Mock config to force batch
	cfg := &mockConfigStore{}

	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(50 * time.Millisecond)

	err = worker.processBatch(ctx, processID, llm.LLMConfig{}, ProcessOptions{})
	if err != nil {
		t.Errorf("processBatch failed: %v", err)
	}

	results, _ := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if len(results) != 1 {
		t.Errorf("Expected 1 completed job from batch, got %d", len(results))
	}
}

func TestQueue_PrepareTaskResume_PreservesBatchJobID(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	taskID := "resume-keep-batch-id"
	if err := q.SubmitTaskRequests(ctx, taskID, "persona_extraction", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	jobs, err := q.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	encodedBatchID, err := json.Marshal(llm.BatchJobID{ID: "batch-existing", Provider: "xai"})
	if err != nil {
		t.Fatalf("marshal batch id failed: %v", err)
	}
	batchIDString := string(encodedBatchID)

	if err := q.UpdateJob(ctx, jobs[0].ID, StatusInProgress, nil, nil, &batchIDString); err != nil {
		t.Fatalf("UpdateJob failed: %v", err)
	}

	if err := q.PrepareTaskResume(ctx, taskID); err != nil {
		t.Fatalf("PrepareTaskResume failed: %v", err)
	}

	resumedJobs, err := q.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests after resume failed: %v", err)
	}
	if len(resumedJobs) != 1 {
		t.Fatalf("expected 1 resumed job, got %d", len(resumedJobs))
	}
	if resumedJobs[0].Status != StatusPending || resumedJobs[0].RequestState != RequestStatePending {
		t.Fatalf("unexpected resumed state: status=%s request_state=%s", resumedJobs[0].Status, resumedJobs[0].RequestState)
	}
	if resumedJobs[0].BatchJobID == nil || *resumedJobs[0].BatchJobID != batchIDString {
		t.Fatalf("expected batch_job_id to be preserved, got %v", resumedJobs[0].BatchJobID)
	}
}

func TestWorker_ProcessBatch_ReconnectsExistingBatchJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "batch-reconnect"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	jobs, err := q.GetTaskRequests(ctx, processID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	encodedBatchID, err := json.Marshal(llm.BatchJobID{ID: "batch-existing", Provider: "xai"})
	if err != nil {
		t.Fatalf("marshal batch id failed: %v", err)
	}
	batchIDString := string(encodedBatchID)
	if err := q.UpdateJob(ctx, jobs[0].ID, StatusPending, nil, nil, &batchIDString); err != nil {
		t.Fatalf("UpdateJob failed: %v", err)
	}

	batchMock := &mockBatchClient{
		status:  llm.BatchStateCompleted,
		results: []llm.Response{{Success: true, Content: "batch results"}},
	}
	manager := &mockLLMManager{batchClient: batchMock}
	worker := NewWorker(q, manager, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	if err := worker.processBatch(ctx, processID, llm.LLMConfig{Provider: "xai", Model: "xai-model"}, ProcessOptions{}); err != nil {
		t.Fatalf("processBatch failed: %v", err)
	}

	if batchMock.submitCalls != 0 {
		t.Fatalf("expected reconnect path to skip SubmitBatch, got %d calls", batchMock.submitCalls)
	}
	if batchMock.statusCalls == 0 || batchMock.resultsCalls == 0 {
		t.Fatalf("expected reconnect path to poll and fetch results, status_calls=%d results_calls=%d", batchMock.statusCalls, batchMock.resultsCalls)
	}

	completedJobs, err := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if err != nil {
		t.Fatalf("GetJobsByStatus completed failed: %v", err)
	}
	if len(completedJobs) != 1 {
		t.Fatalf("expected 1 completed job after reconnect, got %d", len(completedJobs))
	}
}

func TestWorker_ProcessBatch_PartialFailedPersistsSuccessfulResponses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "batch-partial-failed"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q1"}, {UserPrompt: "q2"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	batchMock := &mockBatchClient{
		status: llm.BatchStatePartialFailed,
		results: []llm.Response{
			{Success: true, Content: "ok"},
			{Success: false, Error: "provider error"},
		},
	}
	manager := &mockLLMManager{batchClient: batchMock}
	worker := NewWorker(q, manager, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	if err := worker.processBatch(ctx, processID, llm.LLMConfig{Provider: "xai", Model: "xai-model"}, ProcessOptions{}); err != nil {
		t.Fatalf("processBatch failed: %v", err)
	}

	completedJobs, err := q.GetJobsByStatus(ctx, processID, StatusCompleted)
	if err != nil {
		t.Fatalf("GetJobsByStatus completed failed: %v", err)
	}
	if len(completedJobs) != 1 {
		t.Fatalf("expected 1 completed job in partial_failed, got %d", len(completedJobs))
	}

	failedJobs, err := q.GetJobsByStatus(ctx, processID, StatusFailed)
	if err != nil {
		t.Fatalf("GetJobsByStatus failed failed: %v", err)
	}
	if len(failedJobs) != 1 {
		t.Fatalf("expected 1 failed job in partial_failed, got %d", len(failedJobs))
	}
}

func TestWorker_ProcessBatch_CorrelatesByQueueJobIDOnShuffledResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "batch-correlation-shuffled"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q0"}, {UserPrompt: "q1"}, {UserPrompt: "q2"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	pendingJobs, err := q.GetJobsByStatus(ctx, processID, StatusPending)
	if err != nil {
		t.Fatalf("GetJobsByStatus pending failed: %v", err)
	}
	if len(pendingJobs) != 3 {
		t.Fatalf("expected 3 pending jobs, got %d", len(pendingJobs))
	}

	if _, err := q.db.ExecContext(ctx, `
		UPDATE llm_jobs
		SET resume_cursor = CASE id
			WHEN ? THEN 30
			WHEN ? THEN 10
			WHEN ? THEN 20
			ELSE resume_cursor
		END
		WHERE process_id = ?
	`, pendingJobs[0].ID, pendingJobs[1].ID, pendingJobs[2].ID, processID); err != nil {
		t.Fatalf("failed to update resume cursor: %v", err)
	}

	batchMock := &mockBatchClient{
		status: llm.BatchStateCompleted,
		results: []llm.Response{
			{Success: true, Content: "content-2", Metadata: map[string]interface{}{llm.BatchMetadataQueueJobIDKey: pendingJobs[2].ID}},
			{Success: false, Error: "failed-1", Metadata: map[string]interface{}{llm.BatchMetadataQueueJobIDKey: pendingJobs[1].ID}},
			{Success: true, Content: "content-0", Metadata: map[string]interface{}{llm.BatchMetadataQueueJobIDKey: pendingJobs[0].ID}},
		},
	}
	worker := NewWorker(q, &mockLLMManager{batchClient: batchMock}, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	if err := worker.processBatch(ctx, processID, llm.LLMConfig{Provider: "xai", Model: "xai-model"}, ProcessOptions{}); err != nil {
		t.Fatalf("processBatch failed: %v", err)
	}

	if len(batchMock.submittedReqs) != 3 {
		t.Fatalf("expected 3 submitted requests, got %d", len(batchMock.submittedReqs))
	}
	seenIDs := map[string]bool{}
	for idx, req := range batchMock.submittedReqs {
		queueJobID, ok := req.Metadata[llm.BatchMetadataQueueJobIDKey].(string)
		if !ok || queueJobID == "" {
			t.Fatalf("submitted request %d missing queue_job_id metadata: %+v", idx, req.Metadata)
		}
		if _, ok := req.Metadata[llm.BatchMetadataQueueRequestSeqKey]; !ok {
			t.Fatalf("submitted request %d missing queue_request_seq metadata: %+v", idx, req.Metadata)
		}
		seenIDs[queueJobID] = true
	}
	for _, job := range pendingJobs {
		if !seenIDs[job.ID] {
			t.Fatalf("job %s was not submitted with queue_job_id metadata", job.ID)
		}
	}

	allJobs, err := q.GetResults(ctx, processID)
	if err != nil {
		t.Fatalf("GetResults failed: %v", err)
	}
	if len(allJobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(allJobs))
	}
	jobsByID := map[string]JobRequest{}
	for _, job := range allJobs {
		jobsByID[job.ID] = job
	}

	job0 := jobsByID[pendingJobs[0].ID]
	if job0.Status != StatusCompleted || job0.ResponseJSON == nil {
		t.Fatalf("job0 unexpected state: %+v", job0)
	}
	var job0Resp llm.Response
	if err := json.Unmarshal([]byte(*job0.ResponseJSON), &job0Resp); err != nil {
		t.Fatalf("unmarshal job0 response failed: %v", err)
	}
	if job0Resp.Content != "content-0" {
		t.Fatalf("job0 content = %q, want content-0", job0Resp.Content)
	}

	job1 := jobsByID[pendingJobs[1].ID]
	if job1.Status != StatusFailed || job1.ErrorMessage == nil || *job1.ErrorMessage != "failed-1" {
		t.Fatalf("job1 unexpected state: %+v", job1)
	}

	job2 := jobsByID[pendingJobs[2].ID]
	if job2.Status != StatusCompleted || job2.ResponseJSON == nil {
		t.Fatalf("job2 unexpected state: %+v", job2)
	}
	var job2Resp llm.Response
	if err := json.Unmarshal([]byte(*job2.ResponseJSON), &job2Resp); err != nil {
		t.Fatalf("unmarshal job2 response failed: %v", err)
	}
	if job2Resp.Content != "content-2" {
		t.Fatalf("job2 content = %q, want content-2", job2Resp.Content)
	}
}

func TestJobQueue_LoadFailNoRetryAndUnloadOnce(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "load-fail"
	if err := q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitJobs failed: %v", err)
	}

	client := &mockLLMClient{loadErr: context.DeadlineExceeded}
	worker := NewWorker(q, &mockLLMManager{client: client}, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	err = worker.ProcessProcessID(ctx, processID)
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
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "resume-ok"
	if err := q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitJobs failed: %v", err)
	}
	if err := q.UpdateProcessMetadata(ctx, processID, "lmstudio", "stored-model"); err != nil {
		t.Fatalf("UpdateProcessMetadata failed: %v", err)
	}

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
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "resume-missing"
	if err := q.SubmitJobs(ctx, processID, []any{llm.Request{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitJobs failed: %v", err)
	}
	if _, err := q.db.ExecContext(ctx, "UPDATE llm_jobs SET provider = 'lmstudio', model = '' WHERE process_id = ?", processID); err != nil {
		t.Fatalf("failed to mutate metadata: %v", err)
	}

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	worker := NewWorker(q, manager, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	err = worker.ProcessProcessID(ctx, processID)
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

func TestQueue_DeleteTaskRequests_RemovesOnlyTargetTask(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	if err := q.SubmitTaskRequests(ctx, "task-a", "persona_extraction", []llm.Request{{UserPrompt: "a1"}, {UserPrompt: "a2"}}); err != nil {
		t.Fatalf("SubmitTaskRequests task-a failed: %v", err)
	}
	if err := q.SubmitTaskRequests(ctx, "task-b", "persona_extraction", []llm.Request{{UserPrompt: "b1"}}); err != nil {
		t.Fatalf("SubmitTaskRequests task-b failed: %v", err)
	}

	if err := q.DeleteTaskRequests(ctx, "task-a"); err != nil {
		t.Fatalf("DeleteTaskRequests failed: %v", err)
	}

	taskAJobs, err := q.GetTaskRequests(ctx, "task-a")
	if err != nil {
		t.Fatalf("GetTaskRequests task-a failed: %v", err)
	}
	if len(taskAJobs) != 0 {
		t.Fatalf("expected task-a jobs to be deleted, got %d", len(taskAJobs))
	}

	taskBJobs, err := q.GetTaskRequests(ctx, "task-b")
	if err != nil {
		t.Fatalf("GetTaskRequests task-b failed: %v", err)
	}
	if len(taskBJobs) != 1 {
		t.Fatalf("expected task-b jobs to remain, got %d", len(taskBJobs))
	}
}

func TestWorker_ProcessWithOptions_LoadsLMStudioContextLengthFromConfig(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "ctx-len"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	cfg := &mapConfigStore{
		values: map[string]string{
			"master_persona.llm::selected_provider":       "lmstudio",
			"master_persona.llm.lmstudio::model":          "mock-model",
			"master_persona.llm.lmstudio::endpoint":       "http://localhost:1234",
			"master_persona.llm.lmstudio::context_length": "8192",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		ConfigNamespace:        "master_persona.llm",
		UseConfigProviderModel: true,
		ConfigRead: ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if client.loadCtxLen != 8192 {
		t.Fatalf("expected context_length=8192, got %d", client.loadCtxLen)
	}
}

func TestWorker_ProcessWithOptions_LoadsSyncConcurrencyFromNamespace(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "sync-concurrency"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	cfg := &mapConfigStore{
		values: map[string]string{
			"master_persona.llm::selected_provider":         "lmstudio",
			"master_persona.llm.lmstudio::model":            "mock-model",
			"master_persona.llm.lmstudio::endpoint":         "http://localhost:1234",
			"master_persona.llm::sync_concurrency.lmstudio": "3",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		ConfigNamespace:        "master_persona.llm",
		UseConfigProviderModel: true,
		ConfigRead: ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if manager.lastConfig.Concurrency != 3 {
		t.Fatalf("expected sync concurrency=3, got %d", manager.lastConfig.Concurrency)
	}
}

func TestWorker_CancelKeepsCompletedRequests(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(context.Background(), ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "cancel-keep-completed"
	reqs := []any{
		llm.Request{UserPrompt: "p1"},
		llm.Request{UserPrompt: "p2"},
		llm.Request{UserPrompt: "p3"},
	}
	if err := q.SubmitJobs(context.Background(), processID, reqs); err != nil {
		t.Fatalf("SubmitJobs failed: %v", err)
	}

	client := &mockLLMClient{delay: 80 * time.Millisecond}
	worker := NewWorker(q, &mockLLMManager{client: client}, &mockConfigStore{}, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	err = worker.ProcessProcessID(ctx, processID)
	if err == nil {
		t.Fatalf("expected cancellation error")
	}

	completed, err := q.GetJobsByStatus(context.Background(), processID, StatusCompleted)
	if err != nil {
		t.Fatalf("GetJobsByStatus completed failed: %v", err)
	}
	if len(completed) == 0 {
		t.Fatalf("expected at least one completed job to be preserved on cancel")
	}
}

func TestWorker_ProcessWithOptions_UsesProviderScopedBulkStrategy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "custom-namespace-batch"
	if err := q.SubmitTaskRequests(ctx, processID, "generic_task", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	manager := &mockLLMManager{
		client:      &mockLLMClient{},
		batchClient: &mockBatchClient{status: llm.BatchStateCompleted},
	}
	cfg := &mapConfigStore{
		values: map[string]string{
			"custom.llm::selected_provider":    "xai",
			"custom.llm.xai::model":            "xai-model",
			"custom.llm::bulk_strategy":        "sync",
			"custom.llm.xai::bulk_strategy":    "batch",
			"custom.llm.gemini::bulk_strategy": "sync",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		UseConfigProviderModel: true,
		ConfigRead: ConfigReadOptions{
			Namespace:           "custom.llm",
			DefaultProvider:     "xai",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if manager.batchCalls == 0 {
		t.Fatalf("expected batch client path to be selected by provider-scoped strategy")
	}
	if manager.clientCalls != 0 {
		t.Fatalf("expected sync client path not to run when provider-scoped strategy=batch")
	}
	if manager.lastConfig.Provider != "xai" || manager.lastConfig.Model != "xai-model" {
		t.Fatalf("unexpected config resolved from custom namespace: %+v", manager.lastConfig)
	}
}

func TestWorker_ProcessWithOptions_AppliesExecutionOverrides(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "persona-override-profile"
	if err := q.SubmitTaskRequests(ctx, processID, "persona_extraction", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	client := &mockLLMClient{}
	manager := &mockLLMManager{client: client}
	cfg := &mapConfigStore{
		values: map[string]string{
			"custom.llm::selected_provider": "gemini",
			"custom.llm.gemini::model":      "stale-model",
			"custom.llm.gemini::endpoint":   "https://stale.example",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		UseConfigProviderModel: true,
		ProviderOverride:       "lmstudio",
		ModelOverride:          "lmstudio-community-model",
		EndpointOverride:       "http://127.0.0.1:1234",
		ConfigRead: ConfigReadOptions{
			Namespace:           "custom.llm",
			DefaultProvider:     "gemini",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if manager.lastConfig.Provider != "lmstudio" {
		t.Fatalf("expected provider override to win, got %q", manager.lastConfig.Provider)
	}
	if manager.lastConfig.Model != "lmstudio-community-model" {
		t.Fatalf("expected model override to win, got %q", manager.lastConfig.Model)
	}
	if manager.lastConfig.Endpoint != "http://127.0.0.1:1234" {
		t.Fatalf("expected endpoint override to win, got %q", manager.lastConfig.Endpoint)
	}
}

func TestWorker_ProcessWithOptions_FallsBackToLegacyBulkStrategyKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "legacy-bulk-strategy"
	if err := q.SubmitTaskRequests(ctx, processID, "generic_task", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	manager := &mockLLMManager{
		client:      &mockLLMClient{},
		batchClient: &mockBatchClient{status: llm.BatchStateCompleted},
	}
	cfg := &mapConfigStore{
		values: map[string]string{
			"custom.llm::selected_provider": "xai",
			"custom.llm.xai::model":         "xai-model",
			"custom.llm::bulk_strategy":     "batch",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		UseConfigProviderModel: true,
		ConfigRead: ConfigReadOptions{
			Namespace:           "custom.llm",
			DefaultProvider:     "xai",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if manager.batchCalls == 0 {
		t.Fatalf("expected legacy root bulk_strategy to keep batch behavior")
	}
	if manager.clientCalls != 0 {
		t.Fatalf("expected sync client path not to run when legacy strategy=batch")
	}
}

func TestWorker_ProcessWithOptions_DefaultsBulkStrategyToSync(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	q, err := NewQueue(ctx, ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	defer q.Close()

	processID := "default-sync-strategy"
	if err := q.SubmitTaskRequests(ctx, processID, "generic_task", []llm.Request{{UserPrompt: "q"}}); err != nil {
		t.Fatalf("SubmitTaskRequests failed: %v", err)
	}

	manager := &mockLLMManager{
		client:      &mockLLMClient{},
		batchClient: &mockBatchClient{status: llm.BatchStateCompleted},
	}
	cfg := &mapConfigStore{
		values: map[string]string{
			"custom.llm::selected_provider": "xai",
			"custom.llm.xai::model":         "xai-model",
		},
	}
	worker := NewWorker(q, manager, cfg, &mockSecretStore{}, progress.NewNoopNotifier(), logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	err = worker.ProcessProcessIDWithOptions(ctx, processID, ProcessOptions{
		UseConfigProviderModel: true,
		ConfigRead: ConfigReadOptions{
			Namespace:           "custom.llm",
			DefaultProvider:     "xai",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ProcessProcessIDWithOptions failed: %v", err)
	}
	if manager.clientCalls == 0 {
		t.Fatalf("expected sync client path to run when bulk_strategy is not saved")
	}
	if manager.batchCalls != 0 {
		t.Fatalf("expected batch path not to run when bulk_strategy defaults to sync")
	}
}

func TestWorker_ValidateExecutionProfile_RejectsUnsupportedBatchProvider(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	worker := NewWorker(
		&Queue{},
		&mockLLMManager{client: &mockLLMClient{}, batchClient: &mockBatchClient{status: llm.BatchStateCompleted}},
		&mapConfigStore{values: map[string]string{
			"master_persona.llm::selected_provider":         "lmstudio",
			"master_persona.llm.lmstudio::model":            "local-model",
			"master_persona.llm.lmstudio::bulk_strategy":    "batch",
			"master_persona.llm::sync_concurrency.lmstudio": "1",
		}},
		&mockSecretStore{},
		progress.NewNoopNotifier(),
		logger,
	)

	_, err := worker.ValidateExecutionProfile(ctx, ProcessOptions{
		ConfigRead: ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "lmstudio",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err == nil {
		t.Fatalf("expected unsupported batch provider to be rejected")
	}
}

func TestWorker_ValidateExecutionProfile_AllowsSupportedBatchProvider(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	worker := NewWorker(
		&Queue{},
		&mockLLMManager{client: &mockLLMClient{}, batchClient: &mockBatchClient{status: llm.BatchStateCompleted}},
		&mapConfigStore{values: map[string]string{
			"master_persona.llm::selected_provider":    "xai",
			"master_persona.llm.xai::model":            "xai-model",
			"master_persona.llm.xai::bulk_strategy":    "batch",
			"master_persona.llm::sync_concurrency.xai": "3",
		}},
		&mockSecretStore{},
		progress.NewNoopNotifier(),
		logger,
	)

	profile, err := worker.ValidateExecutionProfile(ctx, ProcessOptions{
		ConfigRead: ConfigReadOptions{
			Namespace:           "master_persona.llm",
			DefaultProvider:     "xai",
			SelectedProviderKey: "selected_provider",
		},
	})
	if err != nil {
		t.Fatalf("ValidateExecutionProfile failed: %v", err)
	}
	if profile.Provider != "xai" || profile.BulkStrategy != llm.BulkStrategyBatch {
		t.Fatalf("unexpected execution profile: %+v", profile)
	}
}
