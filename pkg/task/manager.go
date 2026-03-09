package task

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/telemetry"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Manager struct {
	store  *Store
	logger *slog.Logger
	ctx    context.Context

	activeTasks map[string]*Task
	cancelFuncs map[string]context.CancelFunc
	mu          sync.RWMutex

	lastDBUpdate map[string]time.Time
	throttleMu   sync.Mutex

	runners   map[TaskType]Runner
	runnersMu sync.RWMutex

	completionHooks   map[TaskType][]CompletionHook
	completionHooksMu sync.RWMutex
}

type Runner interface {
	Run(ctx context.Context, task *Task, update func(phase string, progress float64)) error
}

type CompletionHook func(ctx context.Context, task *Task) error

func NewManager(ctx context.Context, logger *slog.Logger, store *Store) *Manager {
	return &Manager{
		ctx:             ctx,
		logger:          logger.With("module", "job_manager"),
		store:           store,
		activeTasks:     make(map[string]*Task),
		cancelFuncs:     make(map[string]context.CancelFunc),
		lastDBUpdate:    make(map[string]time.Time),
		runners:         make(map[TaskType]Runner),
		completionHooks: make(map[TaskType][]CompletionHook),
	}
}

func (m *Manager) RegisterRunner(ttype TaskType, runner Runner) {
	m.runnersMu.Lock()
	defer m.runnersMu.Unlock()
	m.runners[ttype] = runner
}

func (m *Manager) RegisterCompletionHook(ttype TaskType, hook CompletionHook) {
	m.completionHooksMu.Lock()
	defer m.completionHooksMu.Unlock()
	m.completionHooks[ttype] = append(m.completionHooks[ttype], hook)
}

func (m *Manager) ResumeTask(id string) error {
	if err := m.ResumeTaskWithContext(m.ctx, id); err != nil {
		return fmt.Errorf("resume task id=%s: %w", id, err)
	}
	return nil
}

func (m *Manager) ResumeTaskWithContext(ctx context.Context, id string) error {
	logCtx := m.newTaskContext(ctx, id)
	defer telemetry.StartSpan(logCtx, telemetry.ActionTaskManagement)()

	m.logger.InfoContext(logCtx, "task.resume.started", slog.String("task_id", id))

	currentTask, err := m.loadTaskForResume(logCtx, id)
	if err != nil {
		return fmt.Errorf("load task for resume task_id=%s: %w", id, err)
	}
	if currentTask.Status == StatusRunning {
		return nil
	}

	runner, err := m.runnerForTaskType(currentTask.Type)
	if err != nil {
		wrappedErr := fmt.Errorf("resolve runner for task_id=%s task_type=%s: %w", id, currentTask.Type, err)
		m.logger.ErrorContext(logCtx, "task.resume.failed",
			append(telemetry.ErrorAttrs(wrappedErr),
				slog.String("task_id", id),
				slog.String("task_type", string(currentTask.Type)),
				slog.String("task_status", string(currentTask.Status)),
				slog.String("task_phase", currentTask.Phase),
			)...,
		)
		return wrappedErr
	}

	taskCtx, err := m.prepareTaskExecution(logCtx, currentTask)
	if err != nil {
		return fmt.Errorf("prepare task execution task_id=%s: %w", id, err)
	}

	go m.runTask(taskCtx, currentTask, runner, StatusCompleted)
	return nil
}

func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *Manager) Store() *Store {
	return m.store
}

func (m *Manager) Initialize(ctx context.Context) error {
	m.logger.InfoContext(ctx, "task.manager.initialize.started")

	tasks, err := m.store.GetActiveTasks(ctx)
	if err != nil {
		return fmt.Errorf("load active tasks: %w", err)
	}

	for _, t := range tasks {
		task := t
		if task.Status == StatusRunning {
			task.Status = StatusPaused
			task.ErrorMsg = "interrupted by application shutdown"
			if err := m.store.UpdateTask(ctx, task.ID, task.Status, task.Phase, task.Progress, task.ErrorMsg); err != nil {
				m.logger.ErrorContext(ctx, "task.manager.initialize.update_failed",
					append(telemetry.ErrorAttrs(fmt.Errorf("update interrupted task task_id=%s: %w", task.ID, err)),
						slog.String("task_id", task.ID),
					)...,
				)
			}
			m.emitTaskUpdate(&task)
		}

		m.mu.Lock()
		m.activeTasks[task.ID] = &task
		m.mu.Unlock()
	}

	return nil
}

