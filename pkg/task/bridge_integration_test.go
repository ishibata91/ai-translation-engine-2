package task

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
	_ "modernc.org/sqlite"
)

type mockParser struct {
	out *parser.ParserOutput
	err error
}

func (m *mockParser) LoadExtractedJSON(ctx context.Context, path string) (*parser.ParserOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.out, nil
}

type mockPersonaGenerator struct {
	reqs          []llm.Request
	err           error
	mu            sync.Mutex
	saveCalls     int
	savedSpeakers []string
}

func (m *mockPersonaGenerator) ID() string { return "PersonaMock" }
func (m *mockPersonaGenerator) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reqs, nil
}
func (m *mockPersonaGenerator) SaveResults(ctx context.Context, responses []llm.Response) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalls += len(responses)
	for _, resp := range responses {
		if resp.Metadata == nil {
			continue
		}
		if speakerID, ok := resp.Metadata["speaker_id"].(string); ok && speakerID != "" {
			m.savedSpeakers = append(m.savedSpeakers, speakerID)
		}
	}
	return nil
}

func (m *mockPersonaGenerator) SaveResultsWithSummary(ctx context.Context, responses []llm.Response) (persona.SaveResultsSummary, error) {
	if err := m.SaveResults(ctx, responses); err != nil {
		return persona.SaveResultsSummary{}, err
	}
	return persona.SaveResultsSummary{
		SuccessCount: len(responses),
		FailCount:    0,
	}, nil
}

func (m *mockPersonaGenerator) SaveCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveCalls
}

type taskTestLLMClient struct{}

func (m *taskTestLLMClient) ListModels(ctx context.Context) ([]llm.ModelInfo, error) {
	return []llm.ModelInfo{{ID: "mock-model", DisplayName: "mock-model"}}, nil
}
func (m *taskTestLLMClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{
		Content:  "TL: |A brave and thoughtful hunter who speaks plainly.|",
		Success:  true,
		Metadata: req.Metadata,
	}, nil
}
func (m *taskTestLLMClient) GenerateStructured(ctx context.Context, req llm.Request) (llm.Response, error) {
	return m.Complete(ctx, req)
}
func (m *taskTestLLMClient) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	return nil, nil
}
func (m *taskTestLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (m *taskTestLLMClient) HealthCheck(ctx context.Context) error { return nil }

type taskTestLLMManager struct {
	client llm.LLMClient
}

func (m *taskTestLLMManager) GetClient(ctx context.Context, cfg llm.LLMConfig) (llm.LLMClient, error) {
	return m.client, nil
}
func (m *taskTestLLMManager) GetBatchClient(ctx context.Context, cfg llm.LLMConfig) (llm.BatchClient, error) {
	return nil, errors.New("batch not supported in test")
}
func (m *taskTestLLMManager) ResolveBulkStrategy(ctx context.Context, strategy llm.BulkStrategy, provider string) llm.BulkStrategy {
	if strategy == "" {
		return llm.BulkStrategySync
	}
	return strategy
}

type taskTestConfigStore struct{}

func (m *taskTestConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	if key == llm.LLMBulkStrategyKey {
		return string(llm.BulkStrategySync), nil
	}
	switch key {
	case "selected_provider":
		return "lmstudio", nil
	case "model":
		return "mock-model", nil
	case "endpoint":
		return "http://localhost:1234", nil
	default:
		return "", nil
	}
}
func (m *taskTestConfigStore) Set(ctx context.Context, namespace string, key string, value string) error {
	return nil
}
func (m *taskTestConfigStore) Delete(ctx context.Context, namespace string, key string) error { return nil }
func (m *taskTestConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	return nil, nil
}
func (m *taskTestConfigStore) Watch(namespace string, key string, callback config.ChangeCallback) config.UnsubscribeFunc {
	return func() {}
}

type taskTestSecretStore struct{}

func (m *taskTestSecretStore) GetSecret(ctx context.Context, namespace string, key string) (string, error) {
	return "", nil
}
func (m *taskTestSecretStore) SetSecret(ctx context.Context, namespace string, key string, value string) error {
	return nil
}
func (m *taskTestSecretStore) DeleteSecret(ctx context.Context, namespace string, key string) error { return nil }
func (m *taskTestSecretStore) ListSecretKeys(ctx context.Context, namespace string) ([]string, error) {
	return nil, nil
}

type capturedLog struct {
	level slog.Level
	msg   string
	attrs map[string]string
}

type captureHandler struct {
	mu   sync.Mutex
	logs []capturedLog
}

func (h *captureHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (h *captureHandler) Handle(ctx context.Context, rec slog.Record) error {
	row := capturedLog{
		level: rec.Level,
		msg:   rec.Message,
		attrs: map[string]string{},
	}
	rec.Attrs(func(a slog.Attr) bool {
		row.attrs[a.Key] = a.Value.String()
		return true
	})
	h.mu.Lock()
	h.logs = append(h.logs, row)
	h.mu.Unlock()
	return nil
}
func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(name string) slog.Handler       { return h }

func (h *captureHandler) findByMessage(msg string) []capturedLog {
	h.mu.Lock()
	defer h.mu.Unlock()
	var found []capturedLog
	for _, l := range h.logs {
		if l.msg == msg {
			found = append(found, l)
		}
	}
	return found
}

func setupTaskTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:task_bridge_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("failed to migrate task test db: %v", err)
	}
	return db
}

