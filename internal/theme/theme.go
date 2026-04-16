package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme captures the full palette for one color scheme.
type Theme struct {
	Name   string
	Label  string
	IsDark bool

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

type Styles struct {
	Base        lipgloss.Style
	Dim         lipgloss.Style
	Accent      lipgloss.Style
	Header      lipgloss.Style
	HeaderDot   lipgloss.Style
	GroupHeader lipgloss.Style

	SidebarBase     lipgloss.Style
	SidebarSelected lipgloss.Style
	SidebarTitle    lipgloss.Style
	SidebarSubline  lipgloss.Style
	SidebarActivity lipgloss.Style
	FocusBar        lipgloss.Style

	SessionTitle lipgloss.Style
	SessionMeta  lipgloss.Style

	TranscriptBase lipgloss.Style
	UserBar        lipgloss.Style

	InputBase        lipgloss.Style
	InputMode        lipgloss.Style
	InputPlaceholder lipgloss.Style

	Footer    lipgloss.Style
	FooterKey lipgloss.Style

	TooSmall lipgloss.Style

	SettingsTitle    lipgloss.Style
	SettingsCategory lipgloss.Style
	SettingsActive   lipgloss.Style
	SettingsField    lipgloss.Style
	SettingsValue    lipgloss.Style
	SettingsBorder   lipgloss.Style
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
	s.SidebarActivity = lipgloss.NewStyle().Foreground(t.Muted).Italic(true)
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

	s.SettingsTitle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	s.SettingsCategory = lipgloss.NewStyle().Foreground(t.Fg)
	s.SettingsActive = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	s.SettingsField = lipgloss.NewStyle().Foreground(t.Dim)
	s.SettingsValue = lipgloss.NewStyle().Foreground(t.Fg)
	s.SettingsBorder = lipgloss.NewStyle().Foreground(t.Border)
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

// Registry is the ordered list of built-in themes. Cycling in the UI walks
// this list. The first entry is the default.
func Registry() []*Theme {
	return []*Theme{Dark(), Light(), TokyoNightStorm(), GruvboxHard()}
}

// ByName returns the theme with the given name (case-sensitive), falling back
// to the first registered theme if not found.
func ByName(name string) *Theme {
	for _, t := range Registry() {
		if t.Name == name {
			return t
		}
	}
	return Registry()[0]
}

// IsAuto reports whether the name represents the auto-detect sentinel.
func IsAuto(name string) bool { return name == "auto" }

// ResolveAuto picks charm-dark or charm-light based on a dark-background flag.
// Call this only when the config value is "auto".
func ResolveAuto(dark bool) *Theme {
	if dark {
		return ByName("charm-dark")
	}
	return ByName("charm-light")
}

// Next returns the theme after `name` in the registry, wrapping around.
func Next(name string) *Theme {
	all := Registry()
	for i, t := range all {
		if t.Name == name {
			return all[(i+1)%len(all)]
		}
	}
	return all[0]
}
