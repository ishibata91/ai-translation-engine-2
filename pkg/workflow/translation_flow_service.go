package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	runtimeprogress "github.com/ishibata91/ai-translation-engine-2/pkg/foundation/progress"
	terminologyslice "github.com/ishibata91/ai-translation-engine-2/pkg/slice/terminology"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
	taskworkflow "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

const defaultTranslationPreviewPageSize = 50
const terminologyProgressPhase = "terminology"
const terminologyProgressPersistMaxUpdates = 20

// TranslationFlowService orchestrates parser execution and artifact persistence for load phase.
type TranslationFlowService struct {
	parser      skyrim.Parser
	store       translationflow.Service
	terminology terminologyslice.Terminology
	executor    terminologyPhaseExecutor
	notifier    runtimeprogress.ProgressNotifier
}

type terminologyPhaseExecutor interface {
	Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error)
}

type terminologyPhaseExecutorWithProgress interface {
	ExecuteWithProgress(
		ctx context.Context,
		config llmio.ExecutionConfig,
		requests []llmio.Request,
		progress func(completed, total int),
	) ([]llmio.Response, error)
}

// NewTranslationFlowService constructs a translation-flow workflow implementation.
func NewTranslationFlowService(
	parser skyrim.Parser,
	store translationflow.Service,
	terminology terminologyslice.Terminology,
	executor terminologyPhaseExecutor,
	notifier runtimeprogress.ProgressNotifier,
) *TranslationFlowService {
	return &TranslationFlowService{
		parser:      parser,
		store:       store,
		terminology: terminology,
		executor:    executor,
		notifier:    notifier,
	}
}

// LoadFiles parses selected files and stores them under the task boundary.
func (s *TranslationFlowService) LoadFiles(ctx context.Context, input LoadTranslationFlowInput) (TranslationLoadResult, error) {
	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return TranslationLoadResult{}, fmt.Errorf("task_id is required")
	}
	if len(input.FilePaths) == 0 {
		return TranslationLoadResult{}, fmt.Errorf("file_paths is required")
	}

	if err := s.store.EnsureTask(ctx, trimmedTaskID); err != nil {
		return TranslationLoadResult{}, fmt.Errorf("ensure translation-flow task task_id=%s: %w", trimmedTaskID, err)
	}

	for _, sourcePath := range input.FilePaths {
		trimmedPath := strings.TrimSpace(sourcePath)
		if trimmedPath == "" {
			continue
		}

		parsed, err := s.parser.LoadExtractedJSON(ctx, trimmedPath)
		if err != nil {
			return TranslationLoadResult{}, fmt.Errorf("parse source json task_id=%s file=%s: %w", trimmedTaskID, trimmedPath, err)
		}
		if _, err := s.store.SaveParsedOutput(ctx, trimmedTaskID, trimmedPath, parsed); err != nil {
			return TranslationLoadResult{}, fmt.Errorf("save parsed output task_id=%s file=%s: %w", trimmedTaskID, trimmedPath, err)
		}
	}

	if err := s.terminology.UpdatePhaseSummary(ctx, terminologyslice.PhaseSummary{
		TaskID:       trimmedTaskID,
		Status:       "pending",
		ProgressMode: "hidden",
	}); err != nil {
		return TranslationLoadResult{}, fmt.Errorf("reset terminology phase summary task_id=%s: %w", trimmedTaskID, err)
	}

	return s.ListFiles(ctx, trimmedTaskID)
}

// ListFiles returns loaded files with first preview page for each file.
func (s *TranslationFlowService) ListFiles(ctx context.Context, taskID string) (TranslationLoadResult, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return TranslationLoadResult{}, fmt.Errorf("task_id is required")
	}

	files, err := s.store.ListFiles(ctx, trimmedTaskID)
	if err != nil {
		return TranslationLoadResult{}, fmt.Errorf("list translation-flow files task_id=%s: %w", trimmedTaskID, err)
	}

	loadedFiles := make([]TranslationLoadedFile, 0, len(files))
	for _, file := range files {
		previewPage, err := s.store.ListPreviewRows(ctx, file.ID, 1, defaultTranslationPreviewPageSize)
		if err != nil {
			return TranslationLoadResult{}, fmt.Errorf("list file preview task_id=%s file_id=%d: %w", trimmedTaskID, file.ID, err)
		}
		loadedFiles = append(loadedFiles, TranslationLoadedFile{
			FileID:       file.ID,
			FilePath:     file.SourceFilePath,
			FileName:     file.SourceFileName,
			ParseStatus:  file.ParseStatus,
			PreviewCount: file.PreviewRowCount,
			Preview:      mapPreviewPage(previewPage),
		})
	}

	return TranslationLoadResult{
		TaskID: trimmedTaskID,
		Files:  loadedFiles,
	}, nil
}

