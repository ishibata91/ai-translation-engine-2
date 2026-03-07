export type MasterPersonaProvider = 'lmstudio' | 'gemini' | 'openai' | 'xai';

export interface MasterPersonaLLMConfig {
    provider: MasterPersonaProvider;
    model: string;
    endpoint: string;
    apiKey: string;
    temperature: number;
    contextLength: number;
    syncConcurrency: number;
}

export const DEFAULT_MASTER_PERSONA_LLM_CONFIG: MasterPersonaLLMConfig = {
    provider: 'lmstudio',
    model: '',
    endpoint: 'http://localhost:1234',
    apiKey: '',
    temperature: 0.3,
    contextLength: 0,
    syncConcurrency: 1,
};

export interface MasterPersonaPromptConfig {
    userPrompt: string;
    systemPrompt: string;
}

export const DEFAULT_MASTER_PERSONA_PROMPT_CONFIG: MasterPersonaPromptConfig = {
    userPrompt: '以下の会話履歴をもとに、NPC の性格、話し方、背景や関係性を簡潔に要約してください。重要な特徴を優先し、推測しすぎないでください。',
    systemPrompt: `You are a character persona analyzer for RPG dialogue.
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

Keep the total response under 150 words and do not add extra conversational filler.`,
};
