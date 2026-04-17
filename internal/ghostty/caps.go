// Package ghostty contains terminal-capability detection and OSC sequence
// helpers for opportunistic Ghostty integration.
//
// Per the spec (§4) the TUI is designed _for_ Ghostty but must render in any
// modern terminal. Each helper degrades to a no-op when the relevant capability
// is missing.
package ghostty

import (
	"os"
	"strings"
)

// Caps describes the terminal features we care about. Populated once at
// startup by Detect(); never mutated after.
//
// Phase 1 derives every flag from environment variables — no async terminal
// queries, since those would race with Bubble Tea's input loop. Phase 2 can
// extend this with proper APC/OSC query handshakes if the need arises.
type Caps struct {
	IsGhostty      bool // $TERM_PROGRAM == "ghostty"
	TermProgram    string
	TrueColor      bool // $COLORTERM == "truecolor" or "24bit"
	OSC9Progress   bool // assume true on Ghostty ≥ 1.2 + WT/iTerm; safe no-op elsewhere
	OSC52Clipboard bool // safe to attempt on any modern term
	Hyperlinks     bool // OSC 8 — most modern terms support it
	OSC777Notify   bool // desktop notification — Ghostty + a few others
	KittyKeyboard  bool // $TERM contains "kitty" or IsGhostty
	IsTmux         bool // $TERM contains "tmux" or $TMUX is set
}

// Detect returns capabilities derived from the current process environment.
// Detect never blocks and never touches the terminal — call it before
// tea.NewProgram so the result can be passed into the model.
func Detect() Caps {
	term := os.Getenv("TERM")
	termProg := os.Getenv("TERM_PROGRAM")
	colorTerm := os.Getenv("COLORTERM")
	tmux := os.Getenv("TMUX")

	isGhostty := strings.EqualFold(termProg, "ghostty")
	isKitty := strings.Contains(strings.ToLower(term), "kitty")
	isTmux := tmux != "" || strings.Contains(strings.ToLower(term), "tmux") || strings.Contains(strings.ToLower(term), "screen")
	trueColor := colorTerm == "truecolor" || colorTerm == "24bit"

	return Caps{
		IsGhostty:      isGhostty,
		TermProgram:    termProg,
		TrueColor:      trueColor,
		OSC9Progress:   isGhostty || strings.EqualFold(termProg, "WezTerm") || strings.EqualFold(termProg, "iTerm.app"),
		OSC52Clipboard: !isTmux || os.Getenv("DAEMONCTL_FORCE_OSC52") != "",
		Hyperlinks:     true, // emit unconditionally; unsupported terms render the wrapped text fine
		OSC777Notify:   isGhostty || strings.EqualFold(termProg, "WezTerm") || isKitty,
		KittyKeyboard:  isGhostty || isKitty,
		IsTmux:         isTmux,
	}
}

// Label returns a short human-readable summary used in the splash + header
// for debug visibility ("ghostty • truecolor", "xterm", …).
func (c Caps) Label() string {
	switch {
	case c.IsGhostty:
		return "ghostty"
	case c.TermProgram != "":
		return strings.ToLower(c.TermProgram)
	case c.IsTmux:
		return "tmux"
	default:
		t := os.Getenv("TERM")
		if t == "" {
			return "unknown"
		}
		return t
	}
}
