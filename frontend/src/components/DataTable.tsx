import React from 'react';
import { useReactTable, getCoreRowModel, flexRender } from '@tanstack/react-table';
import type { ColumnDef } from '@tanstack/react-table';

interface DataTableProps<TData> {
    /** 列定義 */
    columns: ColumnDef<TData, unknown>[];
    /** 行データ */
    data: TData[];
    /** 選択行の id (行インデックスの文字列) */
    selectedRowId?: string | null;
    /** 行クリック時コールバック。同じ行を再クリックすると null で呼ばれる (トグル) */
    onRowSelect?: (row: TData | null, rowId: string | null) => void;
    /** テーブルヘッダー左上に表示するタイトル */
    title?: string;
    /** テーブルヘッダー右側に追加するアクション要素 */
    headerActions?: React.ReactNode;
    /** コンポーネント自体を折りたたみ可能にするか */
    collapsible?: boolean;
    /** 折りたたみの初期状態 */
    defaultOpen?: boolean;
    /** 列ごとのテキストフィルタを有効にするか */
    enableColumnFilter?: boolean;
}

function DataTable<TData>({
    columns,
    data,
    selectedRowId,
    onRowSelect,
    title,
    headerActions,
    collapsible = false,
    defaultOpen = true,
    enableColumnFilter = false,
}: DataTableProps<TData>) {
    // TanStack の rowSelection は { [rowIndex: string]: boolean } 形式
    const rowSelection: Record<string, boolean> = selectedRowId != null
        ? { [selectedRowId]: true }
        : {};

    const table = useReactTable({
        data,
        columns,
        state: { rowSelection },
        getCoreRowModel: getCoreRowModel(),
        enableRowSelection: true,
        getRowId: (row: any, index: number) => row.id !== undefined ? String(row.id) : String(index),
        // 選択制御は onRowSelect 経由で親が行うため内部更新は無効化
        onRowSelectionChange: () => { },
    });

    const handleRowClick = (rowId: string, rowData: TData) => {
        if (!onRowSelect) return;
        if (selectedRowId === rowId) {
            // 同一行を再クリック → 閉じる
            onRowSelect(null, null);
        } else {
            onRowSelect(rowData, rowId);
        }
    };

    const Wrapper = collapsible ? 'details' : 'div' as const;
    const HeaderWrapper = collapsible ? 'summary' : 'div' as const;

    return (
        <Wrapper
            className={collapsible
                ? "collapse collapse-arrow bg-base-100 border border-base-200 shadow-sm flex flex-col open:flex-1 open:min-h-0"
                : "card bg-base-100 border border-base-200 shadow-sm flex-1 flex flex-col min-h-0"}
            open={collapsible ? defaultOpen : undefined}
            {...(collapsible ? {} : {})}
        >
            {(title || headerActions) && (
                <HeaderWrapper className={collapsible ? "collapse-title flex justify-between items-center px-6 py-3 min-h-0" : "flex justify-between items-center px-6 pt-4 pb-2"}>
                    {title && <h2 className="card-title text-base font-bold">{title}</h2>}
                    {headerActions && <div className="flex gap-2" onClick={e => collapsible && e.stopPropagation()}>{headerActions}</div>}
                </HeaderWrapper>
            )}
            <div className={collapsible ? "collapse-content flex-1 flex flex-col min-h-0 p-0" : "contents"}>
                <div className="overflow-x-auto flex-1 overflow-y-auto">
                    <table className="table table-zebra table-pin-rows w-full">
                        <thead>
                            {table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th key={header.id}>
                                            {header.isPlaceholder
                                                ? null
                                                : flexRender(header.column.columnDef.header, header.getContext())}
                                        </th>
                                    ))}
                                </tr>
                            ))}
                            {enableColumnFilter && table.getHeaderGroups().map((headerGroup) => (
                                <tr key={`filter-${headerGroup.id}`}>
                                    {headerGroup.headers.map((header) => (
                                        <th key={`filter-${header.id}`} className="p-2 py-1 bg-base-200">
                                            {header.isPlaceholder ? null : (
                                                <input
                                                    type="text"
                                                    placeholder="フィルタ..."
                                                    className="input input-xs input-bordered w-full font-normal"
                                                    onChange={() => { /* モック用スタブ */ }}
                                                />
                                            )}
                                        </th>
                                    ))}
                                </tr>
                            ))}
                        </thead>
                        <tbody>
                            {table.getRowModel().rows.map((row) => {
                                const isSelected = row.id === selectedRowId;
                                return (
                                    <tr
                                        key={row.id}
                                        className={`cursor-pointer transition-colors ${isSelected ? 'bg-primary/10 hover:bg-primary/15' : 'hover:bg-base-200'}`}
                                        onClick={() => handleRowClick(row.id, row.original)}
                                    >
                                        {row.getVisibleCells().map((cell) => (
                                            <td key={cell.id}>
                                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                            </td>
                                        ))}
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        </Wrapper>
    );
}

export default DataTable;
