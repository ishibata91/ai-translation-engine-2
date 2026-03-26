import {useMemo, useState} from 'react';
import ModelSettings from '../ModelSettings';
import PromptSettingCard from '../masterPersona/PromptSettingCard';
import {mapMainTranslationRowDetail} from '../../hooks/features/translationFlow/adapters';
import type {
    MainTranslationCategory,
    MainTranslationDraftState,
    MainTranslationRowViewModel,
    MainTranslationRunState,
    MainTranslationSummary,
    TerminologyTargetPreviewPage,
} from '../../hooks/features/translationFlow/types';
import type {MasterPersonaLLMConfig} from '../../types/masterPersona';

interface TranslationPanelProps {
    isActive: boolean;
    taskId: string;
    runState: MainTranslationRunState;
    rows: MainTranslationRowViewModel[];
    selectedCategory: MainTranslationCategory;
    selectedRowId: string;
    draftState: MainTranslationDraftState;
    summary: MainTranslationSummary;
    config: MasterPersonaLLMConfig;
    userPrompt: string;
    errorMessage: string;
    dirtyWarningOpen: boolean;
    nextWarningOpen: boolean;
    nextWarningCount: number;
    isHydrated: boolean;
    terminologyPage: TerminologyTargetPreviewPage;
    onCategoryChange: (next: MainTranslationCategory) => void;
    onSelectRow: (rowId: string) => void;
    onDraftChange: (rowId: string, value: string) => void;
    onConfirmRow: (rowId: string) => void;
    onCancelConfirmed: (rowId: string) => void;
    onRun: () => Promise<void>;
    onRetryFailedOnly: () => Promise<void>;
    onNext: () => void;
    onDiscardAndContinue: () => void;
    onKeepEditing: () => void;
    onConfirmNext: () => void;
    onCancelNext: () => void;
    onConfigChange: (next: MasterPersonaLLMConfig) => void;
    onUserPromptChange: (next: string) => void;
}

const CATEGORY_LABELS: Record<MainTranslationCategory, string> = {
    conversation: '会話',
    quest: 'クエスト',
    other: 'その他',
};

const STATUS_BADGE_CLASS: Record<string, string> = {
    untranslated: 'badge-ghost',
    aiTranslated: 'badge-info badge-outline',
    confirmed: 'badge-success badge-outline',
};

const STATUS_LABELS: Record<string, string> = {
    untranslated: '未翻訳',
    aiTranslated: 'AI翻訳済み',
    confirmed: '確定',
};

const RUN_STATE_LABELS: Record<MainTranslationRunState, string> = {
    hydrating: '読込中',
    loadError: '読み込み失敗',
    empty: '対象なし',
    ready: '準備完了',
    selectionEmpty: '対象未選択',
    selectionReady: '編集可能',
    translating: '翻訳中',
    translateCompleted: '翻訳完了',
    translatePartialFailed: '一部失敗',
    translateFailed: '失敗',
};

const SummaryCard = ({label, value}: {label: string; value: number}) => (
    <div className="rounded-xl border border-base-200 bg-base-100 p-3 shadow-sm">
        <div className="text-[11px] font-bold uppercase tracking-[0.14em] text-base-content/50">{label}</div>
        <div className="mt-1 text-2xl font-bold">{value}</div>
    </div>
);

const isRowActionLocked = (runState: MainTranslationRunState): boolean => runState === 'translating' || runState === 'hydrating';

