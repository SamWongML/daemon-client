# Daemon Client TUI — Implementation Todo

Progress tracker against [tui-spec.md](tui-spec.md). Milestones (§15) are the top-level buckets; nested items are the concrete deliverables.

## Milestones

- [x] **M1 — Skeleton + layout**
  - [x] `appModel` root + router ([internal/app/app.go](../internal/app/app.go))
  - [x] Breakpoint system ([internal/layout/breakpoint.go](../internal/layout/breakpoint.go))
  - [x] `LayoutRect` + size propagation ([internal/layout/rect.go](../internal/layout/rect.go))
  - [x] Placeholder screens wired through router

- [x] **M2 — Theme + styles**
  - [x] `Theme` struct + `Styles` cache ([internal/theme/theme.go](../internal/theme/theme.go))
  - [x] Built-ins: charm-dark, charm-light, tokyonight-storm, gruvbox-hard
  - [x] Live `SetThemeMsg` switching
  - [ ] `theme: "auto"` via OSC 11 background-luminance detection

- [x] **M3 — Splash**
  - [x] Animated ASCII logo + progress bar ([internal/component/splash/splash.go](../internal/component/splash/splash.go))
  - [x] 150ms fade-out → `SplashDoneMsg`
  - [ ] Kitty-graphics PNG logo path (fallback wired, real PNG not loaded)

- [x] **M4 — Onboarding**
  - [x] 11-step huh wizard ([internal/component/onboarding/steps.go](../internal/component/onboarding/steps.go))
  - [x] Config persistence (TOML via xdg)
  - [x] First-run detection on splash-done
  - [ ] Resume-on-restart (`$XDG_STATE_HOME/daemonctl/onboarding.toml`)

- [x] **M5 — Main layout (static)**
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

- [x] **M8 — Dialogs + palette**
  - [x] Alert/Confirm/Question/Permission dialog ([internal/component/dialog/dialog.go](../internal/component/dialog/dialog.go))
  - [x] Command palette ([internal/component/palette/palette.go](../internal/component/palette/palette.go))
  - [ ] Verify all 20 palette actions registered (§7.6)

- [x] **M9 — Mock engine**
  - [x] Fixture loader + script parser ([internal/session/mock/engine.go](../internal/session/mock/engine.go))
  - [x] 6 seed tasks + transcripts ([internal/fixtures/](../internal/fixtures/))
  - [ ] `scripts/demo.jsonl` timed event replay

- [~] **M10 — Toasts + animations**
  - [x] Toast stack ([internal/component/toast/toast.go](../internal/component/toast/toast.go))
  - [x] Seed 3 startup toasts (§14.3)
  - [ ] Global `FrameMsg` / `PulseTickMsg` tick driving pulse + spinner
  - [ ] Slide-in/out, modal fade, sidebar collapse animations

- [ ] **M11 — Ghostty caps**
  - [ ] `internal/ghostty/caps.go` capability detection
  - [ ] OSC 9;4 aggregate progress
  - [ ] OSC 52 clipboard yank
  - [ ] OSC 8 hyperlinks in transcript
  - [ ] OSC 777 desktop notifications
  - [ ] Kitty keyboard protocol push/pop

- [x] **M12 — Settings page**
  - [x] Settings screen + categories ([internal/component/settings/settings.go](../internal/component/settings/settings.go))
  - [ ] Verify all 8 categories editable + `⌃s` save round-trips to TOML

- [~] **M13 — Dev cheats**
  - [x] `--dev` flag gating ([cmd/daemonctl/main.go](../cmd/daemonctl/main.go))
  - [ ] Verify all `⌃⌥{1,2,3,n,p,q,f,c,t}` cheats implemented per §9.6
  - [ ] `--no-mouse` flag wired into program options

- [ ] **M14 — Polish + goldens**
  - [ ] `teatest` golden files: splash, sidebar, main layout × 3 breakpoints, each dialog
  - [ ] `too small` screen (< 60×16)
  - [ ] `defer recover()` → `PanicMsg` on every spawned goroutine
  - [ ] `tea.LogToFile` only; no stdout writes
  - [ ] `runewidth.StringWidth` for all truncation; `…` (U+2026) only

## Acceptance (§13)

Run `./daemonctl --dev` in Ghostty and walk the 17-item checklist in the spec before declaring phase 1 done.
