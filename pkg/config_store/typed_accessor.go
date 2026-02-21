package config_store

import "context"

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
	_ = ctx
	_ = ns
	_ = key
	_ = defaultVal
	panic("not implemented")
}

// GetInt returns an int value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetInt(ctx context.Context, ns string, key string, defaultVal int) int {
	_ = ctx
	_ = ns
	_ = key
	_ = defaultVal
	panic("not implemented")
}

// GetFloat returns a float64 value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetFloat(ctx context.Context, ns string, key string, defaultVal float64) float64 {
	_ = ctx
	_ = ns
	_ = key
	_ = defaultVal
	panic("not implemented")
}

// GetBool returns a bool value for the given namespace/key, or defaultVal if not found.
func (t *TypedAccessor) GetBool(ctx context.Context, ns string, key string, defaultVal bool) bool {
	_ = ctx
	_ = ns
	_ = key
	_ = defaultVal
	panic("not implemented")
}
