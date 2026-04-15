package app

import "testing"

func TestMainFocusGraphTabCycle(t *testing.T) {
	g := mainFocusGraph()
	order := []FocusID{
		FocusSidebar, FocusTranscript, FocusInput,
		FocusHeaderSettings, FocusHeaderHelp, FocusSidebar,
	}
	cur := order[0]
	for i := 1; i < len(order); i++ {
		cur = g.Next(cur)
		if cur != order[i] {
			t.Fatalf("Next step %d: got %q want %q", i, cur, order[i])
		}
	}
	// Shift+Tab walks the cycle backwards.
	for i := len(order) - 2; i >= 0; i-- {
		cur = g.Prev(cur)
		if cur != order[i] {
			t.Fatalf("Prev step %d: got %q want %q", i, cur, order[i])
		}
	}
}

func TestHeaderHorizontalFocus(t *testing.T) {
	g := mainFocusGraph()
	if got := g.Right(FocusHeaderSettings); got != FocusHeaderHelp {
		t.Fatalf("Right(settings) = %q; want help", got)
	}
	if got := g.Left(FocusHeaderHelp); got != FocusHeaderSettings {
		t.Fatalf("Left(help) = %q; want settings", got)
	}
	// No horizontal movement defined for panes — should stay put.
	if got := g.Left(FocusSidebar); got != FocusSidebar {
		t.Fatalf("Left(sidebar) = %q; want unchanged", got)
	}
	if got := g.Right(FocusTranscript); got != FocusTranscript {
		t.Fatalf("Right(transcript) = %q; want unchanged", got)
	}
}

func TestIsHeaderButton(t *testing.T) {
	if !FocusHeaderSettings.IsHeaderButton() || !FocusHeaderHelp.IsHeaderButton() {
		t.Fatal("header buttons should classify as header buttons")
	}
	if FocusSidebar.IsHeaderButton() {
		t.Fatal("sidebar is not a header button")
	}
}
