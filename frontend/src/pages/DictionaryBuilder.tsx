import { HelpCircle } from 'lucide-react';
import DataTable from '../components/DataTable';
import CrossSearchModal from '../components/dictionary/CrossSearchModal';
import DetailPane from '../components/dictionary/DetailPane';
import GridEditor from '../components/dictionary/GridEditor';
import type { GridColumnDef } from '../components/dictionary/GridEditor';
import { useDictionaryBuilder } from '../hooks/features/dictionaryBuilder/useDictionaryBuilder';
import { STATUS_BADGE } from '../hooks/features/dictionaryBuilder/types';
import type { DictEntry } from '../hooks/features/dictionaryBuilder/types';

const ENTRY_COLUMNS: GridColumnDef<DictEntry>[] = [
    { key: 'id', header: 'ID', editable: false, widthClass: 'w-16', type: 'number' },
    { key: 'edid', header: 'Editor ID', editable: true, widthClass: 'w-48' },
    { key: 'recordType', header: 'Record Type', editable: true, widthClass: 'w-32' },
    { key: 'sourceText', header: '原文 (英語)', editable: true, widthClass: 'w-80' },
    { key: 'destText', header: '訳文 (日本語)', editable: true, widthClass: 'w-80' },
];

const CROSS_ENTRY_COLUMNS: GridColumnDef<DictEntry>[] = [
    { key: 'sourceName', header: '辞書ソース', editable: false, widthClass: 'w-40' },
    { key: 'id', header: 'ID', editable: false, widthClass: 'w-16', type: 'number' },
    { key: 'edid', header: 'Editor ID', editable: true, widthClass: 'w-48' },
    { key: 'recordType', header: 'Record Type', editable: true, widthClass: 'w-32' },
    { key: 'sourceText', header: '原文 (英語)', editable: true, widthClass: 'w-80' },
    { key: 'destText', header: '訳文 (日本語)', editable: true, widthClass: 'w-80' },
];

/**
 * 辞書構築の一覧、インポート、編集画面を描画する。
 */
