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

// TranslationFlow defines controller-facing workflow APIs for translation load phase.
type TranslationFlow interface {
	LoadFiles(ctx context.Context, input LoadTranslationFlowInput) (TranslationLoadResult, error)
	ListFiles(ctx context.Context, taskID string) (TranslationLoadResult, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (TranslationPreviewPage, error)
}
