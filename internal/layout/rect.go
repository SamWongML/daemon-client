package layout

type Rect struct{ X, Y, W, H int }

type Layout struct {
	Breakpoint       Breakpoint
	Width, Height    int
	Header           Rect
	Sidebar          Rect
	SessionHeader    Rect
	Transcript       Rect
	Input            Rect
	Footer           Rect
	SidebarCollapsed bool
	TooSmall         bool
}

const (
	HeaderH        = 1
	FooterH        = 1
	SessionHeaderH = 2
	MinInputH      = 3
	MaxInputH      = 10
)

func Compute(w, h int, sidebarCollapsed bool, inputLines int) Layout {
	bp := For(w, h)
	l := Layout{Breakpoint: bp, Width: w, Height: h, SidebarCollapsed: sidebarCollapsed}
	if w < 60 || h < 16 {
		l.TooSmall = true
		return l
	}
	sidebarW := 28
	if sidebarCollapsed || bp == Compact {
		sidebarW = 0
	} else if bp == Wide {
		sidebarW = min(36, w*22/100)
	}

	inputH := clamp(inputLines+2, MinInputH, MaxInputH)
	mainX := sidebarW
	mainW := w - sidebarW

	l.Header = Rect{0, 0, w, HeaderH}
	l.Sidebar = Rect{0, HeaderH, sidebarW, h - HeaderH - FooterH}
	l.SessionHeader = Rect{mainX, HeaderH, mainW, SessionHeaderH}
	transcriptH := h - HeaderH - FooterH - SessionHeaderH - inputH
	if transcriptH < 1 {
		transcriptH = 1
	}
	l.Transcript = Rect{mainX, HeaderH + SessionHeaderH, mainW, transcriptH}
	l.Input = Rect{mainX, HeaderH + SessionHeaderH + transcriptH, mainW, inputH}
	l.Footer = Rect{0, h - FooterH, w, FooterH}
	return l
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
