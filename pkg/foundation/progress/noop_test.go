package progress

import (
	"context"
	"testing"
)

func TestNoopNotifier_OnProgress(t *testing.T) {
	notifier := NewNoopNotifier()

	// パニックしないことを確認
	ctx := context.Background()
	event := ProgressEvent{
		CorrelationID: "test-id",
		Total:         10,
		Completed:     5,
		Status:        StatusInProgress,
		Message:       "Processing...",
	}

	// 何度呼び出しても問題ないこと
	notifier.OnProgress(ctx, event)
	notifier.OnProgress(ctx, ProgressEvent{Status: StatusCompleted})

	// 特にアサーションは不要（何もしないことが正解）
}
