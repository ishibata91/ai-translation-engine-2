import ModelSettings from '../ModelSettings';
import PromptSettingCard from '../masterPersona/PromptSettingCard';
import {
    DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    type MasterPersonaLLMConfig,
    type MasterPersonaPromptConfig,
} from '../../types/masterPersona';
import type {
    PersonaDialogueView,
    PersonaPhaseSummary,
    PersonaTargetPreviewPage,
    PersonaTargetPreviewRow,
    PersonaTargetViewState,
} from '../../hooks/features/translationFlow/types';

interface PersonaPanelProps {
    isActive: boolean;
    taskId?: string;
    summary?: PersonaPhaseSummary;
    statusLabel?: string;
    errorMessage?: string;
    targetPage?: PersonaTargetPreviewPage;
    targetStatus?: PersonaTargetViewState;
    targetErrorMessage?: string;
    isTargetLoading?: boolean;
    isRunning?: boolean;
    llmConfig?: MasterPersonaLLMConfig;
    promptConfig?: MasterPersonaPromptConfig;
    isConfigHydrated?: boolean;
    isPromptHydrated?: boolean;
    selectedTarget?: PersonaTargetPreviewRow | null;
    onConfigChange?: (next: MasterPersonaLLMConfig) => void;
    onPromptChange?: (next: MasterPersonaPromptConfig) => void;
    onSelectTarget?: (sourcePlugin: string, speakerId: string) => void;
    onRun?: () => Promise<void>;
    onRetry?: () => Promise<void>;
    onRefresh?: () => Promise<void>;
    onTargetPageChange?: (page: number) => Promise<void>;
    onNext: () => void;
}

const EMPTY_SUMMARY: PersonaPhaseSummary = {
    taskId: '',
    status: 'loadingTargets',
    detectedCount: 0,
    reusedCount: 0,
    pendingCount: 0,
    generatedCount: 0,
    failedCount: 0,
    progressMode: 'hidden',
    progressCurrent: 0,
    progressTotal: 0,
    progressMessage: '',
};

const EMPTY_PAGE = (taskId = ''): PersonaTargetPreviewPage => ({
    taskId,
    page: 1,
    pageSize: 50,
    totalRows: 0,
    rows: [],
});

const DEFAULT_PERSONA_PROMPT_CONFIG: MasterPersonaPromptConfig = {
    userPrompt: '',
    systemPrompt: '',
};

const ROW_BADGE: Record<PersonaTargetPreviewRow['viewState'], {label: string; className: string}> = {
    reused: {label: '既存 Master Persona', className: 'badge-info badge-outline'},
    pending: {label: '生成対象', className: 'badge-outline'},
    running: {label: '生成中', className: 'badge-warning badge-outline'},
    generated: {label: '生成済み', className: 'badge-success badge-outline'},
    failed: {label: '生成失敗', className: 'badge-error badge-outline'},
};

const SUMMARY_STATUS_BADGE: Record<PersonaTargetViewState, {label: string; className: string}> = {
    loadingTargets: {label: 'loading', className: 'badge-ghost'},
    empty: {label: 'empty', className: 'badge-ghost'},
    ready: {label: 'ready', className: 'badge-primary badge-outline'},
    cachedOnly: {label: 'cachedOnly', className: 'badge-info badge-outline'},
    running: {label: 'running', className: 'badge-warning badge-outline'},
    completed: {label: 'completed', className: 'badge-success badge-outline'},
    partialFailed: {label: 'partialFailed', className: 'badge-warning badge-outline'},
    failed: {label: 'failed', className: 'badge-error badge-outline'},
};

const SummaryCard = ({label, value}: {label: string; value: number}) => (
    <div className="rounded-xl border border-base-200 bg-base-100 p-4 shadow-sm">
        <div className="text-xs font-bold uppercase tracking-[0.16em] text-base-content/50">{label}</div>
        <div className="mt-2 text-3xl font-bold">{value}</div>
    </div>
);

