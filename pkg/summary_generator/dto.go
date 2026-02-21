package summary_generator

import "time"

// DialogueGroupInput is the input for dialogue summary generation.
type DialogueGroupInput struct {
	GroupID    string   `json:"group_id"`
	PlayerText *string  `json:"player_text,omitempty"`
	Lines      []string `json:"lines"`
}

// QuestInput is the input for quest summary generation.
type QuestInput struct {
	QuestID    string   `json:"quest_id"`
	StageTexts []string `json:"stage_texts"`
}

// SummaryResult represents the result of a single summary generation.
type SummaryResult struct {
	RecordID    string `json:"record_id"`
	SummaryType string `json:"summary_type"`
	SummaryText string `json:"summary_text"`
	CacheHit    bool   `json:"cache_hit"`
}

// SummaryRecord represents a persisted summary row in SQLite.
type SummaryRecord struct {
	ID             int64     `json:"id"`
	RecordID       string    `json:"record_id"`
	SummaryType    string    `json:"summary_type"`
	CacheKey       string    `json:"cache_key"`
	InputHash      string    `json:"input_hash"`
	SummaryText    string    `json:"summary_text"`
	InputLineCount int       `json:"input_line_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CacheKeyHasher builds cache keys from record ID and input lines.
type CacheKeyHasher struct{}

// BuildCacheKey generates a cache key from the record ID and input lines.
func (h *CacheKeyHasher) BuildCacheKey(recordID string, lines []string) string {
	_ = recordID
	_ = lines
	panic("not implemented")
}
