package workflow

import (
	"context"
	"fmt"
	"regexp"
	"sort"
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
const personaProgressPhase = "persona"
const terminologyProgressPersistMaxUpdates = 20

var personaSourcePluginPattern = regexp.MustCompile(`(?i)[^\\/:*?"<>|]+\.(esm|esl|esp)`)

// TranslationFlowService orchestrates parser execution and artifact persistence for load phase.
type TranslationFlowService struct {
	parser          skyrim.Parser
	store           translationflow.Service
	terminology     terminologyslice.Terminology
	personaWorkflow MasterPersona
	executor        terminologyPhaseExecutor
	notifier        runtimeprogress.ProgressNotifier
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
	personaWorkflow MasterPersona,
	executor terminologyPhaseExecutor,
	notifier runtimeprogress.ProgressNotifier,
) *TranslationFlowService {
	return &TranslationFlowService{
		parser:          parser,
		store:           store,
		terminology:     terminology,
		personaWorkflow: personaWorkflow,
		executor:        executor,
		notifier:        notifier,
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

	targets, err := s.terminology.ListTargets(ctx, trimmedTaskID)
	if err != nil {
		return TerminologyTargetPreviewPage{}, fmt.Errorf("list terminology targets task_id=%s: %w", trimmedTaskID, err)
	}

	safePage := page
	if safePage <= 0 {
		safePage = 1
	}
	safePageSize := pageSize
	if safePageSize <= 0 {
		safePageSize = defaultTranslationPreviewPageSize
	}

	totalRows := len(targets)
	start := (safePage - 1) * safePageSize
	if start > totalRows {
		start = totalRows
	}
	end := start + safePageSize
	if end > totalRows {
		end = totalRows
	}

	rows := make([]TerminologyTargetPreviewRow, 0, end-start)
	pageEntries := targets[start:end]
	translations, err := s.terminology.GetPreviewTranslations(ctx, pageEntries)
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

// ListTranslationFlowPersonaTargets returns a paged persona-target preview for one task.
func (s *TranslationFlowService) ListTranslationFlowPersonaTargets(ctx context.Context, taskID string, page int, pageSize int) (PersonaTargetPreviewPage, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return PersonaTargetPreviewPage{}, fmt.Errorf("task_id is required")
	}

	plan, err := s.planTranslationFlowPersonaPhase(ctx, trimmedTaskID)
	if err != nil {
		return PersonaTargetPreviewPage{}, fmt.Errorf("plan translation flow persona targets task_id=%s: %w", trimmedTaskID, err)
	}

	safePage := page
	if safePage <= 0 {
		safePage = 1
	}
	safePageSize := pageSize
	if safePageSize <= 0 {
		safePageSize = defaultTranslationPreviewPageSize
	}

	totalRows := len(plan.Rows)
	start := (safePage - 1) * safePageSize
	if start > totalRows {
		start = totalRows
	}
	end := start + safePageSize
	if end > totalRows {
		end = totalRows
	}

	rows := make([]PersonaTargetPreviewRow, 0, end-start)
	rows = append(rows, plan.Rows[start:end]...)

	return PersonaTargetPreviewPage{
		TaskID:    trimmedTaskID,
		Page:      safePage,
		PageSize:  safePageSize,
		TotalRows: totalRows,
		Rows:      rows,
	}, nil
}

// RunTranslationFlowPersonaPhase runs or resumes persona generation for unresolved targets.
func (s *TranslationFlowService) RunTranslationFlowPersonaPhase(ctx context.Context, input RunTranslationFlowPersonaPhaseInput) (PersonaPhaseResult, error) {
	trimmedTaskID := strings.TrimSpace(input.TaskID)
	if trimmedTaskID == "" {
		return PersonaPhaseResult{}, fmt.Errorf("task_id is required")
	}

	plan, err := s.planTranslationFlowPersonaPhase(ctx, trimmedTaskID)
	if err != nil {
		return PersonaPhaseResult{}, fmt.Errorf("plan translation flow persona phase task_id=%s: %w", trimmedTaskID, err)
	}
	if plan.RetryableCount == 0 {
		s.reportPersonaProgress(ctx, plan.Summary)
		return plan.Summary, nil
	}

	if strings.TrimSpace(input.Request.Model) == "" {
		return PersonaPhaseResult{}, fmt.Errorf("request.model is required")
	}
	if s.personaWorkflow == nil {
		return PersonaPhaseResult{}, fmt.Errorf("persona workflow is not configured")
	}

	executionInput := PersonaExecutionInput{
		TaskID:  trimmedTaskID,
		Request: input.Request,
		Prompt:  input.Prompt,
	}
	if plan.RuntimeCount == 0 {
		sourceJSONPath, err := s.resolvePersonaBootstrapSourceJSONPath(ctx, trimmedTaskID)
		if err != nil {
			return PersonaPhaseResult{}, fmt.Errorf("resolve persona bootstrap source_json_path task_id=%s: %w", trimmedTaskID, err)
		}
		executionInput.SourceJSONPath = sourceJSONPath
	}

	runCtx := withPersonaPhaseRunConfig(ctx, input.Request, input.Prompt)
	if err := s.personaWorkflow.RunPersonaPhase(runCtx, executionInput); err != nil {
		return PersonaPhaseResult{}, fmt.Errorf("run persona workflow task_id=%s: %w", trimmedTaskID, err)
	}

	return s.GetTranslationFlowPersonaPhase(ctx, trimmedTaskID)
}

// GetTranslationFlowPersonaPhase returns the current persona phase summary.
func (s *TranslationFlowService) GetTranslationFlowPersonaPhase(ctx context.Context, taskID string) (PersonaPhaseResult, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return PersonaPhaseResult{}, fmt.Errorf("task_id is required")
	}

	plan, err := s.planTranslationFlowPersonaPhase(ctx, trimmedTaskID)
	if err != nil {
		return PersonaPhaseResult{}, fmt.Errorf("plan translation flow persona phase task_id=%s: %w", trimmedTaskID, err)
	}
	s.reportPersonaProgress(ctx, plan.Summary)
	return plan.Summary, nil
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
		baseSummary, err := s.terminology.GetPhaseSummary(ctx, trimmedTaskID)
		if err != nil {
			return TerminologyPhaseResult{}, fmt.Errorf("get prepared terminology summary task_id=%s: %w", trimmedTaskID, err)
		}
		targetCount := baseSummary.TargetCount
		if targetCount <= 0 {
			targetCount = len(requests)
		}
		startCurrent := baseSummary.SavedCount
		if startCurrent < 0 {
			startCurrent = 0
		}
		startSummary := terminologyslice.PhaseSummary{
			TaskID:          trimmedTaskID,
			Status:          "running",
			TargetCount:     targetCount,
			SavedCount:      baseSummary.SavedCount,
			FailedCount:     baseSummary.FailedCount,
			ProgressMode:    "determinate",
			ProgressCurrent: startCurrent,
			ProgressTotal:   targetCount,
			ProgressMessage: buildTerminologyProgressMessage(startCurrent, targetCount),
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
		responses, err := s.executeTerminologyWithProgress(
			ctx,
			trimmedTaskID,
			executionConfig,
			requests,
			startSummary.SavedCount,
			startSummary.TargetCount,
		)
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

type personaPhasePlan struct {
	Rows           []PersonaTargetPreviewRow
	Summary        PersonaPhaseResult
	RetryableCount int
	RuntimeCount   int
}

type personaTargetCandidate struct {
	SourcePlugin string
	SpeakerID    string
	EditorID     string
	NPCName      string
	Race         string
	Sex          string
	VoiceType    string
	Dialogues    []PersonaDialogueView
}

type personaRuntimeSnapshot struct {
	RuntimeByKey map[string]PersonaRuntimeEntry
	TotalCount   int
}

type personaPhaseRunConfig struct {
	Request TranslationRequestConfig
	Prompt  TranslationPromptConfig
}

type personaPhaseRunConfigContextKey struct{}

func (s *TranslationFlowService) planTranslationFlowPersonaPhase(ctx context.Context, taskID string) (personaPhasePlan, error) {
	input, err := s.store.LoadPersonaCandidates(ctx, taskID)
	if err != nil {
		return personaPhasePlan{}, fmt.Errorf("load persona candidates task_id=%s: %w", taskID, err)
	}
	candidates := collectPersonaCandidates(input)
	runtimeSnapshot, err := s.loadPersonaRuntimeSnapshot(ctx, taskID)
	if err != nil {
		return personaPhasePlan{}, fmt.Errorf("load persona runtime snapshot task_id=%s: %w", taskID, err)
	}

	keys := make([]string, 0, len(candidates))
	for key := range candidates {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	rows := make([]PersonaTargetPreviewRow, 0, len(keys))
	detectedCount := len(keys)
	reusedCount := 0
	pendingCount := 0
	generatedCount := 0
	failedCount := 0
	runningCount := 0
	retryableCount := 0

	for _, key := range keys {
		candidate := candidates[key]
		lookupKey := translationflow.PersonaLookupKey{
			SourcePlugin: candidate.SourcePlugin,
			SpeakerID:    candidate.SpeakerID,
		}
		finalPersona, hasFinal, err := s.store.FindPersonaFinal(ctx, lookupKey)
		if err != nil {
			return personaPhasePlan{}, fmt.Errorf("find final persona for planner source_plugin=%s speaker_id=%s: %w", lookupKey.SourcePlugin, lookupKey.SpeakerID, err)
		}

		runtimeEntry, hasRuntime := runtimeSnapshot.RuntimeByKey[key]
		viewState, errorMessage := resolvePersonaViewState(hasFinal, runtimeEntry, hasRuntime)
		if viewState == personaViewStateFailed {
			retryableCount++
		}
		if viewState == personaViewStatePending {
			retryableCount++
		}

		row := PersonaTargetPreviewRow{
			SourcePlugin: candidate.SourcePlugin,
			SpeakerID:    candidate.SpeakerID,
			EditorID:     candidate.EditorID,
			NPCName:      candidate.NPCName,
			Race:         candidate.Race,
			Sex:          candidate.Sex,
			VoiceType:    candidate.VoiceType,
			ViewState:    viewState,
			PersonaText:  "",
			ErrorMessage: errorMessage,
			Dialogues:    append([]PersonaDialogueView(nil), candidate.Dialogues...),
		}
		if hasFinal {
			row.PersonaText = finalPersona.PersonaText
		}
		rows = append(rows, row)

		switch viewState {
		case personaViewStateReused:
			reusedCount++
		case personaViewStatePending:
			pendingCount++
		case personaViewStateGenerated:
			generatedCount++
		case personaViewStateFailed:
			failedCount++
		case personaViewStateRunning:
			runningCount++
		}
	}

	summary := PersonaPhaseResult{
		TaskID:         taskID,
		DetectedCount:  detectedCount,
		ReusedCount:    reusedCount,
		PendingCount:   pendingCount,
		GeneratedCount: generatedCount,
		FailedCount:    failedCount,
	}
	summary.Status = derivePersonaPhaseStatus(summary, runningCount)
	summary.ProgressCurrent, summary.ProgressTotal = derivePersonaPhaseProgress(summary, runningCount)
	summary.ProgressMode, summary.ProgressMessage = derivePersonaPhaseProgressView(summary.Status, summary.ProgressCurrent, summary.ProgressTotal, summary.PendingCount)

	return personaPhasePlan{
		Rows:           rows,
		Summary:        summary,
		RetryableCount: retryableCount,
		RuntimeCount:   runtimeSnapshot.TotalCount,
	}, nil
}

func collectPersonaCandidates(input translationflow.PersonaCandidateInput) map[string]*personaTargetCandidate {
	candidates := make(map[string]*personaTargetCandidate, len(input.Candidates))
	for mapKey, npc := range input.Candidates {
		speakerID := strings.TrimSpace(npc.SpeakerID)
		if speakerID == "" {
			speakerID = strings.TrimSpace(mapKey)
		}
		if speakerID == "" {
			continue
		}
		sourcePlugin := normalizePersonaSourcePlugin(npc.SourcePlugin, npc.SourceHint)
		key := personaLookupCompositeKey(sourcePlugin, speakerID)
		candidate := ensurePersonaCandidate(candidates, key, sourcePlugin, speakerID)
		if candidate.EditorID == "" {
			candidate.EditorID = strings.TrimSpace(npc.EditorID)
		}
		if candidate.NPCName == "" {
			candidate.NPCName = strings.TrimSpace(npc.NPCName)
		}
		if candidate.Race == "" {
			candidate.Race = strings.TrimSpace(npc.Race)
		}
		if candidate.Sex == "" {
			candidate.Sex = strings.TrimSpace(npc.Sex)
		}
		if candidate.VoiceType == "" {
			candidate.VoiceType = strings.TrimSpace(npc.VoiceType)
		}
	}

	for _, dialogue := range input.Dialogues {
		speakerID := strings.TrimSpace(dialogue.SpeakerID)
		if speakerID == "" {
			continue
		}
		sourcePlugin := normalizePersonaSourcePlugin(dialogue.SourcePlugin, dialogue.SourceHint)
		key := personaLookupCompositeKey(sourcePlugin, speakerID)
		candidate := ensurePersonaCandidate(candidates, key, sourcePlugin, speakerID)
		if candidate.EditorID == "" {
			candidate.EditorID = strings.TrimSpace(dialogue.EditorID)
		}

		sourceText := strings.TrimSpace(dialogue.Text)
		if sourceText == "" {
			continue
		}
		candidate.Dialogues = append(candidate.Dialogues, PersonaDialogueView{
			RecordType:       strings.TrimSpace(dialogue.RecordType),
			EditorID:         strings.TrimSpace(dialogue.EditorID),
			SourceText:       sourceText,
			QuestID:          strings.TrimSpace(dialogue.QuestID),
			IsServicesBranch: dialogue.IsServicesBranch,
			Order:            dialogue.Order,
		})
	}

	for _, candidate := range candidates {
		sort.SliceStable(candidate.Dialogues, func(i, j int) bool {
			if candidate.Dialogues[i].Order != candidate.Dialogues[j].Order {
				return candidate.Dialogues[i].Order < candidate.Dialogues[j].Order
			}
			if candidate.Dialogues[i].RecordType != candidate.Dialogues[j].RecordType {
				return candidate.Dialogues[i].RecordType < candidate.Dialogues[j].RecordType
			}
			return candidate.Dialogues[i].EditorID < candidate.Dialogues[j].EditorID
		})
	}

	return candidates
}

func ensurePersonaCandidate(candidates map[string]*personaTargetCandidate, key string, sourcePlugin string, speakerID string) *personaTargetCandidate {
	candidate, ok := candidates[key]
	if ok {
		return candidate
	}
	candidate = &personaTargetCandidate{
		SourcePlugin: sourcePlugin,
		SpeakerID:    speakerID,
		Dialogues:    make([]PersonaDialogueView, 0),
	}
	candidates[key] = candidate
	return candidate
}

func (s *TranslationFlowService) loadPersonaRuntimeSnapshot(ctx context.Context, taskID string) (personaRuntimeSnapshot, error) {
	snapshot := personaRuntimeSnapshot{
		RuntimeByKey: make(map[string]PersonaRuntimeEntry),
	}
	if s.personaWorkflow == nil {
		return snapshot, nil
	}

	runtimeEntries, err := s.personaWorkflow.ListPersonaRuntime(ctx, taskID)
	if err != nil {
		return personaRuntimeSnapshot{}, fmt.Errorf("list persona runtime task_id=%s: %w", taskID, err)
	}
	snapshot.TotalCount = len(runtimeEntries)
	for _, runtimeEntry := range runtimeEntries {
		if !runtimeEntry.HasLookupKey {
			continue
		}
		speakerID := strings.TrimSpace(runtimeEntry.SpeakerID)
		if speakerID == "" {
			continue
		}
		sourcePlugin := normalizePersonaSourcePlugin(runtimeEntry.SourcePlugin, "")
		key := personaLookupCompositeKey(sourcePlugin, speakerID)
		existing, exists := snapshot.RuntimeByKey[key]
		if !exists || runtimeEntry.UpdatedAt.After(existing.UpdatedAt) {
			runtimeEntry.SourcePlugin = sourcePlugin
			runtimeEntry.SpeakerID = speakerID
			snapshot.RuntimeByKey[key] = runtimeEntry
		}
	}
	return snapshot, nil
}

func (s *TranslationFlowService) resolvePersonaBootstrapSourceJSONPath(ctx context.Context, taskID string) (string, error) {
	files, err := s.store.ListFiles(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("list translation-flow files task_id=%s: %w", taskID, err)
	}
	for _, file := range files {
		sourcePath := strings.TrimSpace(file.SourceFilePath)
		if sourcePath != "" {
			return sourcePath, nil
		}
	}
	return "", fmt.Errorf("source_json_path is not available")
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	raw, ok := metadata[key]
	if !ok {
		return ""
	}
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func withPersonaPhaseRunConfig(ctx context.Context, request TranslationRequestConfig, prompt TranslationPromptConfig) context.Context {
	return context.WithValue(ctx, personaPhaseRunConfigContextKey{}, personaPhaseRunConfig{
		Request: request,
		Prompt:  prompt,
	})
}

func personaPhaseRunConfigFromContext(ctx context.Context) (TranslationRequestConfig, TranslationPromptConfig, bool) {
	if ctx == nil {
		return TranslationRequestConfig{}, TranslationPromptConfig{}, false
	}
	raw := ctx.Value(personaPhaseRunConfigContextKey{})
	if raw == nil {
		return TranslationRequestConfig{}, TranslationPromptConfig{}, false
	}
	runConfig, ok := raw.(personaPhaseRunConfig)
	if !ok {
		return TranslationRequestConfig{}, TranslationPromptConfig{}, false
	}
	return runConfig.Request, runConfig.Prompt, true
}

const personaViewStateReused = "reused"
const personaViewStatePending = "pending"
const personaViewStateRunning = "running"
const personaViewStateGenerated = "generated"
const personaViewStateFailed = "failed"

const personaRuntimeStateRunning = "running"
const personaRuntimeStateCompleted = "completed"
const personaRuntimeStateFailed = "failed"
const personaRuntimeStateCanceled = "canceled"

func resolvePersonaViewState(hasFinal bool, runtimeEntry PersonaRuntimeEntry, hasRuntime bool) (string, string) {
	requestState := normalizePersonaRuntimeState(runtimeEntry.RequestState)
	if hasFinal {
		if hasRuntime && requestState == personaRuntimeStateCompleted {
			return personaViewStateGenerated, ""
		}
		return personaViewStateReused, ""
	}
	if !hasRuntime {
		return personaViewStatePending, ""
	}
	switch requestState {
	case personaRuntimeStateRunning:
		return personaViewStateRunning, ""
	case personaRuntimeStateFailed, personaRuntimeStateCanceled:
		return personaViewStateFailed, resolvePersonaErrorMessage(runtimeEntry)
	case personaRuntimeStateCompleted:
		return personaViewStateFailed, "生成結果の保存に失敗しました"
	default:
		return personaViewStatePending, ""
	}
}

func normalizePersonaRuntimeState(state string) string {
	return strings.ToLower(strings.TrimSpace(state))
}

func resolvePersonaErrorMessage(runtimeEntry PersonaRuntimeEntry) string {
	message := strings.TrimSpace(runtimeEntry.ErrorMessage)
	if message == "" {
		return "ペルソナ生成に失敗しました"
	}
	return message
}

func derivePersonaPhaseStatus(summary PersonaPhaseResult, runningCount int) string {
	if summary.DetectedCount == 0 {
		return "empty"
	}
	if summary.ReusedCount == summary.DetectedCount && summary.GeneratedCount == 0 && summary.PendingCount == 0 && summary.FailedCount == 0 && runningCount == 0 {
		return "cached_only"
	}
	if runningCount > 0 {
		return "running"
	}
	if summary.FailedCount > 0 && (summary.GeneratedCount > 0 || summary.ReusedCount > 0) {
		return "partial_failed"
	}
	if summary.FailedCount > 0 {
		return "failed"
	}
	if summary.PendingCount > 0 {
		return "ready"
	}
	return "completed"
}

func derivePersonaPhaseProgress(summary PersonaPhaseResult, runningCount int) (int, int) {
	total := summary.PendingCount + runningCount + summary.GeneratedCount + summary.FailedCount
	current := summary.GeneratedCount + summary.FailedCount
	if total < 0 {
		total = 0
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	return current, total
}

func derivePersonaPhaseProgressView(status string, progressCurrent int, progressTotal int, pendingCount int) (string, string) {
	switch status {
	case "empty":
		return "hidden", "ペルソナ対象 NPC はありません"
	case "cached_only":
		return "hidden", "既存 Master Persona を再利用します"
	case "ready":
		return "hidden", fmt.Sprintf("新規生成対象 %d 件", pendingCount)
	case "running", "completed", "partial_failed", "failed":
		if progressTotal <= 0 {
			return "hidden", ""
		}
		return "determinate", buildPersonaProgressMessage(progressCurrent, progressTotal)
	default:
		return "hidden", ""
	}
}

func buildPersonaProgressMessage(current int, total int) string {
	if total <= 0 {
		return "ペルソナ生成を実行中"
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

func personaLookupCompositeKey(sourcePlugin string, speakerID string) string {
	return strings.TrimSpace(sourcePlugin) + "|" + strings.TrimSpace(speakerID)
}

func normalizePersonaSourcePlugin(sourcePlugin string, sourceHint string) string {
	candidate := strings.TrimSpace(sourcePlugin)
	if candidate == "" {
		candidate = strings.TrimSpace(sourceHint)
	}
	match := personaSourcePluginPattern.FindString(candidate)
	if match != "" {
		return match
	}
	if strings.TrimSpace(sourcePlugin) != "" {
		return strings.TrimSpace(sourcePlugin)
	}
	return "UNKNOWN"
}

func (s *TranslationFlowService) reportPersonaProgress(ctx context.Context, summary PersonaPhaseResult) {
	if s.notifier == nil || strings.TrimSpace(summary.TaskID) == "" {
		return
	}

	status := runtimeprogress.StatusInProgress
	switch summary.Status {
	case "empty", "cached_only", "completed":
		status = runtimeprogress.StatusCompleted
	case "partial_failed", "failed":
		status = runtimeprogress.StatusFailed
	}

	s.notifier.OnProgress(ctx, runtimeprogress.ProgressEvent{
		CorrelationID: summary.TaskID,
		TaskID:        summary.TaskID,
		TaskType:      string(taskworkflow.TypeTranslationProject),
		Phase:         personaProgressPhase,
		Current:       summary.ProgressCurrent,
		Total:         summary.ProgressTotal,
		Completed:     summary.ProgressCurrent,
		Failed:        summary.FailedCount,
		Status:        status,
		Message:       summary.ProgressMessage,
	})
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
	baseSaved int,
	targetCount int,
) ([]llmio.Response, error) {
	executorWithProgress, ok := s.executor.(terminologyPhaseExecutorWithProgress)
	if !ok {
		return s.executor.Execute(ctx, config, requests)
	}

	requestCount := len(requests)
	total := targetCount
	if total <= 0 {
		total = requestCount
	}
	persistStride := terminologyProgressPersistStride(requestCount)
	lastPersisted := 0
	var progressMu sync.Mutex
	return executorWithProgress.ExecuteWithProgress(ctx, config, requests, func(completed, _ int) {
		progressMu.Lock()
		defer progressMu.Unlock()

		safeCompleted := completed
		if safeCompleted < 0 {
			safeCompleted = 0
		}
		if safeCompleted > requestCount {
			safeCompleted = requestCount
		}
		progressCurrent := baseSaved + safeCompleted
		if progressCurrent > total {
			progressCurrent = total
		}
		runningSummary := terminologyslice.PhaseSummary{
			TaskID:          taskID,
			Status:          "running",
			TargetCount:     total,
			ProgressMode:    "determinate",
			ProgressCurrent: progressCurrent,
			ProgressTotal:   total,
			ProgressMessage: buildTerminologyProgressMessage(progressCurrent, total),
		}

		if shouldPersistTerminologyProgress(safeCompleted, requestCount, lastPersisted, persistStride) {
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
