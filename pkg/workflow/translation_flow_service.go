package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	"github.com/ishibata91/ai-translation-engine-2/pkg/foundation/llmio"
	terminologyslice "github.com/ishibata91/ai-translation-engine-2/pkg/slice/terminology"
	"github.com/ishibata91/ai-translation-engine-2/pkg/slice/translationflow"
)

const defaultTranslationPreviewPageSize = 50

// TranslationFlowService orchestrates parser execution and artifact persistence for load phase.
type TranslationFlowService struct {
	parser      skyrim.Parser
	store       translationflow.Service
	terminology terminologyslice.Terminology
	executor    terminologyPhaseExecutor
}

type terminologyPhaseExecutor interface {
	Execute(ctx context.Context, config llmio.ExecutionConfig, requests []llmio.Request) ([]llmio.Response, error)
}

// NewTranslationFlowService constructs a translation-flow workflow implementation.
func NewTranslationFlowService(
	parser skyrim.Parser,
	store translationflow.Service,
	terminology terminologyslice.Terminology,
	executor terminologyPhaseExecutor,
) *TranslationFlowService {
	return &TranslationFlowService{
		parser:      parser,
		store:       store,
		terminology: terminology,
		executor:    executor,
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
	for _, entry := range input.Entries[start:end] {
		rows = append(rows, TerminologyTargetPreviewRow{
			ID:         entry.ID,
			RecordType: entry.RecordType,
			EditorID:   entry.EditorID,
			SourceText: entry.SourceText,
			Variant:    entry.Variant,
			SourceFile: entry.SourceFile,
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
		responses, err := s.executor.Execute(ctx, llmio.ExecutionConfig{
			Provider:        input.Request.Provider,
			Model:           input.Request.Model,
			Endpoint:        input.Request.Endpoint,
			APIKey:          input.Request.APIKey,
			Temperature:     input.Request.Temperature,
			ContextLength:   input.Request.ContextLength,
			SyncConcurrency: input.Request.SyncConcurrency,
			BulkStrategy:    input.Request.BulkStrategy,
		}, requests)
		if err != nil {
			return TerminologyPhaseResult{}, fmt.Errorf("execute terminology llm requests task_id=%s: %w", trimmedTaskID, err)
		}
		if err := s.terminology.SaveResults(ctx, trimmedTaskID, responses); err != nil {
			return TerminologyPhaseResult{}, fmt.Errorf("save terminology results task_id=%s: %w", trimmedTaskID, err)
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
		TaskID:      summary.TaskID,
		Status:      summary.Status,
		SavedCount:  summary.SavedCount,
		FailedCount: summary.FailedCount,
	}, nil
}
