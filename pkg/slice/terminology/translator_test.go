package terminology

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	dictionaryartifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/dictionary_artifact"
	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *sql.DB, func()) {
	t.Helper()

	dictDB, err := sql.Open("sqlite", "file:dict?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open dict db: %v", err)
	}

	_, err = dictDB.Exec(`
		CREATE TABLE artifact_dictionary_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER DEFAULT 0,
			edid TEXT DEFAULT '',
			source_text TEXT,
			dest_text TEXT,
			record_type TEXT
		);
		INSERT INTO artifact_dictionary_entries (source_text, dest_text, record_type) VALUES
			('Iron Sword', '鉄の剣', 'BOOK:FULL'),
			('Uthgerd the Unbroken', '不屈のウスガルド', 'NPC_:FULL'),
			('Skeever', 'スキーヴァー', 'BOOK:FULL'),
			('Broken Tower Redoubt', 'ブロークン・タワー砦', 'BOOK:FULL'),
			('Tower', '塔', 'BOOK:FULL');
	`)
	if err != nil {
		t.Fatalf("failed to init dict db: %v", err)
	}

	modDB, err := sql.Open("sqlite", "file:mod?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open mod db: %v", err)
	}

	return dictDB, modDB, func() {
		_ = dictDB.Close()
		_ = modDB.Close()
	}
}

