package theme

import "charm.land/lipgloss/v2"

// Dark is the default — inspired by Charm's Dracula-like dark palette.
func Dark() *Theme {
	return &Theme{
		Name:   "charm-dark",
		Label:  "Charm (dark)",
		IsDark: true,

		Bg:     lipgloss.Color("#1a1a24"),
		Fg:     lipgloss.Color("#e3e3ea"),
		Dim:    lipgloss.Color("#7d7d8c"),
		Muted:  lipgloss.Color("#4a4a58"),
		Accent: lipgloss.Color("#ff79c6"),
		Border: lipgloss.Color("#3a3a48"),

		Success: lipgloss.Color("#50fa7b"),
		Warn:    lipgloss.Color("#f1fa8c"),
		Danger:  lipgloss.Color("#ff5555"),
		Info:    lipgloss.Color("#8be9fd"),

		StPending:       lipgloss.Color("#7d7d8c"),
		StStarting:      lipgloss.Color("#bd93f9"),
		StRunning:       lipgloss.Color("#8be9fd"),
		StAwaitingInput: lipgloss.Color("#f1fa8c"),
		StAwaitingPerm:  lipgloss.Color("#ffb86c"),
		StIdle:          lipgloss.Color("#6272a4"),
		StPaused:        lipgloss.Color("#bd93f9"),
		StCompleted:     lipgloss.Color("#50fa7b"),
		StFailed:        lipgloss.Color("#ff5555"),
		StDisconnected:  lipgloss.Color("#ff5555"),
	}
}
