package controller

import (
	"context"
	"fmt"

	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
	task2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/task"
)

type taskManager interface {
	GetActiveTasks() []task2.Task
	GetAllTasks(ctx context.Context) ([]task2.Task, error)
	ResumeTask(taskID string) error
	DeleteTask(ctx context.Context, taskID string) error
	CancelTask(taskID string)
	EnsureTranslationProjectTask(ctx context.Context, taskID string) (string, error)
}

type translationFlowWorkflow interface {
	LoadFiles(ctx context.Context, input workflow.LoadTranslationFlowInput) (workflow.TranslationLoadResult, error)
	ListFiles(ctx context.Context, taskID string) (workflow.TranslationLoadResult, error)
	ListPreviewRows(ctx context.Context, fileID int64, page int, pageSize int) (workflow.TranslationPreviewPage, error)
	RunTerminologyPhase(ctx context.Context, input workflow.RunTerminologyPhaseInput) (workflow.TerminologyPhaseResult, error)
	GetTerminologyPhase(ctx context.Context, taskID string) (workflow.TerminologyPhaseResult, error)
}

// TaskController exposes generic Wails-facing task operations.
type TaskController struct {
	ctx             context.Context
	manager         taskManager
	translationFlow translationFlowWorkflow
}

// NewTaskController constructs the task controller adapter.
func NewTaskController(manager taskManager) *TaskController {
	return &TaskController{
		ctx:     context.Background(),
		manager: manager,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *TaskController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// SetTranslationFlowWorkflow injects translation-flow workflow APIs.
func (c *TaskController) SetTranslationFlowWorkflow(translationFlow translationFlowWorkflow) {
	c.translationFlow = translationFlow
}

// GetActiveTasks returns in-memory active tasks for dashboard polling.
func (c *TaskController) GetActiveTasks() []task2.Task {
	return c.manager.GetActiveTasks()
}

// GetAllTasks loads all persisted tasks.
func (c *TaskController) GetAllTasks() ([]task2.Task, error) {
	return c.manager.GetAllTasks(c.ctx)
}

// ResumeTask resumes a generic task through task manager.
func (c *TaskController) ResumeTask(taskID string) error {
	return c.manager.ResumeTask(taskID)
}

// DeleteTask deletes a persisted task through task manager.
func (c *TaskController) DeleteTask(taskID string) error {
	return c.manager.DeleteTask(c.ctx, taskID)
}

// CancelTask cancels a generic task through task manager.
func (c *TaskController) CancelTask(taskID string) {
	c.manager.CancelTask(taskID)
}

// LoadTranslationFlowFiles parses and saves selected files under one translation project task.
func (c *TaskController) LoadTranslationFlowFiles(taskID string, filePaths []string) (workflow.TranslationLoadResult, error) {
	if c.translationFlow == nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("translation flow workflow is not configured")
	}
	resolvedTaskID, err := c.manager.EnsureTranslationProjectTask(c.ctx, taskID)
	if err != nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("ensure translation project task task_id=%s: %w", taskID, err)
	}
	result, err := c.translationFlow.LoadFiles(c.ctx, workflow.LoadTranslationFlowInput{
		TaskID:    resolvedTaskID,
		FilePaths: filePaths,
	})
	if err != nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("load translation flow files task_id=%s: %w", resolvedTaskID, err)
	}
	return result, nil
}

// ListLoadedTranslationFlowFiles returns loaded files and first preview page for each file.
func (c *TaskController) ListLoadedTranslationFlowFiles(taskID string) (workflow.TranslationLoadResult, error) {
	if c.translationFlow == nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("translation flow workflow is not configured")
	}
	resolvedTaskID, err := c.manager.EnsureTranslationProjectTask(c.ctx, taskID)
	if err != nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("ensure translation project task task_id=%s: %w", taskID, err)
	}
	result, err := c.translationFlow.ListFiles(c.ctx, resolvedTaskID)
	if err != nil {
		return workflow.TranslationLoadResult{}, fmt.Errorf("list loaded translation flow files task_id=%s: %w", resolvedTaskID, err)
	}
	return result, nil
}

// ListTranslationFlowPreviewRows returns one paged preview response for a file.
func (c *TaskController) ListTranslationFlowPreviewRows(fileID int64, page int, pageSize int) (workflow.TranslationPreviewPage, error) {
	if c.translationFlow == nil {
		return workflow.TranslationPreviewPage{}, fmt.Errorf("translation flow workflow is not configured")
	}
	preview, err := c.translationFlow.ListPreviewRows(c.ctx, fileID, page, pageSize)
	if err != nil {
		return workflow.TranslationPreviewPage{}, fmt.Errorf("list translation flow preview rows file_id=%d page=%d page_size=%d: %w", fileID, page, pageSize, err)
	}
	return preview, nil
}

// RunTranslationFlowTerminology executes the terminology phase for one translation project task.
func (c *TaskController) RunTranslationFlowTerminology(taskID string, request workflow.TranslationRequestConfig, prompt workflow.TranslationPromptConfig) (workflow.TerminologyPhaseResult, error) {
	if c.translationFlow == nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("translation flow workflow is not configured")
	}
	resolvedTaskID, err := c.manager.EnsureTranslationProjectTask(c.ctx, taskID)
	if err != nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("ensure translation project task task_id=%s: %w", taskID, err)
	}
	result, err := c.translationFlow.RunTerminologyPhase(c.ctx, workflow.RunTerminologyPhaseInput{
		TaskID:  resolvedTaskID,
		Request: request,
		Prompt:  prompt,
	})
	if err != nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("run translation flow terminology task_id=%s: %w", resolvedTaskID, err)
	}
	return result, nil
}

// GetTranslationFlowTerminology returns the current terminology phase summary for one task.
func (c *TaskController) GetTranslationFlowTerminology(taskID string) (workflow.TerminologyPhaseResult, error) {
	if c.translationFlow == nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("translation flow workflow is not configured")
	}
	resolvedTaskID, err := c.manager.EnsureTranslationProjectTask(c.ctx, taskID)
	if err != nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("ensure translation project task task_id=%s: %w", taskID, err)
	}
	result, err := c.translationFlow.GetTerminologyPhase(c.ctx, resolvedTaskID)
	if err != nil {
		return workflow.TerminologyPhaseResult{}, fmt.Errorf("get translation flow terminology task_id=%s: %w", resolvedTaskID, err)
	}
	return result, nil
}
