import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import Dashboard from '../../pages/Dashboard';
import {useTaskStore} from '../../store/taskStore';
import type {FrontendTask} from '../../types/task';
import {DeleteTask} from '../../wailsjs/go/controller/TaskController';

vi.mock('../../wailsjs/runtime/runtime', () => ({
    EventsOn: vi.fn(),
    EventsOff: vi.fn(),
    EventsOffAll: vi.fn(),
}));

vi.mock('../../wailsjs/go/controller/TaskController', () => ({
    GetActiveTasks: vi.fn(),
    GetAllTasks: vi.fn(),
    ResumeTask: vi.fn(),
    DeleteTask: vi.fn(),
    CancelTask: vi.fn(),
}));

vi.mock('../../wailsjs/go/controller/PersonaTaskController', () => ({
    ResumeTask: vi.fn(),
    CancelTask: vi.fn(),
}));

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
    return {
        ...actual,
        useNavigate: () => vi.fn(),
    };
});

const buildTask = (overrides: Partial<FrontendTask>): FrontendTask => ({
    id: 'task-default',
    name: 'default task',
    type: 'translation_project',
    status: 'paused',
    phase: 'phase',
    progress: 42,
    error_msg: '',
    metadata: {},
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T00:00:00.000Z',
    ...overrides,
});

describe('Dashboard', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        useTaskStore.setState({ tasks: {}, isLoading: false });
    });

    afterEach(() => {
        cleanup();
        useTaskStore.setState({ tasks: {}, isLoading: false });
    });

    it('削除可能 status の task にのみ削除ボタンを表示する', () => {
        useTaskStore.setState({
            tasks: {
                paused: buildTask({ id: 'paused', name: 'paused task', status: 'paused' }),
                running: buildTask({ id: 'running', name: 'running task', status: 'running' }),
            },
        });

        render(
            <MemoryRouter>
                <Dashboard />
            </MemoryRouter>,
        );

        expect(screen.getByTestId('dashboard-delete-paused')).toBeInTheDocument();
        expect(screen.queryByTestId('dashboard-delete-running')).not.toBeInTheDocument();
    });

    it('削除確認 modal で確定すると一覧から task を除去する', async () => {
        vi.mocked(DeleteTask).mockResolvedValue(undefined);

        useTaskStore.setState({
            tasks: {
                deletable: buildTask({ id: 'deletable', name: 'deletable task', status: 'failed' }),
            },
        });

        render(
            <MemoryRouter>
                <Dashboard />
            </MemoryRouter>,
        );

        fireEvent.click(screen.getByTestId('dashboard-delete-deletable'));
        expect(screen.getByTestId('dashboard-delete-modal')).toBeInTheDocument();
        expect(screen.getByText('deletable task を削除します。task 本体と task 管理に紐づく中間成果物が削除されます。')).toBeInTheDocument();

        fireEvent.click(screen.getByTestId('dashboard-delete-confirm'));

        await waitFor(() => {
            expect(DeleteTask).toHaveBeenCalledWith('deletable');
        });
        await waitFor(() => {
            expect(screen.queryByText('deletable task')).not.toBeInTheDocument();
        });
    });
});
