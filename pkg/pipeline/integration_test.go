package pipeline

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/progress"
)

// mockSlice implements Slice interface for testing.
type mockSlice struct {
	id           string
	prepareCalls int
	saveCalls    int
	results      []llm.Response
}

func (s *mockSlice) ID() string { return s.id }
func (s *mockSlice) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	s.prepareCalls++
	return []llm.Request{{UserPrompt: "test prompt"}}, nil
}
func (s *mockSlice) SaveResults(ctx context.Context, results []llm.Response) error {
	slog.Info("SaveResults called", slog.String("slice", s.id), slog.Int("results", len(results)))
	s.saveCalls++
	s.results = results
	return nil
}

// Integration test for Pipeline orchestration and recovery.
func TestProcessManager_Integration(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Setup Infrastructure
	q, _ := queue.NewQueue(ctx, ":memory:", logger)
	defer q.Close()

	// Using mocks for other infra (Worker needs real queue but mocked LLM)
	mockLLM := &mockLLMManagerForPM{}
	worker := queue.NewWorker(q, mockLLM, &mockCfgForPM{}, &mockSecForPM{}, &mockNotifierForPM{}, logger)
	worker.SetPollingInterval(10 * time.Millisecond)

	// Setup Store and Manager
	store, _ := NewStore(ctx, ":memory:")
	defer store.Close()

	manager := NewManager(store, q, worker, logger)

	// Register Slice
	slice := &mockSlice{id: "TestSlice"}
	manager.RegisterSlice(slice)

	// 1. Execute Slice (Full Flow)
	processID, err := manager.ExecuteSlice(ctx, "TestSlice", "test-input", "test.json")
	if err != nil {
		t.Fatalf("ExecuteSlice failed: %v", err)
	}

	// 2. Wait for completion (background worker -> handleCompletion)
	// In memory, this should be very fast.
	time.Sleep(100 * time.Millisecond)

	if slice.saveCalls != 1 {
		t.Errorf("Expected SaveResults to be called once, got %d", slice.saveCalls)
	}
	if len(slice.results) != 1 {
		t.Errorf("Expected 1 result saved, got %d", len(slice.results))
	}

	// 3. Verify state cleanup
	state, _ := store.GetState(ctx, processID)
	if state != nil {
		t.Errorf("Expected state to be deleted after completion, but it exists")
	}

	// 4. Test Recovery
	// Manually inject a state and a job
	recoverPID := "recover-123"
	store.SaveState(ctx, ProcessState{
		ProcessID:    recoverPID,
		TargetSlice:  "TestSlice",
		InputFile:    "recover.json",
		CurrentPhase: PhaseDispatched,
	})
	q.SubmitJobs(ctx, recoverPID, []any{llm.Request{UserPrompt: "recover me"}})

	// Trigger Recover
	err = manager.Recover(ctx)
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	// Wait for background resume
	time.Sleep(100 * time.Millisecond)

	if slice.saveCalls != 2 {
		t.Errorf("Expected SaveResults to be called again after recovery, total %d", slice.saveCalls)
	}
}

// Mocks for PM Integration Test
type mockLLMManagerForPM struct{}

func (m *mockLLMManagerForPM) GetClient(ctx context.Context, config llm.LLMConfig) (llm.LLMClient, error) {
	return &mockClientForPM{}, nil
}
func (m *mockLLMManagerForPM) GetBatchClient(ctx context.Context, config llm.LLMConfig) (llm.BatchClient, error) {
	return nil, nil
}
func (m *mockLLMManagerForPM) ResolveBulkStrategy(ctx context.Context, strategy llm.BulkStrategy, provider string) llm.BulkStrategy {
	return llm.BulkStrategySync
}

type mockClientForPM struct{}

func (c *mockClientForPM) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{Success: true, Content: "done"}, nil
}
func (c *mockClientForPM) StreamComplete(ctx context.Context, req llm.Request) (llm.StreamResponse, error) {
	return nil, nil
}
func (c *mockClientForPM) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}
func (c *mockClientForPM) HealthCheck(ctx context.Context) error { return nil }

type mockCfgForPM struct{}

func (m *mockCfgForPM) Get(ctx context.Context, ns, key string) (string, error) { return "sync", nil }
func (m *mockCfgForPM) Set(ctx context.Context, ns, key, val string) error      { return nil }
func (m *mockCfgForPM) Delete(ctx context.Context, ns, key string) error        { return nil }
func (m *mockCfgForPM) GetAll(ctx context.Context, ns string) (map[string]string, error) {
	return nil, nil
}
func (m *mockCfgForPM) Watch(ns, key string, cb config.ChangeCallback) config.UnsubscribeFunc {
	return func() {}
}

type mockSecForPM struct{}

func (m *mockSecForPM) GetSecret(ctx context.Context, ns, key string) (string, error) {
	return "key", nil
}
func (m *mockSecForPM) SetSecret(ctx context.Context, ns, key, val string) error { return nil }
func (m *mockSecForPM) DeleteSecret(ctx context.Context, ns, key string) error   { return nil }
func (m *mockSecForPM) ListSecretKeys(ctx context.Context, ns string) ([]string, error) {
	return nil, nil
}

type mockNotifierForPM struct{}

func (m *mockNotifierForPM) OnProgress(ctx context.Context, event progress.ProgressEvent) {}
