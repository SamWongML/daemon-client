# Daemon Client TUI — Implementation Todo

Progress tracker against [tui-spec.md](tui-spec.md). Milestones (§15) are the top-level buckets; nested items are the concrete deliverables.

Legend: `[x]` done · `[~]` partial · `[ ]` not started.

## Milestones

- [ ] **M0 — Aesthetic refresh** ⚠️ URGENT — do this before any other milestone work
  > Implements spec §6.4 (visual language), revised §7.3.1–7.3.6, and revised §10.
  > Rationale: the current build is too busy — boxed borders, nerd-font icons, status-grouped sidebar, and a dense header all fight for attention. This pass strips the chrome to match modern Charm/opencode aesthetics (dot glyphs, left-accent-bar selection, time-bucketed sidebar, sparse header). See spec for design references.

  - [ ] **M0.1 — Strip borders** ([internal/component/header/header.go](../internal/component/header/header.go), [sidebar.go](../internal/component/sidebar/sidebar.go), [footer.go](../internal/component/footer/footer.go), [input.go](../internal/component/input/input.go), [transcript.go](../internal/component/transcript/transcript.go))
    - [ ] Remove all `lipgloss.RoundedBorder()` / `NormalBorder()` from persistent panes
    - [ ] Add single vertical `│` separator between sidebar and main pane (dim `theme.Border`)
    - [ ] Add dim `─` horizontal rules: header↔body, body↔footer, session-header↔transcript
    - [ ] Keep `RoundedBorder()` only on: command palette, help overlay, dialog modals
  - [ ] **M0.2 — Status glyph swap** ([internal/session/types.go](../internal/session/types.go) or equivalent)
    - [ ] Replace all nerd-font / emoji status glyphs with §6.4.2 dot vocabulary (`●` `◐` `○` `✓` `×` `!` `⋯`)
    - [ ] Add `Severity int` field to status model for sidebar sort order
    - [ ] Remove old `Priority` field (was used for status-group ordering)
  - [ ] **M0.3 — Sidebar redesign** ([internal/component/sidebar/sidebar.go](../internal/component/sidebar/sidebar.go))
    - [ ] Replace status-group ordering with time-bucket ordering (`active > today > this week > older > archive`)
    - [ ] Implement severity-descending secondary sort within each bucket
    - [ ] Reduce items to 2-line comfortable / 1-line compact (drop old 3-line row with activity text)
    - [ ] Replace boxed `+ New session` with borderless accent-bar row
    - [ ] Replace boxed filter bar with single-line `/` prefix row (no border)
    - [ ] Attention-state rows (`!` glyph) get persistent left accent bar in `theme.Warn`
    - [ ] Archive bucket rows render at ~40% opacity
  - [ ] **M0.4 — Header simplification** ([internal/component/header/header.go](../internal/component/header/header.go))
    - [ ] Strip header to: `▌daemonctl   ● agent · model` (left) + `[⚙] [?]` (right)
    - [ ] Remove server URL, session count, clock, cost from header
  - [ ] **M0.5 — Footer status line** ([internal/component/footer/footer.go](../internal/component/footer/footer.go))
    - [ ] Add right-zone to footer: session count, aggregate cost, clock (HH:MM)
    - [ ] Implement calm-state hint rotation (8s cycle when idle)
    - [ ] Right-zone items render in `theme.Dim`, drop right-to-left on overflow
  - [ ] **M0.6 — Session header collapse** ([internal/component/transcript/transcript.go](../internal/component/transcript/transcript.go) or session header component)
    - [ ] Collapse session header from 2 lines to 1 line per revised §7.3.3
    - [ ] Move token bar inline, 12-col `WithSolidFill`, color-ramped by usage %
  - [ ] **M0.7 — Layout math update** ([internal/layout/rect.go](../internal/layout/rect.go))
    - [ ] Update `LayoutRect` to use `sessionHeaderH = 1` (was 2)
    - [ ] Account for `mainX = sidebarW + 1` (separator column)
  - [ ] **M0.8 — Color discipline audit** ([internal/theme/theme.go](../internal/theme/theme.go), all component `View()` methods)
    - [ ] Verify each theme has exactly one accent hue
    - [ ] Audit all `View()` methods: body text must use only `Fg` / `Muted` / `Dim` ramp — no accent on body text
    - [ ] Status colors (`Success`, `Warn`, `Danger`, `Info`) appear only on glyphs + dialog severity

- [x] **M1 — Skeleton + layout**
  - [x] `appModel` root + router ([internal/app/app.go](../internal/app/app.go))
  - [x] Breakpoint system ([internal/layout/breakpoint.go](../internal/layout/breakpoint.go))
  - [x] `LayoutRect` + size propagation ([internal/layout/rect.go](../internal/layout/rect.go))
  - [x] Placeholder screens wired through router

