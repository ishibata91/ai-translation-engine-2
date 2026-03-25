package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	progress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	gatewayllm "github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/persona"
	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
	_ "github.com/mattn/go-sqlite3"
)

func TestMasterPersonaServiceRunPersonaPhaseBootstrapsFreshTask(t *testing.T) {
	ctx := context.Background()
	parser := &stubMasterPersonaParser{output: &skyrim.ParserOutput{}}
	generator := &stubMasterPersonaGenerator{
		prepareRequests: []llmio.Request{
			{
				SystemPrompt: "default-system",
				UserPrompt:   "default-user\n\nNPC Profile:\n- Name: Lydia",
				Temperature:  0.3,
				Metadata: map[string]interface{}{
					"source_plugin": "Skyrim.esm",
					"speaker_id":    "000A2C8E",
				},
			},
		},
	}
	service, queue, _, cleanup := newMasterPersonaServiceHarness(t, parser, generator, nil)
	defer cleanup()
	client := &stubMasterPersonaLLMClient{}
	service.worker = runtimequeue.NewWorker(
		queue,
		&stubMasterPersonaLLMManager{client: client},
		&stubMasterPersonaConfigStore{},
		&stubMasterPersonaSecretStore{},
		progress.NewNoopNotifier(),
		testLogger(),
	)

	input := PersonaExecutionInput{
		TaskID:            "translation-task-1",
		SourceJSONPath:    "dummy.json",
		OverwriteExisting: true,
		Request: TranslationRequestConfig{
			Provider:      "openai",
			Model:         "gpt-4.1-mini",
			Temperature:   0.7,
			BulkStrategy:  "sync",
			ContextLength: 8192,
		},
		Prompt: TranslationPromptConfig{
			UserPrompt:   "custom user prompt",
			SystemPrompt: "custom system prompt",
		},
	}

	if err := service.RunPersonaPhase(ctx, input); err != nil {
		t.Fatalf("RunPersonaPhase failed: %v", err)
	}

	requests, err := queue.GetTaskRequests(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected request count: got=%d want=%d", len(requests), 1)
	}

	var queued llmio.Request
	if err := json.Unmarshal([]byte(requests[0].RequestJSON), &queued); err != nil {
		t.Fatalf("unmarshal queued request failed: %v", err)
	}
	if queued.SystemPrompt != input.Prompt.SystemPrompt {
		t.Fatalf("unexpected system prompt: got=%q want=%q", queued.SystemPrompt, input.Prompt.SystemPrompt)
	}
	if !strings.HasPrefix(queued.UserPrompt, input.Prompt.UserPrompt) {
		t.Fatalf("user prompt override was not applied: got=%q", queued.UserPrompt)
	}
	if !strings.Contains(queued.UserPrompt, "\n\nNPC Profile:\n") {
		t.Fatalf("persona profile suffix must be preserved: got=%q", queued.UserPrompt)
	}
	if queued.Temperature != input.Request.Temperature {
		t.Fatalf("unexpected temperature: got=%f want=%f", queued.Temperature, input.Request.Temperature)
	}
	if got := strings.TrimSpace(metadataString(queued.Metadata, "execution_provider")); got != input.Request.Provider {
		t.Fatalf("unexpected execution provider metadata: got=%q want=%q", got, input.Request.Provider)
	}
	if got := strings.TrimSpace(metadataString(queued.Metadata, "execution_model")); got != input.Request.Model {
		t.Fatalf("unexpected execution model metadata: got=%q want=%q", got, input.Request.Model)
	}
	if client.completeCalls == 0 {
		t.Fatalf("expected bootstrap flow to continue into runtime execution")
	}
}

func TestMasterPersonaServiceRunPersonaPhaseResumesOnlyWhenRuntimeExists(t *testing.T) {
	ctx := context.Background()
	service, queue, _, cleanup := newMasterPersonaServiceHarness(t, nil, nil, nil)
	defer cleanup()

	taskWithRuntime := "translation-task-has-runtime"
	if err := queue.SubmitTaskSharedRequests(ctx, taskWithRuntime, string(task2.TypePersonaExtraction), []llmio.Request{
		{
			SystemPrompt: "system",
			UserPrompt:   "user",
			Metadata: map[string]interface{}{
				"source_plugin": "Skyrim.esm",
				"speaker_id":    "00012345",
			},
		},
	}); err != nil {
		t.Fatalf("SubmitTaskSharedRequests failed: %v", err)
	}

	err := service.RunPersonaPhase(ctx, PersonaExecutionInput{TaskID: taskWithRuntime})
	if err == nil {
		t.Fatalf("RunPersonaPhase unexpectedly succeeded without worker")
	}
	if !strings.Contains(err.Error(), "request queue worker is not configured") {
		t.Fatalf("resume path was not selected: err=%v", err)
	}

	emptyTaskID := "translation-task-empty"
	err = service.RunPersonaPhase(ctx, PersonaExecutionInput{TaskID: emptyTaskID})
	if err == nil {
		t.Fatalf("RunPersonaPhase unexpectedly succeeded for fresh task without source path")
	}
	if !strings.Contains(err.Error(), "source_json_path is required for persona bootstrap") {
		t.Fatalf("bootstrap validation was not selected: err=%v", err)
	}
}

