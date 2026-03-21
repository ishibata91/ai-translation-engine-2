import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import {useTranslationFlow} from './useTranslationFlow';
import {ConfigGetAll, ConfigSet} from '../../../wailsjs/go/controller/ConfigController';
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

function PaginationProbe() {
    const {state, actions} = useTranslationFlow();

    return (
        <div>
            <div data-testid="task-id">{state.taskId}</div>
            <div data-testid="target-page">{state.terminologyTargetPage.page}</div>
            <div data-testid="target-rows">{state.terminologyTargetPage.rows.map((row) => row.id).join(',')}</div>
            <button type="button" onClick={() => void actions.handleTerminologyTargetPageChange(2)}>
                next page
            </button>
        </div>
    );
}

function ConfigProbe() {
    const {state, actions} = useTranslationFlow();

    return (
        <div>
            <div data-testid="terminology-provider">{state.terminologyConfig.provider}</div>
            <div data-testid="terminology-model">{state.terminologyConfig.model}</div>
            <button
                type="button"
                onClick={() =>
                    actions.handleTerminologyConfigChange({
                        ...state.terminologyConfig,
                        provider: 'xai',
                    })}
            >
                switch provider
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

const asConfigMap = (value: Record<string, string>): Record<string, string> => value;

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
                return asConfigMap({model: 'gemini-2.5-flash', provider: 'gemini'});
            }
            return asConfigMap({});
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

describe('useTranslationFlow terminology pagination', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        vi.mocked(ConfigGetAll).mockResolvedValue(asConfigMap({}));
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
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'existing-task',
            files: [],
        }));
        vi.mocked(ListTranslationFlowTerminologyTargets).mockImplementation(async (_taskId: string, page: number) =>
            asTerminologyTargetPage({
                task_id: 'existing-task',
                page,
                page_size: 50,
                total_rows: 60,
                rows: page === 1
                    ? [{
                        id: 'row-1',
                        record_type: 'NPC_:FULL',
                        editor_id: 'NPC_A_01',
                        source_text: 'NPC Name A-01',
                        translated_text: '',
                        translation_state: 'missing',
                        variant: 'full',
                        source_file: 'Update.esm.extract.json',
                    }]
                    : [{
                        id: 'row-51',
                        record_type: 'NPC_:FULL',
                        editor_id: 'NPC_B_01',
                        source_text: 'NPC Name B-01',
                        translated_text: '',
                        translation_state: 'missing',
                        variant: 'full',
                        source_file: 'Update.esm.extract.json',
                    }],
            }),
        );
    });

    afterEach(() => {
        cleanup();
    });

    it('用語対象のページ変更後に task 初期化 effect で 1 ページ目へ戻さない', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PaginationProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });
        await waitFor(() => {
            expect(screen.getByTestId('target-page')).toHaveTextContent('1');
        });

        fireEvent.click(screen.getByRole('button', {name: 'next page'}));

        await waitFor(() => {
            expect(screen.getByTestId('target-page')).toHaveTextContent('2');
        });
        expect(screen.getByTestId('target-rows')).toHaveTextContent('row-51');
        await new Promise((resolve) => window.setTimeout(resolve, 0));
        expect(screen.getByTestId('target-page')).toHaveTextContent('2');
        expect(screen.getByTestId('target-rows')).toHaveTextContent('row-51');
    });
});

describe('useTranslationFlow terminology config namespace', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        vi.mocked(ConfigSet).mockResolvedValue(undefined as never);
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

    it('selected_provider と provider namespace から設定を hydrate する', async () => {
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return asConfigMap({selected_provider: 'xai'});
            }
            if (namespace === 'translation_flow.terminology.llm.xai') {
                return asConfigMap({
                    model: 'grok-3-beta',
                    endpoint: 'https://api.x.ai/v1',
                    api_key: 'x-api-key',
                    temperature: '0.4',
                    context_length: '8192',
                    sync_concurrency: '4',
                    bulk_strategy: 'batch',
                });
            }
            return asConfigMap({});
        });

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <ConfigProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('terminology-provider')).toHaveTextContent('xai');
        });
        expect(screen.getByTestId('terminology-model')).toHaveTextContent('grok-3-beta');
        expect(ConfigGetAll).toHaveBeenCalledWith('translation_flow.terminology.llm');
        expect(ConfigGetAll).toHaveBeenCalledWith('translation_flow.terminology.llm.xai');
    });

    it('provider 切替時に provider namespace を読み直して selected_provider を保存する', async () => {
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return asConfigMap({selected_provider: 'gemini'});
            }
            if (namespace === 'translation_flow.terminology.llm.gemini') {
                return asConfigMap({
                    model: 'gemini-2.5-flash',
                    endpoint: '',
                    api_key: 'gemini-key',
                    temperature: '0.7',
                    context_length: '16384',
                    sync_concurrency: '2',
                    bulk_strategy: 'sync',
                });
            }
            if (namespace === 'translation_flow.terminology.llm.xai') {
                return asConfigMap({
                    model: 'grok-3-beta',
                    endpoint: 'https://api.x.ai/v1',
                    api_key: 'x-api-key',
                    temperature: '0.4',
                    context_length: '8192',
                    sync_concurrency: '4',
                    bulk_strategy: 'batch',
                });
            }
            return asConfigMap({});
        });

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <ConfigProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('terminology-provider')).toHaveTextContent('gemini');
        });

        fireEvent.click(screen.getByRole('button', {name: 'switch provider'}));

        await waitFor(() => {
            expect(screen.getByTestId('terminology-provider')).toHaveTextContent('xai');
        });
        await waitFor(() => {
            expect(ConfigGetAll).toHaveBeenCalledWith('translation_flow.terminology.llm.xai');
        });
        await waitFor(() => {
            expect(ConfigSet).toHaveBeenCalledWith(
                'translation_flow.terminology.llm',
                'selected_provider',
                'xai',
            );
        });
        expect(ConfigSet).toHaveBeenCalledWith(
            'translation_flow.terminology.llm.xai',
            'model',
            'grok-3-beta',
        );
    });
});
