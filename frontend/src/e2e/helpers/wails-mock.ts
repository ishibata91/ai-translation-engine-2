import {Page} from '@playwright/test';

export async function installWailsMocks(page: Page): Promise<void> {
  await page.addInitScript(() => {
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
      GetTaskRequestState: async () => ({ total: 0, completed: 0 }),
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
  });
}
