package sidebar

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/session"
	"github.com/demon/daemon-client/internal/theme"
)

type State struct {
	Sessions   []*session.Session
	Selected   int // index into flattened list
	Focused    bool
	ScrollTop  int
	Now        time.Time // current time for bucket calculation
	PulsePhase float64   // 0.0..1.0 sine wave phase for glyph animation
}

// timeBucket classifies sessions into display groups.
type timeBucket int

const (
	bucketActive timeBucket = iota
	bucketToday
	bucketThisWeek
	bucketOlder
	bucketArchive
)

func bucketLabel(b timeBucket) string {
	switch b {
	case bucketActive:
		return "active"
	case bucketToday:
		return "today"
	case bucketThisWeek:
		return "this week"
	case bucketOlder:
		return "older"
	case bucketArchive:
		return "archive"
	}
	return ""
}

func sessionBucket(sess *session.Session, now time.Time) timeBucket {
	// Active: any session currently in running/awaiting/starting
	switch sess.Status {
	case session.StatusRunning, session.StatusAwaitingInput, session.StatusAwaitingPerm, session.StatusStarting:
		return bucketActive
	}
	age := now.Sub(sess.StartedAt)
	if age < 24*time.Hour {
		return bucketToday
	}
	if age < 7*24*time.Hour {
		return bucketThisWeek
	}
	return bucketOlder
}

// Flatten sorts sessions by time bucket then severity descending within each bucket.
func Flatten(ss []*session.Session) []*session.Session {
	return FlattenAt(ss, time.Now())
}

// FlattenAt sorts sessions with an explicit reference time.
func FlattenAt(ss []*session.Session, now time.Time) []*session.Session {
	out := append([]*session.Session(nil), ss...)
	sort.SliceStable(out, func(i, j int) bool {
		bi := sessionBucket(out[i], now)
		bj := sessionBucket(out[j], now)
		if bi != bj {
			return bi < bj // lower bucket index = higher priority
		}
		// Within same bucket: severity descending
		si := out[i].Status.Severity()
		sj := out[j].Status.Severity()
		if si != sj {
			return si > sj
		}
		// Tie-break: most recent activity first
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out
}

func Render(t *theme.Theme, st *theme.Styles, s State, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	now := s.Now
	if now.IsZero() {
		now = time.Now()
	}

	dim := lipgloss.NewStyle().Foreground(t.Dim)
	muted := lipgloss.NewStyle().Foreground(t.Muted)

	var lines []string

	// "New session" row (sticky, line 0)
	newSessLine := renderNewSession(t, s.Selected == -1, s.Focused, w)
	lines = append(lines, newSessLine)

	// Title row
	title := muted.Render(" sessions") + dim.Render(fmt.Sprintf("   %d", len(s.Sessions)))
	lines = append(lines, title)
	lines = append(lines, dim.Render(strings.Repeat("─", max(0, w))))

	sessions := FlattenAt(s.Sessions, now)
	var currentBucket timeBucket = -1
	for i, sess := range sessions {
		b := sessionBucket(sess, now)
		if b != currentBucket {
			currentBucket = b
			label := bucketLabel(b)
			ruleW := max(0, w-len(label)-5)
			lines = append(lines, dim.Render("─── "+label+" "+strings.Repeat("─", ruleW)))
		}
		lines = append(lines, renderItem(t, sess, i == s.Selected, s.Focused, w, s.PulsePhase)...)
	}

	// Scroll window
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

func renderNewSession(t *theme.Theme, selected, focused bool, w int) string {
	prefix := " "
	labelStyle := lipgloss.NewStyle().Foreground(t.Muted)
	if selected {
		barCol := t.Dim
		if focused {
			barCol = t.Accent
		}
		prefix = lipgloss.NewStyle().Foreground(barCol).Render("▌")
		labelStyle = lipgloss.NewStyle().Foreground(t.Fg).Bold(true)
	}
	label := labelStyle.Render("+ new session")
	hint := lipgloss.NewStyle().Foreground(t.Dim).Render("⌃n")
	gap := max(1, w-lipgloss.Width(prefix)-lipgloss.Width(label)-lipgloss.Width(hint)-1)
	return prefix + label + strings.Repeat(" ", gap) + hint
}

func renderItem(t *theme.Theme, sess *session.Session, selected, focused bool, w int, pulsePhase float64) []string {
	statusCol := theme.StatusColor(t, sess.Status.Name())

	isAttention := sess.Status == session.StatusAwaitingInput || sess.Status == session.StatusAwaitingPerm

	// For awaiting states, binary-pulse the glyph color between status color and dim.
	glyphCol := statusCol
	if isAttention && pulsePhase <= 0.5 {
		glyphCol = t.Dim
	}
	glyph := lipgloss.NewStyle().Foreground(glyphCol).Render(sess.Status.Glyph())

	titleStyle := lipgloss.NewStyle().Foreground(t.Fg)
	if selected {
		titleStyle = titleStyle.Bold(true)
	}
	title := titleStyle.Render(truncate(sess.Title, max(0, w-5)))

	// Line 1: prefix + glyph + title
	prefix := " "
	if selected {
		barCol := t.Dim
		if focused {
			barCol = t.Accent
		}
		prefix = lipgloss.NewStyle().Foreground(barCol).Render("▌")
	} else if isAttention {
		// Attention-state rows get persistent left accent bar in theme.Warn
		prefix = lipgloss.NewStyle().Foreground(t.Warn).Render("▌")
	}
	line1 := prefix + glyph + " " + title

	// Line 2: status label + elapsed
	elapsed := formatElapsed(time.Since(sess.StartedAt))
	sub := lipgloss.NewStyle().Foreground(t.Dim).Render(
		"   " + sess.Status.Name() + " · " + elapsed,
	)
	line2 := sub

	return []string{line1, line2}
}

func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
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
