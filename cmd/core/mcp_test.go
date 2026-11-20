package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/plugins"
	viewsanswer "github.com/arkadiuss/openfief/internal/views/app/answer"
	"github.com/arkadiuss/openfief/internal/views/mcpserver"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// mcpReply is the core's HTTP answer to one JSON-RPC message.
type mcpReply struct {
	status    int
	challenge string
	body      []byte
}

// mcpPost sends one JSON-RPC message to the core's /mcp with token (none when
// empty).
func mcpPost(t *testing.T, srv *httptest.Server, token, message string) mcpReply {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+mcpPath, strings.NewReader(message))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", mcpPath, err)
	}
	body, readErr := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); readErr != nil || closeErr != nil {
		t.Fatalf("reading the answer: %v %v", readErr, closeErr)
	}
	return mcpReply{status: resp.StatusCode, challenge: resp.Header.Get("WWW-Authenticate"), body: body}
}

// The JSON-RPC messages of a client's session.
const (
	mcpInitialize = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"synthetic-client","version":"0"}}}`
	mcpListTools  = `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
)

func mcpCall(name, args string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, name, args)
}

// mcpAnswer is the part of a JSON-RPC answer the test reads.
type mcpAnswer struct {
	Result struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// decode returns r as a JSON-RPC answer, failing unless it is a 200.
func (r mcpReply) decode(t *testing.T) mcpAnswer {
	t.Helper()
	if r.status != http.StatusOK {
		t.Fatalf("status = %d (%s), want 200", r.status, r.body)
	}
	var a mcpAnswer
	if err := json.Unmarshal(r.body, &a); err != nil {
		t.Fatalf("the answer %q is no JSON-RPC message: %v", r.body, err)
	}
	return a
}

// queryer is what a pool and a connection both do.
type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// auditText returns the first column of query's rows, one string per row.
func auditText(ctx context.Context, t *testing.T, db queryer, query string) []string {
	t.Helper()
	rows, err := db.Query(ctx, query)
	if err != nil {
		t.Fatalf("reading the audit log: %v", err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("reading the audit log: %v", err)
	}
	return got
}

// mcpFixture is the real core on a database with the demo data and two script
// tokens, one of which may use MCP.
type mcpFixture struct {
	core        *core
	srv         *httptest.Server
	pool        *pgxpool.Pool
	tokens      *identityapp.Tokens
	agent       identity.ScriptToken
	agentSecret string
	plainSecret string
	// narrowSecret is a token allowed to use MCP and granted the XTB account alone.
	narrowSecret  string
	plain, narrow identity.ScriptToken
}

func newMCPFixture(t *testing.T) *mcpFixture {
	t.Helper()
	ctx := t.Context()
	c, srv, pool := startCore(t, emptyDatabase(t))
	ds := loadDemo(t, pool)
	toks := identityapp.NewTokens(identitypg.New(pool), coreNow, nil)
	agent, agentSecret, err := toks.IssueWith(ctx, "agent", time.Hour, true)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	plain, plainSecret, err := toks.IssueWith(ctx, "script", time.Hour, false)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	narrow, narrowSecret, err := toks.IssueWith(ctx, "narrow", time.Hour, true)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	// Grants are the truth under the policy too (D-209): the agent holds the household, the narrow token one account.
	for token, target := range map[identity.ScriptTokenID]identity.GrantTarget{
		agent.ID:  identity.Household(),
		narrow.ID: {Kind: identity.GrantTargetAccount, ID: string(demo.XTBAccount)},
	} {
		holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(token)}
		if _, err := c.access.grants.Create(ctx, adminActor(), holder, target); err != nil {
			t.Fatalf("granting: %v", err)
		}
	}
	waitValued(t, srv, agentSecret, ds)
	return &mcpFixture{core: c, srv: srv, pool: pool, tokens: toks, agent: agent, agentSecret: agentSecret, plainSecret: plainSecret, narrowSecret: narrowSecret, plain: plain, narrow: narrow}
}

// checkGrantsLimitWhatIsRead checks that a token granted one account reads that
// account and nothing else, over the REST read API and over MCP, while the
// household's token reads both (D-209).
func (f *mcpFixture) checkGrantsLimitWhatIsRead(t *testing.T) {
	t.Helper()
	ids := func(token string) []string {
		var accounts viewsv1.ListAccountsResponse
		getJSON(t, f.srv, "/v1/accounts", token, &accounts)
		out := make([]string, 0, len(accounts.GetAccounts()))
		for _, a := range accounts.GetAccounts() {
			out = append(out, a.GetId())
		}
		slices.Sort(out)
		return out
	}
	if got := ids(f.narrowSecret); !slices.Equal(got, []string{string(demo.XTBAccount)}) {
		t.Errorf("the read API lists %v to a token granted the XTB account alone, want just it", got)
	}
	if got := ids(f.agentSecret); len(got) < 2 {
		t.Errorf("the read API lists %v to the household's token, want every account", got)
	}
	value := func(token string) string {
		call := mcpPost(t, f.srv, token, mcpCall("value", `{}`)).decode(t)
		if call.Result.IsError || len(call.Result.Content) != 1 {
			t.Fatalf("value as %s = %+v", token[:8], call.Result)
		}
		return call.Result.Content[0].Text
	}
	if narrow, household := value(f.narrowSecret), value(f.agentSecret); narrow == household {
		t.Errorf("MCP's value is the same for the narrow and the household token: %s", narrow)
	}
	// What is tied to no account is the owner's alone, over MCP too.
	needs := mcpPost(t, f.srv, f.agentSecret, mcpCall("what_needs_me", `{}`)).decode(t)
	if needs.Result.IsError {
		t.Errorf("what_needs_me = %+v", needs.Result)
	}
}

// checkOffByDefault checks that no endpoint answers while the switch or the
// MCP toggle is off, whatever the token, then turns both on.
func (f *mcpFixture) checkOffByDefault(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	if r := mcpPost(t, f.srv, f.agentSecret, mcpInitialize); r.status != http.StatusNotFound || strings.Contains(string(r.body), "openfief") {
		t.Fatalf("with the switch off: %d %q, want 404 and nothing", r.status, r.body)
	}
	// The switch alone is not enough: the MCP toggle is its own.
	if err := f.core.agents.store.SaveEnabled(ctx, true, nil); err != nil {
		t.Fatalf("SaveEnabled: %v", err)
	}
	if r := mcpPost(t, f.srv, f.agentSecret, mcpInitialize); r.status != http.StatusNotFound {
		t.Fatalf("with the switch on and the MCP toggle off: %d, want 404", r.status)
	}
	if err := f.core.agents.store.SaveToggle(ctx, modelsdomain.FeatureMCP, true, nil); err != nil {
		t.Fatalf("SaveToggle: %v", err)
	}
}

// checkDoor checks the statuses of the refusals, none with a tool list.
func (f *mcpFixture) checkDoor(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		name, token   string
		wantStatus    int
		wantChallenge string
	}{
		{name: "no token", wantStatus: http.StatusUnauthorized, wantChallenge: "Bearer"},
		{name: "a token that is none", token: "ofs_0000000000000000_" + strings.Repeat("A", 43), wantStatus: http.StatusUnauthorized, wantChallenge: `Bearer error="invalid_token"`},
		{name: "a token without the MCP flag", token: f.plainSecret, wantStatus: http.StatusForbidden},
	} {
		r := mcpPost(t, f.srv, tc.token, mcpListTools)
		if r.status != tc.wantStatus || r.challenge != tc.wantChallenge || strings.Contains(string(r.body), "allocation") {
			t.Errorf("%s: %d, %q, %q; want %d, %q and no tool list", tc.name, r.status, r.challenge, r.body, tc.wantStatus, tc.wantChallenge)
		}
	}
}

