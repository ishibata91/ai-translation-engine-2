package summary

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

// Summary is the main entry point for dialogue and quest summary generation.
// It follows a 2-phase model: ProposeJobs and SaveResults.
type Summary interface {
	// ProposeJobs analyzes inputs, checks cache, and returns LLM jobs for missing entries.
	ProposeJobs(ctx context.Context, input SummaryInput) (*ProposeOutput, error)

	// SaveResults persists LLM responses to the SQLite cache.
	SaveResults(ctx context.Context, responses []llm.Response) error

	// GetSummary retrieves a single summary by record ID. Used by Pass 2.
	GetSummary(ctx context.Context, recordID string, summaryType string) (*SummaryResult, error)
}

// SummaryStore manages all operations on the per-source-file summaries SQLite table.
type SummaryStore interface {
	// Init initializes the store, creating tables and setting PRAGMAs.
	Init(ctx context.Context) error

	// Get retrieves a record by its unique cache key.
	Get(ctx context.Context, cacheKey string) (*SummaryRecord, error)

	// GetByRecordID retrieves the latest record for a given record ID and type.
	GetByRecordID(ctx context.Context, recordID string, summaryType string) (*SummaryRecord, error)

	// Upsert inserts or updates a summary record.
	Upsert(ctx context.Context, record SummaryRecord) error

	// Close closes the underlying database connection.
	Close() error
}
