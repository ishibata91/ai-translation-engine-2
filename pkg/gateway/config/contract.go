package config

import (
	config2 "github.com/ishibata91/ai-translation-engine-2/pkg/workflow/config"
)

// ChangeCallback aliases the existing config callback contract at the gateway boundary.
type ChangeCallback = config2.ChangeCallback

// ChangeEvent aliases the existing config change event at the gateway boundary.
type ChangeEvent = config2.ChangeEvent

// UnsubscribeFunc aliases the existing config watcher contract at the gateway boundary.
type UnsubscribeFunc = config2.UnsubscribeFunc

// Config exposes configuration reads and writes as a gateway contract.
type Config = config2.Config

// UIStateStore exposes UI state persistence as a gateway contract.
type UIStateStore = config2.UIStateStore

// SecretStore exposes secret persistence as a gateway contract.
type SecretStore = config2.SecretStore
