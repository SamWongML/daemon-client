package onboarding

// stepKind enumerates the widget shapes the wizard renders. The 11 steps map
// onto a subset of these — the enum stays open for phase 2.
type stepKind int

const (
	stepNote stepKind = iota
	stepInput
	stepPassword
	stepFilepath
	stepSelect
	stepMultiSelect
	stepSegmented
	stepAgents
	stepAdvanced
	stepSummary
)

type step struct {
	ID          string
	Kind        stepKind
	Title       string
	Subtitle    string
	Body        string   // for stepNote
	Placeholder string   // for inputs
	Hint        string   // secondary line under input
	Default     string   // fallback if the buffer is empty on commit
	Options     []string // first group (select / multi / segmented row 1)
	Options2    []string // second group (segmented row 2)
}

// The canonical 11-step list from spec §7.2.
var steps = []step{
	{
		ID:    "welcome",
		Kind:  stepNote,
		Title: "Welcome",
		Body:  "Let's get daemonctl set up. This takes about a minute —\nyou can go back at any step with esc or ←.",
	},
	{
		ID:          "server-url",
		Kind:        stepInput,
		Title:       "Server URL",
		Subtitle:    "Where does the daemon live? ws:// or wss:// only.",
		Placeholder: "wss://your-server.example.com/ws",
		Hint:        "tip: wss:// for TLS, ws:// for local dev",
	},
	{
		ID:          "auth-token",
		Kind:        stepPassword,
		Title:       "Auth token",
		Subtitle:    "Paste the token the server issued you.",
		Placeholder: "sk_xxxxxxxxxxxxxxxxxxxxx",
		Hint:        "press ⌘v / ⌃⇧v to paste — we'll store it under $XDG_DATA_HOME/daemonctl",
	},
	{
		ID:          "workdir",
		Kind:        stepFilepath,
		Title:       "Default working directory",
		Subtitle:    "Where new sessions should start. Defaults to your home dir.",
		Placeholder: "~/projects/…",
		Hint:        "relative paths are resolved against $PWD",
	},
	{
		ID:       "agents",
		Kind:     stepAgents,
		Title:    "Agent binaries",
		Subtitle: "Phase 1 demo: both agents are mocked as found. Phase 2 probes $PATH.",
	},
	{
		ID:       "max-sessions",
		Kind:     stepSelect,
		Title:    "Max concurrent sessions",
		Subtitle: "How many sessions can run at once?",
		Options:  []string{"1", "2", "4", "8"},
	},
	{
		ID:       "default-agent-model",
		Kind:     stepSegmented,
		Title:    "Default agent + model",
		Subtitle: "Two dependent choices — the model list changes when you pick an agent.",
		Options:  []string{"codex", "opencode"},
		// Options2 is computed at render time via modelsFor(Cfg.Agent).
	},
	{
		ID:       "notifications",
		Kind:     stepMultiSelect,
		Title:    "Notifications",
		Subtitle: "When should daemonctl poke you?",
		Options: []string{
			"Terminal bell on attention required",
			"Desktop notification on attention required",
			"Desktop notification on session completion",
			"Desktop notification on session failure",
		},
	},
	{
		ID:       "appearance",
		Kind:     stepSegmented,
		Title:    "Appearance",
		Subtitle: "Theme and density. Both can be changed later in settings.",
		Options:  []string{"auto", "charm-dark", "charm-light", "tokyonight-storm", "gruvbox-hard"},
		Options2: []string{"comfortable", "compact"},
	},
	{
		ID:       "advanced",
		Kind:     stepAdvanced,
		Title:    "Advanced",
		Subtitle: "Logging & telemetry — safe defaults are fine for most people.",
	},
	{
		ID:       "summary",
		Kind:     stepSummary,
		Title:    "Review",
		Subtitle: "Here's what we'll write to config.toml.",
	},
}
