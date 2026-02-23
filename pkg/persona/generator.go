package persona

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
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
	Config      config.Config
	SecretStore config.SecretStore
}

// NewPersonaGenerator creates a new NPCPersonaGenerator.
func NewPersonaGenerator(
	collector DialogueCollector,
	evaluator ContextEvaluator,
	store PersonaStore,
	configStore config.Config,
	secretStore config.SecretStore,
) *DefaultPersonaGenerator {
	return &DefaultPersonaGenerator{
		Collector:   collector,
		Evaluator:   evaluator,
		Store:       store,
		Config:      configStore,
		SecretStore: secretStore,
	}
}

// ID returns the unique identifier of the slice.
func (g *DefaultPersonaGenerator) ID() string {
	return "Persona"
}

// PreparePrompts processes input data and constructs LLM requests (Phase 1).
func (g *DefaultPersonaGenerator) PreparePrompts(
	ctx context.Context,
	input any,
) ([]llm.Request, error) {
	data, ok := input.(PersonaGenInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Persona slice: %T", input)
	}

	slog.DebugContext(ctx, "ENTER PreparePrompts",
		slog.String("slice", "Persona"),
		slog.Int("npc_count", len(data.NPCs)),
	)

	// Note: default config for now. In a real scenario, this might be injected or part of input.
	config := PersonaConfig{
		MinDialogueThreshold: 1,
		ContextWindowLimit:   4000,
		MaxOutputTokens:      500,
	}

	groupedData, err := g.Collector.CollectByNPC(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to collect NPC dialogues: %w", err)
	}

	var requests []llm.Request
	for _, npcData := range groupedData {
		// Check if already generated
		exists, err := g.Store.GetPersona(ctx, npcData.SpeakerID)
		if err == nil && exists != "" {
			slog.DebugContext(ctx, "Skipping already generated persona",
				slog.String("slice", "Persona"),
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

		requests = append(requests, llm.Request{
			SystemPrompt: personaSystemPrompt,
			UserPrompt:   sb.String(),
			MaxTokens:    config.MaxOutputTokens,
			Temperature:  0.3,
			Metadata: map[string]interface{}{
				"speaker_id": npcData.SpeakerID,
				"npc_name":   npcData.NPCName,
				"race":       npcData.Race,
				"editor_id":  npcData.EditorID,
			},
		})
	}

	slog.DebugContext(ctx, "EXIT PreparePrompts",
		slog.String("slice", "Persona"),
		slog.Int("request_count", len(requests)),
	)

	return requests, nil
}

// SaveResults parses LLM responses and persists them to the store (Phase 2).
func (g *DefaultPersonaGenerator) SaveResults(
	ctx context.Context,
	results []llm.Response,
) error {
	slog.DebugContext(ctx, "ENTER SaveResults",
		slog.String("slice", "Persona"),
		slog.Int("response_count", len(results)),
	)

	// Note: This implementation is currently limited because it doesn't have access
	// to the original PersonaGenInput.NPCs needed to create PersonaResult.
	// We might need to store this in state or include it in metadata.
	// For now, we'll try to extract as much as possible from metadata.

	successCount := 0
	failCount := 0

	for _, resp := range results {
		// Identify NPC from metadata
		speakerID, ok := resp.Metadata["speaker_id"].(string)
		if !ok || speakerID == "" {
			slog.WarnContext(ctx, "Missing speaker_id in response metadata", slog.String("slice", "Persona"))
			failCount++
			continue
		}

		if !resp.Success {
			slog.WarnContext(ctx, "LLM response indicates failure",
				slog.String("slice", "Persona"),
				slog.String("speaker_id", speakerID),
				slog.String("error", resp.Error),
			)
			failCount++
			continue
		}

		personaText := g.extractPersona(resp.Content)
		if personaText == "" {
			slog.WarnContext(ctx, "Failed to extract persona text from response",
				slog.String("slice", "Persona"),
				slog.String("speaker_id", speakerID),
			)
			failCount++
			continue
		}

		if len(personaText) <= 5 {
			slog.WarnContext(ctx, "Extracted persona text is too short",
				slog.String("slice", "Persona"),
				slog.String("speaker_id", speakerID),
				slog.String("content", personaText),
			)
			failCount++
			continue
		}

		// Prepare PersonaResult from metadata if available
		// (In a full implementation, we'd ensure all needed fields are in metadata)
		npcName, _ := resp.Metadata["npc_name"].(string)
		race, _ := resp.Metadata["race"].(string)
		editorID, _ := resp.Metadata["editor_id"].(string)

		result := PersonaResult{
			SpeakerID:   speakerID,
			EditorID:    editorID,
			NPCName:     npcName,
			Race:        race,
			PersonaText: personaText,
			Status:      "success",
		}

		if err := g.Store.SavePersona(ctx, result); err != nil {
			slog.ErrorContext(ctx, "Failed to save persona to store",
				slog.String("slice", "Persona"),
				slog.String("speaker_id", speakerID),
				slog.String("error", err.Error()),
			)
			failCount++
			continue
		}

		successCount++
		slog.InfoContext(ctx, "Successfully saved persona",
			slog.String("slice", "Persona"),
			slog.String("speaker_id", speakerID),
		)
	}

	slog.DebugContext(ctx, "EXIT SaveResults",
		slog.String("slice", "Persona"),
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
