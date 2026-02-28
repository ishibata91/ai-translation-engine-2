import React, { useState, useRef, useEffect } from 'react';

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
type RowMeta = 'original' | 'modified' | 'new';

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
    onSave: (rows: TData[]) => void;
    /** 「行追加」で生成する空行ファクトリ (省略時はボタン非表示) */
    newRowFactory?: () => TData;
    /** 変更状態が変わったときのコールバック */
    onDirtyChange?: (isDirty: boolean) => void;
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
    newRowFactory,
    onDirtyChange,
}: GridEditorProps<TData>) {
    // draft: 編集中の行リスト
    const initRows = (): RowWithMeta<TData>[] =>
        initialData.map((d) => ({ data: d, meta: 'original', localId: nextLocalId() }));

    const [rows, setRows] = useState<RowWithMeta<TData>[]>(initRows);

    // 各列のフィルタ文字列 { [columnKey]: filterString }
    const [columnFilters, setColumnFilters] = useState<Record<string, string>>({});

    // フィルタ変更: editingCell もリセット
    const handleFilterChange = (key: string, value: string) => {
        setColumnFilters((prev) => ({ ...prev, [key]: value }));
        setEditingCell(null);
    };

    // フィルタ適用後の行（表示用）
    const filteredRows = rows.filter((row) =>
        columns.every((col) => {
            const filter = (columnFilters[col.key] ?? '').trim().toLowerCase();
            if (!filter) return true;
            const val = String(row.data[col.key as keyof TData] ?? '').toLowerCase();
            return val.includes(filter);
        })
    );

    const isFiltered = Object.values(columnFilters).some((v) => v.trim() !== '');

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

    useEffect(() => {
        onDirtyChange?.(isDirty);
    }, [isDirty, onDirtyChange]);

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
                        meta: r.meta === 'new' ? 'new' : 'modified',
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
    const handleAddRow = () => {
        if (!newRowFactory) return;
        setRows((prev) => [
            ...prev,
            { data: newRowFactory(), meta: 'new', localId: nextLocalId() },
        ]);
    };

    const handleDeleteRow = (localId: number) => {
        setRows((prev) => prev.filter((r) => r.localId !== localId));
    };

    // ── 保存 ────────────────────────────────────────────
    const handleSave = () => {
        onSave(rows.map((r) => r.data));
        setRows((prev) => prev.map((r) => ({ ...r, meta: 'original' })));
    };

    // ── 全リセット ───────────────────────────────────────
    const handleResetAll = () => {
        setRows(initRows());
        setEditingCell(null);
    };

    // ── 行スタイル ───────────────────────────────────────
    const ROW_CLASS: Record<RowMeta, string> = {
        original: '',
        modified: 'bg-warning/10',
        new: 'bg-success/10',
    };

    const modifiedCount = rows.filter((r) => r.meta === 'modified').length;
    const newCount = rows.filter((r) => r.meta === 'new').length;

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            {/* ── ヘッダーナビバー ── */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex items-center gap-4 w-full">
                    {onBack && (
                        <button className="btn btn-ghost btn-sm" onClick={onBack}>
                            ← 戻る
                        </button>
                    )}
                    <span className="text-xl font-bold flex-1 truncate">{title}</span>
                    <div className="flex gap-2 shrink-0">
                        {newRowFactory && (
                            <button className="btn btn-outline btn-sm" onClick={handleAddRow}>
                                ＋ 行追加
                            </button>
                        )}
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
                        {newCount > 0 && `, 追加: ${newCount}行`}
                        {(modifiedCount > 0 || newCount > 0) && ')'}
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
                    <span className="w-3 h-3 rounded-sm bg-success/40 inline-block" />
                    新規追加
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
                            {/* 2行目: 列フィルタ入力 */}
                            <tr className="bg-base-200/60">
                                <th />
                                {columns.map((col) => (
                                    <th key={col.key} className="py-1 px-2">
                                        <input
                                            type="text"
                                            placeholder="🔍 絞り込み"
                                            className="input input-xs input-bordered w-full font-normal"
                                            value={columnFilters[col.key] ?? ''}
                                            onChange={(e) => handleFilterChange(col.key, e.target.value)}
                                        />
                                    </th>
                                ))}
                                <th>
                                    {isFiltered && (
                                        <button
                                            className="btn btn-ghost btn-xs w-full"
                                            onClick={() => {
                                                setColumnFilters({});
                                                setEditingCell(null);
                                            }}
                                            title="全フィルタをクリア"
                                        >
                                            ✕
                                        </button>
                                    )}
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

                                    {/* 操作列 */}
                                    <td className="text-center">
                                        <button
                                            className="btn btn-ghost btn-xs text-error"
                                            onClick={() => handleDeleteRow(row.localId)}
                                            title="この行を削除"
                                        >
                                            削除
                                        </button>
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
                                            : 'エントリがありません。「＋ 行追加」から追加してください。'}
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>

                {/* フッター: 件数 */}
                <div className="px-4 py-2 border-t border-base-200 text-xs text-base-content/60 shrink-0 flex gap-4">
                    {isFiltered
                        ? <span><span className="text-primary font-bold">{filteredRows.length}</span> 件表示中 / 全 {rows.length} 件</span>
                        : <span>合計 {rows.length} 件</span>
                    }
                    {modifiedCount > 0 && <span className="text-warning">変更 {modifiedCount} 件</span>}
                    {newCount > 0 && <span className="text-success">追加 {newCount} 件</span>}
                </div>
            </div>
        </div>
    );
}

export default GridEditor;
