// Package config handles loading and saving the daemonctl user config.
//
// Phase 1 writes a minimal, flat TOML file at
// $XDG_CONFIG_HOME/daemonctl/config.toml (falling back to ~/.config/...).
// Onboarding (§7.2) is the only producer; settings (§7.4) is read-only for
// now. Phase 2 will add live editing + schema validation.
//
// We keep the on-disk format narrow and hand-roll the encoder so we don't
// drag in a TOML library. The shape is strictly key/value + a `[notifications]`
// table — if it grows beyond that, swap in BurntSushi/toml.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config captures every setting the onboarding wizard collects. Field names
// are the TOML keys (lower_snake).
type Config struct {
	ServerURL   string
	AuthToken   string
	Workdir     string
	AgentCodex  string // path to codex binary
	AgentOC     string // path to opencode binary
	MaxSessions int
	Agent       string // default agent: codex | opencode
	Model       string // default model
	Theme       string // charm-dark | charm-light | tokyonight-storm | gruvbox-hard | auto
	Density     string // comfortable | compact
	LogLevel    string // off | error | info | debug
	LogFile     string

	Notifications Notifications
	Telemetry     bool

	// UsedDefaults is true when Default() built this — used by settings as a
	// cue that onboarding never ran. Not persisted.
	UsedDefaults bool `toml:"-"`
}

// Notifications is the flags struct from onboarding step 8.
type Notifications struct {
	BellOnAttention    bool
	DesktopOnAttention bool
	DesktopOnComplete  bool
	DesktopOnFail      bool
}

// Default returns the baseline config used when no file exists and the user
// hasn't completed onboarding.
func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		ServerURL:   "wss://your-server.example.com/ws",
		Workdir:     home,
		AgentCodex:  "/usr/local/bin/codex",
		AgentOC:     "/usr/local/bin/opencode",
		MaxSessions: 4,
		Agent:       "codex",
		Model:       "gpt-5-sonnet",
		Theme:       "charm-dark",
		Density:     "comfortable",
		LogLevel:    "info",
		LogFile:     filepath.Join(home, ".local", "state", "daemonctl", "daemonctl.log"),
		Notifications: Notifications{
			BellOnAttention:    true,
			DesktopOnAttention: true,
		},
		UsedDefaults: true,
	}
}

// Path returns the absolute config file path, honoring $XDG_CONFIG_HOME.
func Path() string {
	if p := os.Getenv("DAEMONCTL_CONFIG"); p != "" {
		return p
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "daemonctl", "config.toml")
}

// Exists reports whether a config file is present on disk.
func Exists() bool {
	_, err := os.Stat(Path())
	return err == nil
}

// Load reads config.toml. If the file is missing a Default() is returned —
// callers distinguish first-run via Exists().
func Load() (Config, error) {
	p := Path()
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}
	defer f.Close()

	c := Default()
	c.UsedDefaults = false

	sc := bufio.NewScanner(f)
	section := ""
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		k, v, ok := parseKV(line)
		if !ok {
			continue
		}
		applyKV(&c, section, k, v)
	}
	return c, sc.Err()
}

// Save writes the config to disk, creating parent directories as needed.
func Save(c Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	// Write atomically via a temp file in the same directory.
	tmp, err := os.CreateTemp(filepath.Dir(p), ".config.toml.*")
	if err != nil {
		return err
	}
	if _, err := tmp.WriteString(Encode(c)); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), p)
}

