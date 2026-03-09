package config

import "context"

// ChangeCallback is a function invoked when a watched config value changes.
type ChangeCallback func(event ChangeEvent)

// UnsubscribeFunc cancels a Watch subscription.
type UnsubscribeFunc func()

// Config provides read/write access to backend configuration values.
// Values are stored as plain text strings keyed by namespace and key.
// JSON values are NOT permitted in this store.
type Config interface {
	Get(ctx context.Context, namespace string, key string) (string, error)
	Set(ctx context.Context, namespace string, key string, value string) error
	Delete(ctx context.Context, namespace string, key string) error
	GetAll(ctx context.Context, namespace string) (map[string]string, error)
	Watch(namespace string, key string, callback ChangeCallback) UnsubscribeFunc
}

// UIStateStore provides read/write access to UI layout and state data.
// JSON-formatted structured data is permitted in this store.
type UIStateStore interface {
	Get(ctx context.Context, namespace string, key string) (string, error)
	SetJSON(ctx context.Context, namespace string, key string, value any) error
	GetJSON(ctx context.Context, namespace string, key string, target any) error
	Delete(ctx context.Context, namespace string, key string) error
	GetAll(ctx context.Context, namespace string) (map[string]string, error)
}

// SecretStore manages sensitive information such as API keys.
// Separated from Config for future encryption and OS Keychain integration.
type SecretStore interface {
	GetSecret(ctx context.Context, namespace string, key string) (string, error)
	SetSecret(ctx context.Context, namespace string, key string, value string) error
	DeleteSecret(ctx context.Context, namespace string, key string) error
	ListSecretKeys(ctx context.Context, namespace string) ([]string, error)
}
