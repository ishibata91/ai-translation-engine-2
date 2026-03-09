package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/llm"
	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/queue"
	telemetry2 "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/telemetry"
)

// Manager coordinates vertical slices and infrastructure.
type Manager struct {
	store    *Store
	jobQueue *queue.Queue
	worker   *queue.Worker
	logger   *slog.Logger

	muSlices sync.RWMutex
	slices   map[string]Slice
}

// NewManager creates a new Pipeline.
func NewManager(
	store *Store,
	jobQueue *queue.Queue,
	worker *queue.Worker,
	logger *slog.Logger,
) *Manager {
	return &Manager{
		store:    store,
		jobQueue: jobQueue,
		worker:   worker,
		logger:   logger.With("slice", "pipeline"),
		slices:   make(map[string]Slice),
	}
}

// RegisterSlice registers a vertical slice to be orchestrated.
func (m *Manager) RegisterSlice(slice Slice) {
	m.muSlices.Lock()
	defer m.muSlices.Unlock()
	m.slices[slice.ID()] = slice
}

func (m *Manager) getSlice(id string) Slice {
	m.muSlices.RLock()
	defer m.muSlices.RUnlock()
	return m.slices[id]
}

// ExecuteSlice initiates the orchestration of a slice.
func (m *Manager) ExecuteSlice(ctx context.Context, sliceID string, input any, inputFile string) (string, error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionPipelineExecute)()
	m.logger.InfoContext(ctx, "pipeline.execute.started",
		slog.String("slice", sliceID),
		slog.String("input_file", inputFile),
	)

	registeredSlice, err := m.resolveSlice(ctx, sliceID)
	if err != nil {
		return "", fmt.Errorf("resolve pipeline slice=%s: %w", sliceID, err)
	}

	requests, err := m.prepareRequests(ctx, registeredSlice, sliceID, input)
	if err != nil {
		return "", fmt.Errorf("prepare pipeline requests slice=%s: %w", sliceID, err)
	}

	processID := uuid.New().String()
	if err := m.persistDispatchedState(ctx, processID, sliceID, inputFile); err != nil {
		return "", fmt.Errorf("persist pipeline state process_id=%s: %w", processID, err)
	}
	if err := m.submitJobs(ctx, processID, requests); err != nil {
		return "", fmt.Errorf("submit pipeline jobs process_id=%s: %w", processID, err)
	}

	go m.runProcessInBackground(ctx, processID)

	m.logger.InfoContext(ctx, "pipeline.execute.dispatched",
		slog.String("process_id", processID),
		slog.String("slice", sliceID),
		slog.Int("job_count", len(requests)),
	)
	return processID, nil
}

func (m *Manager) handleCompletion(ctx context.Context, processID string, workerErr error) {
	defer telemetry2.StartSpan(ctx, telemetry2.ActionPipelineExecute)()
	m.logger.DebugContext(ctx, "pipeline.completion.started", slog.String("process_id", processID))

	if workerErr != nil {
		m.logger.ErrorContext(ctx, "pipeline.worker.failed",
			append(telemetry2.ErrorAttrs(workerErr), slog.String("process_id", processID))...)
		if err := m.updatePhase(ctx, processID, PhaseFailed); err != nil {
			m.logger.WarnContext(ctx, "pipeline.phase_update_failed",
				append(telemetry2.ErrorAttrs(err), slog.String("process_id", processID))...)
		}
		return
	}

	state, err := m.store.GetState(ctx, processID)
	if err != nil {
		m.logger.ErrorContext(ctx, "pipeline.state_load_failed",
			append(telemetry2.ErrorAttrs(err), slog.String("process_id", processID))...)
		return
	}
	if state == nil {
		m.logger.WarnContext(ctx, "pipeline.state_missing", slog.String("process_id", processID))
		return
	}

	registeredSlice := m.getSlice(state.TargetSlice)
	if registeredSlice == nil {
		m.logger.ErrorContext(ctx, "pipeline.slice_missing",
			slog.String("process_id", processID),
			slog.String("slice", state.TargetSlice),
		)
		return
	}

	jobRequests, err := m.jobQueue.GetResults(ctx, processID)
	if err != nil {
		m.logger.ErrorContext(ctx, "pipeline.results_load_failed",
			append(telemetry2.ErrorAttrs(fmt.Errorf("get job results process_id=%s: %w", processID, err)),
				slog.String("process_id", processID),
			)...,
		)
		return
	}

	responses := m.buildResponses(ctx, jobRequests)
	if err := registeredSlice.SaveResults(ctx, responses); err != nil {
		saveErr := fmt.Errorf("save slice results process_id=%s slice=%s: %w", processID, state.TargetSlice, err)
		m.logger.ErrorContext(ctx, "pipeline.results_save_failed",
			append(telemetry2.ErrorAttrs(saveErr),
				slog.String("process_id", processID),
				slog.String("slice", state.TargetSlice),
			)...,
		)
		if err := m.updatePhase(ctx, processID, PhaseFailed); err != nil {
			m.logger.WarnContext(ctx, "pipeline.phase_update_failed",
				append(telemetry2.ErrorAttrs(err), slog.String("process_id", processID))...)
		}
		return
	}

	m.cleanupProcess(ctx, processID)
	m.logger.InfoContext(ctx, "pipeline.completion.succeeded",
		slog.String("process_id", processID),
		slog.String("slice", state.TargetSlice),
	)
}

