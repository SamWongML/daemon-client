package layout

type Breakpoint int

const (
	Compact Breakpoint = iota
	Normal
	Wide
)

func For(w, h int) Breakpoint {
	if w < 100 || h < 24 {
		return Compact
	}
	if w < 160 {
		return Normal
	}
	return Wide
}

func (b Breakpoint) String() string {
	switch b {
	case Compact:
		return "compact"
	case Normal:
		return "normal"
	case Wide:
		return "wide"
	}
	return "?"
}