const DialogueList = ({dialogues}: {dialogues: PersonaDialogueView[]}) => {
    if (dialogues.length === 0) {
        return (
            <div className="rounded-lg border border-dashed border-base-300 p-3 text-sm text-base-content/60">
                会話抜粋はありません。
            </div>
        );
    }

    return (
        <div className="space-y-2">
            {dialogues.map((dialogue, idx) => (
                <div key={`${dialogue.editorId}-${dialogue.order}-${idx}`} className="rounded-lg border border-base-200 p-3">
                    <div className="flex flex-wrap items-center gap-2 text-xs text-base-content/60">
                        <span className="font-mono">{dialogue.recordType || '(recordType未設定)'}</span>
                        <span className="font-mono">{dialogue.editorId || '(editorId未設定)'}</span>
                        {dialogue.questId !== '' && <span>Quest: {dialogue.questId}</span>}
                        {dialogue.isServicesBranch && <span className="badge badge-xs badge-outline">services</span>}
                    </div>
                    <p className="mt-2 text-sm whitespace-pre-wrap">{dialogue.sourceText || '(本文なし)'}</p>
                </div>
            ))}
        </div>
    );
};

const resolveStatusMessage = (status: PersonaTargetViewState): string => {
    switch (status) {
    case 'loadingTargets':
        return 'ペルソナ対象を読込中です。';
    case 'empty':
        return 'ペルソナ対象 NPC はありません。';
    case 'ready':
        return '新規生成対象を確認して実行できます。';
    case 'cachedOnly':
        return '既存 Master Persona を再利用します。';
    case 'running':
        return 'ペルソナ生成を実行中です。';
    case 'completed':
        return 'ペルソナ生成が完了しました。';
    case 'partialFailed':
        return '一部の NPC でペルソナ生成に失敗しました。';
    case 'failed':
        return 'ペルソナ生成に失敗しました。';
    default:
        return '';
    }
};

