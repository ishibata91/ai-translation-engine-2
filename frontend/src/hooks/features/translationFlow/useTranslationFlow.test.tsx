import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {cleanup, render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import {useTranslationFlow} from './useTranslationFlow';
import {ConfigGetAll} from '../../../wailsjs/go/controller/ConfigController';
import {
    GetAllTasks,
    GetTranslationFlowTerminology,
    ListLoadedTranslationFlowFiles,
} from '../../../wailsjs/go/controller/TaskController';
import type {FrontendTask} from '../../../types/task';

vi.mock('../../../wailsjs/go/controller/ConfigController', () => ({
    ConfigGetAll: vi.fn(),
    ConfigSet: vi.fn(),
}));

vi.mock('../../../wailsjs/go/controller/FileDialogController', () => ({
    SelectTranslationInputFiles: vi.fn(),
}));

vi.mock('../../../wailsjs/go/controller/TaskController', () => ({
    GetAllTasks: vi.fn(),
    GetTranslationFlowTerminology: vi.fn(),
    ListLoadedTranslationFlowFiles: vi.fn(),
    ListTranslationFlowPreviewRows: vi.fn(),
    LoadTranslationFlowFiles: vi.fn(),
    RunTranslationFlowTerminology: vi.fn(),
}));

function HookProbe() {
    const {state} = useTranslationFlow();

    return (
        <div>
            <div data-testid="task-id">{state.taskId}</div>
            <div data-testid="error-message">{state.errorMessage}</div>
        </div>
    );
}

const buildTask = (overrides: Partial<FrontendTask>): FrontendTask => ({
    id: 'task-default',
    name: 'default task',
    type: 'translation_project',
    status: 'paused',
    phase: 'load',
    progress: 0,
    error_msg: '',
    metadata: {},
    created_at: '2026-03-18T00:00:00.000Z',
    updated_at: '2026-03-18T00:00:00.000Z',
    ...overrides,
});

const asLoadResult = (
    value: unknown,
): Awaited<ReturnType<typeof ListLoadedTranslationFlowFiles>> => value as Awaited<ReturnType<typeof ListLoadedTranslationFlowFiles>>;

const asTaskList = (
    value: unknown,
): Awaited<ReturnType<typeof GetAllTasks>> => value as Awaited<ReturnType<typeof GetAllTasks>>;

describe('useTranslationFlow task resolution', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(ConfigGetAll).mockResolvedValue({});
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue({
            task_id: 'resolved-task',
            status: 'pending',
            target_count: 0,
            saved_count: 0,
            failed_count: 0,
        });
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'resolved-task',
            files: [],
        }));
    });

    afterEach(() => {
        cleanup();
    });

    it('削除済み route taskId を再利用せず既存 translation task へフォールバックする', async () => {
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([
            buildTask({ id: 'existing-task', updated_at: '2026-03-18T10:00:00.000Z' }),
        ]));
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue({
            task_id: 'existing-task',
            status: 'pending',
            target_count: 0,
            saved_count: 0,
            failed_count: 0,
        });
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'existing-task',
            files: [],
        }));

        render(
            <MemoryRouter initialEntries={[{pathname: '/translation_flow', state: {taskId: 'deleted-task'}}]}>
                <HookProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });
        expect(ListLoadedTranslationFlowFiles).toHaveBeenCalledWith('existing-task');
        expect(GetTranslationFlowTerminology).toHaveBeenCalledWith('existing-task');
    });

    it('translation task が存在しない場合は空 taskId で backend を呼び新規 task を解決する', async () => {
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([]));
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue({
            task_id: 'created-task',
            status: 'pending',
            target_count: 0,
            saved_count: 0,
            failed_count: 0,
        });
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'created-task',
            files: [],
        }));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <HookProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('created-task');
        });
        expect(ListLoadedTranslationFlowFiles).toHaveBeenCalledWith('');
        expect(GetTranslationFlowTerminology).toHaveBeenCalledWith('');
    });
});
