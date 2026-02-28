import React, { useState } from 'react';

type Provider = 'gemini' | 'openai' | 'xai' | 'local';

interface ModelConfig {
    [key: string]: string[];
}

const MODEL_OPTIONS: ModelConfig = {
    gemini: ['gemini-2.0-flash', 'gemini-2.0-pro', 'gemini-1.5-pro', 'gemini-1.5-flash'],
    openai: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-3.5-turbo'],
    xai: ['grok-3', 'grok-3-mini', 'grok-2'],
    local: ['カスタム (手動入力)'],
};

const PROVIDER_LABELS: Record<Provider, string> = {
    gemini: 'Google Gemini',
    openai: 'OpenAI (GPT)',
    xai: 'xAI (Grok)',
    local: 'Local LLM (LM Studio)',
};

interface Props {
    /** カードタイトル。省略時は「モデル設定」 */
    title?: string;
}

const ModelSettings: React.FC<Props> = ({ title = 'モデル設定' }) => {
    const [provider, setProvider] = useState<Provider>('gemini');
    const [model, setModel] = useState<string>(MODEL_OPTIONS['gemini'][0]);
    const [temperature, setTemperature] = useState<number>(0.7);
    const [topK, setTopK] = useState<number>(40);
    const [topP, setTopP] = useState<number>(0.95);
    const [maxContext, setMaxContext] = useState<number>(32768);
    const [batchMode, setBatchMode] = useState<boolean>(false);
    const [parallel, setParallel] = useState<number>(4);

    const handleProviderChange = (next: Provider) => {
        setProvider(next);
        setModel(MODEL_OPTIONS[next][0]);
        setBatchMode(false);
    };

    const isBatchCapable = provider === 'gemini' || provider === 'xai';
    const isLocal = provider === 'local';

    return (
        <details
            className="collapse collapse-arrow bg-base-100 border border-base-200 shadow-sm"
            open
        >
            <summary className="collapse-title text-base font-bold min-h-0 py-3 border-b border-base-200">
                {title}
            </summary>
            <div className="collapse-content pt-4">
                <div className="flex flex-col gap-6">

                    {/* ── プロバイダ ＆ モデル ── */}
                    <div className="grid grid-cols-2 gap-4">
                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">AIプロバイダ</span>
                            </label>
                            <select
                                className="select select-bordered select-sm w-full"
                                value={provider}
                                onChange={(e) => handleProviderChange(e.target.value as Provider)}
                            >
                                {(Object.keys(PROVIDER_LABELS) as Provider[]).map((p) => (
                                    <option key={p} value={p}>{PROVIDER_LABELS[p]}</option>
                                ))}
                            </select>
                        </div>

                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">モデル</span>
                            </label>
                            {isLocal ? (
                                <input
                                    type="text"
                                    className="input input-bordered input-sm w-full"
                                    placeholder="例: lmstudio-community/Meta-Llama-3-8B-Instruct-GGUF"
                                    value={model === 'カスタム (手動入力)' ? '' : model}
                                    onChange={(e) => setModel(e.target.value)}
                                />
                            ) : (
                                <select
                                    className="select select-bordered select-sm w-full"
                                    value={model}
                                    onChange={(e) => setModel(e.target.value)}
                                >
                                    {MODEL_OPTIONS[provider].map((m) => (
                                        <option key={m} value={m}>{m}</option>
                                    ))}
                                </select>
                            )}
                        </div>
                    </div>

                    {/* ── 推論パラメータ ── */}
                    <div className="flex flex-col gap-4">
                        <span className="text-sm font-bold text-base-content/70">推論パラメータ</span>

                        {/* Temperature */}
                        <div className="flex flex-col gap-1">
                            <div className="flex justify-between items-center">
                                <label className="label-text font-bold text-sm">Temperature</label>
                                <span className="badge badge-ghost badge-sm font-mono">{temperature.toFixed(2)}</span>
                            </div>
                            <input
                                type="range" min="0" max="2" step="0.01"
                                className="range range-primary range-sm"
                                value={temperature}
                                onChange={(e) => setTemperature(parseFloat(e.target.value))}
                            />
                            <div className="flex justify-between text-xs text-base-content/50 px-1">
                                <span>0 (決定的)</span>
                                <span>2 (ランダム)</span>
                            </div>
                        </div>

                        {/* Top-K / Top-P */}
                        <div className="grid grid-cols-2 gap-4">
                            <div className="flex flex-col gap-1">
                                <div className="flex justify-between items-center">
                                    <label className="label-text font-bold text-sm">Top-K</label>
                                    <span className="badge badge-ghost badge-sm font-mono">{topK}</span>
                                </div>
                                <input
                                    type="range" min="1" max="100" step="1"
                                    className="range range-secondary range-sm"
                                    value={topK}
                                    onChange={(e) => setTopK(parseInt(e.target.value))}
                                />
                            </div>
                            <div className="flex flex-col gap-1">
                                <div className="flex justify-between items-center">
                                    <label className="label-text font-bold text-sm">Top-P</label>
                                    <span className="badge badge-ghost badge-sm font-mono">{topP.toFixed(2)}</span>
                                </div>
                                <input
                                    type="range" min="0" max="1" step="0.01"
                                    className="range range-secondary range-sm"
                                    value={topP}
                                    onChange={(e) => setTopP(parseFloat(e.target.value))}
                                />
                            </div>
                        </div>
                    </div>

                    {/* ── Local LLM: 最大コンテキスト長 ── */}
                    {isLocal && (
                        <div className="flex flex-col gap-1">
                            <label className="label pb-0">
                                <span className="label-text font-bold">最大コンテキスト長 (トークン)</span>
                                <span className="label-text-alt text-base-content/50">LM Studio側のモデル設定と合わせること</span>
                            </label>
                            <div className="flex gap-2 items-center">
                                <input
                                    type="number"
                                    className="input input-bordered input-sm w-40 font-mono"
                                    value={maxContext}
                                    step={1024}
                                    min={2048}
                                    onChange={(e) => setMaxContext(parseInt(e.target.value))}
                                />
                                <span className="text-sm text-base-content/60">tokens</span>
                                <div className="flex gap-1 ml-2">
                                    {[4096, 8192, 16384, 32768].map((v) => (
                                        <button
                                            key={v}
                                            className={`btn btn-xs btn-outline ${maxContext === v ? 'btn-primary' : ''}`}
                                            onClick={() => setMaxContext(v)}
                                        >{(v / 1024)}K</button>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* ── xAI / Gemini: バッチAPIモード ── */}
                    {isBatchCapable && (
                        <div className="flex flex-col gap-3">
                            <div className="flex items-center justify-between">
                                <div className="flex flex-col gap-0">
                                    <span className="label-text font-bold text-sm">バッチAPIモード</span>
                                    <span className="text-xs text-base-content/50">
                                        {provider === 'gemini' ? 'Gemini Batch API' : 'xAI Batch API'} を使用。コスト削減、低速。
                                    </span>
                                </div>
                                <input
                                    type="checkbox"
                                    className="toggle toggle-primary"
                                    checked={batchMode}
                                    onChange={(e) => setBatchMode(e.target.checked)}
                                />
                            </div>

                            {batchMode && (
                                <div className="alert alert-warning py-2 text-sm">
                                    <svg xmlns="http://www.w3.org/2000/svg" className="stroke-current shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                    </svg>
                                    バッチモードでは並列実行数の設定は無効になります。結果の取得に最大24時間かかる場合があります。
                                </div>
                            )}
                        </div>
                    )}

                    {/* ── 通常モード: 並列実行数 ── */}
                    {!batchMode && (
                        <div className="flex flex-col gap-1">
                            <div className="flex justify-between items-center">
                                <div className="flex flex-col gap-0">
                                    <label className="label-text font-bold text-sm">並列実行数 (Workers)</label>
                                    <span className="text-xs text-base-content/50">APIレート制限に注意。ローカルLLMでは1〜2を推奨。</span>
                                </div>
                                <span className="badge badge-primary badge-sm font-mono">{parallel}</span>
                            </div>
                            <input
                                type="range" min="1" max="16" step="1"
                                className="range range-primary range-sm"
                                value={parallel}
                                onChange={(e) => setParallel(parseInt(e.target.value))}
                            />
                            <div className="flex justify-between text-xs text-base-content/50 px-1">
                                <span>1</span>
                                <span>4</span>
                                <span>8</span>
                                <span>12</span>
                                <span>16</span>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </details>
    );
};

export default ModelSettings;
