import React, { useState, useEffect } from 'react';

// ── フィールド定義 ──────────────────────────────────────
// ページ側で FieldDef<MyType>[] を定義して渡す。
// RecordEditor 自体は TData の形を知らない。
export type FieldType = 'text' | 'number' | 'textarea' | 'readonly' | 'code' | 'badge';

export interface FieldDef<TData> {
    /** TData のキー */
    key: keyof TData & string;
    /** 表示ラベル */
    label: string;
    /** フィールドの種別 */
    type: FieldType;
    /** 2カラムグリッドで全幅を使うか (デフォルト: 'half') */
    span?: 'full' | 'half';
    /** textarea の行数 (type='textarea' のみ) */
    rows?: number;
    /** badge のクラスを動的に決定 (type='badge' のみ) */
    badgeClass?: (value: unknown) => string;
    /** null / undefined のときに表示するフォールバック文字列 */
    fallback?: string;
}

// ── Props ───────────────────────────────────────────────
interface RecordEditorProps<TData extends object> {
    /** ヘッダーに表示するタイトル */
    title: string;
    /** 編集対象レコード (record が変わると draft もリセット) */
    record: TData;
    /** フィールド定義リスト (外から注入。データ型に依存しない) */
    fields: FieldDef<TData>[];
    /** 「← 戻る」ボタンのコールバック */
    onBack: () => void;
    /** 保存ボタンのコールバック。未指定なら保存ボタン非表示 */
    onSave?: (updated: TData) => void;
    /** 削除ボタンのコールバック。未指定なら削除ボタン非表示 */
    onDelete?: () => void;
}

// ── コンポーネント ──────────────────────────────────────
function RecordEditor<TData extends object>({
    title,
    record,
    fields,
    onBack,
    onSave,
    onDelete,
}: RecordEditorProps<TData>) {
    const [draft, setDraft] = useState<TData>(record);

    // record (選択行) が変わったら draft をリセット
    useEffect(() => {
        setDraft(record);
    }, [record]);

    const handleChange = (key: keyof TData, value: unknown) => {
        setDraft((prev) => ({ ...prev, [key]: value }));
    };

    const isDirty = JSON.stringify(draft) !== JSON.stringify(record);

    // フィールドが編集可能かを判定
    const isEditable = (type: FieldType) =>
        type === 'text' || type === 'number' || type === 'textarea';

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            {/* ── ヘッダーナビバー ── */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex items-center gap-4 w-full">
                    <button className="btn btn-ghost btn-sm" onClick={onBack}>
                        ← 戻る
                    </button>
                    <span className="text-xl font-bold flex-1 truncate">{title}</span>
                    <div className="flex gap-2 shrink-0">
                        {onDelete && (
                            <button
                                className="btn btn-outline btn-error btn-sm"
                                onClick={onDelete}
                            >
                                削除
                            </button>
                        )}
                        {onSave && (
                            <button
                                className={`btn btn-primary btn-sm ${!isDirty ? 'btn-disabled opacity-50' : ''}`}
                                onClick={() => isDirty && onSave(draft)}
                            >
                                保存
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* 変更有り通知バナー */}
            {isDirty && (
                <div className="alert alert-warning shrink-0 py-2">
                    <span className="text-sm">変更があります。「保存」ボタンで確定してください。</span>
                    <button
                        className="btn btn-ghost btn-xs ml-auto"
                        onClick={() => setDraft(record)}
                    >
                        元に戻す
                    </button>
                </div>
            )}

            {/* ── フォームカード ── */}
            <div className="card bg-base-100 border border-base-200 shadow-sm flex-1 min-h-0">
                <div className="card-body overflow-y-auto">
                    <div className="grid grid-cols-2 gap-x-8 gap-y-6">
                        {fields.map((field) => {
                            const raw = draft[field.key as keyof TData];
                            const strVal = raw != null ? String(raw) : (field.fallback ?? '—');
                            const colClass = field.span === 'full' ? 'col-span-2' : '';

                            return (
                                <div key={field.key} className={`flex flex-col gap-1 ${colClass}`}>
                                    {/* ラベル */}
                                    <label
                                        htmlFor={`field-${field.key}`}
                                        className="text-xs font-bold uppercase tracking-wide text-base-content/60"
                                    >
                                        {field.label}
                                        {isEditable(field.type) && (
                                            <span className="ml-1 badge badge-outline badge-xs">編集可</span>
                                        )}
                                    </label>

                                    {/* badge */}
                                    {field.type === 'badge' && (
                                        <div
                                            className={`badge badge-sm w-fit ${field.badgeClass ? field.badgeClass(raw) : ''}`}
                                        >
                                            {strVal}
                                        </div>
                                    )}

                                    {/* コードブロック */}
                                    {field.type === 'code' && (
                                        <div className="bg-base-200 rounded px-3 py-2 font-mono text-xs break-all">
                                            {strVal}
                                        </div>
                                    )}

                                    {/* 読み取り専用テキスト */}
                                    {field.type === 'readonly' && (
                                        <span className="text-sm font-mono">{strVal}</span>
                                    )}

                                    {/* 編集可能: text / number */}
                                    {(field.type === 'text' || field.type === 'number') && (
                                        <input
                                            id={`field-${field.key}`}
                                            type={field.type}
                                            className="input input-bordered input-sm"
                                            value={strVal !== '—' ? strVal : ''}
                                            onChange={(e) =>
                                                handleChange(
                                                    field.key as keyof TData,
                                                    field.type === 'number'
                                                        ? Number(e.target.value)
                                                        : e.target.value
                                                )
                                            }
                                        />
                                    )}

                                    {/* 編集可能: textarea */}
                                    {field.type === 'textarea' && (
                                        <textarea
                                            id={`field-${field.key}`}
                                            className="textarea textarea-bordered text-sm"
                                            rows={field.rows ?? 3}
                                            value={strVal !== '—' ? strVal : ''}
                                            onChange={(e) =>
                                                handleChange(
                                                    field.key as keyof TData,
                                                    e.target.value
                                                )
                                            }
                                        />
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </div>
    );
}

export default RecordEditor;
