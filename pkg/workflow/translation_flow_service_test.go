package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"

	dictionaryartifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/dictionary_artifact"
	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	terminologyslice "github.com/ishibata91/ai-translation-engine-2/pkg/slice/terminology"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
	_ "modernc.org/sqlite"
)

func TestTranslationFlowServiceListTerminologyTargetsIncludesTranslations(t *testing.T) {
	service := &TranslationFlowService{
		store: &stubTranslationFlowStore{},
		terminology: &stubTerminology{
			listTargetsResult: []terminologyslice.TerminologyEntry{
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

func TestTranslationFlowServiceListTerminologyTargetsUsesNormalizedTargets(t *testing.T) {
	service := &TranslationFlowService{
		store: &stubTranslationFlowStore{},
		terminology: &stubTerminology{
			listTargetsResult: []terminologyslice.TerminologyEntry{
				{
					ID:         "row-en",
					EditorID:   "EDID_EN",
					RecordType: "BOOK:FULL",
					SourceText: "Iron Sword",
					SourceFile: "Skyrim.esm.extract.json",
					Variant:    "single",
				},
			},
			previewTranslations: map[string]terminologyslice.PreviewTranslation{
				"row-en": {RowID: "row-en", TranslationState: "missing"},
			},
		},
	}

	page, err := service.ListTerminologyTargets(context.Background(), "task-norm", 1, 50)
	if err != nil {
		t.Fatalf("ListTerminologyTargets failed: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Fatalf("unexpected row count: got=%d want=%d", len(page.Rows), 1)
	}
	if page.Rows[0].ID != "row-en" {
		t.Fatalf("unexpected normalized row id: got=%q want=%q", page.Rows[0].ID, "row-en")
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
				SavedCount:      0,
				ProgressMode:    "hidden",
				ProgressCurrent: 0,
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
				SavedCount:      0,
				ProgressMode:    "hidden",
				ProgressCurrent: 0,
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
			SavedCount:      0,
			ProgressMode:    "hidden",
			ProgressCurrent: 0,
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
					SavedCount:      0,
					ProgressMode:    "hidden",
					ProgressCurrent: 0,
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

func TestTranslationFlowServiceRunTerminologyPhaseRestoresFinalSummaryAfterRunningSnapshotReset(t *testing.T) {
	testCases := []struct {
		name       string
		responses  []scriptedTerminologyResponse
		wantStatus string
		wantSaved  int
		wantFailed int
	}{
		{
			name: "completed",
			responses: []scriptedTerminologyResponse{
				{success: true, content: "TL: |鋼鉄の鎧|"},
				{success: true, content: "TL: |銀の盾|"},
				{success: true, content: "TL: |黒檀の弓|"},
			},
			wantStatus: "completed",
			wantSaved:  4,
			wantFailed: 0,
		},
		{
			name: "completed_partial",
			responses: []scriptedTerminologyResponse{
				{success: true, content: "TL: |鋼鉄の鎧|"},
				{success: false, err: "provider timeout"},
				{success: true, content: "TL: |黒檀の弓|"},
			},
			wantStatus: "completed_partial",
			wantSaved:  3,
			wantFailed: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			terminology, store, cleanup := newWorkflowTerminologyHarness(t, "task-cached-mixed")
			defer cleanup()

			service := &TranslationFlowService{
				terminology: terminology,
				executor: &scriptedTerminologyExecutor{
					responses: tc.responses,
					steps:     []int{1, 2, 3},
				},
				notifier: &stubWorkflowProgressNotifier{},
			}

			result, err := service.RunTerminologyPhase(context.Background(), RunTerminologyPhaseInput{
				TaskID: "task-cached-mixed",
				Request: TranslationRequestConfig{
					Model: "gemini-2.5-flash",
				},
			})
			if err != nil {
				t.Fatalf("RunTerminologyPhase failed: %v", err)
			}
			if result.Status != tc.wantStatus {
				t.Fatalf("unexpected status: got=%q want=%q", result.Status, tc.wantStatus)
			}
			if result.SavedCount != tc.wantSaved {
				t.Fatalf("unexpected saved count: got=%d want=%d", result.SavedCount, tc.wantSaved)
			}
			if result.FailedCount != tc.wantFailed {
				t.Fatalf("unexpected failed count: got=%d want=%d", result.FailedCount, tc.wantFailed)
			}

			foundResetSnapshot := false
			for _, summary := range store.updatedSummaries {
				if summary.Status == "running" && summary.SavedCount == 0 && summary.ProgressCurrent == 3 {
					foundResetSnapshot = true
					break
				}
			}
			if !foundResetSnapshot {
				t.Fatalf("running snapshot with saved_count reset was not persisted: %+v", store.updatedSummaries)
			}
		})
	}
}

func newWorkflowTerminologyHarness(t *testing.T, taskID string) (terminologyslice.Terminology, *recordingTerminologyStore, func()) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	dictDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open dictionary db: %v", err)
	}
	dictDB.SetMaxOpenConns(1)
	if _, err := dictDB.Exec(`
		CREATE TABLE artifact_dictionary_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER DEFAULT 0,
			edid TEXT DEFAULT '',
			source_text TEXT,
			dest_text TEXT,
			record_type TEXT
		);
		INSERT INTO artifact_dictionary_entries (source_text, dest_text, record_type) VALUES
			('Iron Sword', '鉄の剣', 'BOOK:FULL');
	`); err != nil {
		t.Fatalf("failed to seed dictionary db: %v", err)
	}

	modDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		_ = dictDB.Close()
		t.Fatalf("failed to open terminology db: %v", err)
	}
	modDB.SetMaxOpenConns(1)

	repo := &workflowTerminologyInputRepository{
		inputs: map[string]translationinput.TerminologyInput{
			taskID: {
				TaskID: taskID,
				Entries: []translationinput.TerminologyEntry{
					{
						ID:         "row-1",
						EditorID:   "EDID_001",
						RecordType: "BOOK:FULL",
						SourceText: "Iron Sword",
						SourceFile: "workflow_cached_mixed.json",
						Variant:    "single",
					},
					{
						ID:         "row-2",
						EditorID:   "EDID_002",
						RecordType: "ARMO:FULL",
						SourceText: "Steel Armor",
						SourceFile: "workflow_cached_mixed.json",
						Variant:    "single",
					},
					{
						ID:         "row-3",
						EditorID:   "EDID_003",
						RecordType: "BOOK:FULL",
						SourceText: "Silver Shield",
						SourceFile: "workflow_cached_mixed.json",
						Variant:    "single",
					},
					{
						ID:         "row-4",
						EditorID:   "EDID_004",
						RecordType: "WEAP:FULL",
						SourceText: "Ebony Bow",
						SourceFile: "workflow_cached_mixed.json",
						Variant:    "single",
					},
				},
			},
		},
	}
	builder := terminologyslice.NewTermRequestBuilder(&terminologyslice.TermRecordConfig{
		TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...),
	})
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := terminologyslice.NewSQLiteTermDictionarySearcher(dictRepo, logger, terminologyslice.NewSnowballStemmer("english"))
	recordingStore := &recordingTerminologyStore{
		ModTermStore: terminologyslice.NewSQLiteModTermStore(modDB, logger),
	}
	promptBuilder, err := terminologyslice.NewTermPromptBuilder("")
	if err != nil {
		_ = dictDB.Close()
		_ = modDB.Close()
		t.Fatalf("failed to create terminology prompt builder: %v", err)
	}
	terminology := terminologyslice.NewTermTranslator(repo, builder, searcher, recordingStore, promptBuilder, logger)

	cleanup := func() {
		_ = dictDB.Close()
		_ = modDB.Close()
	}
	return terminology, recordingStore, cleanup
}

func TestTranslationFlowServiceListTranslationFlowPersonaTargetsExcludesExistingMasterPersona(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-preview",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-a": {
					SpeakerID:    "00012345",
					EditorID:     "NPC_A",
					NPCName:      "Aela",
					Race:         "Nord",
					Sex:          "Female",
					VoiceType:    "FemaleNord",
					SourcePlugin: "",
					SourceHint:   "Skyrim.esm.extract.json",
				},
				"npc-b": {
					SpeakerID:    "00054321",
					EditorID:     "NPC_B",
					NPCName:      "Vilkas",
					Race:         "Nord",
					Sex:          "Male",
					VoiceType:    "MaleNord",
					SourcePlugin: "Update.esm",
				},
			},
			Dialogues: []translationflow.PersonaDialogueExcerpt{
				{
					ID:           "dlg-a-1",
					SpeakerID:    "00012345",
					EditorID:     "DIALOGUE_A",
					RecordType:   "DIAL:INFO",
					Text:         "The Companions stand ready.",
					SourcePlugin: "Skyrim.esm",
					Order:        1,
				},
				{
					ID:           "dlg-b-1",
					SpeakerID:    "00054321",
					EditorID:     "DIALOGUE_B",
					RecordType:   "DIAL:INFO",
					Text:         "The silver hand will fall.",
					SourcePlugin: "Update.esm",
					Order:        1,
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{
			personaLookupCompositeKey("Skyrim.esm", "00012345"): {
				SourcePlugin: "Skyrim.esm",
				SpeakerID:    "00012345",
				PersonaText:  "勇敢で実直な同胞団の戦士。",
			},
		},
	}
	service := &TranslationFlowService{store: store}

	page, err := service.ListTranslationFlowPersonaTargets(context.Background(), "task-persona-preview", 1, 50)
	if err != nil {
		t.Fatalf("ListTranslationFlowPersonaTargets failed: %v", err)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("unexpected persona row count: got=%d want=%d", len(page.Rows), 2)
	}

	var reusedRow PersonaTargetPreviewRow
	var pendingRow PersonaTargetPreviewRow
	for _, row := range page.Rows {
		switch row.SpeakerID {
		case "00012345":
			reusedRow = row
		case "00054321":
			pendingRow = row
		}
	}
	if reusedRow.ViewState != personaViewStateReused {
		t.Fatalf("unexpected reused row view state: got=%q want=%q", reusedRow.ViewState, personaViewStateReused)
	}
	if reusedRow.PersonaText == "" {
		t.Fatalf("reused row must include persona text")
	}
	if pendingRow.ViewState != personaViewStatePending {
		t.Fatalf("unexpected pending row view state: got=%q want=%q", pendingRow.ViewState, personaViewStatePending)
	}

	summary, err := service.GetTranslationFlowPersonaPhase(context.Background(), "task-persona-preview")
	if err != nil {
		t.Fatalf("GetTranslationFlowPersonaPhase failed: %v", err)
	}
	if summary.DetectedCount != 2 || summary.ReusedCount != 1 || summary.PendingCount != 1 {
		t.Fatalf("unexpected persona summary counts: %+v", summary)
	}
}

func TestTranslationFlowServiceRunTranslationFlowPersonaPhaseNoOpWhenAllReused(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-noop",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-a": {
					SpeakerID:    "00012345",
					SourcePlugin: "Skyrim.esm",
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{
			personaLookupCompositeKey("Skyrim.esm", "00012345"): {
				SourcePlugin: "Skyrim.esm",
				SpeakerID:    "00012345",
				PersonaText:  "既存ペルソナ",
			},
		},
	}
	personaWorkflow := &stubMasterPersona{}
	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}

	result, err := service.RunTranslationFlowPersonaPhase(context.Background(), RunTranslationFlowPersonaPhaseInput{
		TaskID: "task-persona-noop",
	})
	if err != nil {
		t.Fatalf("RunTranslationFlowPersonaPhase failed: %v", err)
	}
	if personaWorkflow.resumeCalls != 0 {
		t.Fatalf("persona workflow must not resume on no-op completion: resume_calls=%d", personaWorkflow.resumeCalls)
	}
	if result.PendingCount != 0 {
		t.Fatalf("pending count must be zero for no-op completion: got=%d", result.PendingCount)
	}
	if result.Status != "cached_only" {
		t.Fatalf("unexpected no-op status: got=%q want=%q", result.Status, "cached_only")
	}
}

func TestTranslationFlowServiceRunTranslationFlowPersonaPhaseBootstrapsWhenRuntimeRequestsDoNotExist(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-bootstrap",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-b": {
					SpeakerID:    "00054321",
					SourcePlugin: "Update.esm",
				},
			},
			Dialogues: []translationflow.PersonaDialogueExcerpt{
				{
					ID:           "dlg-b-1",
					SpeakerID:    "00054321",
					RecordType:   "DIAL:INFO",
					Text:         "We stand ready.",
					SourcePlugin: "Update.esm",
					Order:        1,
				},
			},
		},
		loadedFiles: []translationflow.LoadedFile{
			{
				ID:             1,
				TaskID:         "task-persona-bootstrap",
				SourceFilePath: "bootstrap-source.json",
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{},
	}
	personaWorkflow := &stubMasterPersona{}
	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}

	result, err := service.RunTranslationFlowPersonaPhase(context.Background(), RunTranslationFlowPersonaPhaseInput{
		TaskID: "task-persona-bootstrap",
		Request: TranslationRequestConfig{
			Provider: "openai",
			Model:    "gpt-4.1-mini",
		},
		Prompt: TranslationPromptConfig{
			UserPrompt:   "persona user prompt",
			SystemPrompt: "persona system prompt",
		},
	})
	if err != nil {
		t.Fatalf("RunTranslationFlowPersonaPhase failed: %v", err)
	}
	if personaWorkflow.startCalls != 0 {
		t.Fatalf("persona workflow must not start a new persona task: got=%d", personaWorkflow.startCalls)
	}
	if personaWorkflow.resumeCalls != 1 {
		t.Fatalf("translation-flow bootstrap run must execute through persona workflow contract: got=%d", personaWorkflow.resumeCalls)
	}
	if len(personaWorkflow.resumeTaskIDs) != 1 || personaWorkflow.resumeTaskIDs[0] != "task-persona-bootstrap" {
		t.Fatalf("persona workflow must run with translation task id: task_ids=%v", personaWorkflow.resumeTaskIDs)
	}
	if !personaWorkflow.resumeConfigProvided {
		t.Fatalf("run config must be passed to persona workflow contract")
	}
	if personaWorkflow.lastResumeRequest.Provider != "openai" || personaWorkflow.lastResumeRequest.Model != "gpt-4.1-mini" {
		t.Fatalf("unexpected run request config: %+v", personaWorkflow.lastResumeRequest)
	}
	if personaWorkflow.lastResumePrompt.UserPrompt != "persona user prompt" || personaWorkflow.lastResumePrompt.SystemPrompt != "persona system prompt" {
		t.Fatalf("unexpected run prompt config: %+v", personaWorkflow.lastResumePrompt)
	}
	if result.PendingCount != 1 || result.Status != "ready" {
		t.Fatalf("unexpected persona summary after bootstrap run: %+v", result)
	}
}

