package progress

import (
	"context"
	"testing"
	"time"
)

func TestHub_OnProgress_PassthroughsGenericFields(t *testing.T) {
	hub := NewHub()
	ch := make(chan ProgressEvent, 1)
	hub.Subscribe(ch)
	defer hub.Unsubscribe(ch)

	expected := ProgressEvent{
		TaskID:   "task-1",
		TaskType: "generic_task",
		Phase:    "PHASE_CUSTOM",
		Current:  3,
		Total:    10,
		Message:  "custom phase",
	}

	hub.OnProgress(context.Background(), expected)

	select {
	case got := <-ch:
		if got.TaskID != expected.TaskID {
			t.Fatalf("task id mismatch: got=%s want=%s", got.TaskID, expected.TaskID)
		}
		if got.TaskType != expected.TaskType {
			t.Fatalf("task type mismatch: got=%s want=%s", got.TaskType, expected.TaskType)
		}
		if got.Phase != expected.Phase {
			t.Fatalf("phase mismatch: got=%s want=%s", got.Phase, expected.Phase)
		}
		if got.Current != expected.Current || got.Total != expected.Total {
			t.Fatalf("progress mismatch: got=(%d/%d) want=(%d/%d)", got.Current, got.Total, expected.Current, expected.Total)
		}
		if got.Message != expected.Message {
			t.Fatalf("message mismatch: got=%s want=%s", got.Message, expected.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for progress event")
	}
}

