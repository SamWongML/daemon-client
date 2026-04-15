// Package settings renders the settings screen (§7.4).
//
// Phase 1 scope: category navigation with read-only field values. Editing is
// stubbed behind a toast — the huh-backed inline editor lands with M4
// (onboarding) since the two share components.
package settings

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

// Category is one row in the left nav.
type Category struct {
	ID    string
	Title string
	Icon  string
}

// Field is one labelled value row in the right panel.
type Field struct {
	Label string
	Value string
	Hint  string
}

// Categories is the fixed phase-1 list. Order matches §7.4.
var Categories = []Category{
	{ID: "connection", Title: "Connection", Icon: "⚡"},
	{ID: "workspace", Title: "Workspace", Icon: "📁"},
	{ID: "agents", Title: "Agents", Icon: "🤖"},
	{ID: "appearance", Title: "Appearance", Icon: "🎨"},
	{ID: "notifications", Title: "Notifications", Icon: "🔔"},
	{ID: "keyboard", Title: "Keyboard", Icon: "⌨"},
	{ID: "advanced", Title: "Advanced", Icon: "⚙"},
	{ID: "about", Title: "About", Icon: "ℹ"},
}

// State is everything the screen needs to render a frame.
type State struct {
	Selected int
	Theme    *theme.Theme // current theme — surfaced in Appearance

	ServerURL string
	Workdir   string
	MaxSess   int

	Agent string
	Model string

	LogLevel string
	LogFile  string
	Version  string
	Build    string
}

// Fields returns the field list for the given category ID.
func (s State) Fields(id string) []Field {
	switch id {
	case "connection":
		return []Field{
			{"Server URL", s.ServerURL, "wss:// or ws:// scheme required"},
			{"Auth token", "••••••••••••••••", "stored encrypted in $XDG_DATA_HOME/daemonctl"},
			{"Last used", "2026-04-14 18:42", ""},
		}
	case "workspace":
		return []Field{
			{"Working directory", s.Workdir, "default cwd for new sessions"},
			{"Max concurrent sessions", fmt.Sprintf("%d", s.MaxSess), "fleet size cap"},
		}
	case "agents":
		return []Field{
			{"codex binary", "/usr/local/bin/codex", "auto-detected"},
			{"opencode binary", "/usr/local/bin/opencode", "auto-detected"},
			{"Default agent", s.Agent, ""},
			{"Default model", s.Model, ""},
		}
	case "appearance":
		themeLabel := "charm-dark"
		if s.Theme != nil {
			themeLabel = s.Theme.Label
		}
		return []Field{
			{"Theme", themeLabel, "⌃t cycles themes live"},
			{"Density", "comfortable", "⌃d toggles"},
			{"Font hints", "any Nerd-Font-aware monospace", ""},
		}
	case "notifications":
		return []Field{
			{"Terminal bell", "on attention", ""},
			{"Desktop on attention", "enabled", "via OSC 777"},
			{"Desktop on complete", "disabled", ""},
			{"Desktop on fail", "disabled", ""},
		}
	case "keyboard":
		return []Field{
			{"Bindings", "default", "press ? for the full list"},
			{"Leader", "ctrl", ""},
		}
	case "advanced":
		return []Field{
			{"Log level", s.LogLevel, ""},
			{"Log file", s.LogFile, ""},
			{"Telemetry", "off", "opt-in only"},
		}
	case "about":
		return []Field{
			{"Version", s.Version, ""},
			{"Build", s.Build, ""},
			{"Docs", "https://github.com/SamWongML/daemon-client", "OSC 8 link (Ghostty)"},
		}
	}
	return nil
}

// Render returns the full-screen settings view.
func Render(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if w < 60 || h < 16 {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			st.TooSmall.Render("Too small — please resize (min 60×16)"))
	}

	sidebarW := 26
	if sidebarW > w/3 {
		sidebarW = w / 3
	}
	mainW := w - sidebarW - 2

	left := renderCategories(t, st, s.Selected, sidebarW, h)
	right := renderPanel(t, st, s, mainW, h)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)

	title := st.SettingsTitle.Render("  Settings") +
		st.Dim.Render(strings.Repeat(" ", max(0, w-lipgloss.Width("  Settings")-lipgloss.Width(" esc back • ⌃s save "))) + " esc back • ⌃s save ")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}

func renderCategories(t *theme.Theme, st *theme.Styles, selected, w, h int) string {
	var lines []string
	lines = append(lines, st.SettingsField.Render("  CATEGORIES"))
	lines = append(lines, "")
	for i, c := range Categories {
		row := fmt.Sprintf(" %s %s", c.Icon, c.Title)
		if i == selected {
			bar := lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
			lines = append(lines, bar+st.SettingsActive.Render(row))
		} else {
			lines = append(lines, " "+st.SettingsCategory.Render(row))
		}
	}
	block := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(w).
		Height(max(3, h-2)).
		Render(block)
}

func renderPanel(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if s.Selected < 0 || s.Selected >= len(Categories) {
		s.Selected = 0
	}
	cat := Categories[s.Selected]
	fields := s.Fields(cat.ID)

	title := st.SettingsTitle.Render(cat.Icon + "  " + cat.Title)

	var rows []string
	rows = append(rows, title)
	rows = append(rows, "")
	for _, f := range fields {
		label := st.SettingsField.Render(pad(f.Label, 24))
		val := st.SettingsValue.Render(f.Value)
		row := "  " + label + val
		if f.Hint != "" {
			row += "\n  " + st.SettingsField.Render(pad("", 24)) +
				st.SettingsField.Render(f.Hint)
		}
		rows = append(rows, row)
		rows = append(rows, "")
	}
	body := strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(w).
		Height(max(5, h-3)).
		Render(body)
}

func pad(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
