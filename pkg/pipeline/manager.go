package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/queue"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

// Manager coordinates vertical slices and infrastructure.
type Manager struct {
	store    *Store
	jobQueue *queue.Queue
	worker   *queue.Worker
	logger   *slog.Logger

	// Slice registration
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
	m := &Manager{
		store:    store,
		jobQueue: jobQueue,
		worker:   worker,
		logger:   logger.With("slice", "Pipeline"),
		slices:   make(map[string]Slice),
	}
	return m
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
	m.logger.DebugContext(ctx, "ENTER ExecuteSlice", slog.String("slice", sliceID), slog.String("input_file", inputFile))

	// 1. Resolve Slice
	slice := m.getSlice(sliceID)
	if slice == nil {
		return "", fmt.Errorf("slice not registered: %s", sliceID)
	}

	// 2. Phase 1: Prepare Prompts
	reqs, err := slice.PreparePrompts(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to prepare prompts for %s: %w", sliceID, err)
	}

	// 3. Generate ProcessID & Save State
	processID := uuid.New().String()
	state := ProcessState{
		ProcessID:    processID,
		TargetSlice:  sliceID,
		InputFile:    inputFile,
		CurrentPhase: PhaseDispatched,
	}
	if err := m.store.SaveState(ctx, state); err != nil {
		return "", fmt.Errorf("failed to save process state: %w", err)
	}

	// 4. Submit to JobQueue
	anyReqs := make([]any, len(reqs))
	for i, r := range reqs {
		anyReqs[i] = r
	}
	if err := m.jobQueue.SubmitJobs(ctx, processID, anyReqs); err != nil {
		return "", fmt.Errorf("failed to submit jobs to queue: %w", err)
	}

	// 5. Run Worker in background
	go func() {
		bgCtx := context.Background()
		// ProcessID processing blocks until done
		err := m.worker.ProcessProcessID(bgCtx, processID)
		m.handleCompletion(bgCtx, processID, err)
	}()

	m.logger.DebugContext(ctx, "EXIT ExecuteSlice", slog.String("process_id", processID))
	return processID, nil
}

func (m *Manager) handleCompletion(ctx context.Context, processID string, workerErr error) {
	m.logger.DebugContext(ctx, "ENTER handleCompletion", slog.String("process_id", processID), slog.Any("worker_err", workerErr))

	if workerErr != nil {
		m.logger.ErrorContext(ctx, "Process failed in worker", slog.String("process_id", processID), slog.Any("error", workerErr))
		_ = m.updatePhase(ctx, processID, PhaseFailed)
		return
	}

	// 1. Get State
	state, err := m.store.GetState(ctx, processID)
	if err != nil || state == nil {
		m.logger.ErrorContext(ctx, "Failed to fetch state for callback", slog.String("process_id", processID))
		return
	}

	// 2. Resolve Slice
	slice := m.getSlice(state.TargetSlice)
	if slice == nil {
		m.logger.ErrorContext(ctx, "Slice disappeared after completion", slog.String("slice", state.TargetSlice))
		return
	}

	// 3. Get Results from JobQueue
	jobRequests, err := m.jobQueue.GetResults(ctx, processID)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to get results from job queue", slog.String("process_id", processID), slog.Any("error", err))
		return
	}

	// 4. Phase 2: Save Results
	responses := make([]llm.Response, len(jobRequests))
	for i, jr := range jobRequests {
		if jr.ResponseJSON != nil {
			if err := json.Unmarshal([]byte(*jr.ResponseJSON), &responses[i]); err != nil {
				m.logger.ErrorContext(ctx, "Failed to unmarshal response", slog.String("job_id", jr.ID), slog.Any("error", err))
				responses[i].Success = false
				responses[i].Error = fmt.Sprintf("unmarshal error: %v", err)
			}
		} else if jr.ErrorMessage != nil {
			responses[i].Success = false
			responses[i].Error = *jr.ErrorMessage
		}
	}

	if err := slice.SaveResults(ctx, responses); err != nil {
		m.logger.ErrorContext(ctx, "Failed to save slice results", slog.String("process_id", processID), slog.Any("error", err))
		_ = m.updatePhase(ctx, processID, PhaseFailed)
		return
	}

	// 5. Cleanup
	if err := m.jobQueue.DeleteJobs(ctx, processID); err != nil {
		m.logger.WarnContext(ctx, "Failed to hard delete jobs after completion", slog.String("process_id", processID), slog.Any("error", err))
	}
	_ = m.store.DeleteState(ctx, processID)
}

func (m *Manager) updatePhase(ctx context.Context, processID string, phase string) error {
	state, err := m.store.GetState(ctx, processID)
	if err != nil || state == nil {
		return err
	}
	state.CurrentPhase = phase
	return m.store.SaveState(ctx, *state)
}

// Recover matches the orchestration state with the infrastructure state and resumes jobs.
func (m *Manager) Recover(ctx context.Context) error {
	m.logger.InfoContext(ctx, "ENTER Recover")

	// 1. Recover JobQueue infrastructure (resets IN_PROGRESS to PENDING)
	if err := m.worker.Recover(ctx); err != nil {
		return fmt.Errorf("failed to recover job queue infra: %w", err)
	}

	// 2. Load active orchestration states
	states, err := m.store.ListActiveStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active states: %w", err)
	}

	for _, state := range states {
		m.logger.InfoContext(ctx, "Recovering process",
			slog.String("process_id", state.ProcessID),
			slog.String("slice", state.TargetSlice),
			slog.String("input_file", state.InputFile),
			slog.String("phase", state.CurrentPhase))

		// Resume the worker for this processID in background
		go func(pid string) {
			bgCtx := context.Background()
			err := m.worker.ProcessProcessID(bgCtx, pid)
			m.handleCompletion(bgCtx, pid, err)
		}(state.ProcessID)
	}

	m.logger.InfoContext(ctx, "EXIT Recover", slog.Int("recovered_count", len(states)))
	return nil
}
