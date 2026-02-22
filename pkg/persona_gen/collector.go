package persona_gen

import (
	"context"
	"log/slog"
	"time"
)

// DefaultDialogueCollector implements DialogueCollector to group
// dialogues by NPC for persona generation.
type DefaultDialogueCollector struct{}

// NewDefaultDialogueCollector creates a new DialogueCollector.
func NewDefaultDialogueCollector() *DefaultDialogueCollector {
	return &DefaultDialogueCollector{}
}

// CollectByNPC extracts and groups dialogue data by SpeakerID from the raw input data.
func (c *DefaultDialogueCollector) CollectByNPC(ctx context.Context, data PersonaGenInput) ([]NPCDialogueData, error) {
	slog.DebugContext(ctx, "ENTER CollectByNPC",
		slog.String("slice", "PersonaGen"),
		slog.Int("total_dialogues_input", len(data.Dialogues)),
	)
	start := time.Now()

	// 1. Group dialogues by SpeakerID
	speakerDialogues := make(map[string][]DialogueEntry)
	for _, raw := range data.Dialogues {
		// We need a speaker ID to associate dialogue with an NPC
		if raw.SpeakerID == nil || *raw.SpeakerID == "" {
			continue
		}

		speakerID := *raw.SpeakerID

		// We need text to analyze
		if raw.Text == nil || *raw.Text == "" {
			continue
		}

		text := *raw.Text

		// Map the raw dialogue to our internal DTO
		entry := DialogueEntry{
			Text:             text,
			EnglishText:      text, // For now, assuming raw text is English for NLP parsing if English. We might need a separate field if translation source is different.
			IsServicesBranch: raw.IsServicesBranch,
			Order:            raw.Order,
		}
		if raw.QuestID != nil {
			entry.QuestID = *raw.QuestID
		}

		speakerDialogues[speakerID] = append(speakerDialogues[speakerID], entry)
	}

	// 2. Build the result slice
	var result []NPCDialogueData
	for speakerID, dialogues := range speakerDialogues {
		npcData := NPCDialogueData{
			SpeakerID: speakerID,
			Dialogues: dialogues,
		}

		// Enrich with NPC metadata if available
		if npc, found := data.NPCs[speakerID]; found {
			npcData.NPCName = npc.Name
			npcData.Race = npc.Type // Depending on how Type is mapped in the extraction Phase 1
			// Sex and VoiceType might not be directly in PersonaNPC currently,
			// so we leave them empty or map if they become available
			if npc.EditorID != nil {
				npcData.EditorID = *npc.EditorID
			}
		}

		result = append(result, npcData)
	}

	slog.DebugContext(ctx, "EXIT CollectByNPC",
		slog.String("slice", "PersonaGen"),
		slog.Int("unique_speakers", len(result)),
		slog.Duration("elapsed", time.Since(start)),
	)

	return result, nil
}
