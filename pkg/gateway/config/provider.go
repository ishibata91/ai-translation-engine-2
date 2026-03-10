package config

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/wire"
	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/configstore"
)

// SQLiteStore adapts configstore.SQLiteStore to the gateway/config contracts.
type SQLiteStore struct {
	inner *configstore.SQLiteStore
}

// NewSQLiteStore creates the gateway-backed config store.
func NewSQLiteStore(ctx context.Context, db *sql.DB, logger *slog.Logger) (*SQLiteStore, error) {
	store, err := configstore.NewSQLiteStore(ctx, db, logger)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{inner: store}, nil
}

// Get reads a config value.
func (s *SQLiteStore) Get(ctx context.Context, namespace string, key string) (string, error) {
	return s.inner.Get(ctx, namespace, key)
}

// Set writes a config value.
func (s *SQLiteStore) Set(ctx context.Context, namespace string, key string, value string) error {
	return s.inner.Set(ctx, namespace, key, value)
}

// Delete removes a config value.
func (s *SQLiteStore) Delete(ctx context.Context, namespace string, key string) error {
	return s.inner.Delete(ctx, namespace, key)
}

// GetAll returns all config values under one namespace.
func (s *SQLiteStore) GetAll(ctx context.Context, namespace string) (map[string]string, error) {
	return s.inner.GetAll(ctx, namespace)
}

// Watch subscribes to config changes.
func (s *SQLiteStore) Watch(namespace string, key string, callback ChangeCallback) UnsubscribeFunc {
	unsubscribe := s.inner.Watch(namespace, key, func(event configstore.ChangeEvent) {
		callback(ChangeEvent{
			Namespace: event.Namespace,
			Key:       event.Key,
			OldValue:  event.OldValue,
			NewValue:  event.NewValue,
		})
	})
	return func() {
		unsubscribe()
	}
}

// SetJSON writes one UI state value.
func (s *SQLiteStore) SetJSON(ctx context.Context, namespace string, key string, value any) error {
	return s.inner.SetJSON(ctx, namespace, key, value)
}

// GetJSON reads one UI state value into target.
func (s *SQLiteStore) GetJSON(ctx context.Context, namespace string, key string, target any) error {
	return s.inner.GetJSON(ctx, namespace, key, target)
}

// GetSecret reads one secret value.
func (s *SQLiteStore) GetSecret(ctx context.Context, namespace string, key string) (string, error) {
	return s.inner.GetSecret(ctx, namespace, key)
}

// SetSecret writes one secret value.
func (s *SQLiteStore) SetSecret(ctx context.Context, namespace string, key string, value string) error {
	return s.inner.SetSecret(ctx, namespace, key, value)
}

// DeleteSecret deletes one secret value.
func (s *SQLiteStore) DeleteSecret(ctx context.Context, namespace string, key string) error {
	return s.inner.DeleteSecret(ctx, namespace, key)
}

// ListSecretKeys lists secret keys by namespace.
func (s *SQLiteStore) ListSecretKeys(ctx context.Context, namespace string) ([]string, error) {
	return s.inner.ListSecretKeys(ctx, namespace)
}

// ProviderSet exposes the gateway config providers with gateway-local bindings.
var ProviderSet = wire.NewSet(
	NewSQLiteStore,
	wire.Bind(new(Config), new(*SQLiteStore)),
	wire.Bind(new(UIStateStore), new(*SQLiteStore)),
	wire.Bind(new(SecretStore), new(*SQLiteStore)),
)
