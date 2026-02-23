package process_manager

import (
	"github.com/ishibata91/ai-translation-engine-2/pkg/context_engine"
	"github.com/ishibata91/ai-translation-engine-2/pkg/loader_slice"
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona_gen"
	"github.com/ishibata91/ai-translation-engine-2/pkg/term_translator"
)

// ToTermTranslatorInput maps LoaderOutput to TermTranslatorInput.
func ToTermTranslatorInput(out *loader_slice.LoaderOutput) term_translator.TermTranslatorInput {
	input := term_translator.TermTranslatorInput{
		NPCs:      make(map[string]term_translator.TermNPC),
		Items:     make([]term_translator.TermItem, len(out.Items)),
		Magic:     make([]term_translator.TermMagic, len(out.Magic)),
		Locations: make([]term_translator.TermLocation, len(out.Locations)),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = term_translator.TermNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for i, item := range out.Items {
		input.Items[i] = term_translator.TermItem{
			ID:       item.ID,
			EditorID: item.EditorID,
			Type:     item.Type,
			Name:     item.Name,
			Text:     item.Text,
		}
	}

	for i, magic := range out.Magic {
		input.Magic[i] = term_translator.TermMagic{
			ID:       magic.ID,
			EditorID: magic.EditorID,
			Type:     magic.Type,
			Name:     magic.Name,
		}
	}

	for i, loc := range out.Locations {
		input.Locations[i] = term_translator.TermLocation{
			ID:       loc.ID,
			EditorID: loc.EditorID,
			Type:     loc.Type,
			Name:     loc.Name,
		}
	}

	return input
}

// ToPersonaGenInput maps LoaderOutput to PersonaGenInput.
func ToPersonaGenInput(out *loader_slice.LoaderOutput) persona_gen.PersonaGenInput {
	input := persona_gen.PersonaGenInput{
		NPCs: make(map[string]persona_gen.PersonaNPC),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = persona_gen.PersonaNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			input.Dialogues = append(input.Dialogues, persona_gen.PersonaDialogue{
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

// ToContextEngineInput maps LoaderOutput to ContextEngineInput.
func ToContextEngineInput(out *loader_slice.LoaderOutput) context_engine.ContextEngineInput {
	input := context_engine.ContextEngineInput{
		NPCs: make(map[string]context_engine.ContextNPC),
	}

	for id, npc := range out.NPCs {
		input.NPCs[id] = context_engine.ContextNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			input.Dialogues = append(input.Dialogues, context_engine.ContextDialogue{
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
		cq := context_engine.ContextQuest{
			ID:       q.ID,
			EditorID: q.EditorID,
			Type:     q.Type,
			Name:     q.Name,
		}

		for _, s := range q.Stages {
			cq.Stages = append(cq.Stages, context_engine.ContextQuestStage{
				StageIndex:     s.StageIndex,
				LogIndex:       s.LogIndex,
				Type:           s.Type,
				Text:           s.Text,
				ParentID:       s.ParentID,
				ParentEditorID: s.ParentEditorID,
			})
		}

		for _, o := range q.Objectives {
			cq.Objectives = append(cq.Objectives, context_engine.ContextQuestObjective{
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
		input.Items = append(input.Items, context_engine.ContextItem{
			ID:       item.ID,
			EditorID: item.EditorID,
			Type:     item.Type,
			Name:     item.Name,
			Text:     item.Text,
		})
	}

	for _, magic := range out.Magic {
		input.Magic = append(input.Magic, context_engine.ContextMagic{
			ID:       magic.ID,
			EditorID: magic.EditorID,
			Type:     magic.Type,
			Name:     magic.Name,
		})
	}

	for _, loc := range out.Locations {
		input.Locations = append(input.Locations, context_engine.ContextLocation{
			ID:       loc.ID,
			EditorID: loc.EditorID,
			Type:     loc.Type,
			Name:     loc.Name,
		})
	}

	return input
}
