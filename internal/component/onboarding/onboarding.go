// Package onboarding implements the 11-step first-run wizard (spec §7.2).
//
// The spec nominates `charmbracelet/huh` as the form library; phase 1 ships a
// native-bubbletea implementation instead, so we stay dependency-frugal and
// consistent with the other custom components in this repo (dialog, palette,
// settings). Migrating to huh is a phase-2 follow-up that doesn't change the
// wizard's public surface (Model + key/update/view + DoneMsg).
//
// Shape:
//
//   - Steps 0-10 are declared in steps.go as a fixed slice.
//   - The wizard owns a *config.Config pointer it mutates on each field commit.
//   - On step 10 (Summary), pressing Finish writes the config to disk via
//     config.Save and emits DoneMsg.
//   - Esc / left-arrow on the first element of a step rewinds one step.
//     Esc on step 0 emits QuitConfirmMsg.
package onboarding

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/demon/daemon-client/internal/config"
	"github.com/demon/daemon-client/internal/theme"
)

// DoneMsg is emitted when the user finishes the wizard successfully. The
// config has already been written to disk.
type DoneMsg struct{ Cfg config.Config }

// QuitConfirmMsg is emitted when the user presses Esc on step 0.
// The root app decides whether to actually quit or confirm.
type QuitConfirmMsg struct{}

// Model is the wizard state. Construct with New().
type Model struct {
	Cfg  config.Config
	step int

	// Per-step cursor within fields/buttons.
	cursor int

	// Scratch buffer for text input steps.
	buf string

	// For multi-select (step 8) and boolean toggles.
	// index → on/off, indexed into the step's option list.
	multi map[int]bool

	// Error message for the current step (e.g. URL validation).
	err string

	// Terminal size so we can render centered.
	width, height int
}

// New builds a wizard populated with sensible defaults.
func New() *Model {
	c := config.Default()
	c.UsedDefaults = false
	// Step 6 default: min(4, numCPU/2), clamped to >=1.
	if n := runtime.NumCPU() / 2; n > 0 && n < 4 {
		c.MaxSessions = n
	}
	m := &Model{
		Cfg:   c,
		multi: map[int]bool{},
	}
	m.loadStep()
	return m
}

// SetSize is called by the root on WindowSizeMsg.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// Step returns the current 0-indexed step number.
func (m *Model) Step() int { return m.step }

// Total returns the total number of steps.
func (m *Model) Total() int { return len(steps) }

// --- Update ---

// Update handles a single key press and returns an optional outbound command
// (DoneMsg / QuitConfirmMsg). Returns nil if no command.
func (m *Model) Update(key string, text string) tea.Cmd {
	m.err = ""
	s := steps[m.step]
	switch s.Kind {
	case stepNote:
		return m.handleNote(key)
	case stepInput:
		return m.handleInput(key, text)
	case stepPassword:
		return m.handleInput(key, text)
	case stepFilepath:
		return m.handleInput(key, text)
	case stepSelect:
		return m.handleSelect(key, s.Options)
	case stepMultiSelect:
		return m.handleMulti(key, s.Options)
	case stepSegmented:
		return m.handleSegmented(key, s)
	case stepAgents:
		return m.handleAgents(key)
	case stepAdvanced:
		return m.handleAdvanced(key)
	case stepSummary:
		return m.handleSummary(key)
	}
	return nil
}

func (m *Model) handleNote(key string) tea.Cmd {
	switch key {
	case "enter", " ", "right", "l":
		m.advance()
	case "esc", "left":
		return func() tea.Msg { return QuitConfirmMsg{} }
	}
	return nil
}

func (m *Model) handleInput(key, text string) tea.Cmd {
	switch key {
	case "enter":
		return m.commitInput()
	case "esc":
		return m.back()
	case "left":
		// Only rewind if the buffer is empty — otherwise let it pass so users
		// can edit (phase 2 adds cursor movement within the buffer).
		if m.buf == "" {
			return m.back()
		}
	case "backspace":
		if len(m.buf) > 0 {
			r := []rune(m.buf)
			m.buf = string(r[:len(r)-1])
		}
	case "ctrl+u":
		m.buf = ""
	default:
		if text != "" {
			m.buf += text
		}
	}
	return nil
}

