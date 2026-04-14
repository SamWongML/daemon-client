package fixtures

import "embed"

//go:embed tasks.json transcripts/*.md
var FS embed.FS
