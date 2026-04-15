package theme

import "charm.land/lipgloss/v2"

// TokyoNightStorm — deep navy with cyan/purple accents.
func TokyoNightStorm() *Theme {
	return &Theme{
		Name:   "tokyonight-storm",
		Label:  "Tokyo Night (storm)",
		IsDark: true,

		Bg:     lipgloss.Color("#24283b"),
		Fg:     lipgloss.Color("#c0caf5"),
		Dim:    lipgloss.Color("#7982a9"),
		Muted:  lipgloss.Color("#414868"),
		Accent: lipgloss.Color("#7aa2f7"),
		Border: lipgloss.Color("#3b4261"),

		Success: lipgloss.Color("#9ece6a"),
		Warn:    lipgloss.Color("#e0af68"),
		Danger:  lipgloss.Color("#f7768e"),
		Info:    lipgloss.Color("#7dcfff"),

		StPending:       lipgloss.Color("#7982a9"),
		StStarting:      lipgloss.Color("#bb9af7"),
		StRunning:       lipgloss.Color("#7dcfff"),
		StAwaitingInput: lipgloss.Color("#e0af68"),
		StAwaitingPerm:  lipgloss.Color("#ff9e64"),
		StIdle:          lipgloss.Color("#565f89"),
		StPaused:        lipgloss.Color("#bb9af7"),
		StCompleted:     lipgloss.Color("#9ece6a"),
		StFailed:        lipgloss.Color("#f7768e"),
		StDisconnected:  lipgloss.Color("#f7768e"),
	}
}
