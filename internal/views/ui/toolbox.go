package ui

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
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

// ToolboxConfig is what a Toolbox needs.
type ToolboxConfig struct {
	// Read answers the tools' reads, as the core's read API does for a
	// session or a script token. Every call carries the caller's token, so
	// the caller's grants limit what a tool sees.
	Read viewsv1connect.ReadServiceClient
	// Today returns the date of the call; nil is the UTC date of the clock.
	Today func() kernel.Date
}

// Toolbox offers the Ask box's tools, the read-only answerers a model may
// call (D-203), to a caller that is no model of the Ask box: the MCP server
// (D-213). It is the same table and the same answerers the page's loop uses,
// reading the read API with the caller's own token, so nothing a tool shows
// goes beyond what the token's grants cover. It holds nothing between calls,
// logs nothing and changes nothing.
type Toolbox struct {
	s        *server
	commands commandAnswerer
	tools    []askTool
}

// NewToolbox returns a Toolbox over cfg; it fails with ErrMissingDependency
// without a read client.
func NewToolbox(cfg *ToolboxConfig) (*Toolbox, error) {
	if cfg.Read == nil {
		return nil, ErrMissingDependency
	}
	s := &server{read: cfg.Read, today: cfg.Today}
	if s.today == nil {
		s.today = utcToday
	}
	return &Toolbox{s: s, commands: commandAnswerer{s: s}, tools: askTools()}, nil
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
	var tool *askTool
	for i := range t.tools {
		if t.tools[i].Name == name {
			tool = &t.tools[i]
		}
	}
	if tool == nil {
		return ToolResult{Outcome: ToolUnknown, Text: toolFailure("there is no tool called %s", name)}
	}
	q, err := tool.query(args, t.s.today())
	if err != nil {
		return ToolResult{Outcome: ToolBadArguments, Text: toolFailure("%s", err.Error())}
	}
	ctx, cancel := context.WithTimeout(ctx, askToolTimeout)
	defer cancel()
	sess := session{Token: token}
	r, err := t.request(ctx, sess)
	if err == nil {
		var card answerCard
		if card, err = t.commands.run(r, sess, &q); err == nil {
			return ToolResult{Outcome: ToolRan, Text: cardResult(&card, cardID(1))}
		}
	}
	if connect.CodeOf(err) == connect.CodeUnauthenticated {
		return ToolResult{Outcome: ToolUnauthenticated, Text: toolFailure("the token is not accepted")}
	}
	// What failed is not the caller's to see: the result says only that the
	// read did not work.
	return ToolResult{Outcome: ToolReadFailed, Text: toolFailure("the data could not be read")}
}

// request returns the request the answerers read under: ctx and the number
// format of the owner's settings (D-113), which the read API tells the token.
func (t *Toolbox) request(ctx context.Context, sess session) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "/mcp", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("starting the call: %w", err)
	}
	settings, err := t.s.read.GetSettings(ctx, bearer(&viewsv1.GetSettingsRequest{}, sess))
	if err != nil {
		return nil, fmt.Errorf("GetSettings: %w", err)
	}
	return withFormat(r, formatOf(settings.Msg.GetNumberFormat())), nil
}
