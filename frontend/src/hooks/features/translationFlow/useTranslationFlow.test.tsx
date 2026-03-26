import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {act, cleanup, fireEvent, render, screen, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router-dom';
import {useTranslationFlow} from './useTranslationFlow';
import {ConfigGetAll, ConfigSet} from '../../../wailsjs/go/controller/ConfigController';
import {DEFAULT_PERSONA_PROMPT_CONFIG} from '../../../types/masterPersona';
import {
    GetAllTasks,
    GetTranslationFlowTerminology,
    ListLoadedTranslationFlowFiles,
    ListTranslationFlowTerminologyTargets,
    ResumeTask,
    RunTranslationFlowTerminology,
} from '../../../wailsjs/go/controller/TaskController';
import type {FrontendTask} from '../../../types/task';

const runtimeEventHandlers = new Map<string, (payload: unknown) => void>();
const listPersonaTargetsMock = vi.fn();
const runPersonaPhaseMock = vi.fn();
const getPersonaPhaseMock = vi.fn();
const writeTelemetryLogMock = vi.fn().mockResolvedValue(undefined);

const installPersonaBinding = (): void => {
    (window as unknown as {
        go?: {
            controller?: {
                TaskController?: Record<string, unknown>;
                TelemetryController?: {
                    WriteLog: (payload: unknown) => Promise<void>;
                };
            };
        };
    }).go = {
        controller: {
            TaskController: {
                ListTranslationFlowPersonaTargets: listPersonaTargetsMock,
                RunTranslationFlowPersona: runPersonaPhaseMock,
                GetTranslationFlowPersona: getPersonaPhaseMock,
            },
            TelemetryController: {
                WriteLog: writeTelemetryLogMock,
            },
        },
    };
};

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
    ResumeTask: vi.fn(),
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
            <div data-testid="terminology-summary-status">{state.terminologySummary.status}</div>
            <div data-testid="terminology-running-flag">{state.isTerminologyRunning ? 'true' : 'false'}</div>
            <div data-testid="terminology-error">{state.terminologyErrorMessage}</div>
            <div data-testid="terminology-progress">
                {state.terminologySummary.progressCurrent}/{state.terminologySummary.progressTotal}
            </div>
            <div data-testid="terminology-target-total">{state.terminologyTargetPage.totalRows}</div>
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

function PersonaProbe() {
    const {state, actions} = useTranslationFlow();

    return (
        <div>
            <div data-testid="task-id">{state.taskId}</div>
            <div data-testid="active-tab">{state.activeTab}</div>
            <div data-testid="persona-status">{state.personaTargetStatus}</div>
            <div data-testid="persona-label">{state.personaStatusLabel}</div>
            <div data-testid="persona-error">{state.personaErrorMessage}</div>
            <div data-testid="persona-selected-speaker">{state.selectedPersonaTarget?.speakerId ?? ''}</div>
            <div data-testid="terminology-provider">{state.terminologyConfig.provider}</div>
            <div data-testid="persona-provider">{state.personaConfig?.provider ?? ''}</div>
            <button type="button" onClick={() => void actions.handleRunPersonaPhase()}>
                run persona
            </button>
            <button type="button" onClick={() => void actions.handleRetryPersonaPhase()}>
                retry persona
            </button>
            <button type="button" onClick={() => actions.handleTabChange(2)}>
                go persona tab
            </button>
            <button type="button" onClick={() => actions.handleAdvanceFromTerminology()}>
                advance from terminology
            </button>
            <button type="button" onClick={() => actions.handleTabChange(3)}>
                go translation tab
            </button>
            <button type="button" onClick={() => actions.handleAdvanceFromPersona()}>
                advance persona
            </button>
            <button type="button" onClick={() => actions.handleSelectPersonaTarget('Skyrim.esm', 'npc-b')}>
                select npc-b
            </button>
        </div>
    );
}

function MainTranslationProbe() {
    const {state, actions} = useTranslationFlow();

    return (
        <div>
            <div data-testid="task-id">{state.taskId}</div>
            <div data-testid="active-tab">{state.activeTab}</div>
            <div data-testid="main-translation-run-state">{state.mainTranslationRunState}</div>
            <div data-testid="main-translation-row-count">{state.mainTranslationRows.length}</div>
            <div data-testid="main-translation-selected-category">{state.mainTranslationSelectedCategory}</div>
            <div data-testid="main-translation-selected-row">{state.mainTranslationSelectedRowId}</div>
            <div data-testid="main-translation-dirty-row">{state.mainTranslationDraftState.dirtyDraftRowId}</div>
            <div data-testid="main-translation-next-warning-open">{state.mainTranslationNextWarningOpen ? 'true' : 'false'}</div>
            <div data-testid="main-translation-untranslated-count">{state.mainTranslationSummary.untranslatedCount}</div>
            <button type="button" onClick={() => actions.handleMainTranslationDraftChange(state.mainTranslationSelectedRowId, 'edited')}>
                edit draft
            </button>
            <button type="button" onClick={() => actions.handleMainTranslationConfirmRow(state.mainTranslationSelectedRowId)}>
                confirm row
            </button>
            <button type="button" onClick={() => void actions.handleRunMainTranslation()}>
                run main translation
            </button>
            <button type="button" onClick={() => actions.handleAdvanceFromMainTranslation()}>
                advance from main translation
            </button>
            <button type="button" onClick={() => actions.handleMainTranslationConfirmNext()}>
                confirm next
            </button>
            <button type="button" onClick={() => actions.handleMainTranslationDiscardAndContinue()}>
                discard dirty
            </button>
            <button type="button" onClick={() => actions.handleTabChange(4)}>
                go export tab
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

const buildLoadedFilesResult = (taskId: string) => asLoadResult({
    task_id: taskId,
    files: [{
        file_id: 1,
        file_path: 'Data/Update.esm.extract.json',
        file_name: 'Update.esm.extract.json',
        parse_status: 'ready',
        preview_count: 1,
        preview: {
            file_id: 1,
            page: 1,
            page_size: 50,
            total_rows: 1,
            rows: [{
                id: 'load-row-1',
                section: 'dialogue',
                record_type: 'NPC_:FULL',
                editor_id: 'NPC_A_01',
                source_text: 'NPC Name A-01',
            }],
        },
    }],
});

const buildPersonaPhaseResult = (overrides: Record<string, unknown> = {}): Record<string, unknown> => ({
    task_id: 'existing-task',
    status: 'ready',
    detected_count: 2,
    reused_count: 1,
    pending_count: 1,
    generated_count: 0,
    failed_count: 0,
    progress_mode: 'hidden',
    progress_current: 0,
    progress_total: 0,
    progress_message: '',
    ...overrides,
});

const buildPersonaTargetRow = (overrides: Record<string, unknown> = {}): Record<string, unknown> => ({
    source_plugin: 'Skyrim.esm',
    speaker_id: 'npc-a',
    editor_id: 'NPC_A_01',
    npc_name: 'NPC A',
    race: 'Nord',
    sex: 'female',
    voice_type: 'FemaleNord',
    view_state: 'pending',
    persona_text: '',
    error_message: '',
    dialogues: [{
        record_type: 'DIAL',
        editor_id: 'DIAL_001',
        source_text: 'Hello there',
        quest_id: 'QST_001',
        is_services_branch: false,
        order: 1,
    }],
    ...overrides,
});

const buildPersonaTargetPage = (
    taskId: string,
    rows: Record<string, unknown>[] = [buildPersonaTargetRow()],
    overrides: Record<string, unknown> = {},
): Record<string, unknown> => ({
    task_id: taskId,
    page: 1,
    page_size: 50,
    total_rows: rows.length,
    rows,
    ...overrides,
});

const installDefaultPersonaMocks = (taskId: string): void => {
    installPersonaBinding();
    listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage(taskId, []));
    runPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
        task_id: taskId,
        status: 'completed',
        pending_count: 0,
        generated_count: 1,
    }));
    getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
        task_id: taskId,
        status: 'empty',
        detected_count: 0,
        reused_count: 0,
        pending_count: 0,
    }));
};

