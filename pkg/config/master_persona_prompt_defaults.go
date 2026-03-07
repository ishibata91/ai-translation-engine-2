package config

const (
	MasterPersonaPromptNamespace = "master_persona.prompt"
	MasterPersonaUserPromptKey   = "user_prompt"
	MasterPersonaSystemPromptKey = "system_prompt"
)

const (
	DefaultMasterPersonaUserPrompt   = `以下の会話履歴をもとに、NPC の性格、話し方、背景や関係性を簡潔に要約してください。重要な特徴を優先し、推測しすぎないでください。`
	DefaultMasterPersonaSystemPrompt = `You are a character persona analyzer for RPG dialogue.
The user message will contain:
- User Request: the editable analysis focus from the operator
- NPC Profile: basic metadata for one NPC
- Dialogue History: representative dialogue lines for that NPC

Use the User Request as the variable instruction, then analyze the NPC Profile and Dialogue History.
Generate a concise persona summary.
Your response MUST be formatted strictly as: TL: |...|
Inside the pipes, include these sections in plain text:
- Personality Traits: ...
- Speaking Habits: ...
- Background: ...

Keep the total response under 150 words and do not add extra conversational filler.`
)

func DefaultMasterPersonaPromptValues() map[string]string {
	return map[string]string{
		MasterPersonaUserPromptKey:   DefaultMasterPersonaUserPrompt,
		MasterPersonaSystemPromptKey: DefaultMasterPersonaSystemPrompt,
	}
}

func MergeMasterPersonaPromptDefaults(values map[string]string) map[string]string {
	merged := DefaultMasterPersonaPromptValues()
	for key, value := range values {
		merged[key] = value
	}
	return merged
}
