package dialog

import "testing"

func TestEscCancels(t *testing.T) {
	d := Confirm("Delete?", "This cannot be undone.")
	got, done := d.HandleKey("esc")
	if !done {
		t.Fatalf("esc should close the dialog")
	}
	if got != ResultCancelled {
		t.Fatalf("esc should return ResultCancelled, got %d", got)
	}
}

func TestConfirmDefaultFocus(t *testing.T) {
	d := Confirm("Delete?", "yep?")
	// Default focus is on Confirm (index 1) so pressing enter acts safely.
	if d.Selected != 1 {
		t.Fatalf("want default selection=1 (Confirm), got %d", d.Selected)
	}
	got, done := d.HandleKey("enter")
	if !done || got != 1 {
		t.Fatalf("enter should confirm (1,true), got (%d,%v)", got, done)
	}
}

func TestQuestionNumberKey(t *testing.T) {
	d := Question("Pick one", "?", []string{"a", "b", "c"})
	got, done := d.HandleKey("2")
	if !done || got != 1 {
		t.Fatalf("press 2 should select index 1 and close: got (%d,%v)", got, done)
	}
}

func TestPermissionArrowsMove(t *testing.T) {
	d := Permission("bash", "rm -rf /", "")
	if d.Selected != 0 {
		t.Fatalf("want initial selected=0, got %d", d.Selected)
	}
	_, done := d.HandleKey("right")
	if done {
		t.Fatalf("right arrow should not close dialog")
	}
	if d.Selected != 1 {
		t.Fatalf("want selected=1 after right, got %d", d.Selected)
	}
	got, done := d.HandleKey("3")
	if !done || got != 2 {
		t.Fatalf("press 3 should deny (2,true), got (%d,%v)", got, done)
	}
}

func TestQuestionClampsTo9Options(t *testing.T) {
	opts := make([]string, 15)
	for i := range opts {
		opts[i] = "x"
	}
	d := Question("t", "?", opts)
	if len(d.Options) != 9 {
		t.Fatalf("question options should be clamped to 9, got %d", len(d.Options))
	}
}
