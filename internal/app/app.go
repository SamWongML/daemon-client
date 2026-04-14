package app

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/demon/daemon-client/internal/component/footer"
	"github.com/demon/daemon-client/internal/component/header"
	"github.com/demon/daemon-client/internal/component/input"
	"github.com/demon/daemon-client/internal/component/sidebar"
	"github.com/demon/daemon-client/internal/component/splash"
	"github.com/demon/daemon-client/internal/component/transcript"
	"github.com/demon/daemon-client/internal/layout"
	"github.com/demon/daemon-client/internal/session"
	"github.com/demon/daemon-client/internal/session/mock"
	"github.com/demon/daemon-client/internal/theme"
)

type screen int

const (
	screenSplash screen = iota
	screenMain
)

type Model struct {
	screen        screen
	width, height int

	theme  *theme.Theme
	styles *theme.Styles

	store    *session.Store
	engine   *mock.Engine
	selected int // index into flattened sidebar
	current  session.ID

	focus FocusID

	inputBuf string
	scrollTr int
	autoFol  bool

	sidebarCollapsed bool

	lastCtrlC time.Time

	tickSub tea.Cmd
}

func New(store *session.Store, eng *mock.Engine) *Model {
	t := theme.Dark()
	return &Model{
		screen:   screenSplash,
		theme:    t,
		styles:   theme.BuildStyles(t),
		store:    store,
		engine:   eng,
		focus:    FocusSidebar,
		autoFol:  true,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(splash.Tick(), clockTick())
}

type clockTickMsg struct{}

func clockTick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return clockTickMsg{} })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case clockTickMsg:
		return m, clockTick()

	case splash.DoneMsg:
		m.screen = screenMain
		m.selectInitial()
		return m, nil

	case session.AppendMsg:
		if sess := m.store.Get(msg.ID); sess != nil {
			sess.Transcript += msg.Content
		}
		return m, nil

	case session.StatusMsg:
		m.store.SetStatus(msg.ID, msg.Status)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := k.String()

	// Splash: any key skips.
	if m.screen == screenSplash {
		m.screen = screenMain
		m.selectInitial()
		return m, nil
	}

	// Global
	switch key {
	case "ctrl+c":
		if time.Since(m.lastCtrlC) < 500*time.Millisecond {
			return m, tea.Quit
		}
		m.lastCtrlC = time.Now()
		return m, nil
	case "tab":
		m.focus = m.focus.Next()
		return m, nil
	case "shift+tab":
		m.focus = m.focus.Prev()
		return m, nil
	case "ctrl+b":
		m.sidebarCollapsed = !m.sidebarCollapsed
		return m, nil
	}

	switch m.focus {
	case FocusSidebar:
		return m.handleSidebarKey(key)
	case FocusTranscript:
		return m.handleTranscriptKey(key)
	case FocusInput:
		return m.handleInputKey(k, key)
	}
	return m, nil
}

func (m *Model) handleSidebarKey(key string) (tea.Model, tea.Cmd) {
	sessions := sidebar.Flatten(m.store.All())
	switch key {
	case "q":
		return m, tea.Quit
	case "j", "down":
		if m.selected < len(sessions)-1 {
			m.selected++
		}
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
	case "g":
		m.selected = 0
	case "G":
		m.selected = len(sessions) - 1
	case "enter", "l":
		if m.selected >= 0 && m.selected < len(sessions) {
			sess := sessions[m.selected]
			m.current = sess.ID
			m.autoFol = true
			m.scrollTr = 0
			// Live-status sessions with empty transcripts: start scripted replay.
			if sess.Transcript == "" && isLive(sess.Status) {
				m.engine.StartReplay(sess.ID)
			}
		}
	}
	return m, nil
}

func (m *Model) handleTranscriptKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q":
		return m, tea.Quit
	case "j", "down":
		m.scrollTr++
		m.autoFol = false
	case "k", "up":
		if m.scrollTr > 0 {
			m.scrollTr--
		}
		m.autoFol = false
	case "g":
		m.scrollTr = 0
		m.autoFol = false
	case "G":
		m.autoFol = true
	case "i":
		m.focus = FocusInput
	}
	return m, nil
}

func (m *Model) handleInputKey(k tea.KeyPressMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.focus = FocusTranscript
	case "enter":
		// phase 2 will actually submit
		m.inputBuf = ""
	case "backspace":
		if len(m.inputBuf) > 0 {
			r := []rune(m.inputBuf)
			m.inputBuf = string(r[:len(r)-1])
		}
	default:
		if k.Text != "" {
			m.inputBuf += k.Text
		}
	}
	return m, nil
}