func TestTranslationFlowServiceRunTranslationFlowPersonaPhaseFailsWhenBootstrapSourcePathIsUnavailable(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-bootstrap-missing-source",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-c": {
					SpeakerID:    "00077777",
					SourcePlugin: "Skyrim.esm",
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{},
	}
	personaWorkflow := &stubMasterPersona{}
	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}

	_, err := service.RunTranslationFlowPersonaPhase(context.Background(), RunTranslationFlowPersonaPhaseInput{
		TaskID: "task-persona-bootstrap-missing-source",
		Request: TranslationRequestConfig{
			Model: "gpt-4.1-mini",
		},
	})
	if err == nil {
		t.Fatalf("RunTranslationFlowPersonaPhase unexpectedly succeeded without bootstrap source path")
	}
	if !strings.Contains(err.Error(), "source_json_path") {
		t.Fatalf("unexpected bootstrap error: %v", err)
	}
	if personaWorkflow.resumeCalls != 0 {
		t.Fatalf("persona workflow must not run when bootstrap source path is unavailable: resume_calls=%d", personaWorkflow.resumeCalls)
	}
}

func TestTranslationFlowServiceRunTranslationFlowPersonaPhaseResumesWhenRuntimeRequestsAlreadyExist(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-run",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-b": {
					SpeakerID:    "00054321",
					SourcePlugin: "Update.esm",
				},
			},
			Dialogues: []translationflow.PersonaDialogueExcerpt{
				{
					ID:           "dlg-b-1",
					SpeakerID:    "00054321",
					RecordType:   "DIAL:INFO",
					Text:         "We stand ready.",
					SourcePlugin: "Update.esm",
					Order:        1,
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{},
	}
	personaWorkflow := &stubMasterPersona{
		requestsMap: map[string][]runtimequeue.JobRequest{
			"task-persona-run": {
				buildPersonaJobRequest("task-persona-run", runtimequeue.RequestStatePending, "Update.esm", "00054321", nil),
			},
		},
	}

	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}
	result, err := service.RunTranslationFlowPersonaPhase(context.Background(), RunTranslationFlowPersonaPhaseInput{
		TaskID: "task-persona-run",
		Request: TranslationRequestConfig{
			Provider: "openai",
			Model:    "gpt-4.1-mini",
		},
		Prompt: TranslationPromptConfig{
			UserPrompt:   "persona user prompt",
			SystemPrompt: "persona system prompt",
		},
	})
	if err != nil {
		t.Fatalf("RunTranslationFlowPersonaPhase failed: %v", err)
	}
	if personaWorkflow.startCalls != 0 {
		t.Fatalf("persona workflow must not start a new persona task: got=%d", personaWorkflow.startCalls)
	}
	if personaWorkflow.resumeCalls != 1 {
		t.Fatalf("translation-flow run must resume same translation task when runtime exists: got=%d", personaWorkflow.resumeCalls)
	}
	if len(personaWorkflow.resumeTaskIDs) != 1 || personaWorkflow.resumeTaskIDs[0] != "task-persona-run" {
		t.Fatalf("persona workflow must resume translation task id: task_ids=%v", personaWorkflow.resumeTaskIDs)
	}
	if !personaWorkflow.resumeConfigProvided {
		t.Fatalf("run config must be passed to persona workflow resume contract")
	}
	if personaWorkflow.lastResumeRequest.Provider != "openai" || personaWorkflow.lastResumeRequest.Model != "gpt-4.1-mini" {
		t.Fatalf("unexpected resume request config: %+v", personaWorkflow.lastResumeRequest)
	}
	if personaWorkflow.lastResumePrompt.UserPrompt != "persona user prompt" || personaWorkflow.lastResumePrompt.SystemPrompt != "persona system prompt" {
		t.Fatalf("unexpected resume prompt config: %+v", personaWorkflow.lastResumePrompt)
	}
	if result.PendingCount != 1 || result.Status != "ready" {
		t.Fatalf("unexpected persona summary after fresh run: %+v", result)
	}
}

