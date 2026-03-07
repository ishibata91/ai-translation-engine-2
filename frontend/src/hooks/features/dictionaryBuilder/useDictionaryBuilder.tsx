import { useState, useMemo, useEffect } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import {
    DictGetSources,
    DictStartImport,
    DictGetEntriesPaginated,
    DictSearchAllEntriesPaginated,
    DictUpdateEntry,
    DictDeleteEntry,
    SelectFiles,
    DictDeleteSource,
} from '../../../wailsjs/go/main/App';
import * as Events from '../../../wailsjs/runtime/runtime';
import type {
    SourceStatus,
    DictSourceRow,
    DictEntry,
    View,
    DictionaryProgressEvent,
} from './types';
import { STATUS_BADGE } from './types';

const PAGE_SIZE = 500;

const showModal = (id: string) => {
    const modal = document.getElementById(id) as HTMLDialogElement;
    modal?.showModal();
};

export function useDictionaryBuilder() {

    const [view, setView] = useState<View>('list');
    const [selectedRow, setSelectedRow] = useState<DictSourceRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
    const [isImporting, setIsImporting] = useState<boolean>(false);
    const [importMessages, setImportMessages] = useState<Record<string, string>>({});
    const [deletingRowId, setDeletingRowId] = useState<string | null>(null);
    const [showCrossSearch, setShowCrossSearch] = useState(false);

    // 実データ保持用
    const [sources, setSources] = useState<DictSourceRow[]>([]);
    const [entries, setEntries] = useState<DictEntry[]>([]);
    const [entryPage, setEntryPage] = useState(1);
    const [entryTotal, setEntryTotal] = useState(0);
    const [entryQuery, setEntryQuery] = useState('');
    const [entryFilters, setEntryFilters] = useState<Record<string, string>>({});

    // 横断検索結果
    const [crossEntries, setCrossEntries] = useState<DictEntry[]>([]);
    const [crossPage, setCrossPage] = useState(1);
    const [crossTotal, setCrossTotal] = useState(0);
    const [crossQuery, setCrossQuery] = useState('');
    const [crossFilters, setCrossFilters] = useState<Record<string, string>>({});

    // Wails からソース一覧を取得する
    const fetchSources = async () => {
        try {
            const result = await DictGetSources() as any[];
            if (!result) return;
            const formatted = result.map(s => ({
                id: s.id.toString(),
                fileName: s.file_name,
                format: s.format,
                entryCount: s.entry_count,
                status: (s.status === 'COMPLETED' ? '完了' : s.status === 'ERROR' ? 'エラー' : 'インポート中') as SourceStatus,
                updatedAt: s.imported_at ? new Date(s.imported_at).toLocaleString() : '-',
                filePath: s.file_path,
                fileSize: `${(s.file_size_bytes / 1024).toFixed(1)} KB`,
                importDuration: '-',
                errorMessage: s.error_message
            }));
            setSources(formatted);
        } catch (err) {
            console.error('Failed to fetch sources:', err);
        }
    };

    // 初期マウント時に取得し、イベントを購読
    useEffect(() => {
        fetchSources();
        const unsubs = [
            Events.EventsOn('dictionary:import_progress', (payload: DictionaryProgressEvent) => {
                const corrId = payload.CorrelationID;
                if (payload.Status === 'COMPLETED' || payload.Status === 'FAILED') {
                    setImportMessages(prev => {
                        const next = { ...prev };
                        delete next[corrId];
                        return next;
                    });
                    setIsImporting(false);
                    fetchSources();
                } else {
                    setImportMessages(prev => ({ ...prev, [corrId]: payload.Message }));
                    setIsImporting(true);
                    if (payload.Completed > 0 && payload.Completed % 1000 === 0) {
                        fetchSources();
                    }
                }
            })
        ];
        return () => {
            unsubs.forEach(u => typeof u === 'function' ? u() : undefined);
        };
    }, []);

    const handleImport = async () => {
        if (selectedFiles.length === 0) return;
        setIsImporting(true);
        for (const filePath of selectedFiles) {
            try {
                const resultId = await DictStartImport(filePath);
                console.log("Started import with ID:", resultId);
            } catch (e) {
                console.error("Import error:", e);
            }
        }
        setSelectedFiles([]);
    };

    // ── ページネーション付きエントリ取得 ────────────────
    const fetchEntriesPaginated = async (idStr: string, page: number, query: string, filters: Record<string, string>) => {
        try {
            const idNum = parseInt(idStr, 10);
            const result = await DictGetEntriesPaginated(idNum, query, filters, page, PAGE_SIZE) as any;
            if (!result) {
                setEntries([]);
                setEntryTotal(0);
                return;
            }
            const rawEntries: any[] = result.entries ?? result.Entries ?? [];
            const total: number = result.totalCount ?? result.TotalCount ?? 0;
            const mapped = rawEntries.map((r: any) => ({
                id: r.id ?? r.ID,
                sourceId: String(r.source_id ?? r.SourceID ?? ''),
                sourceName: r.source_name ?? r.SourceName ?? '',
                edid: r.edid ?? r.EDID ?? '',
                sourceText: r.source_text ?? r.Source ?? '',
                destText: r.dest_text ?? r.Dest ?? '',
                recordType: r.record_type ?? r.RecordType ?? '',
            }));
            setEntries(mapped);
            setEntryTotal(total);
        } catch (e) {
            console.error('fetchEntriesPaginated failed:', e);
        }
    };

    // ページ変更
    const handleEntryPageChange = (page: number) => {
        if (!selectedRowId) return;
        setEntryPage(page);
        fetchEntriesPaginated(selectedRowId, page, entryQuery, entryFilters);
    };

    const handleRowSelectAndFetch = (row: DictSourceRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
        if (rowId) {
            setEntryPage(1);
            setEntryQuery('');
            setEntryFilters({});
            fetchEntriesPaginated(rowId, 1, '', {});
        }
    };

    // ── 横断検索 ─────────────────────────────────────────
    const fetchCrossSearch = async (query: string, filters: Record<string, string>, page: number) => {
        try {
            const result = await DictSearchAllEntriesPaginated(query, filters, page, PAGE_SIZE) as any;
            if (!result) {
                setCrossEntries([]);
                setCrossTotal(0);
                return;
            }
            const rawEntries: any[] = result.entries ?? result.Entries ?? [];
            const total: number = result.totalCount ?? result.TotalCount ?? 0;
            const mapped = rawEntries.map((r: any) => ({
                id: r.id ?? r.ID,
                sourceId: String(r.source_id ?? r.SourceID ?? ''),
                sourceName: r.source_name ?? r.SourceName ?? '',
                edid: r.edid ?? r.EDID ?? '',
                sourceText: r.source_text ?? r.Source ?? '',
                destText: r.dest_text ?? r.Dest ?? '',
                recordType: r.record_type ?? r.RecordType ?? '',
            }));
            setCrossEntries(mapped);
            setCrossTotal(total);
        } catch (e) {
            console.error('fetchCrossSearch failed:', e);
        }
    };

    const handleCrossSearchExecute = (query: string) => {
        setCrossQuery(query);
        setCrossFilters({});
        setCrossPage(1);
        fetchCrossSearch(query, {}, 1);
        setShowCrossSearch(false);
        setView('cross-search');
    };

    const handleCrossPageChange = (page: number) => {
        setCrossPage(page);
        fetchCrossSearch(crossQuery, crossFilters, page);
    };

    const handleSelectFilesClick = async () => {
        try {
            const files = await SelectFiles();
            if (files && files.length > 0) {
                setSelectedFiles(prev => {
                    const currentPaths = new Set(prev);
                    const uniqueNewFiles = files.filter(f => !currentPaths.has(f));
                    return [...prev, ...uniqueNewFiles];
                });
            }
        } catch (e) {
            console.error('Failed to select files:', e);
        }
    };

    const removeSelectedFile = (pathToRemove: string) => {
        setSelectedFiles(prev => prev.filter(p => p !== pathToRemove));
    };

    const handleDeleteSource = async () => {
        if (!deletingRowId) return;
        try {
            await DictDeleteSource(parseInt(deletingRowId, 10));
            await fetchSources();
            if (selectedRowId === deletingRowId) {
                handleRowSelectAndFetch(null, null);
            }
        } catch (e) {
            console.error('Failed to delete source:', e);
        } finally {
            setDeletingRowId(null);
        }
    };

    // ── GridEditor の保存ハンドラ (entries ビュー) ────────
    const handleEntriesSave = async (modified: DictEntry[], deleted: DictEntry[]) => {
        for (const e of modified) {
            try {
                await DictUpdateEntry({
                    id: e.id, source_id: parseInt(e.sourceId, 10),
                    edid: e.edid, record_type: e.recordType,
                    source_text: e.sourceText, dest_text: e.destText,
                } as any);
            } catch (err) { console.error('UpdateEntry failed:', err); }
        }
        for (const e of deleted) {
            try {
                await DictDeleteEntry(e.id);
            } catch (err) { console.error('DeleteEntry failed:', err); }
        }
        // 保存後にリフレッシュ
        if (selectedRowId) {
            await fetchEntriesPaginated(selectedRowId, entryPage, entryQuery, entryFilters);
        }
    };

    // ── GridEditor の保存ハンドラ (cross-search ビュー) ───
    const handleCrossSave = async (modified: DictEntry[], deleted: DictEntry[]) => {
        for (const e of modified) {
            try {
                await DictUpdateEntry({
                    id: e.id, source_id: parseInt(e.sourceId, 10),
                    edid: e.edid, record_type: e.recordType,
                    source_text: e.sourceText, dest_text: e.destText,
                } as any);
            } catch (err) { console.error('UpdateEntry failed:', err); }
        }
        for (const e of deleted) {
            try {
                await DictDeleteEntry(e.id);
            } catch (err) { console.error('DeleteEntry failed:', err); }
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
                const s = info.getValue() as SourceStatus;
                return <div className={`badge badge-sm ${STATUS_BADGE[s]}`}>{s}</div>;
            },
        },
        {
            id: 'actions',
            header: 'アクション',
            cell: (info) => (
                <button
                    className="btn btn-ghost btn-xs text-error"
                    disabled={isImporting}
                    onClick={(e) => {
                        e.stopPropagation();
                        setDeletingRowId(info.row.original.id);
                        showModal('delete_modal');
                    }}
                >
                    削除
                </button>
            ),
        },
    ], [isImporting]);


    return {
        view,
        setView,
        selectedRow,
        setSelectedRow,
        selectedRowId,
        setSelectedRowId,
        selectedFiles,
        setSelectedFiles,
        isImporting,
        setIsImporting,
        importMessages,
        setImportMessages,
        deletingRowId,
        setDeletingRowId,
        showCrossSearch,
        setShowCrossSearch,
        sources,
        setSources,
        entries,
        setEntries,
        entryPage,
        setEntryPage,
        entryTotal,
        setEntryTotal,
        entryQuery,
        setEntryQuery,
        entryFilters,
        setEntryFilters,
        crossEntries,
        setCrossEntries,
        crossPage,
        setCrossPage,
        crossTotal,
        setCrossTotal,
        crossQuery,
        setCrossQuery,
        crossFilters,
        setCrossFilters,
        fetchSources,
        handleImport,
        fetchEntriesPaginated,
        handleEntryPageChange,
        handleRowSelectAndFetch,
        fetchCrossSearch,
        handleCrossSearchExecute,
        handleCrossPageChange,
        handleSelectFilesClick,
        removeSelectedFile,
        handleDeleteSource,
        handleEntriesSave,
        handleCrossSave,
        sourceColumns,
        PAGE_SIZE,
        showModal
    };
}
