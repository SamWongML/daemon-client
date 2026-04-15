// Package toast implements the bottom-right notification stack (§7.8).
package toast

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Error
)

type Toast struct {
	ID      int64
	Kind    Kind
	Title   string
	Body    string
	Created time.Time
}

type Stack struct {
	items []Toast
	next  int64
}

// MaxVisible is the max simultaneously-shown toasts. Older ones slide off.
const MaxVisible = 3

// TTL is how long a toast stays before auto-dismiss.
const TTL = 4 * time.Second

func (s *Stack) Push(t Toast) {
	s.next++
	t.ID = s.next
	t.Created = time.Now()
	s.items = append(s.items, t)
}

// Tick drops expired toasts. Call from the clock tick in app.Update.
func (s *Stack) Tick(now time.Time) bool {
	changed := false
	filtered := s.items[:0]
	for _, t := range s.items {
		if now.Sub(t.Created) < TTL {
			filtered = append(filtered, t)
		} else {
			changed = true
		}
	}
	s.items = filtered
	return changed
}

func (s *Stack) Visible() []Toast {
	if len(s.items) <= MaxVisible {
		out := make([]Toast, len(s.items))
		copy(out, s.items)
		return out
	}
	tail := s.items[len(s.items)-MaxVisible:]
	out := make([]Toast, len(tail))
	copy(out, tail)
	return out
}

func (s *Stack) Empty() bool { return len(s.items) == 0 }

// Render lays out visible toasts as a vertical stack. Width is clamped to
// min(40, w-4). Caller composites it bottom-right on the base view.
func Render(t *theme.Theme, items []Toast, w int) string {
	if len(items) == 0 {
		return ""
	}
	tw := 40
	if tw > w-4 {
		tw = w - 4
	}
	if tw < 20 {
		tw = 20
	}
	var boxes []string
	for _, ts := range items {
		boxes = append(boxes, renderOne(t, ts, tw))
	}
	return strings.Join(boxes, "\n")
}

func renderOne(t *theme.Theme, ts Toast, w int) string {
	var col color.Color
	var icon string
	switch ts.Kind {
	case Success:
		col = t.Success
		icon = "✓"
	case Warning:
		col = t.Warn
		icon = "⚠"
	case Error:
		col = t.Danger
		icon = "✗"
	default:
		col = t.Info
		icon = "ℹ"
	}
	iconStyle := lipgloss.NewStyle().Foreground(col).Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(t.Fg).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(t.Dim)

	header := iconStyle.Render(icon) + " " + titleStyle.Render(ts.Title)
	lines := []string{header}
	if ts.Body != "" {
		lines = append(lines, bodyStyle.Render(ts.Body))
	}
	body := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(col).
		Background(t.Bg).
		Padding(0, 1).
		Width(w - 4).
		Render(body)
}
