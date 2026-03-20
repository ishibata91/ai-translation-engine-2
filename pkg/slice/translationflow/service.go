package translationflow

import (
	"context"
	"fmt"

	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
)

type service struct {
	repo translationinput.Repository
}

// NewService creates a translation-flow slice service backed by artifact repository.
func NewService(repo translationinput.Repository) Service {
	return &service{repo: repo}
}

// EnsureTask creates or updates task parent row in artifact storage.
func (s *service) EnsureTask(ctx context.Context, taskID string) error {
	if err := s.repo.EnsureTask(ctx, taskID); err != nil {
		return fmt.Errorf("ensure translation-flow task task_id=%s: %w", taskID, err)
	}
	return nil
}

// SaveParsedOutput persists one parser output under the task/file boundary.
func (s *service) SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (LoadedFile, error) {
	stored, err := s.repo.SaveParsedOutput(ctx, taskID, sourceFilePath, output)
	if err != nil {
		return LoadedFile{}, fmt.Errorf("save parsed translation input task_id=%s file=%s: %w", taskID, sourceFilePath, err)
	}
	return LoadedFile{
		ID:              stored.ID,
		TaskID:          stored.TaskID,
		SourceFilePath:  stored.SourceFilePath,
		SourceFileName:  stored.SourceFileName,
		SourceFileHash:  stored.SourceFileHash,
		ParseStatus:     stored.ParseStatus,
		PreviewRowCount: stored.PreviewRowCount,
	}, nil
}

// ListFiles loads all saved files for one translation-flow task.
func (s *service) ListFiles(ctx context.Context, taskID string) ([]LoadedFile, error) {
	storedFiles, err := s.repo.ListFiles(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("list parsed files task_id=%s: %w", taskID, err)
	}
	files := make([]LoadedFile, 0, len(storedFiles))
	for _, stored := range storedFiles {
		files = append(files, LoadedFile{
			ID:              stored.ID,
			TaskID:          stored.TaskID,
			SourceFilePath:  stored.SourceFilePath,
			SourceFileName:  stored.SourceFileName,
			SourceFileHash:  stored.SourceFileHash,
			ParseStatus:     stored.ParseStatus,
			PreviewRowCount: stored.PreviewRowCount,
		})
	}
	return files, nil
}

// ListPreviewRows loads paginated preview rows for one file.
func (s *service) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error) {
	preview, err := s.repo.ListPreviewRows(ctx, fileID, page, pageSize)
	if err != nil {
		return PreviewPage{}, fmt.Errorf("list preview rows file_id=%d page=%d size=%d: %w", fileID, page, pageSize, err)
	}
	rows := make([]PreviewRow, 0, len(preview.Rows))
	for _, row := range preview.Rows {
		rows = append(rows, PreviewRow{
			ID:         row.ID,
			Section:    row.Section,
			RecordType: row.RecordType,
			EditorID:   row.EditorID,
			SourceText: row.SourceText,
		})
	}
	return PreviewPage{
		FileID:    preview.FileID,
		Page:      preview.Page,
		PageSize:  preview.PageSize,
		TotalRows: preview.TotalRows,
		Rows:      rows,
	}, nil
}

// LoadTerminologyInput exposes persisted terminology rows for workflow preview projection.
func (s *service) LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error) {
	input, err := s.repo.LoadTerminologyInput(ctx, taskID)
	if err != nil {
		return translationinput.TerminologyInput{}, fmt.Errorf("load terminology input task_id=%s: %w", taskID, err)
	}
	return input, nil
}
