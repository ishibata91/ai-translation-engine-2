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
	ctx    context.Context // Wails context

	activeTasks map[string]*Task
	cancelFuncs map[string]context.CancelFunc
	mu          sync.RWMutex

	// Throttling for DB updates
	lastDBUpdate map[string]time.Time
	throttleMu   sync.Mutex

	// Task Runners
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
	baseCtx := m.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	logCtx := telemetry.WithTraceID(baseCtx)
	logCtx = telemetry.WithAction(logCtx, telemetry.ActionTaskManagement, telemetry.ResourceTask, id)
	defer telemetry.StartSpan(logCtx, telemetry.ActionTaskManagement)()
	m.logger.InfoContext(logCtx, "resuming task", slog.String("id", id))

	// Load task from DB if not in memory (unexpected case but let's be safe)
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	m.mu.Unlock()

	if !ok {
		// Try loading from DB
		dbTasks, err := m.store.GetAllTasks(context.Background())
		if err != nil {
			m.logger.ErrorContext(logCtx, "resume task failed",
				append(telemetry.ErrorAttrs(err), slog.String("id", id))...,
			)
			return err
		}
		for _, t := range dbTasks {
			if t.ID == id {
				task = &t
				break
			}
		}
	}

	if task == nil {
		err := fmt.Errorf("task not found: %s", id)
		m.logger.ErrorContext(logCtx, "resume task failed",
			slog.String("id", id),
			slog.String("reason", err.Error()),
		)
		return err
	}
	if task.Status == StatusCompleted {
		err := fmt.Errorf("task already completed: %s", id)
		m.logger.InfoContext(logCtx, "resume task skipped",
			slog.String("id", id),
			slog.String("task_type", string(task.Type)),
			slog.String("task_status", string(task.Status)),
			slog.String("task_phase", task.Phase),
			slog.String("reason", err.Error()),
		)
		return err
	}

	if task.Status == StatusRunning {
		m.logger.InfoContext(logCtx, "resume task skipped",
			slog.String("id", id),
			slog.String("task_type", string(task.Type)),
			slog.String("task_status", string(task.Status)),
			slog.String("task_phase", task.Phase),
			slog.String("reason", "task is already running"),
		)
		return nil
	}

	m.runnersMu.RLock()
	runner, ok := m.runners[task.Type]
	m.runnersMu.RUnlock()

	if !ok {
		err := fmt.Errorf("no runner registered for task type: %s", task.Type)
		m.logger.ErrorContext(logCtx, "resume task failed",
			slog.String("id", id),
			slog.String("task_type", string(task.Type)),
			slog.String("task_status", string(task.Status)),
			slog.String("task_phase", task.Phase),
			slog.String("reason", err.Error()),
		)
		return err
	}

	taskCtx, cancel := context.WithCancel(logCtx)
	m.mu.Lock()
	task.Status = StatusRunning
	task.ErrorMsg = ""
	task.UpdatedAt = time.Now().UTC()
	m.activeTasks[id] = task
	m.cancelFuncs[id] = cancel
	m.mu.Unlock()

	// Update DB immediately for status change
	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, "")

	go func() {
		m.emitTaskUpdate(task)

		err := runner.Run(taskCtx, task, func(p string, prog float64) {
			m.UpdateTaskProgress(id, p, prog)
		})

		if err != nil {
			if taskCtx.Err() == context.Canceled {
				m.logger.InfoContext(taskCtx, "task runner canceled",
					slog.String("id", id),
					slog.String("task_type", string(task.Type)),
					slog.String("reason", err.Error()),
				)
				m.handleTaskCancellation(id)
			} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				m.logger.InfoContext(taskCtx, "task runner canceled",
					slog.String("id", id),
					slog.String("task_type", string(task.Type)),
					slog.String("reason", err.Error()),
				)
				m.handleTaskCancellation(id)
			} else {
				m.logger.ErrorContext(taskCtx, "task runner failed",
					append(telemetry.ErrorAttrs(err),
						slog.String("id", id),
						slog.String("task_type", string(task.Type)),
						slog.String("reason", err.Error()),
					)...,
				)
				m.handleTaskFailure(id, err)
			}
		} else {
			m.handleTaskCompletion(id)
		}
	}()

	return nil
}

