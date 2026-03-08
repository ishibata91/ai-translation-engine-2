/**
 * 辞書ソースの取り込み状態を表す。
 */
export type SourceStatus = '完了' | 'インポート中' | 'エラー';

/**
 * 辞書ソース一覧に表示する 1 行分のデータ。
 */
export interface DictSourceRow {
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

/**
 * 辞書エントリ編集に用いる行データ。
 */
export interface DictEntry {
    id: number;
    sourceId: string;
    sourceName?: string;
    edid: string;
    recordType: string;
    sourceText: string;
    destText: string;
}

/**
 * ページネーション済み辞書エントリの結果。
 */
export interface DictEntryPage {
    entries: DictEntry[];
    totalCount: number;
}

/**
 * 辞書エントリ更新 API に渡す payload。
 */
export interface DictUpdateEntryPayload {
    id: number;
    source_id: number;
    edid: string;
    record_type: string;
    source_text: string;
    dest_text: string;
}

export const STATUS_BADGE: Record<SourceStatus, string> = {
    '完了': 'badge-success',
    'インポート中': 'badge-info',
    'エラー': 'badge-error',
};

/**
 * Dictionary Builder の画面モード。
 */
export type View = 'list' | 'entries' | 'cross-search';

/**
 * 辞書インポート進捗イベントの payload。
 */
export interface DictionaryProgressEvent {
    CorrelationID: string;
    Status: 'STARTED' | 'COMPLETED' | 'FAILED' | 'IN_PROGRESS';
    Message: string;
    Total: number;
    Completed: number;
}

/**
 * Dictionary Builder が保持する state 群。
 */
export interface DictionaryBuilderState {
    view: View;
    selectedRow: DictSourceRow | null;
    selectedRowId: string | null;
    selectedFiles: string[];
    isImporting: boolean;
    importMessages: Record<string, string>;
    showCrossSearch: boolean;
    sources: DictSourceRow[];
    entries: DictEntry[];
    entryPage: number;
    entryTotal: number;
    entryQuery: string;
    crossEntries: DictEntry[];
    crossPage: number;
    crossTotal: number;
    crossQuery: string;
}

/**
 * Dictionary Builder から UI に公開する操作群。
 */
export interface DictionaryBuilderActions {
    setView: (view: View) => void;
    openCrossSearch: () => void;
    closeCrossSearch: () => void;
    handleImport: () => Promise<void>;
    handleEntrySearch: (filters: Record<string, string>) => void;
    handleEntryPageChange: (page: number) => void;
    handleRowSelectAndFetch: (row: DictSourceRow | null, rowId: string | null) => void;
    handleCrossSearchExecute: (query: string) => void;
    handleCrossSearchFilter: (filters: Record<string, string>) => void;
    handleCrossPageChange: (page: number) => void;
    handleSelectFilesClick: () => Promise<void>;
    removeSelectedFile: (pathToRemove: string) => void;
    handleDeleteSource: () => Promise<void>;
    handleCancelDelete: () => void;
    handleEntriesSave: (modified: DictEntry[], deleted: DictEntry[]) => Promise<void>;
    handleCrossSave: (modified: DictEntry[], deleted: DictEntry[]) => Promise<void>;
}

/**
 * Dictionary Builder の描画用 UI 契約。
 */
interface DictionaryBuilderUi {
    sourceColumns: import('@tanstack/react-table').ColumnDef<DictSourceRow, unknown>[];
}

/**
 * Dictionary Builder hook の戻り値全体。
 */
export interface UseDictionaryBuilderResult {
    state: DictionaryBuilderState;
    actions: DictionaryBuilderActions;
    ui: DictionaryBuilderUi;
    constants: {
        pageSize: number;
    };
}