func TestTranslationFlowServiceGetTranslationFlowPersonaPhaseRestoresPartialFailure(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-partial",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-a": {
					SpeakerID:    "00012345",
					SourcePlugin: "Skyrim.esm",
				},
				"npc-b": {
					SpeakerID:    "00054321",
					SourcePlugin: "Update.esm",
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{
			personaLookupCompositeKey("Skyrim.esm", "00012345"): {
				SourcePlugin: "Skyrim.esm",
				SpeakerID:    "00012345",
				PersonaText:  "保存済みペルソナ",
			},
		},
	}
	failureMessage := "provider timeout"
	personaWorkflow := &stubMasterPersona{
		requestsMap: map[string][]runtimequeue.JobRequest{
			"task-persona-partial": {
				buildPersonaJobRequest("task-persona-partial", runtimequeue.RequestStateCompleted, "Skyrim.esm", "00012345", nil),
				buildPersonaJobRequest("task-persona-partial", runtimequeue.RequestStateFailed, "Update.esm", "00054321", &failureMessage),
			},
		},
	}
	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}

	summary, err := service.GetTranslationFlowPersonaPhase(context.Background(), "task-persona-partial")
	if err != nil {
		t.Fatalf("GetTranslationFlowPersonaPhase failed: %v", err)
	}
	if summary.Status != "partial_failed" {
		t.Fatalf("unexpected persona summary status: got=%q want=%q", summary.Status, "partial_failed")
	}
	if summary.GeneratedCount != 1 || summary.FailedCount != 1 || summary.PendingCount != 0 {
		t.Fatalf("unexpected persona summary counts: %+v", summary)
	}

	page, err := service.ListTranslationFlowPersonaTargets(context.Background(), "task-persona-partial", 1, 50)
	if err != nil {
		t.Fatalf("ListTranslationFlowPersonaTargets failed: %v", err)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("unexpected persona row count: got=%d want=%d", len(page.Rows), 2)
	}
	var failedRow PersonaTargetPreviewRow
	for _, row := range page.Rows {
		if row.SpeakerID == "00054321" {
			failedRow = row
		}
	}
	if failedRow.ViewState != personaViewStateFailed {
		t.Fatalf("unexpected failed row view state: got=%q want=%q", failedRow.ViewState, personaViewStateFailed)
	}
	if failedRow.ErrorMessage != failureMessage {
		t.Fatalf("unexpected failed row error message: got=%q want=%q", failedRow.ErrorMessage, failureMessage)
	}

	if _, err := service.RunTranslationFlowPersonaPhase(context.Background(), RunTranslationFlowPersonaPhaseInput{
		TaskID: "task-persona-partial",
		Request: TranslationRequestConfig{
			Model: "gemini-2.5-flash",
		},
	}); err != nil {
		t.Fatalf("RunTranslationFlowPersonaPhase retry failed: %v", err)
	}
	if personaWorkflow.resumeCalls != 1 {
		t.Fatalf("persona workflow must resume failed/unresolved targets: resume_calls=%d", personaWorkflow.resumeCalls)
	}
	if len(personaWorkflow.resumeTaskIDs) != 1 || personaWorkflow.resumeTaskIDs[0] != "task-persona-partial" {
		t.Fatalf("persona workflow must resume translation task: task_ids=%v", personaWorkflow.resumeTaskIDs)
	}
}

