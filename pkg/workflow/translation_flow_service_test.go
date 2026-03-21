package workflow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	terminologyslice "github.com/ishibata91/ai-translation-engine-2/pkg/slice/terminology"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
)

func TestTranslationFlowServiceListTerminologyTargetsIncludesTranslations(t *testing.T) {
	service := &TranslationFlowService{
		store: &stubTranslationFlowStore{
			input: translationinput.TerminologyInput{
				TaskID: "task-1",
				Entries: []translationinput.TerminologyEntry{
					{
						ID:         "row-1",
						EditorID:   "EDID_001",
						RecordType: "NPC_:FULL",
						SourceText: "NPC Name",
						SourceFile: "Update.esm.extract.json",
						Variant:    "full",
					},
					{
						ID:         "row-2",
						EditorID:   "EDID_002",
						RecordType: "BOOK:FULL",
						SourceText: "Unreadable Book",
						SourceFile: "Skyrim.esm.extract.json",
						Variant:    "single",
					},
				},
			},
		},
		terminology: &stubTerminology{
			previewTranslations: map[string]terminologyslice.PreviewTranslation{
				"row-1": {
					RowID:            "row-1",
					TranslatedText:   "NPC 名",
					TranslationState: "translated",
				},
				"row-2": {
					RowID:            "row-2",
					TranslationState: "missing",
				},
			},
		},
	}

	page, err := service.ListTerminologyTargets(context.Background(), "task-1", 1, 50)
	if err != nil {
		t.Fatalf("ListTerminologyTargets failed: %v", err)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("unexpected row count: got=%d want=%d", len(page.Rows), 2)
	}
	if page.Rows[0].TranslatedText != "NPC 名" {
		t.Fatalf("unexpected translated text: got=%q want=%q", page.Rows[0].TranslatedText, "NPC 名")
	}
	if page.Rows[0].TranslationState != "translated" {
		t.Fatalf("unexpected translated state: got=%q want=%q", page.Rows[0].TranslationState, "translated")
	}
	if page.Rows[1].TranslationState != "missing" {
		t.Fatalf("unexpected missing state: got=%q want=%q", page.Rows[1].TranslationState, "missing")
	}
}

func TestTranslationFlowServiceLoadFilesResetsTerminologySummary(t *testing.T) {
	terminology := &stubTerminology{}
	service := &TranslationFlowService{
		parser: &stubSkyrimParser{
			output: &skyrim.ParserOutput{},
		},
		store:       &stubTranslationFlowStore{},
		terminology: terminology,
	}

	_, err := service.LoadFiles(context.Background(), LoadTranslationFlowInput{
		TaskID:    "task-reset",
		FilePaths: []string{"example.json"},
	})
	if err != nil {
		t.Fatalf("LoadFiles failed: %v", err)
	}
	if terminology.updatedSummary.TaskID != "task-reset" {
		t.Fatalf("unexpected reset task id: got=%q want=%q", terminology.updatedSummary.TaskID, "task-reset")
	}
	if terminology.updatedSummary.Status != "pending" {
		t.Fatalf("unexpected reset status: got=%q want=%q", terminology.updatedSummary.Status, "pending")
	}
	if terminology.updatedSummary.ProgressMode != "hidden" {
		t.Fatalf("unexpected reset progress mode: got=%q want=%q", terminology.updatedSummary.ProgressMode, "hidden")
	}
	if terminology.updatedSummary.SavedCount != 0 || terminology.updatedSummary.FailedCount != 0 || terminology.updatedSummary.ProgressTotal != 0 {
		t.Fatalf("unexpected reset counters: %+v", terminology.updatedSummary)
	}
}

