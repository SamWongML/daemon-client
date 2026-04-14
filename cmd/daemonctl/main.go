package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/demon/daemon-client/internal/app"
	"github.com/demon/daemon-client/internal/session"
	"github.com/demon/daemon-client/internal/session/mock"
)

func main() {
	store := session.NewStore()
	engine := mock.New(store)
	if err := engine.LoadFixtures(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load fixtures: %v\n", err)
		os.Exit(1)
	}

	m := app.New(store, engine)
	p := tea.NewProgram(m)
	engine.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