func TestTranslationFlowServiceGetTranslationFlowPersonaPhaseRestoresStateAfterServiceRecreation(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-recreate",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-a": {
					SpeakerID:    "00011111",
					SourcePlugin: "Skyrim.esm",
				},
				"npc-b": {
					SpeakerID:    "00022222",
					SourcePlugin: "Update.esm",
				},
			},
		},
		finalPersonasByLookup: map[string]translationflow.PersonaFinalSummary{
			personaLookupCompositeKey("Skyrim.esm", "00011111"): {
				SourcePlugin: "Skyrim.esm",
				SpeakerID:    "00011111",
				PersonaText:  "保存済みペルソナ",
			},
		},
	}
	personaWorkflow := &stubMasterPersona{
		requestsMap: map[string][]runtimequeue.JobRequest{
			"task-persona-recreate": {
				buildPersonaJobRequest("task-persona-recreate", runtimequeue.RequestStateCompleted, "Skyrim.esm", "00011111", nil),
				buildPersonaJobRequest("task-persona-recreate", runtimequeue.RequestStateRunning, "Update.esm", "00022222", nil),
			},
		},
	}

	first := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}
	firstSummary, err := first.GetTranslationFlowPersonaPhase(context.Background(), "task-persona-recreate")
	if err != nil {
		t.Fatalf("GetTranslationFlowPersonaPhase first instance failed: %v", err)
	}
	if firstSummary.Status != "running" || firstSummary.GeneratedCount != 1 || firstSummary.FailedCount != 0 {
		t.Fatalf("unexpected first summary: %+v", firstSummary)
	}

	recreated := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}
	summary, err := recreated.GetTranslationFlowPersonaPhase(context.Background(), "task-persona-recreate")
	if err != nil {
		t.Fatalf("GetTranslationFlowPersonaPhase recreated instance failed: %v", err)
	}
	if summary.Status != "running" || summary.GeneratedCount != 1 || summary.FailedCount != 0 {
		t.Fatalf("unexpected recreated summary: %+v", summary)
	}

	page, err := recreated.ListTranslationFlowPersonaTargets(context.Background(), "task-persona-recreate", 1, 50)
	if err != nil {
		t.Fatalf("ListTranslationFlowPersonaTargets recreated instance failed: %v", err)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("unexpected row count after recreation: got=%d want=%d", len(page.Rows), 2)
	}
	var generatedRow PersonaTargetPreviewRow
	var runningRow PersonaTargetPreviewRow
	for _, row := range page.Rows {
		switch row.SpeakerID {
		case "00011111":
			generatedRow = row
		case "00022222":
			runningRow = row
		}
	}
	if generatedRow.ViewState != personaViewStateGenerated {
		t.Fatalf("unexpected generated row state after recreation: got=%q want=%q", generatedRow.ViewState, personaViewStateGenerated)
	}
	if runningRow.ViewState != personaViewStateRunning {
		t.Fatalf("unexpected running row state after recreation: got=%q want=%q", runningRow.ViewState, personaViewStateRunning)
	}
}

