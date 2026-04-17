package onboarding

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/demon/daemon-client/internal/config"
)

// press walks one key/text pair through the wizard and drains any outbound
// command to a message so tests can assert on DoneMsg / QuitConfirmMsg.
func press(m *Model, key, text string) tea.Msg {
	cmd := m.Update(key, text)
	if cmd == nil {
		return nil
	}
	return cmd()
}

// type_ replays plain text entry a rune at a time using key="<empty>" text="x".
func type_(m *Model, s string) {
	for _, r := range s {
		m.Update("", string(r))
	}
}

func TestStepCountIsEleven(t *testing.T) {
	if got := len(steps); got != 11 {
		t.Fatalf("expected 11 steps, got %d", got)
	}
}

func TestFullHappyPathWritesConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DAEMONCTL_CONFIG", "")

	m := New()

	// Step 0: Welcome → enter.
	press(m, "enter", "")
	if m.step != 1 {
		t.Fatalf("step after welcome = %d, want 1", m.step)
	}

	// Step 1: Server URL. Default buffer is the placeholder-ish default.
	// Clear it and type a valid URL.
	press(m, "ctrl+u", "")
	type_(m, "wss://demo.example.com/ws")
	press(m, "enter", "")
	if m.Cfg.ServerURL != "wss://demo.example.com/ws" {
		t.Fatalf("ServerURL = %q", m.Cfg.ServerURL)
	}

	// Step 2: Auth token.
	type_(m, "sk_test_1234567890")
	press(m, "enter", "")
	if m.Cfg.AuthToken != "sk_test_1234567890" {
		t.Fatalf("AuthToken mismatch: %q", m.Cfg.AuthToken)
	}

	// Step 3: Workdir — accept the default.
	press(m, "enter", "")

	// Step 4: Agent binaries confirmation.
	press(m, "enter", "")

	// Step 5: Max concurrent sessions. Default cursor lands on current value;
	// move down to pick "8" (index 3).
	press(m, "down", "")
	press(m, "down", "")
	press(m, "down", "")
	press(m, "enter", "")
	if m.Cfg.MaxSessions != 8 {
		t.Fatalf("MaxSessions = %d, want 8", m.Cfg.MaxSessions)
	}

	// Step 6: Default agent + model (segmented, 2 rows).
	// Default row is 0 col 0 ("codex"). Move right to opencode, tab to row 2,
	// then enter to commit.
	press(m, "right", "")
	if m.Cfg.Agent != "opencode" {
		t.Fatalf("Agent live-commit = %q, want opencode", m.Cfg.Agent)
	}
	press(m, "tab", "")
	press(m, "enter", "")

	// Step 7: Notifications — toggle one extra, then enter.
	press(m, "down", "")
	press(m, "down", "")
	press(m, " ", "")
	press(m, "enter", "")
	if !m.Cfg.Notifications.DesktopOnComplete {
		t.Fatalf("DesktopOnComplete should be toggled on")
	}

	// Step 8: Appearance (theme + density segmented).
	// Move to "tokyonight-storm" (index 3 — auto, charm-dark, charm-light, tokyonight-storm).
	press(m, "right", "")
	press(m, "right", "")
	press(m, "right", "")
	press(m, "enter", "")
	if m.Cfg.Theme != "tokyonight-storm" {
		t.Fatalf("Theme = %q", m.Cfg.Theme)
	}

	// Step 9: Advanced — default cursor lands on "info" (index 2); one step
	// down is "debug".
	press(m, "down", "")
	press(m, "enter", "")
	if m.Cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q, want debug", m.Cfg.LogLevel)
	}

	// Step 10: Summary. Focus is on Finish by default.
	msg := press(m, "enter", "")
	done, ok := msg.(DoneMsg)
	if !ok {
		t.Fatalf("expected DoneMsg, got %T", msg)
	}
	if done.Cfg.ServerURL != "wss://demo.example.com/ws" {
		t.Fatalf("DoneMsg.Cfg.ServerURL mismatch: %q", done.Cfg.ServerURL)
	}

	// Config should be on disk.
	if !config.Exists() {
		t.Fatalf("config.Exists() false after Finish")
	}
	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.ServerURL != "wss://demo.example.com/ws" ||
		reloaded.MaxSessions != 8 ||
		reloaded.Theme != "tokyonight-storm" ||
		reloaded.LogLevel != "debug" {
		t.Fatalf("reloaded config wrong: %+v", reloaded)
	}
}

func TestBackNavigationRewinds(t *testing.T) {
	m := New()
	press(m, "enter", "") // welcome → server URL
	press(m, "enter", "") // server URL → auth token (default is a valid wss:// placeholder)
	if m.step != 2 {
		t.Fatalf("pre-back step = %d", m.step)
	}
	press(m, "esc", "")
	if m.step != 1 {
		t.Fatalf("after esc step = %d, want 1", m.step)
	}
}

func TestBackOnWelcomeEmitsQuitConfirm(t *testing.T) {
	m := New()
	msg := press(m, "esc", "")
	if _, ok := msg.(QuitConfirmMsg); !ok {
		t.Fatalf("expected QuitConfirmMsg, got %T", msg)
	}
}

func TestInvalidServerURLShowsError(t *testing.T) {
	m := New()
	press(m, "enter", "") // → server URL step
	press(m, "ctrl+u", "")
	type_(m, "not-a-url")
	press(m, "enter", "")
	if m.step != 1 {
		t.Fatalf("should stay on URL step, got %d", m.step)
	}
	if m.err == "" {
		t.Fatalf("expected validation error to be set")
	}
}

func TestSummaryEditJumpsToFirstEditableStep(t *testing.T) {
	// Fast-forward through until we reach the summary.
	m := New()
	for m.step < len(steps)-1 {
		press(m, "enter", "")
	}
	if steps[m.step].ID != "summary" {
		t.Fatalf("expected to land on summary, got %q", steps[m.step].ID)
	}
	// Move focus from Finish (2) → Edit (1) and activate.
	press(m, "left", "")
	press(m, "enter", "")
	if m.step != 1 {
		t.Fatalf("Edit should jump to step 1, got %d", m.step)
	}
}

func TestAgentSwitchRebuildsModelRow(t *testing.T) {
	m := New()
	// Jump to the agent+model step (index 6).
	m.step = 6
	m.loadStep()

	// Start on row 0 col 0 = codex.
	if m.Cfg.Agent != "codex" {
		t.Fatalf("default agent = %q", m.Cfg.Agent)
	}
	// Move right to opencode; the model list should switch too.
	press(m, "right", "")
	if m.Cfg.Agent != "opencode" {
		t.Fatalf("agent not updated on right arrow: %q", m.Cfg.Agent)
	}
	groups := segmentedGroups(m, steps[m.step])
	want := modelsFor("opencode")
	if len(groups[1]) != len(want) || groups[1][0] != want[0] {
		t.Fatalf("model row did not rebuild: %v", groups[1])
	}
}
