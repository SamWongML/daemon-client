package input

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	Buffer      string
	Focused     bool
	Placeholder string
}

func Render(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	border := lipgloss.NewStyle().Foreground(t.Border)
	glyph := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("> ")

	content := s.Buffer
	if content == "" && !s.Focused {
		content = lipgloss.NewStyle().Foreground(t.Dim).Italic(true).Render(s.Placeholder)
	} else if content == "" {
		content = lipgloss.NewStyle().Foreground(t.Dim).Render("▏")
	} else if s.Focused {
		content += lipgloss.NewStyle().Foreground(t.Accent).Render("▏")
	}

	inner := glyph + content
	lines := strings.Split(inner, "\n")
	for len(lines) < h-2 {
		lines = append(lines, "")
	}
	if len(lines) > h-2 {
		lines = lines[:h-2]
	}

	bar := " "
	if s.Focused {
		bar = lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
	}
	top := border.Render("╭" + strings.Repeat("─", max(0, w-2)) + "╮")
	bot := border.Render("╰" + strings.Repeat("─", max(0, w-2)) + "╯")
	body := make([]string, len(lines))
	for i, l := range lines {
		left := bar + border.Render("│")
		right := border.Render("│")
		body[i] = left + padRight(l, max(0, w-3)) + right
	}
	return strings.Join(append([]string{top}, append(body, bot)...), "\n")
}

func padRight(s string, w int) string {
	gap := w - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}