func TestTranslationFlowServiceListTranslationFlowPersonaTargetsHandlesInitialPreviewWithoutRuntimeQueue(t *testing.T) {
	store := &stubTranslationFlowStore{
		personaInput: translationflow.PersonaCandidateInput{
			TaskID: "task-persona-preview-initial",
			Candidates: map[string]translationflow.PersonaCandidate{
				"npc-c": {
					SpeakerID:    "00022222",
					EditorID:     "NPC_C",
					SourcePlugin: "Skyrim.esm",
				},
			},
		},
	}
	personaWorkflow := &stubMasterPersona{
		requestStateErr: errors.New("GetTaskRequestState must not be called"),
	}
	service := &TranslationFlowService{
		store:           store,
		personaWorkflow: personaWorkflow,
	}

	page, err := service.ListTranslationFlowPersonaTargets(context.Background(), "task-persona-preview-initial", 1, 50)
	if err != nil {
		t.Fatalf("ListTranslationFlowPersonaTargets failed: %v", err)
	}
	if len(page.Rows) != 1 {
		t.Fatalf("unexpected persona row count: got=%d want=%d", len(page.Rows), 1)
	}
	if page.Rows[0].ViewState != personaViewStatePending {
		t.Fatalf("initial preview must stay pending without runtime queue: got=%q want=%q", page.Rows[0].ViewState, personaViewStatePending)
	}
	if personaWorkflow.requestStateCalls != 0 {
		t.Fatalf("initial preview must not read task request state: calls=%d", personaWorkflow.requestStateCalls)
	}

	summary, err := service.GetTranslationFlowPersonaPhase(context.Background(), "task-persona-preview-initial")
	if err != nil {
		t.Fatalf("GetTranslationFlowPersonaPhase failed: %v", err)
	}
	if summary.Status != "ready" || summary.PendingCount != 1 {
		t.Fatalf("unexpected persona summary for initial preview: %+v", summary)
	}
}

