package persona

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/telemetry"
)

// DefaultPersonaGenerator implements the NPCPersonaGenerator interface.
type DefaultPersonaGenerator struct {
	Collector   DialogueCollector
	Evaluator   ContextEvaluator
	Store       PersonaStore
	Config      ConfigStore
	SecretStore SecretStore
}

// NewPersonaGenerator creates a new NPCPersonaGenerator.
func NewPersonaGenerator(
	collector DialogueCollector,
	evaluator ContextEvaluator,
	store PersonaStore,
	configStore ConfigStore,
	secretStore SecretStore,
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
) ([]llmio.Request, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionProcessTranslation)() // Persona generation is part of translation process
	data, ok := input.(PersonaGenInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Persona slice: %T", input)
	}

	slog.DebugContext(ctx, "starting persona prompt preparation",
		slog.Int("npc_count", len(data.NPCs)),
	)

	// Note: default config for now. In a real scenario, this might be injected or part of input.
	config := PersonaConfig{
		MinDialogueThreshold: 1,
		ContextWindowLimit:   4000,
		MaxOutputTokens:      500,
	}

	promptCfg, promptErr := loadPromptConfig(ctx, g.Config)
	if promptErr != nil {
		slog.WarnContext(ctx, "failed to load master persona prompt config, using defaults",
			telemetry2.ErrorAttrs(promptErr)...,
		)
	}

	groupedData, err := g.Collector.CollectByNPC(ctx, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to collect NPC dialogues", telemetry2.ErrorAttrs(err)...)
		return nil, fmt.Errorf("failed to collect NPC dialogues: %w", err)
	}

	var requests []llmio.Request
	skippedCount := 0
	thresholdCount := 0

	for _, npcData := range groupedData {
		npcData.SourcePlugin = normalizeSourcePlugin(npcData.SourcePlugin, npcData.SourceHint)

		// Persist base NPC metadata and dialogues before request generation.
		saveState, err := g.Store.SavePersonaBase(ctx, npcData, data.OverwriteExisting)
		if err != nil {
			slog.WarnContext(ctx, "failed to save persona base data",
				slog.String("speaker_id", npcData.SpeakerID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if strings.TrimSpace(saveState.PersonaText) == "" || data.OverwriteExisting {
			if err := g.Store.ReplaceDialogues(ctx, saveState.PersonaID, npcData.SourcePlugin, npcData.SpeakerID, npcData.Dialogues); err != nil {
				slog.WarnContext(ctx, "failed to save persona dialogues",
					slog.String("speaker_id", npcData.SpeakerID),
					slog.String("error", err.Error()),
				)
			}
		}

		if strings.TrimSpace(saveState.PersonaText) != "" && !data.OverwriteExisting {
			skippedCount++
			continue
		}

		if len(npcData.Dialogues) < config.MinDialogueThreshold {
			thresholdCount++
			continue
		}

		estimation, selectedDialogues := g.Evaluator.Evaluate(ctx, npcData, config)
		if len(selectedDialogues) == 0 {
			slog.DebugContext(ctx, "no dialogues selected for persona generation",
				slog.String("speaker_id", npcData.SpeakerID),
				slog.Any("estimation", estimation),
			)
			continue
		}

		request := llmio.Request{
			SystemPrompt: promptCfg.SystemPrompt,
			UserPrompt:   buildPersonaUserPrompt(promptCfg, npcData, selectedDialogues),
			Temperature:  0.3,
			Metadata: map[string]interface{}{
				"speaker_id":         npcData.SpeakerID,
				"npc_name":           npcData.NPCName,
				"race":               npcData.Race,
				"sex":                npcData.Sex,
				"voice_type":         npcData.VoiceType,
				"source_plugin":      npcData.SourcePlugin,
				"editor_id":          npcData.EditorID,
				"overwrite_existing": data.OverwriteExisting,
			},
		}
		if err := g.Store.SaveGenerationRequest(ctx, npcData.SourcePlugin, npcData.SpeakerID, formatGenerationRequest(request)); err != nil {
			slog.WarnContext(ctx, "failed to save persona generation request",
				slog.String("speaker_id", npcData.SpeakerID),
				slog.String("error", err.Error()),
			)
		}
		requests = append(requests, request)

		slog.DebugContext(ctx, "persona request prepared",
			slog.String("speaker_id", npcData.SpeakerID),
			slog.String("npc_name", npcData.NPCName),
			slog.String("system_prompt", request.SystemPrompt),
			slog.String("user_prompt", request.UserPrompt),
			slog.Float64("temperature", float64(request.Temperature)),
		)
	}

	slog.InfoContext(ctx, "persona prompt preparation completed",
		slog.Int("request_count", len(requests)),
		slog.Int("skipped_already_exists", skippedCount),
		slog.Int("skipped_below_threshold", thresholdCount),
	)

	return requests, nil
}

func formatGenerationRequest(request llmio.Request) string {
	return strings.TrimSpace(fmt.Sprintf("System Prompt:\n%s\n\nUser Prompt:\n%s", request.SystemPrompt, request.UserPrompt))
}

// SaveResults parses LLM responses and persists them to the store (Phase 2).
func (g *DefaultPersonaGenerator) SaveResults(
	ctx context.Context,
	results []llmio.Response,
) error {
	_, err := g.SaveResultsWithSummary(ctx, results)
	if err != nil {
		return fmt.Errorf("save persona results: %w", err)
	}
	return nil
}

// SaveResultsWithSummary parses responses, persists valid personas, and returns save counts.
func (g *DefaultPersonaGenerator) SaveResultsWithSummary(
	ctx context.Context,
	results []llmio.Response,
) (SaveResultsSummary, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionProcessTranslation)()
	slog.DebugContext(ctx, "saving persona results",
		slog.Int("response_count", len(results)),
	)

	successCount := 0
	failCount := 0

	for _, resp := range results {
		// Identify NPC from metadata
		speakerID, ok := resp.Metadata["speaker_id"].(string)
		if !ok || speakerID == "" {
			slog.WarnContext(ctx, "missing speaker_id in response metadata")
			failCount++
			continue
		}

		if !resp.Success {
			slog.WarnContext(ctx, "persona generation failed in LLM",
				slog.String("speaker_id", speakerID),
				slog.String("error", resp.Error),
			)
			failCount++
			continue
		}

		personaText := g.extractPersona(resp.Content)
		if personaText == "" {
			slog.WarnContext(ctx, "failed to extract persona text from response",
				slog.String("speaker_id", speakerID),
			)
			failCount++
			continue
		}

		if len(personaText) <= 5 {
			slog.WarnContext(ctx, "extracted persona text is too short",
				slog.String("speaker_id", speakerID),
				slog.String("content", personaText),
			)
			failCount++
			continue
		}

		// Prepare PersonaResult from metadata if available
		npcName, _ := resp.Metadata["npc_name"].(string)
		race, _ := resp.Metadata["race"].(string)
		sex, _ := resp.Metadata["sex"].(string)
		voiceType, _ := resp.Metadata["voice_type"].(string)
		sourcePlugin, _ := resp.Metadata["source_plugin"].(string)
		editorID, _ := resp.Metadata["editor_id"].(string)
		overwriteExisting, _ := resp.Metadata["overwrite_existing"].(bool)

		result := PersonaResult{
			SpeakerID:    speakerID,
			EditorID:     editorID,
			NPCName:      npcName,
			Race:         race,
			Sex:          sex,
			VoiceType:    voiceType,
			PersonaText:  personaText,
			SourcePlugin: sourcePlugin,
		}

		if err := g.Store.SavePersona(ctx, result, overwriteExisting); err != nil {
			slog.ErrorContext(ctx, "failed to save persona to store",
				append(telemetry2.ErrorAttrs(err), slog.String("speaker_id", speakerID))...)
			failCount++
			continue
		}

		successCount++
		slog.InfoContext(ctx, "persona saved successfully",
			slog.String("speaker_id", speakerID),
			slog.String("npc_name", npcName),
		)
	}

	slog.InfoContext(ctx, "persona results saving completed",
		slog.Int("success_count", successCount),
		slog.Int("fail_count", failCount),
	)
	return SaveResultsSummary{
		SuccessCount: successCount,
		FailCount:    failCount,
	}, nil
}

// CleanupTaskArtifacts removes task-scoped intermediate persona artifacts.
func (g *DefaultPersonaGenerator) CleanupTaskArtifacts(ctx context.Context, taskID string) error {
	if strings.TrimSpace(taskID) == "" {
		return nil
	}
	if err := g.Store.CleanupTaskArtifacts(ctx, taskID); err != nil {
		return fmt.Errorf("cleanup persona task artifacts task_id=%s: %w", taskID, err)
	}
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
