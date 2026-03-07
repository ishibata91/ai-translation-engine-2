package persona

import (
	"context"
	"fmt"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
)

type promptConfig struct {
	UserPrompt   string
	SystemPrompt string
}

func defaultPromptConfig() promptConfig {
	return promptConfig{
		UserPrompt:   config.DefaultMasterPersonaUserPrompt,
		SystemPrompt: config.DefaultMasterPersonaSystemPrompt,
	}
}

func loadPromptConfig(ctx context.Context, store config.Config) (promptConfig, error) {
	defaults := defaultPromptConfig()
	if store == nil {
		return defaults, nil
	}

	values, err := store.GetAll(ctx, config.MasterPersonaPromptNamespace)
	if err != nil {
		return defaults, err
	}
	merged := config.MergeMasterPersonaPromptDefaults(values)
	return promptConfig{
		UserPrompt:   strings.TrimSpace(merged[config.MasterPersonaUserPromptKey]),
		SystemPrompt: strings.TrimSpace(merged[config.MasterPersonaSystemPromptKey]),
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
