import {useEffect, useMemo, useState} from 'react';
import {z} from 'zod';
import {ListModels} from '../../../wailsjs/go/controller/ModelCatalogController';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    type MasterPersonaExecutionProfile,
    type MasterPersonaLLMConfig,
    type MasterPersonaModelCapability,
    type MasterPersonaModelOption,
    type MasterPersonaProvider,
} from '../../../types/masterPersona';

export const MASTER_PERSONA_PROVIDERS = ['lmstudio', 'gemini', 'xai'] as const;

const FALLBACK_MODEL_OPTIONS: Record<MasterPersonaProvider, MasterPersonaModelOption[]> = {
    lmstudio: [{ id: '(model-unavailable)', label: '(モデルを取得できませんでした)', capability: { supportsBatch: false } }],
    gemini: [
        { id: 'gemini-2.0-flash', label: 'gemini-2.0-flash', capability: { supportsBatch: true } },
        { id: 'gemini-2.0-pro', label: 'gemini-2.0-pro', capability: { supportsBatch: true } },
        { id: 'gemini-1.5-pro', label: 'gemini-1.5-pro', capability: { supportsBatch: true } },
        { id: 'gemini-1.5-flash', label: 'gemini-1.5-flash', capability: { supportsBatch: true } },
    ],
    xai: [
        { id: 'grok-3', label: 'grok-3', capability: { supportsBatch: true } },
        { id: 'grok-3-mini', label: 'grok-3-mini', capability: { supportsBatch: true } },
        { id: 'grok-2', label: 'grok-2', capability: { supportsBatch: true } },
    ],
};

export const PROVIDER_LABELS: Record<MasterPersonaProvider, string> = {
    lmstudio: 'Local LLM (LM Studio)',
    gemini: 'Google Gemini',
    xai: 'xAI (Grok)',
};

const providerSchema = z.enum(MASTER_PERSONA_PROVIDERS);
const modelCapabilitySchema = z
    .object({
        supports_batch: z.boolean().optional().catch(false),
    })
    .transform((raw) => ({
        supports_batch: raw.supports_batch ?? false,
    }));

const modelOptionSchema = z.object({
    id: z.string().catch(''),
    display_name: z.string().optional().catch(''),
    capability: modelCapabilitySchema.optional(),
});

const modelOptionListSchema = z.array(modelOptionSchema);

const DEFAULT_MODEL_CAPABILITY: MasterPersonaModelCapability = { supportsBatch: false };
const DEFAULT_CLOUD_MODEL_CAPABILITY: MasterPersonaModelCapability = { supportsBatch: true };

interface UseModelSettingsArgs {
    value: MasterPersonaLLMConfig;
    onChange: (next: MasterPersonaLLMConfig) => void;
    enabled: boolean;
    namespace: string;
}

const normalizeProvider = (provider: MasterPersonaLLMConfig['provider']): MasterPersonaProvider => {
    if (provider === 'lmstudio' || provider === 'gemini' || provider === 'xai') {
        return provider;
    }
    return DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
};

const normalizeExecutionProfiles = (capability: MasterPersonaModelCapability): MasterPersonaExecutionProfile[] => {
    if (capability.supportsBatch) {
        return ['sync', 'batch'];
    }
    return ['sync'];
};

const resolveNextModel = (
    provider: MasterPersonaProvider,
    currentModel: string,
    dynamicOptionsByProvider: Partial<Record<MasterPersonaProvider, MasterPersonaModelOption[]>>,
): string => {
    const candidates = dynamicOptionsByProvider[provider] ?? FALLBACK_MODEL_OPTIONS[provider];
    if (candidates.some((option) => option.id === currentModel)) {
        return currentModel;
    }
    return candidates[0]?.id ?? '';
};

const resolveModelCapability = (
    provider: MasterPersonaProvider,
    modelID: string,
    dynamicOptionsByProvider: Partial<Record<MasterPersonaProvider, MasterPersonaModelOption[]>>,
): MasterPersonaModelCapability => {
    const options = dynamicOptionsByProvider[provider] ?? FALLBACK_MODEL_OPTIONS[provider];
    const byID = options.find((option) => option.id === modelID);
    if (byID) {
        return byID.capability;
    }
    if (provider === 'lmstudio') {
        return DEFAULT_MODEL_CAPABILITY;
    }
    // Cloud provider capability may be unknown when catalog fetch fails.
    return DEFAULT_CLOUD_MODEL_CAPABILITY;
};

/**
 * モデル一覧取得と provider 切替ロジックを UI から分離する。
 */
