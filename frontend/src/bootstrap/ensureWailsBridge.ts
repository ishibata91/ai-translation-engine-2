type RuntimeMethod = (...args: unknown[]) => unknown;
type ControllerMethod = (...args: unknown[]) => Promise<unknown>;
type RuntimeMap = Record<string, RuntimeMethod>;
type ControllerMap = Record<string, ControllerMethod>;
type ControllerRegistry = Record<string, ControllerMap>;

type WailsWindow = Window & {
    runtime?: RuntimeMap;
    go?: {
        controller?: ControllerRegistry;
    };
};

const BROWSER_MOCK_TASK_ID = 'browser-mock-translation-project';
const BROWSER_MOCK_FILE_ID = 9001;
const BROWSER_MOCK_SOURCE_FILE = 'F:/mock/translation-flow/dialogue_sample_01.json';

const EMPTY_TERMINOLOGY_RESULT = () => ({task_id: '', status: 'pending', saved_count: 0, failed_count: 0});
const EMPTY_TERMINOLOGY_TARGET_PAGE = () => ({task_id: '', page: 1, page_size: 50, total_rows: 0, rows: []});

const resolveTaskIDFromArgs = (args: unknown[]): string => {
    const candidate = args[0];
    if (typeof candidate === 'string' && candidate.trim() !== '') {
        return candidate.trim();
    }
    return BROWSER_MOCK_TASK_ID;
};

const toInt = (value: unknown, fallback: number): number => {
    if (typeof value === 'number' && Number.isFinite(value)) {
        return Math.floor(value);
    }
    return fallback;
};

const createBrowserMockRows = () => ([
    {
        id: 'mock-row-1',
        section: 'dialogue',
        record_type: 'NPC_',
        editor_id: 'MQ101Farengar',
        source_text: 'Dragonstone? Bring it to me at once.',
    },
    {
        id: 'mock-row-2',
        section: 'dialogue',
        record_type: 'NPC_',
        editor_id: 'MQ101Irileth',
        source_text: 'The Jarl needs your help, now.',
    },
]);

const createBrowserMockPreviewPage = (pageArg?: unknown, pageSizeArg?: unknown) => {
    const rows = createBrowserMockRows();
    const page = Math.max(1, toInt(pageArg, 1));
    const pageSize = Math.max(1, toInt(pageSizeArg, 50));
    const start = (page - 1) * pageSize;
    const end = start + pageSize;
    return {
        file_id: BROWSER_MOCK_FILE_ID,
        page,
        page_size: pageSize,
        total_rows: rows.length,
        rows: rows.slice(start, end),
    };
};

const createBrowserMockLoadedFile = () => {
    const preview = createBrowserMockPreviewPage(1, 50);
    return {
        file_id: BROWSER_MOCK_FILE_ID,
        file_path: BROWSER_MOCK_SOURCE_FILE,
        file_name: 'dialogue_sample_01.json',
        parse_status: 'loaded',
        preview_count: preview.total_rows,
        preview,
    };
};

const createBrowserMockTask = () => {
    const now = new Date().toISOString();
    return {
        id: BROWSER_MOCK_TASK_ID,
        name: 'Browser Mock Translation Flow',
        type: 'translation_project',
        status: 'running',
        phase: 'load',
        progress: 20,
        error_msg: '',
        metadata: {
            source: 'wails-bridge-shim',
            mock_loaded: true,
        },
        created_at: now,
        updated_at: now,
    };
};

const isTranslationFlowRoute = (): boolean => window.location.hash.includes('/translation_flow');
const createBrowserMockTaskList = () => (isTranslationFlowRoute() ? [createBrowserMockTask()] : []);

const RUNTIME_FALLBACKS: RuntimeMap = {
    EventsOn: () => () => undefined,
    EventsOnMultiple: () => () => undefined,
    EventsOff: () => undefined,
    EventsOffAll: () => undefined,
    EventsEmit: () => undefined,
    LogPrint: () => undefined,
    LogTrace: () => undefined,
    LogDebug: () => undefined,
    LogInfo: () => undefined,
    LogWarning: () => undefined,
    LogError: () => undefined,
    LogFatal: () => undefined,
};

