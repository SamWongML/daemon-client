package app

import "github.com/demon/daemon-client/internal/session"

// UI-scoped messages. Session lifecycle events live in internal/session/events.go
// so engines can import them without forming an import cycle with this package.

type FocusChangeMsg struct{ To FocusID }
type OpenSessionMsg struct{ ID session.ID }

// SetThemeMsg switches the active theme by name. If the name is unknown the
// registry's default is used.
type SetThemeMsg struct{ Name string }

// CycleThemeMsg advances to the next theme in the registry.
type CycleThemeMsg struct{}
