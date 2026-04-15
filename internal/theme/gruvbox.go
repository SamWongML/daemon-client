package theme

import "charm.land/lipgloss/v2"

// GruvboxHard — warm earth tones, high contrast.
func GruvboxHard() *Theme {
	return &Theme{
		Name:   "gruvbox-hard",
		Label:  "Gruvbox (hard)",
		IsDark: true,

		Bg:     lipgloss.Color("#1d2021"),
		Fg:     lipgloss.Color("#ebdbb2"),
		Dim:    lipgloss.Color("#928374"),
		Muted:  lipgloss.Color("#504945"),
		Accent: lipgloss.Color("#fabd2f"),
		Border: lipgloss.Color("#3c3836"),

		Success: lipgloss.Color("#b8bb26"),
		Warn:    lipgloss.Color("#fe8019"),
		Danger:  lipgloss.Color("#fb4934"),
		Info:    lipgloss.Color("#83a598"),

		StPending:       lipgloss.Color("#928374"),
		StStarting:      lipgloss.Color("#d3869b"),
		StRunning:       lipgloss.Color("#83a598"),
		StAwaitingInput: lipgloss.Color("#fabd2f"),
		StAwaitingPerm:  lipgloss.Color("#fe8019"),
		StIdle:          lipgloss.Color("#665c54"),
		StPaused:        lipgloss.Color("#d3869b"),
		StCompleted:     lipgloss.Color("#b8bb26"),
		StFailed:        lipgloss.Color("#fb4934"),
		StDisconnected:  lipgloss.Color("#fb4934"),
	}
}