export default function DictionaryBuilder() {
    const { state, actions, ui, constants } = useDictionaryBuilder();

    if (state.view === 'entries' && state.selectedRow) {
        return (
            <GridEditor<DictEntry>
                title={`エントリ編集: ${state.selectedRow.fileName}`}
                initialData={state.entries}
                columns={ENTRY_COLUMNS}
                onBack={() => actions.setView('list')}
                onSave={actions.handleEntriesSave}
                currentPage={state.entryPage}
                totalCount={state.entryTotal}
                pageSize={constants.pageSize}
                onPageChange={actions.handleEntryPageChange}
                onSearch={actions.handleEntrySearch}
            />
        );
    }

    if (state.view === 'cross-search') {
        return (
            <GridEditor<DictEntry>
                title={`横断検索結果: "${state.crossQuery}" (${state.crossTotal.toLocaleString()} 件)`}
                initialData={state.crossEntries}
                columns={CROSS_ENTRY_COLUMNS}
                onBack={() => actions.setView('list')}
                onSave={actions.handleCrossSave}
                currentPage={state.crossPage}
                totalCount={state.crossTotal}
                pageSize={constants.pageSize}
                onPageChange={actions.handleCrossPageChange}
                onSearch={actions.handleCrossSearchFilter}
            />
        );
    }

    return (
        <div className="flex flex-col w-full h-full p-4 gap-4">
            <div className="navbar bg-base-100 rounded-box border border-base-200 shadow-sm px-4 shrink-0">
                <div className="flex justify-between items-center w-full">
                    <span className="text-xl font-bold">辞書構築 (Dictionary Builder)</span>
                    <div className="flex items-center gap-2">
                        <div className="tooltip tooltip-left" data-tip="登録済み辞書ソースを横断して検索出来ます。">
                            <HelpCircle size={18} className="text-base-content/40 cursor-help hover:text-primary transition-colors" />
                        </div>
                        <button className="btn btn-outline btn-sm gap-2" onClick={actions.openCrossSearch}>
                            🔎 横断検索
                        </button>
                    </div>
                </div>
            </div>

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
                <div className="shrink-0">
                    <div className="card bg-base-100 shadow-sm border border-base-200">
                        <div className="card-body py-3 px-4">
                            <h2 className="card-title text-base">XMLインポート (xTranslator形式)</h2>
                            <div className="flex flex-col gap-3">
                                <div className="flex items-center gap-3 flex-wrap">
                                    <span className="text-sm text-base-content/70">SSTMLファイル、または公式翻訳XMLを選択してください。</span>
                                    <button
                                        className="btn btn-outline btn-primary btn-sm w-fit"
                                        onClick={actions.handleSelectFilesClick}
                                        disabled={state.isImporting}
                                    >
                                        ファイルを選択
                                    </button>
                                    <button
                                        className="btn btn-primary btn-sm"
                                        disabled={state.selectedFiles.length === 0 || state.isImporting}
                                        onClick={() => {
                                            if (state.selectedFiles.length === 0) {
                                                return;
                                            }
                                            actions.handleRowSelectAndFetch(null, null);
                                            void actions.handleImport();
                                        }}
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 mr-1">
                                            <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
                                        </svg>
                                        {state.isImporting ? 'インポート実行中...' : '辞書構築を開始'}
                                    </button>
                                </div>

                                {(state.selectedFiles.length > 0 || (state.isImporting && Object.keys(state.importMessages).length > 0)) && (
                                    <div className="flex flex-col gap-2 transition-all">
                                        {state.selectedFiles.length > 0 && (
                                            <div className="flex flex-col gap-1">
                                                <span className="text-xs font-bold text-base-content/70">選択ファイル ({state.selectedFiles.length}件):</span>
                                                <div className="flex flex-wrap gap-2 max-h-24 overflow-y-auto p-2 bg-base-200/50 rounded-lg border border-base-300">
                                                    {state.selectedFiles.map((filePath) => {
                                                        const fileName = filePath.split(/[\\/]/).pop() || filePath;
                                                        return (
                                                            <div key={filePath} className="badge badge-primary badge-outline gap-1 py-3 px-2">
                                                                <span className="truncate max-w-[200px] font-mono text-xs" title={filePath}>{fileName}</span>
                                                                <button
                                                                    className="btn btn-ghost btn-xs btn-circle ml-1 opacity-70 hover:opacity-100"
                                                                    disabled={state.isImporting}
                                                                    onClick={() => actions.removeSelectedFile(filePath)}
                                                                    title="リストから外す"
                                                                >✕</button>
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        )}
                                        {state.isImporting && Object.keys(state.importMessages).length > 0 && (
                                            <div className="flex flex-col gap-2">
                                                <span className="text-xs font-bold block border-b border-base-200 pb-1">インポート進捗</span>
                                                {Object.entries(state.importMessages).map(([corrId, msg]) => (
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

                <div className="flex-1 min-h-0 flex flex-col relative">
                    <DataTable
                        columns={ui.sourceColumns}
                        data={state.sources}
                        title="登録済み辞書ソース一覧"
                        selectedRowId={state.selectedRowId}
                        onRowSelect={actions.handleRowSelectAndFetch}
                    />

                    {state.isImporting && (
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

            <DetailPane
                isOpen={!!state.selectedRow}
                onClose={() => actions.handleRowSelectAndFetch(null, null)}
                title={state.selectedRow ? `詳細: ${state.selectedRow.fileName} (${state.selectedRow.format})` : '詳細'}
                defaultHeight={280}
            >
                {state.selectedRow && (
                    <div className="flex flex-col gap-4 text-sm">
                        <div className="flex gap-2 shrink-0">
                            <button className="btn btn-primary btn-sm" onClick={() => actions.setView('entries')}>
                                📋 エントリを表示・編集
                            </button>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイル名</span>
                                <span className="font-mono">{state.selectedRow.fileName}</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">形式</span>
                                <div className="badge badge-outline badge-sm font-mono w-fit">{state.selectedRow.format}</div>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ステータス</span>
                                <div className={`badge badge-sm w-fit ${STATUS_BADGE[state.selectedRow.status]}`}>{state.selectedRow.status}</div>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">エントリ数</span>
                                <span className="font-mono">{state.selectedRow.entryCount.toLocaleString()} 件</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">最終更新日時</span>
                                <span>{state.selectedRow.updatedAt}</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイルサイズ</span>
                                <span className="font-mono">{state.selectedRow.fileSize}</span>
                            </div>
                        </div>

                        <div className="flex flex-col gap-1">
                            <span className="font-bold text-base-content/60 text-xs uppercase tracking-wide">ファイルパス</span>
                            <div className="bg-base-200 rounded px-3 py-2 font-mono text-xs break-all">
                                {state.selectedRow.filePath}
                            </div>
                        </div>
                    </div>
                )}
            </DetailPane>

            <dialog id="delete_modal" className="modal">
                <div className="modal-box border border-error">
                    <h3 className="font-bold text-lg text-error">削除の確認</h3>
                    <p className="py-4">このソースをデータベースから削除しますか？<br />※この操作は取り消せません。</p>
                    <div className="modal-action">
                        <form method="dialog">
                            <div className="flex gap-2">
                                <button className="btn btn-ghost" onClick={actions.handleCancelDelete}>キャンセル</button>
                                <button className="btn btn-error" onClick={() => void actions.handleDeleteSource()}>削除する</button>
                            </div>
                        </form>
                    </div>
                </div>
            </dialog>

            {state.showCrossSearch && (
                <CrossSearchModal
                    sources={state.sources}
                    onSearch={actions.handleCrossSearchExecute}
                    onClose={actions.closeCrossSearch}
                />
            )}
        </div>
    );
}
