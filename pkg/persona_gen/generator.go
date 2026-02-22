package persona_gen

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

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
	ConfigStore config_store.ConfigStore
	SecretStore config_store.SecretStore
}

// NewPersonaGenerator creates a new NPCPersonaGenerator.
func NewPersonaGenerator(
	collector DialogueCollector,
	evaluator ContextEvaluator,
	store PersonaStore,
	configStore config_store.ConfigStore,
	secretStore config_store.SecretStore,
) *DefaultPersonaGenerator {
	return &DefaultPersonaGenerator{
		Collector:   collector,
		Evaluator:   evaluator,
		Store:       store,
		ConfigStore: configStore,
		SecretStore: secretStore,
	}
}

// PreparePrompts processes input data and constructs LLM requests (Phase 1).
func (g *DefaultPersonaGenerator) PreparePrompts(
	ctx context.Context,
	data PersonaGenInput,
	config PersonaConfig,
) ([]llm_client.Request, error) {
	slog.DebugContext(ctx, "ENTER PreparePrompts",
		slog.String("slice", "PersonaGen"),
		slog.Int("npc_count", len(data.NPCs)),
	)

	groupedData, err := g.Collector.CollectByNPC(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to collect NPC dialogues: %w", err)
	}

	var requests []llm_client.Request
	for _, npcData := range groupedData {
		// Check if already generated
		exists, err := g.Store.GetPersona(ctx, npcData.SpeakerID)
		if err == nil && exists != "" {
			slog.DebugContext(ctx, "Skipping already generated persona",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", npcData.SpeakerID),
			)
			continue
		}

		if len(npcData.Dialogues) < config.MinDialogueThreshold {
			continue
		}

		_, selectedDialogues := g.Evaluator.Evaluate(ctx, npcData, config)
		if len(selectedDialogues) == 0 {
			continue
		}

		// Prepare user prompt
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("NPC Name: %s\n", npcData.NPCName))
		sb.WriteString(fmt.Sprintf("Race: %s\n", npcData.Race))
		sb.WriteString(fmt.Sprintf("Voice Type: %s\n", npcData.VoiceType))
		sb.WriteString("---- Dialogue History ----\n")
		for i, entry := range selectedDialogues {
			sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, entry.EnglishText))
		}
		sb.WriteString("\nFormat Requirement: Output the persona summary strictly within TL: |...| format.")

		requests = append(requests, llm_client.Request{
			SystemPrompt: personaSystemPrompt,
			UserPrompt:   sb.String(),
			MaxTokens:    config.MaxOutputTokens,
			Temperature:  0.3,
			Metadata: map[string]interface{}{
				"speaker_id": npcData.SpeakerID,
			},
		})
	}

	slog.DebugContext(ctx, "EXIT PreparePrompts",
		slog.String("slice", "PersonaGen"),
		slog.Int("request_count", len(requests)),
	)

	return requests, nil
}

// SaveResults parses LLM responses and persists them to the store (Phase 2).
func (g *DefaultPersonaGenerator) SaveResults(
	ctx context.Context,
	data PersonaGenInput,
	results []llm_client.Response,
) error {
	slog.DebugContext(ctx, "ENTER SaveResults",
		slog.String("slice", "PersonaGen"),
		slog.Int("response_count", len(results)),
	)

	successCount := 0
	failCount := 0

	for _, resp := range results {
		// Identify NPC from metadata
		speakerID, ok := resp.Metadata["speaker_id"].(string)
		if !ok || speakerID == "" {
			slog.WarnContext(ctx, "Missing speaker_id in response metadata", slog.String("slice", "PersonaGen"))
			failCount++
			continue
		}

		if !resp.Success {
			slog.WarnContext(ctx, "LLM response indicates failure",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", speakerID),
				slog.String("error", resp.Error),
			)
			failCount++
			continue
		}

		personaText := g.extractPersona(resp.Content)
		if personaText == "" {
			slog.WarnContext(ctx, "Failed to extract persona text from response",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", speakerID),
			)
			failCount++
			continue
		}

		if len(personaText) <= 5 {
			slog.WarnContext(ctx, "Extracted persona text is too short",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", speakerID),
				slog.String("content", personaText),
			)
			failCount++
			continue
		}

		// Prepare PersonaResult
		npc, exists := data.NPCs[speakerID]
		if !exists {
			slog.WarnContext(ctx, "NPC not found in original input",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", speakerID),
			)
			failCount++
			continue
		}

		var editorID string
		if npc.EditorID != nil {
			editorID = *npc.EditorID
		}

		result := PersonaResult{
			SpeakerID:   speakerID,
			EditorID:    editorID,
			NPCName:     npc.Name,
			Race:        npc.Type,
			PersonaText: personaText,
			Status:      "success",
		}

		if err := g.Store.SavePersona(ctx, result); err != nil {
			slog.ErrorContext(ctx, "Failed to save persona to store",
				slog.String("slice", "PersonaGen"),
				slog.String("speaker_id", speakerID),
				slog.String("error", err.Error()),
			)
			failCount++
			continue
		}

		successCount++
		slog.InfoContext(ctx, "Successfully saved persona",
			slog.String("slice", "PersonaGen"),
			slog.String("speaker_id", speakerID),
		)
	}

	slog.DebugContext(ctx, "EXIT SaveResults",
		slog.String("slice", "PersonaGen"),
		slog.Int("success", successCount),
		slog.Int("fail", failCount),
	)
	return nil
}

var personaRegex = regexp.MustCompile(`TL:\s*\|(.*?)\|`)

func (g *DefaultPersonaGenerator) extractPersona(content string) string {
	// 1. Regex search for TL: |...|
	match := personaRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	// 2. Fallback 1: search for TL: prefix and trim
	if idx := strings.Index(content, "TL:"); idx != -1 {
		text := content[idx+3:]
		// strip pipe if exists
		text = strings.TrimLeft(text, " |")
		if pipeIdx := strings.Index(text, "|"); pipeIdx != -1 {
			text = text[:pipeIdx]
		}
		return strings.TrimSpace(text)
	}

	// 3. Fallback 2: search for just |...|
	start := strings.Index(content, "|")
	end := strings.LastIndex(content, "|")
	if start != -1 && end != -1 && end > start {
		return strings.TrimSpace(content[start+1 : end])
	}

	return ""
}