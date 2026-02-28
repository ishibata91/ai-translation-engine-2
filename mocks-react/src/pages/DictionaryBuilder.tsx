import React, { useState, useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import DataTable from '../components/DataTable';
import DetailPane from '../components/DetailPane';
import GridEditor from '../components/GridEditor';
import type { GridColumnDef } from '../components/GridEditor';

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
    edid: string;
    recordType: string;
    sourceText: string;
    destText: string;
}

// ── モックデータ: dlc_sources ────────────────────────────
const DICT_SOURCES: DictSourceRow[] = [
    {
        id: '0', fileName: 'Skyrim.esm', format: 'SSTXML', entryCount: 850231,
        status: '完了', updatedAt: '2026-02-26 12:00',
        filePath: 'C:\\SkyrimData\\Translations\\Skyrim_Japanese.STRINGS.xml',
        fileSize: '218.4 MB', importDuration: '4m 32s', errorMessage: null,
    },
    {
        id: '1', fileName: 'Update.esm', format: 'SSTXML', entryCount: 10023,
        status: '完了', updatedAt: '2026-02-26 12:01',
        filePath: 'C:\\SkyrimData\\Translations\\Update_Japanese.STRINGS.xml',
        fileSize: '2.1 MB', importDuration: '0m 12s', errorMessage: null,
    },
    {
        id: '2', fileName: 'Dawnguard.esm', format: 'SSTXML', entryCount: 150490,
        status: '完了', updatedAt: '2026-02-26 12:05',
        filePath: 'C:\\SkyrimData\\Translations\\Dawnguard_Japanese.STRINGS.xml',
        fileSize: '38.7 MB', importDuration: '0m 54s', errorMessage: null,
    },
    {
        id: '3', fileName: 'HearthFires.esm', format: 'SSTXML', entryCount: 25102,
        status: '完了', updatedAt: '2026-02-26 12:06',
        filePath: 'C:\\SkyrimData\\Translations\\HearthFires_Japanese.STRINGS.xml',
        fileSize: '6.4 MB', importDuration: '0m 08s', errorMessage: null,
    },
    {
        id: '4', fileName: 'Dragonborn.esm', format: 'SSTXML', entryCount: 204666,
        status: '完了', updatedAt: '2026-02-26 12:10',
        filePath: 'C:\\SkyrimData\\Translations\\Dragonborn_Japanese.STRINGS.xml',
        fileSize: '52.3 MB', importDuration: '1m 18s', errorMessage: null,
    },
];