// checkTools checks that tools/list names the Ask box's tools.
func (f *mcpFixture) checkTools(t *testing.T) {
	t.Helper()
	listed := mcpPost(t, f.srv, f.agentSecret, mcpListTools).decode(t).Result.Tools
	got := make([]string, 0, len(listed))
	for _, tool := range listed {
		got = append(got, tool.Name)
	}
	box, err := viewsanswer.NewToolbox(&viewsanswer.ToolboxConfig{Read: viewsv1connect.NewReadServiceClient(http.DefaultClient, "http://core.invalid")})
	if err != nil {
		t.Fatalf("NewToolbox: %v", err)
	}
	asked := box.Tools()
	want := make([]string, 0, len(asked))
	for _, tool := range asked {
		want = append(want, tool.Name)
	}
	slices.Sort(want) // the SDK lists tools by name
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("tools/list is not the Ask box's tools (-want +got):\n%s", diff)
	}
}

// checkSession checks initialize, tools/list as the Ask box's tools and a
// tools/call through the read API, then a call with arguments refused.
func (f *mcpFixture) checkSession(t *testing.T) {
	t.Helper()
	if init := mcpPost(t, f.srv, f.agentSecret, mcpInitialize).decode(t); init.Error != nil {
		t.Fatalf("initialize = %+v", init)
	}
	f.checkTools(t)
	call := mcpPost(t, f.srv, f.agentSecret, mcpCall("allocation", `{"by":"currency"}`))
	if card := call.decode(t); card.Result.IsError || len(card.Result.Content) != 1 || !strings.Contains(card.Result.Content[0].Text, `"rows"`) {
		t.Fatalf("tools/call allocation = %s, want a card with rows", call.body)
	}
	// Arguments the tool does not accept, with an amount in them.
	bad := mcpPost(t, f.srv, f.agentSecret, mcpCall("allocation", `{"by":"12345.67 PLN"}`))
	if refused := bad.decode(t); !refused.Result.IsError || strings.Contains(string(bad.body), "12345.67") {
		t.Errorf("a call with bad arguments = %s, want an error result that quotes nothing", bad.body)
	}
}

