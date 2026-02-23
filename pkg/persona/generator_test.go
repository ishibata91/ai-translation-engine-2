package persona_test

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
	persona "github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	_ "modernc.org/sqlite"
)

// mockConfigStore implements config.Config for testing.
type mockConfigStore struct{}

func (m *mockConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	return "", nil
}
func (m *mockConfigStore) Set(ctx context.Context, namespace string, key string, value string) error {
	return nil
}
func (m *mockConfigStore) Delete(ctx context.Context, namespace string, key string) error { return nil }
func (m *mockConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	return nil, nil
}
func (m *mockConfigStore) Watch(namespace string, key string, callback config.ChangeCallback) config.UnsubscribeFunc {
	return func() {}
}

// mockSecretStore implements config.SecretStore for testing.
type mockSecretStore struct{}

func (m *mockSecretStore) GetSecret(ctx context.Context, namespace string, key string) (string, error) {
	return "", nil
}
func (m *mockSecretStore) SetSecret(ctx context.Context, namespace string, key string, value string) error {
	return nil
}
func (m *mockSecretStore) DeleteSecret(ctx context.Context, namespace string, key string) error {
	return nil
}
func (m *mockSecretStore) ListSecretKeys(ctx context.Context, namespace string) ([]string, error) {
	return nil, nil
}

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", "file:personas?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open personas db: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func strPtr(s string) *string {
	return &s
}

func TestPersonaGenSlice_TableDriven(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Logger setup for test
	slog.SetDefault(slog.Default())

	tests := []struct {
		name                 string
		input                persona.PersonaGenInput
		config               persona.PersonaConfig
		mockLLMOutput        []string
		expectedRequestCount int
		expectedDBCount      int
	}{
		{
			name: "Phase 1 & 2: Normal flow",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"NPC001": {ID: "NPC001", Name: "Aela", Type: "Nord"},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D1", SpeakerID: strPtr("NPC001"), Text: strPtr("Dialogue 1"), Order: 1},
					{ID: "D2", SpeakerID: strPtr("NPC001"), Text: strPtr("Dialogue 2"), Order: 2},
				},
			},
			config: persona.PersonaConfig{
				MinDialogueThreshold: 1,
				ContextWindowLimit:   4000,
				MaxOutputTokens:      500,
			},
			mockLLMOutput: []string{
				"TL: |Personality: Brave, habits: direct|",
			},
			expectedRequestCount: 1,
			expectedDBCount:      1,
		},
		{
			name: "Phase 2: Fallback parsing - simple TL: prefix",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"NPC002": {ID: "NPC002", Name: "Farkas", Type: "Nord"},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D3", SpeakerID: strPtr("NPC002"), Text: strPtr("I am strong."), Order: 1},
				},
			},
			mockLLMOutput: []string{
				"TL: Personality: Simple and loyal.",
			},
			expectedRequestCount: 1,
			expectedDBCount:      1,
		},
		{
			name: "Phase 2: Fallback parsing - just pipes",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"NPC003": {ID: "NPC003", Name: "Vilkas", Type: "Nord"},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D4", SpeakerID: strPtr("NPC003"), Text: strPtr("I am smart."), Order: 1},
				},
			},
			mockLLMOutput: []string{
				"Here is the persona: |Personality: Smart and tactical|",
			},
			expectedRequestCount: 1,
			expectedDBCount:      1,
		},
		{
			name: "Phase 2: Failure - content too short",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"NPC004": {ID: "NPC004", Name: "Kodlak", Type: "Nord"},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D5", SpeakerID: strPtr("NPC004"), Text: strPtr("Old age."), Order: 1},
				},
			},
			mockLLMOutput: []string{
				"TL: |Old|",
			},
			expectedRequestCount: 1,
			expectedDBCount:      0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			store := persona.NewPersonaStore(db)
			if err := store.InitSchema(ctx); err != nil {
				t.Fatalf("Failed to init schema: %v", err)
			}

			collector := persona.NewDefaultDialogueCollector()
			scorer := persona.NewDefaultScorer()
			estimator := persona.NewSimpleTokenEstimator()
			evaluator := persona.NewDefaultContextEvaluator(scorer, estimator)

			configStore := &mockConfigStore{}
			secretStore := &mockSecretStore{}

			generator := persona.NewPersonaGenerator(collector, evaluator, store, configStore, secretStore)

			// Phase 1
			requests, err := generator.PreparePrompts(ctx, tc.input)
			if err != nil {
				t.Fatalf("PreparePrompts failed: %v", err)
			}
			if len(requests) != tc.expectedRequestCount {
				t.Errorf("Expected %d requests, got %d", tc.expectedRequestCount, len(requests))
			}

			// Simulate JobQueue/Pipeline calling LLM
			llmResponses := make([]llm.Response, 0, len(requests))
			for i, content := range tc.mockLLMOutput {
				if i >= len(requests) {
					break
				}
				llmResponses = append(llmResponses, llm.Response{
					Content:  content,
					Success:  true,
					Metadata: requests[i].Metadata,
				})
			}

			// Phase 2
			err = generator.SaveResults(ctx, llmResponses)
			if err != nil {
				t.Fatalf("SaveResults failed: %v", err)
			}

			// Verify DB
			for id := range tc.input.NPCs {
				personaText, err := store.GetPersona(ctx, id)
				if err != nil {
					t.Fatalf("Failed to get persona for %s: %v", id, err)
				}
				if tc.expectedDBCount > 0 && personaText == "" && tc.name != "Phase 2: Failure - content too short" {
					t.Errorf("Expected persona for %s to be saved, but it was empty", id)
				}
				if tc.expectedDBCount == 0 && personaText != "" {
					t.Errorf("Expected persona for %s NOT to be saved, but it was found", id)
				}
			}
		})
	}
}
