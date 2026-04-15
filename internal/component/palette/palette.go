// Package palette implements the command palette (§7.6).
//
// Actions are registered declaratively by the app during init. The palette is
// UI-only — invoking an action returns its ID; the app owns the dispatch.
package palette

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Action struct {
	ID       string
	Title    string
	Icon     string
	Shortcut string
	Category string
}

type Model struct {
	actions  []Action
	query    string
	selected int
	filtered []Action
}

func New(actions []Action) *Model {
	m := &Model{actions: actions}
	m.filter()
	return m
}

func (m *Model) Query() string             { return m.query }
func (m *Model) Filtered() []Action        { return m.filtered }
func (m *Model) SelectedIndex() int        { return m.selected }
func (m *Model) Selected() (Action, bool)  {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return Action{}, false
	}
	return m.filtered[m.selected], true
}

func (m *Model) Reset() {
	m.query = ""
	m.selected = 0
	m.filter()
}

func (m *Model) SetQuery(q string) {
	m.query = q
	m.filter()
}

func (m *Model) AppendRune(r rune) {
	m.query += string(r)
	m.filter()
}

func (m *Model) Backspace() {
	if len(m.query) == 0 {
		return
	}
	r := []rune(m.query)
	m.query = string(r[:len(r)-1])
	m.filter()
}

func (m *Model) Move(delta int) {
	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m *Model) filter() {
	q := strings.ToLower(strings.TrimSpace(m.query))
	if q == "" {
		m.filtered = append([]Action(nil), m.actions...)
		if m.selected >= len(m.filtered) {
			m.selected = 0
		}
		return
	}
	type ranked struct {
		a Action
		r int
	}
	var rs []ranked
	for _, a := range m.actions {
		r := score(a, q)
		if r < 0 {
			continue
		}
		rs = append(rs, ranked{a, r})
	}
	sort.SliceStable(rs, func(i, j int) bool { return rs[i].r < rs[j].r })
	out := make([]Action, len(rs))
	for i, r := range rs {
		out[i] = r.a
	}
	m.filtered = out
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func score(a Action, q string) int {
	title := strings.ToLower(a.Title)
	id := strings.ToLower(a.ID)
	if i := strings.Index(title, q); i >= 0 {
		return i
	}
	if i := strings.Index(id, q); i >= 0 {
		return 100 + i
	}
	if fuzzy(title, q) {
		return 500
	}
	return -1
}

func fuzzy(s, q string) bool {
	i := 0
	for _, c := range s {
		if i < len(q) && byte(c) == q[i] {
			i++
		}
	}
	return i == len(q)
}

// Render draws the palette modal. Caller composites it over the main view.
func Render(t *theme.Theme, m *Model, w, h int) string {
	boxW := 60
	if boxW > w-6 {
		boxW = w - 6
	}
	boxH := 20
	if boxH > h-4 {
		boxH = h - 4
	}
	if boxW < 30 || boxH < 8 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent).
			Background(t.Bg).Foreground(t.Fg).
			Padding(0, 1).
			Render("palette unavailable — terminal too small")
	}

	inner := boxW - 4

	promptStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	inputLine := promptStyle.Render("> ") +
		lipgloss.NewStyle().Foreground(t.Fg).Render(m.query) +
		lipgloss.NewStyle().Foreground(t.Accent).Render("▏")

	divider := lipgloss.NewStyle().Foreground(t.Border).
		Render(strings.Repeat("─", inner))

	maxRows := boxH - 7
	if maxRows < 3 {
		maxRows = 3
	}
	if maxRows > 12 {
		maxRows = 12
	}

	start := 0
	if m.selected >= maxRows {
		start = m.selected - maxRows + 1
	}
	end := start + maxRows
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, renderRow(t, m.filtered[i], i == m.selected, inner))
	}
	if len(rows) == 0 {
		rows = []string{lipgloss.NewStyle().Foreground(t.Dim).Italic(true).
			Render("  no matches")}
	}
	for len(rows) < maxRows {
		rows = append(rows, "")
	}
	list := strings.Join(rows, "\n")

	hint := lipgloss.NewStyle().Foreground(t.Dim).Render(
		"↑↓ navigate   ↵ run   esc close")

	body := lipgloss.JoinVertical(lipgloss.Left,
		inputLine,
		divider,
		list,
		divider,
		hint,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Background(t.Bg).
		Foreground(t.Fg).
		Padding(0, 1).
		Width(inner).
		Render(body)
}

func renderRow(t *theme.Theme, a Action, selected bool, w int) string {
	icon := a.Icon
	if icon == "" {
		icon = "·"
	}
	iconStyle := lipgloss.NewStyle().Foreground(t.Accent)
	titleStyle := lipgloss.NewStyle().Foreground(t.Fg)
	shortStyle := lipgloss.NewStyle().Foreground(t.Dim)
	prefix := "  "
	if selected {
		titleStyle = titleStyle.Bold(true)
		prefix = lipgloss.NewStyle().Foreground(t.Accent).Render("▌ ")
	}
	left := fmt.Sprintf("%s%s  %s", prefix, iconStyle.Render(icon), titleStyle.Render(a.Title))
	right := ""
	if a.Shortcut != "" {
		right = shortStyle.Render(a.Shortcut)
	}
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}
