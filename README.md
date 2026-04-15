# daemonctl

> The coding-agent fleet controller — a TUI for orchestrating `codex` / `opencode` sessions across a remote daemon.

`daemonctl` is the terminal client for the daemon. You get a responsive split-pane view of every session the daemon is running: a sidebar listing what's alive, a live-streaming transcript, an input box for talking to the agent, and overlays (help, command palette, toasts, permission dialogs) that stay out of your way until you need them.

---

## Status

**Phase 1 — Showcase build.** The view layer is wired up against a mock engine that replays canned transcripts from `internal/fixtures/`. No real WebSocket, no PTY, no subprocess spawning yet. See `docs/tui-spec.md` for the full scope.

### What works today

- Splash → main view routing
- Header · sidebar · session header · transcript · input · footer
- Mock engine streaming markdown transcripts in real time
- **Help overlay** (`?`), **command palette** (`⌃p`, 20 actions, fuzzy filter)
- **Dialogs**: alert, confirm, question (1–9), permission
- **Toasts**: bottom-right stack, 4 s TTL
- Dev cheats for injecting modals and state transitions on demand
- Responsive layout (compact / normal / wide breakpoints)

### What's next

Onboarding wizard · settings page · light/auto themes · Ghostty caps (OSC 9;4, 52, 8, 777) · `--dev` flag gating · teatest golden files.

---

## Install & run

```sh
# Build the binary
go build -o daemonctl ./cmd/daemonctl

# Run it (uses mock data)
./daemonctl

# Or just go run it
go run ./cmd/daemonctl
```

Target runtime: **Ghostty ≥ 1.2** on macOS or Linux. It degrades gracefully on Alacritty, Kitty, iTerm2, WezTerm, and tmux-wrapped sessions.

Minimum terminal size: **60 × 16**. Below that you get a centered "too small" message.

---

## Keyboard

### Global

| Key               | Action                            |
| ----------------- | --------------------------------- |
| `?`               | Toggle help overlay               |
| `⌃p`              | Command palette                   |
| `⌃,`              | Settings *(soon)*                 |
| `⌃c` then `⌃c`    | Quit (double-tap within 500 ms)   |
| `⌃b`              | Toggle sidebar                    |
| `tab` / `⇧tab`    | Focus next / previous pane        |
| `esc`             | Close any open modal              |

### Sidebar

`j`/`k` or `↓`/`↑` move · `g`/`G` top/bottom · `↵` or `l` open session · `/` filter.

### Transcript

`j`/`k` scroll · `g`/`G` jump · `i` focus input · `y` yank to clipboard.

### Input

`↵` send · `esc` exit input · `⌫` delete.

### Dialogs

`1`–`9` pick option · `←`/`→` move focus · `↵` confirm · `esc` cancel.

### Dev cheats (phase 1)

These inject mock events so the UI can be exercised without a real backend:

| Key    | Effect                                   |
| ------ | ---------------------------------------- |
| `⌃⌥p`  | Show a permission dialog                 |
| `⌃⌥q`  | Show an agent question dialog            |
| `⌃⌥t`  | Push a random toast                      |
| `⌃⌥c`  | Force-complete the current session       |
| `⌃⌥f`  | Force-fail the current session           |

---

## Development

```sh
# Build all packages
go build ./...

# Run unit tests
go test ./...

# Vet for common issues
go vet ./...

# Run with verbose test output
go test -v ./...

# Tidy module dependencies
go mod tidy

# Format the tree
gofmt -s -w .
```

### Project layout

```
cmd/daemonctl/main.go      entry point
internal/
  app/                     root Bubble Tea model, routing, message dispatch
  component/
    header/  sidebar/  transcript/  input/  footer/  splash/
    help/    palette/  toast/  dialog/      ← overlays
  layout/                  breakpoints + pane rects
  theme/                   Theme, Styles, per-status colors
  keys/                    key.Binding registry
  session/                 Session, Status, Store
    mock/                  fixture loader + transcript replayer
  fixtures/                tasks.json + transcript scripts
docs/
  tui-spec.md              the source of truth for Phase 1
```

### Working against the fixtures

Transcripts are declarative scripts under `internal/fixtures/transcripts/`. The parser understands these directives:

```
@@STATE: running
@@DELAY: 500
Reading the repo structure…

@@STATE: awaiting_perm
@@PERM: {"tool":"bash","args":"rm -rf /tmp/test","diff":null}
```

Each fixture maps 1:1 to a task in `internal/fixtures/tasks.json`. Add a task there, drop a matching `.md` file in `transcripts/`, and the mock engine will pick it up on next launch.

---

## Architecture notes

- Pure Go, **no CGO**, `go 1.23+`.
- Built on [`charm.land/bubbletea/v2`](https://charm.land/bubbletea) + [`charm.land/lipgloss/v2`](https://charm.land/lipgloss) + [`glamour`](https://github.com/charmbracelet/glamour) for markdown.
- Single root `Model`; children get pre-computed sizes via message passing — no component reads the raw window size.
- Overlays composited via `lipgloss.Compositor` with explicit Z-indices.
- Phase-2 seams (`session.Engine`, `session.Transport` interfaces) are already defined so the real PTY / WebSocket implementations can drop in without touching the view layer.

Design non-negotiables live in `docs/tui-spec.md` §16. Read them before making structural changes.

---

## Contributing

1. Pick a milestone from `docs/tui-spec.md` §15 (M1–M14).
2. Keep changes scoped to that milestone — small, reviewable PRs beat big ones.
3. `go test ./...` before opening a PR.
4. Questions that affect product decisions go in `docs/product-decisions.md` (create it if missing), not inline code comments.