// ListPreviewRows returns one paged preview response for one file.
func (s *TranslationFlowService) ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (TranslationPreviewPage, error) {
	previewPage, err := s.store.ListPreviewRows(ctx, fileID, page, pageSize)
	if err != nil {
		return TranslationPreviewPage{}, fmt.Errorf("list preview rows file_id=%d page=%d size=%d: %w", fileID, page, pageSize, err)
	}
	return mapPreviewPage(previewPage), nil
}

func mapPreviewPage(page translationflow.PreviewPage) TranslationPreviewPage {
	rows := make([]TranslationPreviewRow, 0, len(page.Rows))
	for _, row := range page.Rows {
		rows = append(rows, TranslationPreviewRow{
			ID:         row.ID,
			Section:    row.Section,
			RecordType: row.RecordType,
			EditorID:   row.EditorID,
			SourceText: row.SourceText,
		})
	}
	return TranslationPreviewPage{
		FileID:    page.FileID,
		Page:      page.Page,
		PageSize:  page.PageSize,
		TotalRows: page.TotalRows,
		Rows:      rows,
	}
}

// ListTerminologyTargets returns a paged terminology-target preview for one task.
func (s *TranslationFlowService) ListTerminologyTargets(ctx context.Context, taskID string, page int, pageSize int) (TerminologyTargetPreviewPage, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return TerminologyTargetPreviewPage{}, fmt.Errorf("task_id is required")
	}

	input, err := s.store.LoadTerminologyInput(ctx, trimmedTaskID)
	if err != nil {
		return TerminologyTargetPreviewPage{}, fmt.Errorf("load terminology input task_id=%s: %w", trimmedTaskID, err)
	}

	safePage := page
	if safePage <= 0 {
		safePage = 1
	}
	safePageSize := pageSize
	if safePageSize <= 0 {
		safePageSize = defaultTranslationPreviewPageSize
	}

	totalRows := len(input.Entries)
	start := (safePage - 1) * safePageSize
	if start > totalRows {
		start = totalRows
	}
	end := start + safePageSize
	if end > totalRows {
		end = totalRows
	}

	rows := make([]TerminologyTargetPreviewRow, 0, end-start)
	pageEntries := input.Entries[start:end]
	previewEntries := make([]terminologyslice.TerminologyEntry, 0, len(pageEntries))
	for _, entry := range pageEntries {
		previewEntries = append(previewEntries, terminologyslice.TerminologyEntry{
			ID:         entry.ID,
			EditorID:   entry.EditorID,
			RecordType: entry.RecordType,
			SourceText: entry.SourceText,
			SourceFile: entry.SourceFile,
			PairKey:    entry.PairKey,
			Variant:    entry.Variant,
		})
	}
	translations, err := s.terminology.GetPreviewTranslations(ctx, previewEntries)
	if err != nil {
		return TerminologyTargetPreviewPage{}, fmt.Errorf("get terminology preview translations task_id=%s: %w", trimmedTaskID, err)
	}

	for _, entry := range pageEntries {
		translation := translations[entry.ID]
		rows = append(rows, TerminologyTargetPreviewRow{
			ID:               entry.ID,
			RecordType:       entry.RecordType,
			EditorID:         entry.EditorID,
			SourceText:       entry.SourceText,
			TranslatedText:   translation.TranslatedText,
			TranslationState: translation.TranslationState,
			Variant:          entry.Variant,
			SourceFile:       entry.SourceFile,
		})
	}

	return TerminologyTargetPreviewPage{
		TaskID:    trimmedTaskID,
		Page:      safePage,
		PageSize:  safePageSize,
		TotalRows: totalRows,
		Rows:      rows,
	}, nil
}