func TestMasterPersonaProcessOptionsUsesTranslationFlowNamespaceAndOverrides(t *testing.T) {
	ctx := withPersonaPhaseRunConfig(context.Background(), TranslationRequestConfig{
		Provider: "lmstudio",
		Model:    "community-model",
		Endpoint: "http://127.0.0.1:1234",
	}, TranslationPromptConfig{})
	currentTask := &task2.Task{
		ID:   "translation-task",
		Type: task2.TypePersonaExtraction,
		Metadata: task2.TaskMetadata{
			"entrypoint": "translation_flow_persona_phase",
		},
	}

	opts := masterPersonaProcessOptions(ctx, currentTask)
	if opts.ConfigNamespace != translationFlowLLMNS {
		t.Fatalf("unexpected config namespace: got=%q want=%q", opts.ConfigNamespace, translationFlowLLMNS)
	}
	if opts.ConfigRead.Namespace != translationFlowLLMNS {
		t.Fatalf("unexpected config read namespace: got=%q want=%q", opts.ConfigRead.Namespace, translationFlowLLMNS)
	}
	if opts.ProviderOverride != "lmstudio" {
		t.Fatalf("unexpected provider override: got=%q", opts.ProviderOverride)
	}
	if opts.ModelOverride != "community-model" {
		t.Fatalf("unexpected model override: got=%q", opts.ModelOverride)
	}
	if opts.EndpointOverride != "http://127.0.0.1:1234" {
		t.Fatalf("unexpected endpoint override: got=%q", opts.EndpointOverride)
	}
}

func TestMasterPersonaServiceCleanupCompletedTaskRemovesTranslationProjectPersonaRuntime(t *testing.T) {
	ctx := context.Background()
	service, queue, _, cleanup := newMasterPersonaServiceHarness(t, nil, nil, nil)
	defer cleanup()

	taskID := "translation-project-persona"
	if err := queue.SubmitTaskSharedRequests(ctx, taskID, string(task2.TypePersonaExtraction), []llmio.Request{
		{
			SystemPrompt: "system",
			UserPrompt:   "user",
			Metadata: map[string]interface{}{
				"source_plugin": "Skyrim.esm",
				"speaker_id":    "00012345",
			},
		},
	}); err != nil {
		t.Fatalf("SubmitTaskSharedRequests failed: %v", err)
	}

	if err := service.CleanupCompletedTask(ctx, &task2.Task{
		ID:   taskID,
		Type: task2.TypeTranslationProject,
		Metadata: task2.TaskMetadata{
			"entrypoint": "translation_flow_persona_phase",
		},
	}); err != nil {
		t.Fatalf("CleanupCompletedTask failed: %v", err)
	}

	requests, err := queue.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("translation project completion must remove runtime requests: got=%d", len(requests))
	}
}

func TestMasterPersonaServiceCleanupCompletedTaskPreventsStaleRuntimeReuse(t *testing.T) {
	ctx := context.Background()
	service, queue, _, cleanup := newMasterPersonaServiceHarness(t, nil, nil, nil)
	defer cleanup()

	taskID := "translation-project-reuse"
	if err := queue.SubmitTaskSharedRequests(ctx, taskID, string(task2.TypePersonaExtraction), []llmio.Request{
		{
			SystemPrompt: "system",
			UserPrompt:   "user",
			Metadata: map[string]interface{}{
				"source_plugin": "Skyrim.esm",
				"speaker_id":    "00012345",
			},
		},
	}); err != nil {
		t.Fatalf("SubmitTaskSharedRequests failed: %v", err)
	}

	if err := service.CleanupCompletedTask(ctx, &task2.Task{
		ID:   taskID,
		Type: task2.TypeTranslationProject,
		Metadata: task2.TaskMetadata{
			"entrypoint": "translation_flow_persona_phase",
		},
	}); err != nil {
		t.Fatalf("CleanupCompletedTask failed: %v", err)
	}

	err := service.RunPersonaPhase(ctx, PersonaExecutionInput{
		TaskID: taskID,
	})
	if err == nil {
		t.Fatalf("RunPersonaPhase unexpectedly succeeded without bootstrap source path")
	}
	if !strings.Contains(err.Error(), "source_json_path is required for persona bootstrap") {
		t.Fatalf("stale runtime was reused unexpectedly: err=%v", err)
	}
}

