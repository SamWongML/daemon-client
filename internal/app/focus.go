package app

type FocusID string

const (
	FocusNone           FocusID = ""
	FocusSidebar        FocusID = "sidebar"
	FocusTranscript     FocusID = "transcript"
	FocusInput          FocusID = "input"
	FocusHeaderSettings FocusID = "header.settings"
	FocusHeaderHelp     FocusID = "header.help"
)

// FocusNode describes one stop in the focus graph. Next/Prev wire the
// Tab/Shift+Tab cycle (rows); Left/Right wire horizontal movement within a
// row of buttons. Empty fields mean "no movement in that direction".
type FocusNode struct {
	ID         FocusID
	Next, Prev FocusID
	Left       FocusID
	Right      FocusID
}

// FocusGraph is an addressable map of focus nodes plus an entry ID used when
// a screen first mounts.
type FocusGraph struct {
	Entry FocusID
	Nodes map[FocusID]FocusNode
}

func (g FocusGraph) node(id FocusID) (FocusNode, bool) {
	n, ok := g.Nodes[id]
	return n, ok
}

// move returns dst if non-empty, otherwise the current id (no-op).
func move(cur, dst FocusID) FocusID {
	if dst == "" {
		return cur
	}
	return dst
}

func (g FocusGraph) Next(id FocusID) FocusID {
	if n, ok := g.node(id); ok {
		return move(id, n.Next)
	}
	return g.Entry
}

func (g FocusGraph) Prev(id FocusID) FocusID {
	if n, ok := g.node(id); ok {
		return move(id, n.Prev)
	}
	return g.Entry
}

func (g FocusGraph) Left(id FocusID) FocusID {
	if n, ok := g.node(id); ok {
		return move(id, n.Left)
	}
	return id
}

func (g FocusGraph) Right(id FocusID) FocusID {
	if n, ok := g.node(id); ok {
		return move(id, n.Right)
	}
	return id
}

// IsHeaderButton reports whether the id is one of the header chip buttons —
// used so the main handler can route Enter to the button's action.
func (f FocusID) IsHeaderButton() bool {
	return f == FocusHeaderSettings || f == FocusHeaderHelp
}

// mainFocusGraph is the graph for screenMain. Tab cycles the three panes and
// then lands on the header buttons; ←/→ moves within the header button row.
func mainFocusGraph() FocusGraph {
	return FocusGraph{
		Entry: FocusSidebar,
		Nodes: map[FocusID]FocusNode{
			FocusSidebar: {
				ID: FocusSidebar, Next: FocusTranscript, Prev: FocusHeaderHelp,
			},
			FocusTranscript: {
				ID: FocusTranscript, Next: FocusInput, Prev: FocusSidebar,
			},
			FocusInput: {
				ID: FocusInput, Next: FocusHeaderSettings, Prev: FocusTranscript,
			},
			FocusHeaderSettings: {
				ID: FocusHeaderSettings, Next: FocusHeaderHelp, Prev: FocusInput,
				Right: FocusHeaderHelp,
			},
			FocusHeaderHelp: {
				ID: FocusHeaderHelp, Next: FocusSidebar, Prev: FocusHeaderSettings,
				Left: FocusHeaderSettings,
			},
		},
	}
}