func TestTranslationFlowServiceRunTerminologyPhaseMarksRunError(t *testing.T) {
	terminology := &stubTerminology{
		preparePromptsResult: []llmio.Request{
			{
				Metadata: map[string]interface{}{
					"source_text": "NPC Name",
				},
			},
		},
		summary: terminologyslice.PhaseSummary{
			TaskID:          "task-2",
			Status:          "running",
			TargetCount:     1,
			ProgressMode:    "indeterminate",
			ProgressTotal:   1,
			ProgressMessage: "単語翻訳を実行中",
		},
	}
	service := &TranslationFlowService{
		terminology: terminology,
		executor: &stubTerminologyExecutor{
			err: errors.New("executor failed"),
		},
		notifier: &stubWorkflowProgressNotifier{},
	}

	_, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
		TaskID: "task-2",
		Request: TranslationRequestConfig{
			Model: "gemini-2.5-flash",
		},
	})
	if err == nil {
		t.Fatalf("RunTerminologyPhase unexpectedly succeeded")
	}
	if terminology.updatedSummary.Status != "run_error" {
		t.Fatalf("unexpected updated status: got=%q want=%q", terminology.updatedSummary.Status, "run_error")
	}
	if terminology.updatedSummary.ProgressMode != "hidden" {
		t.Fatalf("unexpected updated progress mode: got=%q want=%q", terminology.updatedSummary.ProgressMode, "hidden")
	}
}

func TestTranslationFlowServiceRunTerminologyPhasePublishesRunningProgress(t *testing.T) {
	notifier := &stubWorkflowProgressNotifier{}
	service := &TranslationFlowService{
		terminology: &stubTerminology{
			preparePromptsResult: []llmio.Request{
				{Metadata: map[string]interface{}{"source_text": "NPC Name"}},
			},
			summary: terminologyslice.PhaseSummary{
				TaskID:          "task-3",
				Status:          "completed",
				SavedCount:      1,
				ProgressMode:    "hidden",
				ProgressCurrent: 1,
				ProgressTotal:   1,
				ProgressMessage: "単語翻訳完了",
			},
		},
		executor: &stubTerminologyExecutor{
			responses: []llmio.Response{{Success: true}},
		},
		notifier: notifier,
	}

	if _, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
		TaskID: "task-3",
		Request: TranslationRequestConfig{
			Model: "gemini-2.5-flash",
		},
	}); err != nil {
		t.Fatalf("RunTerminologyPhase failed: %v", err)
	}

	if len(notifier.events) < 2 {
		t.Fatalf("unexpected event count: got=%d want>=%d", len(notifier.events), 2)
	}
	if notifier.events[0].Status != runtimeprogress.StatusInProgress {
		t.Fatalf("unexpected first event status: got=%q want=%q", notifier.events[0].Status, runtimeprogress.StatusInProgress)
	}
	if notifier.events[0].TaskID != "task-3" {
		t.Fatalf("unexpected first event task id: got=%q want=%q", notifier.events[0].TaskID, "task-3")
	}
}

func TestTranslationFlowServiceRunTerminologyPhasePublishesIntermediateProgress(t *testing.T) {
	notifier := &stubWorkflowProgressNotifier{}
	service := &TranslationFlowService{
		terminology: &stubTerminology{
			preparePromptsResult: []llmio.Request{
				{Metadata: map[string]interface{}{"source_text": "A"}},
				{Metadata: map[string]interface{}{"source_text": "B"}},
				{Metadata: map[string]interface{}{"source_text": "C"}},
			},
			summary: terminologyslice.PhaseSummary{
				TaskID:          "task-4",
				Status:          "completed",
				SavedCount:      3,
				ProgressMode:    "hidden",
				ProgressCurrent: 3,
				ProgressTotal:   3,
				ProgressMessage: "単語翻訳完了",
			},
		},
		executor: &stubTerminologyExecutor{
			responses: []llmio.Response{{Success: true}, {Success: true}, {Success: true}},
			steps:     []int{1, 2, 3},
		},
		notifier: notifier,
	}

	if _, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
		TaskID: "task-4",
		Request: TranslationRequestConfig{
			Model: "gemini-2.5-flash",
		},
	}); err != nil {
		t.Fatalf("RunTerminologyPhase failed: %v", err)
	}

	if len(notifier.events) < 4 {
		t.Fatalf("unexpected event count: got=%d want>=%d", len(notifier.events), 4)
	}
	if notifier.events[1].Current != 1 || notifier.events[1].Total != 3 {
		t.Fatalf("unexpected intermediate progress: current=%d total=%d", notifier.events[1].Current, notifier.events[1].Total)
	}
}