func TestMasterPersonaServiceListPersonaRuntimeResolvesAcrossServiceRecreation(t *testing.T) {
	ctx := context.Background()
	service, queue, manager, cleanup := newMasterPersonaServiceHarness(t, nil, nil, nil)
	defer cleanup()

	taskID := "translation-task-runtime-snapshot"
	if err := queue.SubmitTaskSharedRequests(ctx, taskID, string(task2.TypePersonaExtraction), []llmio.Request{
		{
			SystemPrompt: "system-a",
			UserPrompt:   "user-a",
			Metadata: map[string]interface{}{
				"source_plugin": "Skyrim.esm",
				"speaker_id":    "00011111",
			},
		},
		{
			SystemPrompt: "system-b",
			UserPrompt:   "user-b",
			Metadata: map[string]interface{}{
				"speaker_id": "00022222",
			},
		},
	}); err != nil {
		t.Fatalf("SubmitTaskSharedRequests failed: %v", err)
	}

	requests, err := queue.GetTaskRequests(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskRequests failed: %v", err)
	}
	if len(requests) == 0 {
		t.Fatalf("expected seeded requests")
	}
	errorMessage := "provider timeout"
	if err := queue.UpdateJob(ctx, requests[0].ID, runtimequeue.StatusFailed, nil, &errorMessage, nil); err != nil {
		t.Fatalf("UpdateJob failed: %v", err)
	}

	first, err := service.ListPersonaRuntime(ctx, taskID)
	if err != nil {
		t.Fatalf("ListPersonaRuntime first service failed: %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("unexpected runtime entry count: got=%d want=%d", len(first), 2)
	}

	recreated := NewMasterPersonaService(
		manager,
		testLogger(),
		nil,
		nil,
		nil,
		queue,
		nil,
	)
	second, err := recreated.ListPersonaRuntime(ctx, taskID)
	if err != nil {
		t.Fatalf("ListPersonaRuntime recreated service failed: %v", err)
	}
	if len(second) != len(first) {
		t.Fatalf("runtime snapshot must be stable after recreation: first=%d second=%d", len(first), len(second))
	}

	firstByID := map[string]PersonaRuntimeEntry{}
	for _, entry := range first {
		firstByID[entry.RequestID] = entry
	}
	for _, entry := range second {
		prev, ok := firstByID[entry.RequestID]
		if !ok {
			t.Fatalf("unexpected runtime entry after recreation: request_id=%s", entry.RequestID)
		}
		if entry.SourcePlugin != prev.SourcePlugin || entry.SpeakerID != prev.SpeakerID || entry.RequestState != prev.RequestState || entry.ErrorMessage != prev.ErrorMessage {
			t.Fatalf("runtime entry mismatch after recreation: request_id=%s first=%+v second=%+v", entry.RequestID, prev, entry)
		}
	}
}

func newMasterPersonaServiceHarness(
	t *testing.T,
	parser skyrim.Parser,
	generator persona.NPCPersonaGenerator,
	worker *runtimequeue.Worker,
) (*MasterPersonaService, *runtimequeue.Queue, *task2.Manager, func()) {
	t.Helper()

	ctx := context.Background()
	queue, err := runtimequeue.NewQueue(ctx, ":memory:", testLogger())
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}

	taskDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open task db: %v", err)
	}
	if err := task2.Migrate(ctx, taskDB); err != nil {
		t.Fatalf("migrate task db: %v", err)
	}

	// Use nil manager context in tests to avoid Wails runtime event emission.
	manager := task2.NewManager(nil, testLogger(), task2.NewStore(taskDB))
	service := NewMasterPersonaService(manager, testLogger(), parser, generator, nil, queue, worker)
	cleanup := func() {
		_ = queue.Close()
		_ = taskDB.Close()
	}
	return service, queue, manager, cleanup
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubMasterPersonaParser struct {
	output *skyrim.ParserOutput
	err    error
}

func (s *stubMasterPersonaParser) LoadExtractedJSON(ctx context.Context, path string) (*skyrim.ParserOutput, error) {
	_ = ctx
	_ = path
	if s.err != nil {
		return nil, s.err
	}
	if s.output == nil {
		return &skyrim.ParserOutput{}, nil
	}
	return s.output, nil
}

type stubMasterPersonaGenerator struct {
	prepareRequests []llmio.Request
	prepareErr      error
	saveErr         error
}

func (s *stubMasterPersonaGenerator) ID() string {
	return "stub_persona_generator"
}

func (s *stubMasterPersonaGenerator) PreparePrompts(ctx context.Context, input any) ([]llmio.Request, error) {
	_ = ctx
	_ = input
	if s.prepareErr != nil {
		return nil, s.prepareErr
	}
	return append([]llmio.Request(nil), s.prepareRequests...), nil
}

