package task

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
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
	reqs []llm.Request
	err  error
}

func (m *mockPersonaGenerator) ID() string { return "PersonaMock" }
func (m *mockPersonaGenerator) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reqs, nil
}
func (m *mockPersonaGenerator) SaveResults(ctx context.Context, responses []llm.Response) error {
	return nil
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
