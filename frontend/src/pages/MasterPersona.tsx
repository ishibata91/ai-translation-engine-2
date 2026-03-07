import React, { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import PersonaDetail from '../components/PersonaDetail';
import { type NpcRow, type NpcStatus, STATUS_BADGE } from '../types/npc';
import { SelectJSONFile } from '../wailsjs/go/main/App';
import { CancelTask, GetAllTasks, GetTaskRequestState, ResumeTask, StartMasterPersonTask } from '../wailsjs/go/task/Bridge';
import { ConfigGetAll, ConfigSet } from '../wailsjs/go/config/ConfigService';
import type { PhaseCompletedEvent, FrontendTask } from '../types/task';
import { DEFAULT_MASTER_PERSONA_LLM_CONFIG, type MasterPersonaLLMConfig } from '../types/masterPersona';
import * as Events from '../wailsjs/runtime/runtime';

type PersonaProgressStatus = 'IN_PROGRESS' | 'COMPLETED' | 'FAILED';

interface PersonaProgressEvent {
    CorrelationID: string;
    Total: number;
    Completed: number;
    Failed: number;
    Status: PersonaProgressStatus;
    Message: string;
}

interface PersonaRequestSummary {
    request_count: number;
    npc_count: number;
}

const MASTER_PERSONA_LLM_NAMESPACE = 'master_persona.llm';
const SELECTED_PROVIDER_KEY = 'selected_provider';

const normalizeProvider = (value: string | undefined): MasterPersonaLLMConfig['provider'] => {
    if (value === 'lmstudio' || value === 'gemini' || value === 'openai' || value === 'xai') {
        return value;
    }
    return DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
};

const providerNamespace = (provider: MasterPersonaLLMConfig['provider']): string =>
    `${MASTER_PERSONA_LLM_NAMESPACE}.${provider}`;
const syncConcurrencyKey = (provider: MasterPersonaLLMConfig['provider']): string =>
    `sync_concurrency.${provider}`;

const toErrorMessage = (error: unknown, fallback: string): string => {
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

const statusMessageFromTask = (task: FrontendTask): string => {
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

const buildProviderConfigPairs = (cfg: MasterPersonaLLMConfig): Record<string, string> => ({
    model: cfg.model,
    endpoint: cfg.endpoint,
    api_key: cfg.provider === 'lmstudio' ? '' : cfg.apiKey,
    temperature: String(cfg.temperature),
    context_length: String(cfg.contextLength),
});

const parseTaskTimestamp = (value: string | undefined): number => {
    if (!value) {
        return 0;
    }
    const t = Date.parse(value);
    return Number.isFinite(t) ? t : 0;
};

// ── 列定義 ───────────────────────────────────────────────
const NPC_COLUMNS: ColumnDef<NpcRow, unknown>[] = [
    {
        accessorKey: 'formId',
        header: 'FormID',
        cell: (info) => <span className="font-mono text-sm">{info.getValue() as string}</span>,
    },
    {
        accessorKey: 'name',
        header: 'NPC名 (EditorID)',
    },
    {
        accessorKey: 'dialogueCount',
        header: 'セリフ数',
        cell: (info) => <span className="font-mono text-right block">{info.getValue() as number}</span>,
    },
    {
        accessorKey: 'status',
        header: 'ステータス',
        cell: (info) => {
            const s = info.getValue() as NpcStatus;
            return <div className={`badge badge-sm ${STATUS_BADGE[s]}`}>{s}</div>;
        },
    },
    {
        accessorKey: 'updatedAt',
        header: '生成日時',
    },
];

// ── ページコンポーネント ──────────────────────────────────
const MasterPersona: React.FC = () => {
    const location = useLocation();
    const [npcData] = useState<NpcRow[]>([]);
    const [selectedRow, setSelectedRow] = useState<NpcRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [isGenerating, setIsGenerating] = useState<boolean>(false);
    const [jsonPath, setJsonPath] = useState<string>('');
    const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
    const [progressPercent, setProgressPercent] = useState<number>(0);
    const [statusMessage, setStatusMessage] = useState<string>('待機中');
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [progressCounts, setProgressCounts] = useState<{ current: number; total: number }>({ current: 0, total: 0 });
    const [summary, setSummary] = useState<PersonaRequestSummary | null>(null);
    const [activeTaskStatus, setActiveTaskStatus] = useState<FrontendTask['status'] | null>(null);
    const [llmConfig, setLLMConfig] = useState<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const [isLLMConfigHydrated, setIsLLMConfigHydrated] = useState<boolean>(false);
    const lastSavedLLMConfigRef = useRef<Partial<Record<MasterPersonaLLMConfig['provider'], MasterPersonaLLMConfig>>>({});
    const latestLLMConfigRef = useRef<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const selectedProviderRef = useRef<MasterPersonaLLMConfig['provider']>(DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider);
    const saveQueueRef = useRef<Promise<void>>(Promise.resolve());
    const isSwitchingProviderRef = useRef<boolean>(false);
    const previousProviderRef = useRef<MasterPersonaLLMConfig['provider']>(DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider);
    const resumeRequestedRef = useRef<boolean>(false);

    const loadProviderConfig = async (
        provider: MasterPersonaLLMConfig['provider'],
        legacyRoot?: Record<string, string>,
    ): Promise<MasterPersonaLLMConfig> => {
        const root = legacyRoot ?? await ConfigGetAll(MASTER_PERSONA_LLM_NAMESPACE);
        const loaded = await ConfigGetAll(providerNamespace(provider));
        const source = Object.keys(loaded).length > 0 ? loaded : (legacyRoot ?? {});
        const loadedTemperature = Number.parseFloat(source.temperature ?? '');
        const loadedContextLength = Number.parseInt(source.context_length ?? '', 10);
        const loadedSyncConcurrency = Number.parseInt(root[syncConcurrencyKey(provider)] ?? '', 10);
        const config = {
            provider,
            model: source.model ?? DEFAULT_MASTER_PERSONA_LLM_CONFIG.model,
            endpoint: source.endpoint || DEFAULT_MASTER_PERSONA_LLM_CONFIG.endpoint,
            apiKey: source.api_key ?? DEFAULT_MASTER_PERSONA_LLM_CONFIG.apiKey,
            temperature: Number.isFinite(loadedTemperature) ? loadedTemperature : DEFAULT_MASTER_PERSONA_LLM_CONFIG.temperature,
            contextLength: Number.isFinite(loadedContextLength) && loadedContextLength > 0
                ? loadedContextLength
                : DEFAULT_MASTER_PERSONA_LLM_CONFIG.contextLength,
            syncConcurrency: Number.isFinite(loadedSyncConcurrency) && loadedSyncConcurrency > 0
                ? loadedSyncConcurrency
                : DEFAULT_MASTER_PERSONA_LLM_CONFIG.syncConcurrency,
        };
        return config;
    };

    const handleRowSelect = (row: NpcRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
    };

    const handlePickJson = async () => {
        const selected = await SelectJSONFile();
        if (!selected) {
            return;
        }
        setJsonPath(selected);
        setErrorMessage('');
    };

    const handleStart = async () => {
        if (!jsonPath || isGenerating) {
            return;
        }
        setIsGenerating(true);
        setActiveTaskId(null);
        setErrorMessage('');
        setSummary(null);
        setProgressPercent(0);
        setProgressCounts({ current: 0, total: 0 });
        setStatusMessage('タスクを開始しています...');
        resumeRequestedRef.current = false;
        setActiveTaskStatus('pending');

        try {
            const taskID = await StartMasterPersonTask({ source_json_path: jsonPath });
            setActiveTaskId(taskID);
        } catch (error) {
            setIsGenerating(false);
            setStatusMessage('タスク開始に失敗しました');
            setErrorMessage(error instanceof Error ? error.message : '不明なエラーが発生しました');
        }
    };

    const hydrateTaskView = (task: FrontendTask) => {
        setActiveTaskId(task.id);
        setActiveTaskStatus(task.status);
        setProgressPercent(task.progress || 0);
        setStatusMessage(statusMessageFromTask(task));
        setErrorMessage(task.error_msg || '');
        setIsGenerating(task.status === 'running');
        const requestCount = Number(task.metadata?.request_count ?? 0);
        const npcCount = Number(task.metadata?.npc_count ?? 0);
        if (requestCount > 0 || npcCount > 0) {
            setSummary({ request_count: requestCount, npc_count: npcCount });
        }
        const resumeCursor = Number(task.metadata?.resume_cursor ?? 0);
        setProgressCounts({
            current: Number.isFinite(resumeCursor) ? resumeCursor : 0,
            total: requestCount > 0 ? requestCount : 0,
        });
        const sourceJSONPath = String(task.metadata?.source_json_path ?? '');
        if (sourceJSONPath) {
            setJsonPath(sourceJSONPath);
        }
    };

    const handleResumeCurrentTask = async () => {
        if (!activeTaskId) {
            return;
        }
        resumeRequestedRef.current = false;
        setErrorMessage('');
        setStatusMessage('リクエストを実行しています...');
        setIsGenerating(true);
        try {
            await ResumeTask(activeTaskId);
        } catch (error) {
            setIsGenerating(false);
            setStatusMessage('キュー実行の開始に失敗しました');
            setErrorMessage(toErrorMessage(error, 'キュー実行の開始に失敗しました'));
        }
    };

    const handlePauseCurrentTask = async () => {
        if (!activeTaskId) {
            return;
        }
        try {
            await CancelTask(activeTaskId);
            setIsGenerating(false);
            setActiveTaskStatus('cancelled');
            setStatusMessage('一時停止中');
        } catch (error) {
            setErrorMessage(toErrorMessage(error, '一時停止に失敗しました'));
        }
    };

    useEffect(() => {
        let alive = true;
        (async () => {
            for (let attempt = 0; attempt < 3; attempt += 1) {
                try {
                    const root = await ConfigGetAll(MASTER_PERSONA_LLM_NAMESPACE);
                    if (!alive) {
                        return;
                    }

                    const selected = normalizeProvider(root[SELECTED_PROVIDER_KEY] || root.provider);
                    const hydrated = await loadProviderConfig(selected, root);
                    setLLMConfig(hydrated);
                    lastSavedLLMConfigRef.current[selected] = hydrated;
                    latestLLMConfigRef.current = hydrated;
                    selectedProviderRef.current = selected;
                    previousProviderRef.current = selected;
                    setIsLLMConfigHydrated(true);
                    return;
                } catch {
                    if (!alive) {
                        return;
                    }
                    if (attempt < 2) {
                        await new Promise((resolve) => setTimeout(resolve, 150));
                    }
                }
            }
            if (!alive) {
                return;
            }
            setLLMConfig(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
            lastSavedLLMConfigRef.current[DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider] = DEFAULT_MASTER_PERSONA_LLM_CONFIG;
            latestLLMConfigRef.current = DEFAULT_MASTER_PERSONA_LLM_CONFIG;
            selectedProviderRef.current = DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
            previousProviderRef.current = DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
            setIsLLMConfigHydrated(true);
        })();
        return () => {
            alive = false;
        };
    }, []);

    useEffect(() => {
        latestLLMConfigRef.current = llmConfig;
    }, [llmConfig]);

    const persistLLMConfigDiff = (currentRaw: MasterPersonaLLMConfig) => {
        const current = {
            ...currentRaw,
            apiKey: currentRaw.provider === 'lmstudio' ? '' : currentRaw.apiKey,
        };
        const currentPairs = buildProviderConfigPairs(current);
        const previous = lastSavedLLMConfigRef.current[current.provider];
        const previousPairs = previous ? buildProviderConfigPairs(previous) : {};

        const writes: Array<[string, string]> = [];
        for (const key of Object.keys(currentPairs)) {
            const k = key as keyof typeof currentPairs;
            if (previousPairs[k] !== currentPairs[k]) {
                writes.push([k, currentPairs[k]]);
            }
        }
        const persistProvider = writes.length === 0
            ? Promise.resolve()
            : Promise.all(
                writes.map(([key, val]) => ConfigSet(providerNamespace(current.provider), key, val)),
            ).then(() => undefined);
        const persistSyncConcurrency = previous?.syncConcurrency === current.syncConcurrency
            ? Promise.resolve()
            : ConfigSet(
                MASTER_PERSONA_LLM_NAMESPACE,
                syncConcurrencyKey(current.provider),
                String(current.syncConcurrency),
            );
        const persistAll = Promise.all([persistProvider, persistSyncConcurrency]).then(() => {
            lastSavedLLMConfigRef.current[current.provider] = current;
        });

        if (selectedProviderRef.current === current.provider) {
            return persistAll.then(() => undefined);
        }
        return persistAll.then(() =>
            ConfigSet(MASTER_PERSONA_LLM_NAMESPACE, SELECTED_PROVIDER_KEY, current.provider),
        ).then(() => {
            selectedProviderRef.current = current.provider;
        });
    };

    useEffect(() => {
        if (!isLLMConfigHydrated) {
            return;
        }
        const currentProvider = llmConfig.provider;
        const previousProvider = previousProviderRef.current;
        if (currentProvider === previousProvider) {
            return;
        }
        previousProviderRef.current = currentProvider;
        isSwitchingProviderRef.current = true;
        let alive = true;
        void loadProviderConfig(currentProvider)
            .then((nextConfig) => {
                if (!alive) {
                    return;
                }
                setLLMConfig(nextConfig);
                latestLLMConfigRef.current = nextConfig;
                if (lastSavedLLMConfigRef.current[currentProvider] == null) {
                    lastSavedLLMConfigRef.current[currentProvider] = nextConfig;
                }
            })
            .finally(() => {
                if (alive) {
                    isSwitchingProviderRef.current = false;
                    const snapshot = latestLLMConfigRef.current;
                    saveQueueRef.current = saveQueueRef.current
                        .then(() => persistLLMConfigDiff(snapshot))
                        .catch((err) => {
                            console.error('failed to persist provider switch config', err);
                        });
                }
            });
        return () => {
            alive = false;
        };
    }, [isLLMConfigHydrated, llmConfig.provider]);

    useEffect(() => {
        if (!isLLMConfigHydrated) {
            return;
        }
        if (isSwitchingProviderRef.current) {
            return;
        }
        const snapshot = latestLLMConfigRef.current;
        saveQueueRef.current = saveQueueRef.current
            .then(() => persistLLMConfigDiff(snapshot))
            .catch((err) => {
                // 設定保存失敗時は次回変更時に再試行する。
                console.error('failed to persist master_persona.llm config', err);
            });
    }, [isLLMConfigHydrated, llmConfig]);

    useEffect(() => {
        return () => {
            if (!isLLMConfigHydrated) {
                return;
            }
            const snapshot = latestLLMConfigRef.current;
            saveQueueRef.current = saveQueueRef.current
                .then(() => persistLLMConfigDiff(snapshot))
                .catch((err) => {
                    // アンマウント時は失敗しても次回起動時に再入力可能。
                    console.error('failed to flush master_persona.llm config on unmount', err);
                });
        };
    }, [isLLMConfigHydrated]);

    const refreshProgressFromQueueState = async (task: FrontendTask) => {
        try {
            const state = await GetTaskRequestState(task.id);
            const queueTotal = Number(state.total ?? 0);
            const queueCompleted = Number(state.completed ?? 0);
            const metadataTotal = Number(task.metadata?.request_count ?? 0);
            const metadataCompleted = Number(task.metadata?.resume_cursor ?? 0);
            const total = queueTotal > 0 ? queueTotal : (metadataTotal > 0 ? metadataTotal : 0);
            const current = queueTotal > 0 ? queueCompleted : (metadataCompleted > 0 ? metadataCompleted : 0);
            setProgressCounts({
                current,
                total,
            });
            if (total > 0) {
                setProgressPercent(Math.min(100, Math.max(0, (current / total) * 100)));
            }
        } catch (error) {
            console.error('failed to refresh request progress state', { taskId: task.id, error });
        }
    };

    useEffect(() => {
        const navState = location.state as { taskId?: string; resumeFromDashboard?: boolean } | null;
        const taskIdFromNav = navState?.taskId;
        let disposed = false;
        void GetAllTasks()
            .then((tasks) => {
                if (disposed) {
                    return;
                }
                const personaTasks = (tasks as FrontendTask[])
                    .filter((t) => t.type === 'persona_extraction');
                let task: FrontendTask | undefined;
                if (taskIdFromNav) {
                    task = personaTasks.find((t) => t.id === taskIdFromNav);
                } else {
                    const recoverable = personaTasks
                        .filter((t) => t.status !== 'completed')
                        .sort((a, b) => parseTaskTimestamp(b.updated_at) - parseTaskTimestamp(a.updated_at));
                    task = recoverable[0];
                }
                if (!task) {
                    return;
                }
                hydrateTaskView(task);
                return refreshProgressFromQueueState(task);
            })
            .catch((error) => {
                console.error('failed to hydrate task from navigation state', error);
            });
        return () => {
            disposed = true;
        };
    }, [location.state]);

    useEffect(() => {
        const startQueuedExecution = (taskId: string) => {
            if (resumeRequestedRef.current) {
                return;
            }
            resumeRequestedRef.current = true;
            setStatusMessage('リクエストを実行しています...');
            void ResumeTask(taskId).catch((error) => {
                console.error('ResumeTask failed', { taskId, error });
                setIsGenerating(false);
                setStatusMessage('キュー実行の開始に失敗しました');
                setErrorMessage(toErrorMessage(error, 'キュー実行の開始に失敗しました'));
                resumeRequestedRef.current = false;
            });
        };

        const offProgress = Events.EventsOn('persona:progress', (event: PersonaProgressEvent) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? event.CorrelationID : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (currentTaskId && event.CorrelationID !== currentTaskId) {
                return;
            }

            if (event.Total > 0) {
                setProgressCounts({ current: event.Completed, total: event.Total });
                setProgressPercent(Math.min(100, Math.max(0, (event.Completed / event.Total) * 100)));
            }
            setStatusMessage(event.Message || '処理中');

            if (event.Status === 'FAILED') {
                setIsGenerating(false);
                setErrorMessage(event.Message || '生成に失敗しました');
            }
            if (event.Status === 'COMPLETED') {
                setIsGenerating(false);
            }
        });

        const offTaskUpdated = Events.EventsOn('task:updated', (task: FrontendTask) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? task.id : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (!currentTaskId || task.id !== currentTaskId) {
                return;
            }
            setActiveTaskStatus(task.status);
            if (task.status === 'paused' || task.status === 'cancelled' || task.status === 'completed' || task.status === 'failed') {
                void refreshProgressFromQueueState(task);
            }
            if (task.status === 'failed') {
                setIsGenerating(false);
                setErrorMessage(task.error_msg || 'タスク実行に失敗しました');
            }
            if (task.status === 'completed') {
                setIsGenerating(false);
            }
            if (task.status === 'paused' || task.status === 'cancelled') {
                setIsGenerating(false);
                setStatusMessage('一時停止中');
            }
            if (task.status === 'request_generated') {
                setStatusMessage('リクエスト生成完了。実行を開始します...');
            }
        });

        const offPhaseCompleted = Events.EventsOn('task:phase_completed', (payload: PhaseCompletedEvent) => {
            const currentTaskId = activeTaskId ?? (isGenerating ? payload.taskId : null);

            if (!activeTaskId && currentTaskId) {
                setActiveTaskId(currentTaskId);
            }

            if (!currentTaskId || payload.taskId !== currentTaskId || payload.phase !== 'REQUEST_GENERATED') {
                return;
            }
            const nextSummary = payload.summary as PersonaRequestSummary;
            setSummary({
                request_count: nextSummary.request_count ?? 0,
                npc_count: nextSummary.npc_count ?? 0,
            });
            setStatusMessage('リクエスト生成完了。実行を開始します...');
            startQueuedExecution(payload.taskId);
        });

        return () => {
            offProgress();
            offTaskUpdated();
            offPhaseCompleted();
        };
    }, [activeTaskId, isGenerating]);

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            {/* ヘッダー */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">マスターペルソナ構築 (Master Persona Builder)</span>
                </div>
            </div>

            {/* 通知エリア */}
            <div className="alert alert-info shadow-sm shrink-0">
                <span><code>extractData.pas</code> で抽出されたベースゲームのJSONデータからNPCのセリフを解析し、LLMを用いて基本となるペルソナ（性格・口調）を生成・キャッシュします。これによりMod翻訳時の品質と一貫性が向上します。</span>
            </div>

            {/* 上部パネル */}
            <div className="grid grid-cols-2 gap-4 shrink-0">
                {/* 生成設定カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">JSONデータのインポートと生成</h2>
                        <div className="flex flex-col gap-4 mt-2">
                            <span className="text-sm">xEditスクリプト <code>extractData.pas</code> によって抽出された、マスターファイルのJSONデータを選択し、ペルソナ生成を開始します。</span>
                            <div className="flex gap-4 items-center">
                                <input type="text" readOnly value={jsonPath} placeholder="JSONファイルを選択してください" className="input input-bordered w-full max-w-xl font-mono text-xs" />
                                <button className="btn btn-outline btn-primary" onClick={handlePickJson}>JSON選択</button>
                            </div>
                            <div>
                                <span className="mt-2 mb-1 block text-sm text-base-content/70 font-bold">全体進捗</span>
                                <progress className="progress progress-primary w-full" value={progressPercent} max="100"></progress>
                                {progressCounts.total > 0 && (
                                    <span className="text-xs text-base-content/70 mt-1 block">
                                        {progressCounts.current} / {progressCounts.total} 件
                                    </span>
                                )}
                                <span className="text-xs text-base-content/70 mt-1 block">{statusMessage}</span>
                                {errorMessage && <span className="text-xs text-error mt-1 block">{errorMessage}</span>}
                            </div>
                        </div>
                    </div>
                </div>

                {/* 統計カード */}
                <div className="card bg-base-100 border border-base-200 shadow-sm">
                    <div className="card-body">
                        <h2 className="card-title text-base">ペルソナDB ステータス</h2>
                        <div className="grid grid-cols-2 gap-4 mt-2">
                            <div className="stat p-0">
                                <div className="stat-title text-sm">登録済みNPC数</div>
                                <div className="stat-value text-primary font-mono text-3xl">{summary?.npc_count ?? 0}</div>
                            </div>
                            <div className="stat p-0">
                                <div className="stat-title text-sm">生成リクエスト数</div>
                                <div className="stat-value text-secondary font-mono text-3xl">{summary?.request_count ?? 0}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* モデル設定 */}
            <div className="shrink-0">
                {isLLMConfigHydrated ? (
                    <ModelSettings
                        title="ペルソナ生成モデル設定"
                        value={llmConfig}
                        onChange={setLLMConfig}
                        enabled={isLLMConfigHydrated}
                        namespace={providerNamespace(llmConfig.provider)}
                    />
                ) : (
                    <div className="card bg-base-100 border border-base-200 shadow-sm">
                        <div className="card-body py-4">
                            <span className="text-sm text-base-content/60">モデル設定を読み込み中...</span>
                        </div>
                    </div>
                )}
            </div>

            {/* 2ペインレイアウト (左: NPC テーブル, 右: PersonaDetail) */}
            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
                <div className="w-1/2 flex flex-col min-h-0 border border-base-200 rounded-xl bg-base-100 overflow-hidden">
                    <DataTable
                        columns={NPC_COLUMNS}
                        data={npcData}
                        title="NPC処理ステータス (Skyrim.esm)"
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelect}
                        enableColumnFilter
                    />
                </div>

                <div className="w-1/2 flex flex-col min-h-0">
                    <PersonaDetail npc={selectedRow} />
                </div>

                {isGenerating && (
                    <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl border border-base-200">
                        <span className="loading loading-spinner text-primary loading-lg"></span>
                        <div className="flex flex-col items-center gap-1">
                            <span className="font-bold text-lg text-base-content/70">マスターペルソナを一括生成中...</span>
                            <span className="text-sm text-base-content/50">選択されたJSONデータから全NPCのセリフを解析しています</span>
                        </div>
                    </div>
                )}
            </div>

            {/* 下部ステータスバー */}
            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: MasterPersonaGeneration ({isGenerating ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isGenerating ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={handleStart}
                        disabled={isGenerating || !jsonPath || activeTaskStatus === 'running'}
                    >
                        {isGenerating ? '生成中...' : '開始'}
                    </button>
                    {activeTaskId && activeTaskStatus !== 'running' && (
                        <button
                            className="btn btn-success btn-sm"
                            onClick={handleResumeCurrentTask}
                        >
                            再開
                        </button>
                    )}
                    {activeTaskId && activeTaskStatus === 'running' && (
                        <button
                            className="btn btn-warning btn-sm"
                            onClick={handlePauseCurrentTask}
                        >
                            一時停止
                        </button>
                    )}
                    <button className="btn btn-primary btn-sm" disabled={isGenerating}>
                        生成データを確定
                    </button>
                </div>
            </div>

            {/* ペルソナ確認モーダル */}
            <dialog id="persona_modal" className="modal">
                <div className="modal-box w-11/12 max-w-3xl border border-base-300">
                    <h3 className="font-bold text-lg">ペルソナ詳細確認: UlfricStormcloak (00013B9B)</h3>
                    <div className="py-4 flex flex-col gap-4">
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">要約 (Summary)</span></label>
                            <textarea className="textarea textarea-bordered h-24" readOnly value="ストームクロークの反乱軍のリーダー。誇り高く、ノルドの伝統を重んじる。ウィンドヘルムの首長であり、帝国に強い敵対心を抱いている。"></textarea>
                        </div>
                        <div className="form-control">
                            <label className="label"><span className="label-text font-bold">口調・一人称・二人称 (Tone/Pronouns)</span></label>
                            <textarea className="textarea textarea-bordered h-24" readOnly value={`一人称：「俺」\n二人称：「お前」「お前たち」\n口調：威厳があり、力強く、少し乱暴な言葉遣い。命令形をよく使う。`}></textarea>
                        </div>
                    </div>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-outline">再生成</button>
                                <button className="btn btn-primary">閉じる</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {/* 削除確認モーダル */}
            <dialog id="delete_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">削除の確認</h3>
                    <p className="py-4">このNPCデータを一覧およびデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-error">削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>
        </div>
    );
};

export default MasterPersona;
