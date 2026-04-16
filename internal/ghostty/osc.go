package ghostty

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Notify emits an OSC 777 desktop notification. Returns a tea.Cmd that writes
// the escape sequence directly to /dev/tty, bypassing Bubble Tea's renderer
// (which suppresses Printf in alt-screen mode). No-op when caps indicate the
// terminal doesn't support it.
func Notify(caps Caps, title, body string) tea.Cmd {
	if !caps.OSC777Notify {
		return nil
	}
	return func() tea.Msg {
		f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			return nil
		}
		defer f.Close()
		fmt.Fprintf(f, "\x1b]777;notify;%s;%s\x1b\\", title, body)
		return nil
	}
}

// ProgressBar returns a tea.ProgressBar reflecting the aggregate state of all
// sessions. The caller sets View.ProgressBar to this value.
//
// Rules (from spec §4):
//   - any awaiting_* → Indeterminate (pulsing orange in Ghostty)
//   - any failed with nothing running → Error
//   - any running → Default with value = completed/total * 100
//   - everything done → nil (clears the indicator)
func ProgressBar(caps Caps, completed, total int, anyAwaiting, anyFailed, anyRunning bool) *tea.ProgressBar {
	if !caps.OSC9Progress || total == 0 {
		return nil
	}
	if anyAwaiting {
		return tea.NewProgressBar(tea.ProgressBarIndeterminate, 0)
	}
	if anyFailed && !anyRunning {
		return tea.NewProgressBar(tea.ProgressBarError, completed*100/total)
	}
	if anyRunning {
		return tea.NewProgressBar(tea.ProgressBarDefault, completed*100/total)
	}
	if completed == total {
		return nil
	}
	return tea.NewProgressBar(tea.ProgressBarDefault, completed*100/total)
}

// FileHyperlink wraps text with an OSC 8 file:// hyperlink escape sequence.
// Terminals that support OSC 8 make the text clickable; others show it plain.
func FileHyperlink(absPath, display string) string {
	return fmt.Sprintf("\x1b]8;;file://%s\x1b\\%s\x1b]8;;\x1b\\", absPath, display)
}