func TestTermTranslatorSlice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name          string
		input         TerminologyInput
		mockLLMOutput []string
		expectedTerms map[string]string
		expectedReqs  int
		expectedTotal int
	}{
		{
			name: "allowed recs and npc pair are translated",
			input: TerminologyInput{
				TaskID: "task-1",
				Entries: []TerminologyEntry{
					{
						ID:         "001",
						EditorID:   "EditorA",
						RecordType: "BOOK:FULL",
						SourceText: "Iron Sword",
						SourceFile: "test_mod.json",
						Variant:    "single",
					},
					{
						ID:         "002",
						EditorID:   "EditorA",
						RecordType: "ARMO:FULL",
						SourceText: "Steel Armor",
						SourceFile: "test_mod.json",
						Variant:    "single",
					},
					{
						ID:         "101",
						EditorID:   "EditorID1",
						RecordType: "NPC_:FULL",
						SourceText: "Uthgerd the Unbroken",
						SourceFile: "test_mod.json",
						PairKey:    "EditorID1",
						Variant:    "full",
					},
					{
						ID:         "102",
						EditorID:   "EditorID1",
						RecordType: "NPC_:SHRT",
						SourceText: "Uthgerd",
						SourceFile: "test_mod.json",
						PairKey:    "EditorID1",
						Variant:    "short",
					},
					{
						ID:         "900",
						EditorID:   "EditorA",
						RecordType: "SPEL:FULL",
						SourceText: "Fireball",
						SourceFile: "test_mod.json",
						Variant:    "single",
					},
				},
			},
			mockLLMOutput: []string{
				"TL: |鋼鉄の鎧|",
			},
			expectedReqs:  1,
			expectedTotal: 3,
			expectedTerms: map[string]string{
				"Iron Sword":           "鉄の剣",
				"Steel Armor":          "鋼鉄の鎧",
				"Uthgerd the Unbroken": "不屈のウスガルド",
				"Uthgerd":              "不屈のウスガルド",
			},
		},
		{
			name: "duplicate record text is collapsed into one request",
			input: TerminologyInput{
				TaskID: "task-2",
				Entries: []TerminologyEntry{
					{
						ID:         "201",
						EditorID:   "EditorB",
						RecordType: "BOOK:FULL",
						SourceText: "Iron Sword",
						SourceFile: "mod_a.json",
						Variant:    "single",
					},
					{
						ID:         "202",
						EditorID:   "EditorC",
						RecordType: "BOOK:FULL",
						SourceText: "Iron Sword",
						SourceFile: "mod_b.json",
						Variant:    "single",
					},
					{
						ID:         "203",
						EditorID:   "EditorB",
						RecordType: "ARMO:FULL",
						SourceText: "Steel Armor",
						SourceFile: "mod_a.json",
						Variant:    "single",
					},
				},
			},
			mockLLMOutput: []string{
				"TL: |鋼鉄の鎧|",
			},
			expectedReqs:  1,
			expectedTotal: 2,
			expectedTerms: map[string]string{
				"Iron Sword":  "鉄の剣",
				"Steel Armor": "鋼鉄の鎧",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dictDB, modDB, cleanup := setupTestDB(t)
			defer cleanup()

			repo := &fakeTranslationInputRepository{input: tc.input}
			builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
			stemmer := NewSnowballStemmer("english")
			dictRepo := dictionaryartifact.NewRepository(dictDB)
			searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
			store := NewSQLiteModTermStore(modDB, logger)
			promptBuilder, err := NewTermPromptBuilder("")
			if err != nil {
				t.Fatalf("failed to create prompt builder: %v", err)
			}

			translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

			llmRequests, err := translator.PreparePrompts(ctx, tc.input.TaskID, PhaseOptions{})
			if err != nil {
				t.Fatalf("PreparePrompts failed: %v", err)
			}
			if len(llmRequests) != tc.expectedReqs {
				t.Fatalf("PreparePrompts generated %d requests, want %d", len(llmRequests), tc.expectedReqs)
			}

			llmResponses := make([]llmio.Response, 0, len(llmRequests))
			for i, content := range tc.mockLLMOutput {
				llmResponses = append(llmResponses, llmio.Response{
					Content:  content,
					Success:  true,
					Metadata: llmRequests[i].Metadata,
				})
			}

			if err := translator.SaveResults(ctx, tc.input.TaskID, llmResponses); err != nil {
				t.Fatalf("SaveResults failed: %v", err)
			}

			for originalEN, expectedJA := range tc.expectedTerms {
				translatedJA, err := store.GetTerm(ctx, originalEN)
				if err != nil {
					t.Fatalf("GetTerm(%s) failed: %v", originalEN, err)
				}
				if translatedJA != expectedJA {
					t.Fatalf("unexpected translation for %s: got=%q want=%q", originalEN, translatedJA, expectedJA)
				}
			}

			phaseSummary, err := store.GetPhaseSummary(ctx, tc.input.TaskID)
			if err != nil {
				t.Fatalf("GetPhaseSummary failed: %v", err)
			}
			if phaseSummary.TargetCount != tc.expectedTotal {
				t.Fatalf("unexpected target count: got=%d want=%d", phaseSummary.TargetCount, tc.expectedTotal)
			}
			if phaseSummary.Status != "completed" {
				t.Fatalf("unexpected status: got=%q want=%q", phaseSummary.Status, "completed")
			}
			if phaseSummary.ProgressMode != "hidden" {
				t.Fatalf("unexpected progress mode: got=%q want=%q", phaseSummary.ProgressMode, "hidden")
			}

			previewTranslations, err := translator.GetPreviewTranslations(ctx, tc.input.Entries)
			if err != nil {
				t.Fatalf("GetPreviewTranslations failed: %v", err)
			}
			for _, entry := range tc.input.Entries {
				preview, ok := previewTranslations[entry.ID]
				if !ok {
					t.Fatalf("missing preview translation for row_id=%s", entry.ID)
				}
				if expectedJA, exists := tc.expectedTerms[entry.SourceText]; exists {
					if preview.TranslationState != "translated" {
						t.Fatalf("unexpected translation state for row_id=%s: got=%q want=%q", entry.ID, preview.TranslationState, "translated")
					}
					if preview.TranslatedText != expectedJA {
						t.Fatalf("unexpected preview translation for row_id=%s: got=%q want=%q", entry.ID, preview.TranslatedText, expectedJA)
					}
					continue
				}
				if preview.TranslationState != "missing" {
					t.Fatalf("unexpected missing translation state for row_id=%s: got=%q want=%q", entry.ID, preview.TranslationState, "missing")
				}
			}
		})
	}
}

