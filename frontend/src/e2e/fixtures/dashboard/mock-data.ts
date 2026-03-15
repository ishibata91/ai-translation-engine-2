export const DASHBOARD_DELETE_TASK_ID = 'dashboard-delete-task';
export const DASHBOARD_DELETE_TASK_NAME = 'Dashboard Delete Target';

export const DASHBOARD_TASKS = [
  {
    id: DASHBOARD_DELETE_TASK_ID,
    name: DASHBOARD_DELETE_TASK_NAME,
    type: 'translation_project',
    status: 'failed',
    phase: 'review',
    progress: 70,
    error_msg: 'mock failed',
    metadata: {},
    created_at: new Date('2026-01-01T00:00:00.000Z').toISOString(),
    updated_at: new Date('2026-01-02T00:00:00.000Z').toISOString(),
  },
  {
    id: 'dashboard-running-task',
    name: 'Dashboard Running Task',
    type: 'translation_project',
    status: 'running',
    phase: 'load',
    progress: 10,
    error_msg: '',
    metadata: {},
    created_at: new Date('2026-01-03T00:00:00.000Z').toISOString(),
    updated_at: new Date('2026-01-04T00:00:00.000Z').toISOString(),
  },
] as const;