describe('useTranslationFlow task resolution', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        installDefaultPersonaMocks('resolved-task');
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
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'empty',
            detected_count: 0,
            reused_count: 0,
            pending_count: 0,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', []));

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
        installDefaultPersonaMocks('existing-task');
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

    it('未開始 phase では IN_PROGRESS イベント単独で running にしない', async () => {
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
                translated_text: '',
                translation_state: 'missing',
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
        await waitFor(() => {
            expect(screen.getByTestId('terminology-summary-status')).toHaveTextContent('pending');
        });
        expect(screen.getByTestId('terminology-running-flag')).toHaveTextContent('false');

        const handler = runtimeEventHandlers.get('translation_flow.terminology.progress');
        expect(handler).toBeDefined();
        await act(async () => {
            handler?.({
                TaskID: 'existing-task',
                Status: 'IN_PROGRESS',
                Current: 1,
                Total: 8,
                Message: '1 / 8 件を処理中',
            });
        });

        await waitFor(() => {
            expect(screen.getByTestId('terminology-summary-status')).toHaveTextContent('pending');
        });
        expect(screen.getByTestId('terminology-running-flag')).toHaveTextContent('false');
        expect(screen.getByTestId('terminology-progress')).toHaveTextContent('0/0');
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
        await waitFor(() => {
            expect(screen.getByTestId('terminology-target-total')).toHaveTextContent('1');
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
        installDefaultPersonaMocks('existing-task');
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
        installDefaultPersonaMocks('existing-task');
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

describe('useTranslationFlow persona phase', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        vi.mocked(ResumeTask).mockResolvedValue(undefined);
        installPersonaBinding();
        vi.mocked(ConfigSet).mockResolvedValue(undefined as never);
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return asConfigMap({selected_provider: 'xai'});
            }
            if (namespace === 'translation_flow.terminology.llm.xai') {
                return asConfigMap({
                    model: 'grok-3-beta',
                    endpoint: 'https://api.x.ai/v1',
                    api_key: 'terminology-api-key',
                    temperature: '0.2',
                    context_length: '12288',
                    sync_concurrency: '3',
                    bulk_strategy: 'batch',
                });
            }
            if (namespace === 'translation_flow.terminology.prompt') {
                return asConfigMap({
                    user_prompt: 'terminology-user-prompt',
                    system_prompt: 'terminology-system-prompt',
                });
            }
            if (namespace === 'translation_flow.persona.llm') {
                return asConfigMap({selected_provider: 'gemini'});
            }
            if (namespace === 'translation_flow.persona.llm.gemini') {
                return asConfigMap({
                    model: 'gemini-2.5-pro',
                    endpoint: '',
                    api_key: 'persona-api-key',
                    temperature: '0.4',
                    context_length: '32768',
                    sync_concurrency: '2',
                    bulk_strategy: 'sync',
                });
            }
            if (namespace === 'translation_flow.persona.prompt') {
                return asConfigMap({
                    user_prompt: 'persona-user-prompt',
                    system_prompt: 'persona-system-prompt',
                });
            }
            return asConfigMap({});
        });
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([
            buildTask({id: 'existing-task', updated_at: '2026-03-18T10:00:00.000Z'}),
        ]));
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'existing-task',
            status: 'completed',
            saved_count: 8,
            failed_count: 0,
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
                editor_id: 'NPC_A_01',
                source_text: 'NPC Name A-01',
                translated_text: 'NPC 名 A-01',
                translation_state: 'translated',
                variant: 'full',
                source_file: 'Update.esm.extract.json',
            }],
        }));
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(buildLoadedFilesResult('existing-task'));
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'ready',
            detected_count: 2,
            reused_count: 1,
            pending_count: 1,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', [
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-a',
                view_state: 'reused',
                persona_text: 'cached persona',
            }),
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-b',
                editor_id: 'NPC_B_01',
                npc_name: 'NPC B',
                view_state: 'pending',
                persona_text: '',
            }),
        ]));
        runPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'completed',
            detected_count: 2,
            reused_count: 1,
            pending_count: 0,
            generated_count: 1,
            failed_count: 0,
        }));
    });

    afterEach(() => {
        cleanup();
    });

    it('REQUEST_GENERATED イベントが重複しても ResumeTask は 1 回だけ呼ぶ', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });

        const taskUpdatedHandler = runtimeEventHandlers.get('task:updated');
        const phaseCompletedHandler = runtimeEventHandlers.get('task:phase_completed');
        if (!taskUpdatedHandler || !phaseCompletedHandler) {
            throw new Error('runtime event handler is not registered');
        }

        act(() => {
            taskUpdatedHandler(buildTask({
                id: 'existing-task',
                status: 'request_generated',
            }));
            phaseCompletedHandler({
                taskId: 'existing-task',
                phase: 'REQUEST_GENERATED',
            });
            phaseCompletedHandler({
                taskId: 'existing-task',
                phase: 'REQUEST_GENERATED',
            });
        });

        await waitFor(() => {
            expect(ResumeTask).toHaveBeenCalledTimes(1);
        });
        expect(ResumeTask).toHaveBeenCalledWith('existing-task');
    });

    it('resume 時に persona summary と target 一覧を再読込し、選択状態を更新できる', async () => {
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'partial_failed',
            detected_count: 2,
            reused_count: 1,
            pending_count: 0,
            generated_count: 0,
            failed_count: 1,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', [
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-a',
                view_state: 'reused',
                persona_text: 'cached persona',
            }),
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-b',
                editor_id: 'NPC_B_01',
                npc_name: 'NPC B',
                view_state: 'failed',
                error_message: 'llm timeout',
            }),
        ]));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });
        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('partialFailed');
        });
        expect(getPersonaPhaseMock).toHaveBeenCalledWith('existing-task');
        expect(listPersonaTargetsMock).toHaveBeenCalledWith('existing-task', 1, 50);
        expect(screen.getByTestId('persona-selected-speaker')).toHaveTextContent('npc-a');

        fireEvent.click(screen.getByRole('button', {name: 'select npc-b'}));

        await waitFor(() => {
            expect(screen.getByTestId('persona-selected-speaker')).toHaveTextContent('npc-b');
        });
    });

    it('persona status が empty のときは実行不可で、次 phase への advance は可能', async () => {
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'empty',
            detected_count: 0,
            reused_count: 0,
            pending_count: 0,
            generated_count: 0,
            failed_count: 0,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', []));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('empty');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run persona'}));

        await waitFor(() => {
            expect(screen.getByTestId('persona-error')).toHaveTextContent('ペルソナ生成対象 NPC がありません。');
        });
        expect(runPersonaPhaseMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', {name: 'advance persona'}));

        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('3');
        });
    });

    it('persona status が cachedOnly のときは実行せず、次 phase への advance は可能', async () => {
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'cached_only',
            detected_count: 1,
            reused_count: 1,
            pending_count: 0,
            generated_count: 0,
            failed_count: 0,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', [
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-a',
                view_state: 'reused',
                persona_text: 'cached persona',
            }),
        ]));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('cachedOnly');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run persona'}));
        expect(runPersonaPhaseMock).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', {name: 'advance persona'}));

        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('3');
        });
    });

    it('persona status が running のときは再実行も次 phase 遷移もできない', async () => {
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'running',
            detected_count: 2,
            reused_count: 1,
            pending_count: 1,
            generated_count: 0,
            failed_count: 0,
            progress_mode: 'determinate',
            progress_current: 1,
            progress_total: 2,
            progress_message: '1 / 2 件（残り 1 件）',
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', [
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-a',
                view_state: 'reused',
                persona_text: 'cached persona',
            }),
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-b',
                view_state: 'running',
            }),
        ]));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('running');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run persona'}));
        fireEvent.click(screen.getByRole('button', {name: 'advance persona'}));

        expect(runPersonaPhaseMock).not.toHaveBeenCalled();
        expect(screen.getByTestId('active-tab')).toHaveTextContent('0');
    });

    it('partialFailed では retry でき、実行時に persona 設定と prompt を利用する', async () => {
        getPersonaPhaseMock.mockResolvedValue(buildPersonaPhaseResult({
            task_id: 'existing-task',
            status: 'partial_failed',
            detected_count: 2,
            reused_count: 1,
            pending_count: 0,
            generated_count: 0,
            failed_count: 1,
        }));
        listPersonaTargetsMock.mockResolvedValue(buildPersonaTargetPage('existing-task', [
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-a',
                view_state: 'reused',
                persona_text: 'cached persona',
            }),
            buildPersonaTargetRow({
                source_plugin: 'Skyrim.esm',
                speaker_id: 'npc-b',
                view_state: 'failed',
                error_message: 'llm timeout',
            }),
        ]));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('partialFailed');
        });
        await waitFor(() => {
            expect(screen.getByTestId('persona-provider')).toHaveTextContent('gemini');
        });

        fireEvent.click(screen.getByRole('button', {name: 'retry persona'}));

        await waitFor(() => {
            expect(runPersonaPhaseMock).toHaveBeenCalled();
        });
        expect(runPersonaPhaseMock).toHaveBeenCalledWith(
            'existing-task',
            expect.objectContaining({
                provider: 'gemini',
                model: 'gemini-2.5-pro',
                endpoint: 'http://localhost:1234',
                api_key: 'persona-api-key',
                temperature: 0.4,
                context_length: 32768,
                sync_concurrency: 2,
                bulk_strategy: 'sync',
            }),
            expect.objectContaining({
                user_prompt: 'persona-user-prompt',
                system_prompt: 'persona-system-prompt',
            }),
        );
    });

    it('persona prompt namespace 未保存時は master persona default で初回 hydrate して persona namespace へ保存する', async () => {
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return asConfigMap({selected_provider: 'xai'});
            }
            if (namespace === 'translation_flow.terminology.llm.xai') {
                return asConfigMap({
                    model: 'grok-3-beta',
                    endpoint: 'https://api.x.ai/v1',
                    api_key: 'shared-api-key',
                    temperature: '0.3',
                    context_length: '8192',
                    sync_concurrency: '4',
                    bulk_strategy: 'batch',
                });
            }
            if (namespace === 'translation_flow.terminology.prompt') {
                return asConfigMap({
                    user_prompt: 'shared-user-prompt',
                    system_prompt: 'shared-system-prompt',
                });
            }
            return asConfigMap({});
        });

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-provider')).toHaveTextContent('xai');
        });

        await waitFor(() => {
            expect(ConfigSet).toHaveBeenCalledWith(
                'translation_flow.persona.llm',
                'selected_provider',
                'xai',
            );
        });
        expect(ConfigSet).toHaveBeenCalledWith(
            'translation_flow.persona.prompt',
            'user_prompt',
            DEFAULT_PERSONA_PROMPT_CONFIG.userPrompt,
        );
        expect(ConfigSet).toHaveBeenCalledWith(
            'translation_flow.persona.prompt',
            'system_prompt',
            DEFAULT_PERSONA_PROMPT_CONFIG.systemPrompt,
        );
    });

    it('tab 遷移は persona 完了前は本文翻訳を拒否し、完了後に許可する', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('ready');
        });

        fireEvent.click(screen.getByRole('button', {name: 'go translation tab'}));
        expect(screen.getByTestId('active-tab')).toHaveTextContent('0');

        fireEvent.click(screen.getByRole('button', {name: 'go persona tab'}));
        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('2');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run persona'}));
        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('completed');
        });

        fireEvent.click(screen.getByRole('button', {name: 'go translation tab'}));
        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('3');
        });
    });

    it('handleAdvanceFromTerminology は persona refresh 完了待ちにせず先に persona tab を開く', async () => {
        let resolvePersonaSummary: ((value: Record<string, unknown>) => void) | null = null;
        getPersonaPhaseMock.mockImplementation(() =>
            new Promise((resolve) => {
                resolvePersonaSummary = resolve;
            }));

        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <PersonaProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('task-id')).toHaveTextContent('existing-task');
        });

        fireEvent.click(screen.getByRole('button', {name: 'advance from terminology'}));

        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('2');
        });
        expect(getPersonaPhaseMock).toHaveBeenCalledWith('existing-task');

        act(() => {
            resolvePersonaSummary?.(buildPersonaPhaseResult({
                task_id: 'existing-task',
                status: 'ready',
            }));
        });

        await waitFor(() => {
            expect(screen.getByTestId('persona-status')).toHaveTextContent('ready');
        });
    });
});

