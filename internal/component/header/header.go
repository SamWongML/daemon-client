package header

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

// WSStatus represents the simulated WebSocket connection state shown in the
// header dot indicator. Phase 1 uses dev cheats to cycle through states.
type WSStatus int

const (
	WSConnected    WSStatus = iota // green ●
	WSReconnecting                 // yellow ●
	WSDisconnected                 // red ●
)

type State struct {
	Agent string
	Model string
	// Focused is "settings", "help", or "" — drives the focused-button
	// highlight (accent left bar + bold label).
	Focused      string
	TerminalName string // e.g. "ghostty", "kitty" — from ghostty.Caps.Label()
	WS           WSStatus
}

func Render(t *theme.Theme, st *theme.Styles, s State, w int) string {
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	var dotColor color.Color
	switch s.WS {
	case WSReconnecting:
		dotColor = t.Warn
	case WSDisconnected:
		dotColor = t.Danger
	default:
		dotColor = t.Success
	}
	dot := lipgloss.NewStyle().Foreground(dotColor).Render("●")
	fg := lipgloss.NewStyle().Foreground(t.Fg)

	// Left group: wordmark + transport dot + agent · model + optional terminal badge
	termBadge := ""
	if s.TerminalName != "" {
		termBadge = "  " + lipgloss.NewStyle().Foreground(t.Dim).Render(s.TerminalName)
	}
	left := accent.Render("▌daemonctl") + "   " +
		dot + " " + fg.Render(s.Agent+" · "+s.Model) + termBadge

	// Right group: [⚙] [?]
	right := renderBtn(t, "⚙", s.Focused == "settings") + " " +
		renderBtn(t, "?", s.Focused == "help")

	gap := max(1, w-lipgloss.Width(left)-lipgloss.Width(right))
	line := left + strings.Repeat(" ", gap) + right
	if lipgloss.Width(line) > w {
		line = ansiTruncate(line, w)
	}
	return line
}

func renderBtn(t *theme.Theme, glyph string, focused bool) string {
	if focused {
		bar := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("▌")
		label := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(glyph)
		return bar + label
	}
	return lipgloss.NewStyle().Foreground(t.Dim).Render("[" + glyph + "]")
}

func ansiTruncate(s string, width int) string {
	// lipgloss helpfully handles ANSI-aware widths; for simplicity we trust lipgloss.Width
	// and use a rune-level fallback cut. Sufficient for the showcase.
	if lipgloss.Width(s) <= width {
		return s
	}
	out := []rune(s)
	for lipgloss.Width(string(out)) > width && len(out) > 0 {
		out = out[:len(out)-1]
	}
	return string(out)
}