// RunTerminologyPhase executes the terminology phase synchronously and returns the persisted summary.
func (s *TranslationFlowService) RunTerminologyPhase(ctx context.Context, input RunTerminologyPhaseInput) (TerminologyPhaseResult, error) {
	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return TerminologyPhaseResult{}, fmt.Errorf("task_id is required")
	}
	if strings.TrimSpace(input.Request.Model) == "" {
		return TerminologyPhaseResult{}, fmt.Errorf("request.model is required")
	}

	requests, err := s.terminology.PreparePrompts(ctx, trimmedTaskID, terminologyslice.PhaseOptions{
		Request: terminologyslice.RequestConfig{
			Provider:        input.Request.Provider,
			Model:           input.Request.Model,
			Endpoint:        input.Request.Endpoint,
			APIKey:          input.Request.APIKey,
			Temperature:     input.Request.Temperature,
			ContextLength:   input.Request.ContextLength,
			SyncConcurrency: input.Request.SyncConcurrency,
			BulkStrategy:    input.Request.BulkStrategy,
		},
		Prompt: terminologyslice.PromptConfig{
			UserPrompt:   input.Prompt.UserPrompt,
			SystemPrompt: input.Prompt.SystemPrompt,
		},
	})
	if err != nil {
		return TerminologyPhaseResult{}, fmt.Errorf("prepare terminology prompts task_id=%s: %w", trimmedTaskID, err)
	}

	if len(requests) > 0 {
		startSummary := terminologyslice.PhaseSummary{
			TaskID:          trimmedTaskID,
			Status:          "running",
			TargetCount:     len(requests),
			ProgressMode:    "determinate",
			ProgressCurrent: 0,
			ProgressTotal:   len(requests),
			ProgressMessage: buildTerminologyProgressMessage(0, len(requests)),
		}
		if err := s.terminology.UpdatePhaseSummary(ctx, startSummary); err != nil {
			return TerminologyPhaseResult{}, fmt.Errorf("update running terminology summary task_id=%s: %w", trimmedTaskID, err)
		}
		s.reportTerminologyProgress(ctx, startSummary)
		executionConfig := llmio.ExecutionConfig{
			Provider:        input.Request.Provider,
			Model:           input.Request.Model,
			Endpoint:        input.Request.Endpoint,
			APIKey:          input.Request.APIKey,
			Temperature:     input.Request.Temperature,
			ContextLength:   input.Request.ContextLength,
			SyncConcurrency: input.Request.SyncConcurrency,
			BulkStrategy:    input.Request.BulkStrategy,
		}
		responses, err := s.executeTerminologyWithProgress(ctx, trimmedTaskID, executionConfig, requests)
		if err != nil {
			summary, summaryErr := s.terminology.GetPhaseSummary(ctx, trimmedTaskID)
			if summaryErr == nil {
				runErrorSummary := terminologyslice.PhaseSummary{
					TaskID:          trimmedTaskID,
					Status:          "run_error",
					TargetCount:     summary.TargetCount,
					SavedCount:      summary.SavedCount,
					FailedCount:     summary.FailedCount,
					ProgressMode:    "hidden",
					ProgressCurrent: summary.ProgressCurrent,
					ProgressTotal:   summary.ProgressTotal,
					ProgressMessage: "単語翻訳の実行に失敗しました",
				}
				_ = s.terminology.UpdatePhaseSummary(ctx, runErrorSummary)
				s.reportTerminologyProgress(ctx, runErrorSummary)
			}
			return TerminologyPhaseResult{}, fmt.Errorf("execute terminology llm requests task_id=%s: %w", trimmedTaskID, err)
		}
		if err := s.terminology.SaveResults(ctx, trimmedTaskID, responses); err != nil {
			return TerminologyPhaseResult{}, fmt.Errorf("save terminology results task_id=%s: %w", trimmedTaskID, err)
		}
		if summary, summaryErr := s.terminology.GetPhaseSummary(ctx, trimmedTaskID); summaryErr == nil {
			s.reportTerminologyProgress(ctx, summary)
		}
	}

	return s.GetTerminologyPhase(ctx, trimmedTaskID)
}

