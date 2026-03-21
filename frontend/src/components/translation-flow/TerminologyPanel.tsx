import type {ColumnDef} from '@tanstack/react-table';
import DataTable from '../DataTable';
import ModelSettings from '../ModelSettings';
import PromptSettingCard from '../masterPersona/PromptSettingCard';
import type {
    TerminologyPhaseSummary,
    TerminologyTargetPreviewPage,
    TerminologyTargetPreviewRow,
    TerminologyTargetViewState,
} from '../../hooks/features/translationFlow/types';
import type {MasterPersonaLLMConfig, MasterPersonaPromptConfig} from '../../types/masterPersona';

interface TerminologyPanelProps {
    isActive: boolean;
    taskId: string;
    summary: TerminologyPhaseSummary;
    statusLabel: string;
    errorMessage: string;
    targetPage: TerminologyTargetPreviewPage;
    targetStatus: TerminologyTargetViewState;
    targetErrorMessage: string;
    isTargetLoading: boolean;
    isRunning: boolean;
    llmConfig: MasterPersonaLLMConfig;
    promptConfig: MasterPersonaPromptConfig;
    isConfigHydrated: boolean;
    isPromptHydrated: boolean;
    onConfigChange: (next: MasterPersonaLLMConfig) => void;
    onPromptChange: (next: MasterPersonaPromptConfig) => void;
    onRun: () => Promise<void>;
    onRefresh: () => Promise<void>;
    onTargetPageChange: (page: number) => Promise<void>;
    onNext: () => void;
}

const STATUS_BADGE_CLASS: Record<string, string> = {
    completed: 'badge-success badge-outline',
    completed_partial: 'badge-warning badge-outline',
    running: 'badge-warning badge-outline',
    run_error: 'badge-error badge-outline',
    pending: 'badge-ghost',
};

const SummaryCard = ({label, value}: {label: string; value: number}) => (
    <div className="rounded-xl border border-base-200 bg-base-100 p-4 shadow-sm">
        <div className="text-xs font-bold uppercase tracking-[0.16em] text-base-content/50">{label}</div>
        <div className="mt-2 text-3xl font-bold">{value}</div>
    </div>
);

const TARGET_COLUMNS: ColumnDef<TerminologyTargetPreviewRow, unknown>[] = [
    {
        accessorKey: 'recordType',
        header: 'Record Type',
        cell: (info) => <span className="font-mono text-xs">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'editorId',
        header: 'Editor ID',
        cell: (info) => <span className="font-mono text-xs">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'sourceText',
        header: 'Source Text',
    },
    {
        accessorKey: 'translatedText',
        header: 'Translated Text',
        cell: ({row}) => {
            const translationState = row.original.translationState;
            const translatedText = row.original.translatedText;
            if (translationState !== 'translated' || translatedText.trim() === '') {
                return <span className="badge badge-outline badge-warning">未翻訳</span>;
            }
            return <span>{translatedText}</span>;
        },
    },
    {
        accessorKey: 'variant',
        header: 'Variant',
    },
    {
        accessorKey: 'sourceFile',
        header: 'Source File',
    },
];