- [~] **M2 — Theme + styles**
  - [x] `Theme` struct + `Styles` cache ([internal/theme/theme.go](../internal/theme/theme.go))
  - [x] Built-ins: charm-dark, charm-light, tokyonight-storm, gruvbox-hard
  - [x] Live `SetThemeMsg` switching
  - [ ] `theme: "auto"` via OSC 11 background-luminance detection

- [~] **M3 — Splash**
  - [x] Animated ASCII logo + progress bar ([internal/component/splash/splash.go](../internal/component/splash/splash.go))
  - [x] 150ms fade-out → `SplashDoneMsg`
  - [ ] Kitty-graphics PNG logo path (fallback wired, real PNG not loaded)

- [~] **M4 — Onboarding**
  - [x] 11-step huh wizard ([internal/component/onboarding/steps.go](../internal/component/onboarding/steps.go))
  - [x] Config persistence (TOML via xdg)
  - [x] First-run detection on splash-done
  - [ ] Resume-on-restart (`$XDG_STATE_HOME/daemonctl/onboarding.toml`)

- [~] **M5 — Main layout (static)** ← M0 will modify all of these
  - [x] Header ([internal/component/header/header.go](../internal/component/header/header.go))
  - [x] Sidebar w/ mock items ([internal/component/sidebar/sidebar.go](../internal/component/sidebar/sidebar.go))
  - [x] Transcript viewport ([internal/component/transcript/transcript.go](../internal/component/transcript/transcript.go))
  - [x] Input textarea ([internal/component/input/input.go](../internal/component/input/input.go))
  - [x] Footer hints ([internal/component/footer/footer.go](../internal/component/footer/footer.go))

- [x] **M6 — Focus system**
  - [x] `FocusGraph` + traversal ([internal/app/focus.go](../internal/app/focus.go))
  - [x] Horizontal focus within button rows
  - [x] Accent-bar visual feedback

- [~] **M7 — Keyboard bindings**
  - [x] Central bindings table ([internal/keys/keys.go](../internal/keys/keys.go))
  - [x] Global bindings (`?`, `⌃p`, `⌃,`, `⌃b`, `⌃c⌃c`, `⌃n`, `⌃d`, `⌃t`, `⌃l`)
  - [ ] Full sidebar/transcript/input binding coverage per §9.2–9.4 (verify `o` $EDITOR, `⌃e` edit-in-editor, `space` toggle tool-call, `/n/N` search, `y` yank)
  - [ ] Footer hints derived as union of global+screen+pane bindings

- [~] **M8 — Dialogs + palette**
  - [x] Alert/Confirm/Question/Permission dialog ([internal/component/dialog/dialog.go](../internal/component/dialog/dialog.go))
  - [x] Command palette ([internal/component/palette/palette.go](../internal/component/palette/palette.go))
  - [ ] Verify all 20 palette actions registered (§7.6)

- [~] **M9 — Mock engine**
  - [x] Fixture loader + script parser ([internal/session/mock/engine.go](../internal/session/mock/engine.go))
  - [x] 6 seed tasks + transcripts ([internal/fixtures/](../internal/fixtures/))
  - [ ] `scripts/demo.jsonl` timed event replay

- [~] **M10 — Toasts + animations**
  - [x] Toast stack ([internal/component/toast/toast.go](../internal/component/toast/toast.go))
  - [x] Seed 3 startup toasts (§14.3)
  - [ ] Global `FrameMsg` / `PulseTickMsg` tick driving pulse + spinner
  - [ ] Slide-in/out, modal fade, sidebar collapse, footer calm-rotation animations (revised §11)

- [ ] **M11 — Ghostty caps**
  - [ ] `internal/ghostty/caps.go` capability detection
  - [ ] OSC 9;4 aggregate progress
  - [ ] OSC 52 clipboard yank
  - [ ] OSC 8 hyperlinks in transcript
  - [ ] OSC 777 desktop notifications
  - [ ] Kitty keyboard protocol push/pop

- [~] **M12 — Settings page**
  - [x] Settings screen + categories ([internal/component/settings/settings.go](../internal/component/settings/settings.go))
  - [ ] Verify all 8 categories editable + `⌃s` save round-trips to TOML

- [~] **M13 — Dev cheats**
  - [x] `--dev` flag gating ([cmd/daemonctl/main.go](../cmd/daemonctl/main.go))
  - [ ] Verify all `⌃⌥{1,2,3,n,p,q,f,c,t}` cheats implemented per §9.6
  - [x] `--no-mouse` flag

- [ ] **M14 — Polish + goldens**
  - [ ] `teatest` golden files: splash, sidebar, main layout × 3 breakpoints, each dialog
  - [ ] `too small` screen (< 60×16)
  - [ ] `defer recover()` → `PanicMsg` on every spawned goroutine
  - [ ] `tea.LogToFile` only; no stdout writes
  - [ ] `runewidth.StringWidth` for all truncation; `…` (U+2026) only

## Acceptance (§13)

Run `./daemonctl --dev` in Ghostty and walk the 17-item checklist in the spec before declaring phase 1 done.
