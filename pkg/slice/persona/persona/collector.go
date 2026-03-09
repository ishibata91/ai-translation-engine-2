package persona

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
		slog.String("slice", "Persona"),
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
			RecordType:       raw.Type,
			Text:             text,
			EnglishText:      text, // For now, assuming raw text is English for NLP parsing if English. We might need a separate field if translation source is different.
			IsServicesBranch: raw.IsServicesBranch,
			Order:            raw.Order,
		}
		if raw.EditorID != nil {
			entry.EditorID = *raw.EditorID
		} else if raw.GroupEditorID != nil {
			entry.EditorID = *raw.GroupEditorID
		}
		if raw.QuestID != nil {
			entry.QuestID = *raw.QuestID
		}
		if raw.SourcePlugin != nil {
			entry.SourcePlugin = *raw.SourcePlugin
		}

		speakerDialogues[speakerID] = append(speakerDialogues[speakerID], entry)
	}

	// 2. Build the result slice
	var result []NPCDialogueData
	for speakerID, dialogues := range speakerDialogues {
		npcData := NPCDialogueData{
			SpeakerID:  speakerID,
			Dialogues:  dialogues,
			SourceHint: data.SourceJSONPath,
		}

		// Enrich with NPC metadata if available
		if npc, found := data.NPCs[speakerID]; found {
			npcData.NPCName = npc.Name
			npcData.Race = npc.Race
			npcData.Sex = npc.Sex
			npcData.VoiceType = npc.VoiceType
			npcData.SourcePlugin = npc.SourcePlugin
			if npc.EditorID != nil {
				npcData.EditorID = *npc.EditorID
			}
		}
		if npcData.SourcePlugin == "" && len(dialogues) > 0 {
			npcData.SourcePlugin = dialogues[0].SourcePlugin
		}
		if npcData.EditorID == "" && len(dialogues) > 0 {
			npcData.EditorID = dialogues[0].EditorID
		}

		result = append(result, npcData)
	}

	slog.DebugContext(ctx, "EXIT CollectByNPC",
		slog.String("slice", "Persona"),
		slog.Int("unique_speakers", len(result)),
		slog.Duration("elapsed", time.Since(start)),
	)

	return result, nil
}
