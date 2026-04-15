// Package help renders the keyboard-shortcut help overlay (§7.5).
package help

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Binding struct {
	Key  string
	Desc string
}

type Group struct {
	Title    string
	Bindings []Binding
}

// Groups is the canonical list. Bubble Tea's `help.Model` is overkill for a
// static overlay; this is phase 1 — keep it declarative.
var Groups = []Group{
	{"Global", []Binding{
		{"?", "toggle help"},
		{"⌃p", "command palette"},
		{"⌃,", "settings"},
		{"⌃c⌃c", "quit (double-tap)"},
		{"⌃b", "toggle sidebar"},
		{"⌃n", "new session"},
		{"⌃t", "cycle theme"},
		{"tab / S-tab", "next / prev pane"},
	}},
	{"Sidebar", []Binding{
		{"j/k ↓/↑", "move"},
		{"g/G", "top / bottom"},
		{"↵ or l", "open session"},
		{"/", "filter"},
		{"s", "stop"},
		{"r", "resume"},
		{"x", "kill"},
		{"d", "archive"},
	}},
	{"Transcript", []Binding{
		{"j/k", "scroll"},
		{"⌃d/⌃u", "half page"},
		{"g/G", "top / bottom"},
		{"/", "search"},
		{"y", "yank → clipboard"},
		{"space", "toggle tool call"},
		{"i", "focus input"},
	}},
	{"Input", []Binding{
		{"↵", "send"},
		{"⌃j / M-↵", "newline"},
		{"⌃e", "open in $EDITOR"},
		{"esc", "exit input"},
		{"1/2/3", "perm decision"},
	}},
	{"Dialogs", []Binding{
		{"1–9", "select option"},
		{"←/→", "move focus"},
		{"↵", "confirm"},
		{"esc", "cancel"},
	}},
	{"Dev (--dev)", []Binding{
		{"⌃⌥p", "inject permission"},
		{"⌃⌥q", "inject question"},
		{"⌃⌥t", "inject toast"},
		{"⌃⌥c", "force-complete"},
		{"⌃⌥f", "force-fail"},
	}},
}

// Render returns the fully-styled help box. Caller composites it over the
// dimmed main view.
func Render(t *theme.Theme, w, h int) string {
	boxW := 120
	if x := w - 4; x < boxW {
		boxW = x
	}
	boxH := 35
	if y := h - 4; y < boxH {
		boxH = y
	}
	if boxW < 50 || boxH < 12 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent).
			Background(t.Bg).
			Foreground(t.Fg).
			Padding(0, 1).
			Render("help unavailable — terminal too small")
	}

	title := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("Keyboard shortcuts")

	mid := (len(Groups) + 1) / 2
	leftCol := renderColumn(t, Groups[:mid])
	rightCol := renderColumn(t, Groups[mid:])

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		leftCol,
		strings.Repeat(" ", 4),
		rightCol,
	)

	footer := lipgloss.NewStyle().Foreground(t.Dim).Italic(true).
		Render("press ? or esc to close")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		body,
		"",
		footer,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Background(t.Bg).
		Foreground(t.Fg).
		Padding(1, 2).
		Width(boxW - 4).
		Render(content)
}

func renderColumn(t *theme.Theme, gs []Group) string {
	var lines []string
	gStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	kStyle := lipgloss.NewStyle().Foreground(t.Accent)
	dStyle := lipgloss.NewStyle().Foreground(t.Fg)
	for i, g := range gs {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, gStyle.Render(g.Title))
		for _, b := range g.Bindings {
			key := kStyle.Render(pad(b.Key, 14))
			lines = append(lines, "  "+key+dStyle.Render(b.Desc))
		}
	}
	return strings.Join(lines, "\n")
}

func pad(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}
