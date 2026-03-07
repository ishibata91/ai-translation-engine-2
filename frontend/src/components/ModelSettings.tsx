import React, { useEffect, useMemo, useState } from 'react';
import { type MasterPersonaLLMConfig, type MasterPersonaProvider } from '../types/masterPersona';
import { ListModels } from '../wailsjs/go/modelcatalog/ModelCatalogService';

type ModelOptionItem = { id: string; label: string };

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

const PROVIDER_LABELS: Record<MasterPersonaProvider, string> = {
    lmstudio: 'Local LLM (LM Studio)',
    gemini: 'Google Gemini',
    openai: 'OpenAI (GPT)',
    xai: 'xAI (Grok)',
};

interface Props {
    title?: string;
    value: MasterPersonaLLMConfig;
    onChange: (next: MasterPersonaLLMConfig) => void;
    enabled?: boolean;
    namespace: string;
}

const ModelSettings: React.FC<Props> = ({ title = 'モデル設定', value, onChange, enabled = true, namespace }) => {
    const [dynamicOptionsByProvider, setDynamicOptionsByProvider] = useState<Partial<Record<MasterPersonaProvider, ModelOptionItem[]>>>({});
    const [catalogLoading, setCatalogLoading] = useState<boolean>(false);
    const [catalogError, setCatalogError] = useState<string>('');
    const [draftTemperature, setDraftTemperature] = useState<number>(value.temperature);

    const provider = value.provider;
    const endpoint = value.endpoint;
    const apiKey = value.apiKey;
    const currentModel = value.model;
    const isLMStudio = value.provider === 'lmstudio';

    const modelOptions = useMemo(() => {
        const dynamic = dynamicOptionsByProvider[value.provider] ?? [];
        if (dynamic.length > 0) {
            return dynamic;
        }
        return FALLBACK_MODEL_OPTIONS[value.provider];
    }, [dynamicOptionsByProvider, value.provider]);

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
        if (selectableModelOptions.some((x) => x.id === currentModel)) {
            return currentModel;
        }
        const byLabel = selectableModelOptions.find((x) => x.label === currentModel);
        if (byLabel) {
            return byLabel.id;
        }
        return selectableModelOptions[0].id;
    }, [currentModel, selectableModelOptions]);

    useEffect(() => {
        setDraftTemperature(value.temperature);
    }, [value.temperature]);

    const commitTemperature = () => {
        if (value.temperature !== draftTemperature) {
            onChange({ ...value, temperature: draftTemperature });
        }
    };


    useEffect(() => {
        if (!enabled) {
            return;
        }
        let alive = true;
        (async () => {
            setCatalogLoading(true);
            setCatalogError('');
            const targetProvider = provider;
            try {
                const rows = await ListModels({
                    namespace,
                    provider: targetProvider,
                    endpoint,
                    apiKey: isLMStudio ? '' : apiKey,
                });
                if (!alive) {
                    return;
                }
                const seen = new Set<string>();
                const options: ModelOptionItem[] = [];
                for (const row of rows) {
                    const id = (row.id || '').trim();
                    const label = (row.display_name || row.id || '').trim();
                    if (id.length === 0 || seen.has(id)) {
                        continue;
                    }
                    seen.add(id);
                    options.push({ id, label: label.length > 0 ? label : id });
                }
                setDynamicOptionsByProvider((prev) => ({ ...prev, [targetProvider]: options }));
                const hasCurrent = options.some((x) => x.id === currentModel || x.label === currentModel);
                if (options.length > 0 && !hasCurrent) {
                    onChange({ ...value, model: options[0].id });
                }
            } catch {
                if (!alive) {
                    return;
                }
                setDynamicOptionsByProvider((prev) => ({ ...prev, [targetProvider]: [] }));
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
    }, [apiKey, enabled, endpoint, isLMStudio, namespace, onChange, provider]);

    const handleProviderChange = (nextProvider: MasterPersonaProvider) => {
        const candidates = FALLBACK_MODEL_OPTIONS[nextProvider];
        const nextModel = nextProvider === value.provider ? value.model : (candidates[0]?.id ?? '');
        onChange({
            ...value,
            provider: nextProvider,
            model: nextModel,
            endpoint: nextProvider === 'lmstudio' ? value.endpoint || 'http://localhost:1234' : value.endpoint,
        });
    };

    return (
        <details className="collapse collapse-arrow bg-base-100 border border-base-200 shadow-sm" open>
            <summary className="collapse-title text-base font-bold min-h-0 py-3 border-b border-base-200">
                {title}
            </summary>
            <div className="collapse-content pt-4">
                <div className="flex flex-col gap-6">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">AIプロバイダ</span>
                            </label>
                            <select
                                className="select select-bordered select-sm w-full"
                                value={value.provider}
                                onChange={(e) => handleProviderChange(e.target.value as MasterPersonaProvider)}
                            >
                                {(Object.keys(PROVIDER_LABELS) as MasterPersonaProvider[]).map((p) => (
                                    <option key={p} value={p}>{PROVIDER_LABELS[p]}</option>
                                ))}
                            </select>
                        </div>

                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">モデル</span>
                            </label>
                            <select
                                className="select select-bordered select-sm w-full"
                                value={selectedModelValue}
                                onChange={(e) => onChange({ ...value, model: e.target.value })}
                            >
                                {selectableModelOptions.map((m) => (
                                    <option key={m.id} value={m.id}>{m.label}</option>
                                ))}
                            </select>
                            <span className="text-xs text-base-content/50 mt-1 block">
                                {catalogLoading ? 'モデル一覧を取得中...' : catalogError || ''}
                            </span>
                        </div>
                    </div>

                    <div className={`grid gap-4 ${isLMStudio ? 'grid-cols-1' : 'grid-cols-2'}`}>
                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">Endpoint</span>
                            </label>
                            <input
                                type="text"
                                className="input input-bordered input-sm w-full font-mono"
                                placeholder="http://localhost:1234"
                                value={value.endpoint}
                                onChange={(e) => onChange({ ...value, endpoint: e.target.value })}
                            />
                        </div>
                        {!isLMStudio && (
                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">API Key</span>
                                </label>
                                <input
                                    type="password"
                                    className="input input-bordered input-sm w-full"
                                    value={value.apiKey}
                                    onChange={(e) => onChange({ ...value, apiKey: e.target.value })}
                                />
                            </div>
                        )}
                    </div>

                    <div className="grid grid-cols-1 gap-4">
                        <div className="flex flex-col gap-1">
                            <div className="flex justify-between items-center">
                                <label className="label-text font-bold text-sm">Temperature</label>
                                <span className="badge badge-ghost badge-sm font-mono">{draftTemperature.toFixed(2)}</span>
                            </div>
                            <input
                                type="range"
                                min="0"
                                max="2"
                                step="0.01"
                                className="range range-primary range-sm w-full"
                                value={draftTemperature}
                                onChange={(e) => setDraftTemperature(parseFloat(e.target.value))}
                                onMouseUp={commitTemperature}
                                onTouchEnd={commitTemperature}
                                onKeyUp={commitTemperature}
                            />
                        </div>
                    </div>
                </div>
            </div>
        </details>
    );
};

export default ModelSettings;
