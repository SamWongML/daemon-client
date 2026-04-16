# Daemon Client TUI — Implementation Spec (Phase 1: Showcase Build)

> **Audience:** engineers implementing the daemon client TUI.
> **Phase:** 1 of N. Ship a pixel-perfect, interaction-complete TUI with **mock data only**. All network I/O (WebSocket, PTY spawning, session execution) is explicitly **out of scope** — stubbed behind interfaces so phase 2 can slot in real implementations without touching the view layer.
> **Target terminal:** Ghostty ≥ 1.2 (macOS + Linux). Degrade gracefully on tmux/other terminals but do not block on it.
> **Deliverable:** a runnable binary that, when launched, demonstrates the full end-to-end user experience with realistic fake sessions, transcripts, notifications, and state transitions.

---

## 0. Glossary

| Term             | Meaning                                                                                                                                |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Task**         | A unit of work pushed by the server. In phase 1, read from a JSON fixture.                                                             |
| **Session**      | A running coding-agent process (codex / opencode) associated with a task. In phase 1, fake: a goroutine replaying a canned transcript. |
| **Agent binary** | `codex` or `opencode` executable. In phase 1, never actually launched.                                                                 |
| **appModel**     | The single root Bubble Tea model. All state lives here.                                                                                |
| **Pane**         | A focusable region of the main layout (sidebar, transcript, input).                                                                    |
| **Screen**       | A top-level route — `splash`, `onboarding`, `main`, `settings`, `help`.                                                                |

---

## 1. Scope

### In scope (phase 1)

- Splash / landing screen with animated ASCII logo.
- Full onboarding wizard (11 steps) with config persistence to disk.
- Main view: header + sidebar + transcript + input + footer, fully responsive.
- Settings page (accessible from main view).
- Help overlay (`?`).
- Command palette (`⌃p`).
- Permission / question modal.
- Toast notification stack.
- Theme system (dark / light / auto) with 3 built-in themes.
- Full keyboard navigation including horizontal focus traversal.
- Ghostty-specific enhancements (see §6).
- Fake session engine that drives the UI through realistic state transitions.
- Mock fixtures (see §14) — enough to showcase every visible state.

### Out of scope (phase 1, backlog for phase 2+)

- WebSocket client / reconnect logic.
- Real PTY management (`creack/pty`).
- Actual `codex` / `opencode` subprocess spawning.
- Server auth token exchange.
- SQLite persistence of transcripts.
- Cloud response channel.
- Cross-session search.
- Log audit view.

These **must** be isolated behind Go interfaces (§12.2) so phase 2 can implement without view-layer churn.

---

## 2. Tech stack

| Purpose            | Library                                              | Version  | Notes                                                                                                       |
| ------------------ | ---------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| TUI framework      | `charm.land/bubbletea/v2`                            | latest   | v2 is required — we need `FocusMsg`/`BlurMsg`, mouse v2, `tea.View` struct.                                 |
| UI components      | `github.com/charmbracelet/bubbles/v2`                | matching | list, viewport, textinput, textarea, spinner, progress, help, key, paginator, filepicker, stopwatch, timer. |
| Styling            | `github.com/charmbracelet/lipgloss/v2`               | matching | All colors go through `lipgloss.AdaptiveColor`.                                                             |
| Forms (onboarding) | `github.com/charmbracelet/huh`                       | latest   | Used for the wizard.                                                                                        |
| Markdown render    | `github.com/charmbracelet/glamour`                   | latest   | For transcript rendering.                                                                                   |
| Fuzzy search       | `github.com/sahilm/fuzzy`                            | latest   | For command palette.                                                                                        |
| Text wrap          | `github.com/muesli/reflow`                           | latest   | Word-safe wrap for transcripts.                                                                             |
| CJK-safe width     | `github.com/mattn/go-runewidth`                      | latest   | Don't assume ASCII width.                                                                                   |
| Config             | `github.com/adrg/xdg` + `github.com/BurntSushi/toml` | latest   | TOML at `$XDG_CONFIG_HOME/daemonctl/config.toml`.                                                           |
| Logging            | `log/slog` (stdlib)                                  | 1.22+    | To file only, never stdout.                                                                                 |

**No CGO.** Pure Go build. Target: `go 1.23+`.

**Nexus note:** all of the above are hosted on `proxy.golang.org` and should flow through a standard corporate Go module proxy. No npm, no pip. Verify `GOPROXY` at project init but do not block on whitelist.

---

## 3. Project layout

```
cmd/daemonctl/main.go              # entry point, flag parsing, tea.NewProgram
internal/
  app/
    app.go                         # appModel, root Update, View
    router.go                      # screen routing (splash|onboarding|main|settings)
    messages.go                    # all tea.Msg types, one file, typed
  screen/
    splash/                        # animated landing
    onboarding/                    # huh wizard
    main/                          # header + sidebar + chat + input + footer
    settings/                      # config editor
  component/
    header/                        # global app header
    sidebar/                       # session list with custom delegate
    transcript/                    # viewport + glamour
    input/                         # textarea with mode switching
    footer/                        # context-sensitive help hints
    palette/                       # command palette (modal)
    dialog/                        # permission / confirm / question / alert
    toast/                         # notification stack
  theme/
    theme.go                       # Theme interface, AdaptiveColor palette
    dark.go light.go auto.go       # built-ins
  layout/
    breakpoint.go                  # responsive breakpoint system
    focus.go                       # focus graph / traversal
  keys/
    keys.go                        # single source of truth for all key.Binding
  session/
    types.go                       # Session, Status, Task (shared types)
    store.go                       # in-memory store + pubsub
    mock/                          # MockEngine: replays fixtures as live sessions
  config/
    config.go                      # load/save TOML
    defaults.go
  fixtures/
    tasks.json transcripts/        # see §14
  ghostty/
    caps.go                        # terminal capability detection
    osc.go                         # OSC 9;4, OSC 52, OSC 8, OSC 777 helpers
docs/
  tui-spec.md                      # this file
  architecture.md                  # phase 2 notes
```

---

## 4. Terminal target: Ghostty

The TUI is designed **for Ghostty** and opportunistically uses Ghostty-specific features. It must still render correctly (degraded) on Alacritty, Kitty, iTerm2, WezTerm, and tmux-wrapped sessions.

Detection lives in `internal/ghostty/caps.go`:

```go
type Caps struct {
    IsGhostty      bool   // $TERM_PROGRAM == "ghostty"
    TrueColor      bool   // $COLORTERM == "truecolor" || "24bit"
    KittyGraphics  bool   // query via APC <ESC>_Gi=1,a=q;\e\ + response within 100ms
    KittyKeyboard  bool   // push level 1 flags, read back, pop
    OSC9Progress   bool   // assume true for Ghostty ≥ 1.2, else false
    OSC52Clipboard bool   // assume true on modern terms
    Hyperlinks     bool   // OSC 8
    BackgroundRGB  [3]int // OSC 11 query with 100ms timeout
    IsDark         bool   // derived from BackgroundRGB luminance
}
```

**Features we use when available:**

