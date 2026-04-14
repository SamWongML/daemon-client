package splash

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/demon/daemon-client/internal/theme"
)

const logo = `
    ██████╗  █████╗ ███████╗███╗   ███╗ ██████╗ ███╗   ██╗
    ██╔══██╗██╔══██╗██╔════╝████╗ ████║██╔═══██╗████╗  ██║
    ██║  ██║███████║█████╗  ██╔████╔██║██║   ██║██╔██╗ ██║
    ██║  ██║██╔══██║██╔══╝  ██║╚██╔╝██║██║   ██║██║╚██╗██║
    ██████╔╝██║  ██║███████╗██║ ╚═╝ ██║╚██████╔╝██║ ╚████║
    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
`

const (
	Duration = 800 * time.Millisecond
)

type DoneMsg struct{}

type tickMsg struct{}

// Tick is fired once; when Duration elapses we emit DoneMsg so the app routes to main.
func Tick() tea.Cmd {
	return tea.Tick(Duration, func(time.Time) tea.Msg { return DoneMsg{} })
}

func Render(t *theme.Theme, st *theme.Styles, w, h int) string {
	accent := lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	dim := lipgloss.NewStyle().Foreground(t.Dim)
	fg := lipgloss.NewStyle().Foreground(t.Fg).Bold(true)

	logoBlock := accent.Render(strings.TrimRight(logo, "\n"))
	wordmark := fg.Render("d a e m o n c t l")
	tagline := dim.Italic(true).Render("the coding-agent fleet controller")
	version := dim.Render("v0.1.0 • showcase build")
	hint := dim.Render("press any key to skip…")

	content := lipgloss.JoinVertical(lipgloss.Center,
		logoBlock,
		"",
		wordmark,
		tagline,
		"",
		version,
		"",
		hint,
	)
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content)
}

// Suppress unused import warning when we add more msg types later.
var _ = tickMsg{}