func TestTermTranslator_PreparePrompts_AppliesPartialReplacement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, modDB, cleanup := setupTestDB(t)
	defer cleanup()

	input := TerminologyInput{
		TaskID: "task-partial",
		Entries: []TerminologyEntry{
			{
				ID:         "301",
				EditorID:   "EditorD",
				RecordType: "BOOK:FULL",
				SourceText: "Skeever Den",
				SourceFile: "mod_partial.json",
				Variant:    "single",
			},
		},
	}
	repo := &fakeTranslationInputRepository{input: input}
	builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
	stemmer := NewSnowballStemmer("english")
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
	store := NewSQLiteModTermStore(modDB, logger)
	promptBuilder, err := NewTermPromptBuilder("")
	if err != nil {
		t.Fatalf("failed to create prompt builder: %v", err)
	}
	translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

	requests, err := translator.PreparePrompts(ctx, "task-partial", PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected request count: got=%d want=%d", len(requests), 1)
	}
	sourceText, _ := requests[0].Metadata["source_text"].(string)
	replacedText, _ := requests[0].Metadata["replaced_source_text"].(string)
	if sourceText != "Skeever Den" {
		t.Fatalf("unexpected source text metadata: got=%q want=%q", sourceText, "Skeever Den")
	}
	if replacedText != "スキーヴァー Den" {
		t.Fatalf("unexpected replaced source text metadata: got=%q want=%q", replacedText, "スキーヴァー Den")
	}
	if requests[0].SystemPrompt == "" || !contains(requests[0].SystemPrompt, "スキーヴァー Den") {
		t.Fatalf("system prompt must include replaced source text")
	}
}

func TestTermTranslator_PreparePrompts_PrefersLongestPhraseInOverlap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, modDB, cleanup := setupTestDB(t)
	defer cleanup()

	input := TerminologyInput{
		TaskID: "task-longest-first",
		Entries: []TerminologyEntry{
			{
				ID:         "401",
				EditorID:   "EditorE",
				RecordType: "BOOK:FULL",
				SourceText: "Visit Broken Tower Redoubt",
				SourceFile: "mod_overlap.json",
				Variant:    "single",
			},
		},
	}
	repo := &fakeTranslationInputRepository{input: input}
	builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
	stemmer := NewSnowballStemmer("english")
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
	store := NewSQLiteModTermStore(modDB, logger)
	promptBuilder, err := NewTermPromptBuilder("")
	if err != nil {
		t.Fatalf("failed to create prompt builder: %v", err)
	}
	translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

	requests, err := translator.PreparePrompts(ctx, "task-longest-first", PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected request count: got=%d want=%d", len(requests), 1)
	}
	replacedText, _ := requests[0].Metadata["replaced_source_text"].(string)
	if replacedText != "Visit ブロークン・タワー砦" {
		t.Fatalf("unexpected longest-first replacement: got=%q want=%q", replacedText, "Visit ブロークン・タワー砦")
	}
	if strings.Contains(replacedText, "塔") && replacedText != "Visit ブロークン・タワー砦" {
		t.Fatalf("short token replacement must not override long phrase: replaced=%q", replacedText)
	}
}

