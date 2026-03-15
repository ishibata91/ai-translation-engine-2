package translationflow

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
)

// LoadedFile represents one saved translation input file for a task.
type LoadedFile struct {
	ID              int64
	TaskID          string
	SourceFilePath  string
	SourceFileName  string
	SourceFileHash  string
	ParseStatus     string
	PreviewRowCount int
}

// PreviewRow is one row displayed in translation-flow load preview tables.
type PreviewRow struct {
	ID         string
	Section    string
	RecordType string
	EditorID   string
	SourceText string
}

// PreviewPage represents paginated preview rows for one file.
type PreviewPage struct {
	FileID    int64
	Page      int
	PageSize  int
	TotalRows int
	Rows      []PreviewRow
}

// Service defines translation-flow input storage operations exposed to workflow.
type Service interface {
	EnsureTask(ctx context.Context, taskID string) error
	SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (LoadedFile, error)
	ListFiles(ctx context.Context, taskID string) ([]LoadedFile, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error)
}