func (m *Manager) AddTask(name string, ttype TaskType, phase string, metadata TaskMetadata, runner func(ctx context.Context, update func(phase string, progress float64)) error) (string, error) {
	return m.AddTaskWithCompletionStatusContext(m.ctx, name, ttype, phase, metadata, StatusCompleted, func(ctx context.Context, _ string, update func(phase string, progress float64)) error {
		return runner(ctx, update)
	})
}

func (m *Manager) AddTaskWithCompletionStatus(
	name string,
	ttype TaskType,
	phase string,
	metadata TaskMetadata,
	completionStatus TaskStatus,
	runner func(ctx context.Context, taskID string, update func(phase string, progress float64)) error,
) (string, error) {
	return m.AddTaskWithCompletionStatusContext(m.ctx, name, ttype, phase, metadata, completionStatus, runner)
}

func (m *Manager) AddTaskWithCompletionStatusContext(
	ctx context.Context,
	name string,
	ttype TaskType,
	phase string,
	metadata TaskMetadata,
	completionStatus TaskStatus,
	runner func(ctx context.Context, taskID string, update func(phase string, progress float64)) error,
) (string, error) {
	baseCtx := telemetry.WithTraceID(m.contextOrBackground(ctx))
	defer telemetry.StartSpan(baseCtx, telemetry.ActionTaskManagement)()

	id := uuid.New().String()
	currentTask := &Task{
		ID:        id,
		Name:      name,
		Type:      ttype,
		Status:    StatusPending,
		Phase:     phase,
		Progress:  0,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	taskCtx := telemetry.WithAction(baseCtx, telemetry.ActionTaskManagement, telemetry.ResourceTask, id)
	m.logger.InfoContext(taskCtx, "task.create.started",
		slog.String("task_id", id),
		slog.String("task_name", name),
		slog.String("task_type", string(ttype)),
	)

	if err := m.store.InsertTask(taskCtx, *currentTask); err != nil {
		wrappedErr := fmt.Errorf("insert task task_id=%s: %w", id, err)
		m.logger.ErrorContext(taskCtx, "task.create.failed",
			append(telemetry.ErrorAttrs(wrappedErr), slog.String("task_id", id))...)
		return "", wrappedErr
	}

	runCtx, cancel := context.WithCancel(taskCtx)
	m.mu.Lock()
	m.activeTasks[id] = currentTask
	m.cancelFuncs[id] = cancel
	m.mu.Unlock()

	go func() {
		m.emitTaskUpdate(currentTask)
		m.markTaskRunning(runCtx, id)

		err := runner(runCtx, id, func(phase string, progress float64) {
			m.UpdateTaskProgressWithContext(runCtx, id, phase, progress)
		})
		if err != nil {
			if runCtx.Err() == context.Canceled || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				m.handleTaskCancellation(runCtx, id)
				return
			}
			m.handleTaskFailure(runCtx, id, err)
			return
		}

		m.handleTaskCompletionWithStatus(runCtx, id, completionStatus)
	}()

	return id, nil
}

func (m *Manager) markTaskRunning(ctx context.Context, id string) {
	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	currentTask.Status = StatusRunning
	currentTask.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()

	if err := m.persistTaskState(ctx, currentTask, ""); err != nil {
		m.logger.ErrorContext(ctx, "task.status_persist_failed",
			append(telemetry.ErrorAttrs(err), slog.String("task_id", currentTask.ID))...)
	}
	m.emitTaskUpdate(currentTask)
}

func (m *Manager) UpdateTaskProgress(id string, phase string, progress float64) {
	m.UpdateTaskProgressWithContext(m.ctx, id, phase, progress)
}

func (m *Manager) UpdateTaskProgressWithContext(ctx context.Context, id string, phase string, progress float64) {
	logCtx := m.newTaskContext(ctx, id)

	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	oldPhase := currentTask.Phase
	currentTask.Phase = phase
	currentTask.Progress = progress
	currentTask.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()

	if oldPhase != phase {
		m.logger.InfoContext(logCtx, "task.phase_changed",
			slog.String("task_id", id),
			slog.String("old_phase", oldPhase),
			slog.String("new_phase", phase),
		)
	}

	m.emitTaskUpdate(currentTask)
	m.throttleDBUpdate(logCtx, currentTask)
}

func (m *Manager) throttleDBUpdate(ctx context.Context, currentTask *Task) {
	m.throttleMu.Lock()
	defer m.throttleMu.Unlock()

	last, ok := m.lastDBUpdate[currentTask.ID]
	if !ok || time.Since(last) > 3*time.Second {
		if err := m.persistTaskState(ctx, currentTask, currentTask.ErrorMsg); err != nil {
			m.logger.ErrorContext(ctx, "task.progress_persist_failed",
				append(telemetry.ErrorAttrs(err), slog.String("task_id", currentTask.ID))...)
		}
		m.lastDBUpdate[currentTask.ID] = time.Now()
	}
}

