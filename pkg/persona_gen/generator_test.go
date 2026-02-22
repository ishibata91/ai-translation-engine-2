package persona_gen_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
	persona "github.com/ishibata91/ai-translation-engine-2/pkg/persona_gen"
	_ "modernc.org/sqlite"
)

// mockLLMClient implements llm_client.LLMClient for testing.
type mockLLMClient struct {
	responses map[string]string
}

func (m *mockLLMClient) Complete(ctx context.Context, req llm_client.Request) (llm_client.Response, error) {
	for name, response := range m.responses {
		if strings.Contains(req.UserPrompt, name) {
			return llm_client.Response{
				Content: response,
				Success: true,
			}, nil
		}
	}
	return llm_client.Response{
		Content: "Personality Traits: Generic\nSpeaking Habits: Normal\nBackground: None",
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

// mockLLMManager implements llm_client.LLMManager for testing.
type mockLLMManager struct {
	client *mockLLMClient
}

func (m *mockLLMManager) GetClient(ctx context.Context, config llm_client.LLMConfig) (llm_client.LLMClient, error) {
	return m.client, nil
}

func (m *mockLLMManager) GetBatchClient(ctx context.Context, config llm_client.LLMConfig) (llm_client.BatchClient, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockLLMManager) ResolveBulkStrategy(ctx context.Context, strategy llm_client.BulkStrategy, provider string) llm_client.BulkStrategy {
	if provider == "local" {
		return llm_client.BulkStrategySync
	}
	if strategy == "" {
		return llm_client.BulkStrategySync
	}
	return strategy
}

// mockConfigStore implements config_store.ConfigStore for testing.
type mockConfigStore struct {
	configs map[string]string
}

func (m *mockConfigStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	fullKey := namespace + ":" + key
	if val, ok := m.configs[fullKey]; ok {
		return val, nil
	}
	return "", fmt.Errorf("not found")
}

func (m *mockConfigStore) Set(ctx context.Context, namespace string, key string, value string) error {
	return nil
}

func (m *mockConfigStore) Delete(ctx context.Context, namespace string, key string) error {
	return nil
}

func (m *mockConfigStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	return nil, nil
}

func (m *mockConfigStore) Watch(namespace string, key string, callback config_store.ChangeCallback) config_store.UnsubscribeFunc {
	return func() {}
}

// mockSecretStore implements config_store.SecretStore for testing.
type mockSecretStore struct {
	secrets map[string]string
}

func (m *mockSecretStore) GetSecret(ctx context.Context, namespace string, key string) (string, error) {
	fullKey := namespace + ":" + key
	if val, ok := m.secrets[fullKey]; ok {
		return val, nil
	}
	return "", fmt.Errorf("not found")
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

// setupTestDB creates in-memory SQLite DB for testing
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

func TestPersonaGenSlice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		input         persona.PersonaGenInput
		config        persona.PersonaConfig
		mockResponses map[string]string
		expectedRes   map[string]string // SpeakerID -> expected status
		expectedDB    map[string]string // SpeakerID -> expected persona text
	}{
		{
			name: "Generate persona for an NPC meeting threshold",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"Aela": {ID: "Aela", Name: "Aela the Huntress", Type: "Nord", EditorID: strPtr("AelaEditorID")},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D1", SpeakerID: strPtr("Aela"), Text: strPtr("Something has shifted in the moon's light."), Order: 1},
					{ID: "D2", SpeakerID: strPtr("Aela"), Text: strPtr("I am a Companion. We fight for honor."), Order: 2},
				},
			},
			config: persona.PersonaConfig{
				MinDialogueThreshold: 2,
				ContextWindowLimit:   8000,
				SystemPromptOverhead: 100,
				MaxOutputTokens:      500,
			},
			mockResponses: map[string]string{
				"Aela the Huntress": "Personality Traits: Brave\nSpeaking Habits: Direct\nBackground: Companion",
			},
			expectedRes: map[string]string{
				"Aela": "success",
			},
			expectedDB: map[string]string{
				"Aela": "Personality Traits: Brave\nSpeaking Habits: Direct\nBackground: Companion",
			},
		},
		{
			name: "Skip NPC below dialogue threshold",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"Nazeem": {ID: "Nazeem", Name: "Nazeem", Type: "Redguard"},
				},
				Dialogues: []persona.PersonaDialogue{
					{ID: "D1", SpeakerID: strPtr("Nazeem"), Text: strPtr("Do you get to the Cloud District very often?"), Order: 1},
				},
			},
			config: persona.PersonaConfig{
				MinDialogueThreshold: 2, // Needs 2, only has 1
				ContextWindowLimit:   8000,
			},
			mockResponses: map[string]string{},
			expectedRes: map[string]string{
				"Nazeem": "skipped",
			},
			expectedDB: map[string]string{}, // Should not be in DB
		},
		{
			name: "Context limit truncation test",
			input: persona.PersonaGenInput{
				NPCs: map[string]persona.PersonaNPC{
					"Talkative": {ID: "Talkative", Name: "Talkative NPC", Type: "Imperial"},
				},
				Dialogues: []persona.PersonaDialogue{
					// Quest dialogue (higher priority via scoring)
					{ID: "D1", SpeakerID: strPtr("Talkative"), Text: strPtr("This is a very important quest dialogue with Ulfric!"), QuestID: strPtr("MQ101"), Order: 1},
					// Normal generic dialogue
					{ID: "D2", SpeakerID: strPtr("Talkative"), Text: strPtr("Hello there."), Order: 2},
				},
			},
			config: persona.PersonaConfig{
				MinDialogueThreshold: 1,
				ContextWindowLimit:   100 + 40, // System overhead + just enough for one dialogue
				SystemPromptOverhead: 100,
				MaxOutputTokens:      500,
			},
			mockResponses: map[string]string{
				"Talkative NPC": "Summary generated",
			},
			expectedRes: map[string]string{
				"Talkative": "success",
			},
			expectedDB: map[string]string{
				"Talkative": "Summary generated",
			},
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
			scorer.WeightNoun = 2
			scorer.WeightEmotion = 1

			estimator := persona.NewSimpleTokenEstimator()
			estimator.CharsPerToken = 4

			evaluator := persona.NewDefaultContextEvaluator(scorer, estimator)

			llmMock := &mockLLMClient{responses: tc.mockResponses}
			llmManager := &mockLLMManager{client: llmMock}

			configStore := &mockConfigStore{
				configs: map[string]string{
					"llm:default_provider": "mock",
				},
			}
			secretStore := &mockSecretStore{
				secrets: map[string]string{
					"llm:mock_api_key": "dummy",
				},
			}

			generator := persona.NewPersonaGenerator(collector, evaluator, store, llmManager, configStore, secretStore)

			results, err := generator.GeneratePersonas(ctx, tc.input, tc.config, llm_client.LLMConfig{}, nil)
			if err != nil {
				t.Fatalf("GeneratePersonas returned error: %v", err)
			}

			if len(results) != len(tc.expectedRes) {
				t.Errorf("Expected %d results, got %d", len(tc.expectedRes), len(results))
			}

			// Validate results
			for _, res := range results {
				expectedStatus, ok := tc.expectedRes[res.SpeakerID]
				if !ok {
					t.Errorf("Unexpected speaker ID in result: %s", res.SpeakerID)
					continue
				}
				if res.Status != expectedStatus {
					t.Errorf("Speaker %s: expected status %s, got %s: %s", res.SpeakerID, expectedStatus, res.Status, res.ErrorMessage)
				}
			}

			// Validate DB state
			for expectedSpeaker, expectedPersona := range tc.expectedDB {
				actualPersona, err := store.GetPersona(ctx, expectedSpeaker)
				if err != nil {
					t.Errorf("Failed to retrieve persona for %s from DB: %v", expectedSpeaker, err)
					continue
				}
				if actualPersona != expectedPersona {
					t.Errorf("Speaker %s: expected DB persona %q, got %q", expectedSpeaker, expectedPersona, actualPersona)
				}
			}
		})
	}
}
