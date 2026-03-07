import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import type { ColumnDef } from '@tanstack/react-table';
import ModelSettings from '../components/ModelSettings';
import DataTable from '../components/DataTable';
import PersonaDetail from '../components/PersonaDetail';
import PromptSettingCard from '../components/masterPersona/PromptSettingCard';
import { type NpcRow, type NpcStatus, STATUS_BADGE } from '../types/npc';
import { SelectJSONFile } from '../wailsjs/go/main/App';
import { CancelTask, GetAllTasks, GetTaskRequestState, ResumeTask, StartMasterPersonTask } from '../wailsjs/go/task/Bridge';
import { ListDialoguesByPersonaID, ListNPCs } from '../wailsjs/go/persona/Service';
import { ConfigGetAll, ConfigSet } from '../wailsjs/go/config/ConfigService';
import type { PhaseCompletedEvent, FrontendTask } from '../types/task';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    DEFAULT_MASTER_PERSONA_PROMPT_CONFIG,
    type MasterPersonaLLMConfig,
    type MasterPersonaPromptConfig,
} from '../types/masterPersona';
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

interface PersonaNPCRecord {
    persona_id?: number;
    PersonaID?: number;
    speaker_id?: string;
    SpeakerID?: string;
    source_plugin?: string;
    SourcePlugin?: string;
    npc_name?: string;
    NPCName?: string;
    race?: string;
    Race?: string;
    sex?: string;
    Sex?: string;
    voice_type?: string;
    VoiceType?: string;
    dialogue_count?: number;
    DialogueCount?: number;
    persona_text?: string;
    PersonaText?: string;
    updated_at?: string;
    UpdatedAt?: string;
}

interface PersonaDialogueRecord {
    persona_id?: number;
    PersonaID?: number;
    record_type?: string;
    RecordType?: string;
    editor_id?: string;
    EditorID?: string;
    source_text?: string;
    SourceText?: string;
}

const pickString = (value: unknown): string => {
    if (typeof value === 'string') {
        return value;
    }
    return '';
};

const formatUpdatedAt = (raw: string): string => {
    const ts = Date.parse(raw);
    if (!Number.isFinite(ts)) {
        return '';
    }
    return new Date(ts).toLocaleString('ja-JP');
};

const MASTER_PERSONA_LLM_NAMESPACE = 'master_persona.llm';
const MASTER_PERSONA_PROMPT_NAMESPACE = 'master_persona.prompt';
const SELECTED_PROVIDER_KEY = 'selected_provider';
const USER_PROMPT_KEY = 'user_prompt';
const SYSTEM_PROMPT_KEY = 'system_prompt';

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

const buildPromptConfigPairs = (cfg: MasterPersonaPromptConfig): Record<string, string> => ({
    [USER_PROMPT_KEY]: cfg.userPrompt,
    [SYSTEM_PROMPT_KEY]: cfg.systemPrompt,
});

const parseTaskTimestamp = (value: string | undefined): number => {
    if (!value) {
        return 0;
    }
    const t = Date.parse(value);
    return Number.isFinite(t) ? t : 0;
};

const PERSONA_PAGE_SIZE = 100;

