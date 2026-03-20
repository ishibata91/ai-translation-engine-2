package workflow

import (
	"context"
	"errors"
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
	return nil
}

type stubTerminologyExecutor struct {
	err       error
	responses []llmio.Response
}

func (s *stubTerminologyExecutor) Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error) {
	_ = ctx
	_ = config
	_ = requests
	return s.responses, s.err
}

type stubWorkflowProgressNotifier struct {
	events []runtimeprogress.ProgressEvent
}

func (s *stubWorkflowProgressNotifier) OnProgress(ctx context.Context, event runtimeprogress.ProgressEvent) {
	_ = ctx
	s.events = append(s.events, event)
}