type stubTranslationFlowStore struct {
	translationflow.Service
	personaInput          translationflow.PersonaCandidateInput
	finalPersonasByLookup map[string]translationflow.PersonaFinalSummary
	loadedFiles           []translationflow.LoadedFile
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
	return append([]translationflow.LoadedFile(nil), s.loadedFiles...), nil
}

func (s *stubTranslationFlowStore) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (translationflow.PreviewPage, error) {
	_ = ctx
	_ = fileID
	_ = page
	_ = pageSize
	return translationflow.PreviewPage{}, nil
}

func (s *stubTranslationFlowStore) LoadPersonaCandidates(ctx context.Context, taskID string) (translationflow.PersonaCandidateInput, error) {
	_ = ctx
	_ = taskID

	candidates := make(map[string]translationflow.PersonaCandidate, len(s.personaInput.Candidates))
	for key, npc := range s.personaInput.Candidates {
		candidates[key] = translationflow.PersonaCandidate{
			SpeakerID:      npc.SpeakerID,
			SourceRecordID: npc.SourceRecordID,
			NPCKey:         npc.NPCKey,
			EditorID:       npc.EditorID,
			RecordType:     npc.RecordType,
			NPCName:        npc.NPCName,
			Race:           npc.Race,
			Sex:            npc.Sex,
			VoiceType:      npc.VoiceType,
			SourcePlugin:   npc.SourcePlugin,
			SourceHint:     npc.SourceHint,
		}
	}
	dialogues := make([]translationflow.PersonaDialogueExcerpt, 0, len(s.personaInput.Dialogues))
	for _, dialogue := range s.personaInput.Dialogues {
		dialogues = append(dialogues, translationflow.PersonaDialogueExcerpt{
			ID:               dialogue.ID,
			SpeakerID:        dialogue.SpeakerID,
			EditorID:         dialogue.EditorID,
			GroupEditorID:    dialogue.GroupEditorID,
			RecordType:       dialogue.RecordType,
			Text:             dialogue.Text,
			QuestID:          dialogue.QuestID,
			SourcePlugin:     dialogue.SourcePlugin,
			SourceHint:       dialogue.SourceHint,
			IsServicesBranch: dialogue.IsServicesBranch,
			Order:            dialogue.Order,
		})
	}

	return translationflow.PersonaCandidateInput{
		TaskID:     s.personaInput.TaskID,
		Candidates: candidates,
		Dialogues:  dialogues,
	}, nil
}

