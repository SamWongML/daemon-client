package sidebar

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/session"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	Sessions []*session.Session
	Selected int  // index into flattened list
	Focused  bool
	ScrollTop int
}

// Flatten groups sessions by status priority and returns a display order suitable for
// rendering + selection. Group headers are not selectable; they're interleaved in the
// render step only.
func Flatten(ss []*session.Session) []*session.Session {
	out := append([]*session.Session(nil), ss...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Status.Priority() < out[j].Status.Priority()
	})
	return out
}

func Render(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(t.Dim)
	grp := lipgloss.NewStyle().Foreground(t.Dim).Italic(true)

	var lines []string
	active := 0
	for _, sess := range s.Sessions {
		if sess.Status == session.StatusRunning || sess.Status == session.StatusAwaitingInput || sess.Status == session.StatusAwaitingPerm {
			active++
		}
	}
	title := accent.Render("SESSIONS") + dim.Render(fmt.Sprintf("  (%d/%d)", active, len(s.Sessions)))
	lines = append(lines, title)
	lines = append(lines, dim.Render(strings.Repeat("─", max(0, w-1))))

	var currentGroup session.Status = -1
	sessions := Flatten(s.Sessions)
	for i, sess := range sessions {
		if sess.Status != currentGroup {
			currentGroup = sess.Status
			label := groupLabel(sess.Status, countStatus(sessions, sess.Status))
			lines = append(lines, grp.Render("── "+label+" "+strings.Repeat("─", max(0, w-len(label)-4))))
		}
		lines = append(lines, renderItem(t, sess, i == s.Selected, s.Focused, w))
	}

	// scroll window
	window := lines
	if len(lines) > h {
		start := s.ScrollTop
		if start < 0 {
			start = 0
		}
		if start > len(lines)-h {
			start = len(lines) - h
		}
		window = lines[start : start+h]
	}
	for len(window) < h {
		window = append(window, "")
	}
	return strings.Join(window, "\n")
}

func countStatus(ss []*session.Session, st session.Status) int {
	n := 0
	for _, s := range ss {
		if s.Status == st {
			n++
		}
	}
	return n
}

func groupLabel(st session.Status, n int) string {
	name := st.Name()
	switch st {
	case session.StatusAwaitingPerm, session.StatusAwaitingInput:
		name = "needs attention"
	}
	return fmt.Sprintf("%s (%d)", name, n)
}

func renderItem(t *theme.Theme, sess *session.Session, selected, focused bool, w int) string {
	statusCol := theme.StatusColor(t, sess.Status.Name())
	glyph := lipgloss.NewStyle().Foreground(statusCol).Render(sess.Status.Glyph())
	titleStyle := lipgloss.NewStyle().Foreground(t.Fg)
	if selected {
		titleStyle = titleStyle.Bold(true)
	}
	title := titleStyle.Render(truncate(sess.Title, max(0, w-6)))
	sub := lipgloss.NewStyle().Foreground(t.Dim).Render(
		fmt.Sprintf("  %s • %s", sess.Status.Name(), sess.Agent),
	)
	prefix := "  "
	if selected {
		barCol := t.Dim
		if focused {
			barCol = t.Accent
		}
		prefix = lipgloss.NewStyle().Foreground(barCol).Render("▌ ")
	}
	line1 := prefix + glyph + " " + title
	line2 := sub
	return line1 + "\n" + line2
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w < 2 {
		return string(r[:w])
	}
	return string(r[:w-1]) + "…"
}
