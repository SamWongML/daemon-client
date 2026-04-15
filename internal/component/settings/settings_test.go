package settings

import (
	"strings"
	"testing"

	"github.com/demon/daemon-client/internal/theme"
)

func TestCategoriesCoverSpec(t *testing.T) {
	want := []string{
		"connection", "workspace", "agents", "appearance",
		"notifications", "keyboard", "advanced", "about",
	}
	if len(Categories) != len(want) {
		t.Fatalf("got %d categories, want %d", len(Categories), len(want))
	}
	for i, w := range want {
		if Categories[i].ID != w {
			t.Errorf("Categories[%d].ID = %q, want %q", i, Categories[i].ID, w)
		}
	}
}

func TestFieldsPopulatedForEveryCategory(t *testing.T) {
	s := State{
		Theme:     theme.Dark(),
		ServerURL: "wss://example.test",
		Workdir:   "/tmp/proj",
		MaxSess:   4,
		Agent:     "codex",
		Model:     "gpt-5-sonnet",
		LogLevel:  "info",
		LogFile:   "/tmp/log",
		Version:   "v0.1.0",
		Build:     "test",
	}
	for _, c := range Categories {
		got := s.Fields(c.ID)
		if len(got) == 0 {
			t.Errorf("category %q has no fields", c.ID)
		}
	}
}

func TestRenderContainsActiveTheme(t *testing.T) {
	th := theme.TokyoNightStorm()
	s := State{Selected: 3, Theme: th, Version: "v0.1.0"} // 3 = appearance
	out := Render(th, theme.BuildStyles(th), s, 120, 30)
	if !strings.Contains(out, th.Label) {
		t.Errorf("rendered output does not mention active theme label %q", th.Label)
	}
	// Selected category title must appear.
	if !strings.Contains(out, Categories[3].Title) {
		t.Errorf("rendered output missing category title %q", Categories[3].Title)
	}
}

func TestRenderTooSmall(t *testing.T) {
	th := theme.Dark()
	out := Render(th, theme.BuildStyles(th), State{Theme: th}, 40, 10)
	if !strings.Contains(out, "Too small") {
		t.Errorf("expected too-small message, got: %q", out)
	}
}