func TestTermTranslator_PreparePrompts_ReappliesSameReplacementOnRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, modDB, cleanup := setupTestDB(t)
	defer cleanup()

	input := TerminologyInput{
		TaskID: "task-retry-reapply",
		Entries: []TerminologyEntry{
			{
				ID:         "501",
				EditorID:   "EditorF",
				RecordType: "BOOK:FULL",
				SourceText: "Skeever Den",
				SourceFile: "mod_retry.json",
				Variant:    "single",
			},
		},
	}
	repo := &fakeTranslationInputRepository{input: input}
	builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
	stemmer := NewSnowballStemmer("english")
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
	store := NewSQLiteModTermStore(modDB, logger)
	promptBuilder, err := NewTermPromptBuilder("")
	if err != nil {
		t.Fatalf("failed to create prompt builder: %v", err)
	}
	translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

	first, err := translator.PreparePrompts(ctx, "task-retry-reapply", PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts first failed: %v", err)
	}
	second, err := translator.PreparePrompts(ctx, "task-retry-reapply", PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts retry failed: %v", err)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("unexpected request counts: first=%d second=%d", len(first), len(second))
	}
	firstReplaced, _ := first[0].Metadata["replaced_source_text"].(string)
	secondReplaced, _ := second[0].Metadata["replaced_source_text"].(string)
	if firstReplaced != "スキーヴァー Den" {
		t.Fatalf("unexpected first replacement: got=%q want=%q", firstReplaced, "スキーヴァー Den")
	}
	if secondReplaced != firstReplaced {
		t.Fatalf("retry must apply identical replacement rule: first=%q second=%q", firstReplaced, secondReplaced)
	}
}

func TestTermDictionarySearcher_SearchExactKeywords_NoLikeLimitMiss(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, _, cleanup := setupTestDB(t)
	defer cleanup()

	for i := 0; i < 260; i++ {
		_, err := dictDB.Exec(`INSERT INTO artifact_dictionary_entries (source_text, dest_text, record_type) VALUES (?, ?, 'BOOK:FULL')`,
			"NeedleExact Variant "+strconv.Itoa(i), "ノイズ "+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("failed to insert noise row %d: %v", i, err)
		}
	}
	_, err := dictDB.Exec(`INSERT INTO artifact_dictionary_entries (source_text, dest_text, record_type) VALUES ('needleexact', '正解', 'BOOK:FULL')`)
	if err != nil {
		t.Fatalf("failed to insert exact row: %v", err)
	}

	dictRepo := dictionaryartifact.NewRepository(dictDB)
	stemmer := NewSnowballStemmer("english")
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)

	terms, err := searcher.SearchExactKeywords(ctx, []string{"NeedleExact"})
	if err != nil {
		t.Fatalf("SearchExactKeywords failed: %v", err)
	}
	found := false
	for _, term := range terms {
		if strings.EqualFold(term.Source, "NeedleExact") && term.Translation == "正解" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected exact keyword result not found; terms=%v", terms)
	}
}

func TestTermDictionarySearcher_SearchBatch_KeepsInputKeysWithWhitespace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, _, cleanup := setupTestDB(t)
	defer cleanup()

	dictRepo := dictionaryartifact.NewRepository(dictDB)
	stemmer := NewSnowballStemmer("english")
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)

	result, err := searcher.SearchBatch(ctx, []string{" Iron Sword "})
	if err != nil {
		t.Fatalf("SearchBatch failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("unexpected map size: got=%d want=1", len(result))
	}
	if _, ok := result[" Iron Sword "]; !ok {
		t.Fatalf("input key with whitespace must be preserved")
	}
	if terms := result[" Iron Sword "]; len(terms) != 0 {
		t.Fatalf("whitespace key must not match strict exact search: terms=%v", terms)
	}
}

func TestTermDictionarySearcher_SearchBatch_DoesNotCreateUnrequestedKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, _, cleanup := setupTestDB(t)
	defer cleanup()

	dictRepo := dictionaryartifact.NewRepository(dictDB)
	stemmer := NewSnowballStemmer("english")
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)

	result, err := searcher.SearchBatch(ctx, []string{"Iron Sword", " Iron Sword "})
	if err != nil {
		t.Fatalf("SearchBatch failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("unexpected map size: got=%d want=2", len(result))
	}
	if _, ok := result["Iron Sword"]; !ok {
		t.Fatalf("missing requested key: Iron Sword")
	}
	if _, ok := result[" Iron Sword "]; !ok {
		t.Fatalf("missing requested key: space-padded")
	}
	for key := range result {
		if key != "Iron Sword" && key != " Iron Sword " {
			t.Fatalf("unexpected unrequested key generated: %q", key)
		}
	}
}

