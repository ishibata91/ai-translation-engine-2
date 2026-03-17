import {Page} from '@playwright/test';
import {DASHBOARD_TASKS} from '../fixtures/dashboard/mock-data';
import {
  DICTIONARY_BUILDER_ENTRIES_BY_SOURCE_ID,
  DICTIONARY_BUILDER_SOURCES,
} from '../fixtures/dictionary-builder/mock-data';
import {
  MASTER_PERSONA_DIALOGUES_BY_PERSONA_ID,
  MASTER_PERSONA_LLM_CONFIG_BY_NAMESPACE,
  MASTER_PERSONA_LLM_ROOT_CONFIG,
  MASTER_PERSONA_MODEL_CATALOG_BY_PROVIDER,
  MASTER_PERSONA_PROMPT_CONFIG,
  MASTER_PERSONA_REQUIRED_NPCS,
  MASTER_PERSONA_SELECTED_JSON_PATH,
  MASTER_PERSONA_STARTED_TASK_ID,
} from '../fixtures/master-persona/mock-data';
import {
  TRANSLATION_FLOW_FILE_PAYLOADS,
  TRANSLATION_FLOW_SELECTED_FILES,
  TRANSLATION_FLOW_TASK_ID,
  TRANSLATION_FLOW_TERMINOLOGY_COMPLETED_SUMMARY,
  TRANSLATION_FLOW_TERMINOLOGY_LLM_CONFIG,
  TRANSLATION_FLOW_TERMINOLOGY_PROMPT_CONFIG,
  TRANSLATION_FLOW_TERMINOLOGY_SUMMARY,
} from '../fixtures/translation-flow/mock-data';

type DashboardMockFixture = {
  tasks: Array<Record<string, string | number | Record<string, unknown>>>;
};

type DictionaryMockFixture = {
  entriesBySourceId: Record<number, Array<Record<string, string | number | null>>>;
  sources: Array<Record<string, string | number | null>>;
};

type MasterPersonaMockFixture = {
  dialoguesByPersonaId: Record<number, Array<Record<string, string>>>;
  llmConfigByNamespace: Record<string, Record<string, string>>;
  llmRootConfig: Record<string, string>;
  modelCatalogByProvider: Record<string, Array<{
    capability: {
      supports_batch: boolean;
    };
    display_name: string;
    id: string;
  }>>;
  npcs: Array<Record<string, number | string>>;
  promptConfig: Record<string, string>;
  selectedJsonPath: string;
  startedTaskId: string;
};

type TranslationFlowFileFixture = {
  file_id: number;
  file_path: string;
  file_name: string;
  parse_status: string;
  rows: Array<{
    id: string;
    section: string;
    record_type: string;
    editor_id: string;
    source_text: string;
  }>;
};

type TranslationFlowMockFixture = {
  filePayloads: Record<string, TranslationFlowFileFixture>;
  selectedFiles: string[];
  taskId: string;
  terminologyCompletedSummary: Record<string, number | string>;
  terminologyLLMConfig: Record<string, string>;
  terminologyPromptConfig: Record<string, string>;
  terminologySummary: Record<string, number | string>;
};

type WailsMockFixture = {
  dashboard: DashboardMockFixture;
  dictionary: DictionaryMockFixture;
  masterPersona: MasterPersonaMockFixture;
  translationFlow: TranslationFlowMockFixture;
};

