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
	SessionHeaderH = 1
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

	sepCol := 0
	if sidebarW > 0 {
		sepCol = 1
	}

	inputH := clamp(inputLines+2, MinInputH, MaxInputH)
	mainX := sidebarW + sepCol
	mainW := w - sidebarW - sepCol

	// 2 full-width dim ─ rules (header↔body, body↔footer) consume 1 row each
	bodyH := h - HeaderH - 2 - FooterH
	if bodyH < 1 {
		bodyH = 1
	}

	l.Header = Rect{0, 0, w, HeaderH}
	l.Sidebar = Rect{0, HeaderH + 1, sidebarW, bodyH}
	l.SessionHeader = Rect{mainX, HeaderH + 1, mainW, SessionHeaderH}
	// 1 dim ─ rule between session header and transcript
	transcriptH := bodyH - SessionHeaderH - 1 - inputH
	if transcriptH < 1 {
		transcriptH = 1
	}
	l.Transcript = Rect{mainX, HeaderH + 1 + SessionHeaderH + 1, mainW, transcriptH}
	l.Input = Rect{mainX, HeaderH + 1 + SessionHeaderH + 1 + transcriptH, mainW, inputH}
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