func (m *Model) commitInput() tea.Cmd {
	s := steps[m.step]
	val := strings.TrimSpace(m.buf)
	if val == "" {
		val = s.Default
	}
	switch s.ID {
	case "server-url":
		if err := validateWSURL(val); err != nil {
			m.err = err.Error()
			return nil
		}
		m.Cfg.ServerURL = val
	case "auth-token":
		m.Cfg.AuthToken = val
	case "workdir":
		if val == "" {
			m.err = "workdir cannot be empty"
			return nil
		}
		m.Cfg.Workdir = val
	}
	m.advance()
	return nil
}

func (m *Model) handleSelect(key string, opts []string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
	case "enter", " ":
		m.commitSelect(opts[m.cursor])
		m.advance()
	case "esc":
		return m.back()
	case "left":
		return m.back()
	case "right":
		m.commitSelect(opts[m.cursor])
		m.advance()
	}
	return nil
}

func (m *Model) commitSelect(val string) {
	switch steps[m.step].ID {
	case "max-sessions":
		// Value is the string form, e.g. "4".
		n := 4
		fmt.Sscanf(val, "%d", &n)
		m.Cfg.MaxSessions = n
	case "default-agent":
		m.Cfg.Agent = val
		// Reset model to first legal option for this agent.
		m.Cfg.Model = modelsFor(val)[0]
	case "default-model":
		m.Cfg.Model = val
	}
}

func (m *Model) handleMulti(key string, opts []string) tea.Cmd {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
	case " ":
		m.multi[m.cursor] = !m.multi[m.cursor]
	case "enter":
		m.commitMulti()
		m.advance()
	case "esc", "left":
		return m.back()
	}
	return nil
}

func (m *Model) commitMulti() {
	m.Cfg.Notifications = config.Notifications{
		BellOnAttention:    m.multi[0],
		DesktopOnAttention: m.multi[1],
		DesktopOnComplete:  m.multi[2],
		DesktopOnFail:      m.multi[3],
	}
}

// handleSegmented covers any step with two chip rows and horizontal focus
// within a row plus Tab to move between rows. Used by Appearance
// (theme + density) and Default agent + model (dependent selects).
func (m *Model) handleSegmented(key string, s step) tea.Cmd {
	// cursor encodes (row<<8 | col). Row 0 = top group, row 1 = bottom.
	row := m.cursor >> 8
	col := m.cursor & 0xff
	groups := segmentedGroups(m, s)
	switch key {
	case "left":
		if col > 0 {
			col--
		} else {
			return m.back()
		}
	case "right":
		if col < len(groups[row])-1 {
			col++
		}
	case "tab", "down", "j":
		if row < len(groups)-1 {
			row++
			if col >= len(groups[row]) {
				col = len(groups[row]) - 1
			}
		}
	case "shift+tab", "up", "k":
		if row > 0 {
			row--
			if col >= len(groups[row]) {
				col = len(groups[row]) - 1
			}
		}
	case "enter", " ":
		m.commitSegmented(s, groups)
		m.advance()
		return nil
	case "esc":
		return m.back()
	}
	m.cursor = (row << 8) | col
	// Eagerly commit so downstream steps & the summary reflect live choices.
	// When the top row changes on the agent+model step, clamp the model col.
	if s.ID == "default-agent-model" {
		// After moving on row 0, rebuild groups to get fresh model options.
		m.Cfg.Agent = groups[0][m.segCol(0)]
		groups = segmentedGroups(m, s)
		if m.cursor>>8 == 1 {
			c := m.cursor & 0xff
			if c >= len(groups[1]) {
				c = 0
				m.cursor = (1 << 8) | c
			}
		}
	}
	m.commitSegmented(s, groups)
	return nil
}

// segmentedGroups resolves the two row options for a segmented step. The
// Default agent+model step rebuilds row 2 based on the selected agent.
func segmentedGroups(m *Model, s step) [][]string {
	if s.ID == "default-agent-model" {
		return [][]string{s.Options, modelsFor(m.Cfg.Agent)}
	}
	return [][]string{s.Options, s.Options2}
}

// commitSegmented writes the segmented selections into Cfg.
func (m *Model) commitSegmented(s step, groups [][]string) {
	a := groups[0][m.segCol(0)]
	b := groups[1][m.segCol(1)]
	switch s.ID {
	case "appearance":
		m.Cfg.Theme = a
		m.Cfg.Density = b
	case "default-agent-model":
		m.Cfg.Agent = a
		m.Cfg.Model = b
	}
}

