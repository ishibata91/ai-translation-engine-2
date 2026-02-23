package pipeline

import (
	"github.com/ishibata91/ai-translation-engine-2/pkg/lore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/parser"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/terminology"
)

// ToTermTranslatorInput maps ParserOutput to TerminologyInput.
func ToTermTranslatorInput(out *parser.ParserOutput) terminology.TerminologyInput {
	input := terminology.TerminologyInput{
		NPCs:      make(map[string]terminology.TermNPC),
		Items:     make([]terminology.TermItem, len(out.Items)),
		Magic:     make([]terminology.TermMagic, len(out.Magic)),
		Locations: make([]terminology.TermLocation, len(out.Locations)),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = terminology.TermNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for i, item := range out.Items {
		input.Items[i] = terminology.TermItem{
			ID:       item.ID,
			EditorID: item.EditorID,
			Type:     item.Type,
			Name:     item.Name,
			Text:     item.Text,
		}
	}

	for i, magic := range out.Magic {
		input.Magic[i] = terminology.TermMagic{
			ID:       magic.ID,
			EditorID: magic.EditorID,
			Type:     magic.Type,
			Name:     magic.Name,
		}
	}

	for i, loc := range out.Locations {
		input.Locations[i] = terminology.TermLocation{
			ID:       loc.ID,
			EditorID: loc.EditorID,
			Type:     loc.Type,
			Name:     loc.Name,
		}
	}

	return input
}

// ToPersonaGenInput maps ParserOutput to PersonaGenInput.
func ToPersonaGenInput(out *parser.ParserOutput) persona.PersonaGenInput {
	input := persona.PersonaGenInput{
		NPCs: make(map[string]persona.PersonaNPC),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = persona.PersonaNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			input.Dialogues = append(input.Dialogues, persona.PersonaDialogue{
				ID:               resp.ID,
				EditorID:         resp.EditorID,
				Type:             resp.Type,
				SpeakerID:        resp.SpeakerID,
				Text:             &resp.Text,
				QuestID:          group.QuestID,
				IsServicesBranch: group.IsServicesBranch,
				Order:            resp.Order,
			})
		}
	}

	return input
}

// ToContextEngineInput maps ParserOutput to LoreInput.
func ToContextEngineInput(out *parser.ParserOutput) lore.LoreInput {
	input := lore.LoreInput{
		NPCs: make(map[string]lore.ContextNPC),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = lore.ContextNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			input.Dialogues = append(input.Dialogues, lore.ContextDialogue{
				ID:               resp.ID,
				EditorID:         resp.EditorID,
				Type:             resp.Type,
				SpeakerID:        resp.SpeakerID,
				Text:             &resp.Text,
				QuestID:          group.QuestID,
				IsServicesBranch: group.IsServicesBranch,
				Order:            resp.Order,
			})
		}
	}

	for _, q := range out.Quests {
		cq := lore.ContextQuest{
			ID:       q.ID,
			EditorID: q.EditorID,
			Type:     q.Type,
			Name:     q.Name,
		}

		for _, s := range q.Stages {
			cq.Stages = append(cq.Stages, lore.ContextQuestStage{
				StageIndex:     s.StageIndex,
				LogIndex:       s.LogIndex,
				Type:           s.Type,
				Text:           s.Text,
				ParentID:       s.ParentID,
				ParentEditorID: s.ParentEditorID,
			})
		}

		for _, o := range q.Objectives {
			cq.Objectives = append(cq.Objectives, lore.ContextQuestObjective{
				Index:          o.Index,
				Type:           o.Type,
				Text:           o.Text,
				ParentID:       o.ParentID,
				ParentEditorID: o.ParentEditorID,
			})
		}
		input.Quests = append(input.Quests, cq)
	}

	for _, item := range out.Items {
		input.Items = append(input.Items, lore.ContextItem{
			ID:       item.ID,
			EditorID: item.EditorID,
			Type:     item.Type,
			Name:     item.Name,
			Text:     item.Text,
		})
	}

	for _, magic := range out.Magic {
		input.Magic = append(input.Magic, lore.ContextMagic{
			ID:       magic.ID,
			EditorID: magic.EditorID,
			Type:     magic.Type,
			Name:     magic.Name,
		})
	}

	for _, loc := range out.Locations {
		input.Locations = append(input.Locations, lore.ContextLocation{
			ID:       loc.ID,
			EditorID: loc.EditorID,
			Type:     loc.Type,
			Name:     loc.Name,
		})
	}

	return input
}
