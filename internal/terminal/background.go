// Package terminal provides terminal capability detection.
//
// DetectBackground uses termenv to query the terminal's background color via
// OSC 11 and computes relative luminance to decide dark vs light. The query
// has a 100ms timeout — if the terminal doesn't respond (piped stdin, dumb
// term, SSH without forwarding) we fall back to dark.
package terminal

import (
	"os"

	"github.com/muesli/termenv"
)

// IsDarkBackground queries the running terminal for its background color and
// returns true when the background is dark (or when detection fails).
//
// It uses termenv's OSC 11 query under the hood, which respects the terminal's
// actual background — not $COLORFGBG or other heuristics.
func IsDarkBackground() bool {
	return termenv.NewOutput(os.Stdout).HasDarkBackground()
}
