package mcpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/arkadiuss/openfief/internal/views/mcpserver"
)

// Synthetic secrets and IDs: no real token anywhere.
const (
	secretA = "ofs_00000000000000a1_synthetic-key-a"
	secretB = "ofs_00000000000000b2_synthetic-key-b"
	idA     = "00000000000000a1"
	idB     = "00000000000000b2"
	// noMCP is a valid token the owner did not allow to use MCP.
	noMCP   = "ofs_00000000000000c3_synthetic-key-c"
	idNoMCP = "00000000000000c3"
	// cardText is the text of the synthetic card the fake toolbox returns.
	cardText = `{"id":"c1","title":"Synthetic allocation"}`
)

// fakeGate is a Gate the test turns on and off.
type fakeGate struct {
	mu  sync.Mutex
	on  bool
	err error
}

func (g *fakeGate) Enabled(context.Context) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.on, g.err
}

func (g *fakeGate) set(on bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.on = on
}

// fakeTokens knows two tokens, one without the MCP flag, and fails for a
// secret it is told to.
type fakeTokens struct{}

var (
	errTokenStore = errors.New("store unavailable")
	errLogDown    = errors.New("the log is down")
	errDBDown     = errors.New("db down")
)

func (fakeTokens) Authenticate(_ context.Context, secret string) (string, error) {
	switch secret {
	case secretA:
		return idA, nil
	case secretB:
		return idB, nil
	case noMCP:
		return idNoMCP, fmt.Errorf("%w: token c3", mcpserver.ErrNotAllowed)
	case "ofs_broken":
		return "", errTokenStore
	}
	return "", fmt.Errorf("%w: unknown", mcpserver.ErrInvalidToken)
}

// readCall is one call the fake toolbox ran.
type readCall struct{ Secret, Tool, Args string }

// fakeToolbox offers two tools and records the calls it ran, with the secret
// each was made with.
type fakeToolbox struct {
	mu    sync.Mutex
	calls []readCall
}

func (*fakeToolbox) Tools() []mcpserver.Tool {
	return []mcpserver.Tool{
		{Name: "allocation", Description: "The allocation.", Schema: `{"type":"object","properties":{"by":{"type":"string"}},"additionalProperties":false}`},
		{Name: "value", Description: "The value.", Schema: `{"type":"object","properties":{},"additionalProperties":false}`},
	}
}

func (f *fakeToolbox) Call(_ context.Context, secret, name, args string) mcpserver.Result {
	f.mu.Lock()
	f.calls = append(f.calls, readCall{Secret: secret, Tool: name, Args: args})
	f.mu.Unlock()
	if strings.Contains(args, "refuse") {
		return mcpserver.Result{Outcome: mcpserver.OutcomeBadArguments, Text: `{"error":"the argument is none of the values the schema lists"}`}
	}
	if strings.Contains(args, "break") {
		return mcpserver.Result{Outcome: mcpserver.OutcomeReadFailed, Text: `{"error":"the data could not be read"}`}
	}
	return mcpserver.Result{Outcome: mcpserver.OutcomeRan, Text: cardText}
}

func (f *fakeToolbox) ran() []readCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]readCall(nil), f.calls...)
}

// fakeAuditor records the calls it was told, or fails.
type fakeAuditor struct {
	mu    sync.Mutex
	calls []mcpserver.Call
	fail  bool
}

func (a *fakeAuditor) RecordCall(_ context.Context, c mcpserver.Call) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.fail {
		return errLogDown
	}
	a.calls = append(a.calls, c)
	return nil
}

func (a *fakeAuditor) recorded() []mcpserver.Call {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]mcpserver.Call(nil), a.calls...)
}

// fixture is the MCP endpoint served over HTTP with its fakes.
type fixture struct {
	url     string
	gate    *fakeGate
	toolbox *fakeToolbox
	auditor *fakeAuditor
	// unauthenticated counts the requests refused with 401.
	unauthenticated atomic.Int64
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{gate: &fakeGate{on: true}, toolbox: &fakeToolbox{}, auditor: &fakeAuditor{}}
	h, err := mcpserver.New(&mcpserver.Config{
		Gate: f.gate, Tokens: fakeTokens{}, Toolbox: f.toolbox, Auditor: f.auditor,
		OnUnauthenticated: func(context.Context) { f.unauthenticated.Add(1) },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", h)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	f.url = srv.URL + "/mcp"
	return f
}

// reply is an HTTP answer: its status, challenge and body.
type reply struct {
	status    int
	challenge string
	body      string
}

// post sends one JSON-RPC message to the endpoint with the Authorization
// header authorization (none when empty).
func (f *fixture) post(t *testing.T, authorization, message string) reply {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, f.url, strings.NewReader(message))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	return do(t, req)
}

func do(t *testing.T, req *http.Request) reply {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL, err)
	}
	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); err != nil || closeErr != nil {
		t.Fatalf("reading the answer: %v %v", err, closeErr)
	}
	return reply{status: resp.StatusCode, challenge: resp.Header.Get("WWW-Authenticate"), body: string(body)}
}

