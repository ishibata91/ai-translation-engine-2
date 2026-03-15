package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
)

const defaultTranslationPreviewPageSize = 50

// TranslationFlowService orchestrates parser execution and artifact persistence for load phase.
type TranslationFlowService struct {
	parser skyrim.Parser
	store  translationflow.Service
}

// NewTranslationFlowService constructs a translation-flow workflow implementation.
func NewTranslationFlowService(parser skyrim.Parser, store translationflow.Service) *TranslationFlowService {
	return &TranslationFlowService{
		parser: parser,
		store:  store,
	}
}

// LoadFiles parses selected files and stores them under the task boundary.
func (s *TranslationFlowService) LoadFiles(ctx context.Context, input LoadTranslationFlowInput) (TranslationLoadResult, error) {
	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return TranslationLoadResult{}, fmt.Errorf("task_id is required")
	}
	if len(input.FilePaths) == 0 {
		return TranslationLoadResult{}, fmt.Errorf("file_paths is required")
	}

	if err := s.store.EnsureTask(ctx, trimmedTaskID); err != nil {
		return TranslationLoadResult{}, fmt.Errorf("ensure translation-flow task task_id=%s: %w", trimmedTaskID, err)
	}

	for _, sourcePath := range input.FilePaths {
		trimmedPath := strings.TrimSpace(sourcePath)
		if trimmedPath == "" {
			continue
		}

		parsed, err := s.parser.LoadExtractedJSON(ctx, trimmedPath)
		if err != nil {
			return TranslationLoadResult{}, fmt.Errorf("parse source json task_id=%s file=%s: %w", trimmedTaskID, trimmedPath, err)
		}
		if _, err := s.store.SaveParsedOutput(ctx, trimmedTaskID, trimmedPath, parsed); err != nil {
			return TranslationLoadResult{}, fmt.Errorf("save parsed output task_id=%s file=%s: %w", trimmedTaskID, trimmedPath, err)
		}
	}

	return s.ListFiles(ctx, trimmedTaskID)
}

// ListFiles returns loaded files with first preview page for each file.
func (s *TranslationFlowService) ListFiles(ctx context.Context, taskID string) (TranslationLoadResult, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return TranslationLoadResult{}, fmt.Errorf("task_id is required")
	}

	files, err := s.store.ListFiles(ctx, trimmedTaskID)
	if err != nil {
		return TranslationLoadResult{}, fmt.Errorf("list translation-flow files task_id=%s: %w", trimmedTaskID, err)
	}

	loadedFiles := make([]TranslationLoadedFile, 0, len(files))
	for _, file := range files {
		previewPage, err := s.store.ListPreviewRows(ctx, file.ID, 1, defaultTranslationPreviewPageSize)
		if err != nil {
			return TranslationLoadResult{}, fmt.Errorf("list file preview task_id=%s file_id=%d: %w", trimmedTaskID, file.ID, err)
		}
		loadedFiles = append(loadedFiles, TranslationLoadedFile{
			FileID:       file.ID,
			FilePath:     file.SourceFilePath,
			FileName:     file.SourceFileName,
			ParseStatus:  file.ParseStatus,
			PreviewCount: file.PreviewRowCount,
			Preview:      mapPreviewPage(previewPage),
		})
	}

	return TranslationLoadResult{
		TaskID: trimmedTaskID,
		Files:  loadedFiles,
	}, nil
}

// ListPreviewRows returns one paged preview response for one file.
func (s *TranslationFlowService) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (TranslationPreviewPage, error) {
	previewPage, err := s.store.ListPreviewRows(ctx, fileID, page, pageSize)
	if err != nil {
		return TranslationPreviewPage{}, fmt.Errorf("list preview rows file_id=%d page=%d size=%d: %w", fileID, page, pageSize, err)
	}
	return mapPreviewPage(previewPage), nil
}

func mapPreviewPage(page translationflow.PreviewPage) TranslationPreviewPage {
	rows := make([]TranslationPreviewRow, 0, len(page.Rows))
	for _, row := range page.Rows {
		rows = append(rows, TranslationPreviewRow{
			ID:         row.ID,
			Section:    row.Section,
			RecordType: row.RecordType,
			EditorID:   row.EditorID,
			SourceText: row.SourceText,
		})
	}
	return TranslationPreviewPage{
		FileID:    page.FileID,
		Page:      page.Page,
		PageSize:  page.PageSize,
		TotalRows: page.TotalRows,
		Rows:      rows,
	}
}
