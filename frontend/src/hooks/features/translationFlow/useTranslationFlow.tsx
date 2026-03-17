import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {useLocation} from 'react-router-dom';
import {ConfigGetAll, ConfigSet} from '../../../wailsjs/go/controller/ConfigController';
import {SelectTranslationInputFiles} from '../../../wailsjs/go/controller/FileDialogController';
import {
    GetAllTasks,
    GetTranslationFlowTerminology,
    ListLoadedTranslationFlowFiles,
    ListTranslationFlowPreviewRows,
    LoadTranslationFlowFiles,
    RunTranslationFlowTerminology,
} from '../../../wailsjs/go/controller/TaskController';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    type MasterPersonaLLMConfig,
    type MasterPersonaPromptConfig
} from '../../../types/masterPersona';
import {mapLoadResult, mapPreviewPage, mapTerminologyPhaseResult} from './adapters';
import type {TerminologyPhaseSummary, TranslationFlowTab, UseTranslationFlowResult} from './types';

const PREVIEW_PAGE_SIZE = 50;
const TERMINOLOGY_LLM_NAMESPACE = 'translation_flow.terminology.llm';
const TERMINOLOGY_PROMPT_NAMESPACE = 'translation_flow.terminology.prompt';
const TERMINOLOGY_USER_PROMPT_KEY = 'user_prompt';
const TERMINOLOGY_SYSTEM_PROMPT_KEY = 'system_prompt';
const DEFAULT_TERMINOLOGY_PROMPT_CONFIG: MasterPersonaPromptConfig = {
    userPrompt: 'Translate the provided term.',
    systemPrompt: `You are a translator for a Skyrim mod.
Record Type: {{.RecordType}}
Source File: {{.SourceFile}}
Editor ID: {{.EditorID}}

Please translate the following term into Japanese:
"{{.SourceText}}"

{{- if .ShortName }}
Short Name to also translate:
"{{.ShortName}}"
{{- end }}

Context/Reference Terms from Dictionary:
{{- range .ReferenceTerms }}
- {{.Source}}: {{.Translation}}
{{- else}}
None
{{- end }}

Requirements:
1. Translate the text idiomatically for Skyrim (e.g. Katakana for names, appropriate Kanji for titles).
2. Be consistent with the Reference Terms provided.
3. You MUST output the final translation in the following exact format and nothing else:
TL: |[translated_text]|`,
};
const EMPTY_TERMINOLOGY_SUMMARY: TerminologyPhaseSummary = {
    taskId: '',
    status: 'pending',
    targetCount: 0,
    savedCount: 0,
    failedCount: 0,
};

const TABS: TranslationFlowTab[] = [
    {label: 'データロード'},
    {label: '単語翻訳'},
    {label: 'ペルソナ生成'},
    {label: '要約'},
    {label: '本文翻訳'},
    {label: 'エクスポート'},
];

const TO_TASK_RESOLVE_ERROR = 'translation_project task の取得に失敗しました';

const toErrorMessage = (error: unknown, fallback: string): string => {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message;
    }
    if (typeof error === 'string' && error.trim() !== '') {
        return error;
    }
    if (error && typeof error === 'object') {
        const maybeMessage = (error as {message?: unknown}).message;
        if (typeof maybeMessage === 'string' && maybeMessage.trim() !== '') {
            return maybeMessage;
        }
    }
    return fallback;
};

const parseDate = (value: unknown): number => {
    if (typeof value !== 'string' || value.trim() === '') {
        return 0;
    }
    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? 0 : parsed;
};