// ── モックデータ: dlc_dictionary_entries (Skyrim.esm のサンプル) ──
const DICT_ENTRIES: Record<string, DictEntry[]> = {
    '0': [ // Skyrim.esm
        { id: 1, sourceId: '0', edid: 'DialogueUlfric001', recordType: 'INFO:NAM1', sourceText: "What is it? I'm in the middle of something.", destText: '何だ？今、手が離せないんだ。' },
        { id: 2, sourceId: '0', edid: 'DialogueUlfric002', recordType: 'INFO:NAM1', sourceText: 'Victory or Sovngarde!', destText: '勝利か、ソブンガルデか！' },
        { id: 3, sourceId: '0', edid: 'DialogueTullius001', recordType: 'INFO:NAM1', sourceText: "Rikke, get these men moving!", destText: 'リッケ、部下たちを動かせ！' },
        { id: 4, sourceId: '0', edid: 'DialogueTullius002', recordType: 'INFO:NAM1', sourceText: 'For the Empire!', destText: '帝国のために！' },
        { id: 5, sourceId: '0', edid: 'MQNarratorIntro', recordType: 'BOOK:CNAM', sourceText: 'Long ago, when the world was young...', destText: 'その昔、世界がまだ若かりし頃……' },
        { id: 6, sourceId: '0', edid: 'WhiterunCity', recordType: 'CELL:FULL', sourceText: 'Whiterun', destText: 'ホワイトラン' },
        { id: 7, sourceId: '0', edid: 'SkyrimArmor01', recordType: 'ARMO:FULL', sourceText: 'Iron Armor', destText: '鉄の鎧' },
        { id: 8, sourceId: '0', edid: 'SkyrimWeapon01', recordType: 'WEAP:FULL', sourceText: 'Iron Sword', destText: '鉄の剣' },
        { id: 9, sourceId: '0', edid: 'FavorDialogueTalos', recordType: 'INFO:NAM1', sourceText: 'Talos be with you.', destText: 'タロスのお在りを。' },
        { id: 10, sourceId: '0', edid: 'DragonbornNarrator01', recordType: 'BOOK:CNAM', sourceText: 'In the dawn of time, the dragons ruled...', destText: '時の始まり、ドラゴンたちが支配していた……' },
    ],
    '1': [ // Update.esm
        { id: 101, sourceId: '1', edid: 'UpdatePatch001', recordType: 'INFO:NAM1', sourceText: 'The road to Helgen is closed.', destText: 'ヘルゲンへの道は封鎖されている。' },
        { id: 102, sourceId: '1', edid: 'UpdatePatch002', recordType: 'NPC_:FULL', sourceText: 'Guard', destText: '衛兵' },
    ],
    '2': [ // Dawnguard.esm
        { id: 201, sourceId: '2', edid: 'DLC01SeranaMeet', recordType: 'INFO:NAM1', sourceText: "Don't come any closer!", destText: 'これ以上近づかないで！' },
        { id: 202, sourceId: '2', edid: 'DLC01VampireLord', recordType: 'SPEL:FULL', sourceText: 'Vampire Lord', destText: 'ヴァンパイアロード' },
        { id: 203, sourceId: '2', edid: 'DLC01CastleVolkihar', recordType: 'CELL:FULL', sourceText: 'Castle Volkihar', destText: 'ヴォルキハル城' },
    ],
    '3': [], // HearthFires.esm
    '4': [ // Dragonborn.esm
        { id: 401, sourceId: '4', edid: 'DLC02MiraakShout', recordType: 'INFO:NAM1', sourceText: 'Miraak!', destText: 'ミラーク！' },
        { id: 402, sourceId: '4', edid: 'DLC02SolstheimCell', recordType: 'CELL:FULL', sourceText: 'Solstheim', destText: 'ソルスセイム' },
    ],
};

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

// ── ソーステーブル列定義 ─────────────────────────────────
const showModal = (id: string) => {
    const modal = document.getElementById(id) as HTMLDialogElement;
    modal?.showModal();
};

const SOURCE_COLUMNS: ColumnDef<DictSourceRow, unknown>[] = [
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
        cell: () => (
            <button
                className="btn btn-ghost btn-xs text-error"
                onClick={(e) => { e.stopPropagation(); showModal('delete_modal'); }}
            >
                削除
            </button>
        ),
    },
];

// ── ビュー型 ─────────────────────────────────────────────
type View = 'list' | 'entries';

