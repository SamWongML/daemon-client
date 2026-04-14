package keys

// Simple key binding table. Not using bubbles/key since we match on key strings.

type Binding struct {
	Keys []string
	Help string
}

func (b Binding) Matches(s string) bool {
	for _, k := range b.Keys {
		if k == s {
			return true
		}
	}
	return false
}

var (
	Quit         = Binding{[]string{"ctrl+c"}, "quit (double)"}
	Tab          = Binding{[]string{"tab"}, "next pane"}
	ShiftTab     = Binding{[]string{"shift+tab"}, "prev pane"}
	Up           = Binding{[]string{"up", "k"}, "up"}
	Down         = Binding{[]string{"down", "j"}, "down"}
	Top          = Binding{[]string{"g"}, "top"}
	Bottom       = Binding{[]string{"G"}, "bottom"}
	Enter        = Binding{[]string{"enter"}, "open"}
	QNonInput    = Binding{[]string{"q"}, "quit"}
	Help         = Binding{[]string{"?"}, "help"}
	Palette      = Binding{[]string{"ctrl+p"}, "palette"}
	ToggleSide   = Binding{[]string{"ctrl+b"}, "sidebar"}
	Send         = Binding{[]string{"enter"}, "send"}
	FocusInput   = Binding{[]string{"i"}, "input"}
	FocusSidebar = Binding{[]string{"tab"}, "sidebar"}
)
