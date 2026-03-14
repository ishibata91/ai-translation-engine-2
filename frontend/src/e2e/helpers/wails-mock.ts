import {Page} from '@playwright/test';
import {
  DICTIONARY_BUILDER_ENTRIES_BY_SOURCE_ID,
  DICTIONARY_BUILDER_SOURCES,
} from '../fixtures/dictionary-builder/mock-data';

type DictionaryMockFixture = {
  entriesBySourceId: Record<number, Array<Record<string, string | number>>>;
  sources: Array<Record<string, string | number>>;
};

export async function installWailsMocks(page: Page): Promise<void> {
  const dictionaryFixture: DictionaryMockFixture = {
    entriesBySourceId: DICTIONARY_BUILDER_ENTRIES_BY_SOURCE_ID,
    sources: DICTIONARY_BUILDER_SOURCES,
  };

  await page.addInitScript((fixture: DictionaryMockFixture) => {
    const toLowerText = (value: unknown): string =>
      String(value ?? '').toLowerCase();

    const paginateEntries = (
      entries: Array<Record<string, string | number>>,
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
      entries: Array<Record<string, string | number>>,
      filters: Record<string, unknown>,
    ): Array<Record<string, string | number>> =>
      entries.filter((entry) =>
        Object.entries(filters).every(([key, rawFilterValue]) => {
          const filterValue = toLowerText(rawFilterValue).trim();
          if (filterValue.length === 0) {
            return true;
          }

          return toLowerText(entry[key]).includes(filterValue);
        }),
      );

    const runtime = {
      EventsOnMultiple: () => () => undefined,
      EventsOff: () => undefined,
      EventsOffAll: () => undefined,
      EventsEmit: () => undefined,
    };

    const taskController = {
      GetActiveTasks: async () => [],
      GetAllTasks: async () => [],
      ResumeTask: async () => undefined,
      CancelTask: async () => undefined,
      SetContext: async () => undefined,
    };

    const personaTaskController = {
      GetAllTasks: async () => [],
      GetTaskRequestState: async () => ({total: 0, completed: 0}),
      GetTaskRequests: async () => [],
      StartMasterPersonTask: async () => 'persona-task-e2e',
      ResumeMasterPersonaTask: async () => undefined,
      ResumeTask: async () => undefined,
      CancelTask: async () => undefined,
      SetContext: async () => undefined,
    };

    const personaController = {
      ListNPCs: async () => [],
      ListDialoguesByPersonaID: async () => [],
      SetContext: async () => undefined,
    };

    const configController = {
      UIStateGetJSON: async () => '',
      UIStateSetJSON: async () => undefined,
      UIStateDelete: async () => undefined,
      ConfigGet: async () => '',
      ConfigSet: async () => undefined,
      ConfigDelete: async () => undefined,
      ConfigGetAll: async () => ({}),
      ConfigSetMany: async () => undefined,
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
        const baseEntries = fixture.entriesBySourceId[sourceId] ?? [];
        const filteredEntries = applyColumnFilters(baseEntries, filters ?? {});
        return paginateEntries(filteredEntries, pageNumber, pageSize);
      },
      DictGetSources: async () => fixture.sources,
      DictSearchAllEntriesPaginated: async (
        query: string,
        filters: Record<string, unknown>,
        pageNumber: number,
        pageSize: number,
      ) => {
        const normalizedQuery = toLowerText(query).trim();
        const allEntries = Object.values(fixture.entriesBySourceId).flat();
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
    win.go.controller.DictionaryController = {
      ...dictionaryController,
      ...(win.go.controller.DictionaryController as Record<string, unknown> | undefined),
    };
    win.go.controller.FileDialogController = {
      ...fileDialogController,
      ...(win.go.controller.FileDialogController as Record<string, unknown> | undefined),
    };
  }, dictionaryFixture);
}
