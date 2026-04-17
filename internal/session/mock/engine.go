package mock

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/demon/daemon-client/internal/fixtures"
	"github.com/demon/daemon-client/internal/session"
)

type script struct {
	events []event
}

type event struct {
	kind    string // "text", "state", "delay"
	text    string
	state   session.Status
	delayMs int
}

// PanicEvent is sent via program.Send when a replay goroutine recovers from a
// panic. The app layer handles it by showing an error toast.
type PanicEvent struct {
	Err   any
	Stack string
}

// Engine replays canned transcripts via program.Send so the UI sees realistic streaming.
type Engine struct {
	store   *session.Store
	program *tea.Program
	scripts map[session.ID]*script
}

func New(store *session.Store) *Engine {
	return &Engine{store: store, scripts: map[session.ID]*script{}}
}

// LoadFixtures reads tasks.json + transcripts from embedded FS and seeds the store.
// Sessions in a terminal state get their transcript rendered statically; live sessions
// keep empty transcripts pending StartReplay which streams through program.Send.
func (e *Engine) LoadFixtures() error {
	raw, err := fixtures.FS.ReadFile("tasks.json")
	if err != nil {
		return err
	}
	var tasks []session.Task
	if err := json.Unmarshal(raw, &tasks); err != nil {
		return err
	}
	for _, t := range tasks {
		sc, err := loadScript(fixtures.FS, "transcripts/"+t.Transcript)
		if err != nil {
			return err
		}
		id := session.ID(t.ID)
		e.scripts[id] = sc
		status := session.ParseStatus(t.InitialStatus)
		sess := &session.Session{
			ID:      id,
			Title:   t.Title,
			Agent:   t.Agent,
			Model:   t.Model,
			Workdir: t.Workdir,
			Status:  status,
			Tokens:  t.TokensUsed,
			Budget:  t.TokensBudget,
			CostUSD: t.CostUSD,
		}
		if isTerminal(status) {
			var b strings.Builder
			for _, ev := range sc.events {
				if ev.kind == "text" {
					b.WriteString(ev.text)
				}
			}
			sess.Transcript = b.String()
		}
		e.store.Add(sess)
	}
	return nil
}

func (e *Engine) SetProgram(p *tea.Program) { e.program = p }

// StartReplay fires a goroutine that streams the session's scripted transcript.
// Safe to call on any session; no-op if no script is registered.
func (e *Engine) StartReplay(id session.ID) {
	sc, ok := e.scripts[id]
	if !ok || e.program == nil {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				e.program.Send(PanicEvent{
					Err:   r,
					Stack: fmt.Sprintf("%v\n%s", r, buf[:n]),
				})
			}
		}()
		for _, ev := range sc.events {
			switch ev.kind {
			case "delay":
				time.Sleep(time.Duration(ev.delayMs) * time.Millisecond)
			case "text":
				e.program.Send(session.AppendMsg{ID: id, Content: ev.text})
			case "state":
				e.program.Send(session.StatusMsg{ID: id, Status: ev.state})
				if isTerminal(ev.state) || ev.state == session.StatusAwaitingInput || ev.state == session.StatusAwaitingPerm {
					return
				}
			}
		}
	}()
}

func isTerminal(s session.Status) bool {
	return s == session.StatusCompleted || s == session.StatusFailed || s == session.StatusPaused
}

func loadScript(f fs.FS, path string) (*script, error) {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return nil, err
	}
	sc := &script{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	var textBuf strings.Builder
	flushText := func() {
		if textBuf.Len() > 0 {
			sc.events = append(sc.events, event{kind: "text", text: textBuf.String()})
			textBuf.Reset()
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "@@STATE:"):
			flushText()
			name := strings.TrimSpace(strings.TrimPrefix(line, "@@STATE:"))
			sc.events = append(sc.events, event{kind: "state", state: session.ParseStatus(name)})
		case strings.HasPrefix(line, "@@DELAY:"):
			flushText()
			n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "@@DELAY:")))
			sc.events = append(sc.events, event{kind: "delay", delayMs: n})
		case strings.HasPrefix(line, "@@PERM:"), strings.HasPrefix(line, "@@QUESTION:"), strings.HasPrefix(line, "@@RESUME_AFTER_PERM"):
			// Phase 2 will parse these into real SessionPermissionMsg / SessionQuestionMsg events.
			// For showcase we surface them as transcript lines so the script is visible.
			flushText()
			sc.events = append(sc.events, event{kind: "text", text: "> " + line + "\n"})
		default:
			textBuf.WriteString(line)
			textBuf.WriteString("\n")
		}
	}
	flushText()
	return sc, scanner.Err()
}
