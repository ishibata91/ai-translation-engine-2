package dictionaryartifact

import (
	"context"
	"time"
)

// Source represents one shared dictionary source row persisted in artifact storage.
type Source struct {
	ID           int64
	FileName     string
	Format       string
	FilePath     string
	FileSize     int64
	EntryCount   int
	Status       string
	ErrorMessage string
	ImportedAt   *time.Time
	CreatedAt    time.Time
}

// Entry represents one shared dictionary entry row persisted in artifact storage.
type Entry struct {
	ID         int64
	SourceID   int64
	SourceName string
	EDID       string
	RecordType string
	SourceText string
	DestText   string
}

// EntryPage is one paged response for dictionary entries.
type EntryPage struct {
	Entries    []Entry
	TotalCount int
}

// Repository defines shared-dictionary artifact persistence operations.
type Repository interface {
	GetSources(ctx context.Context) ([]Source, error)
	CreateSource(ctx context.Context, source *Source) (int64, error)
	UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error
	DeleteSource(ctx context.Context, id int64) error
	FindExactBySourceText(ctx context.Context, text string) ([]Entry, error)
	FindExactBySourceTextCI(ctx context.Context, text string) ([]Entry, error)
	FindExactBySourceTexts(ctx context.Context, texts []string) ([]Entry, error)
	SearchBySourceTextLike(ctx context.Context, keyword string, limit int, npcOnly bool) ([]Entry, error)
	GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]Entry, error)
	GetEntriesBySourceIDPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, limit int, offset int) (*EntryPage, error)
	SearchAllEntriesPaginated(ctx context.Context, query string, filters map[string]string, limit int, offset int) (*EntryPage, error)
	SaveEntries(ctx context.Context, entries []Entry) error
	UpdateEntry(ctx context.Context, entry Entry) error
	DeleteEntry(ctx context.Context, id int64) error
}
