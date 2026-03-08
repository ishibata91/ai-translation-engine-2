import type { DictEntry, DictEntryPage, DictSourceRow, DictUpdateEntryPayload, SourceStatus } from './types';

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

const toSourceStatus = (value: unknown): SourceStatus => {
    if (value === 'COMPLETED') return '完了';
    if (value === 'ERROR') return 'エラー';
    return 'インポート中';
};

const mapEntryRecord = (value: unknown): DictEntry => {
    const record = asRecord(value) ?? {};
    return {
        id: pickNumber(record.id ?? record.ID),
        sourceId: String(record.source_id ?? record.SourceID ?? ''),
        sourceName: pickString(record.source_name ?? record.SourceName),
        edid: pickString(record.edid ?? record.EDID),
        sourceText: pickString(record.source_text ?? record.Source),
        destText: pickString(record.dest_text ?? record.Dest),
        recordType: pickString(record.record_type ?? record.RecordType),
    };
};

export const mapSourcesResponse = (payload: unknown): DictSourceRow[] => {
    if (!Array.isArray(payload)) {
        return [];
    }

    return payload.map((item) => {
        const row = asRecord(item) ?? {};
        const importedAt = pickString(row.imported_at);
        const bytes = pickNumber(row.file_size_bytes);
        return {
            id: String(row.id ?? ''),
            fileName: pickString(row.file_name),
            format: pickString(row.format),
            entryCount: pickNumber(row.entry_count),
            status: toSourceStatus(row.status),
            updatedAt: importedAt ? new Date(importedAt).toLocaleString() : '-',
            filePath: pickString(row.file_path),
            fileSize: `${(bytes / 1024).toFixed(1)} KB`,
            importDuration: '-',
            errorMessage: pickString(row.error_message) || null,
        };
    });
};

export const mapEntriesPaginatedResponse = (payload: unknown): DictEntryPage => {
    const result = asRecord(payload);
    if (!result) {
        return { entries: [], totalCount: 0 };
    }

    const rawEntries = Array.isArray(result.entries)
        ? result.entries
        : Array.isArray(result.Entries)
            ? result.Entries
            : [];
    const totalCount = pickNumber(result.totalCount ?? result.TotalCount);

    return {
        entries: rawEntries.map(mapEntryRecord),
        totalCount,
    };
};

export const toDictUpdateEntryPayload = (entry: DictEntry): DictUpdateEntryPayload => ({
    id: entry.id,
    source_id: Number.parseInt(entry.sourceId, 10),
    edid: entry.edid,
    record_type: entry.recordType,
    source_text: entry.sourceText,
    dest_text: entry.destText,
});
