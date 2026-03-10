package configaccess

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type configReader interface {
	Get(ctx context.Context, namespace string, key string) (string, error)
}

// TypedAccessor wraps Config and provides type-safe access with default values.
// This is exclusively for the config table and does not support JSON operations.
type TypedAccessor struct {
	store configReader
}

// NewTypedAccessor creates a new TypedAccessor wrapping the given Config.
func NewTypedAccessor(store configReader) *TypedAccessor {
	return &TypedAccessor{store: store}
}

// GetString returns a string value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetString(ctx context.Context, ns string, key string, defaultVal string) string {
	slog.DebugContext(ctx, "ENTER TypedAccessor.GetString", slog.String("ns", ns), slog.String("key", key))

	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return strings.TrimSpace(val)
}

// GetInt returns an int value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetInt(ctx context.Context, ns string, key string, defaultVal int) int {
	slog.DebugContext(ctx, "ENTER TypedAccessor.GetInt", slog.String("ns", ns), slog.String("key", key))

	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return parseIntValue(val, defaultVal)
}

// GetFloat returns a float64 value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetFloat(ctx context.Context, ns string, key string, defaultVal float64) float64 {
	slog.DebugContext(ctx, "ENTER TypedAccessor.GetFloat", slog.String("ns", ns), slog.String("key", key))

	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return parseFloatValue(val, defaultVal)
}

// GetBool returns a bool value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetBool(ctx context.Context, ns string, key string, defaultVal bool) bool {
	slog.DebugContext(ctx, "ENTER TypedAccessor.GetBool", slog.String("ns", ns), slog.String("key", key))

	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return parseBoolValue(val, defaultVal)
}

func parseIntValue(val string, defaultVal int) int {
	var i int
	if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
		return defaultVal
	}
	return i
}

func parseFloatValue(val string, defaultVal float64) float64 {
	var f float64
	if _, err := fmt.Sscanf(val, "%f", &f); err != nil {
		return defaultVal
	}
	return f
}

func parseBoolValue(val string, defaultVal bool) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultVal
	}
}
