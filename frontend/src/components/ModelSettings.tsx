import React, { useEffect, useState } from 'react';
import { type MasterPersonaLLMConfig } from '../types/masterPersona';
import {
    MASTER_PERSONA_PROVIDERS,
    PROVIDER_LABELS,
    useModelSettings,
} from '../hooks/features/modelSettings/useModelSettings';

const CONTEXT_LENGTH_PRESETS: Array<{ label: string; value: number }> = [
    { label: '4096', value: 4096 },
    { label: '8k', value: 8192 },
    { label: '16k', value: 16384 },
    { label: '32k', value: 32768 },
    { label: '64k', value: 65536 },
];

interface Props {
    title?: string;
    value: MasterPersonaLLMConfig;
    onChange: (next: MasterPersonaLLMConfig) => void;
    enabled?: boolean;
    namespace: string;
}

const ModelSettings: React.FC<Props> = ({ title = 'モデル設定', value, onChange, enabled = true, namespace }) => {
    const [draftTemperature, setDraftTemperature] = useState<number>(value.temperature);
    const isLMStudio = value.provider === 'lmstudio';
    const {
        catalogError,
        catalogLoading,
        handleProviderChange,
        selectableModelOptions,
        selectedModelValue,
    } = useModelSettings({ value, onChange, enabled, namespace });

    useEffect(() => {
        setDraftTemperature(value.temperature);
    }, [value.temperature]);

    const commitTemperature = () => {
        if (value.temperature !== draftTemperature) {
            onChange({ ...value, temperature: draftTemperature });
        }
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
                                onChange={(e) => handleProviderChange(e.target.value)}
                            >
                                {MASTER_PERSONA_PROVIDERS.map((p) => (
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
                                <span className="label-text font-bold">エンドポイント</span>
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

                    {isLMStudio && (
                        <div className="grid grid-cols-1 gap-4">
                            <div className="flex flex-col gap-1">
                                <label className="label pb-0">
                                    <span className="label-text font-bold">コンテキスト長</span>
                                </label>
                                <div className="flex flex-wrap gap-2 mb-2">
                                    {CONTEXT_LENGTH_PRESETS.map((preset) => (
                                        <button
                                            key={preset.label}
                                            type="button"
                                            className={`btn btn-xs ${value.contextLength === preset.value ? 'btn-primary' : 'btn-outline'}`}
                                            onClick={() => onChange({ ...value, contextLength: preset.value })}
                                        >
                                            {preset.label}
                                        </button>
                                    ))}
                                </div>
                                <input
                                    type="number"
                                    min={0}
                                    step={1}
                                    className="input input-bordered input-sm w-full font-mono"
                                    value={value.contextLength}
                                    onChange={(e) => {
                                        const parsed = Number.parseInt(e.target.value, 10);
                                        onChange({
                                            ...value,
                                            contextLength: Number.isFinite(parsed) && parsed > 0 ? parsed : 0,
                                        });
                                    }}
                                />
                                <span className="text-xs text-base-content/60">0 は LM Studio の既定値を使用</span>
                            </div>
                        </div>
                    )}

                    <div className="grid grid-cols-1 gap-4">
                        <div className="flex flex-col gap-1">
                            <div className="flex justify-between items-center">
                                <label className="label-text font-bold text-sm">並列実行数</label>
                                <span className="badge badge-ghost badge-sm font-mono">{value.syncConcurrency}</span>
                            </div>
                            <input
                                type="range"
                                min="1"
                                max="64"
                                step="1"
                                className="range range-primary range-sm w-full"
                                value={value.syncConcurrency}
                                onChange={(e) => {
                                    const parsed = Number.parseInt(e.target.value, 10);
                                    onChange({
                                        ...value,
                                        syncConcurrency: Number.isFinite(parsed) && parsed > 0 ? parsed : 1,
                                    });
                                }}
                            />
                            <span className="text-xs text-base-content/60">1〜64 で調整</span>
                        </div>
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