func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *Manager) Initialize(ctx context.Context) error {
	m.logger.InfoContext(ctx, "Initializing TaskManager")

	tasks, err := m.store.GetActiveTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active tasks: %w", err)
	}

	for _, t := range tasks {
		task := t
		// Correct Running to Paused on startup
		if task.Status == StatusRunning {
			task.Status = StatusPaused
			task.ErrorMsg = "interrupted by application shutdown"
			if err := m.store.UpdateTask(ctx, task.ID, task.Status, task.Phase, task.Progress, task.ErrorMsg); err != nil {
				m.logger.ErrorContext(ctx, "failed to correct task status", slog.String("id", task.ID))
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
	return m.AddTaskWithCompletionStatus(name, ttype, phase, metadata, StatusCompleted, func(ctx context.Context, _ string, update func(phase string, progress float64)) error {
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
	defer telemetry.StartSpan(m.ctx, telemetry.ActionTaskManagement)()
	id := uuid.New().String()
	task := &Task{
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

	m.logger.InfoContext(m.ctx, "adding new task",
		slog.String("id", id),
		slog.String("name", name),
		slog.String("type", string(ttype)),
	)

	if err := m.store.InsertTask(context.Background(), *task); err != nil {
		m.logger.ErrorContext(m.ctx, "failed to insert task into store",
			append(telemetry.ErrorAttrs(err), slog.String("id", id))...)
		return "", err
	}

	taskCtx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.activeTasks[id] = task
	m.cancelFuncs[id] = cancel
	m.mu.Unlock()

	go func() {
		m.emitTaskUpdate(task)
		m.markTaskRunning(id)

		err := runner(taskCtx, id, func(p string, prog float64) {
			m.UpdateTaskProgress(id, p, prog)
		})

		if err != nil {
			if taskCtx.Err() == context.Canceled {
				m.handleTaskCancellation(id)
			} else {
				m.handleTaskFailure(id, err)
			}
		} else {
			m.handleTaskCompletionWithStatus(id, completionStatus)
		}
	}()

	return id, nil
}

func (m *Manager) markTaskRunning(id string) {
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Status = StatusRunning
	task.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()

	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, "")
	m.emitTaskUpdate(task)
}

func (m *Manager) UpdateTaskProgress(id string, phase string, progress float64) {
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	oldPhase := task.Phase
	task.Phase = phase
	task.Progress = progress
	task.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()

	if oldPhase != phase {
		m.logger.InfoContext(m.ctx, "task phase changed",
			slog.String("id", id),
			slog.String("old_phase", oldPhase),
			slog.String("new_phase", phase),
		)
	}

	// High frequency emit
	m.emitTaskUpdate(task)

	// Throttled DB update
	m.throttleDBUpdate(task)
}

func (m *Manager) throttleDBUpdate(task *Task) {
	m.throttleMu.Lock()
	defer m.throttleMu.Unlock()

	last, ok := m.lastDBUpdate[task.ID]
	if !ok || time.Since(last) > 3*time.Second {
		err := m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, task.ErrorMsg)
		if err != nil {
			m.logger.Error("failed to update task in DB (throttled)", slog.String("id", task.ID), slog.Any("error", err))
		}
		m.lastDBUpdate[task.ID] = time.Now()
	}
}

func (m *Manager) handleTaskCompletion(id string) {
	m.handleTaskCompletionWithStatus(id, StatusCompleted)
}

func (m *Manager) handleTaskCompletionWithStatus(id string, status TaskStatus) {
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Status = status
	task.Progress = 100.0
	task.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	delete(m.activeTasks, id)
	m.mu.Unlock()

	if status == StatusCompleted {
		m.runCompletionHooks(task)
	}
	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, "")
	m.emitTaskUpdate(task)
}

func (m *Manager) runCompletionHooks(task *Task) {
	m.completionHooksMu.RLock()
	hooks := append([]CompletionHook(nil), m.completionHooks[task.Type]...)
	m.completionHooksMu.RUnlock()

	for _, hook := range hooks {
		if err := hook(context.Background(), task); err != nil {
			m.logger.Error("completion hook failed",
				slog.String("task_id", task.ID),
				slog.String("task_type", string(task.Type)),
				slog.String("error", err.Error()),
			)
		}
	}
}

func (m *Manager) handleTaskFailure(id string, err error) {
	logCtx := m.ctx
	if logCtx == nil {
		logCtx = context.Background()
	}
	logCtx = telemetry.WithTraceID(logCtx)
	logCtx = telemetry.WithAction(logCtx, telemetry.ActionTaskManagement, telemetry.ResourceTask, id)

	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Status = StatusFailed
	task.ErrorMsg = err.Error()
	task.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	// Keep in activeTasks? Maybe keep it for a while or until user acknowledges
	m.mu.Unlock()

	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, task.ErrorMsg)
	m.logger.ErrorContext(logCtx, "task marked failed",
		append(telemetry.ErrorAttrs(err),
			slog.String("id", task.ID),
			slog.String("phase", task.Phase),
			slog.String("reason", task.ErrorMsg),
		)...,
	)
	m.emitTaskUpdate(task)
}

func (m *Manager) handleTaskCancellation(id string) {
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Status = StatusCancelled
	task.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	delete(m.activeTasks, id)
	m.mu.Unlock()

	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, "cancelled by user")
	m.emitTaskUpdate(task)
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
	for _, t := range m.activeTasks {
		tasks = append(tasks, *t)
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