func TestTranslationFlowServiceRunTerminologyPhaseThrottlesSummaryPersistence(t *testing.T) {
	total := 500
	steps := make([]int, 0, total)
	requests := make([]llmio.Request, 0, total)
	for i := 0; i < total; i++ {
		steps = append(steps, i+1)
		requests = append(requests, llmio.Request{Metadata: map[string]interface{}{"source_text": fmt.Sprintf("row-%d", i)}})
	}

	terminology := &stubTerminology{
		preparePromptsResult: requests,
		summary: terminologyslice.PhaseSummary{
			TaskID:          "task-5",
			Status:          "completed",
			SavedCount:      total,
			ProgressMode:    "hidden",
			ProgressCurrent: total,
			ProgressTotal:   total,
			ProgressMessage: "単語翻訳完了",
		},
	}
	notifier := &stubWorkflowProgressNotifier{}
	service := &TranslationFlowService{
		terminology: terminology,
		executor: &stubTerminologyExecutor{
			responses: make([]llmio.Response, total),
			steps:     steps,
		},
		notifier: notifier,
	}

	if _, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
		TaskID: "task-5",
		Request: TranslationRequestConfig{
			Model: "gemini-2.5-flash",
		},
	}); err != nil {
		t.Fatalf("RunTerminologyPhase failed: %v", err)
	}

	if len(terminology.updatedSummaries) >= total {
		t.Fatalf("summary persistence must be throttled: writes=%d total=%d", len(terminology.updatedSummaries), total)
	}
	if len(terminology.updatedSummaries) <= 1 {
		t.Fatalf("summary persistence must keep intermediate snapshots: writes=%d", len(terminology.updatedSummaries))
	}

	progressValues := make([]int, 0, len(notifier.events))
	for _, event := range notifier.events {
		if event.Status != runtimeprogress.StatusInProgress {
			continue
		}
		progressValues = append(progressValues, event.Current)
	}
	if len(progressValues) == 0 {
		t.Fatalf("expected in-progress events")
	}
	if !slices.Contains(progressValues, total) {
		t.Fatalf("expected progress events to include completion value %d", total)
	}
}

func TestTranslationFlowServiceRunTerminologyPhaseThrottlesSummaryPersistenceForSmallTotals(t *testing.T) {
	testCases := []struct {
		name  string
		total int
	}{
		{name: "three_rows", total: 3},
		{name: "ten_rows", total: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := make([]int, 0, tc.total)
			requests := make([]llmio.Request, 0, tc.total)
			for i := 0; i < tc.total; i++ {
				steps = append(steps, i+1)
				requests = append(requests, llmio.Request{Metadata: map[string]interface{}{"source_text": fmt.Sprintf("row-%d", i)}})
			}

			terminology := &stubTerminology{
				preparePromptsResult: requests,
				summary: terminologyslice.PhaseSummary{
					TaskID:          "task-small",
					Status:          "completed",
					SavedCount:      tc.total,
					ProgressMode:    "hidden",
					ProgressCurrent: tc.total,
					ProgressTotal:   tc.total,
					ProgressMessage: "単語翻訳完了",
				},
			}
			service := &TranslationFlowService{
				terminology: terminology,
				executor: &stubTerminologyExecutor{
					responses: make([]llmio.Response, tc.total),
					steps:     steps,
				},
				notifier: &stubWorkflowProgressNotifier{},
			}

			if _, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
				TaskID: "task-small",
				Request: TranslationRequestConfig{
					Model: "gemini-2.5-flash",
				},
			}); err != nil {
				t.Fatalf("RunTerminologyPhase failed: %v", err)
			}

			if len(terminology.updatedSummaries) >= tc.total {
				t.Fatalf("summary persistence must be throttled for small totals: writes=%d total=%d", len(terminology.updatedSummaries), tc.total)
			}
		})
	}
}

