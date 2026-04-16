package ghostty

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDetect_Ghostty(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "ghostty")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("TMUX", "")
	t.Setenv("TERM", "xterm-ghostty")

	c := Detect()
	if !c.IsGhostty {
		t.Error("expected IsGhostty")
	}
	if !c.TrueColor {
		t.Error("expected TrueColor")
	}
	if !c.OSC9Progress {
		t.Error("expected OSC9Progress for Ghostty")
	}
	if !c.OSC777Notify {
		t.Error("expected OSC777Notify for Ghostty")
	}
	if !c.KittyKeyboard {
		t.Error("expected KittyKeyboard for Ghostty")
	}
	if c.IsTmux {
		t.Error("unexpected IsTmux")
	}
	if c.Label() != "ghostty" {
		t.Errorf("Label() = %q, want ghostty", c.Label())
	}
}

func TestDetect_Tmux(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	t.Setenv("TERM", "screen-256color")

	c := Detect()
	if c.IsGhostty {
		t.Error("unexpected IsGhostty")
	}
	if c.TrueColor {
		t.Error("unexpected TrueColor")
	}
	if !c.IsTmux {
		t.Error("expected IsTmux")
	}
	if c.OSC52Clipboard {
		t.Error("expected OSC52Clipboard off for tmux (no override)")
	}
}

func TestDetect_Tmux_ForceOSC52(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TMUX", "/tmp/tmux")
	t.Setenv("DAEMONCTL_FORCE_OSC52", "1")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "tmux-256color")

	c := Detect()
	if !c.OSC52Clipboard {
		t.Error("DAEMONCTL_FORCE_OSC52 should enable OSC52 in tmux")
	}
}

func TestDetect_PlainXterm(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TMUX", "")
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("DAEMONCTL_FORCE_OSC52", "")

	c := Detect()
	if c.IsGhostty {
		t.Error("unexpected IsGhostty")
	}
	if c.TrueColor {
		t.Error("unexpected TrueColor")
	}
	if !c.OSC52Clipboard {
		t.Error("expected OSC52Clipboard on for non-tmux")
	}
	if c.OSC777Notify {
		t.Error("unexpected OSC777Notify on plain xterm")
	}
}

func TestProgressBar_Nil_NoCaps(t *testing.T) {
	caps := Caps{OSC9Progress: false}
	pb := ProgressBar(caps, 2, 6, false, false, true)
	if pb != nil {
		t.Error("expected nil when OSC9Progress disabled")
	}
}

func TestProgressBar_Running(t *testing.T) {
	caps := Caps{OSC9Progress: true}
	pb := ProgressBar(caps, 2, 6, false, false, true)
	if pb == nil {
		t.Fatal("expected non-nil")
	}
	if pb.State != tea.ProgressBarDefault {
		t.Errorf("State = %v, want Default", pb.State)
	}
	if pb.Value != 33 { // 2/6 * 100 = 33
		t.Errorf("Value = %d, want 33", pb.Value)
	}
}

func TestProgressBar_Awaiting(t *testing.T) {
	caps := Caps{OSC9Progress: true}
	pb := ProgressBar(caps, 1, 4, true, false, true)
	if pb == nil {
		t.Fatal("expected non-nil")
	}
	if pb.State != tea.ProgressBarIndeterminate {
		t.Errorf("State = %v, want Indeterminate", pb.State)
	}
}

func TestProgressBar_Failed(t *testing.T) {
	caps := Caps{OSC9Progress: true}
	pb := ProgressBar(caps, 1, 4, false, true, false)
	if pb == nil {
		t.Fatal("expected non-nil")
	}
	if pb.State != tea.ProgressBarError {
		t.Errorf("State = %v, want Error", pb.State)
	}
}

func TestProgressBar_AllDone(t *testing.T) {
	caps := Caps{OSC9Progress: true}
	pb := ProgressBar(caps, 4, 4, false, false, false)
	if pb != nil {
		t.Error("expected nil when all completed")
	}
}

func TestFileHyperlink(t *testing.T) {
	got := FileHyperlink("/home/user/foo.go", "foo.go")
	want := "\x1b]8;;file:///home/user/foo.go\x1b\\foo.go\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("FileHyperlink = %q, want %q", got, want)
	}
}
