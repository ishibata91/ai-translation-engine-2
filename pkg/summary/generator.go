package summary

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
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
	store  SummaryStore
	config SummaryConfig
}

func NewSummaryGenerator(store SummaryStore, config SummaryConfig) Summary {
	return &summaryGenerator{
		store:  store,
		config: config,
	}
}

// dialogueCacheResult holds the outcome of a single parallel cache lookup.
type dialogueCacheResult struct {
	index  int // preserves original order
	job    *llm.Request
	result *SummaryResult
}

func (g *summaryGenerator) ID() string {
	return "Summary"
}

func (g *summaryGenerator) PreparePrompts(ctx context.Context, input any) ([]llm.Request, error) {
	typedInput, ok := input.(SummaryInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for Summary slice: %T", input)
	}
	output, err := g.ProposeJobs(ctx, typedInput)
	if err != nil {
		return nil, err
	}
	return output.Jobs, nil
}

func (g *summaryGenerator) ProposeJobs(ctx context.Context, input SummaryInput) (*ProposeOutput, error) {
	start := time.Now()
	slog.DebugContext(ctx, "ENTER Summary.ProposeJobs",
		slog.Int("dialogue_items", len(input.DialogueItems)),
		slog.Int("quest_items", len(input.QuestItems)),
	)
	defer func() {
		slog.DebugContext(ctx, "EXIT Summary.ProposeJobs",
			slog.Duration("elapsed", time.Since(start)),
		)
	}()

	output := &ProposeOutput{
		Jobs:                 []llm.Request{},
		PreCalculatedResults: []SummaryResult{},
	}

	// ── Phase 1: Dialogue – parallel cache lookup ──────────────────────────
	if err := g.proposeDialogueJobs(ctx, input.DialogueItems, output); err != nil {
		return nil, err
	}

	// ── Phase 2: Quest – sequential (stage order must be preserved) ─────────
	if err := g.proposeQuestJobs(ctx, input.QuestItems, output); err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "ProposeJobs result",
		slog.Int("jobs", len(output.Jobs)),
		slog.Int("cache_hits", len(output.PreCalculatedResults)),
	)
	return output, nil
}

// proposeDialogueJobs runs cache lookups in parallel (bounded by config.Concurrency),
// then appends Jobs/PreCalculatedResults in original order.
func (g *summaryGenerator) proposeDialogueJobs(ctx context.Context, items []DialogueItem, output *ProposeOutput) error {
	if len(items) == 0 {
		return nil
	}

	concurrency := g.config.Effective()
	sem := make(chan struct{}, concurrency)

	results := make([]dialogueCacheResult, len(items))
	hasher := &CacheKeyHasher{}

	eg, egCtx := errgroup.WithContext(ctx)

	for i, item := range items {
		i, item := i, item // capture loop vars
		if len(item.Lines) == 0 {
			// mark as skip (both job and result are nil)
			results[i] = dialogueCacheResult{index: i}
			continue
		}

		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			cacheKey, inputHash := hasher.BuildCacheKey(item.GroupID, item.Lines)
			record, err := g.store.Get(egCtx, cacheKey)
			if err != nil {
				return fmt.Errorf("failed to check cache for dialogue %s: %w", item.GroupID, err)
			}

			r := dialogueCacheResult{index: i}
			if record != nil {
				r.result = &SummaryResult{
					RecordID:    item.GroupID,
					SummaryType: TypeDialogue,
					SummaryText: record.SummaryText,
					CacheHit:    true,
				}
			} else {
				prompt := buildDialoguePrompt(item)
				req := llm.Request{
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
				}
				r.job = &req
			}
			results[i] = r
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	// Collect results in original order
	for _, r := range results {
		if r.result != nil {
			output.PreCalculatedResults = append(output.PreCalculatedResults, *r.result)
		} else if r.job != nil {
			output.Jobs = append(output.Jobs, *r.job)
		}
		// nil both = empty lines, skip
	}
	return nil
}

// proposeQuestJobs processes quest items sequentially (cumulative stage order).
// Different quests are processed in parallel; stages within a quest are sequential.
func (g *summaryGenerator) proposeQuestJobs(ctx context.Context, items []QuestItem, output *ProposeOutput) error {
	if len(items) == 0 {
		return nil
	}

	type questResult struct {
		jobs    []llm.Request
		results []SummaryResult
	}

	concurrency := g.config.Effective()
	sem := make(chan struct{}, concurrency)
	questResults := make([]questResult, len(items))
	hasher := &CacheKeyHasher{}

	eg, egCtx := errgroup.WithContext(ctx)

	for i, item := range items {
		i, item := i, item
		if len(item.StageTexts) == 0 {
			questResults[i] = questResult{}
			continue
		}

		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Sort stages by index (must be sequential within a quest)
			stages := make([]QuestStage, len(item.StageTexts))
			copy(stages, item.StageTexts)
			sort.Slice(stages, func(a, b int) bool {
				return stages[a].Index < stages[b].Index
			})

			var qr questResult
			currentLines := []string{}

			for _, stage := range stages {
				if strings.TrimSpace(stage.Text) == "" {
					continue
				}
				currentLines = append(currentLines, stage.Text)

				stageRecordID := fmt.Sprintf("%s_stage_%d", item.QuestID, stage.Index)
				cacheKey, inputHash := hasher.BuildCacheKey(stageRecordID, currentLines)

				record, err := g.store.Get(egCtx, cacheKey)
				if err != nil {
					return fmt.Errorf("failed to check cache for quest stage %s: %w", stageRecordID, err)
				}

				if record != nil {
					qr.results = append(qr.results, SummaryResult{
						RecordID:    stageRecordID,
						SummaryType: TypeQuest,
						SummaryText: record.SummaryText,
						CacheHit:    true,
					})
					continue
				}

				prompt := buildQuestPrompt(currentLines)
				qr.jobs = append(qr.jobs, llm.Request{
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

			questResults[i] = qr
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	// Collect in original order
	for _, qr := range questResults {
		output.Jobs = append(output.Jobs, qr.jobs...)
		output.PreCalculatedResults = append(output.PreCalculatedResults, qr.results...)
	}
	return nil
}

func (g *summaryGenerator) SaveResults(ctx context.Context, responses []llm.Response) error {
	start := time.Now()
	slog.DebugContext(ctx, "ENTER Summary.SaveResults", slog.Int("responses", len(responses)))
	defer func() {
		slog.DebugContext(ctx, "EXIT Summary.SaveResults",
			slog.Duration("elapsed", time.Since(start)),
		)
	}()

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

	slog.DebugContext(ctx, "SaveResults complete", slog.Int("saved", len(responses)))
	return nil
}

func (g *summaryGenerator) GetSummary(ctx context.Context, recordID string, summaryType string) (*SummaryResult, error) {
	start := time.Now()
	slog.DebugContext(ctx, "ENTER Summary.GetSummary",
		slog.String("record_id", recordID),
		slog.String("type", summaryType),
	)
	defer func() {
		slog.DebugContext(ctx, "EXIT Summary.GetSummary",
			slog.Duration("elapsed", time.Since(start)),
		)
	}()

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

// ── Private prompt builders ──────────────────────────────────────────────────

func buildDialoguePrompt(item DialogueItem) string {
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

func buildQuestPrompt(lines []string) string {
	return "Cumulative Quest Stages:\n" + strings.Join(lines, "\n")
}

// ── SummaryStore implementation ──────────────────────────────────────────────

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
