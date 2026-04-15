package app

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/demon/daemon-client/internal/component/dialog"
	"github.com/demon/daemon-client/internal/component/footer"
	"github.com/demon/daemon-client/internal/component/header"
	"github.com/demon/daemon-client/internal/component/help"
	"github.com/demon/daemon-client/internal/component/input"
	"github.com/demon/daemon-client/internal/component/palette"
	"github.com/demon/daemon-client/internal/component/settings"
	"github.com/demon/daemon-client/internal/component/sidebar"
	"github.com/demon/daemon-client/internal/component/splash"
	"github.com/demon/daemon-client/internal/component/toast"
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
	screenSettings
)

// Options controls runtime features for the app. Populated from CLI flags in
// cmd/daemonctl/main.go.
type Options struct {
	DevMode bool
	Theme   string // name lookup in theme.Registry; falls back to default
	Mouse   bool   // enable mouse capture (MouseModeCellMotion) in tea.View
}

type Model struct {
	screen        screen
	width, height int

	theme  *theme.Theme
	styles *theme.Styles

	store    *session.Store
	engine   *mock.Engine
	selected int
	current  session.ID

	focus FocusID

	inputBuf string
	scrollTr int
	autoFol  bool

	sidebarCollapsed bool

	lastCtrlC time.Time

	// overlays
	helpOpen    bool
	paletteOpen bool
	palette     *palette.Model
	dialog      *dialog.Model
	toasts      toast.Stack

	devMode bool
	mouseOn bool

	// where to restore focus after closing a modal
	focusBeforeModal FocusID

	// settings screen state
	settingsSel int

	// where to return when closing a full-screen route (settings → main)
	prevScreen screen
}

// New constructs the root Model. Pass Options{} for defaults.
func New(store *session.Store, eng *mock.Engine, opts Options) *Model {
	t := theme.ByName(opts.Theme)
	m := &Model{
		screen:  screenSplash,
		theme:   t,
		styles:  theme.BuildStyles(t),
		store:   store,
		engine:  eng,
		focus:   FocusSidebar,
		autoFol: true,
		devMode: opts.DevMode,
		mouseOn: opts.Mouse,
	}
	m.palette = palette.New(defaultActions())
	m.seedToasts()
	return m
}

func (m *Model) seedToasts() {
	m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Connected", Body: "wss://prod.example.com"})
	m.toasts.Push(toast.Toast{Kind: toast.Success, Title: "Session resumed", Body: "refactor auth module"})
	m.toasts.Push(toast.Toast{Kind: toast.Warning, Title: "Awaiting input", Body: "add vitest coverage"})
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
		m.toasts.Tick(time.Now())
		return m, clockTick()

	case splash.DoneMsg:
		if m.screen == screenSplash {
			m.screen = screenMain
			m.selectInitial()
		}
		return m, nil

	case session.AppendMsg:
		if sess := m.store.Get(msg.ID); sess != nil {
			sess.Transcript += msg.Content
		}
		return m, nil

	case session.StatusMsg:
		m.store.SetStatus(msg.ID, msg.Status)
		return m, nil

	case SetThemeMsg:
		m.setTheme(theme.ByName(msg.Name))
		return m, nil

	case CycleThemeMsg:
		m.setTheme(theme.Next(m.theme.Name))
		m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Theme", Body: m.theme.Label})
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) setTheme(t *theme.Theme) {
	m.theme = t
	m.styles = theme.BuildStyles(t)
}

