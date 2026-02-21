package config_store

import (
	"context"
	"fmt"
)

// TypedAccessor wraps ConfigStore and provides type-safe access with default values.
// This is exclusively for the config table and does not support JSON operations.
type TypedAccessor struct {
	store ConfigStore
}

// NewTypedAccessor creates a new TypedAccessor wrapping the given ConfigStore.
func NewTypedAccessor(store ConfigStore) *TypedAccessor {
	return &TypedAccessor{store: store}
}

// GetString returns a string value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetString(ctx context.Context, ns string, key string, defaultVal string) string {
	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	return val
}

// GetInt returns an int value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetInt(ctx context.Context, ns string, key string, defaultVal int) int {
	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	var i int
	if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
		return defaultVal
	}
	return i
}

// GetFloat returns a float64 value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetFloat(ctx context.Context, ns string, key string, defaultVal float64) float64 {
	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	var f float64
	if _, err := fmt.Sscanf(val, "%f", &f); err != nil {
		return defaultVal
	}
	return f
}

// GetBool returns a bool value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetBool(ctx context.Context, ns string, key string, defaultVal bool) bool {
	val, err := t.store.Get(ctx, ns, key)
	if err != nil || val == "" {
		return defaultVal
	}
	switch val {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultVal
	}
}