export function TerminologyPanel({
    isActive,
    taskId,
    summary,
    statusLabel,
    errorMessage,
    targetPage,
    targetStatus,
    targetErrorMessage,
    isTargetLoading,
    isRunning,
    llmConfig,
    promptConfig,
    isConfigHydrated,
    isPromptHydrated,
    onConfigChange,
    onPromptChange,
    onRun,
    onRefresh,
    onTargetPageChange,
    onNext,
}: TerminologyPanelProps) {
    const statusClass = STATUS_BADGE_CLASS[summary.status] ?? STATUS_BADGE_CLASS.pending;
    const canRun = taskId !== '' && llmConfig.model.trim() !== '' && !isRunning && targetStatus === 'ready';
    const canNext = summary.status === 'completed' || summary.status === 'completed_partial';
    const totalPages = Math.max(1, Math.ceil(targetPage.totalRows / Math.max(1, targetPage.pageSize)));
    const showProgress = summary.progressMode !== 'hidden';
    const progressLabel = summary.progressMode === 'determinate' && summary.progressTotal > 0
        ? `${summary.progressCurrent} / ${summary.progressTotal} 件（残り ${Math.max(0, summary.progressTotal - summary.progressCurrent)} 件）`
        : summary.progressMessage;

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>辞書参照を使って単語翻訳を先行実行し、translation flow 専用の terminology 結果を保存します。</span>
            </div>

            <div className="rounded-2xl border border-base-200 bg-base-100 p-5 shadow-sm">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <h2 className="text-lg font-bold">単語翻訳 phase</h2>
                            <span className={`badge ${statusClass}`}>{summary.status || 'pending'}</span>
                        </div>
                        <p className="text-sm text-base-content/70">{statusLabel}</p>
                    </div>
                    <div className="flex flex-wrap gap-2">
                        <button type="button" className="btn btn-outline btn-sm" onClick={() => void onRefresh()} disabled={isRunning}>
                            状態を再読込
                        </button>
                        <div className="flex items-center gap-3">
                            <button type="button" className="btn btn-primary btn-sm" onClick={() => void onRun()} disabled={!canRun}>
                                {isRunning ? '単語翻訳を実行中...' : '単語翻訳を実行'}
                            </button>
                            {showProgress && (
                                <div className="flex min-w-52 flex-col gap-1">
                                    <progress
                                        className="progress progress-primary w-full"
                                        value={summary.progressMode === 'determinate' ? summary.progressCurrent : undefined}
                                        max={summary.progressMode === 'determinate' && summary.progressTotal > 0 ? summary.progressTotal : undefined}
                                    />
                                    {progressLabel !== '' && (
                                        <span className="text-xs text-base-content/70">{progressLabel}</span>
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
                {llmConfig.model.trim() === '' && (
                    <p className="mt-3 text-sm text-warning">実行前にモデルを選択してください。</p>
                )}
                {errorMessage !== '' && <p className="mt-3 text-sm text-error">{errorMessage}</p>}
            </div>

            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <SummaryCard label="保存件数" value={summary.savedCount} />
                <SummaryCard label="失敗件数" value={summary.failedCount} />
            </div>

            <div className="rounded-2xl border border-base-200 bg-base-100 p-5 shadow-sm">
                <div className="mb-4 flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
                    <div>
                        <h3 className="text-base font-bold">対象単語リスト</h3>
                        <p className="text-sm text-base-content/70">Dictionary import と同じ対象集合から抽出した単語です。</p>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                        <span className="badge badge-outline">{targetPage.totalRows} 件</span>
                        <button
                            type="button"
                            className="btn btn-outline btn-xs"
                            onClick={() => void onTargetPageChange(Math.max(1, targetPage.page - 1))}
                            disabled={isRunning || isTargetLoading || targetPage.page <= 1 || targetStatus !== 'ready'}
                        >
                            前へ
                        </button>
                        <span>{targetPage.page} / {totalPages}</span>
                        <button
                            type="button"
                            className="btn btn-outline btn-xs"
                            onClick={() => void onTargetPageChange(Math.min(totalPages, targetPage.page + 1))}
                            disabled={isRunning || isTargetLoading || targetPage.page >= totalPages || targetStatus !== 'ready'}
                        >
                            次へ
                        </button>
                    </div>
                </div>

                {(targetStatus === 'loading' || isRunning) && (
                    <div className="rounded-xl border border-dashed border-base-300 p-4 text-sm text-base-content/60">
                        読込中
                    </div>
                )}
                {targetStatus === 'error' && (
                    <div className="rounded-xl border border-error/30 bg-error/5 p-4 text-sm text-error">
                        {targetErrorMessage || '対象単語リストの取得に失敗しました'}
                    </div>
                )}
                {targetStatus === 'empty' && (
                    <div className="rounded-xl border border-dashed border-base-300 p-4 text-sm text-base-content/60">
                        ロード済みデータに Terminology 対象 REC がありません。
                    </div>
                )}
                {targetStatus === 'ready' && !isRunning && (
                    <DataTable
                        columns={TARGET_COLUMNS}
                        data={targetPage.rows}
                        title={`対象単語リスト (${targetPage.totalRows} 件)`}
                    />
                )}
            </div>

            <ModelSettings
                title="単語翻訳モデル設定"
                value={llmConfig}
                onChange={onConfigChange}
                namespace="translation_flow.terminology"
                enabled={isConfigHydrated}
                locked={isRunning}
                collapsible
                labels={{
                    executionProfile: '翻訳実行方式',
                    syncConcurrency: '単語翻訳の同期並列数',
                }}
            />

            <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
                <PromptSettingCard
                    title="System Prompt"
                    description="terminology slice に渡す system prompt テンプレートです。翻訳ルールや出力形式の制約を調整します。"
                    value={promptConfig.systemPrompt}
                    readOnly
                    badgeLabel="System"
                    footerText="system prompt は terminology slice のテンプレートと一致させるため、単語翻訳フェーズでは閲覧専用です。"
                />
                <PromptSettingCard
                    title="User Prompt"
                    description="各 terminology request に付与する user prompt です。system prompt の補助指示を簡潔に指定します。"
                    value={promptConfig.userPrompt}
                    onChange={!isPromptHydrated || isRunning ? undefined : (value) => onPromptChange({...promptConfig, userPrompt: value})}
                    readOnly={!isPromptHydrated || isRunning}
                    badgeLabel="User"
                    footerText="再訪時も前回保存値をそのまま復元します。"
                />
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-base-content/60 ml-2">
                    Terminology Task: {taskId || '(未選択)'} / {summary.status || 'pending'}
                </span>
                <button type="button" className="btn btn-primary btn-sm" onClick={onNext} disabled={!canNext || isRunning}>
                    単語翻訳を確定して次へ
                </button>
            </div>
        </div>
    );
}