// checkAudit checks the calls are entries of usage only: who, which tool, the
// outcome, and nothing a client sent in the whole log.
func (f *mcpFixture) checkAudit(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	calls := auditText(ctx, t, f.pool, `SELECT actor_kind || ' ' || actor_id || ' ' || outcome || ' ' || coalesce(refs->0->>'id', 'no tool')
		FROM identity.audit_log WHERE action = 'mcp_call' ORDER BY id`)
	id, plain, narrow := string(f.agent.ID), string(f.plain.ID), string(f.narrow.ID)
	// In order: the token not allowed is refused, naming no tool; the agent's two calls of the session; then the
	// narrow token's and the agent's reads of the value and the agent's of what needs the owner.
	wantCalls := []string{
		"script_token " + plain + " refused no tool",
		"script_token " + id + " ok allocation", "script_token " + id + " refused allocation",
		"script_token " + narrow + " ok value", "script_token " + id + " ok value", "script_token " + id + " ok what_needs_me",
	}
	if diff := cmp.Diff(wantCalls, calls); diff != "" {
		t.Errorf("the MCP calls in the audit log (-want +got):\n%s", diff)
	}
	everything := strings.Join(auditText(ctx, t, f.pool, `SELECT row_to_json(a)::text FROM identity.audit_log a ORDER BY id`), "\n")
	for _, leak := range []string{"12345", "currency", f.agentSecret, f.plainSecret} {
		if strings.Contains(everything, leak) {
			t.Errorf("the audit log holds %q, which a client sent or the owner keeps secret", leak)
		}
	}
}

