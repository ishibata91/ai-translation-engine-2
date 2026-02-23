package term_translator

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	_ "modernc.org/sqlite"
)

// setupTestDB creates in-memory SQLite DBs for testing
func setupTestDB(t *testing.T) (*sql.DB, *sql.DB, func()) {
	// Dictionary DB
	dictDB, err := sql.Open("sqlite", "file:dict?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open dict db: %v", err)
	}

	// Create dictionary schema and populate some data
	_, err = dictDB.Exec(`
		CREATE TABLE dictionary_terms (
			original_en TEXT PRIMARY KEY,
			translated_ja TEXT,
			record_type TEXT
		);
		CREATE VIRTUAL TABLE dictionary_terms_fts USING fts5(
			original_en,
			translated_ja,
			content='dictionary_terms',
			content_rowid='rowid'
		);
		INSERT INTO dictionary_terms (original_en, translated_ja, record_type) VALUES ('Iron Sword', '鉄の剣', 'WEAP');
	`)
	if err != nil {
		t.Fatalf("failed to init dict db: %v", err)
	}

	// Mod Term DB
	modDB, err := sql.Open("sqlite", "file:mod?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open mod db: %v", err)
	}

	return dictDB, modDB, func() {
		dictDB.Close()
		modDB.Close()
	}
}

func TestTermTranslatorSlice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name          string
		input         TermTranslatorInput
		config        TermRecordConfig
		mockLLMOutput []string // Resulting LLM content for each request
		expectedTerms map[string]string
	}{
		{
			name: "Two-Phase Translation: Item and magic",
			input: TermTranslatorInput{
				Items: []TermItem{
					{
						ID:   "001",
						Type: "WEAP",
						Name: stringPtr("Iron Sword"),
					},
					{
						ID:   "002",
						Type: "ARMO",
						Name: stringPtr("Steel Armor"),
					},
				},
				Magic: []TermMagic{
					{
						ID:   "003",
						Type: "SPEL",
						Name: stringPtr("Fireball"),
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"WEAP", "ARMO", "SPEL"},
			},
			mockLLMOutput: []string{
				"TL: |鉄の剣|", // Actually matches dict, but PreparePrompts will still generate a request
				"TL: |鋼鉄の鎧|",
				"TL: |ファイアボール|",
			},
			expectedTerms: map[string]string{
				"Iron Sword":  "鉄の剣",
				"Steel Armor": "鋼鉄の鎧",
				"Fireball":    "ファイアボール",
			},
		},
		{
			name: "Two-Phase Translation: Paired NPCs",
			input: TermTranslatorInput{
				NPCs: map[string]TermNPC{
					"npc1": {
						ID:       "101",
						EditorID: stringPtr("EditorID1"),
						Type:     "NPC_:FULL",
						Name:     "Uthgerd the Unbroken",
					},
					"npc2": {
						ID:       "102",
						EditorID: stringPtr("EditorID1"),
						Type:     "NPC_:SHRT",
						Name:     "Uthgerd",
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"NPC_:FULL", "NPC_:SHRT"},
			},
			mockLLMOutput: []string{
				"TL: |不屈のウスガルド 不屈|",
			},
			expectedTerms: map[string]string{
				"Uthgerd the Unbroken": "不屈のウスガルド 不屈",
				"Uthgerd":              "不屈のウスガルド",
			},
		},
		{
			name: "Error handling: Invalid format LLM response",
			input: TermTranslatorInput{
				Items: []TermItem{
					{
						ID:   "004",
						Type: "WEAP",
						Name: stringPtr("Rusty Dagger"),
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"WEAP"},
			},
			mockLLMOutput: []string{
				"I can't translate this.", // Missing TL: |...|
			},
			expectedTerms: map[string]string{}, // Should not be saved
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dictDB, modDB, cleanup := setupTestDB(t)
			defer cleanup()

			builder := NewTermRequestBuilder(&tc.config)
			stemmer := NewSnowballStemmer("english")
			searcher := NewSQLiteTermDictionarySearcher(dictDB, logger, stemmer)
			store := NewSQLiteModTermStore(modDB, logger)
			promptBuilder, err := NewTermPromptBuilder("")
			if err != nil {
				t.Fatalf("failed to create prompt builder: %v", err)
			}

			translator := NewTermTranslator(
				builder,
				searcher,
				store,
				promptBuilder,
				logger,
			)

			// Phase 1: Prepare Prompts
			llmRequests, err := translator.PreparePrompts(ctx, tc.input)
			if err != nil {
				t.Fatalf("PreparePrompts failed: %v", err)
			}

			if len(llmRequests) != len(tc.mockLLMOutput) {
				t.Fatalf("PreparePrompts generated %d requests, but mock has %d outputs", len(llmRequests), len(tc.mockLLMOutput))
			}

			// Simulate JobQueue/ProcessManager calling LLM
			llmResponses := make([]llm_client.Response, 0, len(llmRequests))
			for _, content := range tc.mockLLMOutput {
				llmResponses = append(llmResponses, llm_client.Response{
					Content: content,
					Success: true,
				})
			}

			// Phase 2: Save Results
			err = translator.SaveResults(ctx, tc.input, llmResponses)
			if err != nil {
				t.Fatalf("SaveResults failed: %v", err)
			}

			// Verify in DB
			for originalEN, expectedJA := range tc.expectedTerms {
				translatedJA, err := store.GetTerm(ctx, originalEN)
				if err != nil {
					t.Fatalf("Failed to get term %s: %v", originalEN, err)
				}
				if translatedJA != expectedJA {
					t.Errorf("Expected translation for %s to be '%s', got '%s'", originalEN, expectedJA, translatedJA)
				}
			}

			// Verify records NOT in expectedTerms are NOT in DB
			if len(tc.expectedTerms) == 0 {
				// Simple check for Rusty Dagger in the error case
				translatedJA, _ := store.GetTerm(ctx, "Rusty Dagger")
				if translatedJA != "" {
					t.Errorf("Expected 'Rusty Dagger' NOT to be in DB, but got '%s'", translatedJA)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
