package config_store

import "time"

// ChangeEvent is emitted when a config value is changed.
type ChangeEvent struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}

// SchemaVersion tracks the database schema version for migration.
type SchemaVersion struct {
	Version   int       `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}