func (s *stubTranslationFlowStore) FindPersonaFinal(ctx context.Context, key translationflow.PersonaLookupKey) (translationflow.PersonaFinalSummary, bool, error) {
	_ = ctx
	if s.finalPersonasByLookup == nil {
		return translationflow.PersonaFinalSummary{}, false, nil
	}
	lookupKey := personaLookupCompositeKey(key.SourcePlugin, key.SpeakerID)
	persona, ok := s.finalPersonasByLookup[lookupKey]
	if !ok {
		return translationflow.PersonaFinalSummary{}, false, nil
	}
	return translationflow.PersonaFinalSummary{
		PersonaID:    persona.PersonaID,
		SourcePlugin: persona.SourcePlugin,
		SpeakerID:    persona.SpeakerID,
		PersonaText:  persona.PersonaText,
	}, true, nil
}

type stubTerminology struct {
	preparePromptsResult []llmio.Request
	listTargetsResult    []terminologyslice.TerminologyEntry
	previewTranslations  map[string]terminologyslice.PreviewTranslation
	summary              terminologyslice.PhaseSummary
	updatedSummary       terminologyslice.PhaseSummary
	updatedSummaries     []terminologyslice.PhaseSummary
	savedResponses       []llmio.Response
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
	s.savedResponses = append([]llmio.Response(nil), responses...)
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

func (s *stubTerminology) ListTargets(ctx context.Context, taskID string) ([]terminologyslice.TerminologyEntry, error) {
	_ = ctx
	_ = taskID
	return append([]terminologyslice.TerminologyEntry(nil), s.listTargetsResult...), nil
}

func (s *stubTerminology) UpdatePhaseSummary(ctx context.Context, summary terminologyslice.PhaseSummary) error {
	_ = ctx
	s.updatedSummary = summary
	s.updatedSummaries = append(s.updatedSummaries, summary)
	s.summary = summary
	return nil
}

type workflowTerminologyInputRepository struct {
	inputs map[string]translationinput.TerminologyInput
}

func (r *workflowTerminologyInputRepository) LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error) {
	_ = ctx
	input, ok := r.inputs[taskID]
	if !ok {
		return translationinput.TerminologyInput{}, fmt.Errorf("terminology input is not configured task_id=%s", taskID)
	}
	return input, nil
}

type recordingTerminologyStore struct {
	terminologyslice.ModTermStore
	updatedSummaries []terminologyslice.PhaseSummary
}

func (s *recordingTerminologyStore) UpdatePhaseSummary(ctx context.Context, summary terminologyslice.PhaseSummary) error {
	s.updatedSummaries = append(s.updatedSummaries, summary)
	return s.ModTermStore.UpdatePhaseSummary(ctx, summary)
}

