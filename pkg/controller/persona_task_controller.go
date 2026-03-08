package controller

import (
	"context"
	"fmt"

	runtimequeue "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/task"
	"github.com/ishibata91/ai-translation-engine-2/pkg/workflow"
)

// PersonaTaskController exposes MasterPersona-specific Wails-facing task operations.
type PersonaTaskController struct {
	manager               *task.Manager
	masterPersonaWorkflow workflow.MasterPersona
}

// NewPersonaTaskController constructs the MasterPersona controller adapter.
func NewPersonaTaskController(manager *task.Manager, masterPersonaWorkflow workflow.MasterPersona) *PersonaTaskController {
	return &PersonaTaskController{
		manager:               manager,
		masterPersonaWorkflow: masterPersonaWorkflow,
	}
}

// GetAllTasks loads all persisted tasks so persona-specific tests and screens can hydrate state.
func (c *PersonaTaskController) GetAllTasks() ([]task.Task, error) {
	return c.manager.Store().GetAllTasks(context.Background())
}

// StartMasterPersonTask starts the MasterPersona workflow while keeping the existing Wails API signature.
func (c *PersonaTaskController) StartMasterPersonTask(input task.StartMasterPersonTaskInput) (string, error) {
	if c.masterPersonaWorkflow == nil {
		return "", fmt.Errorf("master persona workflow is not configured")
	}
	return c.masterPersonaWorkflow.StartMasterPersona(context.Background(), workflow.StartMasterPersonaInput{
		SourceJSONPath:    input.SourceJSONPath,
		OverwriteExisting: input.OverwriteExisting,
	})
}

// ResumeTask resumes a MasterPersona task through workflow for UI compatibility.
func (c *PersonaTaskController) ResumeTask(taskID string) error {
	if c.masterPersonaWorkflow == nil {
		return fmt.Errorf("master persona workflow is not configured")
	}
	return c.masterPersonaWorkflow.ResumeMasterPersona(context.Background(), taskID)
}

// ResumeMasterPersonaTask resumes a MasterPersona task explicitly.
func (c *PersonaTaskController) ResumeMasterPersonaTask(taskID string) error {
	return c.ResumeTask(taskID)
}

// CancelTask cancels a MasterPersona task through workflow.
func (c *PersonaTaskController) CancelTask(taskID string) {
	if c.masterPersonaWorkflow == nil {
		return
	}
	_ = c.masterPersonaWorkflow.CancelMasterPersona(context.Background(), taskID)
}

// GetTaskRequestState returns aggregate queued request state for one task.
func (c *PersonaTaskController) GetTaskRequestState(taskID string) (runtimequeue.TaskRequestState, error) {
	if c.masterPersonaWorkflow == nil {
		return runtimequeue.TaskRequestState{}, fmt.Errorf("master persona workflow is not configured")
	}
	return c.masterPersonaWorkflow.GetTaskRequestState(context.Background(), taskID)
}

// GetTaskRequests returns queued requests for one task.
func (c *PersonaTaskController) GetTaskRequests(taskID string) ([]runtimequeue.JobRequest, error) {
	if c.masterPersonaWorkflow == nil {
		return nil, fmt.Errorf("master persona workflow is not configured")
	}
	return c.masterPersonaWorkflow.GetTaskRequests(context.Background(), taskID)
}