func (m *Manager) updatePhase(ctx context.Context, processID string, phase string) error {
	state, err := m.store.GetState(ctx, processID)
	if err != nil {
		return fmt.Errorf("get process state for phase update process_id=%s: %w", processID, err)
	}
	if state == nil {
		return fmt.Errorf("process state not found for process_id=%s", processID)
	}
	state.CurrentPhase = phase
	if err := m.store.SaveState(ctx, *state); err != nil {
		return fmt.Errorf("save process phase process_id=%s phase=%s: %w", processID, phase, err)
	}
	return nil
}

// Recover matches the orchestration state with the infrastructure state and resumes jobs.
func (m *Manager) Recover(ctx context.Context) error {
	m.logger.InfoContext(ctx, "pipeline.recover.started")

	if err := m.worker.Recover(ctx); err != nil {
		return fmt.Errorf("recover job queue infrastructure: %w", err)
	}

	states, err := m.store.ListActiveStates(ctx)
	if err != nil {
		return fmt.Errorf("list active process states: %w", err)
	}

	for _, state := range states {
		m.logger.InfoContext(ctx, "pipeline.recover.process",
			slog.String("process_id", state.ProcessID),
			slog.String("slice", state.TargetSlice),
			slog.String("input_file", state.InputFile),
			slog.String("phase", state.CurrentPhase),
		)
		go m.runProcessInBackground(ctx, state.ProcessID)
	}

	m.logger.InfoContext(ctx, "pipeline.recover.completed", slog.Int("recovered_count", len(states)))
	return nil
}

func (m *Manager) resolveSlice(ctx context.Context, sliceID string) (Slice, error) {
	registeredSlice := m.getSlice(sliceID)
	if registeredSlice == nil {
		err := fmt.Errorf("slice not registered: %s", sliceID)
		m.logger.ErrorContext(ctx, "pipeline.execute.failed", telemetry2.ErrorAttrs(err)...)
		return nil, err
	}
	return registeredSlice, nil
}

func (m *Manager) prepareRequests(ctx context.Context, registeredSlice Slice, sliceID string, input any) ([]llm.Request, error) {
	requests, err := registeredSlice.PreparePrompts(ctx, input)
	if err != nil {
		wrappedErr := fmt.Errorf("prepare prompts for slice=%s: %w", sliceID, err)
		m.logger.ErrorContext(ctx, "pipeline.prepare_prompts_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("slice", sliceID))...)
		return nil, wrappedErr
	}
	return requests, nil
}

func (m *Manager) persistDispatchedState(ctx context.Context, processID, sliceID, inputFile string) error {
	state := ProcessState{
		ProcessID:    processID,
		TargetSlice:  sliceID,
		InputFile:    inputFile,
		CurrentPhase: PhaseDispatched,
	}
	if err := m.store.SaveState(ctx, state); err != nil {
		wrappedErr := fmt.Errorf("save process state process_id=%s: %w", processID, err)
		m.logger.ErrorContext(ctx, "pipeline.state_save_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("process_id", processID))...)
		return wrappedErr
	}
	return nil
}

func (m *Manager) submitJobs(ctx context.Context, processID string, requests []llm.Request) error {
	anyRequests := make([]any, len(requests))
	for i, req := range requests {
		anyRequests[i] = req
	}

	if err := m.jobQueue.SubmitJobs(ctx, processID, anyRequests); err != nil {
		wrappedErr := fmt.Errorf("submit jobs process_id=%s: %w", processID, err)
		m.logger.ErrorContext(ctx, "pipeline.job_submit_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("process_id", processID))...)
		return wrappedErr
	}
	return nil
}

func (m *Manager) runProcessInBackground(ctx context.Context, processID string) {
	bgCtx := telemetry2.WithTraceID(context.WithoutCancel(ctx))
	bgCtx = telemetry2.WithAction(bgCtx, telemetry2.ActionPipelineExecute, telemetry2.ResourceTask, processID)
	defer telemetry2.StartSpan(bgCtx, telemetry2.ActionPipelineExecute)()

	m.logger.InfoContext(bgCtx, "pipeline.worker.started", slog.String("process_id", processID))
	err := m.worker.ProcessProcessID(bgCtx, processID)
	m.handleCompletion(bgCtx, processID, err)
}

func (m *Manager) buildResponses(ctx context.Context, jobRequests []queue.JobRequest) []llm.Response {
	responses := make([]llm.Response, len(jobRequests))
	for i, jobRequest := range jobRequests {
		if jobRequest.ResponseJSON != nil {
			if err := json.Unmarshal([]byte(*jobRequest.ResponseJSON), &responses[i]); err != nil {
				m.logger.WarnContext(ctx, "pipeline.response_decode_failed",
					append(telemetry2.ErrorAttrs(fmt.Errorf("decode job response job_id=%s: %w", jobRequest.ID, err)),
						slog.String("job_id", jobRequest.ID),
					)...,
				)
				responses[i].Success = false
				responses[i].Error = fmt.Sprintf("unmarshal error: %v", err)
			}
			continue
		}
		if jobRequest.ErrorMessage != nil {
			responses[i].Success = false
			responses[i].Error = *jobRequest.ErrorMessage
		}
	}
	return responses
}

func (m *Manager) cleanupProcess(ctx context.Context, processID string) {
	if err := m.jobQueue.DeleteJobs(ctx, processID); err != nil {
		wrappedErr := fmt.Errorf("delete jobs process_id=%s: %w", processID, err)
		m.logger.WarnContext(ctx, "pipeline.cleanup.jobs_delete_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("process_id", processID))...)
	}
	if err := m.store.DeleteState(ctx, processID); err != nil {
		wrappedErr := fmt.Errorf("delete process state process_id=%s: %w", processID, err)
		m.logger.WarnContext(ctx, "pipeline.cleanup.state_delete_failed",
			append(telemetry2.ErrorAttrs(wrappedErr), slog.String("process_id", processID))...)
	}
}
