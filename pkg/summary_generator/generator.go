package summary_generator

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm_client"
)

const (
	DialogueSystemPrompt = `You are a professional game writer.
Summarize the following dialogue group into a single English sentence within 500 characters.
Focus on "who is talking to whom about what".
The summary will be used as context for translating subsequent dialogues.`

	QuestSystemPrompt = `You are a professional game writer.
Summarize the current quest status based on the cumulative stage descriptions.
Generate a single English sentence within 500 characters that captures the "story so far".
This summary will be used as context for translating quest-related texts.`
)

type summaryGenerator struct {
	store SummaryStore
}

func NewSummaryGenerator(store SummaryStore) SummaryGenerator {
	return &summaryGenerator{
		store: store,
	}
}

func (g *summaryGenerator) ProposeJobs(ctx context.Context, input SummaryGeneratorInput) (*ProposeOutput, error) {
	slog.DebugContext(ctx, "ENTER SummaryGenerator.ProposeJobs",
		slog.Int("dialogue_items", len(input.DialogueItems)),
		slog.Int("quest_items", len(input.QuestItems)),
	)
	defer slog.DebugContext(ctx, "EXIT SummaryGenerator.ProposeJobs")

	output := &ProposeOutput{
		Jobs:                 []llm_client.Request{},
		PreCalculatedResults: []SummaryResult{},
	}

	hasher := &CacheKeyHasher{}

	// Process Dialogue Items
	for _, item := range input.DialogueItems {
		if len(item.Lines) == 0 {
			continue
		}

		cacheKey, inputHash := hasher.BuildCacheKey(item.GroupID, item.Lines)
		record, err := g.store.Get(ctx, cacheKey)
		if err != nil {
			return nil, fmt.Errorf("failed to check cache for dialogue %s: %w", item.GroupID, err)
		}

		if record != nil {
			output.PreCalculatedResults = append(output.PreCalculatedResults, SummaryResult{
				RecordID:    item.GroupID,
				SummaryType: TypeDialogue,
				SummaryText: record.SummaryText,
				CacheHit:    true,
			})
			continue
		}

		// Cache MISS: Create LLM Job
		prompt := g.buildDialoguePrompt(item)
		output.Jobs = append(output.Jobs, llm_client.Request{
			SystemPrompt: DialogueSystemPrompt,
			UserPrompt:   prompt,
			Temperature:  0.3,
			Metadata: map[string]interface{}{
				"record_id":    item.GroupID,
				"summary_type": TypeDialogue,
				"cache_key":    cacheKey,
				"input_hash":   inputHash,
				"line_count":   len(item.Lines),
			},
		})
	}

	// Process Quest Items
	for _, item := range input.QuestItems {
		if len(item.StageTexts) == 0 {
			continue
		}

		// Sort stages by index
		sort.Slice(item.StageTexts, func(i, j int) bool {
			return item.StageTexts[i].Index < item.StageTexts[j].Index
		})

		// Cumulative processing: each stage gets a summary of all stages up to it
		currentLines := []string{}
		for _, stage := range item.StageTexts {
			if strings.TrimSpace(stage.Text) == "" {
				continue
			}
			currentLines = append(currentLines, stage.Text)

			// Unique ID for each stage summary
			stageRecordID := fmt.Sprintf("%s_stage_%d", item.QuestID, stage.Index)
			cacheKey, inputHash := hasher.BuildCacheKey(stageRecordID, currentLines)

			record, err := g.store.Get(ctx, cacheKey)
			if err != nil {
				return nil, fmt.Errorf("failed to check cache for quest stage %s: %w", stageRecordID, err)
			}

			if record != nil {
				output.PreCalculatedResults = append(output.PreCalculatedResults, SummaryResult{
					RecordID:    stageRecordID,
					SummaryType: TypeQuest,
					SummaryText: record.SummaryText,
					CacheHit:    true,
				})
				continue
			}

			// Cache MISS: Create LLM Job
			prompt := g.buildQuestPrompt(currentLines)
			output.Jobs = append(output.Jobs, llm_client.Request{
				SystemPrompt: QuestSystemPrompt,
				UserPrompt:   prompt,
				Temperature:  0.3,
				Metadata: map[string]interface{}{
					"record_id":    stageRecordID,
					"summary_type": TypeQuest,
					"cache_key":    cacheKey,
					"input_hash":   inputHash,
					"line_count":   len(currentLines),
				},
			})
		}
	}

	return output, nil
}