type scriptedTerminologyResponse struct {
	success bool
	content string
	err     string
}

type scriptedTerminologyExecutor struct {
	responses []scriptedTerminologyResponse
	steps     []int
}

func (s *scriptedTerminologyExecutor) Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error) {
	_ = ctx
	_ = config
	return s.buildResponses(requests), nil
}

func (s *scriptedTerminologyExecutor) ExecuteWithProgress(
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
	return s.buildResponses(requests), nil
}

func (s *scriptedTerminologyExecutor) buildResponses(requests []llmio.Request) []llmio.Response {
	responses := make([]llmio.Response, 0, len(s.responses))
	for index, response := range s.responses {
		metadata := map[string]interface{}{}
		if index < len(requests) {
			metadata = requests[index].Metadata
		}
		responses = append(responses, llmio.Response{
			Content:  response.content,
			Success:  response.success,
			Error:    response.err,
			Metadata: metadata,
		})
	}
	return responses
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

type stubMasterPersona struct {
	startTaskID          string
	startErr             error
	resumeErr            error
	cancelErr            error
	startCalls           int
	startInputs          []StartMasterPersonaInput
	resumeTaskIDs        []string
	requestState         runtimequeue.TaskRequestState
	requestStateErr      error
	requestStateMap      map[string]runtimequeue.TaskRequestState
	requestStateCalls    int
	requests             []runtimequeue.JobRequest
	requestsErr          error
	requestsMap          map[string][]runtimequeue.JobRequest
	resumeCalls          int
	onResume             func(taskID string)
	lastResumeRequest    TranslationRequestConfig
	lastResumePrompt     TranslationPromptConfig
	resumeConfigProvided bool
}

func (s *stubMasterPersona) StartMasterPersona(ctx context.Context, input StartMasterPersonaInput) (string, error) {
	_ = ctx
	s.startCalls++
	s.startInputs = append(s.startInputs, input)
	if s.startErr != nil {
		return "", s.startErr
	}
	if s.startTaskID == "" {
		return "persona-task", nil
	}
	return s.startTaskID, nil
}

func (s *stubMasterPersona) ResumeMasterPersona(ctx context.Context, taskID string) error {
	request, prompt, ok := personaPhaseRunConfigFromContext(ctx)
	if ok {
		s.resumeConfigProvided = true
		s.lastResumeRequest = request
		s.lastResumePrompt = prompt
	}
	s.resumeCalls++
	s.resumeTaskIDs = append(s.resumeTaskIDs, taskID)
	if s.onResume != nil {
		s.onResume(taskID)
	}
	if s.resumeErr != nil {
		return s.resumeErr
	}
	return nil
}

func (s *stubMasterPersona) CancelMasterPersona(ctx context.Context, taskID string) error {
	_ = ctx
	_ = taskID
	return s.cancelErr
}

func (s *stubMasterPersona) GetTaskRequestState(ctx context.Context, taskID string) (runtimequeue.TaskRequestState, error) {
	_ = ctx
	s.requestStateCalls++
	if s.requestStateMap != nil {
		state, ok := s.requestStateMap[taskID]
		if !ok {
			return runtimequeue.TaskRequestState{}, nil
		}
		return state, nil
	}
	if s.requestStateErr != nil {
		return runtimequeue.TaskRequestState{}, s.requestStateErr
	}
	return s.requestState, nil
}

func (s *stubMasterPersona) GetTaskRequests(ctx context.Context, taskID string) ([]runtimequeue.JobRequest, error) {
	_ = ctx
	if s.requestsMap != nil {
		requests, ok := s.requestsMap[taskID]
		if !ok {
			return nil, nil
		}
		return append([]runtimequeue.JobRequest(nil), requests...), nil
	}
	if s.requestsErr != nil {
		return nil, s.requestsErr
	}
	return append([]runtimequeue.JobRequest(nil), s.requests...), nil
}

func buildPersonaJobRequest(taskID string, requestState string, sourcePlugin string, speakerID string, errorMessage *string) runtimequeue.JobRequest {
	requestPayload := struct {
		Metadata map[string]any `json:"metadata"`
	}{
		Metadata: map[string]any{
			"source_plugin": sourcePlugin,
			"speaker_id":    speakerID,
		},
	}
	rawRequest, err := json.Marshal(requestPayload)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal persona request payload: %v", err))
	}
	return runtimequeue.JobRequest{
		TaskID:       taskID,
		RequestState: requestState,
		RequestJSON:  string(rawRequest),
		ErrorMessage: errorMessage,
	}
}
