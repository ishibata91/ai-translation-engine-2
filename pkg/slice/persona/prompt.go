package persona

import (
	"context"
	"fmt"
	"strings"
)

type promptConfig struct {
	UserPrompt   string
	SystemPrompt string
}

const (
	masterPersonaPromptNamespace = "master_persona.prompt"
	masterPersonaUserPromptKey   = "user_prompt"
	masterPersonaSystemPromptKey = "system_prompt"
)

const (
	defaultMasterPersonaUserPrompt   = `このNPCを他言語へ翻訳する際の「翻訳ガイドライン」を作成せよ。特に、一人称・二人称の選択、文末のニュアンス（敬語の度合い）、および特徴的な語彙（口癖や専門用語）を特定し、翻訳者が一貫性を保てるように分析すること`
	defaultMasterPersonaSystemPrompt = `You are a character persona analyzer for RPG dialogue.
The user message will contain:
- User Request: the editable analysis focus from the operator
- NPC Profile: basic metadata for one NPC
- Dialogue History: representative dialogue lines for that NPC

Use the User Request as the variable instruction, then analyze the NPC Profile and Dialogue History.
Generate a concise persona summary.
Your response MUST be formatted strictly as: TL: |...|
Inside the pipes, include these sections in plain text:

Keep the total response under 150 words and do not add extra conversational filler.`
)

type promptConfigReader interface {
	GetAll(ctx context.Context, namespace string) (map[string]string, error)
}

func defaultPromptConfig() promptConfig {
	return promptConfig{
		UserPrompt:   defaultMasterPersonaUserPrompt,
		SystemPrompt: defaultMasterPersonaSystemPrompt,
	}
}

func mergePromptDefaults(values map[string]string) map[string]string {
	merged := map[string]string{
		masterPersonaUserPromptKey:   defaultMasterPersonaUserPrompt,
		masterPersonaSystemPromptKey: defaultMasterPersonaSystemPrompt,
	}
	for key, value := range values {
		merged[key] = value
	}
	return merged
}

func loadPromptConfig(ctx context.Context, store promptConfigReader) (promptConfig, error) {
	defaults := defaultPromptConfig()
	if store == nil {
		return defaults, nil
	}

	values, err := store.GetAll(ctx, masterPersonaPromptNamespace)
	if err != nil {
		return defaults, err
	}
	merged := mergePromptDefaults(values)
	return promptConfig{
		UserPrompt:   strings.TrimSpace(merged[masterPersonaUserPromptKey]),
		SystemPrompt: strings.TrimSpace(merged[masterPersonaSystemPromptKey]),
	}, nil
}

func buildPersonaUserPrompt(cfg promptConfig, npcData NPCDialogueData, selectedDialogues []DialogueEntry) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(cfg.UserPrompt))
	sb.WriteString("\n\nNPC Profile:\n")
	sb.WriteString(fmt.Sprintf("- Name: %s\n", npcData.NPCName))
	sb.WriteString(fmt.Sprintf("- Race: %s\n", npcData.Race))
	sb.WriteString(fmt.Sprintf("- Voice Type: %s\n", npcData.VoiceType))
	sb.WriteString("Dialogue History:\n")
	for i, entry := range selectedDialogues {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, entry.EnglishText))
	}
	return strings.TrimSpace(sb.String())
}
