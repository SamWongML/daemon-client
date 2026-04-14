package app

type FocusID string

const (
	FocusNone       FocusID = ""
	FocusSidebar    FocusID = "sidebar"
	FocusTranscript FocusID = "transcript"
	FocusInput      FocusID = "input"
)

// Next returns the next pane in the Tab cycle.
func (f FocusID) Next() FocusID {
	switch f {
	case FocusSidebar:
		return FocusTranscript
	case FocusTranscript:
		return FocusInput
	case FocusInput:
		return FocusSidebar
	}
	return FocusSidebar
}

func (f FocusID) Prev() FocusID {
	switch f {
	case FocusSidebar:
		return FocusInput
	case FocusTranscript:
		return FocusSidebar
	case FocusInput:
		return FocusTranscript
	}
	return FocusSidebar
}
