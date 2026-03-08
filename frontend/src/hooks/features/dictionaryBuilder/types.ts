export type SourceStatus = '完了' | 'インポート中' | 'エラー';

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

export interface DictEntry {
    id: number;
    sourceId: string;
    sourceName?: string;
    edid: string;
    recordType: string;
    sourceText: string;
    destText: string;
}

export interface DictEntryPage {
    entries: DictEntry[];
    totalCount: number;
}

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

export type View = 'list' | 'entries' | 'cross-search';

export interface DictionaryProgressEvent {
    CorrelationID: string;
    Status: 'STARTED' | 'COMPLETED' | 'FAILED' | 'IN_PROGRESS';
    Message: string;
    Total: number;
    Completed: number;
}

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

export interface DictionaryBuilderUi {
    sourceColumns: import('@tanstack/react-table').ColumnDef<DictSourceRow, unknown>[];
}

export interface UseDictionaryBuilderResult {
    state: DictionaryBuilderState;
    actions: DictionaryBuilderActions;
    ui: DictionaryBuilderUi;
    constants: {
        pageSize: number;
    };
}
