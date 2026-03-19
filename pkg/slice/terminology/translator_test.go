package terminology

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

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
			source_text TEXT,
			dest_text TEXT,
			record_type TEXT
		);
		INSERT INTO artifact_dictionary_entries (source_text, dest_text, record_type) VALUES
			('Iron Sword', '鉄の剣', 'BOOK:FULL'),
			('Uthgerd the Unbroken', '不屈のウスガルド', 'NPC_:FULL');
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
				"TL: |鉄の剣|",
				"TL: |鋼鉄の鎧|",
				"TL: |不屈のウスガルド 不屈|",
			},
			expectedReqs: 3,
			expectedTerms: map[string]string{
				"Iron Sword":           "鉄の剣",
				"Steel Armor":          "鋼鉄の鎧",
				"Uthgerd the Unbroken": "不屈のウスガルド 不屈",
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
				"TL: |鉄の剣|",
				"TL: |鋼鉄の鎧|",
			},
			expectedReqs: 2,
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
			searcher := NewSQLiteTermDictionarySearcher(dictDB, logger, stemmer)
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
			if phaseSummary.TargetCount != tc.expectedReqs {
				t.Fatalf("unexpected target count: got=%d want=%d", phaseSummary.TargetCount, tc.expectedReqs)
			}
		})
	}
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