func (m *Manager) handleTaskCompletionWithStatus(ctx context.Context, id string, status TaskStatus) {
	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	currentTask.Status = status
	currentTask.Progress = 100.0
	currentTask.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	delete(m.activeTasks, id)
	m.mu.Unlock()

	if status == StatusCompleted {
		m.runCompletionHooks(ctx, currentTask)
	}
	if err := m.persistTaskState(ctx, currentTask, ""); err != nil {
		m.logger.ErrorContext(ctx, "task.status_persist_failed",
			append(telemetry.ErrorAttrs(err),
				slog.String("task_id", currentTask.ID),
				slog.String("task_status", string(currentTask.Status)),
			)...,
		)
	}
	m.emitTaskUpdate(currentTask)
}

func (m *Manager) runCompletionHooks(ctx context.Context, currentTask *Task) {
	m.completionHooksMu.RLock()
	hooks := append([]CompletionHook(nil), m.completionHooks[currentTask.Type]...)
	m.completionHooksMu.RUnlock()

	for _, hook := range hooks {
		if err := hook(ctx, currentTask); err != nil {
			wrappedErr := fmt.Errorf("run completion hook task_id=%s task_type=%s: %w", currentTask.ID, currentTask.Type, err)
			m.logger.ErrorContext(ctx, "task.completion_hook_failed",
				append(telemetry.ErrorAttrs(wrappedErr),
					slog.String("task_id", currentTask.ID),
					slog.String("task_type", string(currentTask.Type)),
				)...,
			)
		}
	}
}

func (m *Manager) handleTaskFailure(ctx context.Context, id string, err error) {
	logCtx := m.newTaskContext(ctx, id)

	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	currentTask.Status = StatusFailed
	currentTask.ErrorMsg = err.Error()
	currentTask.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	m.mu.Unlock()

	if persistErr := m.persistTaskState(logCtx, currentTask, currentTask.ErrorMsg); persistErr != nil {
		m.logger.ErrorContext(logCtx, "task.status_persist_failed",
			append(telemetry.ErrorAttrs(persistErr), slog.String("task_id", currentTask.ID))...)
	}
	m.logger.ErrorContext(logCtx, "task.failed",
		append(telemetry.ErrorAttrs(err),
			slog.String("task_id", currentTask.ID),
			slog.String("task_phase", currentTask.Phase),
			slog.String("reason", currentTask.ErrorMsg),
		)...,
	)
	m.emitTaskUpdate(currentTask)
}

func (m *Manager) handleTaskCancellation(ctx context.Context, id string) {
	logCtx := m.newTaskContext(ctx, id)

	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	currentTask.Status = StatusCancelled
	currentTask.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	delete(m.activeTasks, id)
	m.mu.Unlock()

	if err := m.persistTaskState(logCtx, currentTask, "cancelled by user"); err != nil {
		m.logger.ErrorContext(logCtx, "task.status_persist_failed",
			append(telemetry.ErrorAttrs(err), slog.String("task_id", currentTask.ID))...)
	}
	m.emitTaskUpdate(currentTask)
}

func (m *Manager) CancelTask(id string) {
	m.mu.RLock()
	cancel, ok := m.cancelFuncs[id]
	m.mu.RUnlock()
	if ok {
		cancel()
	}
}

func (m *Manager) emitTaskUpdate(task *Task) {
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "task:updated", task)
	}
}

func (m *Manager) GetActiveTasks() []Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]Task, 0, len(m.activeTasks))
	for _, currentTask := range m.activeTasks {
		tasks = append(tasks, *currentTask)
	}
	return tasks
}

func (m *Manager) EmitPhaseCompleted(taskID string, phaseName string, dataSummary interface{}) {
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "task:phase_completed", map[string]interface{}{
			"taskId":  taskID,
			"phase":   phaseName,
			"summary": dataSummary,
		})
	}
}

func (m *Manager) contextOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *Manager) newTaskContext(ctx context.Context, taskID string) context.Context {
	baseCtx := telemetry.WithTraceID(m.contextOrBackground(ctx))
	return telemetry.WithAction(baseCtx, telemetry.ActionTaskManagement, telemetry.ResourceTask, taskID)
}

