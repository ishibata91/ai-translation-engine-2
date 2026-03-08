package config

import base "github.com/ishibata91/ai-translation-engine-2/pkg/config"

// ChangeCallback aliases the existing config callback contract at the gateway boundary.
type ChangeCallback = base.ChangeCallback

// ChangeEvent aliases the existing config change event at the gateway boundary.
type ChangeEvent = base.ChangeEvent

// UnsubscribeFunc aliases the existing config watcher contract at the gateway boundary.
type UnsubscribeFunc = base.UnsubscribeFunc

// Config exposes configuration reads and writes as a gateway contract.
type Config = base.Config

// UIStateStore exposes UI state persistence as a gateway contract.
type UIStateStore = base.UIStateStore

// SecretStore exposes secret persistence as a gateway contract.
type SecretStore = base.SecretStore