type stubTranslationFlowStore struct {
	input translationinput.TerminologyInput
}

type stubSkyrimParser struct {
	output *skyrim.ParserOutput
	err    error
}

func (s *stubSkyrimParser) LoadExtractedJSON(ctx context.Context, path string) (*skyrim.ParserOutput, error) {
	_ = ctx
	_ = path
	if s.err != nil {
		return nil, s.err
	}
	return s.output, nil
}

func (s *stubTranslationFlowStore) EnsureTask(ctx context.Context, taskID string) error {
	_ = ctx
	_ = taskID
	return nil
}

func (s *stubTranslationFlowStore) SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (translationflow.LoadedFile, error) {
	_ = ctx
	_ = taskID
	_ = sourceFilePath
	_ = output
	return translationflow.LoadedFile{}, nil
}

func (s *stubTranslationFlowStore) ListFiles(ctx context.Context, taskID string) ([]translationflow.LoadedFile, error) {
	_ = ctx
	_ = taskID
	return nil, nil
}

func (s *stubTranslationFlowStore) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (translationflow.PreviewPage, error) {
	_ = ctx
	_ = fileID
	_ = page
	_ = pageSize
	return translationflow.PreviewPage{}, nil
}

func (s *stubTranslationFlowStore) LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error) {
	_ = ctx
	_ = taskID
	return s.input, nil
}

type stubTerminology struct {
	preparePromptsResult []llmio.Request
	previewTranslations  map[string]terminologyslice.PreviewTranslation
	summary              terminologyslice.PhaseSummary
	updatedSummary       terminologyslice.PhaseSummary
	updatedSummaries     []terminologyslice.PhaseSummary
}

func (s *stubTerminology) ID() string {
	return "Terminology"
}

func (s *stubTerminology) PreparePrompts(ctx context.Context, taskID string, options terminologyslice.PhaseOptions) ([]llmio.Request, error) {
	_ = ctx
	_ = taskID
	_ = options
	return s.preparePromptsResult, nil
}

func (s *stubTerminology) SaveResults(ctx context.Context, taskID string, responses []llmio.Response) error {
	_ = ctx
	_ = taskID
	_ = responses
	return nil
}

func (s *stubTerminology) GetPhaseSummary(ctx context.Context, taskID string) (terminologyslice.PhaseSummary, error) {
	_ = ctx
	_ = taskID
	return s.summary, nil
}

func (s *stubTerminology) GetPreviewTranslations(ctx context.Context, entries []terminologyslice.TerminologyEntry) (map[string]terminologyslice.PreviewTranslation, error) {
	_ = ctx
	_ = entries
	return s.previewTranslations, nil
}

func (s *stubTerminology) UpdatePhaseSummary(ctx context.Context, summary terminologyslice.PhaseSummary) error {
	_ = ctx
	s.updatedSummary = summary
	s.updatedSummaries = append(s.updatedSummaries, summary)
	return nil
}

type stubTerminologyExecutor struct {
	err       error
	responses []llmio.Response
	steps     []int
}

func (s *stubTerminologyExecutor) Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error) {
	_ = ctx
	_ = config
	_ = requests
	return s.responses, s.err
}

func (s *stubTerminologyExecutor) ExecuteWithProgress(
	ctx context.Context,
	config llmio.ExecutionConfig,
	requests []llmio.Request,
	progress func(completed, total int),
) ([]llmio.Response, error) {
	_ = ctx
	_ = config
	total := len(requests)
	for _, step := range s.steps {
		progress(step, total)
	}
	return s.responses, s.err
}

type stubWorkflowProgressNotifier struct {
	events []runtimeprogress.ProgressEvent
}

func (s *stubWorkflowProgressNotifier) OnProgress(ctx context.Context, event runtimeprogress.ProgressEvent) {
	_ = ctx
	s.events = append(s.events, event)
}