const resolveTranslationProjectTaskID = (payload: unknown): string => {
    if (!Array.isArray(payload)) {
        return '';
    }

    const candidates = payload
        .filter((entry): entry is Record<string, unknown> => Boolean(entry) && typeof entry === 'object')
        .map((entry) => ({
            id: typeof entry.id === 'string' ? entry.id.trim() : '',
            type: typeof entry.type === 'string' ? entry.type : '',
            status: typeof entry.status === 'string' ? entry.status : '',
            updatedAt: parseDate(entry.updated_at ?? entry.updatedAt),
        }))
        .filter((entry) => entry.type === 'translation_project' && entry.id !== '');

    if (candidates.length === 0) {
        return '';
    }

    const active = candidates
        .filter((entry) => entry.status !== 'completed')
        .sort((a, b) => b.updatedAt - a.updatedAt);

    if (active.length > 0) {
        return active[0].id;
    }

    const sorted = [...candidates].sort((a, b) => b.updatedAt - a.updatedAt);
    return sorted[0]?.id ?? '';
};

const toTaskIDFromRoute = (value: unknown): string => {
    if (typeof value !== 'string') {
        return '';
    }
    return value.trim();
};

const resolveTaskID = async (routeTaskID: string): Promise<string> => {
    if (routeTaskID !== '') {
        return routeTaskID;
    }

    const allTasks = await GetAllTasks();
    return resolveTranslationProjectTaskID(allTasks);
};

const mergeUniquePaths = (base: string[], incoming: string[]): string[] => {
    const existing = new Set(base);
    const next = [...base];
    for (const path of incoming) {
        if (!existing.has(path)) {
            existing.add(path);
            next.push(path);
        }
    }
    return next;
};

const normalizeTerminologyLLMConfig = (loaded: Record<string, string>): MasterPersonaLLMConfig => {
    const temperature = Number.parseFloat(loaded.temperature ?? '');
    const contextLength = Number.parseInt(loaded.context_length ?? '', 10);
    const syncConcurrency = Number.parseInt(loaded.sync_concurrency ?? '', 10);
    const provider = loaded.provider;
    const bulkStrategy = String(loaded.bulk_strategy ?? '').trim().toLowerCase() === 'batch' ? 'batch' : 'sync';

    return {
        provider: provider === 'gemini' || provider === 'xai' || provider === 'lmstudio' ? provider : DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider,
        model: loaded.model ?? DEFAULT_MASTER_PERSONA_LLM_CONFIG.model,
        endpoint: loaded.endpoint || DEFAULT_MASTER_PERSONA_LLM_CONFIG.endpoint,
        apiKey: loaded.api_key ?? DEFAULT_MASTER_PERSONA_LLM_CONFIG.apiKey,
        temperature: Number.isFinite(temperature) ? temperature : DEFAULT_MASTER_PERSONA_LLM_CONFIG.temperature,
        contextLength: Number.isFinite(contextLength) && contextLength > 0 ? contextLength : DEFAULT_MASTER_PERSONA_LLM_CONFIG.contextLength,
        syncConcurrency: Number.isFinite(syncConcurrency) && syncConcurrency > 0 ? syncConcurrency : DEFAULT_MASTER_PERSONA_LLM_CONFIG.syncConcurrency,
        bulkStrategy,
    };
};

const normalizeTerminologyPromptConfig = (loaded: Record<string, string>): MasterPersonaPromptConfig => ({
    userPrompt: loaded[TERMINOLOGY_USER_PROMPT_KEY] ?? DEFAULT_TERMINOLOGY_PROMPT_CONFIG.userPrompt,
    systemPrompt: loaded[TERMINOLOGY_SYSTEM_PROMPT_KEY] ?? DEFAULT_TERMINOLOGY_PROMPT_CONFIG.systemPrompt,
});

const terminologyStatusLabel = (summary: TerminologyPhaseSummary, isRunning: boolean, progressLabel: string): string => {
    if (isRunning && progressLabel !== '') {
        return progressLabel;
    }
    switch (summary.status) {
        case 'completed':
            return summary.failedCount > 0 ? '単語翻訳完了（一部失敗あり）' : '単語翻訳完了';
        case 'running':
            return '単語翻訳を実行中';
        default:
            return '未実行';
    }
};

/**
 * TranslationFlow 画面のロードフェーズ状態を headless に管理する。
 */