export function useModelSettings({ value, onChange, enabled, namespace }: UseModelSettingsArgs) {
    const [dynamicOptionsByProvider, setDynamicOptionsByProvider] = useState<Partial<Record<MasterPersonaProvider, MasterPersonaModelOption[]>>>({});
    const [catalogLoading, setCatalogLoading] = useState<boolean>(false);
    const [catalogError, setCatalogError] = useState<string>('');
    const [lastFetchedProvider, setLastFetchedProvider] = useState<MasterPersonaProvider | null>(null);
    const [lastFetchedApiKey, setLastFetchedApiKey] = useState<string>('');

    const provider = normalizeProvider(value.provider);
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
            return [{
                id: currentModel,
                label: currentModel,
                capability: resolveModelCapability(provider, currentModel, dynamicOptionsByProvider),
            }];
        }
        return [{
            id: '(model-unavailable)',
            label: '(モデルを取得できませんでした)',
            capability: provider === 'lmstudio' ? DEFAULT_MODEL_CAPABILITY : DEFAULT_CLOUD_MODEL_CAPABILITY,
        }];
    }, [currentModel, dynamicOptionsByProvider, modelOptions, provider]);

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

    const selectedModelCapability = useMemo(() => {
        const selected = selectableModelOptions.find((option) => option.id === selectedModelValue);
        if (selected) {
            return selected.capability;
        }
        return DEFAULT_MODEL_CAPABILITY;
    }, [selectableModelOptions, selectedModelValue]);

    const availableExecutionProfiles = useMemo(
        () => normalizeExecutionProfiles(selectedModelCapability),
        [selectedModelCapability],
    );

    useEffect(() => {
        if (value.provider === provider) {
            return;
        }
        const nextModel = resolveNextModel(provider, value.model, dynamicOptionsByProvider);
        onChange({
            ...value,
            provider,
            model: nextModel,
            bulkStrategy: 'sync',
        });
    }, [dynamicOptionsByProvider, onChange, provider, value]);

    useEffect(() => {
        if (value.bulkStrategy !== 'batch') {
            return;
        }
        if (selectedModelCapability.supportsBatch) {
            return;
        }
        onChange({ ...value, bulkStrategy: 'sync' });
    }, [onChange, selectedModelCapability, value]);

    useEffect(() => {
        if (!enabled) {
            return;
        }

        const isFirstLoad = lastFetchedProvider === null;
        const providerChanged = lastFetchedProvider !== provider;
        const apiKeySet = !isLMStudio && lastFetchedApiKey === '' && apiKey !== '';

        if (!isFirstLoad && !providerChanged && !apiKeySet) {
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
                const options: MasterPersonaModelOption[] = [];
                for (const row of rows) {
                    const id = row.id.trim();
                    const label = (row.display_name ?? row.id).trim();
                    if (id.length === 0 || seen.has(id)) {
                        continue;
                    }
                    seen.add(id);
                    options.push({
                        id,
                        label: label.length > 0 ? label : id,
                        capability: {
                            supportsBatch: row.capability?.supports_batch ?? false,
                        },
                    });
                }

                setDynamicOptionsByProvider((prev) => ({ ...prev, [provider]: options }));
                setLastFetchedProvider(provider);
                setLastFetchedApiKey(apiKey);

                const hasCurrent = options.some((option) => option.id === currentModel || option.label === currentModel);
                if (options.length > 0 && !hasCurrent) {
                    const nextModel = options[0].id;
                    const nextCapability = resolveModelCapability(provider, nextModel, { ...dynamicOptionsByProvider, [provider]: options });
                    onChange({
                        ...value,
                        model: nextModel,
                        bulkStrategy: nextCapability.supportsBatch ? value.bulkStrategy : 'sync',
                    });
                }
            } catch {
                if (!alive) {
                    return;
                }
                setDynamicOptionsByProvider((prev) => ({ ...prev, [provider]: [] }));
                setCatalogError('モデル一覧の取得に失敗しました');
                setLastFetchedProvider(provider);
                setLastFetchedApiKey(apiKey);
            } finally {
                if (alive) {
                    setCatalogLoading(false);
                }
            }
        })();

        return () => {
            alive = false;
        };
    }, [
        apiKey,
        currentModel,
        dynamicOptionsByProvider,
        enabled,
        endpoint,
        isLMStudio,
        namespace,
        onChange,
        provider,
        value,
        lastFetchedProvider,
        lastFetchedApiKey,
    ]);

    const handleProviderChange = (rawProvider: string) => {
        const parsed = providerSchema.safeParse(rawProvider);
        if (!parsed.success) {
            return;
        }

        const nextProvider = parsed.data;
        const nextOptions = dynamicOptionsByProvider[nextProvider] ?? FALLBACK_MODEL_OPTIONS[nextProvider];
        const nextModel = nextProvider === provider ? value.model : (nextOptions[0]?.id ?? '');
        const nextCapability = resolveModelCapability(nextProvider, nextModel, dynamicOptionsByProvider);
        onChange({
            ...value,
            provider: nextProvider,
            model: nextModel,
            endpoint: nextProvider === 'lmstudio' ? value.endpoint || 'http://localhost:1234' : value.endpoint,
            bulkStrategy: nextCapability.supportsBatch ? value.bulkStrategy : 'sync',
        });
    };

    return {
        availableExecutionProfiles,
        catalogError,
        catalogLoading,
        handleProviderChange,
        selectableModelOptions,
        selectedModelCapability,
        selectedModelValue,
    };
}
