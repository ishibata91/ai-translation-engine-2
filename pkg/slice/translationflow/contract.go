package translationflow

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/artifact/translationinput"
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

// PersonaCandidateInput groups projected persona candidates and dialogue excerpts for one task.
type PersonaCandidateInput struct {
	TaskID     string
	Candidates map[string]PersonaCandidate
	Dialogues  []PersonaDialogueExcerpt
}

// PersonaCandidate represents one normalized NPC candidate for persona planning.
type PersonaCandidate struct {
	SpeakerID      string
	SourceRecordID string
	NPCKey         string
	EditorID       string
	RecordType     string
	NPCName        string
	Race           string
	Sex            string
	VoiceType      string
	SourcePlugin   string
	SourceHint     string
}

// PersonaDialogueExcerpt represents one dialogue row used for persona candidate planning.
type PersonaDialogueExcerpt struct {
	ID               string
	SpeakerID        string
	EditorID         string
	GroupEditorID    string
	RecordType       string
	Text             string
	QuestID          string
	SourcePlugin     string
	SourceHint       string
	IsServicesBranch bool
	Order            int
}

// PersonaFinalSummary represents the minimum final persona details needed by translation workflow.
type PersonaFinalSummary struct {
	PersonaID    int64
	SourcePlugin string
	SpeakerID    string
	PersonaText  string
}

// PersonaLookupKey identifies one persona final by source plugin and speaker ID.
type PersonaLookupKey struct {
	SourcePlugin string
	SpeakerID    string
}

// Service defines translation-flow input storage operations exposed to workflow.
type Service interface {
	EnsureTask(ctx context.Context, taskID string) error
	SaveParsedOutput(ctx context.Context, taskID string, sourceFilePath string, output *skyrim.ParserOutput) (LoadedFile, error)
	ListFiles(ctx context.Context, taskID string) ([]LoadedFile, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (PreviewPage, error)
	LoadTerminologyInput(ctx context.Context, taskID string) (translationinput.TerminologyInput, error)
	LoadPersonaCandidates(ctx context.Context, taskID string) (PersonaCandidateInput, error)
	FindPersonaFinal(ctx context.Context, key PersonaLookupKey) (PersonaFinalSummary, bool, error)
}