export function useTranslationFlow(): UseTranslationFlowResult {
    const location = useLocation();
    const navState = location.state as {taskId?: string} | null;

    const routeTaskID = useMemo(() => toTaskIDFromRoute(navState?.taskId), [navState?.taskId]);

    const [taskId, setTaskID] = useState(routeTaskID);
    const [activeTab, setActiveTab] = useState(0);
    const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
    const [loadedFiles, setLoadedFiles] = useState<UseTranslationFlowResult['state']['loadedFiles']>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const [terminologySummary, setTerminologySummary] = useState<TerminologyPhaseSummary>(EMPTY_TERMINOLOGY_SUMMARY);
    const [terminologyErrorMessage, setTerminologyErrorMessage] = useState('');
    const [isTerminologyRunning, setIsTerminologyRunning] = useState(false);
    const [terminologyProgressLabel, setTerminologyProgressLabel] = useState('');
    const [terminologyConfig, setTerminologyConfig] = useState<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const [terminologyPromptConfig, setTerminologyPromptConfig] = useState<MasterPersonaPromptConfig>(DEFAULT_TERMINOLOGY_PROMPT_CONFIG);
    const [isTerminologyConfigHydrated, setIsTerminologyConfigHydrated] = useState(false);
    const [isTerminologyPromptHydrated, setIsTerminologyPromptHydrated] = useState(false);
    const llmSaveTimerRef = useRef<number | null>(null);
    const promptSaveTimerRef = useRef<number | null>(null);

    useEffect(() => {
        let active = true;

        const run = async () => {
            try {
                const resolved = await resolveTaskID(routeTaskID);
                if (!active) {
                    return;
                }
                setTaskID(resolved);
                if (resolved === '') {
                    setErrorMessage('');
                    return;
                }
                setErrorMessage('');
            } catch (error) {
                if (!active) {
                    return;
                }
                setErrorMessage(toErrorMessage(error, TO_TASK_RESOLVE_ERROR));
            }
        };

        void run();

        return () => {
            active = false;
        };
    }, [routeTaskID]);

    useEffect(() => {
        let alive = true;

        void (async () => {
            try {
                const loaded = await ConfigGetAll(TERMINOLOGY_LLM_NAMESPACE);
                if (!alive) {
                    return;
                }
                setTerminologyConfig(normalizeTerminologyLLMConfig(loaded));
            } finally {
                if (alive) {
                    setIsTerminologyConfigHydrated(true);
                }
            }
        })();

        void (async () => {
            try {
                const loaded = await ConfigGetAll(TERMINOLOGY_PROMPT_NAMESPACE);
                if (!alive) {
                    return;
                }
                setTerminologyPromptConfig(normalizeTerminologyPromptConfig(loaded));
            } finally {
                if (alive) {
                    setIsTerminologyPromptHydrated(true);
                }
            }
        })();

        return () => {
            alive = false;
        };
    }, []);

    useEffect(() => {
        if (!isTerminologyConfigHydrated) {
            return;
        }
        if (llmSaveTimerRef.current) {
            window.clearTimeout(llmSaveTimerRef.current);
        }
        llmSaveTimerRef.current = window.setTimeout(() => {
            void Promise.all([
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'provider', terminologyConfig.provider),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'model', terminologyConfig.model),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'endpoint', terminologyConfig.endpoint),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'api_key', terminologyConfig.provider === 'lmstudio' ? '' : terminologyConfig.apiKey),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'temperature', String(terminologyConfig.temperature)),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'context_length', String(terminologyConfig.contextLength)),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'sync_concurrency', String(terminologyConfig.syncConcurrency)),
                ConfigSet(TERMINOLOGY_LLM_NAMESPACE, 'bulk_strategy', terminologyConfig.bulkStrategy),
            ]);
        }, 250);

        return () => {
            if (llmSaveTimerRef.current) {
                window.clearTimeout(llmSaveTimerRef.current);
            }
        };
    }, [isTerminologyConfigHydrated, terminologyConfig]);

    useEffect(() => {
        if (!isTerminologyPromptHydrated) {
            return;
        }
        if (promptSaveTimerRef.current) {
            window.clearTimeout(promptSaveTimerRef.current);
        }
        promptSaveTimerRef.current = window.setTimeout(() => {
            void Promise.all([
                ConfigSet(TERMINOLOGY_PROMPT_NAMESPACE, TERMINOLOGY_USER_PROMPT_KEY, terminologyPromptConfig.userPrompt),
                ConfigSet(TERMINOLOGY_PROMPT_NAMESPACE, TERMINOLOGY_SYSTEM_PROMPT_KEY, terminologyPromptConfig.systemPrompt),
            ]);
        }, 250);

        return () => {
            if (promptSaveTimerRef.current) {
                window.clearTimeout(promptSaveTimerRef.current);
            }
        };
    }, [isTerminologyPromptHydrated, terminologyPromptConfig]);

    const handleRefreshTerminologyPhase = useCallback(async () => {
        if (taskId === '') {
            setTerminologySummary(EMPTY_TERMINOLOGY_SUMMARY);
            return;
        }
        try {
            const payload = await GetTranslationFlowTerminology(taskId);
            const summary = mapTerminologyPhaseResult(payload);
            setTerminologySummary({
                ...summary,
                taskId: summary.taskId || taskId,
            });
        } catch (error) {
            setTerminologyErrorMessage(toErrorMessage(error, '単語翻訳の状態取得に失敗しました'));
        }
    }, [taskId]);

    const handleReloadFiles = useCallback(async () => {
        if (taskId === '') {
            setLoadedFiles([]);
            return;
        }
        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await ListLoadedTranslationFlowFiles(taskId);
            const mapped = mapLoadResult(payload);
            if (mapped.taskId !== '' && mapped.taskId !== taskId) {
                setTaskID(mapped.taskId);
            }
            setLoadedFiles(mapped.files);
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ロード済みファイルの取得に失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, [taskId]);

    useEffect(() => {
        void handleReloadFiles();
        void handleRefreshTerminologyPhase();
    }, [handleRefreshTerminologyPhase, handleReloadFiles]);

    const handleSelectFiles = useCallback(async () => {
        setErrorMessage('');
        try {
            const files = await SelectTranslationInputFiles();
            if (!Array.isArray(files) || files.length === 0) {
                return;
            }
            setSelectedFiles((prev) => mergeUniquePaths(prev, files));
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイル選択に失敗しました'));
        }
    }, []);

    const handleRemoveFile = useCallback((pathToRemove: string) => {
        setSelectedFiles((prev) => prev.filter((path) => path !== pathToRemove));
    }, []);

    const handleLoadSelectedFiles = useCallback(async () => {
        if (selectedFiles.length === 0 || taskId === '') {
            return;
        }

        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await LoadTranslationFlowFiles(taskId, selectedFiles);
            const mapped = mapLoadResult(payload);
            if (mapped.taskId !== '' && mapped.taskId !== taskId) {
                setTaskID(mapped.taskId);
            }
            setLoadedFiles(mapped.files);
            setSelectedFiles([]);
            setTerminologySummary({...EMPTY_TERMINOLOGY_SUMMARY, taskId});
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイルロードに失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, [selectedFiles, taskId]);

    const handlePreviewPageChange = useCallback(async (fileId: number, page: number) => {
        if (fileId <= 0) {
            return;
        }

        const safePage = Math.max(1, page);
        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await ListTranslationFlowPreviewRows(fileId, safePage, PREVIEW_PAGE_SIZE);
            const mappedPage = mapPreviewPage(payload);
            setLoadedFiles((prev) =>
                prev.map((file) => {
                    if (file.fileId !== fileId) {
                        return file;
                    }
                    return {
                        ...file,
                        currentPage: mappedPage.page,
                        pageSize: mappedPage.pageSize,
                        totalRows: mappedPage.totalRows,
                        rows: mappedPage.rows,
                    };
                }),
            );
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'プレビューのページ切り替えに失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, []);

    const handleAdvanceFromLoad = useCallback(() => {
        if (loadedFiles.length === 0) {
            return;
        }
        setActiveTab(1);
    }, [loadedFiles.length]);

    const handleRunTerminologyPhase = useCallback(async () => {
        if (taskId === '' || isTerminologyRunning) {
            return;
        }
        setIsTerminologyRunning(true);
        setTerminologyErrorMessage('');
        setTerminologyProgressLabel('terminology ジョブを生成中');
        setTerminologySummary((prev) => ({
            ...prev,
            taskId: taskId,
            status: 'running',
        }));

        try {
            await Promise.resolve();
            setTerminologyProgressLabel('LLM で単語翻訳を実行中');
            const payload = await RunTranslationFlowTerminology(
                taskId,
                {
                    provider: terminologyConfig.provider,
                    model: terminologyConfig.model,
                    endpoint: terminologyConfig.endpoint,
                    api_key: terminologyConfig.provider === 'lmstudio' ? '' : terminologyConfig.apiKey,
                    temperature: terminologyConfig.temperature,
                    context_length: terminologyConfig.contextLength,
                    sync_concurrency: terminologyConfig.syncConcurrency,
                    bulk_strategy: terminologyConfig.bulkStrategy,
                },
                {
                    user_prompt: terminologyPromptConfig.userPrompt,
                    system_prompt: terminologyPromptConfig.systemPrompt,
                },
            );
            setTerminologyProgressLabel('翻訳結果を保存中');
            const summary = mapTerminologyPhaseResult(payload);
            setTerminologySummary({
                ...summary,
                taskId: summary.taskId || taskId,
            });
            setTerminologyProgressLabel('単語翻訳完了');
        } catch (error) {
            setTerminologyErrorMessage(toErrorMessage(error, '単語翻訳の実行に失敗しました'));
            setTerminologyProgressLabel('単語翻訳に失敗しました');
            setTerminologySummary((prev) => ({
                ...prev,
                taskId: taskId,
                status: 'pending',
            }));
        } finally {
            setIsTerminologyRunning(false);
        }
    }, [isTerminologyRunning, taskId, terminologyConfig, terminologyPromptConfig]);

    const handleAdvanceFromTerminology = useCallback(() => {
        if (terminologySummary.status !== 'completed') {
            return;
        }
        setActiveTab(2);
    }, [terminologySummary.status]);

    const handleTerminologyConfigChange = useCallback((next: MasterPersonaLLMConfig) => {
        setTerminologyConfig(next);
    }, []);

    const handleTerminologyPromptChange = useCallback((next: MasterPersonaPromptConfig) => {
        setTerminologyPromptConfig(next);
    }, []);

    const handleTabChange = useCallback((index: number) => {
        if (index < 0 || index >= TABS.length) {
            return;
        }
        if (index > 0 && loadedFiles.length === 0) {
            return;
        }
        if (index > 1 && terminologySummary.status !== 'completed') {
            return;
        }
        setActiveTab(index);
    }, [loadedFiles.length, terminologySummary.status]);

    return {
        state: {
            taskId,
            activeTab,
            tabs: TABS,
            selectedFiles,
            loadedFiles,
            isLoading,
            errorMessage,
            terminologySummary,
            terminologyStatusLabel: terminologyStatusLabel(terminologySummary, isTerminologyRunning, terminologyProgressLabel),
            terminologyErrorMessage,
            isTerminologyRunning,
            terminologyConfig,
            terminologyPromptConfig,
            isTerminologyConfigHydrated,
            isTerminologyPromptHydrated,
        },
        actions: {
            handleTabChange,
            handleSelectFiles,
            handleRemoveFile,
            handleLoadSelectedFiles,
            handleReloadFiles,
            handlePreviewPageChange,
            handleAdvanceFromLoad,
            handleRunTerminologyPhase,
            handleRefreshTerminologyPhase,
            handleTerminologyConfigChange,
            handleTerminologyPromptChange,
            handleAdvanceFromTerminology,
        },
        ui: {
            previewPageSize: PREVIEW_PAGE_SIZE,
        },
    };
}
