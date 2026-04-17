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

// PulseTickMsg carries the current animation phase (0.0..1.0) derived from a
// 1 Hz sine wave. Sent every 50 ms by pulseTick() to drive glyph animation.
type PulseTickMsg struct {
	Phase float64 // 0.0..1.0
}

// PanicMsg is sent when a spawned goroutine recovers from a panic. The app
// shows it as an error toast rather than tearing down the terminal.
type PanicMsg struct {
	Err   any
	Stack string
}
