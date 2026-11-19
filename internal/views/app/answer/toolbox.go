package answer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ToolInfo is a tool as a client is told it: its name, what it does and the
// JSON Schema of its arguments.
type ToolInfo struct {
	Name, Description, Schema string
}

// ToolOutcome is how a call of a tool ended.
type ToolOutcome string

// The outcomes of a tool call, the Ask box's own: its trace names the same
// (callRan, callBadArgs, callReadFailed).
const (
	// ToolRan is a call that read its data: Text is the card as JSON.
	ToolRan ToolOutcome = "ran"
	// ToolUnknown is a call of a tool the box does not have.
	ToolUnknown ToolOutcome = "unknown-tool"
	// ToolBadArguments is a call whose arguments the tool did not accept.
	ToolBadArguments ToolOutcome = "bad-arguments"
	// ToolReadFailed is a call whose data could not be read.
	ToolReadFailed ToolOutcome = "read-failed"
	// ToolUnauthenticated is a call whose token the read API no longer accepts.
	ToolUnauthenticated ToolOutcome = "unauthenticated"
)

// ToolResult is what a tool call returned: how it ended, and the text the
// caller reads, a card as JSON when the call ran and the reason otherwise.
type ToolResult struct {
	Outcome ToolOutcome
	Text    string
}

// ErrMissingDependency reports a Toolbox built without a read client.
var ErrMissingDependency = errors.New("a dependency is missing")

// ToolTimeout bounds one tool, which reads the core a few times.
const ToolTimeout = 20 * time.Second

// ToolboxConfig is what a Toolbox needs.
type ToolboxConfig struct {
	// Read answers the tools' reads, as the core's read API does for a
	// session or a script token. Every call carries the caller's token, so
	// the caller's grants limit what a tool sees.
	Read Reader
	// Today returns the date of the call; nil is the UTC date of the clock.
	Today func() kernel.Date
}

// Toolbox offers the Ask box's tools, the read-only answerers a model may
// call (D-203), to a caller that is no model of the Ask box: the MCP server
// (D-213). It is the same table and the same answerers the page's loop uses,
// reading the read API with the caller's own token, so nothing a tool shows
// goes beyond what the token's grants cover. It holds nothing between calls,
// logs nothing and changes nothing. It draws no page, so its cards have no
// links.
type Toolbox struct {
	read     Reader
	commands *Commands
	tools    []Tool
}

// NewToolbox returns a Toolbox over cfg; it fails with ErrMissingDependency
// without a read client.
func NewToolbox(cfg *ToolboxConfig) (*Toolbox, error) {
	if cfg.Read == nil {
		return nil, ErrMissingDependency
	}
	return &Toolbox{read: cfg.Read, commands: NewCommands(cfg.Read, cfg.Today, NoLinks{}), tools: Tools()}, nil
}

// Tools returns the tools, in the order the Ask box's model is told them.
func (t *Toolbox) Tools() []ToolInfo {
	out := make([]ToolInfo, len(t.tools))
	for i := range t.tools {
		out[i] = ToolInfo{Name: t.tools[i].Name, Description: t.tools[i].Description, Schema: t.tools[i].Schema}
	}
	return out
}

// Call runs the tool name with args, a JSON object text, reading with token,
// a script token's secret or a session's token the read API accepts. The
// result's text is never logged here, and the arguments are in no error.
func (t *Toolbox) Call(ctx context.Context, token, name, args string) ToolResult {
	var tool *Tool
	for i := range t.tools {
		if t.tools[i].Name == name {
			tool = &t.tools[i]
		}
	}
	if tool == nil {
		return ToolResult{Outcome: ToolUnknown, Text: ToolFailure("there is no tool called %s", name)}
	}
	q, err := tool.Query(args, t.commands.Today())
	if err != nil {
		return ToolResult{Outcome: ToolBadArguments, Text: ToolFailure("%s", err.Error())}
	}
	ctx, cancel := context.WithTimeout(ctx, ToolTimeout)
	defer cancel()
	caller, err := t.caller(ctx, token)
	if err == nil {
		var card Card
		if card, err = t.commands.Run(ctx, caller, &q); err == nil {
			return ToolResult{Outcome: ToolRan, Text: CardResult(&card, CardID(1))}
		}
	}
	if connect.CodeOf(err) == connect.CodeUnauthenticated {
		return ToolResult{Outcome: ToolUnauthenticated, Text: ToolFailure("the token is not accepted")}
	}
	// What failed is not the caller's to see: the result says only that the
	// read did not work.
	return ToolResult{Outcome: ToolReadFailed, Text: ToolFailure("the data could not be read")}
}

// caller returns who the answerers read as: the token and the number format
// of the owner's settings (D-113), which the read API tells the token.
func (t *Toolbox) caller(ctx context.Context, token string) (Caller, error) {
	c := Caller{Token: token}
	settings, err := t.read.GetSettings(ctx, bearer(&viewsv1.GetSettingsRequest{}, c))
	if err != nil {
		return Caller{}, fmt.Errorf("GetSettings: %w", err)
	}
	c.Format = FormatOf(settings.Msg.GetNumberFormat())
	return c, nil
}
