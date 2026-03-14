import React, {useEffect, useRef, useState} from 'react';
import {type MasterPersonaLLMConfig} from '../types/masterPersona';
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

const EXECUTION_PROFILE_LABELS: Record<MasterPersonaLLMConfig['bulkStrategy'], string> = {
    sync: '同期実行',
    batch: 'クラウドBatch',
};

interface Props {
    title?: string;
    value: MasterPersonaLLMConfig;
    onChange: (next: MasterPersonaLLMConfig) => void;
    enabled?: boolean;
    namespace: string;
    locked?: boolean;
}

/**
 * LLMモデルの設定を行うコンポーネント。
 *
 * プロバイダ選択、モデル選択、エンドポイント、APIキー、並列実行数、Temperature などの
 * 設定項目を提供し、入力値のデバウンス処理を行う。
 *
 * @param props - コンポーネントのプロパティ
 * @param props.title - セクションタイトル（デフォルト: 'モデル設定'）
 * @param props.value - 現在のLLM設定値
 * @param props.onChange - 設定変更時のコールバック
 * @param props.enabled - モデル一覧取得の有効/無効
 * @param props.namespace - モデル設定の名前空間
 */
const ModelSettings: React.FC<Props> = ({ title = 'モデル設定', value, onChange, enabled = true, namespace, locked = false }) => {
    const [draftTemperature, setDraftTemperature] = useState<number>(value.temperature);
    const [draftEndpoint, setDraftEndpoint] = useState<string>(value.endpoint);
    const [draftApiKey, setDraftApiKey] = useState<string>(value.apiKey);
    const [draftSyncConcurrency, setDraftSyncConcurrency] = useState<number>(value.syncConcurrency);
    const [draftContextLength, setDraftContextLength] = useState<number>(value.contextLength);

    const debounceTimerRef = useRef<number | null>(null);

    const isLMStudio = value.provider === 'lmstudio';
    const controlsDisabled = locked;
    const {
        availableExecutionProfiles,
        catalogError,
        catalogLoading,
        handleProviderChange,
        selectableModelOptions,
        selectedModelCapability,
        selectedModelValue,
    } = useModelSettings({ value, onChange, enabled, namespace });

    useEffect(() => {
        setDraftTemperature(value.temperature);
    }, [value.temperature]);

    useEffect(() => {
        setDraftEndpoint(value.endpoint);
    }, [value.endpoint]);

    useEffect(() => {
        setDraftApiKey(value.apiKey);
    }, [value.apiKey]);

    useEffect(() => {
        setDraftSyncConcurrency(value.syncConcurrency);
    }, [value.syncConcurrency]);

    useEffect(() => {
        setDraftContextLength(value.contextLength);
    }, [value.contextLength]);

    /**
     * 設定値の変更を500msデバウンスして親コンポーネントへ反映する。
     * 頻繁な設定保存を防ぎ、ユーザーの入力完了を待つ。
     *
     * @param updates - 更新する設定項目の部分オブジェクト
     */
    const debouncedOnChange = (updates: Partial<MasterPersonaLLMConfig>) => {
        if (debounceTimerRef.current) {
            clearTimeout(debounceTimerRef.current);
        }
        debounceTimerRef.current = setTimeout(() => {
            onChange({ ...value, ...updates });
        }, 500);
    };

    /**
     * Temperature スライダーの操作完了時（マウスアップ、タッチエンド等）に
     * draft値を親コンポーネントへ即座に反映する。
     */
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
                    {controlsDisabled && (
                        <div className="alert alert-warning py-2 text-sm">
                            <span>再開対象タスクがあるため、モデル設定は固定されています。</span>
                        </div>
                    )}

                    <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">AIプロバイダ</span>
                            </label>
                            <select
                                className="select select-bordered select-sm w-full"
                                value={value.provider}
                                onChange={(e) => handleProviderChange(e.target.value)}
                                disabled={controlsDisabled}
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
                                disabled={controlsDisabled}
                            >
                                {selectableModelOptions.map((m) => (
                                    <option key={m.id} value={m.id}>{m.label}</option>
                                ))}
                            </select>
                            <span className="text-xs text-base-content/50 mt-1 block">
                                {catalogLoading ? 'モデル一覧を取得中...' : catalogError || ''}
                            </span>
                        </div>

                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">実行方式</span>
                            </label>
                            <select
                                className="select select-bordered select-sm w-full"
                                value={value.bulkStrategy}
                                onChange={(e) => onChange({ ...value, bulkStrategy: e.target.value as MasterPersonaLLMConfig['bulkStrategy'] })}
                                disabled={controlsDisabled || availableExecutionProfiles.length <= 1}
                            >
                                {availableExecutionProfiles.map((profile) => (
                                    <option key={profile} value={profile}>{EXECUTION_PROFILE_LABELS[profile]}</option>
                                ))}
                            </select>
                        </div>
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
                                disabled={controlsDisabled}
                            />
                        </div>
                        <span className={`text-xs ${selectedModelCapability.supportsBatch ? 'text-success' : 'text-base-content/70'}`}>
                            {selectedModelCapability.supportsBatch
                                ? 'このモデルは Batch API に対応しています'
                                : 'このモデルは同期実行のみ対応です'}
                        </span>
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
                                value={draftEndpoint}
                                onChange={(e) => {
                                    setDraftEndpoint(e.target.value);
                                    debouncedOnChange({ endpoint: e.target.value });
                                }}
                                disabled={controlsDisabled}
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
                                    value={draftApiKey}
                                    onChange={(e) => {
                                        setDraftApiKey(e.target.value);
                                        debouncedOnChange({ apiKey: e.target.value });
                                    }}
                                    disabled={controlsDisabled}
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
                                            disabled={controlsDisabled}
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
                                    value={draftContextLength}
                                    onChange={(e) => {
                                        const parsed = Number.parseInt(e.target.value, 10);
                                        const newValue = Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
                                        setDraftContextLength(newValue);
                                        debouncedOnChange({ contextLength: newValue });
                                    }}
                                    disabled={controlsDisabled}
                                />
                                <span className="text-xs text-base-content/60">0 は LM Studio の既定値を使用</span>
                            </div>
                        </div>
                    )}

                    {value.bulkStrategy === 'sync' && (
                        <div className="grid grid-cols-1 gap-4">
                            <div className="flex flex-col gap-1">
                                <div className="flex justify-between items-center">
                                    <label className="label-text font-bold text-sm">同期並列数</label>
                                    <span className="badge badge-ghost badge-sm font-mono">{draftSyncConcurrency}</span>
                                </div>
                                <input
                                    type="range"
                                    min="1"
                                    max="64"
                                    step="1"
                                    className="range range-primary range-sm w-full"
                                    value={draftSyncConcurrency}
                                    onChange={(e) => {
                                        const parsed = Number.parseInt(e.target.value, 10);
                                        const newValue = Number.isFinite(parsed) && parsed > 0 ? parsed : 1;
                                        setDraftSyncConcurrency(newValue);
                                        debouncedOnChange({ syncConcurrency: newValue });
                                    }}
                                    disabled={controlsDisabled}
                                />
                                <span className="text-xs text-base-content/60">1〜64 で調整</span>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </details>
    );
};

export default ModelSettings;