func setupRequestQueue(t *testing.T, logger *slog.Logger) *queue.Queue {
	t.Helper()
	q, err := queue.NewQueue(context.Background(), ":memory:", logger)
	if err != nil {
		t.Fatalf("failed to setup queue: %v", err)
	}
	return q
}

func waitTaskStatus(t *testing.T, bridge *Bridge, id string, status TaskStatus, timeout time.Duration) Task {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		tasks, err := bridge.GetAllTasks()
		if err != nil {
			t.Fatalf("GetAllTasks failed: %v", err)
		}
		for _, task := range tasks {
			if task.ID == id && task.Status == status {
				return task
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("task %s did not reach status %s within timeout", id, status)
	return Task{}
}

func TestBridge_StartMasterPersonTask_SuccessStatusAndInfoLog(t *testing.T) {
	db := setupTaskTestDB(t)
	defer db.Close()

	loggerSink := &captureHandler{}
	logger := slog.New(loggerSink)
	requestQueue := setupRequestQueue(t, logger)
	defer requestQueue.Close()

	manager := NewManager(nil, logger, NewStore(db))
	bridge := NewMasterPersonaBridge(
		manager,
		logger,
		&mockParser{
			out: &parser.ParserOutput{
				NPCs: map[string]parser.NPC{
					"npc-1": {BaseExtractedRecord: parser.BaseExtractedRecord{ID: "npc-1", Type: "NPC_"}, Name: "Aela"},
				},
			},
		},
		&mockPersonaGenerator{
			reqs: []llm.Request{
				{
					Metadata: map[string]interface{}{
						"speaker_id": "npc-1",
						"npc_name":   "Aela",
					},
				},
			},
		},
		progress.NewNoopNotifier(),
		requestQueue,
		nil,
	)

	taskID, err := bridge.StartMasterPersonTask(StartMasterPersonTaskInput{SourceJSONPath: "dummy.json"})
	if err != nil {
		t.Fatalf("StartMasterPersonTask failed: %v", err)
	}

	task := waitTaskStatus(t, bridge, taskID, StatusRequestGenerated, 3*time.Second)
	if task.Phase != "REQUEST_GENERATED" {
		t.Fatalf("unexpected task phase: got=%s", task.Phase)
	}

	logs := loggerSink.findByMessage("persona.requests.generated")
	if len(logs) == 0 {
		t.Fatalf("expected persona.requests.generated log")
	}
	var matched bool
	for _, row := range logs {
		if row.attrs["task_id"] == taskID && row.attrs["request_count"] == "1" && row.attrs["npc_count"] == "1" {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("generated log with expected task_id/request_count/npc_count not found")
	}
}

func TestBridge_StartMasterPersonTask_FailureStatusAndErrorLog(t *testing.T) {
	db := setupTaskTestDB(t)
	defer db.Close()

	loggerSink := &captureHandler{}
	logger := slog.New(loggerSink)
	requestQueue := setupRequestQueue(t, logger)
	defer requestQueue.Close()

	manager := NewManager(nil, logger, NewStore(db))
	bridge := NewMasterPersonaBridge(
		manager,
		logger,
		&mockParser{
			err: errors.New("mock parse failed"),
		},
		&mockPersonaGenerator{},
		progress.NewNoopNotifier(),
		requestQueue,
		nil,
	)

	taskID, err := bridge.StartMasterPersonTask(StartMasterPersonTaskInput{SourceJSONPath: "dummy.json"})
	if err != nil {
		t.Fatalf("StartMasterPersonTask failed: %v", err)
	}

	task := waitTaskStatus(t, bridge, taskID, StatusFailed, 3*time.Second)
	if task.ErrorMsg == "" {
		t.Fatalf("expected error message in failed task")
	}

	logs := loggerSink.findByMessage("persona.requests.failed")
	if len(logs) == 0 {
		t.Fatalf("expected persona.requests.failed log")
	}

	var matched bool
	for _, row := range logs {
		if row.attrs["task_id"] == taskID && row.attrs["reason"] != "" {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("failed log with expected task_id/reason not found")
	}
}

func TestBridge_ResumeMasterPersonaTask_SkipsAlreadySavedRequests(t *testing.T) {
	db := setupTaskTestDB(t)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	requestQueue := setupRequestQueue(t, logger)
	defer requestQueue.Close()

	manager := NewManager(nil, logger, NewStore(db))
	personaGen := &mockPersonaGenerator{
		reqs: []llm.Request{
			{Metadata: map[string]interface{}{"speaker_id": "npc-1", "npc_name": "Aela"}},
			{Metadata: map[string]interface{}{"speaker_id": "npc-2", "npc_name": "Farkas"}},
		},
	}
	worker := queue.NewWorker(
		requestQueue,
		&taskTestLLMManager{client: &taskTestLLMClient{}},
		&taskTestConfigStore{},
		&taskTestSecretStore{},
		progress.NewNoopNotifier(),
		logger,
	)
	bridge := NewMasterPersonaBridge(
		manager,
		logger,
		&mockParser{
			out: &parser.ParserOutput{
				NPCs: map[string]parser.NPC{
					"npc-1": {BaseExtractedRecord: parser.BaseExtractedRecord{ID: "npc-1", Type: "NPC_"}, Name: "Aela"},
					"npc-2": {BaseExtractedRecord: parser.BaseExtractedRecord{ID: "npc-2", Type: "NPC_"}, Name: "Farkas"},
				},
			},
		},
		personaGen,
		progress.NewNoopNotifier(),
		requestQueue,
		worker,
	)

	taskID, err := bridge.StartMasterPersonTask(StartMasterPersonTaskInput{SourceJSONPath: "dummy.json"})
	if err != nil {
		t.Fatalf("StartMasterPersonTask failed: %v", err)
	}
	_ = waitTaskStatus(t, bridge, taskID, StatusRequestGenerated, 3*time.Second)

	if err := bridge.ResumeMasterPersonaTask(taskID); err != nil {
		t.Fatalf("ResumeMasterPersonaTask (first) failed: %v", err)
	}
	_ = waitTaskStatus(t, bridge, taskID, StatusCompleted, 3*time.Second)
	firstCalls := personaGen.SaveCallCount()
	if firstCalls != 2 {
		t.Fatalf("expected 2 saved requests on first resume, got %d", firstCalls)
	}

	if err := bridge.ResumeMasterPersonaTask(taskID); err != nil {
		t.Fatalf("ResumeMasterPersonaTask (second) failed: %v", err)
	}
	_ = waitTaskStatus(t, bridge, taskID, StatusCompleted, 3*time.Second)
	secondCalls := personaGen.SaveCallCount()
	if secondCalls != firstCalls {
		t.Fatalf("expected no additional save calls on second resume: first=%d second=%d", firstCalls, secondCalls)
	}

	tasks, err := bridge.GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}
	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}
		if _, ok := task.Metadata["saved_request_ids"]; !ok {
			t.Fatalf("expected saved_request_ids metadata to be persisted")
		}
		return
	}
	t.Fatalf("task not found: %s", taskID)
}
