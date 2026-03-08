package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config"
)

// ConfigController exposes Wails-facing config and UI state operations.
type ConfigController struct {
	ctx          context.Context
	logger       *slog.Logger
	uiStateStore config.UIStateStore
	configStore  config.Config
}

// NewConfigController constructs the config controller adapter.
func NewConfigController(store *config.SQLiteStore, logger *slog.Logger) *ConfigController {
	if logger == nil {
		logger = slog.Default()
	}
	return &ConfigController{
		ctx:          context.Background(),
		logger:       logger.With("module", "config_controller"),
		uiStateStore: store,
		configStore:  store,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *ConfigController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// UIStateGetJSON returns the JSON value for the namespace/key pair.
func (c *ConfigController) UIStateGetJSON(namespace, key string) (string, error) {
	value, err := c.uiStateStore.Get(c.context(), namespace, key)
	if err != nil {
		return "", fmt.Errorf("get ui state namespace=%s key=%s: %w", namespace, key, err)
	}
	return value, nil
}

// UIStateSetJSON persists the JSON value for the namespace/key pair.
func (c *ConfigController) UIStateSetJSON(namespace, key string, value any) error {
	if err := c.uiStateStore.SetJSON(c.context(), namespace, key, value); err != nil {
		wrappedErr := fmt.Errorf("set ui state namespace=%s key=%s: %w", namespace, key, err)
		c.logger.ErrorContext(c.context(), "config.ui_state_set_failed",
			slog.String("namespace", namespace),
			slog.String("key", key),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}
	return nil
}

// UIStateDelete removes the namespace/key pair from UI state.
func (c *ConfigController) UIStateDelete(namespace, key string) error {
	if err := c.uiStateStore.Delete(c.context(), namespace, key); err != nil {
		return fmt.Errorf("delete ui state namespace=%s key=%s: %w", namespace, key, err)
	}
	return nil
}

// ConfigGet returns the config value for the namespace/key pair.
func (c *ConfigController) ConfigGet(namespace, key string) (string, error) {
	value, err := c.configStore.Get(c.context(), namespace, key)
	if err != nil {
		return "", fmt.Errorf("get config namespace=%s key=%s: %w", namespace, key, err)
	}
	if namespace == config.MasterPersonaPromptNamespace && value == "" {
		defaults := config.DefaultMasterPersonaPromptValues()
		return defaults[key], nil
	}
	return value, nil
}

// ConfigSet stores the config value for the namespace/key pair.
func (c *ConfigController) ConfigSet(namespace, key, value string) error {
	if err := c.configStore.Set(c.context(), namespace, key, value); err != nil {
		wrappedErr := fmt.Errorf("set config namespace=%s key=%s: %w", namespace, key, err)
		c.logger.ErrorContext(c.context(), "config.set_failed",
			slog.String("namespace", namespace),
			slog.String("key", key),
			slog.String("reason", wrappedErr.Error()),
		)
		return wrappedErr
	}
	return nil
}

// ConfigSetMany stores multiple config values in the same namespace.
func (c *ConfigController) ConfigSetMany(namespace string, values map[string]string) error {
	ctx := c.context()
	for key, value := range values {
		if err := c.configStore.Set(ctx, namespace, key, value); err != nil {
			return fmt.Errorf("set config namespace=%s key=%s: %w", namespace, key, err)
		}
	}
	return nil
}

// ConfigDelete removes the config value for the namespace/key pair.
func (c *ConfigController) ConfigDelete(namespace, key string) error {
	if err := c.configStore.Delete(c.context(), namespace, key); err != nil {
		return fmt.Errorf("delete config namespace=%s key=%s: %w", namespace, key, err)
	}
	return nil
}

// ConfigGetAll returns all config values for the namespace.
func (c *ConfigController) ConfigGetAll(namespace string) (map[string]string, error) {
	values, err := c.configStore.GetAll(c.context(), namespace)
	if err != nil {
		return nil, fmt.Errorf("get all config namespace=%s: %w", namespace, err)
	}
	if namespace == config.MasterPersonaPromptNamespace {
		return config.MergeMasterPersonaPromptDefaults(values), nil
	}
	return values, nil
}

func (c *ConfigController) context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}