describe('useTranslationFlow main translation workflow', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        runtimeEventHandlers.clear();
        installDefaultPersonaMocks('existing-task');
        vi.mocked(ResumeTask).mockResolvedValue(undefined);
        vi.mocked(ConfigSet).mockResolvedValue(undefined as never);
        vi.mocked(ConfigGetAll).mockImplementation(async (namespace: string) => {
            if (namespace === 'translation_flow.terminology.llm') {
                return asConfigMap({model: 'gemini-2.5-flash', provider: 'gemini'});
            }
            if (namespace === 'translation_flow.translation') {
                return asConfigMap({
                    selected_provider: 'xai',
                    model: 'grok-3-beta',
                    user_prompt: 'translation-user-prompt',
                });
            }
            return asConfigMap({});
        });
        vi.mocked(GetAllTasks).mockResolvedValue(asTaskList([
            buildTask({id: 'existing-task', updated_at: '2026-03-18T10:00:00.000Z'}),
        ]));
        vi.mocked(GetTranslationFlowTerminology).mockResolvedValue(asTerminologyPhaseResult({
            task_id: 'existing-task',
            status: 'completed',
            saved_count: 8,
            failed_count: 0,
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
                editor_id: 'NPC_A_01',
                source_text: 'NPC Name A-01',
                translated_text: 'NPC 名 A-01',
                translation_state: 'translated',
                variant: 'full',
                source_file: 'Update.esm.extract.json',
            }],
        }));
        vi.mocked(ListLoadedTranslationFlowFiles).mockResolvedValue(asLoadResult({
            task_id: 'existing-task',
            files: [{
                file_id: 1,
                file_path: 'Data/Update.esm.extract.json',
                file_name: 'Update.esm.extract.json',
                parse_status: 'ready',
                preview_count: 1,
                preview: {
                    file_id: 1,
                    page: 1,
                    page_size: 50,
                    total_rows: 1,
                    rows: [{
                        id: 'translation-row-1',
                        section: 'dialogue',
                        record_type: 'INFO',
                        editor_id: 'DIALOGUE_001',
                        source_text: 'Hello adventurer.',
                    }],
                },
            }],
        }));
    });

    afterEach(() => {
        cleanup();
    });

    it('main translation の hydrate で ready と選択行を復元する', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <MainTranslationProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('main-translation-run-state')).toHaveTextContent('selectionReady');
        });
        expect(screen.getByTestId('main-translation-row-count')).toHaveTextContent('1');
        expect(screen.getByTestId('main-translation-selected-row')).not.toHaveTextContent(/^$/);
    });

    it('draft がある状態で next は dirty warning を経由し、discard 後に遷移できる', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <MainTranslationProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('main-translation-run-state')).toHaveTextContent('selectionReady');
        });

        fireEvent.click(screen.getByRole('button', {name: 'edit draft'}));
        await waitFor(() => {
            expect(screen.getByTestId('main-translation-dirty-row')).not.toHaveTextContent(/^$/);
        });

        fireEvent.click(screen.getByRole('button', {name: 'advance from main translation'}));
        expect(screen.getByTestId('active-tab')).toHaveTextContent('0');

        fireEvent.click(screen.getByRole('button', {name: 'discard dirty'}));
        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('4');
        });
    });

    it('未翻訳が残る状態で next warning を出し、confirm で export へ進む', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <MainTranslationProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('main-translation-untranslated-count')).toHaveTextContent('1');
        });

        fireEvent.click(screen.getByRole('button', {name: 'advance from main translation'}));
        await waitFor(() => {
            expect(screen.getByTestId('main-translation-next-warning-open')).toHaveTextContent('true');
        });
        expect(screen.getByTestId('active-tab')).toHaveTextContent('0');

        fireEvent.click(screen.getByRole('button', {name: 'confirm next'}));
        await waitFor(() => {
            expect(screen.getByTestId('active-tab')).toHaveTextContent('4');
        });
    });

    it('translating 中は tab change で export へ移動できない', async () => {
        render(
            <MemoryRouter initialEntries={['/translation_flow']}>
                <MainTranslationProbe />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(screen.getByTestId('main-translation-run-state')).toHaveTextContent('selectionReady');
        });

        fireEvent.click(screen.getByRole('button', {name: 'run main translation'}));
        fireEvent.click(screen.getByRole('button', {name: 'go export tab'}));
        expect(screen.getByTestId('active-tab')).toHaveTextContent('0');
    });
});
