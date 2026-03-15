import type {ColumnDef} from '@tanstack/react-table';
import DataTable from '../DataTable';
import type {LoadedTranslationFile, TranslationTargetRow} from '../../hooks/features/translationFlow/types';

interface LoadPanelProps {
    isActive: boolean;
    selectedFiles: string[];
    loadedFiles: LoadedTranslationFile[];
    isLoading: boolean;
    errorMessage: string;
    onSelectFiles: () => Promise<void>;
    onRemoveFile: (pathToRemove: string) => void;
    onLoadSelectedFiles: () => Promise<void>;
    onReloadFiles: () => Promise<void>;
    onPreviewPageChange: (fileId: number, page: number) => Promise<void>;
    onNext: () => void;
}

const PREVIEW_COLUMNS: ColumnDef<TranslationTargetRow, unknown>[] = [
    {
        accessorKey: 'section',
        header: 'Section',
    },
    {
        accessorKey: 'recordType',
        header: 'Record Type',
        cell: (info) => <span className="font-mono text-xs">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'editorId',
        header: 'Editor ID',
        cell: (info) => <span className="font-mono text-xs">{String(info.getValue() ?? '')}</span>,
    },
    {
        accessorKey: 'sourceText',
        header: 'Source Text',
    },
];

const fileLabel = (path: string): string => {
    const parts = path.split(/[\\/]/);
    const last = parts[parts.length - 1];
    return last || path;
};

/**
 * 翻訳フローのデータロードフェーズ UI。
 */
export function LoadPanel({
    isActive,
    selectedFiles,
    loadedFiles,
    isLoading,
    errorMessage,
    onSelectFiles,
    onRemoveFile,
    onLoadSelectedFiles,
    onReloadFiles,
    onPreviewPageChange,
    onNext,
}: LoadPanelProps) {
    const loadedPathSet = new Set(loadedFiles.map((file) => file.filePath));

    return (
        <div className={`tab-content-panel flex-col gap-4 ${isActive ? 'flex' : 'hidden'}`}>
            <div className="alert alert-info shadow-sm shrink-0">
                <span>抽出済み JSON を複数選択し、artifact に保存した翻訳対象をファイル単位で確認します。</span>
            </div>

            <div className="card bg-base-100 border border-base-200 shadow-sm shrink-0">
                <div className="card-body gap-4">
                    <h2 className="card-title text-base">データロード</h2>
                    <div className="flex flex-wrap gap-2">
                        <button type="button" className="btn btn-outline btn-primary btn-sm" onClick={() => void onSelectFiles()} disabled={isLoading}>
                            ファイルを選択
                        </button>
                        <button
                            type="button"
                            className="btn btn-primary btn-sm"
                            onClick={() => void onLoadSelectedFiles()}
                            disabled={isLoading || selectedFiles.length === 0}
                        >
                            {isLoading ? 'ロード中...' : 'ロード実行'}
                        </button>
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => void onReloadFiles()} disabled={isLoading}>
                            再読込
                        </button>
                    </div>

                    <div className="flex flex-col gap-2">
                        <span className="text-sm font-bold text-base-content/70">選択済みファイル ({selectedFiles.length} 件)</span>
                        <div className="flex flex-wrap gap-2">
                            {selectedFiles.length === 0 && <span className="text-sm text-base-content/50">まだファイルが選択されていません</span>}
                            {selectedFiles.map((path) => {
                                const loaded = loadedPathSet.has(path);
                                return (
                                    <div key={path} className={`badge gap-2 py-3 px-3 ${loaded ? 'badge-success badge-outline' : 'badge-primary badge-outline'}`}>
                                        <span className="font-mono text-xs truncate max-w-[260px]" title={path}>{fileLabel(path)}</span>
                                        <span className="text-[10px]">{loaded ? 'ロード済み' : '未ロード'}</span>
                                        {!loaded && (
                                            <button
                                                type="button"
                                                className="btn btn-ghost btn-xs btn-circle"
                                                onClick={() => onRemoveFile(path)}
                                                disabled={isLoading}
                                            >
                                                ✕
                                            </button>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    {errorMessage !== '' && <p className="text-error text-sm">{errorMessage}</p>}
                </div>
            </div>

            <div className="flex flex-col gap-3">
                <span className="text-sm font-bold text-base-content/70">ロード済みプレビュー ({loadedFiles.length} ファイル)</span>
                {loadedFiles.length === 0 && (
                    <div className="p-4 border border-dashed border-base-300 rounded-xl text-sm text-base-content/60">
                        ロード済みファイルはありません。
                    </div>
                )}
                {loadedFiles.map((file, index) => {
                    const totalPages = Math.max(1, Math.ceil(file.totalRows / Math.max(1, file.pageSize)));
                    return (
                        <DataTable
                            key={file.fileId}
                            columns={PREVIEW_COLUMNS}
                            data={file.rows}
                            title={`${file.fileName} (${file.rowCount} 件)`}
                            collapsible
                            defaultOpen={index === 0}
                            headerActions={
                                <div className="flex flex-wrap items-center gap-2 pr-3 xl:pr-0 text-xs">
                                    <span>{file.currentPage} / {totalPages} ページ</span>
                                    <button
                                        type="button"
                                        className="btn btn-outline btn-xs"
                                        onClick={() => void onPreviewPageChange(file.fileId, Math.max(1, file.currentPage - 1))}
                                        disabled={isLoading || file.currentPage <= 1}
                                    >
                                        前へ
                                    </button>
                                    <button
                                        type="button"
                                        className="btn btn-outline btn-xs"
                                        onClick={() => void onPreviewPageChange(file.fileId, Math.min(totalPages, file.currentPage + 1))}
                                        disabled={isLoading || file.currentPage >= totalPages}
                                    >
                                        次へ
                                    </button>
                                </div>
                            }
                        />
                    );
                })}
            </div>

            <div className="flex justify-end bg-base-200 p-2 rounded-xl border shrink-0">
                <button type="button" className="btn btn-primary btn-sm" onClick={onNext} disabled={loadedFiles.length === 0 || isLoading}>
                    ロード完了して次へ
                </button>
            </div>
        </div>
    );
}