func (m *Model) handleKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := k.String()

	if m.screen == screenSplash {
		m.screen = screenMain
		m.selectInitial()
		return m, nil
	}

	// Modal handling first — they trap all keys.
	if m.dialog != nil {
		return m.handleDialogKey(key)
	}
	if m.paletteOpen {
		return m.handlePaletteKey(k, key)
	}
	if m.helpOpen {
		if key == "?" || key == "esc" {
			m.helpOpen = false
		}
		return m, nil
	}

	// Settings screen intercepts its own keys before the usual focus routing.
	if m.screen == screenSettings {
		return m.handleSettingsKey(key)
	}

	// Global, regardless of focus.
	switch key {
	case "ctrl+c":
		if time.Since(m.lastCtrlC) < 500*time.Millisecond {
			return m, tea.Quit
		}
		m.lastCtrlC = time.Now()
		return m, nil
	case "?":
		// `?` is only global when *not* editing input.
		if m.focus != FocusInput {
			m.openHelp()
			return m, nil
		}
	case "ctrl+p":
		m.openPalette()
		return m, nil
	case "ctrl+,":
		m.openSettings()
		return m, nil
	case "ctrl+t":
		return m, func() tea.Msg { return CycleThemeMsg{} }
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

	// Dev cheats.
	if m.devMode {
		if cmd, ok := m.handleDevCheat(key); ok {
			return m, cmd
		}
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

// --- overlays ---

func (m *Model) openHelp() {
	m.helpOpen = true
}

func (m *Model) openPalette() {
	m.paletteOpen = true
	m.focusBeforeModal = m.focus
	m.palette.Reset()
}

func (m *Model) closePalette() {
	m.paletteOpen = false
	if m.focusBeforeModal != FocusNone {
		m.focus = m.focusBeforeModal
	}
}

func (m *Model) openDialog(d *dialog.Model) {
	m.dialog = d
	m.focusBeforeModal = m.focus
}

func (m *Model) closeDialog() {
	m.dialog = nil
	if m.focusBeforeModal != FocusNone {
		m.focus = m.focusBeforeModal
	}
}

func (m *Model) handlePaletteKey(k tea.KeyPressMsg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+p":
		m.closePalette()
		return m, nil
	case "up":
		m.palette.Move(-1)
		return m, nil
	case "down":
		m.palette.Move(1)
		return m, nil
	case "enter":
		if a, ok := m.palette.Selected(); ok {
			m.closePalette()
			return m.runAction(a.ID)
		}
		return m, nil
	case "backspace":
		m.palette.Backspace()
		return m, nil
	}
	if k.Text != "" {
		for _, r := range k.Text {
			m.palette.AppendRune(r)
		}
	}
	return m, nil
}

func (m *Model) handleDialogKey(key string) (tea.Model, tea.Cmd) {
	d := m.dialog
	if d == nil {
		return m, nil
	}
	decision, done := d.HandleKey(key)
	if !done {
		return m, nil
	}
	kind := d.Kind
	// Close before running side effects so toasts render over main view.
	m.closeDialog()
	if decision == dialog.ResultCancelled {
		m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Dismissed", Body: "(cancelled)"})
		return m, nil
	}
	switch kind {
	case dialog.KindPermission:
		labels := []string{"Allow once", "Allow always", "Deny"}
		m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Permission", Body: labels[decision]})
	case dialog.KindQuestion:
		m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Answered", Body: fmt.Sprintf("option %d", decision+1)})
	case dialog.KindConfirm:
		if decision == 1 {
			m.toasts.Push(toast.Toast{Kind: toast.Success, Title: "Confirmed"})
		}
	case dialog.KindAlert:
		// nothing
	}
	return m, nil
}

// --- actions / dev cheats ---

func defaultActions() []palette.Action {
	return []palette.Action{
		{ID: "new-session", Title: "New session", Icon: "＋", Shortcut: "⌃n", Category: "session"},
		{ID: "stop-session", Title: "Stop current session", Icon: "■", Shortcut: "s", Category: "session"},
		{ID: "resume-session", Title: "Resume current session", Icon: "▶", Shortcut: "r", Category: "session"},
		{ID: "kill-session", Title: "Kill current session", Icon: "✗", Shortcut: "x", Category: "session"},
		{ID: "archive-session", Title: "Archive current session", Icon: "📦", Shortcut: "d", Category: "session"},
		{ID: "toggle-sidebar", Title: "Toggle sidebar", Icon: "◧", Shortcut: "⌃b", Category: "view"},
		{ID: "toggle-inspector", Title: "Toggle inspector", Icon: "◨", Shortcut: "I", Category: "view"},
		{ID: "toggle-thinking", Title: "Toggle thinking blocks", Icon: "💭", Shortcut: "⌃x t", Category: "view"},
		{ID: "toggle-compact", Title: "Toggle compact density", Icon: "⇲", Shortcut: "⌃d", Category: "view"},
		{ID: "switch-theme", Title: "Switch theme", Icon: "🎨", Shortcut: "⌃t", Category: "appearance"},
		{ID: "focus-sidebar", Title: "Focus sidebar", Icon: "→", Category: "focus"},
		{ID: "focus-transcript", Title: "Focus transcript", Icon: "→", Category: "focus"},
		{ID: "focus-input", Title: "Focus input", Icon: "→", Shortcut: "i", Category: "focus"},
		{ID: "open-settings", Title: "Open settings", Icon: "⚙", Shortcut: "⌃,", Category: "app"},
		{ID: "open-help", Title: "Open help", Icon: "?", Shortcut: "?", Category: "app"},
		{ID: "open-logs", Title: "Open logs", Icon: "📜", Category: "app"},
		{ID: "search-transcript", Title: "Search transcript", Icon: "🔍", Shortcut: "/", Category: "transcript"},
		{ID: "copy-session-id", Title: "Copy session ID", Icon: "⎘", Category: "clipboard"},
		{ID: "copy-last-output", Title: "Copy last output", Icon: "⎘", Category: "clipboard"},
		{ID: "quit", Title: "Quit", Icon: "⏻", Shortcut: "⌃c⌃c", Category: "app"},
	}
}

func (m *Model) runAction(id string) (tea.Model, tea.Cmd) {
	switch id {
	case "toggle-sidebar":
		m.sidebarCollapsed = !m.sidebarCollapsed
	case "focus-sidebar":
		m.focus = FocusSidebar
	case "focus-transcript":
		m.focus = FocusTranscript
	case "focus-input":
		m.focus = FocusInput
	case "open-help":
		m.openHelp()
	case "quit":
		return m, tea.Quit
	case "switch-theme":
		return m, func() tea.Msg { return CycleThemeMsg{} }
	case "open-settings":
		m.openSettings()
	default:
		m.toasts.Push(toast.Toast{Kind: toast.Info, Title: "Action", Body: id + " — phase 2"})
	}
	return m, nil
}

func (m *Model) handleDevCheat(key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+alt+p":
		m.openDialog(dialog.Permission(
			"bash",
			"rm -rf /tmp/staging",
			"- /tmp/staging/cache.db\n- /tmp/staging/logs/app.log",
		))
		return nil, true
	case "ctrl+alt+q":
		m.openDialog(dialog.Question(
			"Agent is asking",
			"Which test runner should I use?",
			[]string{"vitest", "jest", "bun test"},
		))
		return nil, true
	case "ctrl+alt+t":
		m.injectRandomToast()
		return nil, true
	case "ctrl+alt+c":
		if sess := m.currentSession(); sess != nil {
			m.store.SetStatus(sess.ID, session.StatusCompleted)
			m.toasts.Push(toast.Toast{Kind: toast.Success, Title: "Completed", Body: sess.Title})
		}
		return nil, true
	case "ctrl+alt+f":
		if sess := m.currentSession(); sess != nil {
			m.store.SetStatus(sess.ID, session.StatusFailed)
			m.toasts.Push(toast.Toast{Kind: toast.Error, Title: "Failed", Body: sess.Title})
		}
		return nil, true
	}
	return nil, false
}

