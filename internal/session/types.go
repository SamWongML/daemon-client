package session

import (
	"context"
	"time"
)

type ID string

type Status int

const (
	StatusPending Status = iota
	StatusStarting
	StatusRunning
	StatusAwaitingInput
	StatusAwaitingPerm
	StatusIdle
	StatusPaused
	StatusCompleted
	StatusFailed
	StatusDisconnected
)

func (s Status) Name() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusAwaitingInput:
		return "awaiting_input"
	case StatusAwaitingPerm:
		return "awaiting_perm"
	case StatusIdle:
		return "idle"
	case StatusPaused:
		return "paused"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusDisconnected:
		return "disconnected"
	}
	return "unknown"
}

// Priority controls sidebar grouping order (lower = higher priority).
func (s Status) Priority() int {
	switch s {
	case StatusAwaitingPerm:
		return 0
	case StatusAwaitingInput:
		return 1
	case StatusRunning:
		return 2
	case StatusStarting:
		return 3
	case StatusIdle:
		return 4
	case StatusPaused:
		return 5
	case StatusPending:
		return 6
	case StatusCompleted:
		return 7
	case StatusFailed:
		return 8
	case StatusDisconnected:
		return 9
	}
	return 99
}

func (s Status) Glyph() string {
	switch s {
	case StatusRunning:
		return "⟳"
	case StatusStarting:
		return "◌"
	case StatusAwaitingInput:
		return "?"
	case StatusAwaitingPerm:
		return "⚠"
	case StatusIdle:
		return "◦"
	case StatusPaused:
		return "❚❚"
	case StatusCompleted:
		return "✓"
	case StatusFailed:
		return "✗"
	case StatusDisconnected:
		return "∅"
	}
	return "·"
}

func ParseStatus(s string) Status {
	switch s {
	case "pending":
		return StatusPending
	case "starting":
		return StatusStarting
	case "running":
		return StatusRunning
	case "awaiting_input":
		return StatusAwaitingInput
	case "awaiting_perm":
		return StatusAwaitingPerm
	case "idle":
		return StatusIdle
	case "paused":
		return StatusPaused
	case "completed":
		return StatusCompleted
	case "failed":
		return StatusFailed
	case "disconnected":
		return StatusDisconnected
	}
	return StatusIdle
}

type Task struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Agent         string `json:"agent"`
	Model         string `json:"model"`
	Workdir       string `json:"workdir"`
	InitialStatus string `json:"initial_status"`
	TokensUsed    int    `json:"tokens_used"`
	TokensBudget  int    `json:"tokens_budget"`
	CostUSD       float64 `json:"cost_usd"`
	StartedAt     string  `json:"started_at"`
	Transcript    string  `json:"transcript"`
}

type Session struct {
	ID         ID
	Title      string
	Agent      string
	Model      string
	Workdir    string
	Status     Status
	Activity   string
	Tokens     int
	Budget     int
	CostUSD    float64
	StartedAt  time.Time
	Transcript string // accumulated markdown
}

type Question struct {
	Prompt  string
	Options []string
}

type PermissionReq struct {
	Tool string
	Args string
	Diff string
}

type PermissionDecision int

const (
	DecisionAllowOnce PermissionDecision = iota
	DecisionAllowAlways
	DecisionDeny
)

type Result struct {
	OK      bool
	Summary string
}

// Engine is the phase-2 seam — mock engine implements it now, real PTY/WS engine in phase 2.
type Engine interface {
	Start(ctx context.Context, task Task) (ID, error)
	Stop(id ID) error
	Kill(id ID) error
	Resume(id ID) error
	Respond(id ID, text string) error
	Permit(id ID, req PermissionReq, decision PermissionDecision) error
}
