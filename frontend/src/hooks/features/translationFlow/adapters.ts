import type {
    LoadedTranslationFile,
    PersonaDialogueView,
    PersonaPhaseSummary,
    PersonaTargetStateBadge,
    PersonaTargetPreviewPage,
    PersonaTargetPreviewRow,
    PersonaTargetRowState,
    PersonaTargetViewState,
    TerminologyPhaseSummary,
    TerminologyTargetPreviewPage,
    TerminologyTargetPreviewRow,
    TranslationTargetRow,
    WailsPersonaDialogueView,
    WailsPersonaPhaseResult,
    WailsPersonaTargetPreviewPage,
    WailsPersonaTargetPreviewRow,
    WailsTerminologyPhaseResult,
    WailsTerminologyTargetPreviewPage,
    WailsTerminologyTargetPreviewRow,
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

const mapTerminologyTargetPreviewRow = (payload: WailsTerminologyTargetPreviewRow): TerminologyTargetPreviewRow => ({
    id: pickString(payload.id),
    recordType: pickString(payload.record_type ?? payload.recordType),
    editorId: pickString(payload.editor_id ?? payload.editorId),
    sourceText: pickString(payload.source_text ?? payload.sourceText),
    translatedText: pickString(payload.translated_text ?? payload.translatedText),
    translationState: pickString(payload.translation_state ?? payload.translationState, 'missing'),
    variant: pickString(payload.variant),
    sourceFile: pickString(payload.source_file ?? payload.sourceFile),
});

const mapPersonaTargetViewState = (value: unknown): PersonaTargetViewState => {
    const normalized = pickString(value).trim().toLowerCase();
    switch (normalized) {
    case 'loadingtargets':
    case 'loading_targets':
    case 'loading':
        return 'loadingTargets';
    case 'empty':
        return 'empty';
    case 'ready':
        return 'ready';
    case 'cachedonly':
    case 'cached_only':
        return 'cachedOnly';
    case 'running':
        return 'running';
    case 'completed':
        return 'completed';
    case 'partialfailed':
    case 'partial_failed':
        return 'partialFailed';
    case 'failed':
        return 'failed';
    default:
        return 'loadingTargets';
    }
};

const mapPersonaTargetRowState = (value: unknown): PersonaTargetRowState => {
    const normalized = pickString(value).trim().toLowerCase();
    switch (normalized) {
    case 'reused':
    case 'pending':
    case 'running':
    case 'generated':
    case 'failed':
        return normalized;
    default:
        return 'pending';
    }
};

const mapPersonaTargetStateBadge = (viewState: PersonaTargetRowState): PersonaTargetStateBadge => {
    switch (viewState) {
    case 'reused':
        return {label: '既存 Master Persona', tone: 'info'};
    case 'pending':
        return {label: '生成対象', tone: 'neutral'};
    case 'running':
        return {label: '生成中', tone: 'warning'};
    case 'generated':
        return {label: '生成済み', tone: 'success'};
    case 'failed':
        return {label: '生成失敗', tone: 'error'};
    default:
        return {label: '生成対象', tone: 'neutral'};
    }
};

const mapPersonaDialogueView = (payload: WailsPersonaDialogueView): PersonaDialogueView => ({
    recordType: pickString(payload.record_type ?? payload.recordType),
    editorId: pickString(payload.editor_id ?? payload.editorId),
    sourceText: pickString(payload.source_text ?? payload.sourceText),
    questId: pickString(payload.quest_id ?? payload.questId),
    isServicesBranch: Boolean(payload.is_services_branch ?? payload.isServicesBranch),
    order: Math.max(0, pickNumber(payload.order)),
});

const mapPersonaTargetPreviewRow = (payload: WailsPersonaTargetPreviewRow): PersonaTargetPreviewRow => {
    const rawDialogues = Array.isArray(payload.dialogues) ? payload.dialogues : [];
    const sourcePlugin = pickString(payload.source_plugin ?? payload.sourcePlugin).trim();
    const speakerId = pickString(payload.speaker_id ?? payload.speakerId).trim();
    const rowState = mapPersonaTargetRowState(payload.view_state ?? payload.viewState);
    const updatedAt = pickString(payload.updated_at ?? payload.updatedAt).trim();

    return {
        id: `${sourcePlugin}::${speakerId}`,
        formId: speakerId,
        sourcePlugin,
        speakerId,
        editorId: pickString(payload.editor_id ?? payload.editorId),
        npcName: pickString(payload.npc_name ?? payload.npcName),
        updatedAt: updatedAt !== '' ? updatedAt : undefined,
        race: pickString(payload.race),
        sex: pickString(payload.sex),
        voiceType: pickString(payload.voice_type ?? payload.voiceType),
        viewState: rowState,
        stateBadge: mapPersonaTargetStateBadge(rowState),
        personaText: pickString(payload.persona_text ?? payload.personaText),
        errorMessage: pickString(payload.error_message ?? payload.errorMessage),
        dialogues: rawDialogues.map(mapPersonaDialogueView),
    };
};

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
        savedCount: Math.max(0, pickNumber(resultPayload.saved_count ?? resultPayload.savedCount, 0)),
        failedCount: Math.max(0, pickNumber(resultPayload.failed_count ?? resultPayload.failedCount, 0)),
        progressMode: pickString(resultPayload.progress_mode ?? resultPayload.progressMode, 'hidden'),
        progressCurrent: Math.max(0, pickNumber(resultPayload.progress_current ?? resultPayload.progressCurrent, 0)),
        progressTotal: Math.max(0, pickNumber(resultPayload.progress_total ?? resultPayload.progressTotal, 0)),
        progressMessage: pickString(resultPayload.progress_message ?? resultPayload.progressMessage),
    };
};

