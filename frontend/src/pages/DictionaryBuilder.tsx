import React, { useState, useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import DataTable from '../components/DataTable';
import DetailPane from '../components/dictionary/DetailPane';
import GridEditor from '../components/dictionary/GridEditor';
import type { GridColumnDef } from '../components/dictionary/GridEditor';
import { HelpCircle } from 'lucide-react';
import CrossSearchModal from '../components/dictionary/CrossSearchModal';

// ── 型定義: dlc_sources ──────────────────────────────────
type SourceStatus = '完了' | 'インポート中' | 'エラー';

interface DictSourceRow {
    id: string;
    fileName: string;
    format: string;
    entryCount: number;
    status: SourceStatus;
    updatedAt: string;
    filePath: string;
    fileSize: string;
    importDuration: string;
    errorMessage: string | null;
}

// ── 型定義: dlc_dictionary_entries ──────────────────────
interface DictEntry {
    id: number;
    sourceId: string;
    sourceName?: string;
    edid: string;
    recordType: string;
    sourceText: string;
    destText: string;
}

// ── ステータスバッジ ──────────────────────────────────────
const STATUS_BADGE: Record<SourceStatus, string> = {
    '完了': 'badge-success',
    'インポート中': 'badge-info',
    'エラー': 'badge-error',
};

// ── GridEditor 用列定義 (dlc_dictionary_entries) ─────────
const ENTRY_COLUMNS: GridColumnDef<DictEntry>[] = [
    { key: 'id', header: 'ID', editable: false, widthClass: 'w-16', type: 'number' },
    { key: 'edid', header: 'Editor ID', editable: true, widthClass: 'w-48' },
    { key: 'recordType', header: 'Record Type', editable: true, widthClass: 'w-32' },
    { key: 'sourceText', header: '原文 (英語)', editable: true, widthClass: 'w-80' },
    { key: 'destText', header: '訳文 (日本語)', editable: true, widthClass: 'w-80' },
];

// ── 横断検索用列定義 (sourceName列付き) ──────────────────
const CROSS_ENTRY_COLUMNS: GridColumnDef<DictEntry>[] = [
    { key: 'sourceName', header: '辞書ソース', editable: false, widthClass: 'w-40' },
    { key: 'id', header: 'ID', editable: false, widthClass: 'w-16', type: 'number' },
    { key: 'edid', header: 'Editor ID', editable: true, widthClass: 'w-48' },
    { key: 'recordType', header: 'Record Type', editable: true, widthClass: 'w-32' },
    { key: 'sourceText', header: '原文 (英語)', editable: true, widthClass: 'w-80' },
    { key: 'destText', header: '訳文 (日本語)', editable: true, widthClass: 'w-80' },
];

// ── ビュー型 ─────────────────────────────────────────────
type View = 'list' | 'entries' | 'cross-search';

const showModal = (id: string) => {
    const modal = document.getElementById(id) as HTMLDialogElement;
    modal?.showModal();
};

const PAGE_SIZE = 500;

import {
    DictGetSources,
    DictStartImport,
    DictGetEntriesPaginated,
    DictSearchAllEntriesPaginated,
    DictUpdateEntry,
    DictDeleteEntry,
    SelectFiles,
    DictDeleteSource,
} from '../wailsjs/go/main/App';
import * as Events from '../wailsjs/runtime/runtime';

const DictionaryBuilder: React.FC = () => {
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
    React.useEffect(() => {
        fetchSources();
        const unsubs = [
            Events.EventsOn('dictionary:import_progress', (payload: any) => {
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

    // ── entries ビュー ────────────────────────────────────
    if (view === 'entries' && selectedRow) {
        return (
            <GridEditor<DictEntry>
                title={`エントリ編集: ${selectedRow.fileName}`}
                initialData={entries}
                columns={ENTRY_COLUMNS}
                onBack={() => setView('list')}
                onSave={handleEntriesSave}
                currentPage={entryPage}
                totalCount={entryTotal}
                pageSize={PAGE_SIZE}
                onPageChange={handleEntryPageChange}
                onSearch={(filters) => {
                    setEntryFilters(filters);
                    setEntryPage(1);
                    if (selectedRowId) {
                        fetchEntriesPaginated(selectedRowId, 1, entryQuery, filters);
                    }
                }}
            />
        );
    }

    // ── cross-search ビュー ────────────────────────────────
    if (view === 'cross-search') {
        return (
            <GridEditor<DictEntry>
                title={`横断検索結果: "${crossQuery}" (${crossTotal.toLocaleString()} 件)`}
                initialData={crossEntries}
                columns={CROSS_ENTRY_COLUMNS}
                onBack={() => setView('list')}
                onSave={handleCrossSave}
                currentPage={crossPage}
                totalCount={crossTotal}
                pageSize={PAGE_SIZE}
                onPageChange={handleCrossPageChange}
                onSearch={(filters) => {
                    setCrossFilters(filters);
                    setCrossPage(1);
                    fetchCrossSearch(crossQuery, filters, 1);
                }}
            />
        );
    }

    // ── list ビュー ───────────────────────────────────────
    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            {/* ヘッダー */}
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">辞書構築 (Dictionary Builder)</span>
                    <div className="flex items-center gap-2">
                        <div className="tooltip tooltip-left" data-tip="登録済み辞書ソースを横断して検索出来ます。">
                            <HelpCircle size={18} className="text-base-content/40 cursor-help hover:text-primary transition-colors" />
                        </div>
                        {/* 横断検索ボタン (Task 3.4) */}
                        <button
                            className="btn btn-outline btn-sm gap-2"
                            onClick={() => setShowCrossSearch(true)}
                        >
                            🔎 横断検索
                        </button>
                    </div>
                </div>
            </div>

            {/* 画面説明 */}
            <details className="alert alert-info shadow-sm shrink-0 flex-col items-start gap-2 [&>summary::-webkit-details-marker]:hidden">
                <summary className="flex items-center gap-2 cursor-pointer font-bold select-none list-none">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                    システム辞書の構築について (クリックで展開)
                </summary>
                <div className="text-sm space-y-2 mt-2 pt-2 border-t border-info-content/20">
                    <p>
                        この画面では、公式翻訳や過去の翻訳済みModのデータ（SSTXML形式など）をインポートし、
                        <strong>全プロジェクト共通で利用される「システム辞書(dictionary.db)」</strong>を構築・管理します。
                    </p>
                    <ul className="list-disc list-inside ml-2">
                        <li>ソース行をクリックして選択し、<strong>「エントリを表示・編集」</strong>からインライン編集が行えます。</li>
                        <li><code className="bg-base-100 text-base-content px-1 rounded">Skyrim.esm</code> などの公式マスターファイルを優先してインポートすることを推奨します。</li>
                    </ul>
                </div>
            </details>

            <div className="flex flex-1 flex-col min-h-0 gap-4 relative">
                {/* XMLインポートパネル (Task 3.1: コンパクト化) */}
                <div className="shrink-0">
                    <div className="card bg-base-100 shadow-sm border border-base-200">
                        <div className="card-body py-3 px-4">
                            <h2 className="card-title text-base">XMLインポート (xTranslator形式)</h2>
                            <div className="flex flex-col gap-3">
                                <div className="flex items-center gap-3 flex-wrap">
                                    <span className="text-sm text-base-content/70">SSTMLファイル、または公式翻訳XMLを選択してください。</span>
                                    <button
                                        className="btn btn-outline btn-primary btn-sm w-fit"
                                        onClick={handleSelectFilesClick}
                                        disabled={isImporting}
                                    >
                                        ファイルを選択
                                    </button>
                                    <button
                                        className="btn btn-primary btn-sm"
                                        disabled={selectedFiles.length === 0 || isImporting}
                                        onClick={() => {
                                            if (selectedFiles.length === 0) return;
                                            handleRowSelectAndFetch(null, null);
                                            handleImport();
                                        }}
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 mr-1">
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
                                        </svg>
                                        {isImporting ? 'インポート実行中...' : '辞書構築を開始'}
                                    </button>
                                </div>

                                {/* 選択ファイル一覧 & 進捗: ファイル選択時のみ高さが拡張 */}
                                {(selectedFiles.length > 0 || (isImporting && Object.keys(importMessages).length > 0)) && (
                                    <div className="flex flex-col gap-2 transition-all">
                                        {selectedFiles.length > 0 && (
                                            <div className="flex flex-col gap-1">
                                                <span className="text-xs font-bold text-base-content/70">選択ファイル ({selectedFiles.length}件):</span>
                                                <div className="flex flex-wrap gap-2 max-h-24 overflow-y-auto p-2 bg-base-200/50 rounded-lg border border-base-300">
                                                    {selectedFiles.map(filePath => {
                                                        const fileName = filePath.split(/[\\\/]/).pop() || filePath;
                                                        return (
                                                            <div key={filePath} className="badge badge-primary badge-outline gap-1 py-3 px-2">
                                                                <span className="truncate max-w-[200px] font-mono text-xs" title={filePath}>{fileName}</span>
                                                                <button
                                                                    className="btn btn-ghost btn-xs btn-circle ml-1 opacity-70 hover:opacity-100"
                                                                    disabled={isImporting}
                                                                    onClick={() => removeSelectedFile(filePath)}
                                                                    title="リストから外す"
                                                                >✕</button>
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        )}
                                        {isImporting && Object.keys(importMessages).length > 0 && (
                                            <div className="flex flex-col gap-2">
                                                <span className="text-xs font-bold block border-b border-base-200 pb-1">インポート進捗</span>
                                                {Object.entries(importMessages).map(([corrId, msg]) => (
                                                    <div key={corrId} className="flex flex-col gap-1">
                                                        <div className="flex justify-between text-xs">
                                                            <span className="truncate max-w-full text-primary" title={msg}>{msg}</span>
                                                        </div>
                                                        <progress className="progress progress-primary w-full"></progress>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* ソーステーブル */}
                <div className="flex-1 min-h-0 flex flex-col relative">
                    <DataTable
                        columns={sourceColumns}
                        data={sources}
                        title="登録済み辞書ソース一覧"
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelectAndFetch}
                    />

                    {isImporting && (
                        <div className="absolute inset-0 bg-base-100/50 backdrop-blur-[1px] z-10 flex flex-col items-center justify-center gap-4 rounded-xl border border-base-200">
                            <span className="loading loading-spinner text-primary loading-lg"></span>
                            <div className="flex flex-col items-center gap-1">
                                <span className="font-bold text-lg text-base-content/70">XML辞書データをインポート中...</span>
                                <span className="text-sm text-base-content/50">ファイルの解析とデータベースへのマージを行っています</span>
                            </div>
                        </div>
                    )}
                </div>

            </div>

            {/* 詳細ペイン */}
            <DetailPane
                isOpen={!!selectedRow}
                onClose={() => handleRowSelectAndFetch(null, null)}
                title={selectedRow ? `詳細: ${selectedRow.fileName} (${selectedRow.format})` : '詳細'}
                defaultHeight={280}
            >
                {selectedRow && (
                    <div className="flex flex-col gap-4 text-sm">
                        <div className="flex gap-2 shrink-0">
                            <button
                                className="btn btn-primary btn-sm"
                                onClick={() => setView('entries')}
                            >
                                📋 エントリを表示・編集
                            </button>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイル名</span>
                                <span className="font-mono">{selectedRow.fileName}</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">形式</span>
                                <div className="badge badge-outline badge-sm font-mono w-fit">{selectedRow.format}</div>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ステータス</span>
                                <div className={`badge badge-sm w-fit ${STATUS_BADGE[selectedRow.status]}`}>{selectedRow.status}</div>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">エントリ数</span>
                                <span className="font-mono">{selectedRow.entryCount.toLocaleString()} 件</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">最終更新日時</span>
                                <span>{selectedRow.updatedAt}</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイルサイズ</span>
                                <span className="font-mono">{selectedRow.fileSize}</span>
                            </div>
                        </div>

                        <div className="flex flex-col gap-1">
                            <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイルパス</span>
                            <div className="bg-base-200 rounded px-3 py-2 font-mono text-xs break-all">
                                {selectedRow.filePath}
                            </div>
                        </div>
                    </div>
                )}
            </DetailPane>

            {/* 削除確認モーダル */}
            <dialog id="delete_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">削除の確認</h3>
                    <p className="py-4">このソースをデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost" onClick={() => setDeletingRowId(null)}>キャンセル</button>
                                <button className="btn btn-error" onClick={handleDeleteSource}>削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {/* 横断検索モーダル */}
            {showCrossSearch && (
                <CrossSearchModal
                    sources={sources}
                    onSearch={handleCrossSearchExecute}
                    onClose={() => setShowCrossSearch(false)}
                />
            )}
        </div>
    );
};

export default DictionaryBuilder;