func (s *stubMasterPersonaGenerator) SaveResults(ctx context.Context, responses []llmio.Response) error {
	_ = ctx
	_ = responses
	return s.saveErr
}

// RunPersonaPhase keeps legacy translation-flow tests compiling after MasterPersona contract expansion.
func (s *stubMasterPersona) RunPersonaPhase(ctx context.Context, input PersonaExecutionInput) error {
	return s.ResumeMasterPersona(withPersonaPhaseRunConfig(ctx, input.Request, input.Prompt), input.TaskID)
}

// ListPersonaRuntime keeps legacy translation-flow tests compiling after MasterPersona contract expansion.
func (s *stubMasterPersona) ListPersonaRuntime(ctx context.Context, taskID string) ([]PersonaRuntimeEntry, error) {
	requests, err := s.GetTaskRequests(ctx, taskID)
	if err != nil {
		return nil, err
	}
	entries := make([]PersonaRuntimeEntry, 0, len(requests))
	for _, request := range requests {
		sourcePlugin, speakerID, hasLookupKey := parsePersonaRuntimeLookupFromRequestJSON(request.RequestJSON)
		entry := PersonaRuntimeEntry{
			RequestID:    request.ID,
			SourcePlugin: sourcePlugin,
			SpeakerID:    speakerID,
			RequestState: request.RequestState,
			ResumeCursor: request.ResumeCursor,
			UpdatedAt:    request.UpdatedAt,
			HasResponse:  request.ResponseJSON != nil,
			HasLookupKey: hasLookupKey,
		}
		if request.ErrorMessage != nil {
			entry.ErrorMessage = strings.TrimSpace(*request.ErrorMessage)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

type stubMasterPersonaLLMManager struct {
	client gatewayllm.LLMClient
}

func (s *stubMasterPersonaLLMManager) GetClient(ctx context.Context, config gatewayllm.LLMConfig) (gatewayllm.LLMClient, error) {
	_ = ctx
	_ = config
	if s.client == nil {
		return nil, fmt.Errorf("llm client is not configured")
	}
	return s.client, nil
}

func (s *stubMasterPersonaLLMManager) GetBatchClient(ctx context.Context, config gatewayllm.LLMConfig) (gatewayllm.BatchClient, error) {
	_ = ctx
	_ = config
	return nil, fmt.Errorf("batch client is not configured")
}

func (s *stubMasterPersonaLLMManager) ResolveBulkStrategy(ctx context.Context, strategy gatewayllm.BulkStrategy, provider string) gatewayllm.BulkStrategy {
	_ = ctx
	_ = provider
	if strings.TrimSpace(string(strategy)) == "" {
		return gatewayllm.BulkStrategySync
	}
	return strategy
}

type stubMasterPersonaLLMClient struct {
	completeCalls int
}

func (s *stubMasterPersonaLLMClient) ListModels(ctx context.Context) ([]gatewayllm.ModelInfo, error) {
	_ = ctx
	return []gatewayllm.ModelInfo{{ID: "stub-model", DisplayName: "stub-model"}}, nil
}

func (s *stubMasterPersonaLLMClient) Complete(ctx context.Context, req gatewayllm.Request) (gatewayllm.Response, error) {
	_ = ctx
	_ = req
	s.completeCalls++
	return gatewayllm.Response{Success: true, Content: `{"persona":"ok"}`}, nil
}

func (s *stubMasterPersonaLLMClient) GenerateStructured(ctx context.Context, req gatewayllm.Request) (gatewayllm.Response, error) {
	return s.Complete(ctx, req)
}

func (s *stubMasterPersonaLLMClient) StreamComplete(ctx context.Context, req gatewayllm.Request) (gatewayllm.StreamResponse, error) {
	_ = ctx
	_ = req
	return &stubMasterPersonaStreamResponse{}, nil
}

func (s *stubMasterPersonaLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	_ = text
	return nil, nil
}

func (s *stubMasterPersonaLLMClient) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

type stubMasterPersonaStreamResponse struct{}

func (s *stubMasterPersonaStreamResponse) Next() (gatewayllm.Response, bool) {
	return gatewayllm.Response{}, false
}

func (s *stubMasterPersonaStreamResponse) Close() error {
	return nil
}

type stubMasterPersonaConfigStore struct{}

func (s *stubMasterPersonaConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	_ = ctx
	_ = namespace
	_ = key
	return "", nil
}

type stubMasterPersonaSecretStore struct{}

func (s *stubMasterPersonaSecretStore) GetSecret(ctx context.Context, namespace string, key string) (string, error) {
	_ = ctx
	_ = namespace
	_ = key
	return "", nil
}