export function TranslationPanel({
    isActive,
    taskId,
    runState,
    rows,
    selectedCategory,
    selectedRowId,
    draftState,
    summary,
    config,
    userPrompt,
    errorMessage,
    dirtyWarningOpen,
    nextWarningOpen,
    nextWarningCount,
    isHydrated,
    terminologyPage,
    onCategoryChange,
    onSelectRow,
    onDraftChange,
    onConfirmRow,
    onCancelConfirmed,
    onRun,
    onRetryFailedOnly,
    onNext,
    onDiscardAndContinue,
    onKeepEditing,
    onConfirmNext,
    onCancelNext,
    onConfigChange,
    onUserPromptChange,
}: TranslationPanelProps) {
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState<'all' | 'untranslated' | 'aiTranslated' | 'confirmed'>('all');

    const translatingLocked = runState === 'translating';
    const runBlocked = runState === 'hydrating' || runState === 'loadError' || runState === 'translateFailed';
    const nextDisabled = runState === 'translating' || runState === 'hydrating' || runState === 'translateFailed';

    const categoryRows = useMemo(
        () => rows.filter((row) => row.category === selectedCategory),
        [rows, selectedCategory],
    );

    const visibleRows = useMemo(() => {
        return categoryRows.filter((row) => {
            if (statusFilter !== 'all' && row.status !== statusFilter) {
                return false;
            }
            if (searchTerm.trim() === '') {
                return true;
            }
            const haystack = `${row.primaryLabel} ${row.sourceText} ${row.secondaryMeta.join(' ')}`.toLowerCase();
            return haystack.includes(searchTerm.trim().toLowerCase());
        });
    }, [categoryRows, searchTerm, statusFilter]);

    const selectedRow = useMemo(
        () => categoryRows.find((row) => row.rowId === selectedRowId) ?? null,
        [categoryRows, selectedRowId],
    );

    const selectedRowEffectiveText = selectedRow
        ? (draftState.draftMap[selectedRow.rowId] ?? selectedRow.translatedText)
        : '';
    const rowDetail = selectedRow
        ? mapMainTranslationRowDetail(selectedRow, terminologyPage.rows)
        : null;
    const canConfirmRow = selectedRow !== null
        && !translatingLocked
        && (selectedRow.status === 'aiTranslated' || draftState.dirtyDraftRowId === selectedRow.rowId);
    const canCancelConfirmed = selectedRow !== null && !translatingLocked && selectedRow.status === 'confirmed';

    return (
        <div className={`tab-content-panel flex-col gap-4 h-full min-h-0 overflow-hidden ${isActive ? 'flex' : 'hidden'}`}>
            <div className="rounded-2xl border border-base-200 bg-base-100 p-5 shadow-sm">
                <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div className="space-y-2">
                        <div className="flex items-center gap-2">
                            <h2 className="text-lg font-bold">本文翻訳 phase</h2>
                            <span className="badge badge-outline">{RUN_STATE_LABELS[runState]}</span>
                        </div>
                        <p className="text-sm text-base-content/70">保存済み設定を復元し、対象を選択して本文翻訳を進めます。</p>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            type="button"
                            className="btn btn-primary btn-sm"
                            onClick={() => void onRun()}
                            disabled={runBlocked || rows.length === 0}
                        >
                            翻訳開始
                        </button>
                        <button
                            type="button"
                            className="btn btn-warning btn-sm"
                            onClick={() => void onRetryFailedOnly()}
                            disabled={translatingLocked || summary.untranslatedCount === 0}
                        >
                            failed のみ再実行
                        </button>
                    </div>
                </div>
                <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
                    <SummaryCard label="未翻訳" value={summary.untranslatedCount} />
                    <SummaryCard label="AI翻訳済み" value={summary.aiTranslatedCount} />
                    <SummaryCard label="確定" value={summary.confirmedCount} />
                    <SummaryCard label="失敗" value={summary.failedCount} />
                </div>
                {(runState === 'loadError' || runState === 'translateFailed' || errorMessage !== '') && (
                    <div className="mt-4 rounded-xl border border-error/30 bg-error/5 p-3 text-sm text-error">
                        {errorMessage || '本文翻訳の状態取得に失敗しました。再読込してください。'}
                    </div>
                )}
                {runState === 'translatePartialFailed' && (
                    <div className="mt-4 rounded-xl border border-warning/30 bg-warning/10 p-3 text-sm text-warning-content">
                        失敗行は未翻訳のまま残ります。`failed のみ再実行` を利用できます。
                    </div>
                )}
            </div>

            <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
                <ModelSettings
                    title="翻訳モデル設定"
                    value={config}
                    onChange={onConfigChange}
                    enabled={isHydrated}
                    namespace="translation_flow.translation"
                    locked={translatingLocked}
                    collapsible={false}
                    labels={{
                        executionProfile: '翻訳実行方式',
                        syncConcurrency: '本文翻訳の同期並列数',
                    }}
                />
                <div className="grid grid-cols-1 gap-4">
                    <PromptSettingCard
                        title="System Prompt"
                        description="選択カテゴリと recordType から導出される system prompt です。"
                        value={rowDetail?.systemPrompt ?? '対象を選択すると表示されます。'}
                        readOnly
                        badgeLabel="readOnly"
                        footerText="system prompt は frontend 同梱テンプレートから供給し、保存対象に含めません。"
                    />
                    <PromptSettingCard
                        title="User Prompt"
                        description="本文翻訳 phase 専用の user prompt です。"
                        value={userPrompt}
                        onChange={onUserPromptChange}
                        readOnly={!isHydrated || translatingLocked}
                        badgeLabel={!isHydrated ? '読込中' : translatingLocked ? '固定' : '編集可'}
                        footerText="変更内容は `translation_flow.translation` に保存され、再表示時に復元されます。"
                    />
                </div>
            </div>

            <div className="flex flex-1 min-h-0 gap-4 overflow-hidden">
                <div className="flex w-[45%] min-w-0 flex-col overflow-hidden rounded-xl border border-base-200 bg-base-100">
                    <div className="border-b border-base-200 p-3">
                        <div className="mb-2 flex gap-2">
                            {(Object.keys(CATEGORY_LABELS) as MainTranslationCategory[]).map((category) => (
                                <button
                                    key={category}
                                    type="button"
                                    className={`btn btn-xs ${selectedCategory === category ? 'btn-primary' : 'btn-outline'}`}
                                    onClick={() => onCategoryChange(category)}
                                    disabled={translatingLocked}
                                >
                                    {CATEGORY_LABELS[category]}
                                </button>
                            ))}
                        </div>
                        <div className="flex gap-2">
                            <input
                                type="text"
                                className="input input-sm input-bordered w-full"
                                placeholder="原文・訳文を検索"
                                value={searchTerm}
                                onChange={(event) => setSearchTerm(event.target.value)}
                                disabled={translatingLocked}
                            />
                            <select
                                className="select select-sm select-bordered"
                                value={statusFilter}
                                onChange={(event) => setStatusFilter(event.target.value as typeof statusFilter)}
                                disabled={translatingLocked}
                            >
                                <option value="all">すべて</option>
                                <option value="untranslated">未翻訳</option>
                                <option value="aiTranslated">AI翻訳済み</option>
                                <option value="confirmed">確定</option>
                            </select>
                        </div>
                    </div>
                    <div className="flex-1 overflow-y-auto">
                        {visibleRows.length === 0 ? (
                            <div className="p-4 text-sm text-base-content/60">
                                {rows.length === 0
                                    ? '表示できる翻訳対象はありません。'
                                    : '条件に一致する本文がありません。'}
                            </div>
                        ) : (
                            <ul className="divide-y divide-base-200">
                                {visibleRows.map((row) => (
                                    <li key={row.rowId}>
                                        <button
                                            type="button"
                                            className={`w-full px-3 py-3 text-left hover:bg-base-200/60 ${selectedRowId === row.rowId ? 'bg-base-200' : ''}`}
                                            onClick={() => onSelectRow(row.rowId)}
                                            disabled={translatingLocked}
                                        >
                                            <div className="flex items-start justify-between gap-2">
                                                <div className="line-clamp-2 text-sm font-semibold">{row.primaryLabel || '(本文なし)'}</div>
                                                <span className={`badge badge-xs ${STATUS_BADGE_CLASS[row.status] ?? 'badge-ghost'}`}>
                                                    {STATUS_LABELS[row.status] ?? row.status}
                                                </span>
                                            </div>
                                            <div className="mt-1 text-xs text-base-content/60">{row.secondaryMeta.join(' / ')}</div>
                                        </button>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                </div>

                <div className="flex w-[55%] min-w-0 flex-col overflow-hidden rounded-xl border border-base-200 bg-base-100">
                    <div className="border-b border-base-200 p-3">
                        <div className="flex items-center justify-between gap-2">
                            <div>
                                <div className="text-sm font-bold">{selectedRow?.primaryLabel || '対象を選択してください'}</div>
                                {selectedRow && (
                                    <div className="text-xs text-base-content/60">
                                        {selectedRow.metadata.recordType || '-'} / {selectedRow.metadata.editorId || '-'}
                                    </div>
                                )}
                            </div>
                            {selectedRow && (
                                <div className="flex items-center gap-2">
                                    <span className={`badge ${STATUS_BADGE_CLASS[selectedRow.status] ?? 'badge-ghost'}`}>
                                        {STATUS_LABELS[selectedRow.status] ?? selectedRow.status}
                                    </span>
                                    <button
                                        type="button"
                                        className="btn btn-outline btn-xs"
                                        onClick={() => void onRun()}
                                        disabled={isRowActionLocked(runState)}
                                    >
                                        AI訳を再生成
                                    </button>
                                    <button
                                        type="button"
                                        className="btn btn-primary btn-xs"
                                        onClick={() => onConfirmRow(selectedRow.rowId)}
                                        disabled={!canConfirmRow}
                                    >
                                        確定
                                    </button>
                                    <button
                                        type="button"
                                        className="btn btn-ghost btn-xs"
                                        onClick={() => onCancelConfirmed(selectedRow.rowId)}
                                        disabled={!canCancelConfirmed}
                                    >
                                        取り消し
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>

                    {selectedRow === null ? (
                        <div className="p-4 text-sm text-base-content/60">
                            {rows.length === 0
                                ? '本文翻訳の対象がありません。次 phase へ進めます。'
                                : '対象を選択してください。'}
                        </div>
                    ) : (
                        <div className="flex-1 space-y-4 overflow-y-auto p-4">
                            <div>
                                <div className="mb-1 text-xs font-bold text-base-content/60">原文</div>
                                <div className="rounded-lg border border-base-200 bg-base-100 p-3 text-sm">
                                    {selectedRow.sourceText || '(原文なし)'}
                                </div>
                            </div>
                            <div>
                                <div className="mb-1 text-xs font-bold text-base-content/60">訳文</div>
                                <textarea
                                    className="textarea textarea-bordered w-full min-h-40"
                                    value={selectedRowEffectiveText}
                                    onChange={(event) => onDraftChange(selectedRow.rowId, event.target.value)}
                                    readOnly={translatingLocked}
                                />
                            </div>
                            <div className="rounded-lg border border-base-200 p-3">
                                <div className="mb-2 text-xs font-bold text-base-content/60">メタデータ</div>
                                <div className="text-xs leading-6 text-base-content/80">
                                    <div>recordType: {selectedRow.metadata.recordType || '-'}</div>
                                    <div>editorId: {selectedRow.metadata.editorId || '-'}</div>
                                    <div>sourcePlugin: {selectedRow.metadata.sourcePlugin || '-'}</div>
                                    {selectedRow.category === 'conversation' && (
                                        <div>speakerId / NPC名: {selectedRow.metadata.speakerId || '-'} / {selectedRow.metadata.npcName || '-'}</div>
                                    )}
                                    {selectedRow.category === 'quest' && (
                                        <div>questId / stage / objective: {selectedRow.metadata.questId || '-'} / {selectedRow.metadata.stageIndex ?? '-'} / {selectedRow.metadata.objective || '-'}</div>
                                    )}
                                </div>
                            </div>
                            <div className="grid grid-cols-1 gap-2">
                                {(rowDetail?.referencePanels ?? []).map((panel) => (
                                    <div key={panel.title} className="rounded-lg border border-base-200 p-3">
                                        <div className="mb-1 text-xs font-bold text-base-content/60">{panel.title}</div>
                                        <ul className="space-y-1 text-xs text-base-content/80">
                                            {panel.items.map((item, index) => (
                                                <li key={`${panel.title}-${index}`}>- {item}</li>
                                            ))}
                                        </ul>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>

            <div className="flex items-center justify-between rounded-xl border bg-base-200 p-2">
                <span className="ml-2 text-sm font-bold text-base-content/60">
                    Translation Task: {taskId || '(未選択)'} / {runState}
                </span>
                <button type="button" className="btn btn-primary btn-sm" onClick={onNext} disabled={nextDisabled}>
                    次へ
                </button>
            </div>

            {dirtyWarningOpen && (
                <dialog className="modal modal-open">
                    <div className="modal-box">
                        <h3 className="text-lg font-bold">未確定の変更があります</h3>
                        <p className="py-3 text-sm">この本文には未確定の変更があります。破棄して移動しますか。</p>
                        <div className="modal-action">
                            <button type="button" className="btn btn-ghost" onClick={onKeepEditing}>編集を続ける</button>
                            <button type="button" className="btn btn-primary" onClick={onDiscardAndContinue}>破棄して移動</button>
                        </div>
                    </div>
                </dialog>
            )}

            {nextWarningOpen && (
                <dialog className="modal modal-open">
                    <div className="modal-box">
                        <h3 className="text-lg font-bold">未翻訳が残っています</h3>
                        <p className="py-3 text-sm">
                            未翻訳の本文が {nextWarningCount} 件残っています。このまま次へ進みますか。
                        </p>
                        <div className="modal-action">
                            <button type="button" className="btn btn-ghost" onClick={onCancelNext}>キャンセル</button>
                            <button type="button" className="btn btn-primary" onClick={onConfirmNext}>このまま進む</button>
                        </div>
                    </div>
                </dialog>
            )}
        </div>
    );
}
