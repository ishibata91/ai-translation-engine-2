package task

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
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
}

type Runner interface {
	Run(ctx context.Context, task *Task, update func(phase string, progress float64)) error
}

func NewManager(ctx context.Context, logger *slog.Logger, store *Store) *Manager {
	return &Manager{
		ctx:          ctx,
		logger:       logger.With("module", "job_manager"),
		store:        store,
		activeTasks:  make(map[string]*Task),
		cancelFuncs:  make(map[string]context.CancelFunc),
		lastDBUpdate: make(map[string]time.Time),
		runners:      make(map[TaskType]Runner),
	}
}

func (m *Manager) RegisterRunner(ttype TaskType, runner Runner) {
	m.runnersMu.Lock()
	defer m.runnersMu.Unlock()
	m.runners[ttype] = runner
}

func (m *Manager) ResumeTask(id string) error {
	m.logger.Info("Resuming task", slog.String("id", id))

	// Load task from DB if not in memory (unexpected case but let's be safe)
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	m.mu.Unlock()

	if !ok {
		// Try loading from DB
		dbTasks, err := m.store.GetAllTasks(context.Background())
		if err != nil {
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
		return fmt.Errorf("task not found: %s", id)
	}

	if task.Status == StatusRunning {
		return fmt.Errorf("task is already running: %s", id)
	}

	m.runnersMu.RLock()
	runner, ok := m.runners[task.Type]
	m.runnersMu.RUnlock()

	if !ok {
		return fmt.Errorf("no runner registered for task type: %s", task.Type)
	}

	taskCtx, cancel := context.WithCancel(context.Background())
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
				m.handleTaskCancellation(id)
			} else {
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
	id := uuid.New().String()
	task := &Task{
		ID:        id,
		Name:      name,
		Type:      ttype,
		Status:    StatusRunning,
		Phase:     phase,
		Progress:  0,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := m.store.InsertTask(context.Background(), *task); err != nil {
		return "", err
	}

	taskCtx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.activeTasks[id] = task
	m.cancelFuncs[id] = cancel
	m.mu.Unlock()

	go func() {
		m.emitTaskUpdate(task)

		err := runner(taskCtx, func(p string, prog float64) {
			m.UpdateTaskProgress(id, p, prog)
		})

		if err != nil {
			if taskCtx.Err() == context.Canceled {
				m.handleTaskCancellation(id)
			} else {
				m.handleTaskFailure(id, err)
			}
		} else {
			m.handleTaskCompletion(id)
		}
	}()

	return id, nil
}

func (m *Manager) UpdateTaskProgress(id string, phase string, progress float64) {
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Phase = phase
	task.Progress = progress
	task.UpdatedAt = time.Now().UTC()
	m.mu.Unlock()

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
	m.mu.Lock()
	task, ok := m.activeTasks[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	task.Status = StatusCompleted
	task.Progress = 100.0
	task.UpdatedAt = time.Now().UTC()
	delete(m.cancelFuncs, id)
	delete(m.activeTasks, id)
	m.mu.Unlock()

	_ = m.store.UpdateTask(context.Background(), task.ID, task.Status, task.Phase, task.Progress, "")
	m.emitTaskUpdate(task)
}

func (m *Manager) handleTaskFailure(id string, err error) {
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

func (m *Manager) EmitPhaseCompleted(taskId string, phaseName string, dataSummary interface{}) {
	if m.ctx != nil {
		runtime.EventsEmit(m.ctx, "task:phase_completed", map[string]interface{}{
			"taskId":  taskId,
			"phase":   phaseName,
			"summary": dataSummary,
		})
	}
}
