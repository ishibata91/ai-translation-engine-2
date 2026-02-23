package summary_generator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

// Record types
const (
	TypeDialogue = "Dialogue"
	TypeQuest    = "Quest"
)

// SummaryGeneratorConfig holds configuration for the SummaryGenerator.
type SummaryGeneratorConfig struct {
	// Concurrency controls the number of parallel goroutines for cache lookups in ProposeJobs.
	// Defaults to 10 if not set or set to 0.
	Concurrency int
}

// DefaultConcurrency is the default number of parallel goroutines.
const DefaultConcurrency = 10

// Effective returns the concurrency value, applying the default if needed.
func (c SummaryGeneratorConfig) Effective() int {
	if c.Concurrency <= 0 {
		return DefaultConcurrency
	}
	return c.Concurrency
}

// SummaryGeneratorInput is the independent DTO for summary generation.
type SummaryGeneratorInput struct {
	DialogueItems []DialogueItem `json:"dialogue_items"`
	QuestItems    []QuestItem    `json:"quest_items"`
}

// DialogueItem represents a dialogue group to be summarized.
type DialogueItem struct {
	GroupID    string   `json:"group_id"`
	PlayerText *string  `json:"player_text,omitempty"`
	Lines      []string `json:"lines"`
}

// QuestItem represents a quest with its stages to be summarized.
type QuestItem struct {
	QuestID    string       `json:"quest_id"`
	StageTexts []QuestStage `json:"stage_texts"`
}

// QuestStage represents a single stage text of a quest.
type QuestStage struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// ProposeOutput contains the results of ProposeJobs.
type ProposeOutput struct {
	Jobs                 []llm_client.Request `json:"jobs"`
	PreCalculatedResults []SummaryResult      `json:"pre_calculated_results"`
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
	SummaryType    string    `json:"summary_type"` // Dialogue or Quest
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
// Format: {record_id}|{sha256_hash}
func (h *CacheKeyHasher) BuildCacheKey(recordID string, lines []string) (string, string) {
	joined := strings.Join(lines, "\n")
	hash := sha256.Sum256([]byte(joined))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s|%s", recordID, hashStr), hashStr
}
