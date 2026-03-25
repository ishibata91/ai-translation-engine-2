package task

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestStoreGetActiveTasks_ContainsFailedAndExcludesCompleted(t *testing.T) {
	ctx := context.Background()
	_, store := buildManagerForDeleteTest(ctx, t)

	testTasks := []Task{
		{ID: "task-pending", Name: "pending", Type: TypeTranslationProject, Status: StatusPending, Phase: "phase", Progress: 0, Metadata: TaskMetadata{}},
		{ID: "task-running", Name: "running", Type: TypeTranslationProject, Status: StatusRunning, Phase: "phase", Progress: 10, Metadata: TaskMetadata{}},
		{ID: "task-paused", Name: "paused", Type: TypeTranslationProject, Status: StatusPaused, Phase: "phase", Progress: 20, Metadata: TaskMetadata{}},
		{ID: "task-request-generated", Name: "request-generated", Type: TypeTranslationProject, Status: StatusRequestGenerated, Phase: "phase", Progress: 30, Metadata: TaskMetadata{}},
		{ID: "task-cancelled", Name: "cancelled", Type: TypeTranslationProject, Status: StatusCancelled, Phase: "phase", Progress: 40, Metadata: TaskMetadata{}},
		{ID: "task-failed", Name: "failed", Type: TypeTranslationProject, Status: StatusFailed, Phase: "phase", Progress: 50, Metadata: TaskMetadata{}},
		{ID: "task-completed", Name: "completed", Type: TypeTranslationProject, Status: StatusCompleted, Phase: "phase", Progress: 100, Metadata: TaskMetadata{}},
	}
	for _, currentTask := range testTasks {
		insertTask(ctx, t, store, currentTask)
	}

	activeTasks, err := store.GetActiveTasks(ctx)
	if err != nil {
		t.Fatalf("GetActiveTasks returned error: %v", err)
	}

	gotByID := make(map[string]Task, len(activeTasks))
	for _, currentTask := range activeTasks {
		gotByID[currentTask.ID] = currentTask
	}

	mustContain := []string{
		"task-pending",
		"task-running",
		"task-paused",
		"task-request-generated",
		"task-cancelled",
		"task-failed",
	}
	for _, taskID := range mustContain {
		if _, ok := gotByID[taskID]; !ok {
			t.Fatalf("active tasks does not contain %s", taskID)
		}
	}

	if _, ok := gotByID["task-completed"]; ok {
		t.Fatalf("active tasks must not contain completed task")
	}
}

func TestManagerInitialize_RestoresFailedTask(t *testing.T) {
	ctx := context.Background()
	manager, store := buildManagerForDeleteTest(ctx, t)
	manager.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	insertTask(ctx, t, store, Task{
		ID:       "task-failed",
		Name:     "failed",
		Type:     TypeTranslationProject,
		Status:   StatusFailed,
		Phase:    "phase",
		Progress: 10,
		ErrorMsg: "failed",
		Metadata: TaskMetadata{},
	})

	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}

	if _, ok := manager.activeTasks["task-failed"]; !ok {
		t.Fatalf("manager activeTasks does not contain failed task after initialize")
	}
}
