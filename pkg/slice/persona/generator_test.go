package persona

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"testing"
	"time"

	master_persona_artifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/master_persona_artifact"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// mockConfigStore implements config.Config for testing.
type mockConfigStore struct {
	values map[string]map[string]string
}

func (m *mockConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	if m.values == nil {
		return "", nil
	}
	if bucket, ok := m.values[namespace]; ok {
		return bucket[key], nil
	}
	return "", nil
}
func (m *mockConfigStore) Set(ctx context.Context, namespace string, key string, value string) error {
	return nil
}
func (m *mockConfigStore) Delete(ctx context.Context, namespace string, key string) error { return nil }
func (m *mockConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	if m.values == nil {
		return nil, nil
	}
	if bucket, ok := m.values[namespace]; ok {
		out := make(map[string]string, len(bucket))
		for k, v := range bucket {
			out[k] = v
		}
		return out, nil
	}
	return nil, nil
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
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
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
		input                PersonaGenInput
		config               PersonaConfig
		mockLLMOutput        []string
		expectedRequestCount int
		expectedDBCount      int
	}{
		{
			name: "Phase 1 & 2: Normal flow",
			input: PersonaGenInput{
				NPCs: map[string]PersonaNPC{
					"NPC001": {ID: "NPC001", Name: "Aela", Type: "Nord"},
				},
				Dialogues: []PersonaDialogue{
					{ID: "D1", SpeakerID: strPtr("NPC001"), Text: strPtr("Dialogue 1"), Order: 1},
					{ID: "D2", SpeakerID: strPtr("NPC001"), Text: strPtr("Dialogue 2"), Order: 2},
				},
			},
			config: PersonaConfig{
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
			input: PersonaGenInput{
				NPCs: map[string]PersonaNPC{
					"NPC002": {ID: "NPC002", Name: "Farkas", Type: "Nord"},
				},
				Dialogues: []PersonaDialogue{
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
			input: PersonaGenInput{
				NPCs: map[string]PersonaNPC{
					"NPC003": {ID: "NPC003", Name: "Vilkas", Type: "Nord"},
				},
				Dialogues: []PersonaDialogue{
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
			input: PersonaGenInput{
				NPCs: map[string]PersonaNPC{
					"NPC004": {ID: "NPC004", Name: "Kodlak", Type: "Nord"},
				},
				Dialogues: []PersonaDialogue{
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

			require.NoError(t, master_persona_artifact.Migrate(ctx, db))
			store := NewPersonaStore(master_persona_artifact.NewRepository(db))
			if err := store.InitSchema(ctx); err != nil {
				t.Fatalf("Failed to init schema: %v", err)
			}

			collector := NewDefaultDialogueCollector()
			scorer := NewDefaultScorer()
			estimator := NewSimpleTokenEstimator()
			evaluator := NewDefaultContextEvaluator(scorer, estimator)

			configStore := &mockConfigStore{}
			secretStore := &mockSecretStore{}

			generator := NewPersonaGenerator(collector, evaluator, store, configStore, secretStore)

			// Phase 1
			requests, err := generator.PreparePrompts(ctx, tc.input)
			if err != nil {
				t.Fatalf("PreparePrompts failed: %v", err)
			}
			if len(requests) != tc.expectedRequestCount {
				t.Errorf("Expected %d requests, got %d", tc.expectedRequestCount, len(requests))
			}

			// Simulate JobQueue/Pipeline calling LLM
			llmResponses := make([]llmio.Response, 0, len(requests))
			for i, content := range tc.mockLLMOutput {
				if i >= len(requests) {
					break
				}
				llmResponses = append(llmResponses, llmio.Response{
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
				personaText, err := store.GetPersona(ctx, "", id)
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

func TestPersonaGenSlice_UsesConfiguredPromptSplit(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	require.NoError(t, master_persona_artifact.Migrate(ctx, db))
	store := NewPersonaStore(master_persona_artifact.NewRepository(db))
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	collector := NewDefaultDialogueCollector()
	scorer := NewDefaultScorer()
	estimator := NewSimpleTokenEstimator()
	evaluator := NewDefaultContextEvaluator(scorer, estimator)

	configStore := &mockConfigStore{
		values: map[string]map[string]string{
			masterPersonaPromptNamespace: {
				masterPersonaUserPromptKey:   "会話から口調と性格を抽出してください。",
				masterPersonaSystemPromptKey: "SYSTEM RULES",
			},
		},
	}
	secretStore := &mockSecretStore{}
	generator := NewPersonaGenerator(collector, evaluator, store, configStore, secretStore)

	requests, err := generator.PreparePrompts(ctx, PersonaGenInput{
		NPCs: map[string]PersonaNPC{
			"NPC001": {ID: "NPC001", Name: "Aela", Race: "Nord", VoiceType: "FemaleYoungEager"},
		},
		Dialogues: []PersonaDialogue{
			{ID: "D1", SpeakerID: strPtr("NPC001"), Text: strPtr("We hunt as one."), Order: 1},
		},
	})
	if err != nil {
		t.Fatalf("PreparePrompts failed: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one request, got %d", len(requests))
	}
	if requests[0].SystemPrompt != "SYSTEM RULES" {
		t.Fatalf("unexpected system prompt: %q", requests[0].SystemPrompt)
	}
	if requests[0].UserPrompt == "" || requests[0].UserPrompt[:len("会話から口調と性格を抽出してください。")] != "会話から口調と性格を抽出してください。" {
		t.Fatalf("unexpected user prompt prefix: %q", requests[0].UserPrompt)
	}
	if !strings.Contains(requests[0].UserPrompt, "Dialogue History:") {
		t.Fatalf("expected dialogue history block in user prompt: %q", requests[0].UserPrompt)
	}

}
