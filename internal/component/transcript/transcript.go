package transcript

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	Markdown   string
	ScrollTop  int
	AutoFollow bool
	Focused    bool
}

// Render pipes the transcript markdown through glamour and returns a height-h window.
// Re-renders on every call; for a phase-1 showcase this is fine, phase 2 should cache.
func Render(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(max(20, w-2)),
	)
	var body string
	if err == nil && strings.TrimSpace(s.Markdown) != "" {
		out, rerr := renderer.Render(s.Markdown)
		if rerr == nil {
			body = out
		}
	}
	if body == "" {
		if s.Markdown == "" {
			body = lipgloss.NewStyle().Foreground(t.Dim).Italic(true).Render("  (select a session to view its transcript)")
		} else {
			body = s.Markdown
		}
	}

	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	start := s.ScrollTop
	if s.AutoFollow && len(lines) > h {
		start = len(lines) - h
	}
	if start < 0 {
		start = 0
	}
	if start > max(0, len(lines)-h) {
		start = max(0, len(lines)-h)
	}
	end := start + h
	if end > len(lines) {
		end = len(lines)
	}
	window := lines[start:end]
	for len(window) < h {
		window = append(window, "")
	}

	bar := " "
	if s.Focused {
		bar = lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
	}
	for i, l := range window {
		window[i] = bar + l
	}
	return strings.Join(window, "\n")
}