const (
	initialize = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"synthetic-client","version":"0"}}}`
	listTools  = `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
)

// callTool returns the JSON-RPC message of a tools/call of name with args.
func callTool(name, args string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, name, args)
}

// rpc is the part of a JSON-RPC answer the tests read.
type rpc struct {
	Result struct {
		ServerInfo struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
			Annotations struct {
				ReadOnlyHint    bool  `json:"readOnlyHint"`
				DestructiveHint *bool `json:"destructiveHint"`
			} `json:"annotations"`
		} `json:"tools"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func decode(t *testing.T, r reply) rpc {
	t.Helper()
	if r.status != http.StatusOK {
		t.Fatalf("status = %d (%s), want 200", r.status, r.body)
	}
	var out rpc
	if err := json.NewDecoder(bytes.NewReader([]byte(r.body))).Decode(&out); err != nil {
		t.Fatalf("the answer %q is no JSON-RPC message: %v", r.body, err)
	}
	return out
}

// TestEndpointRefusesBeforeAnyTool checks the door of /mcp (D-213, D-212): a
// request is answered only with the switch and the toggle on and a script
// token the owner allowed to use MCP, with the status each refusal has and no
// tool, token or data in the answer, and the toolbox never reached.
func TestEndpointRefusesBeforeAnyTool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		authorization string
		prepare       func(f *fixture)
		wantStatus    int
		wantChallenge string
	}{
		{name: "the switch or the toggle off", authorization: "Bearer " + secretA, prepare: func(f *fixture) { f.gate.set(false) }, wantStatus: http.StatusNotFound},
		{name: "off, and no token either", prepare: func(f *fixture) { f.gate.set(false) }, wantStatus: http.StatusNotFound},
		{name: "the switch cannot be checked", authorization: "Bearer " + secretA, prepare: func(f *fixture) { f.gate.err = errDBDown }, wantStatus: http.StatusServiceUnavailable},
		{name: "no Authorization header", wantStatus: http.StatusUnauthorized, wantChallenge: "Bearer"},
		{name: "another scheme", authorization: "Basic " + secretA, wantStatus: http.StatusUnauthorized, wantChallenge: "Bearer"},
		{name: "an empty bearer token", authorization: "Bearer  ", wantStatus: http.StatusUnauthorized, wantChallenge: "Bearer"},
		{name: "an unknown token", authorization: "Bearer ofs_ffffffffffffffff_nope", wantStatus: http.StatusUnauthorized, wantChallenge: `Bearer error="invalid_token"`},
		{name: "a session token", authorization: "Bearer ofss_synthetic", wantStatus: http.StatusUnauthorized, wantChallenge: `Bearer error="invalid_token"`},
		{name: "a token without the MCP flag", authorization: "Bearer " + noMCP, wantStatus: http.StatusForbidden},
		{name: "the token cannot be checked", authorization: "Bearer ofs_broken", wantStatus: http.StatusInternalServerError},
	}
	for _, tc := range tests {
		for _, message := range []string{initialize, listTools, callTool("allocation", `{}`)} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				f := newFixture(t)
				if tc.prepare != nil {
					tc.prepare(f)
				}
				f.checkRefused(t, f.post(t, tc.authorization, message), tc.wantStatus, tc.wantChallenge)
			})
		}
	}
}

// checkRefused checks that got is a refusal of the status and challenge,
// with no tool, token or card in its body, that reached no tool and wrote no
// audit entry.
func (f *fixture) checkRefused(t *testing.T, got reply, wantStatus int, wantChallenge string) {
	t.Helper()
	if got.status != wantStatus || got.challenge != wantChallenge {
		t.Errorf("status, challenge = %d, %q; want %d, %q", got.status, got.challenge, wantStatus, wantChallenge)
	}
	for _, leak := range []string{"allocation", "value", "tools", secretA, noMCP, cardText} {
		if strings.Contains(got.body, leak) {
			t.Errorf("the refusal's body %q mentions %q", got.body, leak)
		}
	}
	if len(f.toolbox.ran()) != 0 {
		t.Error("a refused request reached a tool")
	}
	// Only the refusal of a valid token not allowed to use MCP is an entry, by that token, naming no tool.
	var want []mcpserver.Call
	if wantStatus == http.StatusForbidden {
		want = []mcpserver.Call{{Token: idNoMCP, Outcome: mcpserver.OutcomeNotAllowed}}
	}
	if diff := cmp.Diff(want, f.auditor.recorded()); diff != "" {
		t.Errorf("the audit log after a %d (-want +got):\n%s", wantStatus, diff)
	}
	if counted := f.unauthenticated.Load(); (counted == 1) != (wantStatus == http.StatusUnauthorized) || counted > 1 {
		t.Errorf("a %d was counted as unauthenticated %d times", wantStatus, counted)
	}
}

// TestTurningOffStopsAnswersAtOnce checks that the switch is read at every
// request: the same token that was answered is refused the moment the owner
// turns the MCP server off, and answered again when it is back on.
func TestTurningOffStopsAnswersAtOnce(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	auth := "Bearer " + secretA
	for i, step := range []struct {
		on   bool
		want int
	}{{true, 200}, {false, 404}, {false, 404}, {true, 200}} {
		f.gate.set(step.on)
		if got := f.post(t, auth, listTools).status; got != step.want {
			t.Errorf("step %d (on=%v): status = %d, want %d", i, step.on, got, step.want)
		}
	}
}

// TestMCPSession speaks MCP JSON-RPC to the endpoint, the way a client does:
// initialize, tools/list and tools/call, as the one token.
func TestMCPSession(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	auth := "Bearer " + secretA

	init := decode(t, f.post(t, auth, initialize))
	if init.Error != nil || init.Result.ServerInfo.Name != "openfief" {
		t.Fatalf("initialize = %+v, want the server's name", init)
	}

	listed := decode(t, f.post(t, auth, listTools))
	checkToolList(t, &listed)

	call := decode(t, f.post(t, auth, callTool("allocation", `{"by":"currency"}`)))
	if call.Error != nil || call.Result.IsError || len(call.Result.Content) != 1 || call.Result.Content[0].Text != cardText {
		t.Fatalf("tools/call = %+v, want the card as text", call)
	}
	wantRead := []readCall{{Secret: secretA, Tool: "allocation", Args: `{"by":"currency"}`}}
	if diff := cmp.Diff(wantRead, f.toolbox.ran()); diff != "" {
		t.Errorf("the tool ran (-want +got):\n%s", diff)
	}
	wantAudit := []mcpserver.Call{{Token: idA, Tool: "allocation", Outcome: mcpserver.OutcomeRan}}
	if diff := cmp.Diff(wantAudit, f.auditor.recorded()); diff != "" {
		t.Errorf("the audit log (-want +got):\n%s", diff)
	}
}

// checkToolList checks the tools a client is told: the toolbox's, each
// described, with an object schema, and marked read-only and not destructive.
func checkToolList(t *testing.T, list *rpc) {
	t.Helper()
	names := make([]string, 0, len(list.Result.Tools))
	for _, tool := range list.Result.Tools {
		names = append(names, tool.Name)
		if !tool.Annotations.ReadOnlyHint || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
			t.Errorf("tool %s is not marked read-only and not destructive: %+v", tool.Name, tool.Annotations)
		}
		if tool.Description == "" || tool.InputSchema["type"] != "object" {
			t.Errorf("tool %s lacks a description or an object schema: %+v", tool.Name, tool)
		}
	}
	if diff := cmp.Diff([]string{"allocation", "value"}, names); diff != "" {
		t.Errorf("tools/list (-want +got):\n%s", diff)
	}
}

// TestEachCallReadsAsItsOwnToken checks that two tokens' calls reach the
// toolbox with their own secrets and are entered under their own IDs, so the
// read API filters each by its own grants (D-209).
func TestEachCallReadsAsItsOwnToken(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	for _, secret := range []string{secretA, secretB, secretA} {
		decode(t, f.post(t, "Bearer "+secret, callTool("value", `{}`)))
	}
	want := []readCall{{secretA, "value", "{}"}, {secretB, "value", "{}"}, {secretA, "value", "{}"}}
	if diff := cmp.Diff(want, f.toolbox.ran()); diff != "" {
		t.Errorf("the tools ran (-want +got):\n%s", diff)
	}
	recorded := f.auditor.recorded()
	ids := make([]string, 0, len(recorded))
	for _, c := range recorded {
		ids = append(ids, c.Token)
	}
	if diff := cmp.Diff([]string{idA, idB, idA}, ids); diff != "" {
		t.Errorf("the audit entries' tokens (-want +got):\n%s", diff)
	}
}

// TestCallOutcomesAreAudited checks that a call whose arguments were refused
// or whose data could not be read is an entry with that outcome and an error
// result, an unknown tool is a protocol error with no entry, and that the
// entry holds no argument.
func TestCallOutcomesAreAudited(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	auth := "Bearer " + secretA
	refused := decode(t, f.post(t, auth, callTool("allocation", `{"by":"refuse 1234.56 PLN"}`)))
	if !refused.Result.IsError {
		t.Errorf("a refused call: result = %+v, want an error result", refused.Result)
	}
	broken := decode(t, f.post(t, auth, callTool("value", `{"break":true}`)))
	if !broken.Result.IsError {
		t.Errorf("a failed read: result = %+v, want an error result", broken.Result)
	}
	unknown := decode(t, f.post(t, auth, callTool("delete_everything", `{}`)))
	if unknown.Error == nil {
		t.Errorf("an unknown tool: %+v, want a protocol error", unknown)
	}
	want := []mcpserver.Call{
		{Token: idA, Tool: "allocation", Outcome: mcpserver.OutcomeBadArguments},
		{Token: idA, Tool: "value", Outcome: mcpserver.OutcomeReadFailed},
	}
	if diff := cmp.Diff(want, f.auditor.recorded()); diff != "" {
		t.Errorf("the audit log (-want +got):\n%s", diff)
	}
	if got := len(f.toolbox.ran()); got != 2 {
		t.Errorf("the toolbox ran %d calls, want the two known tools only", got)
	}
}

// TestACallTheLogCannotRecordShowsNothing checks that the audit log is
// fail-closed (D-211): when the entry can't be written the client gets an
// error result and none of the card.
func TestACallTheLogCannotRecordShowsNothing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.auditor.fail = true
	got := f.post(t, "Bearer "+secretA, callTool("allocation", `{}`))
	res := decode(t, got)
	if !res.Result.IsError {
		t.Errorf("result = %+v, want an error result", res.Result)
	}
	if strings.Contains(got.body, "Synthetic allocation") {
		t.Errorf("the answer %q shows the card although the call was not recorded", got.body)
	}
}

// TestOnlyPostIsAnswered checks that the stateless endpoint offers no stream
// to open and no session to delete.
func TestOnlyPostIsAnswered(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	for _, method := range []string{http.MethodGet, http.MethodDelete, http.MethodPut} {
		req, err := http.NewRequestWithContext(t.Context(), method, f.url, http.NoBody)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+secretA)
		req.Header.Set("Accept", "application/json, text/event-stream")
		if got := do(t, req); got.status != http.StatusMethodNotAllowed {
			t.Errorf("%s: status = %d, want 405", method, got.status)
		}
	}
}

// TestNewNeedsItsPorts checks that an endpoint without a port is refused.
func TestNewNeedsItsPorts(t *testing.T) {
	t.Parallel()
	if _, err := mcpserver.New(&mcpserver.Config{}); !errors.Is(err, mcpserver.ErrMissingDependency) {
		t.Errorf("New() error = %v, want ErrMissingDependency", err)
	}
}

// bearerTransport adds a script token to every request a client makes.
type bearerTransport struct{ secret string }

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.secret)
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}
	return resp, nil
}

// checkSDKCalls lists the tools and calls one through the SDK's client session.
func checkSDKCalls(t *testing.T, session *sdk.ClientSession) {
	t.Helper()
	tools, err := session.ListTools(t.Context(), nil)
	if err != nil || len(tools.Tools) != 2 {
		t.Fatalf("ListTools = %v, %v; want the two tools", tools, err)
	}
	res, err := session.CallTool(t.Context(), &sdk.CallToolParams{Name: "allocation", Arguments: map[string]any{"by": "currency"}})
	if err != nil || res.IsError || len(res.Content) != 1 {
		t.Fatalf("CallTool = %v, %v; want the card", res, err)
	}
	if text, ok := res.Content[0].(*sdk.TextContent); !ok || text.Text != cardText {
		t.Errorf("content = %+v, want the card as text", res.Content[0])
	}
}

// TestTheSDKsOwnClientWorks connects the official SDK's client, which does
// the handshake a real agent does, and lists and calls a tool.
func TestTheSDKsOwnClientWorks(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	client := sdk.NewClient(&sdk.Implementation{Name: "synthetic-client", Version: "0"}, nil)
	session, err := client.Connect(t.Context(), &sdk.StreamableClientTransport{
		Endpoint: f.url, HTTPClient: &http.Client{Transport: bearerTransport{secret: secretB}}, DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := session.Close(); closeErr != nil {
			t.Errorf("Close: %v", closeErr)
		}
	})
	checkSDKCalls(t, session)
	if got := f.auditor.recorded(); len(got) != 1 || got[0].Token != idB {
		t.Errorf("the audit log = %+v, want one entry under the client's token", got)
	}
}
