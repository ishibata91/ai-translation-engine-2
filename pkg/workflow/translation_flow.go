package workflow

import "context"

// LoadTranslationFlowInput is the workflow entry DTO for load-phase file ingestion.
type LoadTranslationFlowInput struct {
	TaskID    string   `json:"task_id"`
	FilePaths []string `json:"file_paths"`
}

// TranslationPreviewRow is one row shown in load-phase preview tables.
type TranslationPreviewRow struct {
	ID         string `json:"id"`
	Section    string `json:"section"`
	RecordType string `json:"record_type"`
	EditorID   string `json:"editor_id"`
	SourceText string `json:"source_text"`
}

// TranslationPreviewPage is one paged preview response for one loaded file.
type TranslationPreviewPage struct {
	FileID    int64                   `json:"file_id"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
	TotalRows int                     `json:"total_rows"`
	Rows      []TranslationPreviewRow `json:"rows"`
}

// TranslationLoadedFile is one file block rendered in the load panel.
type TranslationLoadedFile struct {
	FileID       int64                  `json:"file_id"`
	FilePath     string                 `json:"file_path"`
	FileName     string                 `json:"file_name"`
	ParseStatus  string                 `json:"parse_status"`
	PreviewCount int                    `json:"preview_count"`
	Preview      TranslationPreviewPage `json:"preview"`
}

// TranslationLoadResult is the aggregate response for load-phase file list retrieval.
type TranslationLoadResult struct {
	TaskID string                  `json:"task_id"`
	Files  []TranslationLoadedFile `json:"files"`
}

// TranslationRequestConfig is the workflow DTO for terminology request settings.
type TranslationRequestConfig struct {
	Provider        string  `json:"provider"`
	Model           string  `json:"model"`
	Endpoint        string  `json:"endpoint"`
	APIKey          string  `json:"api_key"`
	Temperature     float32 `json:"temperature"`
	ContextLength   int     `json:"context_length"`
	SyncConcurrency int     `json:"sync_concurrency"`
	BulkStrategy    string  `json:"bulk_strategy"`
}

// TranslationPromptConfig is the workflow DTO for terminology prompt settings.
type TranslationPromptConfig struct {
	UserPrompt   string `json:"user_prompt"`
	SystemPrompt string `json:"system_prompt"`
}

// TerminologyTargetPreviewRow is one terminology target row shown before execution.
type TerminologyTargetPreviewRow struct {
	ID               string `json:"id"`
	RecordType       string `json:"record_type"`
	EditorID         string `json:"editor_id"`
	SourceText       string `json:"source_text"`
	TranslatedText   string `json:"translated_text"`
	TranslationState string `json:"translation_state"`
	Variant          string `json:"variant"`
	SourceFile       string `json:"source_file"`
}

// TerminologyTargetPreviewPage is one paged response for terminology targets.
type TerminologyTargetPreviewPage struct {
	TaskID    string                        `json:"task_id"`
	Page      int                           `json:"page"`
	PageSize  int                           `json:"page_size"`
	TotalRows int                           `json:"total_rows"`
	Rows      []TerminologyTargetPreviewRow `json:"rows"`
}

// RunTerminologyPhaseInput is the controller-facing DTO for terminology phase execution.
type RunTerminologyPhaseInput struct {
	TaskID  string                   `json:"task_id"`
	Request TranslationRequestConfig `json:"request"`
	Prompt  TranslationPromptConfig  `json:"prompt"`
}

// TerminologyPhaseResult is the aggregate response for terminology phase status.
type TerminologyPhaseResult struct {
	TaskID          string `json:"task_id"`
	Status          string `json:"status"`
	SavedCount      int    `json:"saved_count"`
	FailedCount     int    `json:"failed_count"`
	ProgressMode    string `json:"progress_mode"`
	ProgressCurrent int    `json:"progress_current"`
	ProgressTotal   int    `json:"progress_total"`
	ProgressMessage string `json:"progress_message"`
}

// PersonaDialogueView is one dialogue excerpt rendered in persona detail panes.
type PersonaDialogueView struct {
	RecordType       string `json:"record_type"`
	EditorID         string `json:"editor_id"`
	SourceText       string `json:"source_text"`
	QuestID          string `json:"quest_id"`
	IsServicesBranch bool   `json:"is_services_branch"`
	Order            int    `json:"order"`
}

// PersonaTargetPreviewRow is one persona-target row shown before or after execution.
type PersonaTargetPreviewRow struct {
	SourcePlugin string                `json:"source_plugin"`
	SpeakerID    string                `json:"speaker_id"`
	EditorID     string                `json:"editor_id"`
	NPCName      string                `json:"npc_name"`
	Race         string                `json:"race"`
	Sex          string                `json:"sex"`
	VoiceType    string                `json:"voice_type"`
	ViewState    string                `json:"view_state"`
	PersonaText  string                `json:"persona_text"`
	ErrorMessage string                `json:"error_message"`
	Dialogues    []PersonaDialogueView `json:"dialogues"`
}

// PersonaTargetPreviewPage is one paged response for persona targets.
type PersonaTargetPreviewPage struct {
	TaskID    string                    `json:"task_id"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
	TotalRows int                       `json:"total_rows"`
	Rows      []PersonaTargetPreviewRow `json:"rows"`
}

// RunTranslationFlowPersonaPhaseInput is the controller-facing DTO for persona phase execution.
type RunTranslationFlowPersonaPhaseInput struct {
	TaskID  string                   `json:"task_id"`
	Request TranslationRequestConfig `json:"request"`
	Prompt  TranslationPromptConfig  `json:"prompt"`
}

// PersonaPhaseResult is the aggregate response for persona phase status.
type PersonaPhaseResult struct {
	TaskID          string `json:"task_id"`
	Status          string `json:"status"`
	DetectedCount   int    `json:"detected_count"`
	ReusedCount     int    `json:"reused_count"`
	PendingCount    int    `json:"pending_count"`
	GeneratedCount  int    `json:"generated_count"`
	FailedCount     int    `json:"failed_count"`
	ProgressMode    string `json:"progress_mode"`
	ProgressCurrent int    `json:"progress_current"`
	ProgressTotal   int    `json:"progress_total"`
	ProgressMessage string `json:"progress_message"`
}

// TranslationFlow defines controller-facing workflow APIs for translation-flow phases.
type TranslationFlow interface {
	LoadFiles(ctx context.Context, input LoadTranslationFlowInput) (TranslationLoadResult, error)
	ListFiles(ctx context.Context, taskID string) (TranslationLoadResult, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (TranslationPreviewPage, error)
	ListTerminologyTargets(ctx context.Context, taskID string, page int, pageSize int) (TerminologyTargetPreviewPage, error)
	RunTerminologyPhase(ctx context.Context, input RunTerminologyPhaseInput) (TerminologyPhaseResult, error)
	GetTerminologyPhase(ctx context.Context, taskID string) (TerminologyPhaseResult, error)
	ListTranslationFlowPersonaTargets(ctx context.Context, taskID string, page int, pageSize int) (PersonaTargetPreviewPage, error)
	RunTranslationFlowPersonaPhase(ctx context.Context, input RunTranslationFlowPersonaPhaseInput) (PersonaPhaseResult, error)
	GetTranslationFlowPersonaPhase(ctx context.Context, taskID string) (PersonaPhaseResult, error)
}
