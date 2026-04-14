package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Name   string
	Bg     color.Color
	Fg     color.Color
	Dim    color.Color
	Muted  color.Color
	Accent color.Color
	Border color.Color

	Success color.Color
	Warn    color.Color
	Danger  color.Color
	Info    color.Color

	StPending       color.Color
	StStarting      color.Color
	StRunning       color.Color
	StAwaitingInput color.Color
	StAwaitingPerm  color.Color
	StIdle          color.Color
	StPaused        color.Color
	StCompleted     color.Color
	StFailed        color.Color
	StDisconnected  color.Color
}

func Dark() *Theme {
	return &Theme{
		Name:   "charm-dark",
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

type Styles struct {
	Base        lipgloss.Style
	Dim         lipgloss.Style
	Accent      lipgloss.Style
	Header      lipgloss.Style
	HeaderDot   lipgloss.Style
	GroupHeader lipgloss.Style

	SidebarBase      lipgloss.Style
	SidebarSelected  lipgloss.Style
	SidebarTitle     lipgloss.Style
	SidebarSubline   lipgloss.Style
	SidebarActivity  lipgloss.Style
	FocusBar         lipgloss.Style

	SessionTitle  lipgloss.Style
	SessionMeta   lipgloss.Style

	TranscriptBase lipgloss.Style
	UserBar        lipgloss.Style

	InputBase    lipgloss.Style
	InputMode    lipgloss.Style
	InputPlaceholder lipgloss.Style

	Footer       lipgloss.Style
	FooterKey    lipgloss.Style

	TooSmall     lipgloss.Style
}

func BuildStyles(t *Theme) *Styles {
	s := &Styles{}
	s.Base = lipgloss.NewStyle().Foreground(t.Fg)
	s.Dim = lipgloss.NewStyle().Foreground(t.Dim)
	s.Accent = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	s.Header = lipgloss.NewStyle().Foreground(t.Fg)
	s.HeaderDot = lipgloss.NewStyle().Foreground(t.Success)
	s.GroupHeader = lipgloss.NewStyle().Foreground(t.Dim).Italic(true)

	s.SidebarBase = lipgloss.NewStyle().Foreground(t.Fg)
	s.SidebarSelected = lipgloss.NewStyle().Foreground(t.Fg).Bold(true)
	s.SidebarTitle = lipgloss.NewStyle().Foreground(t.Fg)
	s.SidebarSubline = lipgloss.NewStyle().Foreground(t.Dim)
	s.SidebarActivity = lipgloss.NewStyle().Foreground(t.Info).Italic(true)
	s.FocusBar = lipgloss.NewStyle().Foreground(t.Accent)

	s.SessionTitle = lipgloss.NewStyle().Foreground(t.Fg).Bold(true)
	s.SessionMeta = lipgloss.NewStyle().Foreground(t.Dim)

	s.TranscriptBase = lipgloss.NewStyle().Foreground(t.Fg)
	s.UserBar = lipgloss.NewStyle().Foreground(t.Info)

	s.InputBase = lipgloss.NewStyle().Foreground(t.Fg)
	s.InputMode = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	s.InputPlaceholder = lipgloss.NewStyle().Foreground(t.Dim).Italic(true)

	s.Footer = lipgloss.NewStyle().Foreground(t.Dim)
	s.FooterKey = lipgloss.NewStyle().Foreground(t.Accent)

	s.TooSmall = lipgloss.NewStyle().Foreground(t.Warn).Bold(true)
	return s
}

func StatusColor(t *Theme, name string) color.Color {
	switch name {
	case "pending":
		return t.StPending
	case "starting":
		return t.StStarting
	case "running":
		return t.StRunning
	case "awaiting_input":
		return t.StAwaitingInput
	case "awaiting_perm":
		return t.StAwaitingPerm
	case "idle":
		return t.StIdle
	case "paused":
		return t.StPaused
	case "completed":
		return t.StCompleted
	case "failed":
		return t.StFailed
	case "disconnected":
		return t.StDisconnected
	}
	return t.Fg
}
