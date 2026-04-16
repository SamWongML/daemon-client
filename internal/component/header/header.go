package header

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	ServerURL    string
	Active       int
	Max          int
	TotalCost    float64
	Now          time.Time
	TerminalName string // e.g. "ghostty", "kitty" — from ghostty.Caps.Label()
}

func Render(t *theme.Theme, st *theme.Styles, s State, w int) string {
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	dot := lipgloss.NewStyle().Foreground(t.Success).Render("●")
	dim := lipgloss.NewStyle().Foreground(t.Dim)
	fg := lipgloss.NewStyle().Foreground(t.Fg)

	parts := []string{
		accent.Render("▌daemonctl"),
		dot + " " + fg.Render(truncMiddle(s.ServerURL, 32)),
		fg.Render(fmt.Sprintf("%d/%d sessions", s.Active, s.Max)),
		dim.Render("⏱ " + s.Now.Format("15:04")),
		dim.Render(fmt.Sprintf("$%.2f", s.TotalCost)),
	}
	if s.TerminalName != "" {
		parts = append(parts, dim.Render(s.TerminalName))
	}
	parts = append(parts, dim.Render("[⚙]"), dim.Render("[?]"))

	sep := dim.Render(" • ")
	line := strings.Join(parts, sep)
	if lipgloss.Width(line) > w {
		line = ansiTruncate(line, w)
	}
	return line + strings.Repeat(" ", max(0, w-lipgloss.Width(line)))
}

func truncMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	half := (max - 1) / 2
	return s[:half] + "…" + s[len(s)-half:]
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
