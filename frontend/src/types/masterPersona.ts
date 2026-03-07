export type MasterPersonaProvider = 'lmstudio' | 'gemini' | 'openai' | 'xai';

export interface MasterPersonaLLMConfig {
    provider: MasterPersonaProvider;
    model: string;
    endpoint: string;
    apiKey: string;
    temperature: number;
    maxTokens: number;
}

export const DEFAULT_MASTER_PERSONA_LLM_CONFIG: MasterPersonaLLMConfig = {
    provider: 'lmstudio',
    model: '',
    endpoint: 'http://localhost:1234',
    apiKey: '',
    temperature: 0.3,
    maxTokens: 500,
};
