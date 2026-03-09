package pipeline

import (
	"github.com/ishibata91/ai-translation-engine-2/pkg/persona"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/parser"
	translator2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/translator/translator"
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
			ID:           npc.ID,
			EditorID:     npc.EditorID,
			Type:         npc.Type,
			Name:         npc.Name,
			Race:         npc.Race,
			Sex:          npc.Sex,
			VoiceType:    npc.Voice,
			SourcePlugin: derefString(npc.Source),
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			editorID := resp.EditorID
			if editorID == nil || *editorID == "" {
				editorID = group.EditorID
			}
			sourcePlugin := resp.Source
			if sourcePlugin == nil || *sourcePlugin == "" {
				sourcePlugin = group.Source
			}
			input.Dialogues = append(input.Dialogues, persona.PersonaDialogue{
				ID:               resp.ID,
				EditorID:         editorID,
				GroupEditorID:    group.EditorID,
				Type:             resp.Type,
				SpeakerID:        resp.SpeakerID,
				Text:             &resp.Text,
				QuestID:          group.QuestID,
				SourcePlugin:     sourcePlugin,
				IsServicesBranch: group.IsServicesBranch,
				Order:            resp.Order,
			})
		}
	}

	return input
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// ToTranslatorInput maps ParserOutput to TranslatorInput.
func ToTranslatorInput(out *parser.ParserOutput) translator2.TranslatorInput {
	input := translator2.TranslatorInput{
		GameData: translator2.ContextEngineInput{
			NPCs: make(map[string]translator2.ContextNPC),
		},
	}

	for id, npc := range out.NPCs {
		input.GameData.NPCs[id] = translator2.ContextNPC{
			ID:       npc.ID,
			EditorID: npc.EditorID,
			Type:     npc.Type,
			Name:     npc.Name,
		}
	}

	for _, group := range out.DialogueGroups {
		for _, resp := range group.Responses {
			input.GameData.Dialogues = append(input.GameData.Dialogues, translator2.ContextDialogue{
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
		cq := translator2.ContextQuest{
			ID:       q.ID,
			EditorID: q.EditorID,
			Type:     q.Type,
			Name:     q.Name,
		}

		for _, s := range q.Stages {
			cq.Stages = append(cq.Stages, translator2.ContextQuestStage{
				StageIndex:     s.StageIndex,
				LogIndex:       s.LogIndex,
				Type:           s.Type,
				Text:           s.Text,
				ParentID:       s.ParentID,
				ParentEditorID: s.ParentEditorID,
			})
		}

		for _, o := range q.Objectives {
			cq.Objectives = append(cq.Objectives, translator2.ContextQuestObjective{
				Index:          o.Index,
				Type:           o.Type,
				Text:           o.Text,
				ParentID:       o.ParentID,
				ParentEditorID: o.ParentEditorID,
			})
		}
		input.GameData.Quests = append(input.GameData.Quests, cq)
	}

	for _, item := range out.Items {
		input.GameData.Items = append(input.GameData.Items, translator2.ContextItem{
			ID:       item.ID,
			EditorID: item.EditorID,
			Type:     item.Type,
			Name:     item.Name,
			Text:     item.Text,
		})
	}

	for _, magic := range out.Magic {
		input.GameData.Magic = append(input.GameData.Magic, translator2.ContextMagic{
			ID:       magic.ID,
			EditorID: magic.EditorID,
			Type:     magic.Type,
			Name:     magic.Name,
		})
	}

	for _, loc := range out.Locations {
		input.GameData.Locations = append(input.GameData.Locations, translator2.ContextLocation{
			ID:       loc.ID,
			EditorID: loc.EditorID,
			Type:     loc.Type,
			Name:     loc.Name,
		})
	}

	return input
}
