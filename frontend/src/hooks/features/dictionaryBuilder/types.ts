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
