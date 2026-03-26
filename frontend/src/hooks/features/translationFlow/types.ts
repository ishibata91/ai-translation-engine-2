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
    translatedText: string;
    translationState: string;
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
 * ペルソナ詳細ペインに表示する会話抜粋 1 行。
 */
export interface PersonaDialogueView {
    recordType: string;
    editorId: string;
    sourceText: string;
    questId: string;
    isServicesBranch: boolean;
    order: number;
}

/**
 * ペルソナ対象一覧 1 行の実行状態。
 */
export type PersonaTargetRowState = 'reused' | 'pending' | 'running' | 'generated' | 'failed';

/**
 * shared persona list で使う state badge の tone。
 */
type PersonaTargetStateTone = 'neutral' | 'info' | 'warning' | 'success' | 'error';

/**
 * shared persona list で使う state badge。
 */
export interface PersonaTargetStateBadge {
    label: string;
    tone?: PersonaTargetStateTone;
}

/**
 * ペルソナ生成 phase の UI state machine。
 */
export type PersonaTargetViewState =
    | 'loadingTargets'
    | 'empty'
    | 'ready'
    | 'cachedOnly'
    | 'running'
    | 'completed'
    | 'partialFailed'
    | 'failed';

/**
 * ペルソナ生成 phase の対象一覧 1 行。
 */
export interface PersonaTargetPreviewRow {
    id: string;
    formId: string;
    sourcePlugin: string;
    speakerId: string;
    editorId: string;
    npcName: string;
    updatedAt?: string;
    race: string;
    sex: string;
    voiceType: string;
    viewState: PersonaTargetRowState;
    stateBadge?: PersonaTargetStateBadge | null;
    personaText: string;
    errorMessage: string;
    dialogues: PersonaDialogueView[];
}

/**
 * ペルソナ生成 phase の対象一覧ページ。
 */
export interface PersonaTargetPreviewPage {
    taskId: string;
    page: number;
    pageSize: number;
    totalRows: number;
    rows: PersonaTargetPreviewRow[];
}

/**
 * ペルソナ生成 phase の実行サマリ。
 */
export interface PersonaPhaseSummary {
    taskId: string;
    status: PersonaTargetViewState;
    detectedCount: number;
    reusedCount: number;
    pendingCount: number;
    generatedCount: number;
    failedCount: number;
    progressMode: string;
    progressCurrent: number;
    progressTotal: number;
    progressMessage: string;
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
    progressMode: string;
    progressCurrent: number;
    progressTotal: number;
    progressMessage: string;
}

/**
 * 単語翻訳 phase の対象一覧の表示状態。
 */
export type TerminologyTargetViewState = 'loading' | 'ready' | 'empty' | 'error';

/**
 * 本文翻訳 phase のカテゴリ。
 */
type TranslationMainCategory = 'conversation' | 'quest' | 'other';

/**
 * 本文 1 行の状態。
 */
type TranslationMainRowStatus = 'untranslated' | 'aiTranslated' | 'confirmed';

/**
 * 本文翻訳 phase の view state。
 */
type TranslationMainViewState =
    | 'hydrating'
    | 'loadError'
    | 'empty'
    | 'ready'
    | 'selectionEmpty'
    | 'selectionReady'
    | 'translating'
    | 'translateCompleted'
    | 'translatePartialFailed'
    | 'translateFailed';

/**
 * 参照コンテキスト panel の 1 セクション。
 */
interface TranslationReferencePanel {
    title: string;
    items: string[];
}

/**
 * 本文翻訳 row の補助メタデータ。
 */
interface TranslationMainRowMetadata {
    sourcePlugin: string;
    recordType: string;
    editorId: string;
    section: string;
    speakerId: string;
    npcName: string;
    questId: string;
    stageIndex: number | null;
    objective: string;
}

/**
 * 本文翻訳 row の表示用 DTO。
 */
