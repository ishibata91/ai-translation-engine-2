package persona

import (
	"context"
	"strings"
)

type taskIDContextKey struct{}

// WithTaskID returns a child context carrying master-persona task_id.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	normalizedTaskID := strings.TrimSpace(taskID)
	if normalizedTaskID == "" {
		return ctx
	}
	return context.WithValue(ctx, taskIDContextKey{}, normalizedTaskID)
}

func taskIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	raw := ctx.Value(taskIDContextKey{})
	taskID, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(taskID)
}