const CONTROLLER_FALLBACKS: Record<string, ControllerMap> = {
    ConfigController: {
        ConfigDelete: async () => undefined,
        ConfigGet: async () => '',
        ConfigGetAll: async () => ({}),
        ConfigSet: async () => undefined,
        ConfigSetMany: async () => undefined,
        SetContext: async () => undefined,
        UIStateDelete: async () => undefined,
        UIStateGetJSON: async () => '',
        UIStateSetJSON: async () => undefined,
    },
    TaskController: {
        CancelTask: async () => undefined,
        DeleteTask: async () => undefined,
        GetActiveTasks: async () => createBrowserMockTaskList(),
        GetAllTasks: async () => createBrowserMockTaskList(),
        GetTranslationFlowTerminology: async (...args) => ({
            ...EMPTY_TERMINOLOGY_RESULT(),
            task_id: resolveTaskIDFromArgs(args),
        }),
        ListLoadedTranslationFlowFiles: async (...args) =>
            isTranslationFlowRoute()
                ? {
                    task_id: resolveTaskIDFromArgs(args),
                    files: [createBrowserMockLoadedFile()],
                }
                : {task_id: '', files: []},
        ListTranslationFlowTerminologyTargets: async (...args) =>
            isTranslationFlowRoute()
                ? {
                    task_id: resolveTaskIDFromArgs(args),
                    page: Math.max(1, toInt(args[1], 1)),
                    page_size: Math.max(1, toInt(args[2], 50)),
                    total_rows: 2,
                    rows: [
                        {
                            id: 'term-1',
                            record_type: 'NPC_:FULL',
                            editor_id: 'MQ101Farengar',
                            source_text: 'Farengar Secret-Fire',
                            variant: 'full',
                            source_file: 'dialogue_sample_01.json',
                        },
                        {
                            id: 'term-2',
                            record_type: 'NPC_:SHRT',
                            editor_id: 'MQ101Farengar',
                            source_text: 'Farengar',
                            variant: 'short',
                            source_file: 'dialogue_sample_01.json',
                        },
                    ],
                }
                : EMPTY_TERMINOLOGY_TARGET_PAGE(),
        ListTranslationFlowPreviewRows: async (...args) => {
            if (!isTranslationFlowRoute()) {
                return {
                    ...createBrowserMockPreviewPage(1, 50),
                    total_rows: 0,
                    rows: [],
                };
            }
            const fileId = Math.max(1, toInt(args[0], BROWSER_MOCK_FILE_ID));
            const preview = createBrowserMockPreviewPage(args[1], args[2]);
            return {
                ...preview,
                file_id: fileId,
            };
        },
        LoadTranslationFlowFiles: async (...args) =>
            isTranslationFlowRoute()
                ? {
                    task_id: resolveTaskIDFromArgs(args),
                    files: [createBrowserMockLoadedFile()],
                }
                : {task_id: '', files: []},
        ResumeTask: async () => undefined,
        RunTranslationFlowTerminology: async (...args) => ({
            task_id: resolveTaskIDFromArgs(args),
            status: isTranslationFlowRoute() ? 'completed' : 'pending',
            saved_count: isTranslationFlowRoute() ? 12 : 0,
            failed_count: 0,
        }),
        SetContext: async () => undefined,
        SetTranslationFlowWorkflow: async () => undefined,
    },
    ModelCatalogController: {
        ListModels: async () => [],
        SetContext: async () => undefined,
    },
    FileDialogController: {
        SelectFiles: async () => [],
        SelectJSONFile: async () => '',
        SelectTranslationInputFiles: async () => [],
        SetContext: async () => undefined,
    },
    PersonaTaskController: {
        CancelTask: async () => undefined,
        GetAllTasks: async () => [],
        GetTaskRequestState: async () => ({total: 0, completed: 0, failed: 0, canceled: 0}),
        GetTaskRequests: async () => [],
        ResumeMasterPersonaTask: async () => undefined,
        ResumeTask: async () => undefined,
        SetContext: async () => undefined,
        StartMasterPersonTask: async () => '',
    },
    PersonaController: {
        ListDialoguesByPersonaID: async () => [],
        ListNPCs: async () => [],
        SetContext: async () => undefined,
    },
    DictionaryController: {
        DictDeleteEntry: async () => undefined,
        DictDeleteSource: async () => undefined,
        DictGetEntries: async () => [],
        DictGetEntriesPaginated: async () => ({entries: [], totalCount: 0}),
        DictGetSources: async () => [],
        DictSearchAllEntriesPaginated: async () => ({entries: [], totalCount: 0}),
        DictStartImport: async () => '',
        DictUpdateEntry: async () => undefined,
        SetContext: async () => undefined,
    },
};