interface TranslationMainRow {
    id: string;
    category: TranslationMainCategory;
    displayLabel: string;
    secondaryLabel: string;
    sourceText: string;
    translatedText: string;
    status: TranslationMainRowStatus;
    metadata: TranslationMainRowMetadata;
    referencePanels: TranslationReferencePanel[];
    systemPrompt: string;
}

/**
 * 本文翻訳一覧のカテゴリ別 grouping。
 */
interface TranslationMainCategoryGroup {
    category: TranslationMainCategory;
    label: string;
    rows: TranslationMainRow[];
}

/**
 * 本文翻訳 phase のサマリ。
 */
interface TranslationMainSummary {
    totalCount: number;
    untranslatedCount: number;
    aiTranslatedCount: number;
    confirmedCount: number;
    failedCount: number;
}

/**
 * 破棄確認の遷移意図。
 */
interface TranslationPendingNavigationIntent {
    type: 'category' | 'row' | 'tab';
    nextCategory?: TranslationMainCategory;
    nextRowId?: string;
    nextTab?: number;
}

/**
 * 本文翻訳 phase のカテゴリ。
 */
export type MainTranslationCategory = TranslationMainCategory;
/**
 * 本文翻訳 phase の行状態。
 */
export type MainTranslationRowStatus = TranslationMainRowStatus;
/**
 * 本文翻訳 phase の実行状態。
 */
export type MainTranslationRunState = TranslationMainViewState;
/**
 * 本文翻訳 phase の遷移意図。
 */
export type MainTranslationNavigationIntent = 'selectRow' | 'switchCategory' | 'search' | 'page' | 'phaseChange' | 'next';
/**
 * 本文翻訳行のメタデータ。
 */
export type MainTranslationRowMetadata = TranslationMainRowMetadata;
/**
 * 本文翻訳行の normalized view-model。
 */
export interface MainTranslationRowViewModel {
    rowId: string;
    category: MainTranslationCategory;
    primaryLabel: string;
    secondaryMeta: string[];
    sourceText: string;
    translatedText: string;
    status: MainTranslationRowStatus;
    metadata: MainTranslationRowMetadata;
}
/**
 * 本文翻訳 phase のドラフト状態。
 */
export interface MainTranslationDraftState {
    selectedCategory: MainTranslationCategory;
    selectedRowId: string;
    dirtyDraftRowId: string;
    pendingNavigationIntent: MainTranslationNavigationIntent | null;
    draftMap: Record<string, string>;
    confirmedMap: Record<string, string>;
}
/**
 * 本文翻訳 phase の summary。
 */
export interface MainTranslationSummary {
    untranslatedCount: number;
    aiTranslatedCount: number;
    confirmedCount: number;
    failedCount: number;
}

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
    personaConfig?: MasterPersonaLLMConfig;
    personaPromptConfig?: MasterPersonaPromptConfig;
    isPersonaConfigHydrated?: boolean;
    isPersonaPromptHydrated?: boolean;
    translationConfig?: MasterPersonaLLMConfig;
    translationUserPrompt?: string;
    isTranslationConfigHydrated?: boolean;
    translationViewState?: TranslationMainViewState;
    translationSummary?: TranslationMainSummary;
    translationStatusLabel?: string;
    translationErrorMessage?: string;
    translationCategoryGroups?: TranslationMainCategoryGroup[];
    selectedTranslationCategory?: TranslationMainCategory;
    selectedTranslationRowId?: string;
    selectedTranslationRow?: TranslationMainRow | null;
    translationDraftText?: string;
    translationIsDirty?: boolean;
    translationPendingNavigationIntent?: TranslationPendingNavigationIntent | null;
    showTranslationDirtyWarning?: boolean;
    showTranslationNextWarning?: boolean;
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
    handlePersonaConfigChange?: (next: MasterPersonaLLMConfig) => void;
    handlePersonaPromptChange?: (next: MasterPersonaPromptConfig) => void;
    handleAdvanceFromTerminology: () => void;
    handleTranslationConfigChange?: (next: MasterPersonaLLMConfig) => void;
    handleTranslationUserPromptChange?: (next: string) => void;
    handleTranslationCategoryChange?: (next: TranslationMainCategory) => void;
    handleTranslationSelectRow?: (rowId: string) => void;
    handleTranslationDraftChange?: (next: string) => void;
    handleRunTranslationPhase?: () => Promise<void>;
    handleRetryTranslationPhase?: () => Promise<void>;
    handleConfirmTranslationDraft?: () => void;
    handleCancelConfirmedTranslation?: () => void;
    handleConfirmTranslationNavigation?: () => void;
    handleDismissTranslationNavigation?: () => void;
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
    progress_mode?: string;
    progressMode?: string;
    progress_current?: number;
    progressCurrent?: number;
    progress_total?: number;
    progressTotal?: number;
    progress_message?: string;
    progressMessage?: string;
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
    translated_text?: string;
    translatedText?: string;
    translation_state?: string;
    translationState?: string;
    variant?: string;
    source_file?: string;
    sourceFile?: string;
}