func (g *summaryGenerator) SaveResults(ctx context.Context, responses []llm_client.Response) error {
	slog.DebugContext(ctx, "ENTER SummaryGenerator.SaveResults", slog.Int("responses", len(responses)))
	defer slog.DebugContext(ctx, "EXIT SummaryGenerator.SaveResults")

	for _, resp := range responses {
		if !resp.Success {
			slog.WarnContext(ctx, "Skipping failed LLM response", slog.String("error", resp.Error))
			continue
		}

		recordID, _ := resp.Metadata["record_id"].(string)
		summaryType, _ := resp.Metadata["summary_type"].(string)
		cacheKey, _ := resp.Metadata["cache_key"].(string)
		inputHash, _ := resp.Metadata["input_hash"].(string)

		var lineCount int
		switch v := resp.Metadata["line_count"].(type) {
		case int:
			lineCount = v
		case float64:
			lineCount = int(v)
		case string:
			lineCount, _ = strconv.Atoi(v)
		}

		record := SummaryRecord{
			RecordID:       recordID,
			SummaryType:    summaryType,
			CacheKey:       cacheKey,
			InputHash:      inputHash,
			SummaryText:    resp.Content,
			InputLineCount: lineCount,
		}

		if err := g.store.Upsert(ctx, record); err != nil {
			return fmt.Errorf("failed to save summary result for %s: %w", recordID, err)
		}
	}

	return nil
}

func (g *summaryGenerator) GetSummary(ctx context.Context, recordID string, summaryType string) (*SummaryResult, error) {
	slog.DebugContext(ctx, "ENTER SummaryGenerator.GetSummary",
		slog.String("record_id", recordID),
		slog.String("type", summaryType),
	)
	defer slog.DebugContext(ctx, "EXIT SummaryGenerator.GetSummary")

	record, err := g.store.GetByRecordID(ctx, recordID, summaryType)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	return &SummaryResult{
		RecordID:    record.RecordID,
		SummaryType: record.SummaryType,
		SummaryText: record.SummaryText,
		CacheHit:    true,
	}, nil
}

func (g *summaryGenerator) buildDialoguePrompt(item DialogueItem) string {
	var sb strings.Builder
	if item.PlayerText != nil && *item.PlayerText != "" {
		sb.WriteString(fmt.Sprintf("Context (Player's choice/action): %s\n\n", *item.PlayerText))
	}
	sb.WriteString("Dialogue lines:\n")
	for _, line := range item.Lines {
		sb.WriteString("- ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (g *summaryGenerator) buildQuestPrompt(lines []string) string {
	return "Cumulative Quest Stages:\n" + strings.Join(lines, "\n")
}

type summaryStore struct {
	db *sql.DB
}

func NewSummaryStore(db *sql.DB) SummaryStore {
	return &summaryStore{
		db: db,
	}
}

func (s *summaryStore) Init(ctx context.Context) error {
	slog.DebugContext(ctx, "ENTER SummaryStore.Init")
	defer slog.DebugContext(ctx, "EXIT SummaryStore.Init")

	// Set PRAGMAs for performance and concurrency
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}
	for _, p := range pragmas {
		if _, err := s.db.ExecContext(ctx, p); err != nil {
			return fmt.Errorf("failed to set pragma %s: %w", p, err)
		}
	}

	// Create table
	query := `
	CREATE TABLE IF NOT EXISTS summaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id TEXT NOT NULL,
		summary_type TEXT NOT NULL,
		cache_key TEXT NOT NULL UNIQUE,
		input_hash TEXT NOT NULL,
		summary_text TEXT NOT NULL,
		input_line_count INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_summaries_cache_key ON summaries(cache_key);
	CREATE INDEX IF NOT EXISTS idx_summaries_record_id_type ON summaries(record_id, summary_type);
	`
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create summaries table: %w", err)
	}

	return nil
}

func (s *summaryStore) Get(ctx context.Context, cacheKey string) (*SummaryRecord, error) {
	query := `
	SELECT id, record_id, summary_type, cache_key, input_hash, summary_text, input_line_count, created_at, updated_at
	FROM summaries
	WHERE cache_key = ?
	`
	row := s.db.QueryRowContext(ctx, query, cacheKey)
	var r SummaryRecord
	err := row.Scan(&r.ID, &r.RecordID, &r.SummaryType, &r.CacheKey, &r.InputHash, &r.SummaryText, &r.InputLineCount, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *summaryStore) GetByRecordID(ctx context.Context, recordID string, summaryType string) (*SummaryRecord, error) {
	query := `
	SELECT id, record_id, summary_type, cache_key, input_hash, summary_text, input_line_count, created_at, updated_at
	FROM summaries
	WHERE record_id = ? AND summary_type = ?
	ORDER BY updated_at DESC
	LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, query, recordID, summaryType)
	var r SummaryRecord
	err := row.Scan(&r.ID, &r.RecordID, &r.SummaryType, &r.CacheKey, &r.InputHash, &r.SummaryText, &r.InputLineCount, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *summaryStore) Upsert(ctx context.Context, record SummaryRecord) error {
	query := `
	INSERT INTO summaries (record_id, summary_type, cache_key, input_hash, summary_text, input_line_count, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(cache_key) DO UPDATE SET
		summary_text = excluded.summary_text,
		updated_at = excluded.updated_at
	`
	now := time.Now()
	_, err := s.db.ExecContext(ctx, query,
		record.RecordID,
		record.SummaryType,
		record.CacheKey,
		record.InputHash,
		record.SummaryText,
		record.InputLineCount,
		now,
	)
	return err
}

func (s *summaryStore) Close() error {
	return s.db.Close()
}