const UNKNOWN_CONTROLLER_METHOD = async (): Promise<unknown> => undefined;
let shimNoticePrinted = false;

const ensureRuntimeMap = (runtime: RuntimeMap): boolean => {
    let patched = false;
    for (const [methodName, fallback] of Object.entries(RUNTIME_FALLBACKS)) {
        if (typeof runtime[methodName] !== 'function') {
            runtime[methodName] = fallback;
            patched = true;
        }
    }
    return patched;
};

const withControllerMethodProxy = (controller: ControllerMap): ControllerMap =>
    new Proxy(controller, {
        get(target, property, receiver) {
            if (typeof property !== 'string') {
                return Reflect.get(target, property, receiver) as ControllerMethod;
            }
            const existing = Reflect.get(target, property, receiver);
            if (typeof existing === 'function') {
                return existing as ControllerMethod;
            }
            Reflect.set(target, property, UNKNOWN_CONTROLLER_METHOD);
            return UNKNOWN_CONTROLLER_METHOD;
        },
    });

const ensureControllerRegistry = (registry: ControllerRegistry): {registry: ControllerRegistry; patched: boolean} => {
    let patched = false;
    const controllerMap: ControllerRegistry = registry;

    for (const [controllerName, fallbackMethods] of Object.entries(CONTROLLER_FALLBACKS)) {
        const existingController = controllerMap[controllerName];
        const normalizedController: ControllerMap =
            existingController && typeof existingController === 'object' ? existingController : {};
        if (normalizedController !== existingController) {
            controllerMap[controllerName] = normalizedController;
            patched = true;
        }
        for (const [methodName, fallback] of Object.entries(fallbackMethods)) {
            if (typeof normalizedController[methodName] !== 'function') {
                normalizedController[methodName] = fallback;
                patched = true;
            }
        }
        if (!('$$shimProxy' in normalizedController)) {
            const proxied = withControllerMethodProxy(normalizedController);
            Reflect.set(proxied, '$$shimProxy', true);
            controllerMap[controllerName] = proxied;
            patched = true;
        }
    }

    const proxiedRegistry = new Proxy(controllerMap, {
        get(target, property, receiver) {
            if (typeof property !== 'string') {
                return Reflect.get(target, property, receiver) as ControllerMap;
            }
            const existing = Reflect.get(target, property, receiver);
            if (existing && typeof existing === 'object') {
                if (!Reflect.has(existing as object, '$$shimProxy')) {
                    const proxied = withControllerMethodProxy(existing as ControllerMap);
                    Reflect.set(proxied, '$$shimProxy', true);
                    Reflect.set(target, property, proxied);
                    return proxied;
                }
                return existing as ControllerMap;
            }
            const fallbackController = withControllerMethodProxy({});
            Reflect.set(fallbackController, '$$shimProxy', true);
            Reflect.set(target, property, fallbackController);
            return fallbackController;
        },
    });

    return {registry: proxiedRegistry, patched};
};

/**
 * Wails runtime が無いブラウザ実行時のみ、安全な no-op bridge を注入する。
 * 既存の WebView 注入オブジェクトがある場合は上書きしない。
 */
export function ensureWailsBridge(): void {
    const win = window as WailsWindow;
    const runtime =
        win.runtime && typeof win.runtime === 'object'
            ? win.runtime
            : ((win.runtime = {}) as RuntimeMap);

    const goRoot =
        win.go && typeof win.go === 'object'
            ? win.go
            : ((win.go = {}) as NonNullable<WailsWindow['go']>);
    const controllerRegistry =
        goRoot.controller && typeof goRoot.controller === 'object'
            ? goRoot.controller
            : ((goRoot.controller = {}) as ControllerRegistry);

    const runtimePatched = ensureRuntimeMap(runtime);
    const controllerResult = ensureControllerRegistry(controllerRegistry);
    goRoot.controller = controllerResult.registry;

    if ((runtimePatched || controllerResult.patched) && !shimNoticePrinted) {
        console.info('[wails-bridge-shim] Wails runtime 未注入のため no-op bridge を有効化しました。');
        shimNoticePrinted = true;
    }
}
