import { useEffect, useMemo, useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import {
    DictDeleteEntry,
    DictDeleteSource,
    DictGetEntriesPaginated,
    DictGetSources,
    DictSearchAllEntriesPaginated,
    DictStartImport,
    DictUpdateEntry,
    SelectFiles,
} from '../../../wailsjs/go/main/App';
import { useWailsEvent } from '../../useWailsEvent';
import {
    mapEntriesPaginatedResponse,
    mapSourcesResponse,
    toDictUpdateEntryPayload,
} from './adapters';
import type {
    DictEntry,
    DictSourceRow,
    DictionaryProgressEvent,
    DictionaryBuilderActions,
    DictionaryBuilderState,
    UseDictionaryBuilderResult,
    SourceStatus,
    View,
} from './types';
import { STATUS_BADGE } from './types';

const PAGE_SIZE = 500;

const showModal = (id: string) => {
    const modal = document.getElementById(id) as HTMLDialogElement | null;
    modal?.showModal();
};

/**
 * 辞書構築画面の state、action、UI 契約をまとめて返す。
 */
export function useDictionaryBuilder(): UseDictionaryBuilderResult {
    const [view, setView] = useState<View>('list');
    const [selectedRow, setSelectedRow] = useState<DictSourceRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
    const [isImporting, setIsImporting] = useState(false);
    const [importMessages, setImportMessages] = useState<Record<string, string>>({});
    const [deletingRowId, setDeletingRowId] = useState<string | null>(null);
    const [showCrossSearch, setShowCrossSearch] = useState(false);

    const [sources, setSources] = useState<DictSourceRow[]>([]);
    const [entries, setEntries] = useState<DictEntry[]>([]);
    const [entryPage, setEntryPage] = useState(1);
    const [entryTotal, setEntryTotal] = useState(0);
    const [entryQuery] = useState('');
    const [entryFilters, setEntryFilters] = useState<Record<string, string>>({});

    const [crossEntries, setCrossEntries] = useState<DictEntry[]>([]);
    const [crossPage, setCrossPage] = useState(1);
    const [crossTotal, setCrossTotal] = useState(0);
    const [crossQuery, setCrossQuery] = useState('');
    const [crossFilters, setCrossFilters] = useState<Record<string, string>>({});

    const fetchSources = async () => {
        try {
            const response = await DictGetSources();
            setSources(mapSourcesResponse(response));
        } catch (error) {
            console.error('Failed to fetch sources:', error);
        }
    };

    const fetchEntriesPaginated = async (
        idStr: string,
        page: number,
        query: string,
        filters: Record<string, string>,
    ) => {
        try {
            const idNum = Number.parseInt(idStr, 10);
            const response = await DictGetEntriesPaginated(idNum, query, filters, page, PAGE_SIZE);
            const { entries: mappedEntries, totalCount } = mapEntriesPaginatedResponse(response);
            setEntries(mappedEntries);
            setEntryTotal(totalCount);
        } catch (error) {
            console.error('fetchEntriesPaginated failed:', error);
        }
    };

    const fetchCrossSearch = async (
        query: string,
        filters: Record<string, string>,
        page: number,
    ) => {
        try {
            const response = await DictSearchAllEntriesPaginated(query, filters, page, PAGE_SIZE);
            const { entries: mappedEntries, totalCount } = mapEntriesPaginatedResponse(response);
            setCrossEntries(mappedEntries);
            setCrossTotal(totalCount);
        } catch (error) {
            console.error('fetchCrossSearch failed:', error);
        }
    };

    useEffect(() => {
        void fetchSources();
    }, []);

    useWailsEvent<DictionaryProgressEvent>('dictionary:import_progress', (payload) => {
        const corrId = payload.CorrelationID;
        if (payload.Status === 'COMPLETED' || payload.Status === 'FAILED') {
            setImportMessages((prev) => {
                const next = { ...prev };
                delete next[corrId];
                return next;
            });
            setIsImporting(false);
            void fetchSources();
            return;
        }

        setImportMessages((prev) => ({ ...prev, [corrId]: payload.Message }));
        setIsImporting(true);
        if (payload.Completed > 0 && payload.Completed % 1000 === 0) {
            void fetchSources();
        }
    });

    const handleImport = async () => {
        if (selectedFiles.length === 0) {
            return;
        }
        setIsImporting(true);
        for (const filePath of selectedFiles) {
            try {
                const resultId = await DictStartImport(filePath);
                console.warn('Started import with ID:', resultId);
            } catch (error) {
                console.error('Import error:', error);
            }
        }
        setSelectedFiles([]);
    };

    const handleEntryPageChange = (page: number) => {
        if (!selectedRowId) {
            return;
        }
        setEntryPage(page);
        void fetchEntriesPaginated(selectedRowId, page, entryQuery, entryFilters);
    };

    const handleRowSelectAndFetch = (row: DictSourceRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
        if (!rowId) {
            return;
        }

        setEntryPage(1);
        setEntryFilters({});
        void fetchEntriesPaginated(rowId, 1, '', {});
    };

    const handleEntrySearch = (filters: Record<string, string>) => {
        setEntryFilters(filters);
        setEntryPage(1);
        if (!selectedRowId) {
            return;
        }
        void fetchEntriesPaginated(selectedRowId, 1, entryQuery, filters);
    };

    const handleCrossSearchExecute = (query: string) => {
        setCrossQuery(query);
        setCrossFilters({});
        setCrossPage(1);
        void fetchCrossSearch(query, {}, 1);
        setShowCrossSearch(false);
        setView('cross-search');
    };

    const handleCrossSearchFilter = (filters: Record<string, string>) => {
        setCrossFilters(filters);
        setCrossPage(1);
        void fetchCrossSearch(crossQuery, filters, 1);
    };

    const handleCrossPageChange = (page: number) => {
        setCrossPage(page);
        void fetchCrossSearch(crossQuery, crossFilters, page);
    };

    const handleSelectFilesClick = async () => {
        try {
            const files = await SelectFiles();
            if (!files || files.length === 0) {
                return;
            }
            setSelectedFiles((prev) => {
                const currentPaths = new Set(prev);
                const uniqueNewFiles = files.filter((filePath) => !currentPaths.has(filePath));
                return [...prev, ...uniqueNewFiles];
            });
        } catch (error) {
            console.error('Failed to select files:', error);
        }
    };

    const removeSelectedFile = (pathToRemove: string) => {
        setSelectedFiles((prev) => prev.filter((path) => path !== pathToRemove));
    };

    const handleDeleteSource = async () => {
        if (!deletingRowId) {
            return;
        }

        try {
            await DictDeleteSource(Number.parseInt(deletingRowId, 10));
            await fetchSources();
            if (selectedRowId === deletingRowId) {
                handleRowSelectAndFetch(null, null);
            }
        } catch (error) {
            console.error('Failed to delete source:', error);
        } finally {
            setDeletingRowId(null);
        }
    };

    const handleCancelDelete = () => {
        setDeletingRowId(null);
    };

    const handleEntriesSave = async (modified: DictEntry[], deleted: DictEntry[]) => {
        for (const entry of modified) {
            try {
                await DictUpdateEntry(toDictUpdateEntryPayload(entry));
            } catch (error) {
                console.error('UpdateEntry failed:', error);
            }
        }

        for (const entry of deleted) {
            try {
                await DictDeleteEntry(entry.id);
            } catch (error) {
                console.error('DeleteEntry failed:', error);
            }
        }

        if (selectedRowId) {
            await fetchEntriesPaginated(selectedRowId, entryPage, entryQuery, entryFilters);
        }
    };

    const handleCrossSave = async (modified: DictEntry[], deleted: DictEntry[]) => {
        for (const entry of modified) {
            try {
                await DictUpdateEntry(toDictUpdateEntryPayload(entry));
            } catch (error) {
                console.error('UpdateEntry failed:', error);
            }
        }

        for (const entry of deleted) {
            try {
                await DictDeleteEntry(entry.id);
            } catch (error) {
                console.error('DeleteEntry failed:', error);
            }
        }

        await fetchCrossSearch(crossQuery, crossFilters, crossPage);
    };

    const sourceColumns = useMemo<ColumnDef<DictSourceRow, unknown>[]>(() => [
        {
            accessorKey: 'fileName',
            header: 'ソース名 (ファイル名)',
            cell: (info) => <span className="font-mono text-sm">{info.getValue() as string}</span>,
        },
        {
            accessorKey: 'format',
            header: '形式',
            cell: (info) => (
                <div className="badge badge-outline badge-sm font-mono">{info.getValue() as string}</div>
            ),
        },
        {
            accessorKey: 'entryCount',
            header: 'エントリ数',
            cell: (info) => (
                <span className="font-mono text-right block">
                    {(info.getValue() as number).toLocaleString()}
                </span>
            ),
        },
        { accessorKey: 'updatedAt', header: '最終更新日時' },
        {
            accessorKey: 'status',
            header: 'ステータス',
            cell: (info) => {
                const status = info.getValue() as SourceStatus;
                return <div className={`badge badge-sm ${STATUS_BADGE[status]}`}>{status}</div>;
            },
        },
        {
            id: 'actions',
            header: 'アクション',
            cell: (info) => (
                <button
                    className="btn btn-ghost btn-xs text-error"
                    disabled={isImporting}
                    onClick={(event) => {
                        event.stopPropagation();
                        setDeletingRowId(info.row.original.id);
                        showModal('delete_modal');
                    }}
                >
                    削除
                </button>
            ),
        },
    ], [isImporting]);

    const state: DictionaryBuilderState = {
        view,
        selectedRow,
        selectedRowId,
        selectedFiles,
        isImporting,
        importMessages,
        showCrossSearch,
        sources,
        entries,
        entryPage,
        entryTotal,
        entryQuery,
        crossEntries,
        crossPage,
        crossTotal,
        crossQuery,
    };

    const actions: DictionaryBuilderActions = {
        setView,
        openCrossSearch: () => setShowCrossSearch(true),
        closeCrossSearch: () => setShowCrossSearch(false),
        handleImport,
        handleEntrySearch,
        handleEntryPageChange,
        handleRowSelectAndFetch,
        handleCrossSearchExecute,
        handleCrossSearchFilter,
        handleCrossPageChange,
        handleSelectFilesClick,
        removeSelectedFile,
        handleDeleteSource,
        handleCancelDelete,
        handleEntriesSave,
        handleCrossSave,
    };

    return {
        state,
        actions,
        ui: {
            sourceColumns,
        },
        constants: {
            pageSize: PAGE_SIZE,
        },
    };
}

