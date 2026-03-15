import type {FrontendTask} from '../../../types/task';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    type MasterPersonaLLMConfig,
    type MasterPersonaPromptConfig,
} from '../../../types/masterPersona';

export const MASTER_PERSONA_LLM_NAMESPACE = 'master_persona.llm';
export const MASTER_PERSONA_PROMPT_NAMESPACE = 'master_persona.prompt';
export const SELECTED_PROVIDER_KEY = 'selected_provider';
export const USER_PROMPT_KEY = 'user_prompt';
export const SYSTEM_PROMPT_KEY = 'system_prompt';

export const pickString = (value: unknown): string => {
    if (typeof value === 'string') {
        return value;
    }
    return '';
};

export const formatUpdatedAt = (raw: string): string => {
    const ts = Date.parse(raw);
    if (!Number.isFinite(ts)) {
        return '';
    }
    return new Date(ts).toLocaleString('ja-JP');
};

export const normalizeProvider = (value: string | undefined): MasterPersonaLLMConfig['provider'] => {
    if (value === 'lmstudio' || value === 'gemini' || value === 'xai') {
        return value;
    }
    return DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
};

export const providerNamespace = (provider: MasterPersonaLLMConfig['provider']): string =>
    `${MASTER_PERSONA_LLM_NAMESPACE}.${provider}`;

export const syncConcurrencyKey = (provider: MasterPersonaLLMConfig['provider']): string =>
    `sync_concurrency.${provider}`;

export const toErrorMessage = (error: unknown, fallback: string): string => {
    if (typeof error === 'string' && error.trim() !== '') {
        return error;
    }
    if (error && typeof error === 'object') {
        const message = (error as { message?: unknown }).message;
        if (typeof message === 'string' && message.trim() !== '') {
            return message;
        }
    }
    return fallback;
};

export const statusMessageFromTask = (task: FrontendTask): string => {
    switch (task.status) {
        case 'running':
            return 'リクエストを実行しています...';
        case 'paused':
        case 'cancelled':
            return '一時停止中';
        case 'request_generated':
            return 'リクエスト生成完了';
        case 'failed':
            return 'タスク実行に失敗しました';
        case 'completed':
            return '処理完了';
        default:
            return '待機中';
    }
};

export const buildProviderConfigPairs = (cfg: MasterPersonaLLMConfig): Record<string, string> => ({
    model: cfg.model,
    endpoint: cfg.endpoint,
    api_key: cfg.provider === 'lmstudio' ? '' : cfg.apiKey,
    temperature: String(cfg.temperature),
    context_length: String(cfg.contextLength),
    bulk_strategy: cfg.bulkStrategy,
});

export const buildPromptConfigPairs = (cfg: MasterPersonaPromptConfig): Record<string, string> => ({
    [USER_PROMPT_KEY]: cfg.userPrompt,
    [SYSTEM_PROMPT_KEY]: cfg.systemPrompt,
});

export const parseTaskTimestamp = (value: string | undefined): number => {
    if (!value) {
        return 0;
    }
    const t = Date.parse(value);
    return Number.isFinite(t) ? t : 0;
};
