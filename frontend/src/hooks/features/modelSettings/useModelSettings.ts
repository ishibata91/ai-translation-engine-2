import { useEffect, useMemo, useState } from 'react';
import { z } from 'zod';
import { ListModels } from '../../../wailsjs/go/modelcatalog/ModelCatalogService';
import { type MasterPersonaLLMConfig, type MasterPersonaProvider } from '../../../types/masterPersona';

type ModelOptionItem = { id: string; label: string };

export const MASTER_PERSONA_PROVIDERS = ['lmstudio', 'gemini', 'openai', 'xai'] as const;

const FALLBACK_MODEL_OPTIONS: Record<MasterPersonaProvider, ModelOptionItem[]> = {
    lmstudio: [{ id: '(model-unavailable)', label: '(モデルを取得できませんでした)' }],
    gemini: [
        { id: 'gemini-2.0-flash', label: 'gemini-2.0-flash' },
        { id: 'gemini-2.0-pro', label: 'gemini-2.0-pro' },
        { id: 'gemini-1.5-pro', label: 'gemini-1.5-pro' },
        { id: 'gemini-1.5-flash', label: 'gemini-1.5-flash' },
    ],
    openai: [
        { id: 'gpt-4o', label: 'gpt-4o' },
        { id: 'gpt-4o-mini', label: 'gpt-4o-mini' },
        { id: 'gpt-4-turbo', label: 'gpt-4-turbo' },
        { id: 'gpt-3.5-turbo', label: 'gpt-3.5-turbo' },
    ],
    xai: [
        { id: 'grok-3', label: 'grok-3' },
        { id: 'grok-3-mini', label: 'grok-3-mini' },
        { id: 'grok-2', label: 'grok-2' },
    ],
};

export const PROVIDER_LABELS: Record<MasterPersonaProvider, string> = {
    lmstudio: 'Local LLM (LM Studio)',
    gemini: 'Google Gemini',
    openai: 'OpenAI (GPT)',
    xai: 'xAI (Grok)',
};

const providerSchema = z.enum(MASTER_PERSONA_PROVIDERS);
const modelOptionSchema = z.object({
    id: z.string().catch(''),
    display_name: z.string().optional().catch(''),
});
const modelOptionListSchema = z.array(modelOptionSchema);

interface UseModelSettingsArgs {
    value: MasterPersonaLLMConfig;
    onChange: (next: MasterPersonaLLMConfig) => void;
    enabled: boolean;
    namespace: string;
}

/**
 * モデル一覧取得と provider 切替ロジックを UI から分離する。
 */
export function useModelSettings({ value, onChange, enabled, namespace }: UseModelSettingsArgs) {
    const [dynamicOptionsByProvider, setDynamicOptionsByProvider] = useState<Partial<Record<MasterPersonaProvider, ModelOptionItem[]>>>({});
    const [catalogLoading, setCatalogLoading] = useState<boolean>(false);
    const [catalogError, setCatalogError] = useState<string>('');

    const provider = value.provider;
    const endpoint = value.endpoint;
    const apiKey = value.apiKey;
    const currentModel = value.model;
    const isLMStudio = provider === 'lmstudio';

    const modelOptions = useMemo(() => {
        const dynamic = dynamicOptionsByProvider[provider] ?? [];
        if (dynamic.length > 0) {
            return dynamic;
        }
        return FALLBACK_MODEL_OPTIONS[provider];
    }, [dynamicOptionsByProvider, provider]);

    const selectableModelOptions = useMemo(() => {
        if (modelOptions.length > 0) {
            return modelOptions;
        }
        if (currentModel.trim() !== '') {
            return [{ id: currentModel, label: currentModel }];
        }
        return [{ id: '(model-unavailable)', label: '(モデルを取得できませんでした)' }];
    }, [currentModel, modelOptions]);

    const selectedModelValue = useMemo(() => {
        if (selectableModelOptions.some((option) => option.id === currentModel)) {
            return currentModel;
        }
        const byLabel = selectableModelOptions.find((option) => option.label === currentModel);
        if (byLabel) {
            return byLabel.id;
        }
        return selectableModelOptions[0].id;
    }, [currentModel, selectableModelOptions]);

    useEffect(() => {
        if (!enabled) {
            return;
        }

        let alive = true;
        void (async () => {
            setCatalogLoading(true);
            setCatalogError('');

            try {
                const rowsRaw: unknown = await ListModels({
                    namespace,
                    provider,
                    endpoint,
                    apiKey: isLMStudio ? '' : apiKey,
                });
                const rows = modelOptionListSchema.parse(rowsRaw);
                if (!alive) {
                    return;
                }

                const seen = new Set<string>();
                const options: ModelOptionItem[] = [];
                for (const row of rows) {
                    const id = row.id.trim();
                    const label = (row.display_name ?? row.id).trim();
                    if (id.length === 0 || seen.has(id)) {
                        continue;
                    }
                    seen.add(id);
                    options.push({ id, label: label.length > 0 ? label : id });
                }

                setDynamicOptionsByProvider((prev) => ({ ...prev, [provider]: options }));
                const hasCurrent = options.some((option) => option.id === currentModel || option.label === currentModel);
                if (options.length > 0 && !hasCurrent) {
                    onChange({ ...value, model: options[0].id });
                }
            } catch {
                if (!alive) {
                    return;
                }
                setDynamicOptionsByProvider((prev) => ({ ...prev, [provider]: [] }));
                setCatalogError('モデル一覧の取得に失敗しました');
            } finally {
                if (alive) {
                    setCatalogLoading(false);
                }
            }
        })();

        return () => {
            alive = false;
        };
    }, [apiKey, currentModel, enabled, endpoint, isLMStudio, namespace, onChange, provider, value]);

    const handleProviderChange = (rawProvider: string) => {
        const parsed = providerSchema.safeParse(rawProvider);
        if (!parsed.success) {
            return;
        }

        const nextProvider = parsed.data;
        const candidates = FALLBACK_MODEL_OPTIONS[nextProvider];
        const nextModel = nextProvider === provider ? value.model : (candidates[0]?.id ?? '');
        onChange({
            ...value,
            provider: nextProvider,
            model: nextModel,
            endpoint: nextProvider === 'lmstudio' ? value.endpoint || 'http://localhost:1234' : value.endpoint,
        });
    };

    return {
        catalogError,
        catalogLoading,
        handleProviderChange,
        selectableModelOptions,
        selectedModelValue,
    };
}
