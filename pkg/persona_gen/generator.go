package persona_gen

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

const personaSystemPrompt = `You are a character persona analyzer.
Based on the provided dialogue history of an NPC, generate a concise persona summary.
Your response MUST be formatted strictly with the following sections:
- Personality Traits: (brief summary of character's personality)
- Speaking Habits: (notes on tone, vocabulary, verbal tics)
- Background: (implicit backstory or relationships inferred from dialogue)

Keep the total response under 150 words. Do not include extra conversational filler.
Format:
Personality Traits: ...
Speaking Habits: ...
Background: ...`

// DefaultPersonaGenerator implements the NPCPersonaGenerator interface.
type DefaultPersonaGenerator struct {
	Collector   DialogueCollector
	Evaluator   ContextEvaluator
	Store       PersonaStore
	LLMManager  llm_client.LLMManager
	ConfigStore config_store.ConfigStore
	SecretStore config_store.SecretStore
}

// NewPersonaGenerator creates a new NPCPersonaGenerator.
func NewPersonaGenerator(
	collector DialogueCollector,
	evaluator ContextEvaluator,
	store PersonaStore,
	llmManager llm_client.LLMManager,
	configStore config_store.ConfigStore,
	secretStore config_store.SecretStore,
) *DefaultPersonaGenerator {
	return &DefaultPersonaGenerator{
		Collector:   collector,
		Evaluator:   evaluator,
		Store:       store,
		LLMManager:  llmManager,
		ConfigStore: configStore,
		SecretStore: secretStore,
	}
}

// GeneratePersonas processes input data to generate and store personas for all NPCs.
// Note: It ignores the passed llmConfig parameter and dynamically fetches it.
func (g *DefaultPersonaGenerator) GeneratePersonas(
	ctx context.Context,
	data PersonaGenInput,
	config PersonaConfig,
	_ llm_client.LLMConfig, // Kept for interface compatibility if needed, but we override it
	notifier ProgressNotifier,
) ([]PersonaResult, error) {
	slog.DebugContext(ctx, "ENTER GeneratePersonas",
		slog.String("slice", "PersonaGen"),
		slog.Int("npc_count", len(data.NPCs)),
	)
	start := time.Now()

	llmConfig, err := g.fetchLLMConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LLM config: %w", err)
	}

	client, err := g.LLMManager.GetClient(ctx, llmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM client: %w", err)
	}

	groupedData, err := g.Collector.CollectByNPC(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to collect NPC dialogues: %w", err)
	}

	var results []PersonaResult
	total := len(groupedData)
	completed := 0

	for _, npcData := range groupedData {
		result := g.processNPC(ctx, client, npcData, config)
		results = append(results, result)

		if result.Status == "success" {
			// Save to store
			err := g.Store.SavePersona(ctx, result)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to save persona",
					slog.String("slice", "PersonaGen"),
					slog.String("speaker_id", result.SpeakerID),
					slog.String("error", err.Error()),
				)
				result.Status = "error"
				result.ErrorMessage = fmt.Sprintf("db save failed: %s", err.Error())
			}
		}

		completed++
		if notifier != nil {
			notifier.OnProgress(completed, total)
		}
	}

	slog.DebugContext(ctx, "EXIT GeneratePersonas",
		slog.String("slice", "PersonaGen"),
		slog.Int("processed_count", len(results)),
		slog.Duration("elapsed", time.Since(start)),
	)

	return results, nil
}

func (g *DefaultPersonaGenerator) processNPC(
	ctx context.Context,
	client llm_client.LLMClient,
	npcData NPCDialogueData,
	config PersonaConfig,
) PersonaResult {
	slog.DebugContext(ctx, "ENTER processNPC",
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", npcData.SpeakerID),
		slog.Int("raw_dialogue_count", len(npcData.Dialogues)),
	)

	result := PersonaResult{
		SpeakerID: npcData.SpeakerID,
		EditorID:  npcData.EditorID,
		NPCName:   npcData.NPCName,
		Race:      npcData.Race,
		Sex:       npcData.Sex,
		VoiceType: npcData.VoiceType,
		Status:    "processing",
	}

	if len(npcData.Dialogues) < config.MinDialogueThreshold {
		result.Status = "skipped"
		result.ErrorMessage = fmt.Sprintf("below minimum dialogue threshold (%d < %d)", len(npcData.Dialogues), config.MinDialogueThreshold)
		return result
	}

	estimation, selectedDialogues := g.Evaluator.Evaluate(ctx, npcData, config)

	if len(selectedDialogues) == 0 {
		result.Status = "skipped"
		result.ErrorMessage = "no dialogues remained after context evaluation"
		return result
	}

	// Prepare user prompt
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("NPC Name: %s\n", npcData.NPCName))
	sb.WriteString(fmt.Sprintf("Race: %s\n", npcData.Race))
	sb.WriteString(fmt.Sprintf("Voice Type: %s\n", npcData.VoiceType))
	sb.WriteString("---- Dialogue History ----\n")
	for i, entry := range selectedDialogues {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, entry.EnglishText)) // Use EnglishText for generating persona
	}

	userPrompt := sb.String()

	req := llm_client.Request{
		SystemPrompt: personaSystemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    config.MaxOutputTokens,
		Temperature:  0.3, // Low temperature for more analytical/consistent extraction
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("LLM completion failed: %s", err.Error())
		return result
	}

	if !resp.Success {
		result.Status = "error"
		result.ErrorMessage = fmt.Sprintf("LLM API returned failure: %s", resp.Error)
		return result
	}

	result.PersonaText = resp.Content
	result.DialogueCount = len(selectedDialogues)
	result.EstimatedTokens = estimation.TotalTokens
	result.Status = "success"

	slog.DebugContext(ctx, "EXIT processNPC",
		slog.String("slice", "PersonaGen"),
		slog.String("speaker_id", npcData.SpeakerID),
		slog.String("status", result.Status),
	)

	return result
}

// fetchLLMConfig fetches the LLM configuration from the ConfigStore and SecretStore.
func (g *DefaultPersonaGenerator) fetchLLMConfig(ctx context.Context) (llm_client.LLMConfig, error) {
	// First check PersonaGen specific settings, fallback to global LLM settings

	// Provider
	provider, err := g.ConfigStore.Get(ctx, "persona_gen", "llm_provider")
	if err != nil || provider == "" {
		provider, _ = g.ConfigStore.Get(ctx, "llm", "default_provider")
		if provider == "" {
			provider = "gemini" // Default fallback
		}
	}

	// Model
	model, err := g.ConfigStore.Get(ctx, "persona_gen", "llm_model")
	if err != nil || model == "" {
		model, _ = g.ConfigStore.Get(ctx, "llm", provider+"_default_model")
	}

	// Endpoint (optional, useful for local/custom endpoints)
	endpoint, err := g.ConfigStore.Get(ctx, "persona_gen", "llm_endpoint")
	if err != nil || endpoint == "" {
		endpoint, _ = g.ConfigStore.Get(ctx, "llm", provider+"_endpoint")
	}

	// API Key
	apiKey, err := g.SecretStore.GetSecret(ctx, "persona_gen", "llm_api_key")
	if err != nil || apiKey == "" {
		apiKey, _ = g.SecretStore.GetSecret(ctx, "llm", provider+"_api_key")
	}

	return llm_client.LLMConfig{
		Provider: provider,
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
	}, nil
}
