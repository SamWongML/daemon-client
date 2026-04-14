package app

import "github.com/demon/daemon-client/internal/session"

// UI-scoped messages. Session lifecycle events live in internal/session/events.go
// so engines can import them without forming an import cycle with this package.

type FocusChangeMsg struct{ To FocusID }
type OpenSessionMsg struct{ ID session.ID }