// ── 列定義 ───────────────────────────────────────────────
const NPC_COLUMNS: ColumnDef<NpcRow, unknown>[] = [
    {
        accessorKey: 'formId',
        header: 'FormID',
        cell: (info) => <span className="font-mono text-sm">{info.getValue() as string}</span>,
    },
    {
        accessorKey: 'sourcePlugin',
        header: 'プラグイン名',
        cell: (info) => <span className="font-mono text-xs">{info.getValue() as string}</span>,
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
    const [allNpcData, setAllNpcData] = useState<NpcRow[]>([]);
    const [selectedRow, setSelectedRow] = useState<NpcRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [npcSearchInput, setNpcSearchInput] = useState<string>('');
    const [pluginFilterInput, setPluginFilterInput] = useState<string>('');
    const [appliedNpcSearch, setAppliedNpcSearch] = useState<string>('');
    const [appliedPluginFilter, setAppliedPluginFilter] = useState<string>('');
    const [npcPage, setNpcPage] = useState<number>(1);
    const [isGenerating, setIsGenerating] = useState<boolean>(false);
    const [jsonPath, setJsonPath] = useState<string>('');
    const [overwriteExisting, setOverwriteExisting] = useState<boolean>(false);
    const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
    const [progressPercent, setProgressPercent] = useState<number>(0);
    const [statusMessage, setStatusMessage] = useState<string>('待機中');
    const [errorMessage, setErrorMessage] = useState<string>('');
    const [progressCounts, setProgressCounts] = useState<{ current: number; total: number }>({ current: 0, total: 0 });
    const [activeTaskStatus, setActiveTaskStatus] = useState<FrontendTask['status'] | null>(null);
    const [llmConfig, setLLMConfig] = useState<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const [isLLMConfigHydrated, setIsLLMConfigHydrated] = useState<boolean>(false);
    const [promptConfig, setPromptConfig] = useState<MasterPersonaPromptConfig>(DEFAULT_MASTER_PERSONA_PROMPT_CONFIG);
    const [isPromptConfigHydrated, setIsPromptConfigHydrated] = useState<boolean>(false);
    const lastSavedLLMConfigRef = useRef<Partial<Record<MasterPersonaLLMConfig['provider'], MasterPersonaLLMConfig>>>({});
    const latestLLMConfigRef = useRef<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const lastSavedPromptConfigRef = useRef<MasterPersonaPromptConfig>(DEFAULT_MASTER_PERSONA_PROMPT_CONFIG);
    const latestPromptConfigRef = useRef<MasterPersonaPromptConfig>(DEFAULT_MASTER_PERSONA_PROMPT_CONFIG);
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

    const loadPromptConfig = async (): Promise<MasterPersonaPromptConfig> => {
        const loaded = await ConfigGetAll(MASTER_PERSONA_PROMPT_NAMESPACE);
        return {
            userPrompt: loaded[USER_PROMPT_KEY] ?? DEFAULT_MASTER_PERSONA_PROMPT_CONFIG.userPrompt,
            systemPrompt: loaded[SYSTEM_PROMPT_KEY] ?? DEFAULT_MASTER_PERSONA_PROMPT_CONFIG.systemPrompt,
        };
    };

    const handleRowSelect = (row: NpcRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
    };

    const pluginOptions = useMemo(() => {
        const unique = new Set<string>();
        for (const row of allNpcData) {
            if (row.sourcePlugin.trim() !== '') {
                unique.add(row.sourcePlugin);
            }
        }
        return Array.from(unique).sort((a, b) => a.localeCompare(b, 'ja'));
    }, [allNpcData]);

    const filteredNpcData = useMemo(() => {
        const keyword = appliedNpcSearch.trim().toLowerCase();
        const plugin = appliedPluginFilter.trim().toLowerCase();
        return allNpcData.filter((row) => {
            if (plugin !== '' && row.sourcePlugin.toLowerCase() !== plugin) {
                return false;
            }
            if (keyword === '') {
                return true;
            }
            return [
                row.formId,
                row.sourcePlugin,
                row.name,
                row.race,
                row.sex,
                row.voiceType,
                row.personaText,
            ].some((value) => value.toLowerCase().includes(keyword));
        });
    }, [allNpcData, appliedNpcSearch, appliedPluginFilter]);

    const totalNpcPages = Math.max(1, Math.ceil(filteredNpcData.length / PERSONA_PAGE_SIZE));
    const pagedNpcData = useMemo(() => {
        const start = (npcPage - 1) * PERSONA_PAGE_SIZE;
        return filteredNpcData.slice(start, start + PERSONA_PAGE_SIZE);
    }, [filteredNpcData, npcPage]);

    const applyNPCFilters = () => {
        setAppliedNpcSearch(npcSearchInput);
        setAppliedPluginFilter(pluginFilterInput);
        setNpcPage(1);
    };

    const clearNPCFilters = () => {
        setNpcSearchInput('');
        setPluginFilterInput('');
        setAppliedNpcSearch('');
        setAppliedPluginFilter('');
        setNpcPage(1);
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
        setProgressPercent(0);
        setProgressCounts({ current: 0, total: 0 });
        setStatusMessage('タスクを開始しています...');
        resumeRequestedRef.current = false;
        setActiveTaskStatus('pending');

        try {
            const taskID = await StartMasterPersonTask({ source_json_path: jsonPath, overwrite_existing: overwriteExisting });
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
        const resumeCursor = Number(task.metadata?.resume_cursor ?? 0);
        setProgressCounts({
            current: Number.isFinite(resumeCursor) ? resumeCursor : 0,
            total: requestCount > 0 ? requestCount : 0,
        });
        const sourceJSONPath = String(task.metadata?.source_json_path ?? '');
        if (sourceJSONPath) {
            setJsonPath(sourceJSONPath);
        }
        setOverwriteExisting(Boolean(task.metadata?.overwrite_existing));
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

    useEffect(() => {
        latestPromptConfigRef.current = promptConfig;
    }, [promptConfig]);

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

    const persistPromptConfigDiff = (current: MasterPersonaPromptConfig) => {
        const currentPairs = buildPromptConfigPairs(current);
        const previousPairs = buildPromptConfigPairs(lastSavedPromptConfigRef.current);
        const writes: Array<[string, string]> = [];

        for (const key of Object.keys(currentPairs)) {
            const typedKey = key as keyof typeof currentPairs;
            if (previousPairs[typedKey] !== currentPairs[typedKey]) {
                writes.push([typedKey, currentPairs[typedKey]]);
            }
        }

        if (writes.length === 0) {
            return Promise.resolve();
        }

        return Promise.all(
            writes.map(([key, value]) => ConfigSet(MASTER_PERSONA_PROMPT_NAMESPACE, key, value)),
        ).then(() => {
            lastSavedPromptConfigRef.current = current;
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
        let alive = true;
        void loadPromptConfig()
            .then((loaded) => {
                if (!alive) {
                    return;
                }
                setPromptConfig(loaded);
                lastSavedPromptConfigRef.current = loaded;
                latestPromptConfigRef.current = loaded;
                setIsPromptConfigHydrated(true);
            })
            .catch((error) => {
                console.error('failed to hydrate master_persona.prompt config', error);
                if (!alive) {
                    return;
                }
                setPromptConfig(DEFAULT_MASTER_PERSONA_PROMPT_CONFIG);
                lastSavedPromptConfigRef.current = DEFAULT_MASTER_PERSONA_PROMPT_CONFIG;
                latestPromptConfigRef.current = DEFAULT_MASTER_PERSONA_PROMPT_CONFIG;
                setIsPromptConfigHydrated(true);
            });
        return () => {
            alive = false;
        };
    }, []);

    useEffect(() => {
        if (!isPromptConfigHydrated) {
            return;
        }
        const snapshot = latestPromptConfigRef.current;
        saveQueueRef.current = saveQueueRef.current
            .then(() => persistPromptConfigDiff(snapshot))
            .catch((err) => {
                console.error('failed to persist master_persona.prompt config', err);
            });
    }, [isPromptConfigHydrated, promptConfig]);

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

    useEffect(() => {
        return () => {
            if (!isPromptConfigHydrated) {
                return;
            }
            const snapshot = latestPromptConfigRef.current;
            saveQueueRef.current = saveQueueRef.current
                .then(() => persistPromptConfigDiff(snapshot))
                .catch((err) => {
                    console.error('failed to flush master_persona.prompt config on unmount', err);
                });
        };
    }, [isPromptConfigHydrated]);

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

    const refreshNPCDataFromService = async () => {
        try {
            const records = (await ListNPCs() as unknown as PersonaNPCRecord[]) || [];
            const rows: NpcRow[] = records
                .map((record) => {
                    const speakerID = pickString(record.speaker_id ?? record.SpeakerID);
                    const personaID = Number(record.persona_id ?? record.PersonaID ?? 0);
                    const npcName = pickString(record.npc_name ?? record.NPCName);
                    const updatedAt = formatUpdatedAt(pickString(record.updated_at ?? record.UpdatedAt));
                    const dialogueCount = Number(record.dialogue_count ?? record.DialogueCount ?? 0);
                    return {
                        id: String(personaID),
                        personaId: personaID,
                        formId: speakerID,
                        sourcePlugin: pickString(record.source_plugin ?? record.SourcePlugin),
                        name: npcName || 'Unknown NPC',
                        race: pickString(record.race ?? record.Race),
                        sex: pickString(record.sex ?? record.Sex),
                        voiceType: pickString(record.voice_type ?? record.VoiceType),
                        dialogueCount: Number.isFinite(dialogueCount) ? dialogueCount : 0,
                        status: '完了' as NpcStatus,
                        updatedAt,
                        personaText: pickString(record.persona_text ?? record.PersonaText),
                        rawResponse: '',
                        dialogues: [],
                    };
                })
                .filter((row) => row.personaId > 0);
            setAllNpcData(rows);
            if (rows.length === 0) {
                setSelectedRow(null);
                setSelectedRowId(null);
            } else if (!rows.some((row) => row.id === selectedRowId)) {
                setSelectedRow(rows[0]);
                setSelectedRowId(rows[0].id);
            }
        } catch (error) {
            console.error('failed to refresh npc rows from persona service', { error });
        }
    };

    const loadDialoguesForPersona = async (personaID: number) => {
        try {
            const records = (await ListDialoguesByPersonaID(personaID) as unknown as PersonaDialogueRecord[]) || [];
            const dialogues = records.map((row) => ({
                recordType: pickString(row.record_type ?? row.RecordType),
                editorId: pickString(row.editor_id ?? row.EditorID),
                source: pickString(row.source_text ?? row.SourceText),
            }));
            setAllNpcData((prev) =>
                prev.map((row) => (row.personaId === personaID ? { ...row, dialogues } : row)),
            );
            setSelectedRow((prev) => {
                if (!prev || prev.personaId !== personaID) {
                    return prev;
                }
                return { ...prev, dialogues };
            });
        } catch (error) {
            console.error('failed to load dialogues from persona service', { personaID, error });
        }
    };

    const resetTaskView = () => {
        setIsGenerating(false);
        setActiveTaskStatus(null);
        setActiveTaskId(null);
        setStatusMessage('待機中');
        resumeRequestedRef.current = false;
    };

    useEffect(() => {
        if (!selectedRowId) {
            return;
        }
        const personaID = Number.parseInt(selectedRowId, 10);
        if (Number.isFinite(personaID) && personaID > 0) {
            void loadDialoguesForPersona(personaID);
        }
    }, [selectedRowId]);

    useEffect(() => {
        if (npcPage > totalNpcPages) {
            setNpcPage(totalNpcPages);
        }
    }, [npcPage, totalNpcPages]);

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
                return Promise.all([
                    refreshProgressFromQueueState(task),
                    refreshNPCDataFromService(),
                ]).then(() => undefined);
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
                void refreshNPCDataFromService();
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
                void refreshNPCDataFromService();
            }
            if (task.status === 'failed') {
                setIsGenerating(false);
                setErrorMessage(task.error_msg || 'タスク実行に失敗しました');
            }
            if (task.status === 'completed') {
                setStatusMessage('処理完了');
                setErrorMessage('');
                resetTaskView();
            }
            if (task.status === 'paused' || task.status === 'cancelled') {
                setIsGenerating(false);
                setStatusMessage('一時停止中');
            }
            if (task.status === 'running') {
                setIsGenerating(true);
                setStatusMessage('リクエストを実行しています...');
            }
            if (task.status === 'request_generated') {
                setStatusMessage('リクエスト生成完了。実行を開始します...');
                startQueuedExecution(task.id);
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
            <div className="grid grid-cols-1 gap-4 shrink-0">
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
                            <label className="label cursor-pointer justify-start gap-3">
                                <input
                                    type="checkbox"
                                    className="checkbox checkbox-primary checkbox-sm"
                                    checked={overwriteExisting}
                                    onChange={(event) => setOverwriteExisting(event.target.checked)}
                                    disabled={isGenerating}
                                />
                                <span className="label-text">重複時に既存ペルソナを上書きする</span>
                            </label>
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
            </div>

            {/* Prompt 設定 */}
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 shrink-0">
                {isPromptConfigHydrated ? (
                    <>
                        <PromptSettingCard
                            title="ユーザープロンプト"
                            description="可変の指示だけを編集します。NPC メタデータや会話履歴は送信時に別途付与されます。"
                            value={promptConfig.userPrompt}
                            onChange={(value) => setPromptConfig((prev) => ({ ...prev, userPrompt: value }))}
                            badgeLabel="編集可能"
                        />
                        <PromptSettingCard
                            title="システムプロンプト"
                            description="固定の分析ルールと出力形式です。画面表示と実際の送信内容は同じ system prompt を参照します。"
                            value={promptConfig.systemPrompt}
                            readOnly
                            badgeLabel="Read Only"
                        />
                    </>
                ) : (
                    <div className="card bg-base-100 border border-base-200 shadow-sm xl:col-span-2">
                        <div className="card-body py-4">
                            <span className="text-sm text-base-content/60">プロンプト設定を読み込み中...</span>
                        </div>
                    </div>
                )}
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
                        data={pagedNpcData}
                        title="NPC処理ステータス (Skyrim.esm)"
                        headerActions={
                            <div className="flex flex-wrap items-center gap-2">
                                <input
                                    type="text"
                                    className="input input-bordered input-xs w-44"
                                    placeholder="NPC / FormID / ペルソナ検索"
                                    value={npcSearchInput}
                                    onChange={(event) => setNpcSearchInput(event.target.value)}
                                />
                                <select
                                    className="select select-bordered select-xs w-40"
                                    value={pluginFilterInput}
                                    onChange={(event) => setPluginFilterInput(event.target.value)}
                                >
                                    <option value="">全プラグイン</option>
                                    {pluginOptions.map((plugin) => (
                                        <option key={plugin} value={plugin}>{plugin}</option>
                                    ))}
                                </select>
                                <button className="btn btn-primary btn-xs" onClick={applyNPCFilters}>
                                    検索
                                </button>
                                <button className="btn btn-ghost btn-xs" onClick={clearNPCFilters}>
                                    解除
                                </button>
                                <span className="text-xs text-base-content/60">
                                    {filteredNpcData.length.toLocaleString()} 件 / {npcPage} / {totalNpcPages} ページ
                                </span>
                                <button
                                    className="btn btn-outline btn-xs"
                                    disabled={npcPage <= 1}
                                    onClick={() => setNpcPage((prev) => Math.max(1, prev - 1))}
                                >
                                    前へ
                                </button>
                                <button
                                    className="btn btn-outline btn-xs"
                                    disabled={npcPage >= totalNpcPages}
                                    onClick={() => setNpcPage((prev) => Math.min(totalNpcPages, prev + 1))}
                                >
                                    次へ
                                </button>
                            </div>
                        }
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelect}
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