/**
 * Wails の persona dialogue payload。
 */
export interface WailsPersonaDialogueView {
    record_type?: string;
    recordType?: string;
    editor_id?: string;
    editorId?: string;
    source_text?: string;
    sourceText?: string;
    quest_id?: string;
    questId?: string;
    is_services_branch?: boolean;
    isServicesBranch?: boolean;
    order?: number;
}

/**
 * Wails の persona target preview row payload。
 */
export interface WailsPersonaTargetPreviewRow {
    source_plugin?: string;
    sourcePlugin?: string;
    speaker_id?: string;
    speakerId?: string;
    editor_id?: string;
    editorId?: string;
    npc_name?: string;
    npcName?: string;
    updated_at?: string;
    updatedAt?: string;
    race?: string;
    sex?: string;
    voice_type?: string;
    voiceType?: string;
    view_state?: string;
    viewState?: string;
    persona_text?: string;
    personaText?: string;
    error_message?: string;
    errorMessage?: string;
    dialogues?: WailsPersonaDialogueView[];
}

/**
 * Wails の persona target preview page payload。
 */
export interface WailsPersonaTargetPreviewPage {
    task_id?: string;
    taskId?: string;
    page?: number;
    pageSize?: number;
    page_size?: number;
    totalRows?: number;
    total_rows?: number;
    rows?: WailsPersonaTargetPreviewRow[];
}

/**
 * Wails の persona phase payload。
 */
export interface WailsPersonaPhaseResult {
    task_id?: string;
    taskId?: string;
    status?: string;
    detected_count?: number;
    detectedCount?: number;
    reused_count?: number;
    reusedCount?: number;
    pending_count?: number;
    pendingCount?: number;
    generated_count?: number;
    generatedCount?: number;
    failed_count?: number;
    failedCount?: number;
    progress_mode?: string;
    progressMode?: string;
    progress_current?: number;
    progressCurrent?: number;
    progress_total?: number;
    progressTotal?: number;
    progress_message?: string;
    progressMessage?: string;
}

/**
 * terminology progress bridge event payload。
 */
export interface WailsTerminologyProgressEvent {
    task_id?: string;
    taskId?: string;
    TaskID?: string;
    status?: string;
    Status?: string;
    current?: number;
    Current?: number;
    completed?: number;
    Completed?: number;
    total?: number;
    Total?: number;
    failed?: number;
    Failed?: number;
    message?: string;
    Message?: string;
}

/**
 * 本文翻訳 preview row payload。
 */
export interface WailsMainTranslationPreviewRow {
    id?: string;
    section?: string;
    record_type?: string;
    recordType?: string;
    editor_id?: string;
    editorId?: string;
    source_plugin?: string;
    sourcePlugin?: string;
    speaker_id?: string;
    speakerId?: string;
    npc_name?: string;
    npcName?: string;
    quest_id?: string;
    questId?: string;
    stage_index?: number;
    stageIndex?: number;
    objective?: string;
    source_text?: string;
    sourceText?: string;
    translated_text?: string;
    translatedText?: string;
    translation_state?: string;
    translationState?: string;
}
