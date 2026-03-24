import type {ColumnDef} from '@tanstack/react-table';
import React from 'react';
import DataTable from '../DataTable';
import type {PersonaListPager, PersonaListRow, PersonaListStateTone} from './types';

interface SharedPersonaListProps {
    rows: PersonaListRow[];
    selectedRowId?: string | null;
    title?: string;
    totalCount?: number;
    pager?: PersonaListPager;
    headerActions?: React.ReactNode;
    isLoading?: boolean;
    loadingMessage?: string;
    emptyMessage?: string;
    onSelectRow?: (row: PersonaListRow | null, rowId: string | null) => void;
}

const TONE_CLASS: Record<PersonaListStateTone, string> = {
    neutral: 'badge-outline',
    info: 'badge-info badge-outline',
    warning: 'badge-warning badge-outline',
    success: 'badge-success badge-outline',
    error: 'badge-error badge-outline',
};

const COLUMNS: ColumnDef<PersonaListRow, unknown>[] = [
    {
        accessorKey: 'formId',
        header: 'FormID',
        cell: (info) => <span className="font-mono text-sm">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'sourcePlugin',
        header: 'プラグイン名',
        cell: (info) => <span className="font-mono text-xs">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'npcName',
        header: 'NPC名',
        cell: (info) => {
            const row = info.row.original;
            const hasMeta = Boolean(row.editorId) || Boolean(row.updatedAt) || Boolean(row.stateBadge?.label);

            return (
                <div className="min-w-0">
                    <div className="truncate font-semibold">{row.npcName || '(名称なし)'}</div>
                    {hasMeta && (
                        <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-base-content/70">
                            {row.editorId && (
                                <span className="font-mono">EditorID: {row.editorId}</span>
                            )}
                            {row.updatedAt && (
                                <span>更新: {row.updatedAt}</span>
                            )}
                            {row.stateBadge?.label && (
                                <span className={`badge badge-xs ${TONE_CLASS[row.stateBadge.tone ?? 'neutral']}`}>
                                    {row.stateBadge.label}
                                </span>
                            )}
                        </div>
                    )}
                </div>
            );
        },
    },
];

const renderColumnHeader = (column: ColumnDef<PersonaListRow, unknown>, index: number): React.ReactNode => {
    if (typeof column.header === 'string') {
        return column.header;
    }

    return `column-${index + 1}`;
};

/**
 * MasterPersona と TranslationFlow persona phase の一覧 shell を共有する。
 * 一覧表示と選択イベントのみを担当し、詳細ペインや phase 制御は扱わない。
 */
export function SharedPersonaList({
    rows,
    selectedRowId = null,
    title = 'ペルソナ一覧',
    totalCount,
    pager,
    headerActions,
    isLoading = false,
    loadingMessage = '読込中です。',
    emptyMessage = '表示できる NPC がありません。',
    onSelectRow,
}: SharedPersonaListProps) {
    const totalLabel = typeof totalCount === 'number' ? `${totalCount.toLocaleString()} 件` : `${rows.length.toLocaleString()} 件`;
    const shouldShowStateTable = isLoading || rows.length === 0;
    const stateMessage = isLoading ? loadingMessage : emptyMessage;
    const pagerLabel = pager
        ? `${pager.page} / ${Math.max(1, pager.totalPages)}` 
        : null;

    return (
        <div className="flex h-full min-h-[28rem] flex-col">
            {shouldShowStateTable ? (
                <div className="card bg-base-100 border border-base-200 shadow-sm flex-1 flex flex-col min-h-0">
                    <div className="flex flex-col xl:flex-row justify-between xl:items-center gap-4 px-6 pt-4 pb-2">
                        <h2 className="card-title text-base font-bold whitespace-nowrap shrink-0">{title}</h2>
                        <div className="flex flex-wrap gap-2 w-full xl:w-auto xl:justify-end">
                            <div className="flex flex-wrap items-center gap-2">
                                {headerActions}
                                <span className="text-xs text-base-content/60">{totalLabel}</span>
                                {pager && (
                                    <div className="flex items-center gap-2 text-xs">
                                        <button
                                            type="button"
                                            className="btn btn-outline btn-xs"
                                            onClick={pager.onPrevPage}
                                            disabled={Boolean(pager.disablePrev)}
                                        >
                                            前へ
                                        </button>
                                        <span>{pagerLabel}</span>
                                        <button
                                            type="button"
                                            className="btn btn-outline btn-xs"
                                            onClick={pager.onNextPage}
                                            disabled={Boolean(pager.disableNext)}
                                        >
                                            次へ
                                        </button>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                    <div className="flex-1 flex flex-col min-h-0">
                        <div className="overflow-x-auto overflow-y-auto flex-1 min-h-0">
                            <table className="table table-zebra table-pin-rows w-full">
                                <thead>
                                    <tr>
                                        {COLUMNS.map((column, index) => (
                                            <th key={`column-${index + 1}`}>
                                                {renderColumnHeader(column, index)}
                                            </th>
                                        ))}
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td colSpan={COLUMNS.length} className="py-6 text-sm text-base-content/60">
                                            {stateMessage}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            ) : (
                <DataTable
                    columns={COLUMNS}
                    data={rows}
                    title={title}
                    selectedRowId={selectedRowId}
                    onRowSelect={onSelectRow}
                    headerActions={(
                        <div className="flex flex-wrap items-center gap-2">
                            {headerActions}
                            <span className="text-xs text-base-content/60">{totalLabel}</span>
                            {pager && (
                                <div className="flex items-center gap-2 text-xs">
                                    <button
                                        type="button"
                                        className="btn btn-outline btn-xs"
                                        onClick={pager.onPrevPage}
                                        disabled={Boolean(pager.disablePrev)}
                                    >
                                        前へ
                                    </button>
                                    <span>{pagerLabel}</span>
                                    <button
                                        type="button"
                                        className="btn btn-outline btn-xs"
                                        onClick={pager.onNextPage}
                                        disabled={Boolean(pager.disableNext)}
                                    >
                                        次へ
                                    </button>
                                </div>
                            )}
                        </div>
                    )}
                />
            )}
        </div>
    );
}
