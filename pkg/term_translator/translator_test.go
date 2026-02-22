package term_translator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	_ "modernc.org/sqlite"
)

// mockLLMClient implements llm_client.LLMClient for testing
type mockLLMClient struct {
	responses map[string]string
}

func (m *mockLLMClient) Complete(ctx context.Context, req llm_client.Request) (llm_client.Response, error) {
	// Find the source text in the prompt to return a mocked translation
	for source, trans := range m.responses {
		if strings.Contains(req.SystemPrompt, source) {
			return llm_client.Response{
				Content: fmt.Sprintf("TL: |%s|", trans),
				Success: true,
			}, nil
		}
	}
	return llm_client.Response{
		Content: "TL: |Mocked Translation|",
		Success: true,
	}, nil
}

func (m *mockLLMClient) StreamComplete(ctx context.Context, req llm_client.Request) (llm_client.StreamResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockLLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockLLMClient) HealthCheck(ctx context.Context) error {
	return nil
}

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
	// 5.2 Require OpenTelemetry Context (Simulated with standard context here, ideally with otel SDK)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name          string
		input         models.ExtractedData
		config        TermRecordConfig
		mockResponses map[string]string
		expectedTerms map[string]string
	}{
		{
			name: "Translate Item and magic with exact dictionary match and LLM",
			input: models.ExtractedData{
				Items: []models.Item{
					{
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "001", Type: "WEAP"},
						Name:                stringPtr("Iron Sword"),
					},
					{
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "002", Type: "ARMO"},
						Name:                stringPtr("Steel Armor"),
					},
				},
				Magic: []models.Magic{
					{
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "003", Type: "SPEL"},
						Name:                stringPtr("Fireball"),
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"WEAP", "ARMO", "SPEL"},
			},
			mockResponses: map[string]string{
				"Steel Armor": "鋼鉄の鎧",
				"Fireball":    "ファイアボール",
			},
			expectedTerms: map[string]string{
				"Iron Sword":  "鉄の剣",     // From Dict DB direct match
				"Steel Armor": "鋼鉄の鎧",    // From LLM
				"Fireball":    "ファイアボール", // From LLM
			},
		},
		{
			name: "Translate Paired NPCs",
			input: models.ExtractedData{
				NPCs: map[string]models.NPC{
					"npc1": {
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "101", EditorID: stringPtr("EditorID1"), Type: "NPC_:FULL"},
						Name:                "Uthgerd the Unbroken",
					},
					"npc2": {
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "102", EditorID: stringPtr("EditorID1"), Type: "NPC_:SHRT"},
						Name:                "Uthgerd",
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"NPC_:FULL", "NPC_:SHRT"},
			},
			mockResponses: map[string]string{
				"Uthgerd the Unbroken": "不屈のウスガルド 不屈", // mock combined result for split logic
			},
			expectedTerms: map[string]string{
				"Uthgerd the Unbroken": "不屈のウスガルド 不屈",
				"Uthgerd":              "不屈のウスガルド", // strings.Split("不屈のウスガルド 不屈", " ")[0]
			},
		},
		{
			name: "Filtered records are ignored",
			input: models.ExtractedData{
				Items: []models.Item{
					{
						BaseExtractedRecord: models.BaseExtractedRecord{ID: "001", Type: "MISC"},
						Name:                stringPtr("Gold Coin"),
					},
				},
			},
			config: TermRecordConfig{
				TargetRecordTypes: []string{"WEAP", "ARMO"}, // MISC is filtered
			},
			mockResponses: map[string]string{
				"Gold Coin": "ゴールドコイン",
			},
			expectedTerms: map[string]string{}, // Should be empty
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
			llm := &mockLLMClient{responses: tc.mockResponses}
			promptBuilder, err := NewTermPromptBuilder("")
			if err != nil {
				t.Fatalf("failed to create prompt builder: %v", err)
			}

			translator := NewTermTranslator(
				builder,
				searcher,
				store,
				llm,
				promptBuilder,
				logger,
				nil,
			)

			// Execute translation
			results, err := translator.TranslateTerms(ctx, tc.input)
			if err != nil {
				t.Fatalf("TranslateTerms failed: %v", err)
			}

			if len(tc.expectedTerms) == 0 && len(results) > 0 {
				t.Fatalf("Expected no results, got %d", len(results))
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
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