1. **Kitty graphics protocol** — splash screen can render a real PNG logo (see §7.1). Fallback: ASCII art.
2. **OSC 9;4 progress reporting** — when any session is `running`, emit `ESC ] 9 ; 4 ; 1 ; <pct> ST`; when `awaiting_*`, emit state `2` (indeterminate error = pulsing orange in Ghostty's tab bar). Aggregate across all sessions — progress = `count(completed) / count(total)` this session. This gives users a native Ghostty tab-bar progress bar for free.
3. **OSC 52 clipboard** — `y` in transcript yanks selection to system clipboard via OSC 52, no external deps.
4. **OSC 8 hyperlinks** — in transcripts, file paths become `OSC 8 ; ; file:///abs/path ST<text>ESC ] 8 ; ; ST`. Ghostty makes them clickable; other terms show plain text.
5. **OSC 777 notifications** — `ESC ] 777 ; notify ; <title> ; <body> ST` when a session enters `awaiting_*`. Ghostty shows a native desktop notification; silent fallback elsewhere.
6. **Kitty keyboard protocol** — push with `CSI > 1 u` on startup, pop on exit. Lets us distinguish Tab vs Shift+Tab vs Ctrl+Tab cleanly, and detect key release events for hold-to-repeat.
7. **OSC 11 background color query + OSC 4** — `theme: "auto"` uses the actual background luminance, not guesses.
8. **Synchronized rendering (DECSET 2026)** — Bubble Tea v2 emits this automatically; do not disable.
9. **Undercurl** (`CSI 4 : 3 m`) — used for "unread output" indicator on sidebar rows.

**Forbidden:**

- Do not hardcode ANSI colors below 256. Always use `lipgloss.AdaptiveColor` with true-color values.
- Do not use box-drawing characters that don't exist in half-width fonts; stick to `╭╮╰╯│─┤├┬┴┼` and double variants.

---

## 5. Responsive layout system

### 5.1 Breakpoints

Define three breakpoints in `internal/layout/breakpoint.go`:

| Name      | Width range                 | Layout                                                                                |
| --------- | --------------------------- | ------------------------------------------------------------------------------------- |
| `Compact` | `< 100` cols OR `< 24` rows | single-pane stack: only the active screen visible; sidebar becomes an overlay toggle  |
| `Normal`  | `100–159` cols              | sidebar (28 cols) + main; footer collapses to 1 line                                  |
| `Wide`    | `≥ 160` cols                | sidebar (`min(36, 22% width)`) + main + optional inspector (32 cols, toggle with `I`) |

```go
type Breakpoint int
const (
    Compact Breakpoint = iota
    Normal
    Wide
)

func For(w, h int) Breakpoint {
    if w < 100 || h < 24 { return Compact }
    if w < 160          { return Normal }
    return Wide
}
```

`appModel` holds `breakpoint Breakpoint` and recomputes it on every `tea.WindowSizeMsg`. Each child component's `SetSize(w, h int)` is called with pre-computed dimensions — **never** let a child guess its own size from the raw window.

### 5.2 Size propagation rules

1. Exactly one place computes dimensions: `appModel.layout()`. It runs on `WindowSizeMsg` and produces a `LayoutRect` struct with explicit `{X, Y, W, H}` for every pane.
2. After computing, the model calls `SetSize` on every child with **content area** (frame size already subtracted). Use `lipgloss.Style.GetFrameSize()` — do not subtract margins manually.
3. Children store their dimensions and use them in `View()`. They **must not** call `lipgloss.Width(m.window)` or similar.
4. A `ResizeDebounceMs = 16` buffer coalesces rapid resize events (important when dragging the terminal window edge): stash the latest `WindowSizeMsg`, schedule a `tea.Tick` of 16ms, only apply on tick. This is the single biggest source of flicker.

### 5.3 Layout math (Normal breakpoint)

All chrome is borderless (§6.4.1). Separators are dim `─` / `│` rules that occupy 0 extra lines (they replace a padding line). The diagram below uses `─` and `│` for separators, **not** box-drawing borders:

```
 HEADER (h=1, w=W)
─────────────────────────────────────────────────────────────────
 SIDEBAR (w=28)  │  SESSION HEADER (h=1)
                 │ ─────────────────────────────────────────────
                 │  TRANSCRIPT (flex)
                 │
                 │ ─────────────────────────────────────────────
                 │  INPUT (h=dynamic, min 3, max 10)
─────────────────────────────────────────────────────────────────
 FOOTER (h=1, w=W)
```

Formula:

```
headerH = 1
footerH = 1
sessionHeaderH = 1      // was 2 — now single-line per §7.3.3
separatorH = 0           // dim rules are drawn inside padding, not additional rows
sidebarW = 28 (normal) | min(36, W*22/100) (wide) | 0 (compact-collapsed)
sidebarH = H - headerH - footerH

mainX = sidebarW + 1     // +1 for the vertical │ separator
mainW = W - sidebarW - 1
inputH = clamp(m.input.LineCount() + 2, 3, 10)
transcriptH = H - headerH - footerH - sessionHeaderH - inputH
```

In Wide mode, an optional `inspectorW = 32` is subtracted from `mainW`.

### 5.4 Terminal-too-small screen

If `W < 60` or `H < 16`, render a centered `Too small — please resize (min 60×16)` message in lipgloss and short-circuit the normal View.

---

## 6. Focus system

### 6.1 Focus graph

There is a **single focus pointer** in `appModel.focus FocusID`. A `FocusGraph` is a directed graph with Tab/Shift+Tab as the default traversal order and optional arrow-key overrides per node.

```go
type FocusID string

type FocusNode struct {
    ID        FocusID
    Next      FocusID // tab / right
    Prev      FocusID // shift-tab / left
    Up, Down  FocusID // optional; if empty, arrow keys pass through
    OnFocus   tea.Cmd // e.g. textinput.Focus()
    OnBlur    tea.Cmd
}

type FocusGraph map[FocusID]FocusNode
```

Each screen builds its own graph in its `Init()`. Switching screen resets the focus to the graph's `EntryID`.

### 6.2 Horizontal focus within a line ("change focused position in a line")

When a row contains multiple interactive elements (e.g. the action buttons at the bottom of an onboarding step, the header's `[settings] [help] [theme]` chip group), `←`/`→` move focus within the row and `Tab`/`Shift+Tab` move to the next/previous row.

Implementation: a `HRow` focus node has `Prev`/`Next` wired left/right for arrow keys, and a separate `ParentGroup` pointer so `Tab` can jump to the next group instead of cycling within the row.

Example (onboarding summary step):

```
[ ← Back ]  [ Edit ]  [ ✓ Finish ]   ← ↔ arrows here
     ↑                      │
     └── Tab goes to ───────┘
         the title above
```

### 6.3 Focus visual feedback

A focused element gets **all three** of:

1. A left accent bar (`▌` in accent color) — works where borders would waste space.
2. Slightly brighter foreground on its label.
3. If it's a text input, a visible block cursor (`Cursor.Shape = tea.CursorBlock, Blink = true` via `tea.View.Cursor`).

Buttons render differently based on focus × enabled:

| State               | Render                                     |
| ------------------- | ------------------------------------------ |
| enabled + unfocused | `[ Finish ]` dim fg                        |
| enabled + focused   | `▌ Finish ▐` accent bg, bold               |
| disabled            | `[ Finish ]` strikethrough + very dim      |
| primary + focused   | `▌ ✓ Finish ▐` accent bg, bold, with glyph |

---

## 6.4 Visual language & density principles

Phase 1 follows a **minimalist, chrome-light aesthetic** inspired by opencode (sst/opencode) and Crush (charmbracelet/crush). These rules are load-bearing — every section below defers to them. If a lower-section sketch appears to violate one of these rules, the rule wins.

**References:** opencode's spartan header + toggleable chrome ([opencode.ai/docs/tui](https://opencode.ai/docs/tui/)); Crush's `BorderThick = ▌` left-edge accent instead of boxed frames ([Crush styling system](https://deepwiki.com/charmbracelet/crush/5.8-styling-system)); Claude Code's "no chrome, conversation first"; community trend toward single-accent palettes (Catppuccin, Charmtone).

### 6.4.1 Borders

- **No box borders on persistent chrome.** Header, sidebar, transcript, input, footer, session header — none of these get `lipgloss.RoundedBorder()` or `NormalBorder()`.
- **Separation is achieved by:**
  - A single vertical rule `│` (dim `theme.Border`) between sidebar and main pane.
  - A single dim horizontal rule `─` (full width) between header ↔ body and body ↔ footer. Use `lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false)` (bottom only) — never four-sided.
  - Whitespace (1-col horizontal padding inside each pane).
- **Box borders are reserved for modals:** command palette, help overlay, all four dialog types, filter bar popover. These may use `lipgloss.RoundedBorder()` in accent color.
- The "New session" row at the top of the sidebar is **not** a boxed button — it is a single row with a `+` glyph and a left accent bar when focused. See §7.3.2.

### 6.4.2 Status glyphs

All session-state indicators use a **single-cell dot vocabulary** (Crush + Charmtone convention). No nerd-font icons, no emoji, no mixed metaphors.

| Glyph | Meaning                                       | Color source       |
| ----- | --------------------------------------------- | ------------------ |
| `●`   | Running / active tool use                     | `theme.StRunning`  |
| `◐`   | Pending / starting (pair with pulse, 1 Hz)    | `theme.StPending`  |
| `○`   | Idle / paused                                 | `theme.Muted`      |
| `✓`   | Completed                                     | `theme.Success`    |
| `×`   | Failed (`U+00D7`, not `✗`)                    | `theme.Danger`     |
| `!`   | Awaiting input / permission (pulse, 1 Hz)     | `theme.Warn`       |
| `⋯`   | Spinner fallback when `◐` would churn rapidly | `theme.StPending`  |

Rendering rule: one glyph + one space, never two glyphs in a row, never glyph-on-background. Color carries the meaning; the glyph shape is secondary.

### 6.4.3 Color discipline

- **One accent.** Pick one hue per theme (Charple-violet for `charm-dark`, Catppuccin mauve for `tokyonight-storm`, gruvbox orange for `gruvbox-hard`). It marks focus, primary buttons, and the sidebar selection bar — **nothing else**.
- Four semantic colors for status (`Success`, `Warn`, `Danger`, `Info`). These are used only on status glyphs and dialog severity — never on body text.
- Everything else is a foreground ramp: `Fg` (primary), `Muted` (~60%), `Dim` (~35%). Resist coloring more than ~5% of visible cells.

### 6.4.4 Density

- Default density is **comfortable**: 1 col horizontal padding, sidebar rows are 2 lines (not 3). `compact` density collapses sidebar rows to 1 line (glyph + title only).
- The "3-line sidebar row" from earlier drafts is removed — line 3's activity text lives in the session header instead.
- Header, session header, and footer are each **exactly 1 line**. No 2-line variants.

---

## 7. Screens

### 7.1 Splash / landing screen

First screen on every launch (even after onboarding). Duration: **minimum 800ms, maximum 1800ms**, then auto-advances. User can press any key to skip.

**Layout:** fullscreen alt-screen, everything centered via `lipgloss.Place(w, h, Center, Center, ...)`.

**Content (top to bottom):**

1. **Logo** — 8-row ASCII block in the accent gradient. The ASCII is rendered through a per-character gradient using `lipgloss`'s 24-bit color (compute per-column hue from a sine). On Ghostty with Kitty graphics detected, replace ASCII with a 120×60 px PNG loaded from `assets/logo.png` via the Kitty protocol; `a=T, f=100` (PNG direct transmission), `c=24, r=8`.
2. **Wordmark** — app name in `lipgloss.NewStyle().Bold(true)`, 2 rows tall (figlet-style).
3. **Tagline** — `the coding-agent fleet controller` in dim italic.
4. **Version + build** — `v0.1.0 • 2026-04-14 • ghostty` in dim.
5. **Loading bar** — animated `bubbles/progress`, 40 cols wide, completes from 0→100% over the splash duration. Label cycles: `loading config…` → `checking agents…` → `connecting…` (all fake in phase 1).
6. **Hint** — `press any key to skip` at the very bottom, dim.

**Animation:** the logo uses a 3-frame cycle at 8fps where the gradient phase shifts, creating a gentle shimmer. Use `tea.Tick(125ms)` driving a frame counter — **not** `harmonica`, overkill.

**Ready-state transition:** when the loading bar hits 100%, do a 150ms fade-out using a `FadeMsg{phase: 0..1}` where each tick dims the foreground color. Then emit `SplashDoneMsg` which the router switches on.

**First-run detection:** on `SplashDoneMsg`, router checks `config.Exists()`:

- If no config → route to `onboarding`.
- If config exists → route to `main`.

### 7.2 Onboarding wizard (11 steps)

Built on `huh`. Each step is a `huh.Group`. Steps:

1. **Welcome** — `huh.NewNote` with a re-used (smaller) version of the ASCII logo and a one-line intro. Action: `[ Let's go → ]`.
2. **Server URL** — `huh.NewInput` with `Validate` = URL parser that requires `wss://` or `ws://` scheme. Placeholder: `wss://your-server.example.com/ws`.
3. **Auth token** — `huh.NewInput` with `.EchoMode(EchoPassword)`. Side hint: `press ⌘v / ⌃⇧v to paste`. Optionally a `[ ] import from ~/.netrc` confirm.
4. **Workdir** — custom filepicker bubble wrapped as a `huh.Field`. Default `$PWD`. Create-if-missing checkbox.
5. **Agent binaries** — two-row custom field: auto-detect `codex` and `opencode` on `$PATH`. Each row shows `codex ✓ /usr/local/bin/codex` or `codex ✗ not found [ override… ]`. In phase 1, always pretend both are found.
6. **Max concurrent sessions** — `huh.NewSelect` with options `[1, 2, 4, 8, custom…]`. Default `min(4, runtime.NumCPU()/2)`.
7. **Default agent + model** — two dependent `huh.NewSelect`s. Agent options: `codex | opencode`. Model options depend on agent (for phase 1, hardcoded lists; phase 2 will fetch from server).
8. **Notifications** — `huh.NewMultiSelect`:
    - `[x] terminal bell on attention required`
    - `[x] desktop notification on attention required`
    - `[ ] desktop notification on session completion`
    - `[ ] desktop notification on session failure`
9. **Appearance** — two-column row (use `HRow` focus):
    - Left: theme selector `[ dark | light | auto ]` (segmented control).
    - Right: density `[ comfortable | compact ]`.
10. **Advanced** (collapsible) — log level `[ off | error | info | debug ]`, log file path, telemetry opt-in (default off).
11. **Summary** — renders the resolved config as syntax-highlighted TOML through glamour, plus a `[ ← Back ] [ Edit section… ] [ ✓ Finish ]` button row (horizontal focus).

**Step indicator:** top of each step, `● ● ● ○ ○ ○ ○ ○ ○ ○ ○   (3/11 — Workdir)`. Filled = accent color, empty = dim.

**Back navigation:** `Esc` or `←` on the first element goes back a step. First step's back is a confirm-quit.

**Persist partial state:** on every field commit, write the in-progress state to `$XDG_STATE_HOME/daemonctl/onboarding.toml`. On next launch, detect and offer `Resume onboarding? [Y/n]`.

### 7.3 Main view

The default screen. Composed of Header + Sidebar + Session header + Transcript + Input + Footer, with all sizing done by the layout rules in §5.

#### 7.3.1 Header (height 1)

Follows §6.4.1 (no border, one line only). Layout is two flushed groups separated by whitespace — **no `•` separators inside a group**; spacing alone does the work.

```
 ▌daemonctl   ● codex · gpt-5                                        [⚙] [?]
```

Left group:

- `▌daemonctl` — wordmark, accent color.
- `● <agent> · <model>` — active session's agent/model. The `●` is the WebSocket/transport status dot (green/yellow/red from §6.4.2). Phase 1 default: green; `⌃⌥1/2/3` cycles states.

Right group:

- `[⚙]` — settings, `FocusID = "header.settings"`, hotkey `⌃,`.
- `[?]` — help, `FocusID = "header.help"`, hotkey `?`.

**Everything else moved out of the header:**

- Server URL → Settings → Connection, and visible on hover of the status dot (OSC 8 title attribute on supporting terminals).
- Session count `active/max` → footer status line (§7.3.6).
- Clock → footer status line (right-aligned). Rationale: Ghostty already shows the OS clock; duplicating it here adds noise. If footer hides (Compact breakpoint), the clock disappears entirely — that is intentional.
- Aggregate cost → footer status line, and on-demand via `/cost` palette action.

> **Note to executor:** `[⚙]` in the header remains the canonical settings entry point and must be visible in every breakpoint. In Compact mode the header still renders `▌daemonctl   [⚙] [?]` (agent/model drops); the transport dot moves to the footer.

#### 7.3.2 Sidebar

28 cols wide (Normal). Custom list renderer (not the default `bubbles/list` delegate). No border around the sidebar itself — separated from the main pane by a single vertical `│` rule in `theme.Border` (§6.4.1).

**Title row** (line 1, followed by one dim `─` separator line):

```
 sessions   3
```

Lowercase, no `/max` (that lives in the footer). Column title is `theme.Muted`.

**Each item = 2 lines** (comfortable density) / **1 line** (compact density):

```
 ● refactor auth module
   running · 2m14s
```

- Line 1: `glyph` (§6.4.2) + title. Title truncates with `…`, padded to column width.
- Line 2: `status-label · elapsed` in `theme.Muted`. No token count here — moves to the session header. Compact density drops this line entirely.
- **Selected**: left accent bar `▌` replaces the leading space. Foreground brightens to `theme.Fg`. **No background tint** — the accent bar alone communicates selection (inspired by Crush).
- **Unread output** (new content since last focus): small `•` in accent color prepended to the title, no undercurl. Undercurl is reserved for search-match highlighting.
- **Attention state** (`awaiting_input` / `awaiting_perm`): glyph is `!`, pulsing at 1 Hz (§11). Row gains a 1-cell left accent bar in `theme.Warn` even when unselected, so it's scannable without focus.

**"New session" row** (sticky, line 0 of the list, above the title row):

```
 + new session                        ⌃n
```

One line. No box. Focused state renders `▌` accent bar + bright fg. Hotkey right-aligned in `theme.Dim`. Following §6.4.1, boxes are reserved for modals.

**Filter bar** appears above the title row when `/` pressed (slide-down 120ms):

```
 / auth_
```

Plain row, single-cell `/` prefix in accent, no border. Dismissed by `Esc` or empty query + `Enter`. (The earlier rounded-border popup sketch is retired — `/` filter is a modal input, but a *lightweight* one; the border would break the sidebar's vertical rhythm.)

Match highlights inline using `lipgloss.NewStyle().Underline(true)`.

**Ordering — time buckets with severity as secondary sort:**

Items are **not** grouped by status. They are grouped by **last-activity time** into up to five buckets, in this fixed order. Empty buckets are not rendered:

1. `active`    — any session currently in `running` / `awaiting_*` / `starting` (pinned to top regardless of age)
2. `today`     — last-activity within the last 24h
3. `this week` — last-activity within the last 7d
4. `older`     — everything else still alive
5. `archive`   — terminal states (`completed`, `failed`, `disconnected`) explicitly archived by the user

Bucket headers are 1-line dim lowercase text with a trailing dim rule, no border:

```
─── today ──────────────────
```

**Within each bucket**, rows are sorted by **severity descending**, then by last-activity descending. The severity ladder (highest first, see §10 `Priority`):

`awaiting_perm > awaiting_input > failed > disconnected > running > starting > pending > idle > paused > completed`

This surfaces what needs attention at the top of the visible bucket without fragmenting the list by status. Pattern matches Todoist/TickTick "group by time, sort by priority".

**Completed + archived dimming:** rows in `archive` render at ~40% opacity (`theme.Dim` fg, muted glyph). Density collapses to 1 line in archive regardless of the global density setting.

#### 7.3.3 Session header (height 1)

Single line. Border-less (§6.4.1). Left/right flushed with whitespace as the separator; no `•` inside a group.

```
 ● refactor auth module   2m14s   14.2k/128k ▓▓▓▓▓▓▓░░░░░░
```

- Left: status glyph (§6.4.2) + title (truncates with `…`).
- Middle: elapsed time in `theme.Muted`.
- Right: token usage `used/budget` + a 12-col inline context-window bar (`bubbles/progress`, `WithSolidFill`, no percent text). Bar color = `theme.Accent` at low usage, `theme.Warn` above 80%, `theme.Danger` above 95%.

Agent/model is already shown in the app header (§7.3.1) — do not repeat it here. Cost is in the footer status line.

A single dim `─` rule below this row separates it from the transcript.

#### 7.3.4 Transcript pane

`bubbles/viewport` wrapping `glamour`-rendered markdown. The fake session engine feeds it markdown chunks via `TranscriptAppendMsg{SessionID, Content}`.

**Rendering rules:**

- User messages: `▌` left bar in blue, no background.
- Agent messages: no bar, markdown rendered through glamour.
- Tool calls: collapsible blocks with a header `▼ bash  · 230ms · exit 0` followed by a dim code block. Toggle with `Space` when cursor over the header.
- File diffs: rendered as unified diff with green `+` / red `-` lines.
- Thinking blocks: rendered in a dim italic style, togglable globally via `⌃x t`.
- File paths: wrapped in OSC 8 hyperlinks when Caps.Hyperlinks.

**Scroll state:**

- `auto_follow` bool: when new content arrives, if the viewport is at bottom (`YOffset >= TotalLines - Height`), auto-scroll. If the user has scrolled up, don't.
- Scroll-to-top `g`, scroll-to-bottom `G`, half-page `⌃d`/`⌃u`.
- `/` enters search mode (incremental).
- `y` yanks selection to clipboard via OSC 52.

**Throttling:** chunks are coalesced via a 33ms `tea.Tick`. The mock engine respects this.

#### 7.3.5 Input pane

`bubbles/textarea` with custom rendering. Height grows with content, clamped `[3, 10]`.

Mode indicator on the left edge of line 1 of the input frame:

| Mode      | Glyph | Color  | Notes                                     |
| --------- | ----- | ------ | ----------------------------------------- |
| `prompt`  | `>`   | accent | normal message to agent                   |
| `shell`   | `!`   | yellow | typed as first char; submit runs as shell |
| `slash`   | `/`   | cyan   | slash command completion                  |
| `respond` | `↵`   | green  | responding to a pending question          |
| `perm`    | `⚠`   | orange | approving/denying a tool permission       |

Mode switches automatically based on first char. `Esc` clears mode.

Autocomplete dropdown appears above the input as a floating panel (max 6 rows, scrollable). Uses `fuzzy` for ranking.

Placeholder text rotates every 4 seconds through a small list: `ask the agent…`, `describe the task…`, `press / for commands…`, `press @ to mention a file…`.

#### 7.3.6 Footer (height 1)

Single line, borderless, split into two zones with whitespace between them:

```
 tab next   ↵ send   ⌃p palette   ⌃b sidebar   ? help        3/4  $0.42  14:32
```

**Left zone — keybind hints.** Context-sensitive from the focused component's `key.Bindings`, separated by two spaces (no `·` / `|`). Bubbles' `help.Model` renders this; the binding list is a union of: global + current screen + focused pane. Duplicates removed. Order: pane-specific → screen → global. Truncate with `…` on overflow.

**Right zone — status line.** Ambient state that used to live in the header (§7.3.1). Items are rendered only if they fit (right-to-left drop order: clock → cost → session-count):

1. `active/max` session count (e.g. `3/4`).
2. Aggregate cost (e.g. `$0.42`).
3. Local time `HH:MM` (updates every 60s — not every second; no need for second-precision here).

All right-zone items are `theme.Dim`. They carry no focus and no click target — they're passive readouts.

**Calm-state rule** (opencode-inspired): when no session is active and no state has changed in 5s, the left zone fades to a single hint (`press ⌃p for commands`) and gently rotates every ~8s through `press ? for help`, `press ⌃n for a new session`. Animation reuses the global `FrameMsg` tick — do not schedule a dedicated ticker.

### 7.4 Settings page

Entered via header `[⚙]` or `⌃,`. Layout: fullscreen with a left sidebar of categories and a right form panel.

**Categories** (vertical list, focusable):

1. Connection (server, token — redacted, last-used timestamp)
2. Workspace (workdir, max sessions)
3. Agents (binaries, default agent, default model)
4. Appearance (theme, density, font hints)
5. Notifications
6. Keyboard (view current bindings, reset to default)
7. Advanced (log level, telemetry, config file path)
8. About (version, build, link to docs via OSC 8)

Right panel renders fields from the same `huh` components used in onboarding, but editable inline. `⌃s` saves. `Esc` returns to main.

Visual framing: use `lipgloss.RoundedBorder()` around the right panel, not the sidebar.

### 7.5 Help overlay

Triggered by `?` anywhere. Renders as a **centered modal** (not fullscreen), `min(120, width-4) × min(35, height-4)`.

Content:

- Title: `Keyboard shortcuts`
- Two columns of binding groups: `Global`, `Sidebar`, `Transcript`, `Input`, `Dialogs`.
- Each binding rendered as `  ⌃p    command palette`.
- Footer: `press ? or esc to close`.

Background: semi-transparent dim over the main view (`lipgloss.NewStyle().Faint(true)` on the underlying view).

### 7.6 Command palette

Triggered by `⌃p`. Modal, centered, `60 × 20`.

- Top: text input with `> ` prompt.
- Middle: fuzzy-filtered action list, max 12 visible, each row `icon name   shortcut`.
- Bottom: dim hint `↑↓ navigate   ↵ run   esc close`.

Actions are registered globally from a central `palette.Register(Action{...})` called by each screen/pane during init. Each action has: `ID`, `Title`, `Icon`, `Shortcut`, `Category`, `Run func() tea.Cmd`, `Visible func() bool`.

Phase 1 registers at least these 20:

- `new-session`, `stop-session`, `resume-session`, `kill-session`, `archive-session`
- `toggle-sidebar`, `toggle-inspector`, `toggle-thinking`, `toggle-compact`
- `switch-theme`, `focus-sidebar`, `focus-transcript`, `focus-input`
- `open-settings`, `open-help`, `open-logs`, `search-transcript`
- `copy-session-id`, `copy-last-output`, `quit`

### 7.7 Dialogs (modal)

Four types, all built on a shared `dialog.Model`:

1. **Alert** — title + message + `[ OK ]` — for errors.
2. **Confirm** — title + message + `[ Cancel ] [ Confirm ]` — for destructive actions.
3. **Question** — title + message + 1–9 numbered options (single-key selection) — for agent questions.
4. **Permission** — title + tool name + args + optional diff preview + `[1 Allow once] [2 Allow always] [3 Deny]`.

All modals dim the underlying view, trap focus, and return focus to the previous pane on close.

### 7.8 Toasts

Bottom-right stack, max 3 visible, auto-dismiss after `4s`. Slide-in from right (150ms), slide-out to right.

Types: `info` (blue), `success` (green), `warning` (yellow), `error` (red). Each shows an icon, a title, and an optional message. Click to dismiss.

---

## 8. Theme system

```go
type Theme struct {
    Name      string
    Bg        lipgloss.AdaptiveColor
    Fg        lipgloss.AdaptiveColor
    Dim       lipgloss.AdaptiveColor
    Accent    lipgloss.AdaptiveColor
    AccentAlt lipgloss.AdaptiveColor
    Success   lipgloss.AdaptiveColor
    Warn      lipgloss.AdaptiveColor
    Danger    lipgloss.AdaptiveColor
    Info      lipgloss.AdaptiveColor
    Border    lipgloss.AdaptiveColor
    Muted     lipgloss.AdaptiveColor
    // Status colors — one per session state
    StPending, StStarting, StRunning, StAwaitingInput,
    StAwaitingPerm, StIdle, StPaused, StCompleted,
    StFailed, StDisconnected lipgloss.AdaptiveColor
}
```

**Built-ins (phase 1):**

1. **`charm-dark`** — inspired by Charm's default (pinks, purples, bright accent).
2. **`tokyonight-storm`** — deep navy with cyan accents.
3. **`gruvbox-hard`** — warm earth tones.

Switching theme is instant (no restart) via `SetThemeMsg{Name}`.

`theme: "auto"` queries OSC 11, computes background luminance, picks the dark/light variant of the currently-selected theme.

All styles are created lazily per-theme and cached. Do **not** create styles in `View()` — expensive. Use a `Styles` struct populated once per theme change.

---

## 9. Keyboard bindings

Single source of truth: `internal/keys/keys.go`. Every binding is a `key.Binding` with Help.

### 9.1 Global (work from anywhere)

| Binding        | Action                                     |
| -------------- | ------------------------------------------ |
| `?`            | Toggle help overlay                        |
| `⌃p`           | Command palette                            |
| `⌃,`           | Open settings                              |
| `⌃c` then `⌃c` | Quit (double-press, 500ms window)          |
| `⌃b`           | Toggle sidebar                             |
| `⌃l`           | Clear current transcript view (not buffer) |
| `⌃n`           | New session (opens composer dialog)        |
| `⌃d`           | Toggle density compact/comfortable         |
| `⌃t`           | Cycle theme                                |

### 9.2 Sidebar focus

| Binding            | Action                         |
| ------------------ | ------------------------------ |
| `j`/`k` or `↓`/`↑` | Move                           |
| `g`/`G`            | Top / bottom                   |
| `/`                | Filter                         |
| `↵` or `l`         | Open session                   |
| `s`                | Stop                           |
| `r`                | Resume                         |
| `x`                | Kill (confirm)                 |
| `d`                | Archive (terminal states only) |
| `D`                | Delete (confirm)               |
| `tab`              | Focus transcript               |

### 9.3 Transcript focus

| Binding                     | Action                                        |
| --------------------------- | --------------------------------------------- |
| `j`/`k`, `⌃d`/`⌃u`, `g`/`G` | Scroll                                        |
| `/`                         | Search                                        |
| `n`/`N`                     | Next / prev match                             |
| `y`                         | Yank selection to clipboard                   |
| `o`                         | Open last-edited file in `$EDITOR`            |
| `space`                     | Toggle expand on tool-call block under cursor |
| `i`                         | Focus input                                   |
| `tab`                       | Focus sidebar                                 |

### 9.4 Input focus

| Binding         | Action                                           |
| --------------- | ------------------------------------------------ |
| `↵`             | Send                                             |
| `⌃j` or `alt+↵` | Newline                                          |
| `⌃e`            | Open input in `$EDITOR`                          |
| `⌃a`/`⌃e`       | Line start / end                                 |
| `⌃k`            | Kill to end of line                              |
| `⌃u`            | Kill to start of line                            |
| `⌃w`            | Kill previous word                               |
| `esc`           | Exit input focus                                 |
| `1`/`2`/`3`     | (only in `perm` mode) allow once / always / deny |

### 9.5 Horizontal focus (within button rows, chip groups)

| Binding           | Action                      |
| ----------------- | --------------------------- |
| `←`/`→`           | Move focus within row       |
| `tab`/`shift+tab` | Move focus to next/prev row |
| `↵` or `space`    | Activate focused button     |

### 9.6 Dev cheats (phase 1 only, `--dev` flag)

| Binding | Action                                         |
| ------- | ---------------------------------------------- |
| `⌃⌥1`   | Simulate WS connected                          |
| `⌃⌥2`   | Simulate WS reconnecting                       |
| `⌃⌥3`   | Simulate WS disconnected                       |
| `⌃⌥n`   | Inject a new mock session                      |
| `⌃⌥p`   | Inject a permission request on current session |
| `⌃⌥q`   | Inject a question on current session           |
| `⌃⌥f`   | Force-fail current session                     |
| `⌃⌥c`   | Force-complete current session                 |
| `⌃⌥t`   | Inject a toast of random severity              |

These make the demo interactive without wiring real events.

---

## 10. Session status model

```go
type Status int

const (
    StatusPending        Status = iota // queued, not started
    StatusStarting                     // spawning
    StatusRunning                      // agent working
    StatusAwaitingInput                // agent asked a question
    StatusAwaitingPerm                 // tool wants permission
    StatusIdle                         // alive but waiting
    StatusPaused                       // user-stopped, resumable
    StatusCompleted                    // task finished ok
    StatusFailed                       // error
    StatusDisconnected                 // PTY / ws lost
)
```

Each status has:

| Field      | Purpose                                                                      |
| ---------- | ---------------------------------------------------------------------------- |
| `Glyph`    | Single-cell dot from the §6.4.2 vocabulary — **no nerd-font / emoji icons** |
| `Color`    | via `Theme.St*`                                                              |
| `Label`    | human-readable lowercase label (e.g. `running`, `awaiting input`)            |
| `Pulse`    | bool — should the glyph pulse (awaiting + starting states only)              |
| `Severity` | int — descending sort weight for within-bucket sidebar ordering (§7.3.2)     |

Concrete glyph + severity assignments:

| Status             | Glyph | Pulse | Severity | Notes                                |
| ------------------ | ----- | ----- | -------- | ------------------------------------ |
| `StatusPending`    | `◐`   | yes   | 30       |                                      |
| `StatusStarting`   | `◐`   | yes   | 35       | may also use `⋯` spinner variant    |
| `StatusRunning`    | `●`   | no    | 40       |                                      |
| `StatusAwaitingInput` | `!` | yes   | 90       |                                      |
| `StatusAwaitingPerm`  | `!` | yes   | 100      | highest — demands immediate response |
| `StatusIdle`       | `○`   | no    | 20       |                                      |
| `StatusPaused`     | `○`   | no    | 15       |                                      |
| `StatusCompleted`  | `✓`   | no    | 5        |                                      |
| `StatusFailed`     | `×`   | no    | 80       |                                      |
| `StatusDisconnected` | `×` | no    | 70       |                                      |

The old `Priority` field (used for status-group ordering) is replaced by `Severity`. Sidebar no longer groups by status — it groups by time bucket and sorts within each bucket by severity descending (§7.3.2).

Pulse animation: 1 Hz sine, alpha 0.5 → 1.0. Use a global `tea.Tick(50ms)` that emits a `PulseTickMsg{Phase float64}` consumed by all pulsing components.

---

## 11. Animations & polish

Kept cheap and purposeful. Everything is driven by a single global `tea.Tick(50ms)` timer emitting `FrameMsg{TickN int64}`. Components read `TickN` and compute their frame without scheduling their own tickers.

> **Aesthetic cross-ref:** all visual chrome follows §6.4. In particular: no box borders on persistent panes, status glyphs from the §6.4.2 dot vocabulary, one accent color, body text in foreground ramp only.

1. **Splash logo shimmer** — hue shift, 8 fps (every 3rd frame).
2. **Pulse** (awaiting / starting states) — 1 Hz sine on foreground alpha. Applies to glyph `!` and `◐` only.
3. **Spinner** (starting state) — `bubbles/spinner` with custom glyphs `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`, 12 fps. Used as an alternative to `◐` when the session header shows starting state.
4. **Toast slide** — `x` translates from `width+3` to final position over 150ms, ease-out.
5. **Modal fade** — backdrop from `0` to `dim(0.4)` over 100ms. Applies only to modal overlays that have borders (§6.4.1).
6. **Sidebar collapse/expand** — `width 28 → 0` over 150ms, with the vertical `│` separator disappearing on the last frame.
7. **Filter bar slide-down** — `height 0 → 1` over 120ms (was 0→3 when the filter had a border box — now borderless, single line).
8. **Focus-change accent bar** — 1 frame flash bright, then settle.
9. **Footer calm rotation** — gentle 8s cycle between hint strings when idle (§7.3.6). No flicker — cross-fade by dimming the outgoing string over 2 frames, then brightening the incoming.

No easing library needed; write a tiny `ease.OutCubic(t float64) float64` helper.

**Absolutely forbidden:** per-character typing animations in the transcript. They feel novel once, then users beg for them to stop.

---

## 12. Architecture details

### 12.1 appModel and message routing

```go
type appModel struct {
    screen    Screen            // splash | onboarding | main | settings
    caps      ghostty.Caps
    theme     *theme.Theme
    styles    *Styles
    layout    LayoutRect
    bp        layout.Breakpoint
    focus     layout.FocusGraph
    focusID   layout.FocusID
    cfg       config.Config
    width, height int

    splash      splash.Model
    onboarding  onboarding.Model
    main        main.Model
    settings    settings.Model

    dialog   *dialog.Model // nil when none
    palette  *palette.Model
    help     *help.Model
    toasts   toast.Stack

    sessions session.Store       // in-memory
    mockEng  *mock.Engine        // phase 1 only
    devMode  bool
}
```

The root `Update` dispatches in this order:

1. Global `⌃c⌃c`, `?`, `⌃p`, `⌃,`, `⌃b` — handled here regardless of focus.
2. If a modal is open → route **only** to the modal. Modal can `return, ClearModalCmd` to close.
3. `tea.WindowSizeMsg` → recompute `layout`, propagate `SetSize` to children.
4. `tea.FocusMsg` / `tea.BlurMsg` → pause pulse timers on blur, resume on focus.
5. `FrameMsg` / `PulseTickMsg` → broadcast to all animated components.
6. Screen-scoped message → delegate to current screen's Update.

Return a single batched `tea.Cmd`.

### 12.2 Phase 2 seams

Define these interfaces in `internal/session/types.go`:

```go
type Engine interface {
    Start(ctx context.Context, task Task) (SessionID, error)
    Stop(id SessionID) error
    Kill(id SessionID) error
    Resume(id SessionID) error
    Respond(id SessionID, text string) error
    Permit(id SessionID, req PermissionReq, decision PermissionDecision) error
    Subscribe() <-chan Event // pubsub
}

type Transport interface {
    Run(ctx context.Context) error
    Send(msg ClientMsg) error
    Events() <-chan ServerMsg
    Status() TransportStatus
}
```

Phase 1 provides `mock.Engine` and `mock.Transport`. Phase 2 adds `pty.Engine` and `ws.Transport`. The view layer imports only the interfaces.

### 12.3 Message types

Define every message in `internal/app/messages.go`, typed. Examples:

```go
// Splash
type SplashTickMsg struct{ Frame int }
type SplashDoneMsg struct{}

// Framing
type FrameMsg struct{ TickN int64 }
type PulseTickMsg struct{ Phase float64 }

// Session lifecycle (from engine)
type SessionCreatedMsg   struct{ Session session.Session }
type SessionStatusMsg    struct{ ID session.ID; Status session.Status }
type SessionAppendMsg    struct{ ID session.ID; Content string }
type SessionQuestionMsg  struct{ ID session.ID; Q session.Question }
type SessionPermissionMsg struct{ ID session.ID; P session.PermissionReq }
type SessionDoneMsg      struct{ ID session.ID; Result session.Result }

// Transport
type TransportStatusMsg  struct{ Status session.TransportStatus }

// UI
type FocusChangeMsg      struct{ To layout.FocusID }
type ToastMsg            struct{ Toast toast.Toast }
type ClearModalMsg       struct{}
type SetThemeMsg         struct{ Name string }
type DevCheatMsg         struct{ Kind string }
```

Never use `interface{}` / `any` — type-switch in Update should cover every `case`.

### 12.4 Mock engine

`internal/session/mock/engine.go` is the phase 1 replacement for the real backend.

On startup, it:

1. Loads fixtures from `internal/fixtures/tasks.json` (see §14).
2. Creates one session per task.
3. Spawns a goroutine per session that replays its transcript from `internal/fixtures/transcripts/<id>.md` line-by-line, with scripted delays and state transitions encoded as special lines (`@@STATE: awaiting_perm {tool:"bash", args:"rm -rf /tmp/test"}`).
4. Sends events via `program.Send()` — **not** via a channel the model polls.

Parser syntax for transcript scripts:

```
@@STATE: running
@@DELAY: 500
Reading the repo structure…

@@DELAY: 1000
Found `src/auth/oauth.go`. Let me check the current imports.

@@STATE: awaiting_perm
@@PERM: {"tool": "bash", "args": "rm -rf /tmp/test", "diff": null}

@@RESUME_AFTER_PERM
OK, proceeding.

@@STATE: completed
```

This keeps fixtures declarative and lets phase 2 reuse them as integration test data.

---

## 13. Acceptance criteria (Definition of Done for phase 1)

The build is done when a reviewer, running `./daemonctl --dev` in Ghostty, can:

1. See the animated splash screen with shimmer, progress bar, and smooth fade-out.
2. Complete the full 11-step onboarding wizard, backtrack, and finish — resulting in a written `config.toml`.
3. Close and relaunch — the splash plays, then the main view loads directly (skips onboarding).
4. See at least **6 mock sessions** in the sidebar, grouped correctly by status, including one each of `running`, `awaiting_perm`, `awaiting_input`, `completed`, `failed`, `paused`.
5. Navigate the sidebar with `j/k`, open a session with `↵`, and watch the transcript stream in real time (mock replay).
6. Click `[⚙]` in the header (mouse) **and** press `⌃,` — both open settings.
7. Switch theme via palette (`⌃p → switch-theme`) — colors change instantly without flicker.
8. Trigger `⌃⌥p` to inject a permission request — the permission modal appears, is navigable with `1`/`2`/`3`, and correctly returns focus.
9. Resize the terminal from wide → normal → compact → tiny (< 60 cols): layout reflows without corruption at every size, and the `too small` screen appears below threshold.
10. Press `Tab` / `Shift+Tab` through every focusable element; `←`/`→` correctly move horizontally within button rows on the onboarding summary and the dialog button bars.
11. Press `⌃b` — sidebar collapses (in normal) or toggles overlay (in compact) with animation.
12. Press `?` anywhere — help overlay appears with all current-context bindings.
13. Press `y` in the transcript — content is copied to the system clipboard via OSC 52 (verified with `pbpaste` / `wl-paste`).
14. Trigger a session completion (dev cheat `⌃⌥c`) — a native Ghostty desktop notification fires via OSC 777.
15. Observe Ghostty's tab-bar progress indicator updating as sessions progress (OSC 9;4).
16. Run `TERM=xterm-256color ./daemonctl --dev` in a non-Ghostty terminal — everything still works, Ghostty-specific features silently degrade.
17. `go test ./...` passes, including `teatest` golden-file tests for splash, sidebar, main layout at all three breakpoints, and each dialog.

---

## 14. Mock fixtures

Put these under `internal/fixtures/`. They're what makes the demo feel alive; don't skimp.

### 14.1 `tasks.json`

Six tasks. Each matches one visible state on first load:

```json
[
	{
		"id": "sess_01HX3A",
		"title": "refactor auth module to use OAuth2 PKCE",
		"agent": "codex",
		"model": "gpt-5-sonnet",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "running",
		"tokens_used": 14234,
		"tokens_budget": 128000,
		"cost_usd": 0.12,
		"started_at": "-00:02:14",
		"transcript": "refactor-auth.md"
	},
	{
		"id": "sess_01HX3B",
		"title": "migrate database schema to v7",
		"agent": "opencode",
		"model": "gpt-5-sonnet",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "awaiting_perm",
		"tokens_used": 8921,
		"tokens_budget": 128000,
		"cost_usd": 0.08,
		"started_at": "-00:05:47",
		"transcript": "migrate-db.md"
	},
	{
		"id": "sess_01HX3C",
		"title": "add vitest coverage for retriever",
		"agent": "codex",
		"model": "gpt-5-mini",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "awaiting_input",
		"tokens_used": 4102,
		"tokens_budget": 128000,
		"cost_usd": 0.02,
		"started_at": "-00:01:30",
		"transcript": "add-tests.md"
	},
	{
		"id": "sess_01HW9D",
		"title": "fix lint errors in ingestion worker",
		"agent": "codex",
		"model": "gpt-5-mini",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "completed",
		"tokens_used": 2847,
		"tokens_budget": 128000,
		"cost_usd": 0.01,
		"started_at": "-00:12:00",
		"transcript": "lint-fix.md"
	},
	{
		"id": "sess_01HW9E",
		"title": "upgrade drizzle to 0.33",
		"agent": "opencode",
		"model": "gpt-5-sonnet",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "failed",
		"tokens_used": 17203,
		"tokens_budget": 128000,
		"cost_usd": 0.18,
		"started_at": "-00:34:11",
		"transcript": "upgrade-drizzle.md"
	},
	{
		"id": "sess_01HW9F",
		"title": "write CHANGELOG entries for v0.4",
		"agent": "codex",
		"model": "gpt-5-mini",
		"workdir": "/Users/demon/projects/rag-broker",
		"initial_status": "paused",
		"tokens_used": 1203,
		"tokens_budget": 128000,
		"cost_usd": 0.0,
		"started_at": "-01:02:00",
		"transcript": "changelog.md"
	}
]
```

### 14.2 Transcript fixtures

One per task under `internal/fixtures/transcripts/`. Each file uses the script syntax from §12.4.

Minimum content requirements per fixture:

- `refactor-auth.md` — ~40 lines, active streaming, 2 tool calls (`grep`, `edit`), a code diff block, finishes `running`.
- `migrate-db.md` — ~25 lines, ends at a permission request for `bash: rm migrations/*.sql.bak`.
- `add-tests.md` — ~20 lines, ends at a question: `Which test runner should I use? 1) vitest 2) jest 3) bun test`.
- `lint-fix.md` — complete transcript, 15 lines, final state `completed`.
- `upgrade-drizzle.md` — complete, ~30 lines, final state `failed` with a stack trace.
- `changelog.md` — 8 lines, final state `paused`.

Tasks also drive **live demo events** via a `scripts/demo.jsonl` that the dev-mode harness replays on a timer — e.g. at `t=30s` inject a new task, at `t=60s` flip `refactor-auth` to `completed`, at `t=90s` fire a toast. This gives reviewers a moving demo without keystrokes.

### 14.3 Mock notifications

Seed 3 toasts in the stack at startup so the demo immediately shows them stacked:

1. `info` — `Connected to wss://prod.example.com`.
2. `success` — `Session "refactor auth module" resumed`.
3. `warning` — `Session "migrate db" is waiting for your input`.

---

## 15. Implementation milestones

Suggested ordering — each milestone is an independently reviewable PR.

| #   | Milestone            | Key deliverables                                                                                   |
| --- | -------------------- | -------------------------------------------------------------------------------------------------- |
| M1  | Skeleton + layout    | `appModel`, WindowSizeMsg routing, breakpoint system, 3 placeholder screens, tests for layout math |
| M2  | Theme + styles       | `theme` pkg, 3 built-ins, live switching, OSC 11 auto detection                                    |
| M3  | Splash               | Animated splash with shimmer + Kitty graphics fallback                                             |
| M4  | Onboarding           | Full 11-step wizard with huh, state persistence, resume-on-restart                                 |
| M5  | Main layout (static) | Header, sidebar with mock items, transcript, input, footer — no interactivity yet                  |
| M6  | Focus system         | FocusGraph, horizontal focus, visual accent bar                                                    |
| M7  | Keyboard bindings    | All of §9 wired through, footer hint integration                                                   |
| M8  | Dialogs + palette    | Alert, confirm, question, permission, command palette with 20 actions                              |
| M9  | Mock engine          | Fixture loader, transcript script parser, live replay                                              |
| M10 | Toasts + animations  | Toast stack, pulse, spinner, slide-in/out                                                          |
| M11 | Ghostty caps         | OSC 9;4, OSC 52, OSC 8, OSC 777, Kitty keyboard                                                    |
| M12 | Settings page        | Full editable settings, save to TOML                                                               |
| M13 | Dev cheats           | `--dev` flag, all `⌃⌥` cheats, demo script runner                                                  |
| M14 | Polish + goldens     | teatest golden files at each breakpoint, regression tests                                          |

Total estimated effort: **~3–4 engineer-weeks**, depending on how much bike-shedding happens on colors.

---

## 16. Non-negotiables

These are easier to get right on day one than day 60:

1. **Every goroutine** spawned from the TUI process uses `defer recover()` posting a `PanicMsg{err, stack}`. A panic in a worker must never tear down the terminal without restoring it.
2. **`tea.LogToFile`** is used exclusively. No `fmt.Println`. No `slog` to stdout.
3. Styles are **never** constructed in `View()`. They live in a per-theme `Styles` struct built once.
4. `View()` must be O(visible) not O(total). Transcript uses a windowed viewport; sidebar uses the list bubble's built-in virtualization.
5. All messages are **typed**, declared in one file, and exhaustively handled in `Update` switches.
6. **No** direct mutation of child component state from the root. Always go through messages or explicit setters.
7. CJK-safe width using `runewidth.StringWidth`, not `len()`, for **every** truncation.
8. All truncation uses the `…` character (U+2026), never `...`.
9. No hardcoded ANSI escapes in application code — everything goes through lipgloss or `internal/ghostty/osc.go`.
10. Mouse events are **opt-in**: `--no-mouse` flag disables capture so users keep native terminal selection when they want it.

---

## 17. Open questions for product

Flag these to the product owner before starting M8:

1. On session completion, should the transcript auto-archive to a read-only state after N seconds, or stay interactive?
2. Permission decisions — do we need a global "remember this tool" store, or always ask per-session?
3. Cost display — roll up per-session or per-run? What counts a "run"?
4. Sidebar grouping — by status (current spec) or by project workdir? Make configurable?
5. Is there a plan for sharing a session to a teammate (read-only URL)? If yes, reserve a `[↗ share]` slot in the session header.

File answers in `docs/product-decisions.md` as they're resolved.

---

_End of spec. Questions or scope changes → ping in `#daemonctl` Slack before implementing._