// Encode returns the TOML serialization of c. Exposed so tests can round-trip
// without touching disk.
func Encode(c Config) string {
	var b strings.Builder
	b.WriteString("# daemonctl config — written by the onboarding wizard\n")
	b.WriteString("# edit by hand or re-run `daemonctl --reset-config`\n\n")

	writeStr(&b, "server_url", c.ServerURL)
	writeStr(&b, "auth_token", c.AuthToken)
	writeStr(&b, "workdir", c.Workdir)
	writeStr(&b, "agent_codex", c.AgentCodex)
	writeStr(&b, "agent_opencode", c.AgentOC)
	writeInt(&b, "max_sessions", c.MaxSessions)
	writeStr(&b, "agent", c.Agent)
	writeStr(&b, "model", c.Model)
	writeStr(&b, "theme", c.Theme)
	writeStr(&b, "density", c.Density)
	writeStr(&b, "log_level", c.LogLevel)
	writeStr(&b, "log_file", c.LogFile)
	writeBool(&b, "telemetry", c.Telemetry)

	b.WriteString("\n[notifications]\n")
	writeBool(&b, "bell_on_attention", c.Notifications.BellOnAttention)
	writeBool(&b, "desktop_on_attention", c.Notifications.DesktopOnAttention)
	writeBool(&b, "desktop_on_complete", c.Notifications.DesktopOnComplete)
	writeBool(&b, "desktop_on_fail", c.Notifications.DesktopOnFail)

	return b.String()
}

func writeStr(b *strings.Builder, k, v string) {
	fmt.Fprintf(b, "%s = %q\n", k, v)
}
func writeInt(b *strings.Builder, k string, v int) {
	fmt.Fprintf(b, "%s = %d\n", k, v)
}
func writeBool(b *strings.Builder, k string, v bool) {
	fmt.Fprintf(b, "%s = %t\n", k, v)
}

func parseKV(line string) (string, string, bool) {
	eq := strings.IndexRune(line, '=')
	if eq < 0 {
		return "", "", false
	}
	k := strings.TrimSpace(line[:eq])
	v := strings.TrimSpace(line[eq+1:])
	// Strip trailing inline comments. When the value is quoted, preserve
	// everything up to and including the closing quote.
	if strings.HasPrefix(v, "\"") {
		// Walk past the opening quote, honoring backslash escapes, until we
		// find the matching close quote.
		esc := false
		for i := 1; i < len(v); i++ {
			if esc {
				esc = false
				continue
			}
			if v[i] == '\\' {
				esc = true
				continue
			}
			if v[i] == '"' {
				v = v[:i+1]
				break
			}
		}
	} else if hash := strings.Index(v, "#"); hash >= 0 {
		v = strings.TrimSpace(v[:hash])
	}
	return k, v, k != ""
}

func applyKV(c *Config, section, k, v string) {
	switch section {
	case "":
		switch k {
		case "server_url":
			c.ServerURL = unquote(v)
		case "auth_token":
			c.AuthToken = unquote(v)
		case "workdir":
			c.Workdir = unquote(v)
		case "agent_codex":
			c.AgentCodex = unquote(v)
		case "agent_opencode":
			c.AgentOC = unquote(v)
		case "max_sessions":
			if n, err := strconv.Atoi(v); err == nil {
				c.MaxSessions = n
			}
		case "agent":
			c.Agent = unquote(v)
		case "model":
			c.Model = unquote(v)
		case "theme":
			c.Theme = unquote(v)
		case "density":
			c.Density = unquote(v)
		case "log_level":
			c.LogLevel = unquote(v)
		case "log_file":
			c.LogFile = unquote(v)
		case "telemetry":
			c.Telemetry = v == "true"
		}
	case "notifications":
		b := v == "true"
		switch k {
		case "bell_on_attention":
			c.Notifications.BellOnAttention = b
		case "desktop_on_attention":
			c.Notifications.DesktopOnAttention = b
		case "desktop_on_complete":
			c.Notifications.DesktopOnComplete = b
		case "desktop_on_fail":
			c.Notifications.DesktopOnFail = b
		}
	}
}

func unquote(v string) string {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		// Handle the small set of escapes our encoder emits via strconv.
		if s, err := strconv.Unquote(v); err == nil {
			return s
		}
		return v[1 : len(v)-1]
	}
	return v
}
