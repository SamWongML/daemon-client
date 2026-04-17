package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/demon/daemon-client/internal/app"
	"github.com/demon/daemon-client/internal/ghostty"
	"github.com/demon/daemon-client/internal/session"
	"github.com/demon/daemon-client/internal/session/mock"
)

func main() {
	devMode := flag.Bool("dev", false, "enable dev cheats (⌃⌥{p,q,t,c,f}) and extra diagnostics")
	noMouse := flag.Bool("no-mouse", false, "disable mouse capture (keep native terminal selection)")
	themeName := flag.String("theme", "charm-dark", "initial theme name (charm-dark, charm-light, tokyonight-storm, gruvbox-hard)")
	resetConfig := flag.Bool("reset-config", false, "ignore existing config and re-run the onboarding wizard")
	flag.Parse()

	caps := ghostty.Detect()

	store := session.NewStore()
	engine := mock.New(store)
	if err := engine.LoadFixtures(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load fixtures: %v\n", err)
		os.Exit(1)
	}

	m := app.New(store, engine, caps, app.Options{
		DevMode:         *devMode,
		Theme:           *themeName,
		Mouse:           !*noMouse,
		ForceOnboarding: *resetConfig,
	})

	p := tea.NewProgram(m)
	engine.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