export function PersonaPanel({
    isActive,
    taskId = '',
    summary = EMPTY_SUMMARY,
    statusLabel = '',
    errorMessage = '',
    targetPage = EMPTY_PAGE(taskId),
    targetStatus,
    targetErrorMessage = '',
    isTargetLoading = false,
    isRunning = false,
    llmConfig = DEFAULT_MASTER_PERSONA_LLM_CONFIG,
    promptConfig = DEFAULT_PERSONA_PROMPT_CONFIG,
    isConfigHydrated = false,
    isPromptHydrated = false,
    selectedTarget = null,
    onConfigChange,
    onPromptChange,
    onSelectTarget,
    onRun,
    onRetry,
    onRefresh,
    onTargetPageChange,
    onNext,
}: PersonaPanelProps) {
    const effectiveStatus = targetStatus ?? summary.status;
    const statusBadge = SUMMARY_STATUS_BADGE[effectiveStatus];
    const listRows = targetPage.rows;
    const selectedRow = selectedTarget ?? listRows[0] ?? null;
    const totalPages = Math.max(1, Math.ceil(targetPage.totalRows / Math.max(1, targetPage.pageSize)));
    const canRun = effectiveStatus === 'ready' && !isRunning && !isTargetLoading && typeof onRun === 'function';
    const canRetry = (effectiveStatus === 'partialFailed' || effectiveStatus === 'failed')
        && !isRunning
        && typeof onRetry === 'function';
    const canNext = !isRunning && (
        effectiveStatus === 'empty'
        || effectiveStatus === 'cachedOnly'
        || effectiveStatus === 'completed'
        || effectiveStatus === 'partialFailed'
    );
    const settingsLocked = isRunning;

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full overflow-y-auto ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>既存 Master Persona は再利用し、新規生成対象のみを実行します。</span>
            </div>

            <div className="rounded-2xl border border-base-200 bg-base-100 p-5 shadow-sm">
                <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <h2 className="text-lg font-bold">ペルソナ生成 phase</h2>
                            <span className={`badge ${statusBadge.className}`}>{statusBadge.label}</span>
                        </div>
                        <p className="text-sm text-base-content/70">{statusLabel || resolveStatusMessage(effectiveStatus)}</p>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            type="button"
                            className="btn btn-outline btn-sm"
                            onClick={() => void onRefresh?.()}
                            disabled={isRunning || typeof onRefresh !== 'function'}
                        >
                            状態を再読込
                        </button>
                        <button
                            type="button"
                            className="btn btn-primary btn-sm"
                            onClick={() => void onRun?.()}
                            disabled={!canRun}
                        >
                            ペルソナ生成を開始
                        </button>
                        <button
                            type="button"
                            className="btn btn-warning btn-sm"
                            onClick={() => void onRetry?.()}
                            disabled={!canRetry}
                        >
                            再試行
                        </button>
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
                    <SummaryCard label="検出 NPC 数" value={summary.detectedCount} />
                    <SummaryCard label="再利用数" value={summary.reusedCount} />
                    <SummaryCard label="新規生成数" value={summary.pendingCount} />
                    <SummaryCard label="失敗数" value={summary.failedCount} />
                </div>

                {summary.progressMode !== 'hidden' && (
                    <div className="mt-4 rounded-xl border border-base-200 bg-base-100 p-3">
                        <progress
                            className="progress progress-primary w-full"
                            value={summary.progressMode === 'determinate' ? summary.progressCurrent : undefined}
                            max={summary.progressMode === 'determinate' && summary.progressTotal > 0 ? summary.progressTotal : undefined}
                        />
                        {(summary.progressMessage || summary.progressTotal > 0) && (
                            <p className="mt-2 text-xs text-base-content/70">
                                {summary.progressMessage || `${summary.progressCurrent} / ${summary.progressTotal}`}
                            </p>
                        )}
                    </div>
                )}

                {(effectiveStatus === 'partialFailed' || effectiveStatus === 'failed' || errorMessage !== '') && (
                    <div className="mt-4 rounded-xl border border-error/30 bg-error/5 p-4 text-sm text-error">
                        {errorMessage || targetErrorMessage || resolveStatusMessage(effectiveStatus)}
                        {effectiveStatus === 'partialFailed' && (
                            <div className="mt-2 text-xs">失敗行は persona なしのまま次 phase に進むことができます。</div>
                        )}
                    </div>
                )}
            </div>

            <div className="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
                <ModelSettings
                    title="ペルソナ生成モデル設定"
                    value={llmConfig}
                    onChange={onConfigChange ?? (() => undefined)}
                    enabled={isConfigHydrated}
                    namespace="translation_flow.persona"
                    locked={settingsLocked}
                    collapsible={false}
                    labels={{
                        executionProfile: '生成方式',
                    }}
                />

                <div className="grid grid-cols-1 gap-4">
                    <PromptSettingCard
                        title="System Prompt"
                        description="ペルソナ生成 phase で共通に使う system prompt です。"
                        value={promptConfig.systemPrompt}
                        onChange={onPromptChange ? (value) => onPromptChange({...promptConfig, systemPrompt: value}) : undefined}
                        readOnly={!isPromptHydrated || settingsLocked}
                        badgeLabel={!isPromptHydrated ? '読込中' : settingsLocked ? '固定' : '編集可'}
                        footerText="変更内容は persona phase 専用 prompt として自動保存されます。"
                    />
                    <PromptSettingCard
                        title="User Prompt"
                        description="NPC ごとの会話抜粋へ差し込む user prompt です。"
                        value={promptConfig.userPrompt}
                        onChange={onPromptChange ? (value) => onPromptChange({...promptConfig, userPrompt: value}) : undefined}
                        readOnly={!isPromptHydrated || settingsLocked}
                        badgeLabel={!isPromptHydrated ? '読込中' : settingsLocked ? '固定' : '編集可'}
                        footerText="terminology phase とは別 namespace に保存され、再表示時にも復元されます。"
                    />
                </div>
            </div>

            <div className="flex gap-4 flex-1 min-h-0 overflow-hidden relative">
                <div className="w-2/5 border rounded-xl bg-base-100 flex flex-col min-h-0 overflow-hidden">
                    <div className="flex items-center justify-between border-b border-base-200 px-3 py-2">
                        <span className="text-sm font-bold">NPC 一覧 ({targetPage.totalRows} 件)</span>
                        <div className="flex items-center gap-2 text-xs">
                            <button
                                type="button"
                                className="btn btn-outline btn-xs"
                                onClick={() => void onTargetPageChange?.(Math.max(1, targetPage.page - 1))}
                                disabled={targetPage.page <= 1 || isRunning || isTargetLoading || typeof onTargetPageChange !== 'function'}
                            >
                                前へ
                            </button>
                            <span>{targetPage.page} / {totalPages}</span>
                            <button
                                type="button"
                                className="btn btn-outline btn-xs"
                                onClick={() => void onTargetPageChange?.(Math.min(totalPages, targetPage.page + 1))}
                                disabled={targetPage.page >= totalPages || isRunning || isTargetLoading || typeof onTargetPageChange !== 'function'}
                            >
                                次へ
                            </button>
                        </div>
                    </div>

                    {(effectiveStatus === 'loadingTargets' || isTargetLoading) && (
                        <div className="p-4 text-sm text-base-content/60">読込中</div>
                    )}
                    {effectiveStatus === 'empty' && (
                        <div className="p-4 text-sm text-base-content/60">ペルソナ対象 NPC はありません。</div>
                    )}
                    {listRows.length > 0 && (
                        <ul className="menu w-full bg-base-100 flex-1 overflow-y-auto p-2 gap-1">
                            {listRows.map((row) => {
                                const rowKey = `${row.sourcePlugin}::${row.speakerId}`;
                                const isSelected = selectedRow !== null && rowKey === `${selectedRow.sourcePlugin}::${selectedRow.speakerId}`;
                                const badge = ROW_BADGE[row.viewState];
                                return (
                                    <li key={rowKey}>
                                        <button
                                            type="button"
                                            className={`${isSelected ? 'active' : ''} flex-col items-start gap-1`}
                                            onClick={() => onSelectTarget?.(row.sourcePlugin, row.speakerId)}
                                        >
                                            <div className="flex w-full items-center justify-between gap-2">
                                                <span className="font-semibold truncate">{row.npcName || '(名称なし)'}</span>
                                                <span className={`badge badge-xs ${badge.className}`}>{badge.label}</span>
                                            </div>
                                            <div className="w-full text-left text-xs opacity-70 font-mono truncate">
                                                {row.sourcePlugin || '(source plugin 未設定)'} / {row.speakerId || '(speaker未設定)'}
                                            </div>
                                        </button>
                                    </li>
                                );
                            })}
                        </ul>
                    )}
                </div>

                <div className="w-3/5 flex flex-col min-h-0 rounded-xl border bg-base-100">
                    <div className="border-b border-base-200 px-4 py-3">
                        <h3 className="font-bold">詳細</h3>
                    </div>

                    {selectedRow === null && (
                        <div className="p-4 text-sm text-base-content/60">NPC を選択してください。</div>
                    )}

                    {selectedRow !== null && (
                        <div className="p-4 flex flex-col gap-4 overflow-y-auto">
                            <div className="space-y-1">
                                <div className="text-base font-bold">{selectedRow.npcName || '(名称なし)'}</div>
                                <div className="text-xs text-base-content/70 font-mono">
                                    {selectedRow.sourcePlugin || '(source plugin 未設定)'} / {selectedRow.speakerId || '(speaker未設定)'}
                                </div>
                                <div className="text-xs text-base-content/70">
                                    editor: {selectedRow.editorId || '(未設定)'} / race: {selectedRow.race || '(未設定)'} / sex: {selectedRow.sex || '(未設定)'} / voice: {selectedRow.voiceType || '(未設定)'}
                                </div>
                            </div>

                            {(selectedRow.viewState === 'reused' || selectedRow.viewState === 'generated') ? (
                                <div className="rounded-xl border border-base-200 bg-base-100 p-4">
                                    <div className="mb-2 text-sm font-bold">Persona 本文</div>
                                    <pre className="whitespace-pre-wrap text-sm leading-relaxed">{selectedRow.personaText || '(本文なし)'}</pre>
                                </div>
                            ) : (
                                <div className="space-y-3">
                                    <div className="rounded-xl border border-dashed border-base-300 p-3 text-sm text-base-content/70">
                                        {selectedRow.viewState === 'failed'
                                            ? (selectedRow.errorMessage || 'この NPC のペルソナ生成は失敗しました。再試行してください。')
                                            : 'この NPC はまだ生成されていません。'}
                                    </div>
                                    <div>
                                        <div className="mb-2 text-sm font-bold">会話抜粋</div>
                                        <DialogueList dialogues={selectedRow.dialogues} />
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0 mt-auto">
                <span className="text-sm font-bold text-base-content/60 ml-2">
                    Persona Task: {taskId || '(未選択)'} / {effectiveStatus}
                </span>
                <button type="button" className="btn btn-primary btn-sm" onClick={onNext} disabled={!canNext}>
                    次へ
                </button>
            </div>
        </div>
    );
}
