import type {
    LoadedTranslationFile,
    TerminologyPhaseSummary,
    TranslationTargetRow,
    WailsTerminologyPhaseResult,
    WailsTranslationLoadedFile,
    WailsTranslationLoadResult,
    WailsTranslationPreviewPage,
    WailsTranslationPreviewRow,
} from './types';

const asRecord = (value: unknown): Record<string, unknown> | null => {
    if (value && typeof value === 'object') {
        return value as Record<string, unknown>;
    }
    return null;
};

const pickString = (value: unknown, fallback = ''): string =>
    typeof value === 'string' ? value : fallback;

const pickNumber = (value: unknown, fallback = 0): number =>
    typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const mapPreviewRow = (payload: WailsTranslationPreviewRow): TranslationTargetRow => ({
    id: pickString(payload.id),
    section: pickString(payload.section),
    recordType: pickString(payload.record_type ?? payload.recordType),
    editorId: pickString(payload.editor_id ?? payload.editorId),
    sourceText: pickString(payload.source_text ?? payload.sourceText),
});

export const mapPreviewPage = (payload: unknown): {
    fileId: number;
    page: number;
    pageSize: number;
    totalRows: number;
    rows: TranslationTargetRow[];
} => {
    const pagePayload = (asRecord(payload) ?? {}) as WailsTranslationPreviewPage;
    const rawRows = Array.isArray(pagePayload.rows) ? pagePayload.rows : [];

    return {
        fileId: pickNumber(pagePayload.file_id ?? pagePayload.fileId),
        page: Math.max(1, pickNumber(pagePayload.page, 1)),
        pageSize: Math.max(1, pickNumber(pagePayload.page_size ?? pagePayload.pageSize, 50)),
        totalRows: Math.max(0, pickNumber(pagePayload.total_rows ?? pagePayload.totalRows, 0)),
        rows: rawRows.map(mapPreviewRow),
    };
};

const mapLoadedFile = (payload: WailsTranslationLoadedFile): LoadedTranslationFile => {
    const preview = mapPreviewPage(payload.preview ?? {});
    const fileName = pickString(payload.file_name ?? payload.fileName);
    const filePath = pickString(payload.file_path ?? payload.filePath);

    return {
        fileId: pickNumber(payload.file_id ?? payload.fileId),
        filePath,
        fileName: fileName !== '' ? fileName : filePath,
        parseStatus: pickString(payload.parse_status ?? payload.parseStatus),
        rowCount: Math.max(0, pickNumber(payload.preview_count ?? payload.previewCount)),
        currentPage: preview.page,
        pageSize: preview.pageSize,
        totalRows: preview.totalRows,
        rows: preview.rows,
    };
};

export const mapLoadResult = (payload: unknown): {
    taskId: string;
    files: LoadedTranslationFile[];
} => {
    const resultPayload = (asRecord(payload) ?? {}) as WailsTranslationLoadResult;
    const rawFiles = Array.isArray(resultPayload.files) ? resultPayload.files : [];

    return {
        taskId: pickString(resultPayload.task_id ?? resultPayload.taskId),
        files: rawFiles
            .map((entry) => asRecord(entry))
            .filter((entry): entry is Record<string, unknown> => entry !== null)
            .map((entry) => mapLoadedFile(entry as WailsTranslationLoadedFile))
            .filter((file) => file.fileId > 0),
    };
};

export const mapTerminologyPhaseResult = (payload: unknown): TerminologyPhaseSummary => {
    const resultPayload = (asRecord(payload) ?? {}) as WailsTerminologyPhaseResult;
    return {
        taskId: pickString(resultPayload.task_id ?? resultPayload.taskId),
        status: pickString(resultPayload.status, 'pending'),
        targetCount: Math.max(0, pickNumber(resultPayload.target_count ?? resultPayload.targetCount, 0)),
        savedCount: Math.max(0, pickNumber(resultPayload.saved_count ?? resultPayload.savedCount, 0)),
        failedCount: Math.max(0, pickNumber(resultPayload.failed_count ?? resultPayload.failedCount, 0)),
    };
};