func (m *Model) currentSession() *session.Session {
	if m.current == "" {
		return nil
	}
	return m.store.Get(m.current)
}

func (m *Model) injectRandomToast() {
	samples := []toast.Toast{
		{Kind: toast.Info, Title: "Heads up", Body: "a new mock event"},
		{Kind: toast.Success, Title: "Session done", Body: "lint-fix finished ok"},
		{Kind: toast.Warning, Title: "Rate limit", Body: "slowing replays"},
		{Kind: toast.Error, Title: "Transport", Body: "ws flapped, reconnecting…"},
	}
	m.toasts.Push(samples[rand.Intn(len(samples))])
}

// --- settings screen ---

func (m *Model) openSettings() {
	if m.screen == screenSettings {
		return
	}
	m.prevScreen = m.screen
	m.screen = screenSettings
}

func (m *Model) closeSettings() {
	if m.prevScreen == 0 {
		m.screen = screenMain
	} else {
		m.screen = m.prevScreen
	}
}

func (m *Model) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+,":
		m.closeSettings()
		return m, nil
	case "ctrl+c":
		if time.Since(m.lastCtrlC) < 500*time.Millisecond {
			return m, tea.Quit
		}
		m.lastCtrlC = time.Now()
		return m, nil
	case "ctrl+s":
		m.toasts.Push(toast.Toast{Kind: toast.Success, Title: "Saved", Body: "(phase 2 writes to config.toml)"})
		return m, nil
	case "ctrl+t":
		return m, func() tea.Msg { return CycleThemeMsg{} }
	case "ctrl+p":
		m.openPalette()
		return m, nil
	case "j", "down":
		if m.settingsSel < len(settings.Categories)-1 {
			m.settingsSel++
		}
	case "k", "up":
		if m.settingsSel > 0 {
			m.settingsSel--
		}
	case "g":
		m.settingsSel = 0
	case "G":
		m.settingsSel = len(settings.Categories) - 1
	case "?":
		m.openHelp()
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
	if m.mouseOn {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

func (m *Model) render() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.screen == screenSplash {
		return splash.Render(m.theme, m.styles, m.width, m.height)
	}
	var base string
	if m.screen == screenSettings {
		base = m.renderSettings()
	} else {
		base = m.renderMain()
	}

	// Composite overlays: toasts (bottom-right), then modal (centered).
	layers := []*lipgloss.Layer{lipgloss.NewLayer(base).Z(0)}

	if toasts := m.toasts.Visible(); len(toasts) > 0 {
		block := toast.Render(m.theme, toasts, m.width)
		tw := lipgloss.Width(block)
		th := lipgloss.Height(block)
		x := m.width - tw - 2
		y := m.height - th - 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		layers = append(layers, lipgloss.NewLayer(block).X(x).Y(y).Z(1))
	}

	if m.helpOpen {
		box := help.Render(m.theme, m.width, m.height)
		layers = append(layers, centered(box, m.width, m.height, 10))
	} else if m.paletteOpen {
		box := palette.Render(m.theme, m.palette, m.width, m.height)
		layers = append(layers, centered(box, m.width, m.height, 10))
	} else if m.dialog != nil {
		box := dialog.Render(m.theme, m.dialog, m.width, m.height)
		layers = append(layers, centered(box, m.width, m.height, 10))
	}

	if len(layers) == 1 {
		return base
	}
	return lipgloss.NewCompositor(layers...).Render()
}

