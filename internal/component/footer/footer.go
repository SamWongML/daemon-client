package footer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Hint struct {
	Key  string
	Desc string
}

func Render(t *theme.Theme, st *theme.Styles, hints []Hint, w int) string {
	keyStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(t.Dim)

	var parts []string
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Key)+" "+descStyle.Render(h.Desc))
	}
	sep := descStyle.Render(" │ ")
	line := strings.Join(parts, sep)
	if lipgloss.Width(line) > w {
		// truncate with ellipsis
		r := []rune(line)
		for lipgloss.Width(string(r)) > w-1 && len(r) > 0 {
			r = r[:len(r)-1]
		}
		line = string(r) + "…"
	}
	return " " + line + strings.Repeat(" ", max(0, w-1-lipgloss.Width(line)))
}