func (m *Model) selectInitial() {
	sessions := sidebar.Flatten(m.store.All())
	if len(sessions) == 0 {
		return
	}
	m.selected = 0
	sess := sessions[0]
	m.current = sess.ID
	if sess.Transcript == "" && isLive(sess.Status) {
		m.engine.StartReplay(sess.ID)
	}
}

func isLive(s session.Status) bool {
	return s == session.StatusRunning || s == session.StatusAwaitingInput || s == session.StatusAwaitingPerm || s == session.StatusStarting
}

// --- View ---

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m *Model) render() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.screen == screenSplash {
		return splash.Render(m.theme, m.styles, m.width, m.height)
	}
	return m.renderMain()
}

func (m *Model) renderMain() string {
	inputLines := 1 + strings.Count(m.inputBuf, "\n")
	if inputLines > 8 {
		inputLines = 8
	}
	l := layout.Compute(m.width, m.height, m.sidebarCollapsed, inputLines)
	if l.TooSmall {
		msg := m.styles.TooSmall.Render("Too small — please resize (min 60×16)")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}

	sessions := sidebar.Flatten(m.store.All())
	active := 0
	totalCost := 0.0
	for _, s := range sessions {
		if isLive(s.Status) {
			active++
		}
		totalCost += s.CostUSD
	}

	hdr := header.Render(m.theme, m.styles, header.State{
		ServerURL: "wss://prod.example.com",
		Active:    active,
		Max:       4,
		TotalCost: totalCost,
		Now:       time.Now(),
	}, l.Width)

	side := ""
	if l.Sidebar.W > 0 {
		side = sidebar.Render(m.theme, m.styles, sidebar.State{
			Sessions: sessions,
			Selected: m.selected,
			Focused:  m.focus == FocusSidebar,
		}, l.Sidebar.W, l.Sidebar.H)
	}

	var sess *session.Session
	if m.current != "" {
		sess = m.store.Get(m.current)
	}
	sessHdr := m.renderSessionHeader(sess, l.SessionHeader.W)

	trState := transcript.State{
		Focused:    m.focus == FocusTranscript,
		ScrollTop:  m.scrollTr,
		AutoFollow: m.autoFol,
	}
	if sess != nil {
		trState.Markdown = sess.Transcript
	}
	tr := transcript.Render(m.theme, m.styles, trState, l.Transcript.W, l.Transcript.H)

	inp := input.Render(m.theme, m.styles, input.State{
		Buffer:      m.inputBuf,
		Focused:     m.focus == FocusInput,
		Placeholder: "ask the agent…",
	}, l.Input.W, l.Input.H)

	ftr := footer.Render(m.theme, m.styles, m.hints(), l.Width)

	// Compose: join sidebar vertically with (sessHdr/tr/inp) then stack header above and footer below.
	mainCol := lipgloss.JoinVertical(lipgloss.Left, sessHdr, tr, inp)
	var body string
	if side == "" {
		body = mainCol
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, side, mainCol)
	}
	return lipgloss.JoinVertical(lipgloss.Left, hdr, body, ftr)
}

func (m *Model) renderSessionHeader(sess *session.Session, w int) string {
	if sess == nil {
		return m.styles.Dim.Render("  (no session selected)") + "\n"
	}
	statusCol := theme.StatusColor(m.theme, sess.Status.Name())
	title := m.styles.SessionTitle.Render("  " + sess.Title)
	right := m.styles.SessionMeta.Render(fmt.Sprintf("%s • %s • $%.2f  ", sess.Agent, sess.Model, sess.CostUSD))
	line1 := fitLine(title, right, w)
	status := lipgloss.NewStyle().Foreground(statusCol).Bold(true).Render(sess.Status.Glyph() + " " + sess.Status.Name())
	tokens := m.styles.SessionMeta.Render(fmt.Sprintf(" • %d / %d tokens", sess.Tokens, sess.Budget))
	line2 := "  " + status + tokens
	return line1 + "\n" + line2
}

func fitLine(left, right string, w int) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) hints() []footer.Hint {
	base := []footer.Hint{
		{Key: "tab", Desc: "next pane"},
		{Key: "↵", Desc: "open/send"},
		{Key: "⌃b", Desc: "sidebar"},
		{Key: "⌃c⌃c", Desc: "quit"},
		{Key: "?", Desc: "help"},
	}
	switch m.focus {
	case FocusSidebar:
		return append([]footer.Hint{{Key: "j/k", Desc: "move"}, {Key: "g/G", Desc: "top/bot"}}, base...)
	case FocusTranscript:
		return append([]footer.Hint{{Key: "j/k", Desc: "scroll"}, {Key: "i", Desc: "input"}}, base...)
	case FocusInput:
		return append([]footer.Hint{{Key: "esc", Desc: "exit input"}}, base...)
	}
	return base
}
