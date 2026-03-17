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

// TerminologyInput groups loaded terminology targets for one translation task.
type TerminologyInput struct {
	TaskID    string
	FileNames []string
	NPCs      []TerminologyNPC
	Items     []TerminologyItem
	Magic     []TerminologyMagic
	Locations []TerminologyLocation
	Messages  []TerminologyMessage
	Quests    []TerminologyQuest
}

// TerminologyNPC represents one NPC row projected for terminology phase.
type TerminologyNPC struct {
	ID         string
	EditorID   string
	RecordType string
	Name       string
	SourceFile string
}

// TerminologyItem represents one item row projected for terminology phase.
type TerminologyItem struct {
	ID         string
	EditorID   string
	RecordType string
	Name       string
	Text       string
	SourceFile string
}

// TerminologyMagic represents one magic row projected for terminology phase.
type TerminologyMagic struct {
	ID         string
	EditorID   string
	RecordType string
	Name       string
	SourceFile string
}

// TerminologyLocation represents one location/cell/world row projected for terminology phase.
type TerminologyLocation struct {
	ID         string
	EditorID   string
	RecordType string
	Name       string
	SourceFile string
}

// TerminologyMessage represents one message row projected for terminology phase.
type TerminologyMessage struct {
	ID         string
	EditorID   string
	RecordType string
	Title      string
	SourceFile string
}

// TerminologyQuest represents one quest row projected for terminology phase.
type TerminologyQuest struct {
	ID         string
	EditorID   string
	RecordType string
	Name       string
	SourceFile string
}

// Repository defines artifact persistence operations for translation-flow input data.
type Repository interface {
	EnsureTask(ctx context.Context, taskID string) error
	SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (InputFile, error)
	ListFiles(ctx context.Context, taskID string) ([]InputFile, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error)
	LoadTerminologyInput(ctx context.Context, taskID string) (TerminologyInput, error)
}