export async function installWailsMocks(page: Page): Promise<void> {
  const fixture: WailsMockFixture = {
    dashboard: {
      tasks: DASHBOARD_TASKS.map((task) => ({...task})),
    },
    dictionary: {
      entriesBySourceId: DICTIONARY_BUILDER_ENTRIES_BY_SOURCE_ID,
      sources: DICTIONARY_BUILDER_SOURCES,
    },
    masterPersona: {
      dialoguesByPersonaId: MASTER_PERSONA_DIALOGUES_BY_PERSONA_ID,
      llmConfigByNamespace: MASTER_PERSONA_LLM_CONFIG_BY_NAMESPACE,
      llmRootConfig: MASTER_PERSONA_LLM_ROOT_CONFIG,
      modelCatalogByProvider: MASTER_PERSONA_MODEL_CATALOG_BY_PROVIDER,
      npcs: MASTER_PERSONA_REQUIRED_NPCS,
      promptConfig: MASTER_PERSONA_PROMPT_CONFIG,
      selectedJsonPath: MASTER_PERSONA_SELECTED_JSON_PATH,
      startedTaskId: MASTER_PERSONA_STARTED_TASK_ID,
    },
    translationFlow: {
      filePayloads: TRANSLATION_FLOW_FILE_PAYLOADS,
      selectedFiles: [...TRANSLATION_FLOW_SELECTED_FILES],
      taskId: TRANSLATION_FLOW_TASK_ID,
      terminologyCompletedSummary: TRANSLATION_FLOW_TERMINOLOGY_COMPLETED_SUMMARY,
      terminologyLLMConfig: TRANSLATION_FLOW_TERMINOLOGY_LLM_CONFIG,
      terminologyPromptConfig: TRANSLATION_FLOW_TERMINOLOGY_PROMPT_CONFIG,
      terminologySummary: TRANSLATION_FLOW_TERMINOLOGY_SUMMARY,
    },
  };

  await page.addInitScript((mockFixture: WailsMockFixture) => {
    const toLowerText = (value: unknown): string =>
      String(value ?? '').toLowerCase();

    const paginateEntries = (
      entries: Array<Record<string, string | number | null>>,
      pageNumber: number,
      pageSize: number,
    ) => {
      const safePage = Number.isFinite(pageNumber) && pageNumber > 0 ? pageNumber : 1;
      const safePageSize = Number.isFinite(pageSize) && pageSize > 0 ? pageSize : 500;
      const start = (safePage - 1) * safePageSize;
      const end = start + safePageSize;

      return {
        entries: entries.slice(start, end),
        totalCount: entries.length,
      };
    };

    const applyColumnFilters = (
      entries: Array<Record<string, string | number | null>>,
      filters: Record<string, unknown>,
    ): Array<Record<string, string | number | null>> =>
      entries.filter((entry) =>
        Object.entries(filters).every(([key, rawFilterValue]) => {
          const filterValue = toLowerText(rawFilterValue).trim();
          if (filterValue.length === 0) {
            return true;
          }

          return toLowerText(entry[key]).includes(filterValue);
        }),
      );

    const configStore = new Map<string, Record<string, string>>();
    configStore.set('master_persona.llm', {...mockFixture.masterPersona.llmRootConfig});
    configStore.set('master_persona.prompt', {...mockFixture.masterPersona.promptConfig});
    configStore.set('translation_flow.terminology.llm', {...mockFixture.translationFlow.terminologyLLMConfig});
    configStore.set('translation_flow.terminology.prompt', {...mockFixture.translationFlow.terminologyPromptConfig});

    for (const [namespace, values] of Object.entries(mockFixture.masterPersona.llmConfigByNamespace)) {
      configStore.set(namespace, {...values});
    }

    let personaTask: Record<string, unknown> | null = null;
    let terminologySummary = {...mockFixture.translationFlow.terminologySummary};

    const dashboardTasks = [...mockFixture.dashboard.tasks];
    const translationFilesByTask = new Map<string, TranslationFlowFileFixture[]>();

    const normalizePage = (page: number): number => {
      if (!Number.isFinite(page) || page < 1) {
        return 1;
      }
      return Math.floor(page);
    };

    const normalizePageSize = (pageSize: number): number => {
      if (!Number.isFinite(pageSize) || pageSize < 1) {
        return 50;
      }
      return Math.floor(pageSize);
    };

    const buildTranslationPreviewPage = (
      file: TranslationFlowFileFixture,
      page: number,
      pageSize: number,
    ) => {
      const safePage = normalizePage(page);
      const safePageSize = normalizePageSize(pageSize);
      const start = (safePage - 1) * safePageSize;
      const end = start + safePageSize;
      return {
        file_id: file.file_id,
        page: safePage,
        page_size: safePageSize,
        total_rows: file.rows.length,
        rows: file.rows.slice(start, end),
      };
    };

    const buildLoadedTranslationFile = (
      file: TranslationFlowFileFixture,
      page: number,
      pageSize: number,
    ) => ({
      file_id: file.file_id,
      file_path: file.file_path,
      file_name: file.file_name,
      parse_status: file.parse_status,
      preview_count: file.rows.length,
      preview: buildTranslationPreviewPage(file, page, pageSize),
    });

    const runtime = {
      EventsOn: () => () => undefined,
      EventsOnMultiple: () => () => undefined,
      EventsOff: () => undefined,
      EventsOffAll: () => undefined,
      EventsEmit: () => undefined,
    };

    const taskController = {
      GetActiveTasks: async () => [...dashboardTasks],
      GetAllTasks: async () => [...dashboardTasks],
      GetTranslationFlowTerminology: async () => ({...terminologySummary}),
      ListLoadedTranslationFlowFiles: async (taskID: string) => {
        const files = translationFilesByTask.get(taskID) ?? [];
        return {
          task_id: taskID,
          files: files.map((file) => buildLoadedTranslationFile(file, 1, 50)),
        };
      },
      ListTranslationFlowPreviewRows: async (fileID: number, page: number, pageSize: number) => {
        for (const files of translationFilesByTask.values()) {
          const file = files.find((entry) => entry.file_id === fileID);
          if (file) {
            return buildTranslationPreviewPage(file, page, pageSize);
          }
        }
        return {
          file_id: fileID,
          page: normalizePage(page),
          page_size: normalizePageSize(pageSize),
          total_rows: 0,
          rows: [],
        };
      },
      LoadTranslationFlowFiles: async (taskID: string, filePaths: string[]) => {
        const loadedFiles = filePaths
          .map((path) => mockFixture.translationFlow.filePayloads[path])
          .filter((file): file is TranslationFlowFileFixture => Boolean(file));
        translationFilesByTask.set(taskID, loadedFiles);
        terminologySummary = {
          ...mockFixture.translationFlow.terminologySummary,
          task_id: taskID,
        };
        return {
          task_id: taskID,
          files: loadedFiles.map((file) => buildLoadedTranslationFile(file, 1, 50)),
        };
      },
      RunTranslationFlowTerminology: async (taskID: string) => {
        terminologySummary = {
          ...mockFixture.translationFlow.terminologyCompletedSummary,
          task_id: taskID,
        };
        return {...terminologySummary};
      },
      ResumeTask: async () => undefined,
      CancelTask: async () => undefined,
      DeleteTask: async (taskID: string) => {
        const targetIndex = dashboardTasks.findIndex((task) => task.id === taskID);
        if (targetIndex >= 0) {
          dashboardTasks.splice(targetIndex, 1);
        }
      },
      SetContext: async () => undefined,
    };

    const personaTaskController = {
      GetAllTasks: async () => (personaTask ? [personaTask] : []),
      GetTaskRequestState: async () => ({total: 2, completed: 0}),
      GetTaskRequests: async () => [],
      StartMasterPersonTask: async (payload: Record<string, unknown>) => {
        personaTask = {
          id: mockFixture.masterPersona.startedTaskId,
          type: 'persona_extraction',
          status: 'running',
          progress: 0,
          metadata: {
            overwrite_existing: Boolean(payload.overwrite_existing),
            request_count: 2,
            resume_cursor: 0,
            source_json_path: String(payload.source_json_path ?? ''),
          },
          updated_at: new Date().toISOString(),
        };
        return mockFixture.masterPersona.startedTaskId;
      },
      ResumeMasterPersonaTask: async () => undefined,
      ResumeTask: async (taskID: string) => {
        if (personaTask && personaTask.id === taskID) {
          personaTask = {
            ...personaTask,
            status: 'running',
            updated_at: new Date().toISOString(),
          };
        }
      },
      CancelTask: async (taskID: string) => {
        if (personaTask && personaTask.id === taskID) {
          personaTask = {
            ...personaTask,
            status: 'cancelled',
            updated_at: new Date().toISOString(),
          };
        }
      },
      SetContext: async () => undefined,
    };

    const personaController = {
      ListNPCs: async () => mockFixture.masterPersona.npcs,
      ListDialoguesByPersonaID: async (personaID: number) =>
        mockFixture.masterPersona.dialoguesByPersonaId[personaID] ?? [],
      SetContext: async () => undefined,
    };

    const configController = {
      UIStateGetJSON: async () => '',
      UIStateSetJSON: async () => undefined,
      UIStateDelete: async () => undefined,
      ConfigGet: async (namespace: string, key: string) => {
        const namespaceStore = configStore.get(namespace) ?? {};
        return namespaceStore[key] ?? '';
      },
      ConfigSet: async (namespace: string, key: string, value: string) => {
        const namespaceStore = configStore.get(namespace) ?? {};
        namespaceStore[key] = value;
        configStore.set(namespace, namespaceStore);
      },
      ConfigDelete: async (namespace: string, key: string) => {
        const namespaceStore = configStore.get(namespace) ?? {};
        delete namespaceStore[key];
        configStore.set(namespace, namespaceStore);
      },
      ConfigGetAll: async (namespace: string) => ({...(configStore.get(namespace) ?? {})}),
      ConfigSetMany: async (namespace: string, values: Record<string, string>) => {
        const namespaceStore = configStore.get(namespace) ?? {};
        configStore.set(namespace, {
          ...namespaceStore,
          ...values,
        });
      },
      SetContext: async () => undefined,
    };

    const modelCatalogController = {
      ListModels: async (request: Record<string, unknown>) => {
        const provider = String(request.provider ?? '');
        const namespace = String(request.namespace ?? '');
        if (namespace === 'translation_flow.terminology' && provider === 'lmstudio') {
          return [{
            id: 'local-terminology-model',
            display_name: 'local-terminology-model',
            capability: {
              supports_batch: false,
            },
          }];
        }
        return mockFixture.masterPersona.modelCatalogByProvider[provider] ?? [];
      },
      SetContext: async () => undefined,
    };

    const dictionaryController = {
      DictDeleteEntry: async () => undefined,
      DictDeleteSource: async () => undefined,
      DictGetEntries: async () => [],
      DictGetEntriesPaginated: async (
        sourceId: number,
        _query: string,
        filters: Record<string, unknown>,
        pageNumber: number,
        pageSize: number,
      ) => {
        const baseEntries = mockFixture.dictionary.entriesBySourceId[sourceId] ?? [];
        const filteredEntries = applyColumnFilters(baseEntries, filters ?? {});
        return paginateEntries(filteredEntries, pageNumber, pageSize);
      },
      DictGetSources: async () => mockFixture.dictionary.sources,
      DictSearchAllEntriesPaginated: async (
        query: string,
        filters: Record<string, unknown>,
        pageNumber: number,
        pageSize: number,
      ) => {
        const normalizedQuery = toLowerText(query).trim();
        const allEntries = Object.values(mockFixture.dictionary.entriesBySourceId).flat();
        const queryMatchedEntries = normalizedQuery.length === 0
          ? allEntries
          : allEntries.filter((entry) =>
            [entry.edid, entry.source_text, entry.dest_text]
              .map((value) => toLowerText(value))
              .some((value) => value.includes(normalizedQuery)),
          );
        const filteredEntries = applyColumnFilters(queryMatchedEntries, filters ?? {});
        return paginateEntries(filteredEntries, pageNumber, pageSize);
      },
      DictStartImport: async () => 'dictionary-task-e2e',
      DictUpdateEntry: async () => undefined,
      SetContext: async () => undefined,
    };

    const fileDialogController = {
      SelectFiles: async () => [],
      SelectJSONFile: async () => mockFixture.masterPersona.selectedJsonPath,
      SelectTranslationInputFiles: async () => [...mockFixture.translationFlow.selectedFiles],
      SetContext: async () => undefined,
    };

    const win = window as Window & {
      runtime?: Record<string, unknown>;
      go?: {
        controller?: Record<string, unknown>;
      };
    };

    win.runtime = {
      ...runtime,
      ...(win.runtime ?? {}),
    };

    win.go = win.go ?? {};
    win.go.controller = win.go.controller ?? {};
    win.go.controller.TaskController = {
      ...taskController,
      ...(win.go.controller.TaskController as Record<string, unknown> | undefined),
    };
    win.go.controller.PersonaTaskController = {
      ...personaTaskController,
      ...(win.go.controller.PersonaTaskController as Record<string, unknown> | undefined),
    };
    win.go.controller.PersonaController = {
      ...personaController,
      ...(win.go.controller.PersonaController as Record<string, unknown> | undefined),
    };
    win.go.controller.ConfigController = {
      ...configController,
      ...(win.go.controller.ConfigController as Record<string, unknown> | undefined),
    };
    win.go.controller.ModelCatalogController = {
      ...modelCatalogController,
      ...(win.go.controller.ModelCatalogController as Record<string, unknown> | undefined),
    };
    win.go.controller.DictionaryController = {
      ...dictionaryController,
      ...(win.go.controller.DictionaryController as Record<string, unknown> | undefined),
    };
    win.go.controller.FileDialogController = {
      ...fileDialogController,
      ...(win.go.controller.FileDialogController as Record<string, unknown> | undefined),
    };
  }, fixture);
}
