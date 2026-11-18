package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// newToolbox returns a Toolbox over the fake read API served on a real
// Connect server, with the fake to read what it was asked.
func newToolbox(t *testing.T) (*Toolbox, *fakeRead) {
	t.Helper()
	read := &fakeRead{}
	mux := http.NewServeMux()
	mux.Handle(viewsv1connect.NewReadServiceHandler(read))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	box, err := NewToolbox(&ToolboxConfig{
		Read:  viewsv1connect.NewReadServiceClient(server.Client(), server.URL),
		Today: func() kernel.Date { return mustDate(t, askToday) },
	})
	if err != nil {
		t.Fatalf("NewToolbox: %v", err)
	}
	return box, read
}

// TestToolboxOffersTheAskBoxsTools checks that the toolbox lists exactly the
// tools the Ask box's model is given, with the same descriptions and
// schemas: one table, no parallel copy (D-213, D-203).
func TestToolboxOffersTheAskBoxsTools(t *testing.T) {
	t.Parallel()
	box, _ := newToolbox(t)
	tools := askTools()
	want := make([]ToolInfo, 0, len(tools))
	for _, tool := range tools {
		want = append(want, ToolInfo{Name: tool.Name, Description: tool.Description, Schema: tool.Schema})
	}
	if diff := cmp.Diff(want, box.Tools()); diff != "" {
		t.Errorf("Tools() (-want +got):\n%s", diff)
	}
}

// TestToolboxCallsReadWithTheCallersOwnToken checks that a call reads the
// read API only with the token it was given, so that token's grants limit
// what the tool sees, and that the card comes back as the model's tool result
// does, in the owner's number format.
func TestToolboxCallsReadWithTheCallersOwnToken(t *testing.T) {
	t.Parallel()
	box, read := newToolbox(t)
	got := box.Call(t.Context(), readSession, "allocation", `{"by":"currency"}`)
	if got.Outcome != ToolRan {
		t.Fatalf("Call() outcome = %q (%s), want %q", got.Outcome, got.Text, ToolRan)
	}
	var card toolResult
	if err := json.Unmarshal([]byte(got.Text), &card); err != nil {
		t.Fatalf("the result is not a card: %v", err)
	}
	if card.ID != "c1" || card.Title == "" || len(card.Rows) == 0 {
		t.Errorf("card = %+v, want an id, a title and rows", card)
	}
	if !slices.ContainsFunc(card.Rows, func(r toolResultRow) bool { return strings.Contains(r.Figure, ",") }) {
		t.Errorf("rows = %+v, want amounts in the owner's decimal comma", card.Rows)
	}
	read.mu.Lock()
	defer read.mu.Unlock()
	if len(read.bearers) == 0 {
		t.Fatal("the call read nothing")
	}
	for _, b := range read.bearers {
		if b != "Bearer "+readSession {
			t.Errorf("the read API was called with %q, want only the caller's token", b)
		}
	}
}

// TestToolboxCallOutcomes checks how a call ends when the tool is unknown,
// the arguments are refused or the token is not accepted, and that no result
// quotes the arguments (GR-12).
func TestToolboxCallOutcomes(t *testing.T) {
	t.Parallel()
	box, _ := newToolbox(t)
	tests := []struct {
		name, token, tool, args string
		want                    ToolOutcome
	}{
		{name: "a tool that does not exist", token: readSession, tool: "delete_account", args: `{}`, want: ToolUnknown},
		{name: "an argument the schema does not list", token: readSession, tool: "allocation", args: `{"surprise":"1234.56"}`, want: ToolBadArguments},
		{name: "a value outside the enum", token: readSession, tool: "allocation", args: `{"by":"1234.56"}`, want: ToolBadArguments},
		{name: "arguments that are no object", token: readSession, tool: "value", args: `[1234.56]`, want: ToolBadArguments},
		{name: "a token the read API does not accept", token: "not-a-token", tool: "allocation", args: `{}`, want: ToolUnauthenticated},
		{name: "no arguments", token: readSession, tool: "allocation", args: ``, want: ToolRan},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := box.Call(t.Context(), tc.token, tc.tool, tc.args)
			if got.Outcome != tc.want {
				t.Errorf("Call() outcome = %q (%s), want %q", got.Outcome, got.Text, tc.want)
			}
			if got.Outcome != ToolRan && strings.Contains(got.Text, "1234.56") {
				t.Errorf("the refusal quotes the arguments: %s", got.Text)
			}
		})
	}
}

// TestToolboxNeedsARead checks that a toolbox without a read client is
// refused.
func TestToolboxNeedsARead(t *testing.T) {
	t.Parallel()
	if _, err := NewToolbox(&ToolboxConfig{}); err == nil {
		t.Error("NewToolbox without a read client succeeded")
	}
}
