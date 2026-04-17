package footer

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Hint struct {
	Key  string
	Desc string
}

// Status holds the right-zone ambient info that moved from the header.
type Status struct {
	Active    int
	Max       int
	TotalCost float64
	Now       time.Time
}

func Render(t *theme.Theme, st *theme.Styles, hints []Hint, status Status, w int) string {
	keyStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(t.Dim)
	dim := lipgloss.NewStyle().Foreground(t.Dim)

	// Left zone: keybind hints separated by two spaces
	var parts []string
	for _, h := range hints {
		parts = append(parts, keyStyle.Render(h.Key)+" "+descStyle.Render(h.Desc))
	}
	left := " " + strings.Join(parts, "  ")

	// Right zone: session count, cost, clock — drop right-to-left on overflow
	rightItems := []string{
		fmt.Sprintf("%d/%d", status.Active, status.Max),
		fmt.Sprintf("$%.2f", status.TotalCost),
		status.Now.Format("15:04"),
	}
	right := ""
	for _, item := range rightItems {
		candidate := dim.Render(item)
		if right == "" {
			candidate = candidate + " "
		} else {
			candidate = right + "  " + candidate + " "
		}
		if lipgloss.Width(left)+lipgloss.Width(candidate)+2 > w {
			break
		}
		right = strings.TrimRight(candidate, " ")
	}

	gap := max(1, w-lipgloss.Width(left)-lipgloss.Width(right)-1)
	line := left + strings.Repeat(" ", gap) + right

	if lipgloss.Width(line) > w {
		r := []rune(line)
		for lipgloss.Width(string(r)) > w-1 && len(r) > 0 {
			r = r[:len(r)-1]
		}
		line = string(r) + "…"
	}
	return line
}