func TestTermTranslator_SaveResults_DoesNotCountCachedAsFailed_WhenRunningSummaryResetsSavedCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, modDB, cleanup := setupTestDB(t)
	defer cleanup()

	input := TerminologyInput{
		TaskID: "task-cached-summary-reset-success",
		Entries: []TerminologyEntry{
			{
				ID:         "601",
				EditorID:   "EditorG",
				RecordType: "BOOK:FULL",
				SourceText: "Iron Sword",
				SourceFile: "mod_cached_reset.json",
				Variant:    "single",
			},
			{
				ID:         "602",
				EditorID:   "EditorG",
				RecordType: "ARMO:FULL",
				SourceText: "Steel Armor",
				SourceFile: "mod_cached_reset.json",
				Variant:    "single",
			},
		},
	}
	repo := &fakeTranslationInputRepository{input: input}
	builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
	stemmer := NewSnowballStemmer("english")
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
	store := NewSQLiteModTermStore(modDB, logger)
	promptBuilder, err := NewTermPromptBuilder("")
	if err != nil {
		t.Fatalf("failed to create prompt builder: %v", err)
	}
	translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

	requests, err := translator.PreparePrompts(ctx, input.TaskID, PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected request count: got=%d want=1", len(requests))
	}

	preparedSummary, err := store.GetPhaseSummary(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetPhaseSummary before reset failed: %v", err)
	}
	if preparedSummary.TargetCount != 2 || preparedSummary.SavedCount != 1 || preparedSummary.FailedCount != 0 {
		t.Fatalf("unexpected prepared summary: %+v", preparedSummary)
	}

	if err := translator.UpdatePhaseSummary(ctx, PhaseSummary{
		TaskID:          input.TaskID,
		Status:          "running",
		TargetCount:     preparedSummary.TargetCount,
		ProgressMode:    "determinate",
		ProgressCurrent: preparedSummary.SavedCount,
		ProgressTotal:   preparedSummary.TargetCount,
		ProgressMessage: "running",
	}); err != nil {
		t.Fatalf("UpdatePhaseSummary reset failed: %v", err)
	}
	resetSummary, err := store.GetPhaseSummary(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetPhaseSummary after reset failed: %v", err)
	}
	if resetSummary.SavedCount != 0 {
		t.Fatalf("saved_count reset simulation failed: got=%d want=0", resetSummary.SavedCount)
	}

	responses := []llmio.Response{
		{
			Content:  "TL: |鋼鉄の鎧|",
			Success:  true,
			Metadata: requests[0].Metadata,
		},
	}
	if err := translator.SaveResults(ctx, input.TaskID, responses); err != nil {
		t.Fatalf("SaveResults failed: %v", err)
	}

	summary, err := store.GetPhaseSummary(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetPhaseSummary after save failed: %v", err)
	}
	if summary.Status != "completed" {
		t.Fatalf("unexpected status: got=%q want=%q", summary.Status, "completed")
	}
	if summary.SavedCount != 2 {
		t.Fatalf("unexpected saved count: got=%d want=2", summary.SavedCount)
	}
	if summary.FailedCount != 0 {
		t.Fatalf("unexpected failed count: got=%d want=0", summary.FailedCount)
	}
}

