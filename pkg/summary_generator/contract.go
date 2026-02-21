package summary_generator

import "context"

// SummaryGenerator is the main entry point for dialogue and quest summary generation.
// It orchestrates LLM-based summarization and caching to SQLite.
type SummaryGenerator interface {
	GenerateDialogueSummaries(ctx context.Context, groups []DialogueGroupInput, progress func(done, total int)) ([]SummaryResult, error)
	GenerateQuestSummaries(ctx context.Context, quests []QuestInput, progress func(done, total int)) ([]SummaryResult, error)
}

// SummaryStore manages all operations on the per-source-file summaries SQLite table,
// including schema creation, cache lookup, and UPSERT.
type SummaryStore interface {
	InitTable(ctx context.Context) error
	Get(ctx context.Context, cacheKey string) (*SummaryRecord, error)
	Upsert(ctx context.Context, record SummaryRecord) error
	GetByRecordID(ctx context.Context, recordID string, summaryType string) (*SummaryRecord, error)
	Close() error
}
