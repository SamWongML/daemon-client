package session

// Event messages emitted by engines (mock or real PTY). App's Update routes these into the store.

type AppendMsg struct {
	ID      ID
	Content string
}

type StatusMsg struct {
	ID     ID
	Status Status
}

type ActivityMsg struct {
	ID       ID
	Activity string
}

type QuestionMsg struct {
	ID ID
	Q  Question
}

type PermissionMsg struct {
	ID ID
	P  PermissionReq
}

type DoneMsg struct {
	ID     ID
	Result Result
}

type CreatedMsg struct{ Session *Session }
