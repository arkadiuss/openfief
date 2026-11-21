package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// callerToken is the one token the fake read API accepts.
const callerToken = "session-read"

// fakeRead is the part of the read API the allocation tool reads, with the
// token check of the core's: every call is recorded with its bearer token, and
// a token other than callerToken is refused.
type fakeRead struct {
	viewsv1connect.UnimplementedReadServiceHandler

	mu      sync.Mutex
	bearers []string
}

func (f *fakeRead) GetSettings(_ context.Context, r *connect.Request[viewsv1.GetSettingsRequest]) (*connect.Response[viewsv1.GetSettingsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.GetSettingsResponse{
		NumberFormat: &viewsv1.NumberFormat{Key: "space-comma", GroupSeparator: " ", DecimalSeparator: ","},
	}), nil
}

func (f *fakeRead) ListValuationStatuses(_ context.Context, r *connect.Request[viewsv1.ListValuationStatusesRequest]) (*connect.Response[viewsv1.ListValuationStatusesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListValuationStatusesResponse{
		Statuses: []*viewsv1.ValuationStatus{{AccountId: "acc-1", ValuedThrough: askToday}},
	}), nil
}

func (f *fakeRead) GetAllocation(_ context.Context, r *connect.Request[viewsv1.GetAllocationRequest]) (*connect.Response[viewsv1.GetAllocationResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	pln := func(amount string) *viewsv1.Money { return &viewsv1.Money{Amount: amount, Currency: "PLN"} }
	return connect.NewResponse(&viewsv1.GetAllocationResponse{
		Date: r.Msg.GetDate(), By: r.Msg.GetBy(), TotalReporting: pln("1500.00"),
		Groups: []*viewsv1.AllocationGroup{
			{Key: "listed", ValueReporting: pln("1000.00")},
			{Key: "cash", ValueReporting: pln("500.00")},
		},
	}), nil
}

func (f *fakeRead) check(h http.Header) error {
	f.mu.Lock()
	f.bearers = append(f.bearers, h.Get("Authorization"))
	f.mu.Unlock()
	if h.Get("Authorization") == "Bearer "+callerToken {
		return nil
	}
	return fmt.Errorf("fake core: %w", connect.NewError(connect.CodeUnauthenticated, nil))
}

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
		Today: func() kernel.Date { return mustDay(t, askToday) },
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
	tools := Tools()
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
	got := box.Call(t.Context(), callerToken, "allocation", `{"by":"currency"}`)
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
		if b != "Bearer "+callerToken {
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
		{name: "a tool that does not exist", token: callerToken, tool: "delete_account", args: `{}`, want: ToolUnknown},
		{name: "an argument the schema does not list", token: callerToken, tool: "allocation", args: `{"surprise":"1234.56"}`, want: ToolBadArguments},
		{name: "a value outside the enum", token: callerToken, tool: "allocation", args: `{"by":"1234.56"}`, want: ToolBadArguments},
		{name: "arguments that are no object", token: callerToken, tool: "value", args: `[1234.56]`, want: ToolBadArguments},
		{name: "a token the read API does not accept", token: "not-a-token", tool: "allocation", args: `{}`, want: ToolUnauthenticated},
		{name: "no arguments", token: callerToken, tool: "allocation", args: ``, want: ToolRan},
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

// TestToolboxCalcComputesWithoutReadingData checks that the calc tool answers
// from its arguments alone: the call asks the read API only for the owner's
// number format, with the caller's token, shows the expression with its exact
// result, says the figures are the ones given, and a token the read API does
// not accept is refused like any tool's.
func TestToolboxCalcComputesWithoutReadingData(t *testing.T) {
	t.Parallel()
	box, read := newToolbox(t)
	got := box.Call(t.Context(), callerToken, "calc", `{"expression":"0.1 + 0.2"}`)
	if got.Outcome != ToolRan {
		t.Fatalf("Call() outcome = %q (%s), want %q", got.Outcome, got.Text, ToolRan)
	}
	var card toolResult
	if err := json.Unmarshal([]byte(got.Text), &card); err != nil {
		t.Fatalf("the result is not a card: %v", err)
	}
	if card.Title != "0.1 + 0.2 = 0.3" || !strings.Contains(card.Text, "Computed from the figures given") {
		t.Errorf("card = %+v, want the expression with its exact result and the figures-given wording", card)
	}
	if len(card.Rows) != 2 || card.Rows[1].Figure != "0,3" {
		t.Errorf("rows = %+v, want the result in the owner's decimal comma", card.Rows)
	}
	read.mu.Lock()
	n := len(read.bearers)
	read.mu.Unlock()
	if n != 1 {
		t.Errorf("the call made %d reads, want only the settings' number format", n)
	}
}

// TestToolboxCalcRefusesWhatCannotBeComputed checks that an expression that
// divides by zero is a card with the reason, and that a token the read API does
// not accept is refused like any tool's.
func TestToolboxCalcRefusesWhatCannotBeComputed(t *testing.T) {
	t.Parallel()
	box, _ := newToolbox(t)
	refused := box.Call(t.Context(), callerToken, "calc", `{"expression":"1/0"}`)
	if refused.Outcome != ToolRan || !strings.Contains(refused.Text, "divides by zero") {
		t.Errorf("1/0 = %q (%s), want a card that refuses with its reason", refused.Outcome, refused.Text)
	}
	if bad := box.Call(t.Context(), "not-a-token", "calc", `{"expression":"1+1"}`); bad.Outcome != ToolUnauthenticated {
		t.Errorf("calc with a bad token = %q, want %q", bad.Outcome, ToolUnauthenticated)
	}
}
