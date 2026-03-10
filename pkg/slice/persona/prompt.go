package persona

import (
	"context"
	"fmt"
	"strings"

	config2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

type promptConfig struct {
	UserPrompt   string
	SystemPrompt string
}

func defaultPromptConfig() promptConfig {
	return promptConfig{
		UserPrompt:   config2.DefaultMasterPersonaUserPrompt,
		SystemPrompt: config2.DefaultMasterPersonaSystemPrompt,
	}
}

func loadPromptConfig(ctx context.Context, store config2.Config) (promptConfig, error) {
	defaults := defaultPromptConfig()
	if store == nil {
		return defaults, nil
	}

	values, err := store.GetAll(ctx, config2.MasterPersonaPromptNamespace)
	if err != nil {
		return defaults, err
	}
	merged := config2.MergeMasterPersonaPromptDefaults(values)
	return promptConfig{
		UserPrompt:   strings.TrimSpace(merged[config2.MasterPersonaUserPromptKey]),
		SystemPrompt: strings.TrimSpace(merged[config2.MasterPersonaSystemPromptKey]),
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
