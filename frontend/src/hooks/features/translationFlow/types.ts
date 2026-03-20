import type {MasterPersonaLLMConfig, MasterPersonaPromptConfig} from '../../../types/masterPersona';

/**
 * 翻訳フローのタブ表示に使う定義。
 */
export interface TranslationFlowTab {
    label: string;
}

/**
 * ロード済み preview テーブルの 1 行。
 */
export interface TranslationTargetRow {
    id: string;
    section: string;
    recordType: string;
    editorId: string;
    sourceText: string;
}

/**
 * 単語翻訳 phase の対象一覧 1 行。
 */
export interface TerminologyTargetPreviewRow {
    id: string;
    recordType: string;
    editorId: string;
    sourceText: string;
    variant: string;
    sourceFile: string;
}

/**
 * 単語翻訳 phase の対象一覧ページ。
 */
export interface TerminologyTargetPreviewPage {
    taskId: string;
    page: number;
    pageSize: number;
    totalRows: number;
    rows: TerminologyTargetPreviewRow[];
}

/**
 * ロード済みファイルと preview ページング状態。
 */
export interface LoadedTranslationFile {
    fileId: number;
    filePath: string;
    fileName: string;
    parseStatus: string;
    rowCount: number;
    currentPage: number;
    pageSize: number;
    totalRows: number;
    rows: TranslationTargetRow[];
}

/**
 * 単語翻訳 phase の実行サマリ。
 */
export interface TerminologyPhaseSummary {
    taskId: string;
    status: string;
    savedCount: number;
    failedCount: number;
}

/**
 * 単語翻訳 phase の対象一覧の表示状態。
 */
export type TerminologyTargetViewState = 'loading' | 'ready' | 'empty' | 'error';

/**
 * TranslationFlow が保持する state 群。
 */
interface TranslationFlowState {
    taskId: string;
    activeTab: number;
    tabs: TranslationFlowTab[];
    selectedFiles: string[];
    loadedFiles: LoadedTranslationFile[];
    isLoading: boolean;
    errorMessage: string;
    terminologySummary: TerminologyPhaseSummary;
    terminologyStatusLabel: string;
    terminologyErrorMessage: string;
    terminologyTargetPage: TerminologyTargetPreviewPage;
    terminologyTargetStatus: TerminologyTargetViewState;
    terminologyTargetErrorMessage: string;
    isTerminologyTargetLoading: boolean;
    isTerminologyRunning: boolean;
    terminologyConfig: MasterPersonaLLMConfig;
    terminologyPromptConfig: MasterPersonaPromptConfig;
    isTerminologyConfigHydrated: boolean;
    isTerminologyPromptHydrated: boolean;
}

/**
 * TranslationFlow から UI に公開する操作群。
 */
interface TranslationFlowActions {
    handleTabChange: (index: number) => void;
    handleSelectFiles: () => Promise<void>;
    handleRemoveFile: (pathToRemove: string) => void;
    handleLoadSelectedFiles: () => Promise<void>;
    handleReloadFiles: () => Promise<void>;
    handlePreviewPageChange: (fileId: number, page: number) => Promise<void>;
    handleAdvanceFromLoad: () => void;
    handleRunTerminologyPhase: () => Promise<void>;
    handleRefreshTerminologyPhase: () => Promise<void>;
    handleTerminologyTargetPageChange: (page: number) => Promise<void>;
    handleTerminologyConfigChange: (next: MasterPersonaLLMConfig) => void;
    handleTerminologyPromptChange: (next: MasterPersonaPromptConfig) => void;
    handleAdvanceFromTerminology: () => void;
}

/**
 * TranslationFlow hook の戻り値。
 */
export interface UseTranslationFlowResult {
    state: TranslationFlowState;
    actions: TranslationFlowActions;
    ui: {
        previewPageSize: number;
    };
}

/**
 * Wails の load result payload。
 */
export interface WailsTranslationLoadResult {
    task_id?: string;
    taskId?: string;
    files?: unknown[];
}

/**
 * Wails の loaded file payload。
 */
export interface WailsTranslationLoadedFile {
    file_id?: number;
    fileId?: number;
    file_path?: string;
    filePath?: string;
    file_name?: string;
    fileName?: string;
    parse_status?: string;
    parseStatus?: string;
    preview_count?: number;
    previewCount?: number;
    preview?: WailsTranslationPreviewPage;
}

/**
 * Wails の preview page payload。
 */
export interface WailsTranslationPreviewPage {
    file_id?: number;
    fileId?: number;
    page?: number;
    pageSize?: number;
    page_size?: number;
    totalRows?: number;
    total_rows?: number;
    rows?: WailsTranslationPreviewRow[];
}

/**
 * Wails の preview row payload。
 */
export interface WailsTranslationPreviewRow {
    id?: string;
    section?: string;
    record_type?: string;
    recordType?: string;
    editor_id?: string;
    editorId?: string;
    source_text?: string;
    sourceText?: string;
}

/**
 * Wails の terminology phase payload。
 */
export interface WailsTerminologyPhaseResult {
    task_id?: string;
    taskId?: string;
    status?: string;
    saved_count?: number;
    savedCount?: number;
    failed_count?: number;
    failedCount?: number;
}

/**
 * Wails の terminology target preview page payload。
 */
export interface WailsTerminologyTargetPreviewPage {
    task_id?: string;
    taskId?: string;
    page?: number;
    pageSize?: number;
    page_size?: number;
    totalRows?: number;
    total_rows?: number;
    rows?: WailsTerminologyTargetPreviewRow[];
}

/**
 * Wails の terminology target preview row payload。
 */
export interface WailsTerminologyTargetPreviewRow {
    id?: string;
    record_type?: string;
    recordType?: string;
    editor_id?: string;
    editorId?: string;
    source_text?: string;
    sourceText?: string;
    variant?: string;
    source_file?: string;
    sourceFile?: string;
}
