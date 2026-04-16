package header

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	Agent string
	Model string
	// Focused is "settings", "help", or "" — drives the focused-button
	// highlight (accent left bar + bold label).
	Focused string
}

func Render(t *theme.Theme, st *theme.Styles, s State, w int) string {
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	dot := lipgloss.NewStyle().Foreground(t.Success).Render("●")
	fg := lipgloss.NewStyle().Foreground(t.Fg)

	// Left group: wordmark + transport dot + agent · model
	left := accent.Render("▌daemonctl") + "   " +
		dot + " " + fg.Render(s.Agent+" · "+s.Model)

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