// GetTerminologyPhase returns the current terminology phase summary.
func (s *TranslationFlowService) GetTerminologyPhase(ctx context.Context, taskID string) (TerminologyPhaseResult, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return TerminologyPhaseResult{}, fmt.Errorf("task_id is required")
	}
	summary, err := s.terminology.GetPhaseSummary(ctx, trimmedTaskID)
	if err != nil {
		return TerminologyPhaseResult{}, fmt.Errorf("get terminology phase summary task_id=%s: %w", trimmedTaskID, err)
	}
	return TerminologyPhaseResult{
		TaskID:          summary.TaskID,
		Status:          summary.Status,
		SavedCount:      summary.SavedCount,
		FailedCount:     summary.FailedCount,
		ProgressMode:    summary.ProgressMode,
		ProgressCurrent: summary.ProgressCurrent,
		ProgressTotal:   summary.ProgressTotal,
		ProgressMessage: summary.ProgressMessage,
	}, nil
}

func (s *TranslationFlowService) reportTerminologyProgress(ctx context.Context, summary terminologyslice.PhaseSummary) {
	if s.notifier == nil || strings.TrimSpace(summary.TaskID) == "" {
		return
	}

	status := runtimeprogress.StatusInProgress
	switch summary.Status {
	case "completed":
		status = runtimeprogress.StatusCompleted
	case "completed_partial", "run_error":
		status = runtimeprogress.StatusFailed
	}

	s.notifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
		CorrelationID: summary.TaskID,
		TaskID:        summary.TaskID,
		TaskType:      string(taskworkflow.TypeTranslationProject),
		Phase:         terminologyProgressPhase,
		Current:       summary.ProgressCurrent,
		Total:         summary.ProgressTotal,
		Completed:     summary.ProgressCurrent,
		Failed:        summary.FailedCount,
		Status:        status,
		Message:       summary.ProgressMessage,
	})
}

func (s *TranslationFlowService) executeTerminologyWithProgress(
	ctx context.Context,
	taskID string,
	config llmio.ExecutionConfig,
	requests []llmio.Request,
) ([]llmio.Response, error) {
	executorWithProgress, ok := s.executor.(terminologyPhaseExecutorWithProgress)
	if !ok {
		return s.executor.Execute(ctx, config, requests)
	}

	total := len(requests)
	persistStride := terminologyProgressPersistStride(total)
	lastPersisted := 0
	var progressMu sync.Mutex
	return executorWithProgress.ExecuteWithProgress(ctx, config, requests, func(completed, _ int) {
		progressMu.Lock()
		defer progressMu.Unlock()

		safeCompleted := completed
		if safeCompleted < 0 {
			safeCompleted = 0
		}
		if safeCompleted > total {
			safeCompleted = total
		}
		runningSummary := terminologyslice.PhaseSummary{
			TaskID:          taskID,
			Status:          "running",
			TargetCount:     total,
			ProgressMode:    "determinate",
			ProgressCurrent: safeCompleted,
			ProgressTotal:   total,
			ProgressMessage: buildTerminologyProgressMessage(safeCompleted, total),
		}

		if shouldPersistTerminologyProgress(safeCompleted, total, lastPersisted, persistStride) {
			_ = s.terminology.UpdatePhaseSummary(ctx, runningSummary)
			lastPersisted = safeCompleted
		}
		s.reportTerminologyProgress(ctx, runningSummary)
	})
}

func terminologyProgressPersistStride(total int) int {
	if total <= 1 {
		return 1
	}
	stride := total / terminologyProgressPersistMaxUpdates
	if total%terminologyProgressPersistMaxUpdates != 0 {
		stride++
	}
	// Always throttle intermediate persistence updates.
	if stride < 2 {
		return 2
	}
	return stride
}

func shouldPersistTerminologyProgress(current int, total int, lastPersisted int, stride int) bool {
	if current <= 0 {
		return false
	}
	// Completion summary is persisted by result saving path.
	if current >= total {
		return false
	}
	if stride < 2 {
		stride = 2
	}
	return current-lastPersisted >= stride
}

func buildTerminologyProgressMessage(current int, total int) string {
	if total <= 0 {
		return "単語翻訳を実行中"
	}
	safeCurrent := current
	if safeCurrent < 0 {
		safeCurrent = 0
	}
	if safeCurrent > total {
		safeCurrent = total
	}
	remaining := total - safeCurrent
	return fmt.Sprintf("%d / %d 件（残り %d 件）", safeCurrent, total, remaining)
}