func TestTermTranslator_SaveResults_KeepsPartialFailureContract_WhenRunningSummaryResetsSavedCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	dictDB, modDB, cleanup := setupTestDB(t)
	defer cleanup()

	input := TerminologyInput{
		TaskID: "task-cached-summary-reset-partial",
		Entries: []TerminologyEntry{
			{
				ID:         "701",
				EditorID:   "EditorH",
				RecordType: "BOOK:FULL",
				SourceText: "Iron Sword",
				SourceFile: "mod_cached_partial.json",
				Variant:    "single",
			},
			{
				ID:         "702",
				EditorID:   "EditorH",
				RecordType: "ARMO:FULL",
				SourceText: "Steel Armor",
				SourceFile: "mod_cached_partial.json",
				Variant:    "single",
			},
		},
	}
	repo := &fakeTranslationInputRepository{input: input}
	builder := NewTermRequestBuilder(&TermRecordConfig{TargetRecordTypes: append([]string(nil), foundation.DictionaryImportRECTypes...)})
	stemmer := NewSnowballStemmer("english")
	dictRepo := dictionaryartifact.NewRepository(dictDB)
	searcher := NewSQLiteTermDictionarySearcher(dictRepo, logger, stemmer)
	store := NewSQLiteModTermStore(modDB, logger)
	promptBuilder, err := NewTermPromptBuilder("")
	if err != nil {
		t.Fatalf("failed to create prompt builder: %v", err)
	}
	translator := NewTermTranslator(repo, builder, searcher, store, promptBuilder, logger)

	requests, err := translator.PreparePrompts(ctx, input.TaskID, PhaseOptions{})
	if err != nil {
		t.Fatalf("PreparePrompts failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("unexpected request count: got=%d want=1", len(requests))
	}

	preparedSummary, err := store.GetPhaseSummary(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetPhaseSummary before reset failed: %v", err)
	}
	if err := translator.UpdatePhaseSummary(ctx, PhaseSummary{
		TaskID:          input.TaskID,
		Status:          "running",
		TargetCount:     preparedSummary.TargetCount,
		ProgressMode:    "determinate",
		ProgressCurrent: preparedSummary.SavedCount,
		ProgressTotal:   preparedSummary.TargetCount,
		ProgressMessage: "running",
	}); err != nil {
		t.Fatalf("UpdatePhaseSummary reset failed: %v", err)
	}

	responses := []llmio.Response{
		{
			Success:  false,
			Error:    "upstream timeout",
			Metadata: requests[0].Metadata,
		},
	}
	if err := translator.SaveResults(ctx, input.TaskID, responses); err != nil {
		t.Fatalf("SaveResults failed: %v", err)
	}

	summary, err := store.GetPhaseSummary(ctx, input.TaskID)
	if err != nil {
		t.Fatalf("GetPhaseSummary after save failed: %v", err)
	}
	if summary.Status != "completed_partial" {
		t.Fatalf("unexpected status: got=%q want=%q", summary.Status, "completed_partial")
	}
	if summary.SavedCount != 1 {
		t.Fatalf("unexpected saved count: got=%d want=1", summary.SavedCount)
	}
	if summary.FailedCount != 1 {
		t.Fatalf("unexpected failed count: got=%d want=1", summary.FailedCount)
	}
}

func contains(text string, want string) bool {
	return strings.Contains(text, want)
}

type fakeTranslationInputRepository struct {
	input TerminologyInput
}

func (f *fakeTranslationInputRepository) LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error) {
	_ = ctx
	_ = taskID
	return toArtifactInput(f.input), nil
}

func toArtifactInput(input TerminologyInput) translationinput.TerminologyInput {
	entries := make([]translationinput.TerminologyEntry, 0, len(input.Entries))
	for _, entry := range input.Entries {
		entries = append(entries, translationinput.TerminologyEntry{
			ID:         entry.ID,
			EditorID:   entry.EditorID,
			RecordType: entry.RecordType,
			SourceText: entry.SourceText,
			SourceFile: entry.SourceFile,
			PairKey:    entry.PairKey,
			Variant:    entry.Variant,
		})
	}
	return translationinput.TerminologyInput{
		TaskID:    input.TaskID,
		FileNames: append([]string(nil), input.FileNames...),
		Entries:   entries,
	}
}
