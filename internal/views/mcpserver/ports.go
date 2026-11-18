package mcpserver

import (
	"context"
	"errors"
)

// Why a token is not accepted. Tokens.Authenticate returns them wrapped.
var (
	// ErrInvalidToken reports a secret that may not be used: malformed,
	// unknown, expired, revoked or not a script token's.
	ErrInvalidToken = errors.New("invalid script token")
	// ErrNotAllowed reports a valid script token the owner did not allow to
	// use MCP (D-213).
	ErrNotAllowed = errors.New("script token not allowed to use MCP")
)

// Gate says whether the MCP server answers (D-212): the one switch for agents
// and models is on and the MCP toggle under it is on. It is asked at every
// request, so turning either off stops the answers at once.
type Gate interface {
	Enabled(ctx context.Context) (bool, error)
}

// Tokens checks the script token of a request.
type Tokens interface {
	// Authenticate returns the ID of the token secret stands for when the
	// token may authenticate to the MCP server, or an error wrapping
	// ErrInvalidToken or ErrNotAllowed; with ErrNotAllowed the token's ID is
	// returned too, for the audit entry of the refusal. Any other error is a
	// failure to check. The error never holds the secret.
	Authenticate(ctx context.Context, secret string) (tokenID string, err error)
}

// Tool is a tool a client is told: its name, what it does and the JSON Schema
// of its arguments.
type Tool struct {
	Name, Description, Schema string
}

// Outcome is how a tool call ended.
type Outcome string

// The outcomes of a tool call.
const (
	// OutcomeRan is a call that read its data.
	OutcomeRan Outcome = "ran"
	// OutcomeBadArguments is a call whose arguments the tool did not accept.
	OutcomeBadArguments Outcome = "bad-arguments"
	// OutcomeReadFailed is a call whose data could not be read, or whose
	// token the read API no longer accepts.
	OutcomeReadFailed Outcome = "read-failed"
	// OutcomeNotAllowed is a request of a valid token the owner did not allow
	// to use MCP; it names no tool.
	OutcomeNotAllowed Outcome = "not-allowed"
)

// Result is what a tool call returned: how it ended and the text a client
// reads, a card as JSON when the call ran and the reason otherwise.
type Result struct {
	Outcome Outcome
	Text    string
}

// Toolbox offers the tools and runs one with the caller's own token, so what
// the token's grants cover limits what the tool sees. It never changes
// anything.
type Toolbox interface {
	Tools() []Tool
	// Call runs tool name with args, a JSON object text, reading with
	// secret. A tool it does not have is a call whose arguments were not
	// accepted.
	Call(ctx context.Context, secret, name, args string) Result
}

// Call is what the audit log learns of a tool call, or of the refusal of a
// token not allowed to use MCP (D-211): who, which tool (none for the
// refusal) and how it ended. It has no field for an argument or a result.
type Call struct {
	// Token is the ID of the script token that called, never its secret.
	Token string
	// Tool is the tool's name.
	Tool    string
	Outcome Outcome
}

// Auditor enters a tool call in the audit log as usage. A call it can't
// record shows nothing.
type Auditor interface {
	RecordCall(ctx context.Context, c Call) error
}