func (m *Manager) loadTaskForResume(ctx context.Context, id string) (*Task, error) {
	m.mu.Lock()
	currentTask, ok := m.activeTasks[id]
	m.mu.Unlock()
	if ok {
		if err := m.validateTaskForResume(ctx, currentTask); err != nil {
			return nil, err
		}
		return currentTask, nil
	}

	tasks, err := m.store.GetAllTasks(ctx)
	if err != nil {
		wrappedErr := fmt.Errorf("load persisted tasks for task_id=%s: %w", id, err)
		m.logger.ErrorContext(ctx, "task.resume.failed",
			append(telemetry.ErrorAttrs(wrappedErr), slog.String("task_id", id))...)
		return nil, wrappedErr
	}
	for _, persistedTask := range tasks {
		if persistedTask.ID != id {
			continue
		}
		taskCopy := persistedTask
		if err := m.validateTaskForResume(ctx, &taskCopy); err != nil {
			return nil, err
		}
		return &taskCopy, nil
	}

	err = fmt.Errorf("task_id=%s not found", id)
	m.logger.ErrorContext(ctx, "task.resume.failed",
		slog.String("task_id", id),
		slog.String("reason", err.Error()),
	)
	return nil, err
}

func (m *Manager) validateTaskForResume(ctx context.Context, currentTask *Task) error {
	if currentTask.Status == StatusCompleted {
		err := fmt.Errorf("task_id=%s already completed", currentTask.ID)
		m.logger.InfoContext(ctx, "task.resume.skipped",
			slog.String("task_id", currentTask.ID),
			slog.String("task_type", string(currentTask.Type)),
			slog.String("task_status", string(currentTask.Status)),
			slog.String("task_phase", currentTask.Phase),
			slog.String("reason", err.Error()),
		)
		return err
	}
	if currentTask.Status == StatusRunning {
		m.logger.InfoContext(ctx, "task.resume.skipped",
			slog.String("task_id", currentTask.ID),
			slog.String("task_type", string(currentTask.Type)),
			slog.String("task_status", string(currentTask.Status)),
			slog.String("task_phase", currentTask.Phase),
			slog.String("reason", "task is already running"),
		)
		return nil
	}
	return nil
}

func (m *Manager) runnerForTaskType(taskType TaskType) (Runner, error) {
	m.runnersMu.RLock()
	runner, ok := m.runners[taskType]
	m.runnersMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no runner registered")
	}
	return runner, nil
}

func (m *Manager) prepareTaskExecution(ctx context.Context, currentTask *Task) (context.Context, error) {
	taskCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	currentTask.Status = StatusRunning
	currentTask.ErrorMsg = ""
	currentTask.UpdatedAt = time.Now().UTC()
	m.activeTasks[currentTask.ID] = currentTask
	m.cancelFuncs[currentTask.ID] = cancel
	m.mu.Unlock()

	if err := m.persistTaskState(taskCtx, currentTask, ""); err != nil {
		wrappedErr := fmt.Errorf("persist running task task_id=%s: %w", currentTask.ID, err)
		m.logger.ErrorContext(taskCtx, "task.resume.failed",
			append(telemetry.ErrorAttrs(wrappedErr),
				slog.String("task_id", currentTask.ID),
				slog.String("task_type", string(currentTask.Type)),
			)...,
		)
		return nil, wrappedErr
	}

	return taskCtx, nil
}

func (m *Manager) runTask(ctx context.Context, currentTask *Task, runner Runner, completionStatus TaskStatus) {
	m.emitTaskUpdate(currentTask)

	err := runner.Run(ctx, currentTask, func(phase string, progress float64) {
		m.UpdateTaskProgressWithContext(ctx, currentTask.ID, phase, progress)
	})
	if err != nil {
		if ctx.Err() == context.Canceled || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			m.logger.InfoContext(ctx, "task.runner_canceled",
				slog.String("task_id", currentTask.ID),
				slog.String("task_type", string(currentTask.Type)),
				slog.String("reason", err.Error()),
			)
			m.handleTaskCancellation(ctx, currentTask.ID)
			return
		}

		m.logger.ErrorContext(ctx, "task.runner_failed",
			append(telemetry.ErrorAttrs(err),
				slog.String("task_id", currentTask.ID),
				slog.String("task_type", string(currentTask.Type)),
				slog.String("reason", err.Error()),
			)...,
		)
		m.handleTaskFailure(ctx, currentTask.ID, err)
		return
	}

	m.handleTaskCompletionWithStatus(ctx, currentTask.ID, completionStatus)
}

func (m *Manager) persistTaskState(ctx context.Context, currentTask *Task, errorMessage string) error {
	persistCtx := context.WithoutCancel(m.contextOrBackground(ctx))
	if err := m.store.UpdateTask(persistCtx, currentTask.ID, currentTask.Status, currentTask.Phase, currentTask.Progress, errorMessage); err != nil {
		return fmt.Errorf("update task task_id=%s status=%s phase=%s: %w", currentTask.ID, currentTask.Status, currentTask.Phase, err)
	}
	return nil
}