export const mapTerminologyTargetPreviewPage = (payload: unknown): TerminologyTargetPreviewPage => {
    const pagePayload = (asRecord(payload) ?? {}) as WailsTerminologyTargetPreviewPage;
    const rawRows = Array.isArray(pagePayload.rows) ? pagePayload.rows : [];

    return {
        taskId: pickString(pagePayload.task_id ?? pagePayload.taskId),
        page: Math.max(1, pickNumber(pagePayload.page, 1)),
        pageSize: Math.max(1, pickNumber(pagePayload.pageSize ?? pagePayload.page_size, 50)),
        totalRows: Math.max(0, pickNumber(pagePayload.totalRows ?? pagePayload.total_rows, 0)),
        rows: rawRows.map(mapTerminologyTargetPreviewRow),
    };
};

export const mapPersonaTargetPreviewPage = (payload: unknown): PersonaTargetPreviewPage => {
    const pagePayload = (asRecord(payload) ?? {}) as WailsPersonaTargetPreviewPage;
    const rawRows = Array.isArray(pagePayload.rows) ? pagePayload.rows : [];

    return {
        taskId: pickString(pagePayload.task_id ?? pagePayload.taskId),
        page: Math.max(1, pickNumber(pagePayload.page, 1)),
        pageSize: Math.max(1, pickNumber(pagePayload.pageSize ?? pagePayload.page_size, 50)),
        totalRows: Math.max(0, pickNumber(pagePayload.totalRows ?? pagePayload.total_rows, 0)),
        rows: rawRows.map(mapPersonaTargetPreviewRow),
    };
};

export const mapPersonaPhaseResult = (payload: unknown): PersonaPhaseSummary => {
    const resultPayload = (asRecord(payload) ?? {}) as WailsPersonaPhaseResult;

    return {
        taskId: pickString(resultPayload.task_id ?? resultPayload.taskId),
        status: mapPersonaTargetViewState(resultPayload.status),
        detectedCount: Math.max(0, pickNumber(resultPayload.detected_count ?? resultPayload.detectedCount, 0)),
        reusedCount: Math.max(0, pickNumber(resultPayload.reused_count ?? resultPayload.reusedCount, 0)),
        pendingCount: Math.max(0, pickNumber(resultPayload.pending_count ?? resultPayload.pendingCount, 0)),
        generatedCount: Math.max(0, pickNumber(resultPayload.generated_count ?? resultPayload.generatedCount, 0)),
        failedCount: Math.max(0, pickNumber(resultPayload.failed_count ?? resultPayload.failedCount, 0)),
        progressMode: pickString(resultPayload.progress_mode ?? resultPayload.progressMode, 'hidden'),
        progressCurrent: Math.max(0, pickNumber(resultPayload.progress_current ?? resultPayload.progressCurrent, 0)),
        progressTotal: Math.max(0, pickNumber(resultPayload.progress_total ?? resultPayload.progressTotal, 0)),
        progressMessage: pickString(resultPayload.progress_message ?? resultPayload.progressMessage),
    };
};