func centered(content string, w, h, z int) *lipgloss.Layer {
	cw := lipgloss.Width(content)
	ch := lipgloss.Height(content)
	x := (w - cw) / 2
	y := (h - ch) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return lipgloss.NewLayer(content).X(x).Y(y).Z(z)
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

	mainCol := lipgloss.JoinVertical(lipgloss.Left, sessHdr, tr, inp)
	var body string
	if side == "" {
		body = mainCol
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, side, mainCol)
	}
	return lipgloss.JoinVertical(lipgloss.Left, hdr, body, ftr)
}

func (m *Model) renderSettings() string {
	st := settings.State{
		Selected:  m.settingsSel,
		Theme:     m.theme,
		ServerURL: "wss://prod.example.com",
		Workdir:   "/Users/demon/projects/rag-broker",
		MaxSess:   4,
		Agent:     "codex",
		Model:     "gpt-5-sonnet",
		LogLevel:  "info",
		LogFile:   "~/.local/state/daemonctl/daemonctl.log",
		Version:   "v0.1.0",
		Build:     "showcase",
	}
	body := settings.Render(m.theme, m.styles, st, m.width, m.height-1)
	ftr := footer.Render(m.theme, m.styles, []footer.Hint{
		{Key: "j/k", Desc: "move"},
		{Key: "⌃s", Desc: "save"},
		{Key: "⌃t", Desc: "cycle theme"},
		{Key: "esc", Desc: "back"},
		{Key: "?", Desc: "help"},
	}, m.width)
	return lipgloss.JoinVertical(lipgloss.Left, body, ftr)
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
		{Key: "⌃p", Desc: "palette"},
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
