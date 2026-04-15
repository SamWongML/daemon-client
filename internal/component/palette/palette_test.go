package palette

import "testing"

func TestFilterSubstring(t *testing.T) {
	m := New([]Action{
		{ID: "quit", Title: "Quit"},
		{ID: "open-help", Title: "Open help"},
		{ID: "open-settings", Title: "Open settings"},
		{ID: "switch-theme", Title: "Switch theme"},
	})
	m.SetQuery("open")
	got := m.Filtered()
	if len(got) != 2 {
		t.Fatalf("want 2 matches, got %d (%v)", len(got), got)
	}
	// Both "Open help" and "Open settings" should match; the one starting at
	// position 0 comes first — either is fine, but Quit must not appear.
	for _, a := range got {
		if a.ID == "quit" {
			t.Fatalf("unexpected quit in results: %v", got)
		}
	}
}

func TestFilterFuzzy(t *testing.T) {
	m := New([]Action{
		{ID: "switch-theme", Title: "Switch theme"},
		{ID: "quit", Title: "Quit"},
	})
	m.SetQuery("sch") // substring 's' 'c' 'h' in "switch"
	got := m.Filtered()
	if len(got) == 0 || got[0].ID != "switch-theme" {
		t.Fatalf("want switch-theme first, got %v", got)
	}
}

func TestMoveClampsToFiltered(t *testing.T) {
	m := New([]Action{{ID: "a", Title: "A"}, {ID: "b", Title: "B"}})
	m.Move(10)
	if m.SelectedIndex() != 1 {
		t.Fatalf("want selected=1 after over-move, got %d", m.SelectedIndex())
	}
	m.Move(-5)
	if m.SelectedIndex() != 0 {
		t.Fatalf("want selected=0 after under-move, got %d", m.SelectedIndex())
	}
}

func TestSetQueryResetsInvalidSelection(t *testing.T) {
	m := New([]Action{{ID: "open-help", Title: "Open help"}, {ID: "quit", Title: "Quit"}})
	m.Move(1)
	m.SetQuery("open")
	if m.SelectedIndex() != 0 {
		t.Fatalf("want selected=0 after filter shrinks list, got %d", m.SelectedIndex())
	}
}