// checkStopsAtOnce checks that a revoked token, and a toggle turned off for a
// token that is fine, are refused at once.
func (f *mcpFixture) checkStopsAtOnce(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	if err := f.tokens.Revoke(ctx, f.agent.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if r := mcpPost(t, f.srv, f.agentSecret, mcpListTools); r.status != http.StatusUnauthorized {
		t.Errorf("a revoked token: %d, want 401", r.status)
	}
	_, fresh, err := f.tokens.IssueWith(ctx, "again", time.Hour, true)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	if r := mcpPost(t, f.srv, fresh, mcpListTools); r.status != http.StatusOK {
		t.Fatalf("a new allowed token: %d, want 200", r.status)
	}
	if err = f.core.agents.store.SaveToggle(ctx, modelsdomain.FeatureMCP, false, nil); err != nil {
		t.Fatalf("SaveToggle: %v", err)
	}
	if r := mcpPost(t, f.srv, fresh, mcpListTools); r.status != http.StatusNotFound {
		t.Errorf("after turning the MCP toggle off: %d, want 404", r.status)
	}
}

// TestMCPServedOnTheBrowserPort runs D-213 over the real core and database:
// /mcp answers nothing while the switch or the MCP toggle is off, answers MCP
// JSON-RPC to a script token the owner allowed to use MCP and to no other,
// offers the Ask box's tools, runs them through the read API with the token's
// own identity, enters each call in the audit log as usage only, and stops
// answering at once when the toggle is turned off or the token is revoked.
func TestMCPServedOnTheBrowserPort(t *testing.T) {
	t.Parallel()
	f := newMCPFixture(t)
	f.checkOffByDefault(t)
	f.checkDoor(t)
	f.checkSession(t)
	f.checkGrantsLimitWhatIsRead(t)
	f.checkAudit(t)
	f.checkStopsAtOnce(t)
}

// TestMCPPathBelongsToTheCore checks that /mcp is a path the core serves itself
// (D-81): no view plugin's browser path may claim it, so the front door sends it
// to the core and never to a plugin.
func TestMCPPathBelongsToTheCore(t *testing.T) {
	t.Parallel()
	for _, p := range []string{mcpPath, mcpPath + "/"} {
		if !plugins.CoreServesPath(p) {
			t.Errorf("the core does not serve %s itself", p)
		}
	}
}

var errMCPDown = errors.New("down")

// grantless is an authorizer that does not limit script tokens to their grants.
type grantless struct{ identityapp.Policy }

// fakeSwitch is the models switch with a fixed answer, or none.
type fakeSwitch struct {
	sw  modelsdomain.Switch
	err error
}

func (f fakeSwitch) Switch(context.Context) (modelsdomain.Switch, error) { return f.sw, f.err }

// TestMCPGate checks that the MCP server answers only with the one switch on and
// the MCP toggle under it on (D-212), and that a switch that can't be read is an
// error, never an answer.
func TestMCPGate(t *testing.T) {
	t.Parallel()
	on := func(enabled bool, features ...modelsdomain.Feature) modelsdomain.Switch {
		sw := modelsdomain.Switch{Enabled: enabled, Toggles: map[modelsdomain.Feature]bool{}}
		for _, f := range features {
			sw.Toggles[f] = true
		}
		return sw
	}
	tests := []struct {
		name       string
		sw         fakeSwitch
		authorizer identityapp.Authorizer
		want       bool
		wantErr    bool
	}{
		{name: "the zero switch", sw: fakeSwitch{}},
		{name: "the switch on and no toggle", sw: fakeSwitch{sw: on(true)}},
		{name: "the switch on, other toggles on", sw: fakeSwitch{sw: on(true, modelsdomain.FeatureAsk, modelsdomain.FeatureMaintainer)}},
		{name: "the toggle on, the switch off", sw: fakeSwitch{sw: on(false, modelsdomain.FeatureMCP)}},
		{name: "the switch and the toggle on", sw: fakeSwitch{sw: on(true, modelsdomain.FeatureMCP)}, want: true},
		{name: "on, but the authorizer ignores a token's grants", sw: fakeSwitch{sw: on(true, modelsdomain.FeatureMCP)}, authorizer: grantless{}, wantErr: true},
		{name: "off, and the authorizer ignores grants too: still nothing answers", sw: fakeSwitch{sw: on(false, modelsdomain.FeatureMCP)}, authorizer: grantless{}},
		{name: "the switch cannot be read", sw: fakeSwitch{sw: on(true, modelsdomain.FeatureMCP), err: errMCPDown}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			authorizer := tt.authorizer
			if authorizer == nil {
				authorizer = identityapp.NewGrantPolicy(nil, nil) // the policy under --authz=policy
			}
			got, err := mcpGate{settings: tt.sw, authorizer: authorizer}.Enabled(t.Context())
			if got != tt.want || (err != nil) != tt.wantErr {
				t.Errorf("Enabled() = %v, %v; want %v, error: %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// mcpLog is an auditAppender that keeps what it is given, or fails.
type mcpLog struct {
	entries []identity.AuditEntry
	err     error
}

func (l *mcpLog) AppendAudit(_ context.Context, e *identity.AuditEntry) error {
	if l.err != nil {
		return l.err
	}
	l.entries = append(l.entries, *e)
	return nil
}

// TestMCPAuditEntry checks the audit entry of each outcome of a call (D-211):
// the token is the actor, the tool the one record named, ok when the call ran,
// refused when its arguments were not accepted and failed when its data could
// not be read; and a log that can't write is an error, so the call shows nothing.
func TestMCPAuditEntry(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		outcome mcpserver.Outcome
		want    identity.AuditOutcome
	}{
		{mcpserver.OutcomeRan, identity.OutcomeOK},
		{mcpserver.OutcomeBadArguments, identity.OutcomeRefused},
		{mcpserver.OutcomeReadFailed, identity.OutcomeFailed},
	}
	for _, tt := range tests {
		t.Run(string(tt.outcome), func(t *testing.T) {
			t.Parallel()
			log := &mcpLog{}
			a := mcpAudit{log: log, now: func() time.Time { return at }}
			if err := a.RecordCall(t.Context(), mcpserver.Call{Token: "00000000000000a1", Tool: "value", Outcome: tt.outcome}); err != nil {
				t.Fatalf("RecordCall: %v", err)
			}
			want := []identity.AuditEntry{{
				At: at, Actor: identity.Actor{Kind: identity.ActorScriptToken, ID: "00000000000000a1"}, Action: identity.ActionMCPCall,
				Outcome: tt.want, Refs: []identity.Ref{{Kind: identity.RefMCPTool, ID: "value"}},
			}}
			if diff := cmp.Diff(want, log.entries); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
	if err := (mcpAudit{log: &mcpLog{err: errMCPDown}, now: func() time.Time { return at }}).RecordCall(t.Context(), mcpserver.Call{Token: "00000000000000a1", Tool: "value", Outcome: mcpserver.OutcomeRan}); !errors.Is(err, errMCPDown) {
		t.Errorf("RecordCall on a log that is down = %v, want its error", err)
	}
	if err := (mcpAudit{log: &mcpLog{}, now: func() time.Time { return at }}).RecordCall(t.Context(), mcpserver.Call{Token: "not-an-id", Tool: "value", Outcome: mcpserver.OutcomeRan}); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("RecordCall for a token that is no ID = %v, want ErrInvalidAuditEntry", err)
	}
}

// TestMCPTokens checks how the core's token check is told to the MCP server:
// a valid allowed token is its ID, one the owner did not allow is not allowed,
// and anything the token check calls invalid is invalid.
func TestMCPTokens(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, _, pool := startCore(t, dsn)
	toks := identityapp.NewTokens(identitypg.New(pool), coreNow, nil)
	allowed, allowedSecret, err := toks.IssueWith(t.Context(), "agent", time.Hour, true)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	_, plainSecret, err := toks.IssueWith(t.Context(), "script", time.Hour, false)
	if err != nil {
		t.Fatalf("IssueWith: %v", err)
	}
	check := mcpTokens{tokens: toks}
	if id, err := check.Authenticate(t.Context(), allowedSecret); err != nil || id != string(allowed.ID) {
		t.Errorf("Authenticate(allowed) = %q, %v; want %q", id, err, allowed.ID)
	}
	if _, err := check.Authenticate(t.Context(), plainSecret); !errors.Is(err, mcpserver.ErrNotAllowed) || errors.Is(err, mcpserver.ErrInvalidToken) {
		t.Errorf("Authenticate(no flag) = %v, want ErrNotAllowed only", err)
	}
	for _, secret := range []string{"", "nope", "ofss_0000000000000000_" + strings.Repeat("A", 43)} {
		if _, err := check.Authenticate(t.Context(), secret); !errors.Is(err, mcpserver.ErrInvalidToken) {
			t.Errorf("Authenticate(%q) = %v, want ErrInvalidToken", secret, err)
		}
	}
}
