package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := Config{
		ServerURL:   "wss://example.com/ws",
		AuthToken:   "abc\"def\n",
		Workdir:     "/tmp/work dir",
		AgentCodex:  "/usr/local/bin/codex",
		AgentOC:     "/usr/local/bin/opencode",
		MaxSessions: 8,
		Agent:       "opencode",
		Model:       "gpt-5-sonnet",
		Theme:       "tokyonight-storm",
		Density:     "compact",
		LogLevel:    "debug",
		LogFile:     "/tmp/log",
		Telemetry:   false,
		Notifications: Notifications{
			BellOnAttention:    true,
			DesktopOnAttention: true,
			DesktopOnComplete:  false,
			DesktopOnFail:      true,
		},
	}
	encoded := Encode(in)

	// Sanity: every expected key is present.
	for _, k := range []string{"server_url", "auth_token", "workdir", "max_sessions",
		"agent", "model", "theme", "density", "log_level", "log_file", "telemetry",
		"[notifications]", "bell_on_attention", "desktop_on_fail"} {
		if !strings.Contains(encoded, k) {
			t.Fatalf("encoded output missing %q:\n%s", k, encoded)
		}
	}

	// Save → Load via temp XDG_CONFIG_HOME.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DAEMONCTL_CONFIG", "") // ensure the override doesn't leak
	if Exists() {
		t.Fatalf("fresh temp dir should not have a config")
	}
	if err := Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	if !Exists() {
		t.Fatalf("Exists() false after Save")
	}
	out, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.UsedDefaults {
		t.Fatalf("Load should clear UsedDefaults when the file exists")
	}
	out.UsedDefaults = in.UsedDefaults
	if out != in {
		t.Fatalf("round-trip mismatch:\n want: %+v\n got:  %+v", in, out)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("DAEMONCTL_CONFIG", "")
	c, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !c.UsedDefaults {
		t.Fatalf("missing file should flag UsedDefaults=true")
	}
	if c.MaxSessions == 0 {
		t.Fatalf("default should set MaxSessions")
	}
}

func TestPathHonorsOverride(t *testing.T) {
	dir := t.TempDir()
	custom := filepath.Join(dir, "nested", "custom.toml")
	t.Setenv("DAEMONCTL_CONFIG", custom)
	if got := Path(); got != custom {
		t.Fatalf("Path() = %q, want %q", got, custom)
	}
}

func TestLoadIgnoresCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("DAEMONCTL_CONFIG", "")
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `# hello
server_url = "wss://a"   # inline comment

[notifications]
bell_on_attention = true
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.ServerURL != "wss://a" {
		t.Fatalf("ServerURL = %q", c.ServerURL)
	}
	if !c.Notifications.BellOnAttention {
		t.Fatalf("bell_on_attention should parse as true")
	}
}