// segCol extracts the column on a given row from the packed cursor, defaulting
// to whatever matches the current config value if the user hasn't visited the
// row yet.
func (m *Model) segCol(row int) int {
	s := steps[m.step]
	groups := segmentedGroups(m, s)
	if m.cursor>>8 == row {
		c := m.cursor & 0xff
		if c >= len(groups[row]) {
			return 0
		}
		return c
	}
	// Recover from current config.
	var target string
	switch s.ID {
	case "appearance":
		if row == 0 {
			target = m.Cfg.Theme
		} else {
			target = m.Cfg.Density
		}
	case "default-agent-model":
		if row == 0 {
			target = m.Cfg.Agent
		} else {
			target = m.Cfg.Model
		}
	}
	for i, opt := range groups[row] {
		if opt == target {
			return i
		}
	}
	return 0
}

// handleAgents covers step 5: we pretend both binaries were auto-detected and
// the user just confirms. No branching in phase 1.
func (m *Model) handleAgents(key string) tea.Cmd {
	switch key {
	case "enter", " ", "right", "l":
		m.advance()
	case "esc", "left":
		return m.back()
	}
	return nil
}

// handleAdvanced: step 10 is a compact form with log level select + telemetry
// toggle. cursor maps 0..n-1 over log levels, then len is telemetry toggle.
func (m *Model) handleAdvanced(key string) tea.Cmd {
	levels := []string{"off", "error", "info", "debug"}
	total := len(levels) + 1 // +1 for telemetry row
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < total-1 {
			m.cursor++
		}
	case " ":
		if m.cursor < len(levels) {
			m.Cfg.LogLevel = levels[m.cursor]
		} else {
			m.Cfg.Telemetry = !m.Cfg.Telemetry
		}
	case "enter":
		if m.cursor < len(levels) {
			m.Cfg.LogLevel = levels[m.cursor]
		}
		m.advance()
	case "esc", "left":
		return m.back()
	}
	return nil
}

// handleSummary covers step 11: three buttons [ ← Back ] [ Edit ] [ ✓ Finish ].
// cursor is the button index 0..2.
func (m *Model) handleSummary(key string) tea.Cmd {
	switch key {
	case "left":
		if m.cursor > 0 {
			m.cursor--
		} else {
			return m.back()
		}
	case "right":
		if m.cursor < 2 {
			m.cursor++
		}
	case "tab":
		m.cursor = (m.cursor + 1) % 3
	case "shift+tab":
		m.cursor = (m.cursor + 2) % 3
	case "esc":
		return m.back()
	case "enter", " ":
		switch m.cursor {
		case 0: // Back
			return m.back()
		case 1: // Edit — jump to step 1 (server URL), the first editable field.
			m.step = 1
			m.loadStep()
			return nil
		case 2: // Finish
			if err := config.Save(m.Cfg); err != nil {
				m.err = "failed to save config: " + err.Error()
				return nil
			}
			out := m.Cfg
			return func() tea.Msg { return DoneMsg{Cfg: out} }
		}
	}
	return nil
}

// --- Navigation helpers ---

func (m *Model) advance() {
	if m.step >= len(steps)-1 {
		return
	}
	m.step++
	m.loadStep()
}

func (m *Model) back() tea.Cmd {
	if m.step == 0 {
		return func() tea.Msg { return QuitConfirmMsg{} }
	}
	m.step--
	m.loadStep()
	return nil
}

// loadStep resets cursor/buf for the current step, seeded from the in-progress
// config so going Back preserves earlier answers.
func (m *Model) loadStep() {
	s := steps[m.step]
	m.cursor = 0
	m.buf = ""
	switch s.Kind {
	case stepInput, stepPassword, stepFilepath:
		switch s.ID {
		case "server-url":
			m.buf = m.Cfg.ServerURL
		case "auth-token":
			m.buf = m.Cfg.AuthToken
		case "workdir":
			m.buf = m.Cfg.Workdir
		}
	case stepSelect:
		switch s.ID {
		case "max-sessions":
			for i, opt := range s.Options {
				if opt == fmt.Sprintf("%d", m.Cfg.MaxSessions) {
					m.cursor = i
					break
				}
			}
		}
	case stepSegmented:
		// Focus row 0 by default; segCol resolves from config.
		m.cursor = 0
	case stepMultiSelect:
		m.multi[0] = m.Cfg.Notifications.BellOnAttention
		m.multi[1] = m.Cfg.Notifications.DesktopOnAttention
		m.multi[2] = m.Cfg.Notifications.DesktopOnComplete
		m.multi[3] = m.Cfg.Notifications.DesktopOnFail
	case stepAdvanced:
		levels := []string{"off", "error", "info", "debug"}
		for i, lv := range levels {
			if lv == m.Cfg.LogLevel {
				m.cursor = i
				break
			}
		}
	case stepSummary:
		m.cursor = 2 // focus Finish by default
	}
}

