// Package dialog implements the four modal dialog types: alert, confirm,
// question (1–9 numbered), and permission (§7.7).
package dialog

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

type Kind int

const (
	KindAlert Kind = iota
	KindConfirm
	KindQuestion
	KindPermission
)

type Model struct {
	Kind    Kind
	Title   string
	Message string
	// Options — for question: arbitrary 1..9; for confirm: ["Cancel","Confirm"];
	// for alert: ["OK"]; for permission: always ["Allow once","Allow always","Deny"].
	Options []string
	// Permission-only.
	Tool string
	Args string
	Diff string

	Selected int
}

// Result values returned by HandleKey's first return.
const (
	ResultCancelled = -1
	ResultPending   = -2
)

func Alert(title, message string) *Model {
	return &Model{Kind: KindAlert, Title: title, Message: message, Options: []string{"OK"}}
}

func Confirm(title, message string) *Model {
	return &Model{
		Kind:     KindConfirm,
		Title:    title,
		Message:  message,
		Options:  []string{"Cancel", "Confirm"},
		Selected: 1,
	}
}

func Question(title, prompt string, opts []string) *Model {
	if len(opts) > 9 {
		opts = opts[:9]
	}
	return &Model{Kind: KindQuestion, Title: title, Message: prompt, Options: opts}
}

func Permission(tool, args, diff string) *Model {
	return &Model{
		Kind:    KindPermission,
		Title:   "Tool permission required",
		Tool:    tool,
		Args:    args,
		Diff:    diff,
		Options: []string{"Allow once", "Allow always", "Deny"},
	}
}

func (m *Model) Move(d int) {
	if len(m.Options) == 0 {
		return
	}
	m.Selected += d
	if m.Selected < 0 {
		m.Selected = 0
	}
	if m.Selected >= len(m.Options) {
		m.Selected = len(m.Options) - 1
	}
}

// HandleKey processes a single key press. Returns (decision, done). If done
// is true the caller closes the modal; decision is the option index selected
// or ResultCancelled for esc.
func (m *Model) HandleKey(key string) (int, bool) {
	switch key {
	case "esc":
		return ResultCancelled, true
	case "left", "h", "shift+tab":
		m.Move(-1)
		return ResultPending, false
	case "right", "l", "tab":
		m.Move(1)
		return ResultPending, false
	case "enter", " ":
		return m.Selected, true
	}
	if m.Kind == KindQuestion || m.Kind == KindPermission {
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			idx := int(key[0] - '1')
			if idx < len(m.Options) {
				m.Selected = idx
				return idx, true
			}
		}
	}
	return ResultPending, false
}

func Render(t *theme.Theme, m *Model, w, h int) string {
	boxW := 80
	if boxW > w-6 {
		boxW = w - 6
	}
	if boxW < 40 {
		boxW = 40
		if boxW > w-4 {
			boxW = w - 4
		}
	}
	inner := boxW - 4

	titleStyle := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(t.Dim)
	fg := lipgloss.NewStyle().Foreground(t.Fg)

	var lines []string
	lines = append(lines, titleStyle.Render(m.Title))
	lines = append(lines, "")

	switch m.Kind {
	case KindPermission:
		lines = append(lines, dim.Render("tool"))
		lines = append(lines, "  "+fg.Render(m.Tool))
		lines = append(lines, "")
		lines = append(lines, dim.Render("args"))
		lines = append(lines, "  "+fg.Render(m.Args))
		if strings.TrimSpace(m.Diff) != "" {
			lines = append(lines, "", dim.Render("diff"))
			for _, ln := range strings.Split(m.Diff, "\n") {
				lines = append(lines, "  "+fg.Render(ln))
			}
		}

	case KindQuestion:
		if m.Message != "" {
			for _, ln := range wrap(m.Message, inner) {
				lines = append(lines, fg.Render(ln))
			}
			lines = append(lines, "")
		}
		for i, opt := range m.Options {
			num := fmt.Sprintf(" %d) ", i+1)
			style := fg
			if i == m.Selected {
				style = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
			}
			lines = append(lines, style.Render(num+opt))
		}

	default: // alert, confirm
		if m.Message != "" {
			for _, ln := range wrap(m.Message, inner) {
				lines = append(lines, fg.Render(ln))
			}
		}
	}

	if m.Kind != KindQuestion {
		lines = append(lines, "")
		lines = append(lines, renderButtons(t, m))
	}

	body := strings.Join(lines, "\n")

	borderCol := borderColor(t, m.Kind)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Background(t.Bg).
		Foreground(t.Fg).
		Padding(1, 2).
		Width(inner).
		Render(body)
}

func borderColor(t *theme.Theme, k Kind) color.Color {
	switch k {
	case KindPermission:
		return t.Warn
	case KindAlert:
		return t.Danger
	case KindConfirm:
		return t.Accent
	}
	return t.Accent
}

func renderButtons(t *theme.Theme, m *Model) string {
	var parts []string
	for i, opt := range m.Options {
		label := opt
		if m.Kind == KindPermission {
			label = fmt.Sprintf("%d %s", i+1, opt)
		}
		var style lipgloss.Style
		if i == m.Selected {
			style = lipgloss.NewStyle().
				Background(t.Accent).
				Foreground(t.Bg).
				Bold(true).
				Padding(0, 1)
			parts = append(parts, style.Render(label))
		} else {
			style = lipgloss.NewStyle().Foreground(t.Dim).Padding(0, 1)
			parts = append(parts, style.Render("["+label+"]"))
		}
	}
	return strings.Join(parts, "  ")
}

func wrap(s string, w int) []string {
	if w <= 0 {
		return []string{s}
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		for len(para) > w {
			cut := w
			if idx := strings.LastIndex(para[:cut], " "); idx > 0 {
				cut = idx
			}
			out = append(out, para[:cut])
			para = strings.TrimLeft(para[cut:], " ")
		}
		out = append(out, para)
	}
	return out
}
