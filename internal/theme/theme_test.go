package theme

import "testing"

func TestRegistryDefaults(t *testing.T) {
	all := Registry()
	if len(all) < 3 {
		t.Fatalf("expected at least 3 themes, got %d", len(all))
	}

	// First theme must be a dark theme (spec: default is charm-dark).
	if all[0].Name != "charm-dark" {
		t.Errorf("registry[0] = %q, want %q", all[0].Name, "charm-dark")
	}

	seen := map[string]bool{}
	for _, th := range all {
		if seen[th.Name] {
			t.Errorf("duplicate theme name: %q", th.Name)
		}
		seen[th.Name] = true
		if th.Label == "" {
			t.Errorf("theme %q missing Label", th.Name)
		}
	}
}

func TestByNameFallback(t *testing.T) {
	got := ByName("not-a-theme")
	if got == nil {
		t.Fatal("ByName returned nil for unknown theme")
	}
	if got.Name != Registry()[0].Name {
		t.Errorf("fallback = %q, want default %q", got.Name, Registry()[0].Name)
	}
	// Known names resolve correctly.
	for _, want := range []string{"charm-dark", "charm-light", "tokyonight-storm", "gruvbox-hard"} {
		if got := ByName(want); got.Name != want {
			t.Errorf("ByName(%q).Name = %q", want, got.Name)
		}
	}
}

func TestNextWraps(t *testing.T) {
	all := Registry()
	start := all[0].Name
	cur := start
	for i := 0; i < len(all); i++ {
		cur = Next(cur).Name
	}
	if cur != start {
		t.Errorf("Next did not wrap: went %q → %q after %d steps", start, cur, len(all))
	}
}

func TestBuildStylesCovers(t *testing.T) {
	// Make sure BuildStyles returns a populated struct for every theme so a
	// live switch doesn't return zero values that render invisible text.
	for _, th := range Registry() {
		s := BuildStyles(th)
		if s == nil {
			t.Fatalf("BuildStyles returned nil for %q", th.Name)
		}
		// Spot-check a couple of settings-specific styles that are new.
		if s.SettingsTitle.GetForeground() == nil {
			t.Errorf("%q: SettingsTitle has no foreground", th.Name)
		}
	}
}

func TestStatusColorCoverage(t *testing.T) {
	th := Dark()
	for _, name := range []string{
		"pending", "starting", "running", "awaiting_input",
		"awaiting_perm", "idle", "paused", "completed",
		"failed", "disconnected",
	} {
		if StatusColor(th, name) == nil {
			t.Errorf("StatusColor(%q) returned nil", name)
		}
	}
	// Unknown name falls back to Fg, not nil.
	if StatusColor(th, "nonesuch") == nil {
		t.Errorf("StatusColor fallback returned nil")
	}
}
