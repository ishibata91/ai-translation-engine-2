package task

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestManagerDeleteTask_TableDriven(t *testing.T) {
	testCases := []struct {
		name                     string
		taskID                   string
		prepare                  func(t *testing.T, manager *Manager, store *Store, ctx context.Context, taskID string)
		queueCleanerErr          error
		wantErrSubstr            string
		wantPersisted            bool
		wantInMemory             bool
		wantCancelFunc           bool
		wantThrottle             bool
		wantQueueDeleteCallCount int
	}{
		{
			name:   "停止済みタスクを削除すると永続と管理対象から消える",
			taskID: "task-paused",
			prepare: func(t *testing.T, manager *Manager, store *Store, ctx context.Context, taskID string) {
				t.Helper()
				insertTask(ctx, t, store, Task{ID: taskID, Name: "paused", Type: TypeTranslationProject, Status: StatusPaused, Phase: "phase", Progress: 50, Metadata: TaskMetadata{}})
				manager.activeTasks[taskID] = &Task{ID: taskID, Status: StatusPaused}
				_, cancel := context.WithCancel(context.Background())
				manager.cancelFuncs[taskID] = cancel
				manager.lastDBUpdate[taskID] = time.Now().UTC()
			},
			wantPersisted:            false,
			wantInMemory:             false,
			wantCancelFunc:           false,
			wantThrottle:             false,
			wantQueueDeleteCallCount: 1,
		},
		{
			name:   "実行中タスクは削除拒否される",
			taskID: "task-running",
			prepare: func(t *testing.T, manager *Manager, store *Store, ctx context.Context, taskID string) {
				t.Helper()
				insertTask(ctx, t, store, Task{ID: taskID, Name: "running", Type: TypeTranslationProject, Status: StatusRunning, Phase: "phase", Progress: 10, Metadata: TaskMetadata{}})
				manager.activeTasks[taskID] = &Task{ID: taskID, Status: StatusRunning}
				_, cancel := context.WithCancel(context.Background())
				manager.cancelFuncs[taskID] = cancel
				manager.lastDBUpdate[taskID] = time.Now().UTC()
			},
			wantErrSubstr:            "stop task before deleting",
			wantPersisted:            true,
			wantInMemory:             true,
			wantCancelFunc:           true,
			wantThrottle:             true,
			wantQueueDeleteCallCount: 0,
		},
		{
			name:                     "存在しないタスクは not found を返す",
			taskID:                   "task-missing",
			prepare:                  func(_ *testing.T, _ *Manager, _ *Store, _ context.Context, _ string) {},
			wantErrSubstr:            "not found",
			wantPersisted:            false,
			wantInMemory:             false,
			wantCancelFunc:           false,
			wantThrottle:             false,
			wantQueueDeleteCallCount: 0,
		},
		{
			name:   "queue 削除失敗時は task 削除を中断する",
			taskID: "task-queue-fail",
			prepare: func(t *testing.T, manager *Manager, store *Store, ctx context.Context, taskID string) {
				t.Helper()
				insertTask(ctx, t, store, Task{ID: taskID, Name: "paused", Type: TypeTranslationProject, Status: StatusPaused, Phase: "phase", Progress: 50, Metadata: TaskMetadata{}})
				manager.activeTasks[taskID] = &Task{ID: taskID, Status: StatusPaused}
				_, cancel := context.WithCancel(context.Background())
				manager.cancelFuncs[taskID] = cancel
				manager.lastDBUpdate[taskID] = time.Now().UTC()
			},
			queueCleanerErr:          fmt.Errorf("queue delete failed"),
			wantErrSubstr:            "delete queue requests",
			wantPersisted:            true,
			wantInMemory:             true,
			wantCancelFunc:           true,
			wantThrottle:             true,
			wantQueueDeleteCallCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			manager, store := buildManagerForDeleteTest(ctx, t)
			queueCleaner := &fakeTaskRequestCleaner{err: tc.queueCleanerErr}
			manager.SetTaskRequestCleaner(queueCleaner)
			tc.prepare(t, manager, store, ctx, tc.taskID)

			err := manager.DeleteTask(ctx, tc.taskID)
			if tc.wantErrSubstr == "" {
				if err != nil {
					t.Fatalf("DeleteTask returned error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("DeleteTask expected error containing %q, got nil", tc.wantErrSubstr)
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("DeleteTask error = %q, want contains %q", err.Error(), tc.wantErrSubstr)
				}
			}

			persisted := hasPersistedTask(ctx, t, store, tc.taskID)
			if persisted != tc.wantPersisted {
				t.Fatalf("persisted task exists = %v, want %v", persisted, tc.wantPersisted)
			}

			_, inMemory := manager.activeTasks[tc.taskID]
			if inMemory != tc.wantInMemory {
				t.Fatalf("in-memory task exists = %v, want %v", inMemory, tc.wantInMemory)
			}
			_, hasCancel := manager.cancelFuncs[tc.taskID]
			if hasCancel != tc.wantCancelFunc {
				t.Fatalf("cancel func exists = %v, want %v", hasCancel, tc.wantCancelFunc)
			}
			_, hasThrottle := manager.lastDBUpdate[tc.taskID]
			if hasThrottle != tc.wantThrottle {
				t.Fatalf("throttle state exists = %v, want %v", hasThrottle, tc.wantThrottle)
			}

			if len(queueCleaner.calledTaskIDs) != tc.wantQueueDeleteCallCount {
				t.Fatalf("queue delete call count = %d, want %d", len(queueCleaner.calledTaskIDs), tc.wantQueueDeleteCallCount)
			}
			if tc.wantQueueDeleteCallCount > 0 && queueCleaner.calledTaskIDs[0] != tc.taskID {
				t.Fatalf("queue delete task_id = %s, want %s", queueCleaner.calledTaskIDs[0], tc.taskID)
			}
		})
	}
}

type fakeTaskRequestCleaner struct {
	err           error
	calledTaskIDs []string
}

func (f *fakeTaskRequestCleaner) DeleteTaskRequests(_ context.Context, taskID string) error {
	f.calledTaskIDs = append(f.calledTaskIDs, taskID)
	return f.err
}

func buildManagerForDeleteTest(ctx context.Context, t *testing.T) (*Manager, *Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "task-delete-test.db")
	dsn := fmt.Sprintf("file:%s", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("migrate task db: %v", err)
	}

	store := NewStore(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	manager := NewManager(ctx, logger, store)
	return manager, store
}

func insertTask(ctx context.Context, t *testing.T, store *Store, currentTask Task) {
	t.Helper()
	if err := store.InsertTask(ctx, currentTask); err != nil {
		t.Fatalf("insert task %s: %v", currentTask.ID, err)
	}
}

func hasPersistedTask(ctx context.Context, t *testing.T, store *Store, taskID string) bool {
	t.Helper()
	tasks, err := store.GetAllTasks(ctx)
	if err != nil {
		t.Fatalf("get all tasks: %v", err)
	}
	for _, currentTask := range tasks {
		if currentTask.ID == taskID {
			return true
		}
	}
	return false
}