// ── ページコンポーネント ──────────────────────────────────
const DictionaryBuilder: React.FC = () => {
    const [view, setView] = useState<View>('list');
    const [selectedRow, setSelectedRow] = useState<DictSourceRow | null>(null);
    const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
    const [isImporting, setIsImporting] = useState<boolean>(true);

    const handleRowSelect = (row: DictSourceRow | null, rowId: string | null) => {
        setSelectedRow(row);
        setSelectedRowId(rowId);
    };

    const tableHeaderActions = useMemo(() => (
        <button className="btn btn-outline btn-error btn-sm" onClick={() => showModal('delete_all_modal')}>
            全て削除
        </button>
    ), []);

    // 選択ソースのエントリデータ
    const currentEntries: DictEntry[] = selectedRow
        ? (DICT_ENTRIES[selectedRow.id] ?? [])
        : [];

    // ── entries ビュー ────────────────────────────────────
    if (view === 'entries' && selectedRow) {
        let nextEntryId = 900; // モック用の新規ID採番
        return (
            <GridEditor<DictEntry>
                title={`エントリ編集: ${selectedRow.fileName} (${currentEntries.length.toLocaleString()} 件表示中)`}
                initialData={currentEntries}
                columns={ENTRY_COLUMNS}
                onBack={() => setView('list')}
                onSave={(rows) => {
                    console.log('[DictionaryBuilder] エントリ保存:', rows);
                    setView('list');
                }}
                newRowFactory={() => ({
                    id: nextEntryId++,
                    sourceId: selectedRow.id,
                    edid: '',
                    recordType: '',
                    sourceText: '',
                    destText: '',
                })}
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
                </div>
            </div>

            {/* 画面説明 */}
            <div className="alert alert-info shadow-sm shrink-0 flex-col items-start gap-2">
                <div className="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                    <h3 className="font-bold">システム辞書の構築について</h3>
                </div>
                <div className="text-sm space-y-2">
                    <p>
                        この画面では、公式翻訳や過去の翻訳済みModのデータ（SSTXML形式など）をインポートし、
                        <strong>全プロジェクト共通で利用される「システム辞書(dictionary.db)」</strong>を構築・管理します。
                    </p>
                    <ul className="list-disc list-inside ml-2">
                        <li>ソース行をクリックして選択し、<strong>「エントリを表示・編集」</strong>からインライン編集が行えます。</li>
                        <li><code className="bg-base-200 px-1 rounded">Skyrim.esm</code> などの公式マスターファイルを優先してインポートすることを推奨します。</li>
                    </ul>
                </div>
            </div>

            <div className="flex flex-1 flex-col min-h-0 gap-4 relative">
                {/* 上部パネル */}
                <div className="grid grid-cols-2 gap-4 shrink-0">
                    <div className="card bg-base-100 shadow-sm border border-base-200">
                        <div className="card-body">
                            <h2 className="card-title text-base">XMLインポート (xTranslator形式)</h2>
                            <div className="flex flex-col gap-4 mt-2">
                                <span className="text-sm">SSTXMLファイル、または公式DLCの翻訳XMLを選択してください。</span>
                                <div className="flex gap-4">
                                    <input type="file" className="file-input file-input-bordered file-input-primary w-full max-w-xs" />
                                </div>
                                <div>
                                    <span className="text-sm font-bold block mb-2">インポート進捗</span>
                                    <progress className="progress progress-primary w-full" value="0" max="100"></progress>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div className="card bg-base-100 shadow-sm border border-base-200">
                        <div className="card-body">
                            <h2 className="card-title text-base">システム辞書ステータス</h2>
                            <div className="flex flex-col gap-4 mt-2">
                                <div className="stat px-0 py-2">
                                    <div className="stat-title text-sm">総エントリ数</div>
                                    <div className="stat-value text-primary text-3xl font-mono">1,240,512</div>
                                </div>
                                <div className="stat px-0 py-2 border-t border-base-200">
                                    <div className="stat-title text-sm">登録済みソース</div>
                                    <div className="stat-value text-xl">5</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* ソーステーブル */}
                <div className="flex-1 min-h-0 flex flex-col">
                    <DataTable
                        columns={SOURCE_COLUMNS}
                        data={DICT_SOURCES}
                        title="登録済み辞書ソース一覧"
                        selectedRowId={selectedRowId}
                        onRowSelect={handleRowSelect}
                        headerActions={tableHeaderActions}
                    />
                </div>

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

            {/* 下部ステータスバー */}
            <div className="flex justify-between items-center bg-base-200 p-2 rounded-xl border shrink-0">
                <span className="text-sm font-bold text-gray-500 ml-2">Job: DictionaryImport ({isImporting ? 'Running' : 'Stopped'})</span>
                <div className="flex gap-2">
                    <button
                        className={`btn btn-sm ${isImporting ? 'btn-ghost' : 'btn-outline'}`}
                        onClick={() => setIsImporting(!isImporting)}
                    >
                        {isImporting ? '一時停止' : '再開 (デモ)'}
                    </button>
                </div>
            </div>

            {/* 詳細ペイン */}
            <DetailPane
                isOpen={!!selectedRow}
                onClose={() => handleRowSelect(null, null)}
                title={selectedRow ? `詳細: ${selectedRow.fileName} (${selectedRow.format})` : '詳細'}
                defaultHeight={280}
            >
                {selectedRow && (
                    <div className="flex flex-col gap-4 text-sm">
                        {/* アクションボタン群 */}
                        <div className="flex gap-2 shrink-0">
                            <button
                                className="btn btn-primary btn-sm"
                                onClick={() => setView('entries')}
                            >
                                📋 エントリを表示・編集
                            </button>
                        </div>

                        {/* プレビューグリッド */}
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
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-error">削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            <dialog id="delete_all_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">全ソース削除の確認</h3>
                    <p className="py-4">登録されている全ての辞書ソースをデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost">キャンセル</button>
                                <button className="btn btn-error">全て削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>
        </div>
    );
};

export default DictionaryBuilder;
