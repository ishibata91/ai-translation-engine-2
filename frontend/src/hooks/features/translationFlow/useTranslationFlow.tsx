import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {useLocation} from 'react-router-dom';
import {ConfigGetAll, ConfigSet} from '../../../wailsjs/go/controller/ConfigController';
import {SelectTranslationInputFiles} from '../../../wailsjs/go/controller/FileDialogController';
import {useWailsEvent} from '../../useWailsEvent';
import {
    GetAllTasks,
    GetTranslationFlowTerminology,
    ListLoadedTranslationFlowFiles,
    ListTranslationFlowPreviewRows,
    ListTranslationFlowTerminologyTargets,
    LoadTranslationFlowFiles,
    RunTranslationFlowTerminology,
} from '../../../wailsjs/go/controller/TaskController';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    type MasterPersonaLLMConfig,
    type MasterPersonaPromptConfig,
} from '../../../types/masterPersona';
import {
    mapLoadResult,
    mapPersonaPhaseResult,
    mapPersonaTargetPreviewPage,
    mapPreviewPage,
    mapTerminologyPhaseResult,
    mapTerminologyTargetPreviewPage,
} from './adapters';
import type {
    PersonaPhaseSummary,
    PersonaTargetPreviewPage,
    PersonaTargetPreviewRow,
    PersonaTargetViewState,
    TerminologyPhaseSummary,
    TerminologyTargetPreviewPage,
    TerminologyTargetViewState,
    TranslationFlowTab,
    UseTranslationFlowResult,
    WailsTerminologyProgressEvent,
} from './types';