// modelsFor returns the model options for a given agent (phase 1 hardcoded;
// phase 2 fetches from server).
func modelsFor(agent string) []string {
	switch agent {
	case "codex":
		return []string{"gpt-5-sonnet", "gpt-5-mini", "gpt-5-turbo"}
	case "opencode":
		return []string{"gpt-5-sonnet", "claude-sonnet-4-6", "claude-haiku-4-5"}
	}
	return []string{"gpt-5-sonnet"}
}

func validateWSURL(s string) error {
	if s == "" {
		return fmt.Errorf("URL is required")
	}
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("scheme must be ws:// or wss://")
	}
	if u.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}

// --- View ---

// Render returns the full-screen wizard view.
func (m *Model) Render(t *theme.Theme, st *theme.Styles) string {
	if m.width < 60 || m.height < 16 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			st.TooSmall.Render("Too small — please resize (min 60×16)"))
	}
	s := steps[m.step]

	indicator := m.renderIndicator(t, st)
	title := lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(s.Title)
	subtitle := ""
	if s.Subtitle != "" {
		subtitle = st.Dim.Render(s.Subtitle)
	}

	var body string
	switch s.Kind {
	case stepNote:
		body = m.renderNote(t, st, s)
	case stepInput:
		body = m.renderInput(t, st, s, false)
	case stepPassword:
		body = m.renderInput(t, st, s, true)
	case stepFilepath:
		body = m.renderInput(t, st, s, false)
	case stepSelect:
		body = m.renderSelect(t, st, s)
	case stepMultiSelect:
		body = m.renderMulti(t, st, s)
	case stepSegmented:
		body = m.renderSegmented(t, st, s)
	case stepAgents:
		body = m.renderAgents(t, st)
	case stepAdvanced:
		body = m.renderAdvanced(t, st)
	case stepSummary:
		body = m.renderSummary(t, st)
	}

	footer := m.renderHints(st)

	parts := []string{indicator, "", title}
	if subtitle != "" {
		parts = append(parts, subtitle)
	}
	parts = append(parts, "", body)
	if m.err != "" {
		parts = append(parts,
			lipgloss.NewStyle().Foreground(t.Danger).Render("  ! "+m.err))
	}
	block := lipgloss.JoinVertical(lipgloss.Left, parts...)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 3).
		Width(minInt(90, m.width-8)).
		Render(block)

	outer := lipgloss.JoinVertical(lipgloss.Center, box, "", footer)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, outer)
}

func (m *Model) renderIndicator(t *theme.Theme, st *theme.Styles) string {
	var b strings.Builder
	for i := 0; i < len(steps); i++ {
		glyph := "○"
		var s lipgloss.Style
		switch {
		case i == m.step:
			glyph = "●"
			s = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
		case i < m.step:
			glyph = "●"
			s = lipgloss.NewStyle().Foreground(t.Accent)
		default:
			s = st.Dim
		}
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(s.Render(glyph))
	}
	label := st.Dim.Render(fmt.Sprintf("   (%d/%d — %s)",
		m.step+1, len(steps), steps[m.step].Title))
	return b.String() + label
}

func (m *Model) renderNote(t *theme.Theme, st *theme.Styles, s step) string {
	body := s.Body
	hint := st.Dim.Render("↵ continue    esc quit")
	if m.step == 0 {
		logo := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(miniLogo)
		return lipgloss.JoinVertical(lipgloss.Left, logo, "", body, "", hint)
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, "", hint)
}

