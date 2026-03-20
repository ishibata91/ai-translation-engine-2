import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import {useTranslationFlow} from './useTranslationFlow';
import {ConfigGetAll} from '../../../wailsjs/go/controller/ConfigController';
import {
    GetAllTasks,
    GetTranslationFlowTerminology,
    ListLoadedTranslationFlowFiles,
    ListTranslationFlowTerminologyTargets,
    RunTranslationFlowTerminology,
} from '../../../wailsjs/go/controller/TaskController';
import type {FrontendTask} from '../../../types/task';

const runtimeEventHandlers = new Map<string, (payload: unknown) => void>();

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
    ListTranslationFlowTerminologyTargets: vi.fn(),
    ListTranslationFlowPreviewRows: vi.fn(),
    LoadTranslationFlowFiles: vi.fn(),
    RunTranslationFlowTerminology: vi.fn(),
}));

vi.mock('../../../wailsjs/runtime/runtime', () => ({
    EventsOn: vi.fn((eventName: string, callback: (payload: unknown) => void) => {
        runtimeEventHandlers.set(eventName, callback);
        return () => {
            runtimeEventHandlers.delete(eventName);
        };
    }),
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

function RunProbe() {
    const {state, actions} = useTranslationFlow();

    return (
        <div>
            <div data-testid="task-id">{state.taskId}</div>
            <div data-testid="terminology-status">{state.terminologyStatusLabel}</div>
            <div data-testid="terminology-error">{state.terminologyErrorMessage}</div>
            <div data-testid="terminology-progress">
                {state.terminologySummary.progressCurrent}/{state.terminologySummary.progressTotal}
            </div>
            <button type="button" onClick={() => void actions.handleRunTerminologyPhase()}>
                run terminology
            </button>
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

const asTerminologyResult = (
    value: unknown,
): Awaited<ReturnType<typeof RunTranslationFlowTerminology>> => value as Awaited<ReturnType<typeof RunTranslationFlowTerminology>>;

const asTerminologyPhaseResult = (
    value: unknown,
): Awaited<ReturnType<typeof GetTranslationFlowTerminology>> => value as Awaited<ReturnType<typeof GetTranslationFlowTerminology>>;

const asTerminologyTargetPage = (
    value: unknown,
): Awaited<ReturnType<typeof ListTranslationFlowTerminologyTargets>> =>
    value as Awaited<ReturnType<typeof ListTranslationFlowTerminologyTargets>>;

describe('useTranslationFlow task resolution', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        vi.mocked(ConfigGetAll).mockResolvedValue({});
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'resolved-task',
            status: 'pending',
            saved_count: 0,
            failed_count: 0,
            progress_mode: 'hidden',
            progress_current: 0,
            progress_total: 0,
            progress_message: '',
        }));
        vi.mocked(ListTranslationFlowTerminologyTargets).mockResolvedValue(asTerminologyTargetPage({
            task_id: 'resolved-task',
            page: 1,
            page_size: 50,
            total_rows: 0,
            rows: [],
        }));
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
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'existing-task',
            status: 'pending',
            saved_count: 0,
            failed_count: 0,
            progress_mode: 'hidden',
            progress_current: 0,
            progress_total: 0,
            progress_message: '',
        }));
        vi.mocked(ListTranslationFlowTerminologyTargets).mockResolvedValue(asTerminologyTargetPage({
            task_id: 'existing-task',
            page: 1,
            page_size: 50,
            total_rows: 0,
            rows: [],
        }));
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

    it('translation task が存在しない場合は空 taskId のまま backend を呼ばない', async () => {
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([]));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <HookProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent(/^$/);
        });
        expect(ListLoadedTranslationFlowFiles).not.toHaveBeenCalled();
        expect(GetTranslationFlowTerminology).not.toHaveBeenCalled();
    });
});

