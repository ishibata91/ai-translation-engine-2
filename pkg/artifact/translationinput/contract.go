package translationinput

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
)

// InputFile represents one parsed source file persisted in artifact storage.
type InputFile struct {
	ID              int64
	TaskID          string
	SourceFilePath  string
	SourceFileName  string
	SourceFileHash  string
	ParseStatus     string
	PreviewRowCount int
}

// PreviewRow is one translation-target row projected for load-phase tables.
type PreviewRow struct {
	ID         string
	Section    string
	RecordType string
	EditorID   string
	SourceText string
}

// PreviewPage is one paged preview response for a file.
type PreviewPage struct {
	FileID    int64
	Page      int
	PageSize  int
	TotalRows int
	Rows      []PreviewRow
}

// Repository defines artifact persistence operations for translation-flow input data.
type Repository interface {
	EnsureTask(ctx context.Context, taskID string) error
	SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (InputFile, error)
	ListFiles(ctx context.Context, taskID string) ([]InputFile, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error)
}