const PREVIEW_PAGE_SIZE = 50;
const TERMINOLOGY_LLM_NAMESPACE = 'translation_flow.terminology.llm';
const PERSONA_LLM_NAMESPACE = 'translation_flow.persona.llm';
const TERMINOLOGY_SELECTED_PROVIDER_KEY = 'selected_provider';
const TERMINOLOGY_PROMPT_NAMESPACE = 'translation_flow.terminology.prompt';
const PERSONA_PROMPT_NAMESPACE = 'translation_flow.persona.prompt';
const TERMINOLOGY_USER_PROMPT_KEY = 'user_prompt';
const TERMINOLOGY_SYSTEM_PROMPT_KEY = 'system_prompt';
const TERMINOLOGY_PROGRESS_EVENT = 'translation_flow.terminology.progress';
const DUPLICATE_FILE_MESSAGE = '同じプラグインを2重で処理しないため、同名ファイルは追加できません。';
const NO_TERMINOLOGY_TARGETS_MESSAGE = 'ロード済みデータに Terminology 対象 REC がありません。';

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
TL: |translated_text|`,
};

const EMPTY_TERMINOLOGY_SUMMARY: TerminologyPhaseSummary = {
    taskId: '',
    status: 'pending',
    savedCount: 0,
    failedCount: 0,
    progressMode: 'hidden',
    progressCurrent: 0,
    progressTotal: 0,
    progressMessage: '',
};

const EMPTY_TERMINOLOGY_TARGET_PAGE = (
    taskId = '',
    page = 1,
    pageSize = PREVIEW_PAGE_SIZE,
): TerminologyTargetPreviewPage => ({
    taskId,
    page,
    pageSize,
    totalRows: 0,
    rows: [],
});

const EMPTY_PERSONA_SUMMARY: PersonaPhaseSummary = {
    taskId: '',
    status: 'loadingTargets',
    detectedCount: 0,
    reusedCount: 0,
    pendingCount: 0,
    generatedCount: 0,
    failedCount: 0,
    progressMode: 'hidden',
    progressCurrent: 0,
    progressTotal: 0,
    progressMessage: '',
};

const EMPTY_PERSONA_TARGET_PAGE = (
    taskId = '',
    page = 1,
    pageSize = PREVIEW_PAGE_SIZE,
): PersonaTargetPreviewPage => ({
    taskId,
    page,
    pageSize,
    totalRows: 0,
    rows: [],
});

const TABS: TranslationFlowTab[] = [
    {label: 'データロード'},
    {label: '単語翻訳'},
    {label: 'ペルソナ生成'},
    {label: '要約'},
    {label: '本文翻訳'},
    {label: 'エクスポート'},
];

interface PersonaHookState {
    personaSummary: PersonaPhaseSummary;
    personaStatusLabel: string;
    personaErrorMessage: string;
    personaTargetPage: PersonaTargetPreviewPage;
    personaTargetStatus: PersonaTargetViewState;
    personaTargetErrorMessage: string;
    isPersonaTargetLoading: boolean;
    isPersonaRunning: boolean;
    selectedPersonaTargetKey: string;
    selectedPersonaTarget: PersonaTargetPreviewRow | null;
}

interface PersonaHookActions {
    handleRunPersonaPhase: () => Promise<void>;
    handleRetryPersonaPhase: () => Promise<void>;
    handleRefreshPersonaPhase: () => Promise<void>;
    handlePersonaTargetPageChange: (page: number) => Promise<void>;
    handleSelectPersonaTarget: (sourcePlugin: string, speakerId: string) => void;
    handleAdvanceFromPersona: () => void;
}

type UseTranslationFlowWithPersonaResult = UseTranslationFlowResult & {
    state: UseTranslationFlowResult['state'] & PersonaHookState;
    actions: UseTranslationFlowResult['actions'] & PersonaHookActions;
};

const TO_TASK_RESOLVE_ERROR = 'translation_project task の取得に失敗しました';
const NO_PERSONA_TARGETS_MESSAGE = 'ペルソナ生成対象 NPC がありません。';

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

const pickProgressEventString = (value: unknown, fallback = ''): string =>
    typeof value === 'string' ? value : fallback;

const pickProgressEventNumber = (value: unknown, fallback = 0): number =>
    typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const parseDate = (value: unknown): number => {
    if (typeof value !== 'string' || value.trim() === '') {
        return 0;
    }
    const parsed = Date.parse(value);
    return Number.isNaN(parsed) ? 0 : parsed;
};

const normalizeFileName = (path: string): string => {
    const trimmed = path.trim();
    if (trimmed === '') {
        return '';
    }
    const parts = trimmed.split(/[\\/]/);
    return (parts[parts.length - 1] ?? trimmed).trim().toLowerCase();
};

const mergeUniqueFilesByName = (
    base: string[],
    incoming: string[],
    loadedFileNames: Set<string>,
): {files: string[]; duplicateBlocked: boolean} => {
    const existingNames = new Set(base.map(normalizeFileName).filter((value) => value !== ''));
    const next = [...base];
    let duplicateBlocked = false;

    for (const path of incoming) {
        const fileName = normalizeFileName(path);
        if (fileName === '') {
            continue;
        }
        if (existingNames.has(fileName) || loadedFileNames.has(fileName)) {
            duplicateBlocked = true;
            continue;
        }
        existingNames.add(fileName);
        next.push(path);
    }

    return {files: next, duplicateBlocked};
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

const hasTaskID = (payload: unknown, taskID: string): boolean => {
    if (taskID === '' || !Array.isArray(payload)) {
        return false;
    }

    return payload.some((entry) => {
        if (!entry || typeof entry !== 'object') {
            return false;
        }
        return typeof (entry as {id?: unknown}).id === 'string' && (entry as {id: string}).id.trim() === taskID;
    });
};

const resolveTaskID = async (routeTaskID: string): Promise<string> => {
    const allTasks = await GetAllTasks();
    if (hasTaskID(allTasks, routeTaskID)) {
        return routeTaskID;
    }
    return resolveTranslationProjectTaskID(allTasks);
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

type TaskControllerPersonaBinding = {
    ListTranslationFlowPersonaTargets?: (taskID: string, page: number, pageSize: number) => Promise<unknown>;
    RunTranslationFlowPersona?: (taskID: string, request: unknown, prompt: unknown) => Promise<unknown>;
    GetTranslationFlowPersona?: (taskID: string) => Promise<unknown>;
};

const getTaskControllerPersonaBinding = (): TaskControllerPersonaBinding | null => {
    const goBridge = (window as unknown as {
        go?: {
            controller?: {
                TaskController?: TaskControllerPersonaBinding;
            };
        };
    }).go;
    return goBridge?.controller?.TaskController ?? null;
};

const listTranslationFlowPersonaTargetsBinding = async (
    taskID: string,
    page: number,
    pageSize: number,
): Promise<unknown> => {
    const binding = getTaskControllerPersonaBinding();
    if (!binding?.ListTranslationFlowPersonaTargets) {
        throw new Error('ListTranslationFlowPersonaTargets binding is not available');
    }
    return binding.ListTranslationFlowPersonaTargets(taskID, page, pageSize);
};

const runTranslationFlowPersonaBinding = async (
    taskID: string,
    request: unknown,
    prompt: unknown,
): Promise<unknown> => {
    const binding = getTaskControllerPersonaBinding();
    if (!binding?.RunTranslationFlowPersona) {
        throw new Error('RunTranslationFlowPersona binding is not available');
    }
    return binding.RunTranslationFlowPersona(taskID, request, prompt);
};

const getTranslationFlowPersonaBinding = async (taskID: string): Promise<unknown> => {
    const binding = getTaskControllerPersonaBinding();
    if (!binding?.GetTranslationFlowPersona) {
        throw new Error('GetTranslationFlowPersona binding is not available');
    }
    return binding.GetTranslationFlowPersona(taskID);
};

const isTerminologyCompleted = (status: string): boolean =>
    status === 'completed' || status === 'completed_partial';

const isPersonaAdvanceAllowed = (status: PersonaTargetViewState): boolean =>
    status === 'completed' || status === 'cachedOnly' || status === 'empty' || status === 'partialFailed';

const normalizePersonaRowKey = (row: PersonaTargetPreviewRow): string =>
    `${row.sourcePlugin}::${row.speakerId}`.trim();

const normalizeTerminologyProvider = (value: string | undefined): MasterPersonaLLMConfig['provider'] => {
    if (value === 'gemini' || value === 'xai' || value === 'lmstudio') {
        return value;
    }
    return DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider;
};

const hasStoredConfig = (loaded: Record<string, string>): boolean => Object.keys(loaded).length > 0;

const llmProviderNamespace = (
    namespace: string,
    provider: MasterPersonaLLMConfig['provider'],
): string => `${namespace}.${provider}`;

const terminologyProviderNamespace = (provider: MasterPersonaLLMConfig['provider']): string =>
    llmProviderNamespace(TERMINOLOGY_LLM_NAMESPACE, provider);

const personaProviderNamespace = (provider: MasterPersonaLLMConfig['provider']): string =>
    llmProviderNamespace(PERSONA_LLM_NAMESPACE, provider);

const normalizePromptConfig = (
    loaded: Record<string, string>,
    fallback: MasterPersonaPromptConfig,
): MasterPersonaPromptConfig => ({
    userPrompt: loaded[TERMINOLOGY_USER_PROMPT_KEY] ?? fallback.userPrompt,
    systemPrompt: loaded[TERMINOLOGY_SYSTEM_PROMPT_KEY] ?? fallback.systemPrompt,
});

const terminologyStatusLabel = (
    summary: TerminologyPhaseSummary,
    targetStatus: TerminologyTargetViewState,
): string => {
    const isNoTargetsState = targetStatus === 'empty'
        || (targetStatus === 'ready'
            && summary.status === 'pending'
            && summary.savedCount === 0
            && summary.failedCount === 0
            && summary.progressTotal === 0);
    if (summary.status === 'running' && summary.progressMessage !== '') {
        return summary.progressMessage;
    }
    if (targetStatus === 'loading') {
        return '読込中';
    }
    if (targetStatus === 'error') {
        return '対象単語リスト取得失敗';
    }
    if (isNoTargetsState) {
        return '用語翻訳対象なし';
    }
    switch (summary.status) {
        case 'completed_partial':
            return '単語翻訳完了（一部失敗あり）';
        case 'completed':
            return '単語翻訳完了';
        case 'running':
            return '単語翻訳を実行中';
        case 'run_error':
            return '単語翻訳の実行に失敗しました';
        default:
            return '未実行';
    }
};

const personaStatusLabel = (
    summary: PersonaPhaseSummary,
    targetStatus: PersonaTargetViewState,
): string => {
    if (summary.status === 'running' && summary.progressMessage !== '') {
        return summary.progressMessage;
    }
    switch (targetStatus) {
    case 'loadingTargets':
        return 'ペルソナ対象を読込中';
    case 'empty':
        return 'ペルソナ対象 NPC はありません';
    case 'cachedOnly':
        return '既存 Master Persona を再利用';
    case 'ready':
        return 'ペルソナ生成を開始できます';
    case 'running':
        return 'ペルソナ生成を実行中';
    case 'completed':
        return 'ペルソナ生成完了';
    case 'partialFailed':
        return 'ペルソナ生成完了（一部失敗あり）';
    case 'failed':
        return 'ペルソナ生成に失敗しました';
    default:
        return 'ペルソナ対象を読込中';
    }
};

/**
 * TranslationFlow 画面のロードフェーズ状態を headless に管理する。
 */
export function useTranslationFlow(): UseTranslationFlowWithPersonaResult {
    const location = useLocation();
    const navState = location.state as {taskId?: string} | null;

    const routeTaskID = useMemo(() => toTaskIDFromRoute(navState?.taskId), [navState?.taskId]);

    const [taskId, setTaskID] = useState(routeTaskID);
    const [isTaskIDResolved, setIsTaskIDResolved] = useState(false);
    const [activeTab, setActiveTab] = useState(0);
    const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
    const [loadedFiles, setLoadedFiles] = useState<UseTranslationFlowResult['state']['loadedFiles']>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const [terminologySummary, setTerminologySummary] = useState<TerminologyPhaseSummary>(EMPTY_TERMINOLOGY_SUMMARY);
    const [terminologyErrorMessage, setTerminologyErrorMessage] = useState('');
    const [terminologyTargetPage, setTerminologyTargetPage] = useState<TerminologyTargetPreviewPage>(
        EMPTY_TERMINOLOGY_TARGET_PAGE(routeTaskID),
    );
    const [terminologyTargetStatus, setTerminologyTargetStatus] = useState<TerminologyTargetViewState>('loading');
    const [terminologyTargetErrorMessage, setTerminologyTargetErrorMessage] = useState('');
    const [isTerminologyTargetLoading, setIsTerminologyTargetLoading] = useState(false);
    const [isTerminologyRunning, setIsTerminologyRunning] = useState(false);
    const isTerminologyRunInFlightRef = useRef(false);
    const [personaSummary, setPersonaSummary] = useState<PersonaPhaseSummary>(EMPTY_PERSONA_SUMMARY);
    const [personaErrorMessage, setPersonaErrorMessage] = useState('');
    const [personaTargetPage, setPersonaTargetPage] = useState<PersonaTargetPreviewPage>(
        EMPTY_PERSONA_TARGET_PAGE(routeTaskID),
    );
    const [personaTargetStatus, setPersonaTargetStatus] = useState<PersonaTargetViewState>('loadingTargets');
    const [personaTargetErrorMessage, setPersonaTargetErrorMessage] = useState('');
    const [isPersonaTargetLoading, setIsPersonaTargetLoading] = useState(false);
    const [isPersonaRunning, setIsPersonaRunning] = useState(false);
    const [selectedPersonaTargetKey, setSelectedPersonaTargetKey] = useState('');
    const [terminologyConfig, setTerminologyConfig] = useState<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const [terminologyPromptConfig, setTerminologyPromptConfig] = useState<MasterPersonaPromptConfig>(DEFAULT_TERMINOLOGY_PROMPT_CONFIG);
    const [personaConfig, setPersonaConfig] = useState<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const [personaPromptConfig, setPersonaPromptConfig] = useState<MasterPersonaPromptConfig>(DEFAULT_TERMINOLOGY_PROMPT_CONFIG);
    const [isTerminologyConfigHydrated, setIsTerminologyConfigHydrated] = useState(false);
    const [isTerminologyPromptHydrated, setIsTerminologyPromptHydrated] = useState(false);
    const [isPersonaConfigHydrated, setIsPersonaConfigHydrated] = useState(false);
    const [isPersonaPromptHydrated, setIsPersonaPromptHydrated] = useState(false);
    const terminologyLLMSaveTimerRef = useRef<number | null>(null);
    const terminologyPromptSaveTimerRef = useRef<number | null>(null);
    const personaLLMSaveTimerRef = useRef<number | null>(null);
    const personaPromptSaveTimerRef = useRef<number | null>(null);
    const latestTerminologyConfigRef = useRef<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const latestPersonaConfigRef = useRef<MasterPersonaLLMConfig>(DEFAULT_MASTER_PERSONA_LLM_CONFIG);
    const previousTerminologyProviderRef = useRef<MasterPersonaLLMConfig['provider']>(DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider);
    const previousPersonaProviderRef = useRef<MasterPersonaLLMConfig['provider']>(DEFAULT_MASTER_PERSONA_LLM_CONFIG.provider);
    const isSwitchingTerminologyProviderRef = useRef(false);
    const isSwitchingPersonaProviderRef = useRef(false);

    const persistLLMConfig = useCallback((namespace: string, current: MasterPersonaLLMConfig): Promise<void> =>
        Promise.all([
            ConfigSet(namespace, TERMINOLOGY_SELECTED_PROVIDER_KEY, current.provider),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'model', current.model),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'endpoint', current.endpoint),
            ConfigSet(
                llmProviderNamespace(namespace, current.provider),
                'api_key',
                current.provider === 'lmstudio' ? '' : current.apiKey,
            ),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'temperature', String(current.temperature)),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'context_length', String(current.contextLength)),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'sync_concurrency', String(current.syncConcurrency)),
            ConfigSet(llmProviderNamespace(namespace, current.provider), 'bulk_strategy', current.bulkStrategy),
        ]).then(() => undefined), []);

    const persistTerminologyLLMConfig = useCallback((current: MasterPersonaLLMConfig): Promise<void> =>
        persistLLMConfig(TERMINOLOGY_LLM_NAMESPACE, current), [persistLLMConfig]);

    const persistPersonaLLMConfig = useCallback((current: MasterPersonaLLMConfig): Promise<void> =>
        persistLLMConfig(PERSONA_LLM_NAMESPACE, current), [persistLLMConfig]);

    const persistPromptConfig = useCallback((namespace: string, current: MasterPersonaPromptConfig): Promise<void> =>
        Promise.all([
            ConfigSet(namespace, TERMINOLOGY_USER_PROMPT_KEY, current.userPrompt),
            ConfigSet(namespace, TERMINOLOGY_SYSTEM_PROMPT_KEY, current.systemPrompt),
        ]).then(() => undefined), []);

    useWailsEvent<WailsTerminologyProgressEvent>(TERMINOLOGY_PROGRESS_EVENT, (payload) => {
        const eventTaskId = pickProgressEventString(payload.task_id ?? payload.taskId ?? payload.TaskID);
        const eventStatus = pickProgressEventString(payload.status ?? payload.Status);
        const progressTotal = Math.max(0, pickProgressEventNumber(payload.total ?? payload.Total));
        const progressCurrent = Math.max(
            0,
            pickProgressEventNumber(payload.current ?? payload.Current ?? payload.completed ?? payload.Completed),
        );
        const incomingMessage = pickProgressEventString(payload.message ?? payload.Message);
        if (eventTaskId === '' || (taskId !== '' && eventTaskId !== taskId)) {
            return;
        }

        if (eventStatus !== 'IN_PROGRESS') {
            return;
        }

        const shouldApplyProgress = isTerminologyRunInFlightRef.current
            || isTerminologyRunning
            || terminologySummary.status === 'running';
        if (!shouldApplyProgress) {
            return;
        }

        setTerminologySummary((prev) => {
            const monotonicCurrent = Math.max(prev.progressCurrent, progressCurrent);
            const normalizedTotal = Math.max(progressTotal, monotonicCurrent);
            const remaining = normalizedTotal > 0 ? Math.max(0, normalizedTotal - monotonicCurrent) : 0;
            const fallbackMessage = normalizedTotal > 0
                ? `${monotonicCurrent} / ${normalizedTotal} 件（残り ${remaining} 件）`
                : '単語翻訳を実行中';
            return {
                ...prev,
                taskId: eventTaskId,
                status: 'running',
                progressMode: normalizedTotal > 0 ? 'determinate' : 'indeterminate',
                progressCurrent: monotonicCurrent,
                progressTotal: normalizedTotal,
                progressMessage: incomingMessage || fallbackMessage,
            };
        });
        setTerminologyErrorMessage('');
        setTerminologyTargetStatus('loading');
        setIsTerminologyRunning(true);
    });

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

    const handleRefreshTerminologySummary = useCallback(async (nextTaskId: string): Promise<string> => {
        if (nextTaskId === '') {
            setTerminologySummary(EMPTY_TERMINOLOGY_SUMMARY);
            setTerminologyErrorMessage('');
            return '';
        }

        try {
            const payload = await GetTranslationFlowTerminology(nextTaskId);
            const summary = mapTerminologyPhaseResult(payload);
            const resolvedTaskId = summary.taskId || nextTaskId;
            const nextIsTerminologyRunning = summary.status === 'running';

            if (resolvedTaskId !== '' && resolvedTaskId !== taskId) {
                setTaskID(resolvedTaskId);
            }

            setTerminologySummary({
                ...summary,
                taskId: resolvedTaskId,
            });
            setIsTerminologyRunning(nextIsTerminologyRunning);
            setTerminologyErrorMessage('');
            return resolvedTaskId;
        } catch (error) {
            setIsTerminologyRunning(false);
            const message = toErrorMessage(error, '単語翻訳の状態取得に失敗しました');
            setTerminologyErrorMessage(message);
            return nextTaskId;
        }
    }, [taskId]);

    const handleRefreshTerminologyTargets = useCallback(async (
        nextTaskId: string,
        page: number,
        pageSize: number,
    ): Promise<TerminologyTargetPreviewPage> => {
        if (nextTaskId === '') {
            const emptyPage = EMPTY_TERMINOLOGY_TARGET_PAGE('', page, pageSize);
            setTerminologyTargetPage(emptyPage);
            setTerminologyTargetStatus('empty');
            setTerminologyTargetErrorMessage('');
            setIsTerminologyTargetLoading(false);
            return emptyPage;
        }

        setIsTerminologyTargetLoading(true);
        setTerminologyTargetStatus('loading');
        setTerminologyTargetErrorMessage('');
        setTerminologyTargetPage(EMPTY_TERMINOLOGY_TARGET_PAGE(nextTaskId, page, pageSize));

        try {
            const payload = await ListTranslationFlowTerminologyTargets(nextTaskId, page, pageSize);
            const mapped = mapTerminologyTargetPreviewPage(payload);
            const resolvedTaskId = mapped.taskId || nextTaskId;
            const nextPage = {
                ...mapped,
                taskId: resolvedTaskId,
            };

            if (resolvedTaskId !== '' && resolvedTaskId !== taskId) {
                setTaskID(resolvedTaskId);
            }

            setTerminologyTargetPage(nextPage);
            const hasTargets = nextPage.totalRows > 0;
            setTerminologyTargetStatus('ready');
            setTerminologyErrorMessage(hasTargets ? '' : NO_TERMINOLOGY_TARGETS_MESSAGE);
            return nextPage;
        } catch (error) {
            const message = toErrorMessage(error, '対象単語リストの取得に失敗しました');
            setTerminologyTargetErrorMessage(message);
            setTerminologyTargetStatus('error');
            const emptyPage = EMPTY_TERMINOLOGY_TARGET_PAGE(nextTaskId, page, pageSize);
            setTerminologyTargetPage(emptyPage);
            return emptyPage;
        } finally {
            setIsTerminologyTargetLoading(false);
        }
    }, [taskId]);

    const handleRefreshPersonaSummary = useCallback(async (nextTaskId: string): Promise<PersonaPhaseSummary> => {
        if (nextTaskId === '') {
            setPersonaSummary(EMPTY_PERSONA_SUMMARY);
            setPersonaErrorMessage('');
            return EMPTY_PERSONA_SUMMARY;
        }

        try {
            const payload = await getTranslationFlowPersonaBinding(nextTaskId);
            const summary = mapPersonaPhaseResult(payload);
            const resolvedTaskID = summary.taskId || nextTaskId;
            const nextSummary = {
                ...summary,
                taskId: resolvedTaskID,
            };
            if (resolvedTaskID !== '' && resolvedTaskID !== taskId) {
                setTaskID(resolvedTaskID);
            }
            setPersonaSummary(nextSummary);
            setPersonaTargetStatus(nextSummary.status);
            setPersonaErrorMessage('');
            setIsPersonaRunning(nextSummary.status === 'running');
            return nextSummary;
        } catch (error) {
            const message = toErrorMessage(error, 'ペルソナ生成の状態取得に失敗しました');
            setPersonaErrorMessage(message);
            setIsPersonaRunning(false);
            return {
                ...EMPTY_PERSONA_SUMMARY,
                taskId: nextTaskId,
                status: 'failed',
            };
        }
    }, [taskId]);

    const handleRefreshPersonaTargets = useCallback(async (
        nextTaskId: string,
        page: number,
        pageSize: number,
        statusFallback: PersonaTargetViewState,
    ): Promise<PersonaTargetPreviewPage> => {
        if (nextTaskId === '') {
            const emptyPage = EMPTY_PERSONA_TARGET_PAGE('', page, pageSize);
            setPersonaTargetPage(emptyPage);
            setPersonaTargetStatus('empty');
            setPersonaTargetErrorMessage('');
            setSelectedPersonaTargetKey('');
            setIsPersonaTargetLoading(false);
            return emptyPage;
        }

        setIsPersonaTargetLoading(true);
        setPersonaTargetStatus('loadingTargets');
        setPersonaTargetErrorMessage('');
        setPersonaTargetPage(EMPTY_PERSONA_TARGET_PAGE(nextTaskId, page, pageSize));

        try {
            const payload = await listTranslationFlowPersonaTargetsBinding(nextTaskId, page, pageSize);
            const mapped = mapPersonaTargetPreviewPage(payload);
            const resolvedTaskID = mapped.taskId || nextTaskId;
            const nextPage = {
                ...mapped,
                taskId: resolvedTaskID,
            };
            if (resolvedTaskID !== '' && resolvedTaskID !== taskId) {
                setTaskID(resolvedTaskID);
            }
            setPersonaTargetPage(nextPage);
            setSelectedPersonaTargetKey((previous) => {
                if (nextPage.rows.length === 0) {
                    return '';
                }
                const hasPrevious = nextPage.rows.some((row) => normalizePersonaRowKey(row) === previous);
                if (hasPrevious) {
                    return previous;
                }
                return normalizePersonaRowKey(nextPage.rows[0]);
            });
            const nextStatus = nextPage.totalRows > 0 ? statusFallback : 'empty';
            setPersonaTargetStatus(nextStatus);
            return nextPage;
        } catch (error) {
            const message = toErrorMessage(error, 'ペルソナ対象一覧の取得に失敗しました');
            setPersonaTargetErrorMessage(message);
            setPersonaTargetStatus('failed');
            const emptyPage = EMPTY_PERSONA_TARGET_PAGE(nextTaskId, page, pageSize);
            setPersonaTargetPage(emptyPage);
            setSelectedPersonaTargetKey('');
            return emptyPage;
        } finally {
            setIsPersonaTargetLoading(false);
        }
    }, [taskId]);

    const handleRefreshTerminologyPhase = useCallback(async (nextTaskId = taskId): Promise<void> => {
        if (nextTaskId === '') {
            setTerminologySummary(EMPTY_TERMINOLOGY_SUMMARY);
            setTerminologyErrorMessage('');
            setTerminologyTargetPage(EMPTY_TERMINOLOGY_TARGET_PAGE('', 1, PREVIEW_PAGE_SIZE));
            setTerminologyTargetStatus('empty');
            setTerminologyTargetErrorMessage('');
            return;
        }

        const resolvedTaskId = await handleRefreshTerminologySummary(nextTaskId);
        await handleRefreshTerminologyTargets(resolvedTaskId || nextTaskId, 1, PREVIEW_PAGE_SIZE);
    }, [handleRefreshTerminologySummary, handleRefreshTerminologyTargets, taskId]);

    const handleRefreshPersonaPhase = useCallback(async (nextTaskId = taskId): Promise<void> => {
        if (nextTaskId === '') {
            setPersonaSummary(EMPTY_PERSONA_SUMMARY);
            setPersonaErrorMessage('');
            setPersonaTargetPage(EMPTY_PERSONA_TARGET_PAGE('', 1, PREVIEW_PAGE_SIZE));
            setPersonaTargetStatus('empty');
            setPersonaTargetErrorMessage('');
            setSelectedPersonaTargetKey('');
            setIsPersonaRunning(false);
            return;
        }

        const summary = await handleRefreshPersonaSummary(nextTaskId);
        const resolvedTaskID = summary.taskId || nextTaskId;
        await handleRefreshPersonaTargets(resolvedTaskID, 1, PREVIEW_PAGE_SIZE, summary.status);
    }, [handleRefreshPersonaSummary, handleRefreshPersonaTargets, taskId]);

    const handlePersonaTargetPageChange = useCallback(async (page: number): Promise<void> => {
        if (taskId === '') {
            return;
        }
        const safePage = Math.max(1, page);
        await handleRefreshPersonaTargets(taskId, safePage, personaTargetPage.pageSize, personaSummary.status);
    }, [handleRefreshPersonaTargets, personaSummary.status, personaTargetPage.pageSize, taskId]);

    const handleSelectPersonaTarget = useCallback((sourcePlugin: string, speakerId: string) => {
        const normalizedSourcePlugin = sourcePlugin.trim();
        const normalizedSpeakerID = speakerId.trim();
        if (normalizedSpeakerID === '') {
            return;
        }
        setSelectedPersonaTargetKey(`${normalizedSourcePlugin}::${normalizedSpeakerID}`);
    }, []);

    const handleRunPersonaPhase = useCallback(async (): Promise<void> => {
        if (taskId === '' || isPersonaRunning) {
            return;
        }
        if (personaTargetStatus === 'empty') {
            setPersonaErrorMessage(NO_PERSONA_TARGETS_MESSAGE);
            return;
        }
        if (personaTargetStatus === 'cachedOnly' || personaTargetStatus === 'completed') {
            return;
        }

        setIsPersonaRunning(true);
        setPersonaErrorMessage('');
        setPersonaTargetErrorMessage('');
        const initialTotal = Math.max(0, personaTargetPage.totalRows);
        setPersonaSummary((previous) => ({
            ...previous,
            taskId,
            status: 'running',
            progressMode: initialTotal > 0 ? 'determinate' : 'indeterminate',
            progressCurrent: 0,
            progressTotal: initialTotal > 0 ? initialTotal : Math.max(previous.progressTotal, 0),
            progressMessage: initialTotal > 0
                ? `0 / ${initialTotal} 件（残り ${initialTotal} 件）`
                : 'ペルソナ生成を実行中',
        }));
        setPersonaTargetStatus('running');

        try {
            const payload = await runTranslationFlowPersonaBinding(
                taskId,
                {
                    provider: personaConfig.provider,
                    model: personaConfig.model,
                    endpoint: personaConfig.endpoint,
                    api_key: personaConfig.provider === 'lmstudio' ? '' : personaConfig.apiKey,
                    temperature: personaConfig.temperature,
                    context_length: personaConfig.contextLength,
                    sync_concurrency: personaConfig.syncConcurrency,
                    bulk_strategy: personaConfig.bulkStrategy,
                },
                {
                    user_prompt: personaPromptConfig.userPrompt,
                    system_prompt: personaPromptConfig.systemPrompt,
                },
            );
            const summary = mapPersonaPhaseResult(payload);
            const resolvedTaskID = summary.taskId || taskId;
            const nextSummary = {
                ...summary,
                taskId: resolvedTaskID,
            };
            setPersonaSummary(nextSummary);
            await handleRefreshPersonaTargets(
                resolvedTaskID,
                personaTargetPage.page,
                personaTargetPage.pageSize,
                nextSummary.status,
            );
        } catch (error) {
            setPersonaErrorMessage(toErrorMessage(error, 'ペルソナ生成の実行に失敗しました'));
            await handleRefreshPersonaPhase(taskId);
        } finally {
            setIsPersonaRunning(false);
        }
    }, [
        handleRefreshPersonaPhase,
        handleRefreshPersonaTargets,
        isPersonaRunning,
        personaTargetPage.page,
        personaTargetPage.pageSize,
        personaTargetPage.totalRows,
        personaTargetStatus,
        taskId,
        personaConfig.apiKey,
        personaConfig.bulkStrategy,
        personaConfig.contextLength,
        personaConfig.endpoint,
        personaConfig.model,
        personaConfig.provider,
        personaConfig.syncConcurrency,
        personaConfig.temperature,
        personaPromptConfig.systemPrompt,
        personaPromptConfig.userPrompt,
    ]);

    const handleRetryPersonaPhase = useCallback(async (): Promise<void> => {
        if (personaSummary.status !== 'partialFailed' && personaSummary.status !== 'failed') {
            return;
        }
        await handleRunPersonaPhase();
    }, [handleRunPersonaPhase, personaSummary.status]);

    useEffect(() => {
        let active = true;
        setIsTaskIDResolved(false);

        const run = async () => {
            try {
                const resolved = await resolveTaskID(routeTaskID);
                if (!active) {
                    return;
                }
                setTaskID(resolved);
                setIsTaskIDResolved(true);
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
                setIsTaskIDResolved(true);
            }
        };

        void run();

        return () => {
            active = false;
        };
    }, [routeTaskID]);

    useEffect(() => {
        latestTerminologyConfigRef.current = terminologyConfig;
    }, [terminologyConfig]);

    useEffect(() => {
        latestPersonaConfigRef.current = personaConfig;
    }, [personaConfig]);

    useEffect(() => {
        let alive = true;

        const loadProviderConfig = async (
            namespace: string,
            provider: MasterPersonaLLMConfig['provider'],
            fallback?: Record<string, string>,
        ): Promise<MasterPersonaLLMConfig> => {
            const loaded = await ConfigGetAll(llmProviderNamespace(namespace, provider));
            const source = hasStoredConfig(loaded) ? loaded : (fallback ?? {});
            return {
                ...normalizeTerminologyLLMConfig(source),
                provider,
            };
        };

        void (async () => {
            try {
                const root = await ConfigGetAll(TERMINOLOGY_LLM_NAMESPACE);
                const selectedProvider = normalizeTerminologyProvider(
                    root[TERMINOLOGY_SELECTED_PROVIDER_KEY] || root.provider,
                );
                const loaded = await loadProviderConfig(TERMINOLOGY_LLM_NAMESPACE, selectedProvider, root);
                if (!alive) {
                    return;
                }
                setTerminologyConfig(loaded);
                latestTerminologyConfigRef.current = loaded;
                previousTerminologyProviderRef.current = selectedProvider;

                const personaRoot = await ConfigGetAll(PERSONA_LLM_NAMESPACE);
                const personaSelectedProvider = normalizeTerminologyProvider(
                    personaRoot[TERMINOLOGY_SELECTED_PROVIDER_KEY] || personaRoot.provider || loaded.provider,
                );
                const personaProviderLoaded = await ConfigGetAll(personaProviderNamespace(personaSelectedProvider));
                const personaFallback = hasStoredConfig(personaRoot)
                    ? personaRoot
                    : {
                        provider: loaded.provider,
                        model: loaded.model,
                        endpoint: loaded.endpoint,
                        api_key: loaded.provider === 'lmstudio' ? '' : loaded.apiKey,
                        temperature: String(loaded.temperature),
                        context_length: String(loaded.contextLength),
                        sync_concurrency: String(loaded.syncConcurrency),
                        bulk_strategy: loaded.bulkStrategy,
                    };
                const loadedPersonaConfig = {
                    ...normalizeTerminologyLLMConfig(
                        hasStoredConfig(personaProviderLoaded) ? personaProviderLoaded : personaFallback,
                    ),
                    provider: personaSelectedProvider,
                };
                if (!alive) {
                    return;
                }
                setPersonaConfig(loadedPersonaConfig);
                latestPersonaConfigRef.current = loadedPersonaConfig;
                previousPersonaProviderRef.current = personaSelectedProvider;
                if (!hasStoredConfig(personaRoot) && !hasStoredConfig(personaProviderLoaded)) {
                    void persistPersonaLLMConfig(loadedPersonaConfig);
                }
            } finally {
                if (alive) {
                    setIsTerminologyConfigHydrated(true);
                    setIsPersonaConfigHydrated(true);
                }
            }
        })();

        void (async () => {
            try {
                const loaded = await ConfigGetAll(TERMINOLOGY_PROMPT_NAMESPACE);
                const normalizedTerminologyPrompt = normalizePromptConfig(loaded, DEFAULT_TERMINOLOGY_PROMPT_CONFIG);
                if (!alive) {
                    return;
                }
                setTerminologyPromptConfig(normalizedTerminologyPrompt);

                const personaLoaded = await ConfigGetAll(PERSONA_PROMPT_NAMESPACE);
                const normalizedPersonaPrompt = normalizePromptConfig(
                    personaLoaded,
                    normalizedTerminologyPrompt,
                );
                if (!alive) {
                    return;
                }
                setPersonaPromptConfig(normalizedPersonaPrompt);
                if (!hasStoredConfig(personaLoaded)) {
                    void persistPromptConfig(PERSONA_PROMPT_NAMESPACE, normalizedPersonaPrompt);
                }
            } finally {
                if (alive) {
                    setIsTerminologyPromptHydrated(true);
                    setIsPersonaPromptHydrated(true);
                }
            }
        })();

        return () => {
            alive = false;
        };
    }, [persistPersonaLLMConfig, persistPromptConfig]);

    useEffect(() => {
        if (!isTerminologyConfigHydrated) {
            return;
        }
        const currentProvider = terminologyConfig.provider;
        const previousProvider = previousTerminologyProviderRef.current;
        if (currentProvider === previousProvider) {
            return;
        }
        previousTerminologyProviderRef.current = currentProvider;
        isSwitchingTerminologyProviderRef.current = true;
        let alive = true;

        void ConfigGetAll(terminologyProviderNamespace(currentProvider))
            .then((loaded) => {
                if (!alive) {
                    return;
                }
                const nextConfig = {
                    ...normalizeTerminologyLLMConfig(loaded),
                    provider: currentProvider,
                };
                setTerminologyConfig(nextConfig);
                latestTerminologyConfigRef.current = nextConfig;
            })
            .finally(() => {
                if (alive) {
                    isSwitchingTerminologyProviderRef.current = false;
                    void persistTerminologyLLMConfig(latestTerminologyConfigRef.current);
                }
            });

        return () => {
            alive = false;
        };
    }, [isTerminologyConfigHydrated, persistTerminologyLLMConfig, terminologyConfig.provider]);

    useEffect(() => {
        if (!isPersonaConfigHydrated) {
            return;
        }
        const currentProvider = personaConfig.provider;
        const previousProvider = previousPersonaProviderRef.current;
        if (currentProvider === previousProvider) {
            return;
        }
        previousPersonaProviderRef.current = currentProvider;
        isSwitchingPersonaProviderRef.current = true;
        let alive = true;

        void ConfigGetAll(personaProviderNamespace(currentProvider))
            .then((loaded) => {
                if (!alive) {
                    return;
                }
                const nextConfig = {
                    ...normalizeTerminologyLLMConfig(loaded),
                    provider: currentProvider,
                };
                setPersonaConfig(nextConfig);
                latestPersonaConfigRef.current = nextConfig;
            })
            .finally(() => {
                if (alive) {
                    isSwitchingPersonaProviderRef.current = false;
                    void persistPersonaLLMConfig(latestPersonaConfigRef.current);
                }
            });

        return () => {
            alive = false;
        };
    }, [isPersonaConfigHydrated, persistPersonaLLMConfig, personaConfig.provider]);

    useEffect(() => {
        if (!isTerminologyConfigHydrated) {
            return;
        }
        if (isSwitchingTerminologyProviderRef.current) {
            return;
        }
        if (terminologyLLMSaveTimerRef.current) {
            window.clearTimeout(terminologyLLMSaveTimerRef.current);
        }
        terminologyLLMSaveTimerRef.current = window.setTimeout(() => {
            void persistTerminologyLLMConfig(latestTerminologyConfigRef.current);
        }, 250);

        return () => {
            if (terminologyLLMSaveTimerRef.current) {
                window.clearTimeout(terminologyLLMSaveTimerRef.current);
            }
        };
    }, [isTerminologyConfigHydrated, persistTerminologyLLMConfig, terminologyConfig]);

    useEffect(() => {
        if (!isPersonaConfigHydrated) {
            return;
        }
        if (isSwitchingPersonaProviderRef.current) {
            return;
        }
        if (personaLLMSaveTimerRef.current) {
            window.clearTimeout(personaLLMSaveTimerRef.current);
        }
        personaLLMSaveTimerRef.current = window.setTimeout(() => {
            void persistPersonaLLMConfig(latestPersonaConfigRef.current);
        }, 250);

        return () => {
            if (personaLLMSaveTimerRef.current) {
                window.clearTimeout(personaLLMSaveTimerRef.current);
            }
        };
    }, [isPersonaConfigHydrated, persistPersonaLLMConfig, personaConfig]);

    useEffect(() => {
        if (!isTerminologyPromptHydrated) {
            return;
        }
        if (terminologyPromptSaveTimerRef.current) {
            window.clearTimeout(terminologyPromptSaveTimerRef.current);
        }
        terminologyPromptSaveTimerRef.current = window.setTimeout(() => {
            void persistPromptConfig(TERMINOLOGY_PROMPT_NAMESPACE, terminologyPromptConfig);
        }, 250);

        return () => {
            if (terminologyPromptSaveTimerRef.current) {
                window.clearTimeout(terminologyPromptSaveTimerRef.current);
            }
        };
    }, [isTerminologyPromptHydrated, persistPromptConfig, terminologyPromptConfig]);

    useEffect(() => {
        if (!isPersonaPromptHydrated) {
            return;
        }
        if (personaPromptSaveTimerRef.current) {
            window.clearTimeout(personaPromptSaveTimerRef.current);
        }
        personaPromptSaveTimerRef.current = window.setTimeout(() => {
            void persistPromptConfig(PERSONA_PROMPT_NAMESPACE, personaPromptConfig);
        }, 250);

        return () => {
            if (personaPromptSaveTimerRef.current) {
                window.clearTimeout(personaPromptSaveTimerRef.current);
            }
        };
    }, [isPersonaPromptHydrated, persistPromptConfig, personaPromptConfig]);

    useEffect(() => {
        if (!isTaskIDResolved) {
            return;
        }
        if (taskId === '') {
            setLoadedFiles([]);
            setTerminologySummary(EMPTY_TERMINOLOGY_SUMMARY);
            setTerminologyErrorMessage('');
            setTerminologyTargetPage(EMPTY_TERMINOLOGY_TARGET_PAGE('', 1, PREVIEW_PAGE_SIZE));
            setTerminologyTargetStatus('empty');
            setTerminologyTargetErrorMessage('');
            setPersonaSummary(EMPTY_PERSONA_SUMMARY);
            setPersonaErrorMessage('');
            setPersonaTargetPage(EMPTY_PERSONA_TARGET_PAGE('', 1, PREVIEW_PAGE_SIZE));
            setPersonaTargetStatus('empty');
            setPersonaTargetErrorMessage('');
            setSelectedPersonaTargetKey('');
            setErrorMessage('');
            setActiveTab(0);
            setIsTerminologyRunning(false);
            setIsPersonaRunning(false);
            return;
        }
        void handleReloadFiles();
        void handleRefreshTerminologyPhase(taskId);
        void handleRefreshPersonaPhase(taskId);
    }, [handleRefreshPersonaPhase, handleRefreshTerminologyPhase, handleReloadFiles, isTaskIDResolved, taskId]);

    const handleSelectFiles = useCallback(async () => {
        setErrorMessage('');
        try {
            const files = await SelectTranslationInputFiles();
            if (!Array.isArray(files) || files.length === 0) {
                return;
            }

            const loadedFileNames = new Set(loadedFiles.map((file) => normalizeFileName(file.fileName || file.filePath)));
            const {files: merged, duplicateBlocked} = mergeUniqueFilesByName(selectedFiles, files, loadedFileNames);

            setSelectedFiles(merged);
            if (duplicateBlocked) {
                setErrorMessage(DUPLICATE_FILE_MESSAGE);
            }
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイル選択に失敗しました'));
        }
    }, [loadedFiles, selectedFiles]);

    const handleRemoveFile = useCallback((pathToRemove: string) => {
        setSelectedFiles((prev) => prev.filter((path) => path !== pathToRemove));
    }, []);

    const handleLoadSelectedFiles = useCallback(async () => {
        if (selectedFiles.length === 0) {
            return;
        }

        setIsLoading(true);
        setErrorMessage('');
        try {
            const payload = await LoadTranslationFlowFiles(taskId, selectedFiles);
            const mapped = mapLoadResult(payload);
            const resolvedTaskId = mapped.taskId !== '' ? mapped.taskId : taskId;
            if (resolvedTaskId !== '' && resolvedTaskId !== taskId) {
                setTaskID(resolvedTaskId);
            }
            setLoadedFiles(mapped.files);
            setSelectedFiles([]);
            setTerminologySummary({...EMPTY_TERMINOLOGY_SUMMARY, taskId: resolvedTaskId});
            await handleRefreshTerminologyPhase(resolvedTaskId);
        } catch (error) {
            setErrorMessage(toErrorMessage(error, 'ファイルロードに失敗しました'));
        } finally {
            setIsLoading(false);
        }
    }, [handleRefreshTerminologyPhase, selectedFiles, taskId]);

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

    const handleTerminologyTargetPageChange = useCallback(async (page: number) => {
        if (taskId === '') {
            return;
        }
        await handleRefreshTerminologyTargets(taskId, Math.max(1, page), terminologyTargetPage.pageSize);
    }, [handleRefreshTerminologyTargets, taskId, terminologyTargetPage.pageSize]);

    const handleRunTerminologyPhase = useCallback(async () => {
        if (taskId === '' || isTerminologyRunning) {
            return;
        }
        if (terminologyTargetPage.totalRows === 0) {
            setTerminologyErrorMessage(NO_TERMINOLOGY_TARGETS_MESSAGE);
            return;
        }
        const runStartedAt = Date.now();
        isTerminologyRunInFlightRef.current = true;
        setIsTerminologyRunning(true);
        setTerminologyErrorMessage('');
        const initialTotal = Math.max(0, terminologyTargetPage.totalRows);
        setTerminologySummary((prev) => ({
            ...prev,
            taskId,
            status: 'running',
            progressMode: initialTotal > 0 ? 'determinate' : 'indeterminate',
            progressCurrent: 0,
            progressTotal: initialTotal > 0 ? initialTotal : Math.max(prev.progressTotal, 0),
            progressMessage: initialTotal > 0
                ? `0 / ${initialTotal} 件（残り ${initialTotal} 件）`
                : '単語翻訳を実行中',
        }));
        setTerminologyTargetErrorMessage('');

        try {
            setTerminologyTargetStatus('loading');
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
            const summary = mapTerminologyPhaseResult(payload);
            const resolvedTaskId = summary.taskId || taskId;
            setTerminologySummary({
                ...summary,
                taskId: resolvedTaskId,
            });
            if (summary.status === 'pending' && summary.savedCount === 0 && summary.failedCount === 0) {
                setTerminologyErrorMessage(NO_TERMINOLOGY_TARGETS_MESSAGE);
                setTerminologyTargetStatus('ready');
                setTerminologyTargetPage(EMPTY_TERMINOLOGY_TARGET_PAGE(resolvedTaskId, terminologyTargetPage.page, terminologyTargetPage.pageSize));
                return;
            }
            await handleRefreshTerminologyTargets(resolvedTaskId, terminologyTargetPage.page, terminologyTargetPage.pageSize);
        } catch (error) {
            setTerminologyErrorMessage(toErrorMessage(error, '単語翻訳の実行に失敗しました'));
            await handleRefreshTerminologyPhase(taskId);
        } finally {
            const minimumRunningMs = 300;
            const elapsedMs = Date.now() - runStartedAt;
            if (elapsedMs < minimumRunningMs) {
                await new Promise<void>((resolve) => {
                    window.setTimeout(resolve, minimumRunningMs - elapsedMs);
                });
            }
            isTerminologyRunInFlightRef.current = false;
            setIsTerminologyRunning(false);
        }
    }, [
        handleRefreshTerminologyPhase,
        handleRefreshTerminologyTargets,
        isTerminologyRunning,
        taskId,
        terminologyConfig,
        terminologyPromptConfig,
        terminologyTargetPage.page,
        terminologyTargetPage.pageSize,
        terminologyTargetPage.totalRows,
    ]);

    const handleAdvanceFromTerminology = useCallback(() => {
        if (!isTerminologyCompleted(terminologySummary.status)) {
            return;
        }
        setActiveTab(2);
        void handleRefreshPersonaPhase(taskId);
    }, [handleRefreshPersonaPhase, taskId, terminologySummary.status]);

    const handleAdvanceFromPersona = useCallback(() => {
        if (!isPersonaAdvanceAllowed(personaSummary.status)) {
            return;
        }
        setActiveTab(3);
    }, [personaSummary.status]);

    const handleTerminologyConfigChange = useCallback((next: MasterPersonaLLMConfig) => {
        setTerminologyConfig(next);
    }, []);

    const handleTerminologyPromptChange = useCallback((next: MasterPersonaPromptConfig) => {
        setTerminologyPromptConfig(next);
    }, []);

    const handlePersonaConfigChange = useCallback((next: MasterPersonaLLMConfig) => {
        setPersonaConfig(next);
    }, []);

    const handlePersonaPromptChange = useCallback((next: MasterPersonaPromptConfig) => {
        setPersonaPromptConfig(next);
    }, []);

    const handleTabChange = useCallback((index: number) => {
        if (index < 0 || index >= TABS.length) {
            return;
        }
        if (index > 0 && loadedFiles.length === 0) {
            return;
        }
        if (index > 1 && !isTerminologyCompleted(terminologySummary.status)) {
            return;
        }
        if (index > 2 && !isPersonaAdvanceAllowed(personaSummary.status)) {
            return;
        }
        if (index === 2 && taskId !== '') {
            void handleRefreshPersonaPhase(taskId);
        }
        setActiveTab(index);
    }, [
        handleRefreshPersonaPhase,
        loadedFiles.length,
        personaSummary.status,
        taskId,
        terminologySummary.status,
    ]);

    const selectedPersonaTarget = useMemo<PersonaTargetPreviewRow | null>(() => {
        if (selectedPersonaTargetKey === '') {
            return personaTargetPage.rows[0] ?? null;
        }
        const matched = personaTargetPage.rows.find((row) => normalizePersonaRowKey(row) === selectedPersonaTargetKey);
        return matched ?? personaTargetPage.rows[0] ?? null;
    }, [personaTargetPage.rows, selectedPersonaTargetKey]);

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
            terminologyStatusLabel: terminologyStatusLabel(
                terminologySummary,
                terminologyTargetStatus,
            ),
            terminologyErrorMessage,
            terminologyTargetPage,
            terminologyTargetStatus,
            terminologyTargetErrorMessage,
            isTerminologyTargetLoading,
            isTerminologyRunning,
            terminologyConfig,
            terminologyPromptConfig,
            isTerminologyConfigHydrated,
            isTerminologyPromptHydrated,
            personaConfig,
            personaPromptConfig,
            isPersonaConfigHydrated,
            isPersonaPromptHydrated,
            personaSummary,
            personaStatusLabel: personaStatusLabel(personaSummary, personaTargetStatus),
            personaErrorMessage,
            personaTargetPage,
            personaTargetStatus,
            personaTargetErrorMessage,
            isPersonaTargetLoading,
            isPersonaRunning,
            selectedPersonaTargetKey,
            selectedPersonaTarget,
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
            handleTerminologyTargetPageChange,
            handleTerminologyConfigChange,
            handleTerminologyPromptChange,
            handlePersonaConfigChange,
            handlePersonaPromptChange,
            handleAdvanceFromTerminology,
            handleRunPersonaPhase,
            handleRetryPersonaPhase,
            handleRefreshPersonaPhase,
            handlePersonaTargetPageChange,
            handleSelectPersonaTarget,
            handleAdvanceFromPersona,
        },
        ui: {
            previewPageSize: PREVIEW_PAGE_SIZE,
        },
    };
}
