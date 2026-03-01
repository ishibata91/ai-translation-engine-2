import React, { useState, useRef, useEffect, useCallback } from 'react';

// ── 列定義 ──────────────────────────────────────────────
export interface GridColumnDef<TData> {
    /** TData のキー */
    key: keyof TData & string;
    /** ヘッダーラベル */
    header: string;
    /** 編集可能か (デフォルト: false = 読み取り専用) */
    editable?: boolean;
    /** セル幅の Tailwind クラス (e.g. 'w-24', 'w-64') */
    widthClass?: string;
    /** 入力タイプ */
    type?: 'text' | 'number';
}

// ── 行メタ情報（変更追跡用） ─────────────────────────────
type RowMeta = 'original' | 'modified' | 'deleted';

interface RowWithMeta<TData> {
    data: TData;
    meta: RowMeta;
    /** React key 用ローカルID (DB の id と独立) */
    localId: number;
}

// ── Props ───────────────────────────────────────────────
interface GridEditorProps<TData extends object> {
    /** ヘッダーのタイトル */
    title: string;
    /** 初期データ (変更時に draft をリセット) */
    initialData: TData[];
    /** 列定義 (データ型に依存しない外部注入) */
    columns: GridColumnDef<TData>[];
    /** 「← 戻る」コールバック (省略時はボタン非表示) */
    onBack?: () => void;
    /** 「保存」コールバック */
    onSave: (modified: TData[], deleted: TData[]) => void;
    /** 「検索」実行時のコールバック (サーバーサイド検索用)。各列ごとのフィルタ状態を渡す。 */
    onSearch?: (filters: Record<string, string>) => void;
    /** 変更状態が変わったときのコールバック */
    onDirtyChange?: (isDirty: boolean) => void;
    // ── ページネーション (省略時はローカルフィルタのみ) ──
    /** 現在のページ番号 (1始まり) */
    currentPage?: number;
    /** 全体件数 */
    totalCount?: number;
    /** 1ページの表示件数 */
    pageSize?: number;
    /** ページ切り替えコールバック */
    onPageChange?: (page: number) => void;
}

// ── ローカルID採番 ───────────────────────────────────────
let _localId = 0;
const nextLocalId = () => ++_localId;