describe('useTranslationFlow terminology run', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return {model: 'gemini-2.5-flash', provider: 'gemini'};
            }
            return {} as Record<string, string>;
        });
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([
            buildTask({id: 'existing-task', updated_at: '2026-03-18T10:00:00.000Z'}),
        ]));
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'existing-task',
            status: 'pending',
            saved_count: 0,
            failed_count: 0,
            progress_mode: 'hidden',
            progress_current: 0,
            progress_total: 0,
            progress_message: '',
        }));
        vi.mocked(ListTranslationFlowTerminologyTargets).mockResolvedValue(asTerminologyTargetPage({
            task_id: 'existing-task',
            page: 1,
            page_size: 50,
            total_rows: 0,
            rows: [],
        }));
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'existing-task',
            files: [],
        }));
    });

    afterEach(() => {
        cleanup();
    });

    it('用語翻訳対象が 0 件のとき未実行ではなく理由を表示する', async () => {
        vi.mocked(RunTranslationFlowTerminology).mockResolvedValue(asTerminologyResult({
            task_id: 'existing-task',
            status: 'pending',
            saved_count: 0,
            failed_count: 0,
            progress_mode: 'hidden',
            progress_current: 0,
            progress_total: 0,
            progress_message: '',
        }));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <RunProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run terminology'}));

        await waitFor(() => {
            expect(screen.getByTestId('terminology-status')).toHaveTextContent('用語翻訳対象なし');
        });
        expect(screen.getByTestId('terminology-error')).toHaveTextContent(
            'ロード済みデータに Terminology 対象 REC がありません。',
        );
    });

    it('completed_partial を完了表示として扱う', async () => {
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'existing-task',
            status: 'completed_partial',
            saved_count: 7,
            failed_count: 1,
            progress_mode: 'hidden',
            progress_current: 8,
            progress_total: 8,
            progress_message: '',
        }));
        vi.mocked(ListTranslationFlowTerminologyTargets).mockResolvedValue(asTerminologyTargetPage({
            task_id: 'existing-task',
            page: 1,
            page_size: 50,
            total_rows: 1,
            rows: [{
                id: 'row-1',
                record_type: 'NPC_:FULL',
                editor_id: 'NPC_B_03',
                source_text: 'NPC Name B-03',
                translated_text: 'NPC 名 B-03',
                translation_state: 'translated',
                variant: 'full',
                source_file: 'Update.esm.extract.json',
            }],
        }));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <RunProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('terminology-status')).toHaveTextContent('単語翻訳完了（一部失敗あり）');
        });
    });

    it('terminology progress イベントで実行中ラベルと進捗値を更新する', async () => {
        vi.mocked(RunTranslationFlowTerminology).mockImplementation(
            async () => new Promise((resolve) => {
                const handler = runtimeEventHandlers.get('translation_flow.terminology.progress');
                handler?.({
                    TaskID: 'existing-task',
                    Status: 'IN_PROGRESS',
                    Current: 2,
                    Total: 8,
                    Message: '2 / 8 件を処理中',
                });
                window.setTimeout(() => {
                    resolve(asTerminologyResult({
                        task_id: 'existing-task',
                        status: 'completed',
                        saved_count: 8,
                        failed_count: 0,
                        progress_mode: 'hidden',
                        progress_current: 8,
                        progress_total: 8,
                        progress_message: '',
                    }));
                }, 10);
            }),
        );
        vi.mocked(ListTranslationFlowTerminologyTargets).mockResolvedValue(asTerminologyTargetPage({
            task_id: 'existing-task',
            page: 1,
            page_size: 50,
            total_rows: 1,
            rows: [{
                id: 'row-1',
                record_type: 'NPC_:FULL',
                editor_id: 'NPC_B_03',
                source_text: 'NPC Name B-03',
                translated_text: 'NPC 名 B-03',
                translation_state: 'translated',
                variant: 'full',
                source_file: 'Update.esm.extract.json',
            }],
        }));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <RunProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run terminology'}));

        await waitFor(() => {
            expect(screen.getByTestId('terminology-status')).toHaveTextContent('2 / 8 件を処理中');
        });
        expect(screen.getByTestId('terminology-progress')).toHaveTextContent('2/8');
    });
});
