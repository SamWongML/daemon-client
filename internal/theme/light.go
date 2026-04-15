package theme

import "charm.land/lipgloss/v2"

// Light is a paper-ish light theme with violet accents.
func Light() *Theme {
	return &Theme{
		Name:   "charm-light",
		Label:  "Charm (light)",
		IsDark: false,

		Bg:     lipgloss.Color("#faf8f5"),
		Fg:     lipgloss.Color("#1d1b26"),
		Dim:    lipgloss.Color("#6d6a7c"),
		Muted:  lipgloss.Color("#a9a4b5"),
		Accent: lipgloss.Color("#b3326b"),
		Border: lipgloss.Color("#c9c4d1"),

		Success: lipgloss.Color("#2f8f3f"),
		Warn:    lipgloss.Color("#a67b00"),
		Danger:  lipgloss.Color("#c0392b"),
		Info:    lipgloss.Color("#2f6ab3"),

		StPending:       lipgloss.Color("#6d6a7c"),
		StStarting:      lipgloss.Color("#7c4dff"),
		StRunning:       lipgloss.Color("#2f6ab3"),
		StAwaitingInput: lipgloss.Color("#a67b00"),
		StAwaitingPerm:  lipgloss.Color("#c45500"),
		StIdle:          lipgloss.Color("#5b6b9a"),
		StPaused:        lipgloss.Color("#7c4dff"),
		StCompleted:     lipgloss.Color("#2f8f3f"),
		StFailed:        lipgloss.Color("#c0392b"),
		StDisconnected:  lipgloss.Color("#c0392b"),
	}
}