// ── コンポーネント ──────────────────────────────────────
function GridEditor<TData extends object>({
    title,
    initialData,
    columns,
    onBack,
    onSave,
    onDirtyChange,
    currentPage,
    totalCount,
    pageSize = 500,
    onPageChange,
    onSearch,
}: GridEditorProps<TData>) {
    // draft: 編集中の行リスト
    const initRows = (): RowWithMeta<TData>[] =>
        initialData.map((d) => ({ data: d, meta: 'original', localId: nextLocalId() }));

    const [rows, setRows] = useState<RowWithMeta<TData>[]>(initRows);

    // 各列のフィルタ文字列の入力中状態 { [columnKey]: draftString }
    const [filterDraft, setFilterDraft] = useState<Record<string, string>>({});
    // 実際に適用済みのフィルタ
    const [appliedFilters, setAppliedFilters] = useState<Record<string, string>>({});
    // 確認モーダル（戻る時の未保存警告）の表示フラグ
    const [showBackModal, setShowBackModal] = useState(false);

    // フィルタ「検索」ボタン押下
    const handleApplyFilters = () => {
        setAppliedFilters({ ...filterDraft });
        setEditingCell(null);

        // onSearch が提供されている場合は、親(サーバー側)で検索を行う
        if (onSearch) {
            onSearch(filterDraft);
        }
    };

    // フィルタ「クリア」ボタン押下
    const handleClearFilters = () => {
        setFilterDraft({});
        setAppliedFilters({});
        setEditingCell(null);
        if (onSearch) {
            onSearch({});
        }
    };

    // フィルタ適用後の行（表示用、削除予定行は薄く表示するため除外しない）
    const filteredRows = rows.filter((row) =>
        columns.every((col) => {
            const filter = (appliedFilters[col.key] ?? '').trim().toLowerCase();
            if (!filter) return true;
            const val = String(row.data[col.key as keyof TData] ?? '').toLowerCase();
            return val.includes(filter);
        })
    );

    const isFiltered = Object.values(appliedFilters).some((v) => v.trim() !== '');

    // initialData が変わったら draft をリセット
    useEffect(() => {
        setRows(initRows());
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [initialData]);

    // 現在編集中のセル
    const [editingCell, setEditingCell] = useState<{ localId: number; key: string } | null>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    // editingCell が変わるたびに input にフォーカス
    useEffect(() => {
        if (editingCell && inputRef.current) {
            inputRef.current.focus();
            inputRef.current.select();
        }
    }, [editingCell]);

    const isDirty = rows.some((r) => r.meta !== 'original');
    const deletedCount = rows.filter((r) => r.meta === 'deleted').length;

    useEffect(() => {
        onDirtyChange?.(isDirty);
    }, [isDirty, onDirtyChange]);

    // 「戻る」ボタン押下: 未保存があれば警告
    const handleBackClick = useCallback(() => {
        if (isDirty) {
            setShowBackModal(true);
        } else {
            onBack?.();
        }
    }, [isDirty, onBack]);

    // ── セル操作 ────────────────────────────────────────
    const handleCellClick = (localId: number, key: string, editable?: boolean) => {
        if (!editable) return;
        setEditingCell({ localId, key });
    };

    const handleCellChange = (localId: number, key: string, value: string | number) => {
        setRows((prev) =>
            prev.map((r) =>
                r.localId === localId
                    ? {
                        ...r,
                        data: { ...r.data, [key]: value },
                        meta: r.meta === 'deleted' ? 'deleted' : 'modified',
                    }
                    : r
            )
        );
    };

    const handleCellBlur = () => setEditingCell(null);

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' || e.key === 'Tab' || e.key === 'Escape') {
            e.preventDefault();
            handleCellBlur();
        }
    };

    // ── 行操作 ──────────────────────────────────────────
    // 削除: 即座に消すのではなく「削除予定」状態にマーク
    const handleMarkDeleteRow = (localId: number) => {
        setRows((prev) =>
            prev.map((r) =>
                r.localId === localId ? { ...r, meta: 'deleted' } : r
            )
        );
    };

    // 削除予定を取り消す
    const handleUnmarkDeleteRow = (localId: number) => {
        setRows((prev) =>
            prev.map((r) =>
                r.localId === localId ? { ...r, meta: 'original' } : r
            )
        );
    };

    // ── 保存 ────────────────────────────────────────────
    const handleSave = () => {
        const modified = rows.filter((r) => r.meta === 'modified').map((r) => r.data);
        const deleted = rows.filter((r) => r.meta === 'deleted').map((r) => r.data);
        onSave(modified, deleted);
        // 保存後: deleted行を除去し、残りをoriginalに戻す
        setRows((prev) => prev.filter((r) => r.meta !== 'deleted').map((r) => ({ ...r, meta: 'original' })));
    };

    // ── 全リセット ───────────────────────────────────────
    const handleResetAll = () => {
        setRows(initRows());
        setEditingCell(null);
        setFilterDraft({});
        setAppliedFilters({});
    };

    // ── 行スタイル ───────────────────────────────────────
    const ROW_CLASS: Record<RowMeta, string> = {
        original: '',
        modified: 'bg-warning/10',
        deleted: 'bg-error/10 opacity-60',
    };

    const modifiedCount = rows.filter((r) => r.meta === 'modified').length;

    // ── ページ計算 ─────────────────────────────────────
    const isPaginated = onPageChange !== undefined && totalCount !== undefined;
    const totalPages = isPaginated ? Math.max(1, Math.ceil(totalCount! / pageSize)) : 1;

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            {/* ── ヘッダーナビバー ── */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex items-center gap-4 w-full">
                    {onBack && (
                        <button className="btn btn-ghost btn-sm" onClick={handleBackClick}>
                            ← 戻る
                        </button>
                    )}
                    <span className="text-xl font-bold flex-1 truncate">{title}</span>
                    <div className="flex gap-2 shrink-0">
                        <button
                            className={`btn btn-primary btn-sm ${!isDirty ? 'btn-disabled opacity-50' : ''}`}
                            onClick={() => isDirty && handleSave()}
                        >
                            保存
                        </button>
                    </div>
                </div>
            </div>

            {/* ── 変更状態バナー ── */}
            {isDirty && (
                <div className="alert alert-warning shrink-0 py-2 flex items-center">
                    <span className="text-sm">
                        未保存の変更があります
                        {modifiedCount > 0 && ` (変更: ${modifiedCount}行`}
                        {deletedCount > 0 && `, 削除予定: ${deletedCount}行`}
                        {(modifiedCount > 0 || deletedCount > 0) && ')'}
                    </span>
                    <button className="btn btn-ghost btn-xs ml-auto" onClick={handleResetAll}>
                        全て元に戻す
                    </button>
                </div>
            )}

            {/* ── 凡例 ── */}
            <div className="flex gap-6 text-xs text-base-content/60 shrink-0">
                <span className="flex items-center gap-1">
                    <span className="w-3 h-3 rounded-sm bg-warning/40 inline-block" />
                    変更済み
                </span>
                <span className="flex items-center gap-1">
                    <span className="w-3 h-3 rounded-sm bg-error/40 inline-block" />
                    削除予定
                </span>
                <span className="flex items-center gap-1 ml-4">
                    セル（✎）をクリックで編集 ／ Enter・Tab・Esc で確定
                </span>
            </div>

            {/* ── グリッドテーブル ── */}
            <div className="card bg-base-100 border border-base-200 shadow-sm flex-1 min-h-0 overflow-hidden flex flex-col">
                <div className="overflow-auto flex-1">
                    <table className="table table-pin-rows table-sm w-full">
                        <thead>
                            {/* 1行目: 列ヘッダー */}
                            <tr>
                                <th className="w-8">#</th>
                                {columns.map((col) => (
                                    <th key={col.key} className={col.widthClass ?? ''}>
                                        {col.header}
                                    </th>
                                ))}
                                <th className="w-16 text-center">操作</th>
                            </tr>
                            {/* 2行目: 列フィルタ入力（検索ボタン押下で適用） */}
                            <tr className="bg-base-200/60">
                                <th />
                                {columns.map((col) => (
                                    <th key={col.key} className="py-1 px-2">
                                        <input
                                            type="text"
                                            placeholder="🔍 絞り込み（検索ボタンで適用）"
                                            className="input input-xs input-bordered w-full font-normal"
                                            value={filterDraft[col.key] ?? ''}
                                            onChange={(e) =>
                                                setFilterDraft((prev) => ({ ...prev, [col.key]: e.target.value }))
                                            }
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') handleApplyFilters();
                                            }}
                                        />
                                    </th>
                                ))}
                                <th className="py-1 px-1">
                                    <div className="flex flex-col gap-1">
                                        <button
                                            className="btn btn-primary btn-xs w-full"
                                            onClick={handleApplyFilters}
                                            title="フィルタを適用"
                                        >
                                            検索
                                        </button>
                                        {isFiltered && (
                                            <button
                                                className="btn btn-ghost btn-xs w-full"
                                                onClick={handleClearFilters}
                                                title="フィルタをクリア"
                                            >
                                                クリア
                                            </button>
                                        )}
                                    </div>
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredRows.map((row, rowIdx) => (
                                <tr
                                    key={row.localId}
                                    className={`hover ${ROW_CLASS[row.meta]}`}
                                >
                                    {/* 行番号 */}
                                    <td className="text-base-content/40 font-mono text-xs select-none text-right">
                                        {rowIdx + 1}
                                    </td>

                                    {/* データセル */}
                                    {columns.map((col) => {
                                        const isEditing =
                                            editingCell?.localId === row.localId &&
                                            editingCell?.key === col.key;
                                        const rawVal = row.data[col.key as keyof TData];
                                        const strVal = rawVal != null ? String(rawVal) : '';

                                        return (
                                            <td
                                                key={col.key}
                                                className={`p-0 ${col.editable ? 'cursor-text' : ''}`}
                                                onClick={() =>
                                                    handleCellClick(row.localId, col.key, col.editable)
                                                }
                                            >
                                                {isEditing ? (
                                                    <input
                                                        ref={inputRef}
                                                        type={col.type ?? 'text'}
                                                        className="input input-sm w-full rounded-none border-x-0 border-t-0 border-b-2 border-primary focus:outline-none bg-base-100"
                                                        value={strVal}
                                                        onChange={(e) =>
                                                            handleCellChange(
                                                                row.localId,
                                                                col.key,
                                                                col.type === 'number'
                                                                    ? Number(e.target.value)
                                                                    : e.target.value
                                                            )
                                                        }
                                                        onBlur={handleCellBlur}
                                                        onKeyDown={handleKeyDown}
                                                    />
                                                ) : (
                                                    <span
                                                        className={`block px-3 py-2 text-sm min-h-8 ${col.editable ? 'hover:bg-primary/5' : ''
                                                            }`}
                                                    >
                                                        {strVal || (
                                                            <span className="text-base-content/25 italic text-xs">
                                                                {col.editable ? 'クリックして入力' : '—'}
                                                            </span>
                                                        )}
                                                    </span>
                                                )}
                                            </td>
                                        );
                                    })}

                                    {/* 操作列: 削除予定マーク / 取り消し */}
                                    <td className="text-center">
                                        {row.meta === 'deleted' ? (
                                            <button
                                                className="btn btn-ghost btn-xs text-base-content/50"
                                                onClick={() => handleUnmarkDeleteRow(row.localId)}
                                                title="削除を取り消す"
                                            >
                                                戻す
                                            </button>
                                        ) : (
                                            <button
                                                className="btn btn-ghost btn-xs text-error"
                                                onClick={() => handleMarkDeleteRow(row.localId)}
                                                title="削除予定としてマーク"
                                            >
                                                削除
                                            </button>
                                        )}
                                    </td>
                                </tr>
                            ))}

                            {/* 空状態 */}
                            {filteredRows.length === 0 && (
                                <tr>
                                    <td
                                        colSpan={columns.length + 2}
                                        className="text-center text-base-content/40 py-12"
                                    >
                                        {isFiltered
                                            ? '絞り込み条件に一致するエントリがありません。'
                                            : 'エントリがありません。'}
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>

                {/* フッター: 件数 + ページネーション */}
                <div className="px-4 py-2 border-t border-base-200 text-xs text-base-content/60 shrink-0 flex items-center gap-4 flex-wrap">
                    <div className="flex gap-4 flex-1">
                        {isFiltered
                            ? <span><span className="text-primary font-bold">{filteredRows.filter(r => r.meta !== 'deleted').length}</span> 件表示中</span>
                            : isPaginated
                                ? <span>全 <span className="text-primary font-bold">{totalCount!.toLocaleString()}</span> 件中 {((currentPage! - 1) * pageSize + 1).toLocaleString()}〜{Math.min(currentPage! * pageSize, totalCount!).toLocaleString()} 件表示</span>
                                : <span>合計 {rows.length.toLocaleString()} 件</span>
                        }
                        {modifiedCount > 0 && <span className="text-warning">変更 {modifiedCount} 件</span>}
                        {deletedCount > 0 && <span className="text-error">削除予定 {deletedCount} 件</span>}
                    </div>
                    {/* ページネーション UI */}
                    {isPaginated && totalPages > 1 && (
                        <div className="flex items-center gap-1">
                            <button
                                className="btn btn-xs btn-ghost"
                                disabled={currentPage! <= 1}
                                onClick={() => onPageChange!(currentPage! - 1)}
                            >«</button>
                            {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                                // 前後2ページを表示
                                const mid = Math.min(Math.max(currentPage!, 4), totalPages - 3);
                                const page = totalPages <= 7 ? i + 1 :
                                    i === 0 ? 1 :
                                        i === 6 ? totalPages :
                                            i === 1 && mid > 3 ? -1 :
                                                i === 5 && mid < totalPages - 3 ? -1 :
                                                    mid + i - 3;
                                if (page === -1) return <span key={i} className="px-1">…</span>;
                                return (
                                    <button
                                        key={i}
                                        className={`btn btn-xs ${currentPage === page ? 'btn-primary' : 'btn-ghost'}`}
                                        onClick={() => onPageChange!(page)}
                                    >{page}</button>
                                );
                            })}
                            <button
                                className="btn btn-xs btn-ghost"
                                disabled={currentPage! >= totalPages}
                                onClick={() => onPageChange!(currentPage! + 1)}
                            >»</button>
                        </div>
                    )}
                </div>
            </div>

            {/* ── 「戻る」未保存警告モーダル ── */}
            {showBackModal && (
                <dialog open className="modal modal-open">
                    <div className="modal-box border border-warning">
                        <h3 className="font-bold text-lg text-warning">未保存の変更があります</h3>
                        <p className="py-4">
                            変更内容（{modifiedCount > 0 ? `編集: ${modifiedCount}行` : ''}
                            {deletedCount > 0 ? `${modifiedCount > 0 ? '、' : ''}削除予定: ${deletedCount}行` : ''}）が破棄されます。<br />
                            本当に戻りますか？
                        </p>
                        <div className="modal-action">
                            <button className="btn btn-ghost" onClick={() => setShowBackModal(false)}>キャンセル</button>
                            <button
                                className="btn btn-warning"
                                onClick={() => { setShowBackModal(false); onBack?.(); }}
                            >
                                変更を破棄して戻る
                            </button>
                        </div>
                    </div>
                </dialog>
            )}
        </div>
    );
}

export default GridEditor;