func (m *Model) renderInput(t *theme.Theme, st *theme.Styles, s step, password bool) string {
	shown := m.buf
	if password {
		shown = strings.Repeat("•", len([]rune(m.buf)))
	}
	if shown == "" {
		shown = st.InputPlaceholder.Render(s.Placeholder)
	}
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Accent).
		Padding(0, 1).
		Width(60).
		Render(shown + lipgloss.NewStyle().Foreground(t.Accent).Render("█"))

	var hints []string
	if s.Hint != "" {
		hints = append(hints, st.Dim.Render(s.Hint))
	}
	hints = append(hints, st.Dim.Render("↵ continue    esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, append([]string{inputBox, ""}, hints...)...)
}

func (m *Model) renderSelect(t *theme.Theme, st *theme.Styles, s step) string {
	opts := s.Options
	if s.ID == "default-model" {
		opts = modelsFor(m.Cfg.Agent)
		if m.cursor >= len(opts) {
			m.cursor = 0
		}
	}
	var rows []string
	for i, opt := range opts {
		label := opt
		if s.ID == "max-sessions" {
			label = opt + " concurrent sessions"
		}
		if i == m.cursor {
			bar := lipgloss.NewStyle().Foreground(t.Accent).Render("▌ ")
			rows = append(rows, bar+lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(label))
		} else {
			rows = append(rows, "  "+st.Dim.Render(label))
		}
	}
	rows = append(rows, "", st.Dim.Render("↑/↓ move   ↵ continue   esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderMulti(t *theme.Theme, st *theme.Styles, s step) string {
	var rows []string
	for i, opt := range s.Options {
		box := "[ ]"
		if m.multi[i] {
			box = "[x]"
		}
		line := fmt.Sprintf(" %s  %s", box, opt)
		if i == m.cursor {
			bar := lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
			rows = append(rows, bar+lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(line))
		} else {
			rows = append(rows, " "+st.Dim.Render(line))
		}
	}
	rows = append(rows, "", st.Dim.Render("space toggle   ↵ continue   esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderSegmented(t *theme.Theme, st *theme.Styles, s step) string {
	groups := segmentedGroups(m, s)
	labels := []string{"Theme", "Density"}
	if s.ID == "default-agent-model" {
		labels = []string{"Agent", "Model"}
	}
	curRow := m.cursor >> 8
	curCol := m.cursor & 0xff
	var rows []string
	for row, opts := range groups {
		var chips []string
		for col, opt := range opts {
			chip := " " + opt + " "
			if row == curRow && col == curCol {
				chip = lipgloss.NewStyle().Background(t.Accent).Foreground(t.Bg).Bold(true).Render(" " + opt + " ")
			} else {
				// Mark the currently-committed choice on non-focused rows so the
				// user can see what's selected without walking the whole group.
				committed := groups[row][m.segCol(row)]
				if opt == committed {
					chip = lipgloss.NewStyle().Foreground(t.Accent).Render("[" + opt + "]")
				} else {
					chip = st.Dim.Render(" " + opt + " ")
				}
			}
			chips = append(chips, chip)
		}
		head := st.Dim.Render(pad(labels[row]+":", 10))
		rows = append(rows, head+strings.Join(chips, "  "))
	}
	rows = append(rows, "", st.Dim.Render("←/→ within row   tab next row   ↵ continue   esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderAgents(t *theme.Theme, st *theme.Styles) string {
	ok := lipgloss.NewStyle().Foreground(t.Success).Bold(true).Render("✓")
	rows := []string{
		fmt.Sprintf("  codex     %s  %s", ok, m.Cfg.AgentCodex),
		fmt.Sprintf("  opencode  %s  %s", ok, m.Cfg.AgentOC),
		"",
		st.Dim.Render("(auto-detected on $PATH; phase 1 always succeeds)"),
		"",
		st.Dim.Render("↵ continue   esc back"),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderAdvanced(t *theme.Theme, st *theme.Styles) string {
	levels := []string{"off", "error", "info", "debug"}
	var rows []string
	rows = append(rows, st.Dim.Render("  Log level"))
	for i, lv := range levels {
		line := "   " + lv
		if lv == m.Cfg.LogLevel {
			line = "  ● " + lv
		} else {
			line = "  ○ " + lv
		}
		if i == m.cursor {
			bar := lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
			rows = append(rows, bar+lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(line))
		} else {
			rows = append(rows, " "+st.Dim.Render(line))
		}
	}
	rows = append(rows, "")
	rows = append(rows, st.Dim.Render("  Telemetry"))
	box := "[ ]"
	if m.Cfg.Telemetry {
		box = "[x]"
	}
	telemetryRow := fmt.Sprintf("  %s  send anonymous usage stats", box)
	if m.cursor == len(levels) {
		bar := lipgloss.NewStyle().Foreground(t.Accent).Render("▌")
		rows = append(rows, bar+lipgloss.NewStyle().Foreground(t.Fg).Bold(true).Render(telemetryRow))
	} else {
		rows = append(rows, " "+st.Dim.Render(telemetryRow))
	}
	rows = append(rows, "", st.Dim.Render("↑/↓ move   space toggle   ↵ continue   esc back"))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m *Model) renderSummary(t *theme.Theme, st *theme.Styles) string {
	// Syntax-ish highlighted TOML (poor man's: keys in dim, strings in info).
	cfg := m.Cfg
	lines := []string{
		kvLine(st, t, "server_url", fmt.Sprintf("%q", cfg.ServerURL)),
		kvLine(st, t, "auth_token", fmt.Sprintf("%q", maskToken(cfg.AuthToken))),
		kvLine(st, t, "workdir", fmt.Sprintf("%q", cfg.Workdir)),
		kvLine(st, t, "max_sessions", fmt.Sprintf("%d", cfg.MaxSessions)),
		kvLine(st, t, "agent", fmt.Sprintf("%q", cfg.Agent)),
		kvLine(st, t, "model", fmt.Sprintf("%q", cfg.Model)),
		kvLine(st, t, "theme", fmt.Sprintf("%q", cfg.Theme)),
		kvLine(st, t, "density", fmt.Sprintf("%q", cfg.Density)),
		kvLine(st, t, "log_level", fmt.Sprintf("%q", cfg.LogLevel)),
		kvLine(st, t, "telemetry", fmt.Sprintf("%t", cfg.Telemetry)),
		"",
		st.Dim.Render("[notifications]"),
		kvLine(st, t, "bell_on_attention", fmt.Sprintf("%t", cfg.Notifications.BellOnAttention)),
		kvLine(st, t, "desktop_on_attention", fmt.Sprintf("%t", cfg.Notifications.DesktopOnAttention)),
		kvLine(st, t, "desktop_on_complete", fmt.Sprintf("%t", cfg.Notifications.DesktopOnComplete)),
		kvLine(st, t, "desktop_on_fail", fmt.Sprintf("%t", cfg.Notifications.DesktopOnFail)),
	}
	body := strings.Join(lines, "\n")

	// Button row.
	btn := func(label string, focused bool, primary bool) string {
		if focused {
			s := lipgloss.NewStyle().Background(t.Accent).Foreground(t.Bg).Bold(true).Padding(0, 2)
			glyph := " "
			if primary {
				glyph = "✓ "
			}
			return s.Render(glyph + label)
		}
		return lipgloss.NewStyle().Foreground(t.Fg).Padding(0, 2).Render("[ " + label + " ]")
	}
	buttons := strings.Join([]string{
		btn("← Back", m.cursor == 0, false),
		btn("Edit", m.cursor == 1, false),
		btn("Finish", m.cursor == 2, true),
	}, "  ")

	return lipgloss.JoinVertical(lipgloss.Left,
		body, "", buttons, "",
		st.Dim.Render("←/→ move   ↵ activate   esc back"))
}

func kvLine(st *theme.Styles, t *theme.Theme, k, v string) string {
	key := st.Dim.Render(pad(k+" =", 22))
	val := lipgloss.NewStyle().Foreground(t.Info).Render(v)
	return "  " + key + " " + val
}

func (m *Model) renderHints(st *theme.Styles) string {
	return st.Dim.Render(fmt.Sprintf("step %d/%d    tab/shift+tab next/prev row    ? help",
		m.step+1, len(steps)))
}

func maskToken(t string) string {
	if t == "" {
		return ""
	}
	if len(t) <= 4 {
		return strings.Repeat("•", len(t))
	}
	return strings.Repeat("•", len(t)-4) + t[len(t)-4:]
}

func pad(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const miniLogo = `  ╭──╮  ╭──╮
  │  │  │  │  d a e m o n c t l
  ╰──╯  ╰──╯  the coding-agent fleet controller`
