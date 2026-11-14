package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

var (
	errBoom      = errors.New("a failure of no class")
	errSpiceDown = errors.New("spicedb down")
)

// providerFake is the only model provider of a route: it answers with content,
// or fails with err, and keeps the requests it was sent.
type providerFake struct {
	content string
	err     error
	entry   modelsdomain.ModelEntry
	sent    []modelsdomain.Request
}

func (p *providerFake) Instance(context.Context, string) error { return nil }
func (p *providerFake) Describe(context.Context, string) (modelsdomain.ProviderInstance, error) {
	return modelsdomain.ProviderInstance{Location: modelsdomain.LocationOwnerHardware}, nil
}

func (p *providerFake) Models(context.Context, string) ([]modelsdomain.ModelEntry, error) {
	return []modelsdomain.ModelEntry{p.entry}, nil
}

func (p *providerFake) Generate(_ context.Context, _, _ string, req *modelsdomain.Request) (modelsdomain.Result, error) {
	p.sent = append(p.sent, *req)
	if p.err != nil {
		return modelsdomain.Result{}, p.err
	}
	return modelsdomain.Result{Content: p.content, Finish: modelsdomain.FinishStop}, nil
}

// callLog keeps the router's audit entries by their kind and role.
type callLog struct {
	sent    []modelsapp.CallSent
	results []modelsapp.CallResult
}

func (l *callLog) RecordSent(_ context.Context, c *modelsapp.CallSent) error {
	l.sent = append(l.sent, *c)
	return nil
}

func (l *callLog) RecordResult(_ context.Context, c *modelsapp.CallResult) error {
	l.results = append(l.results, *c)
	return nil
}

type noRaise struct{}

func (noRaise) Raise(context.Context, string, string, string) error { return nil }
func (noRaise) Clear(context.Context, string, string) error         { return nil }

// maintainerRoute is the settings of a maintainer whose route is one model,
// with the switch and the toggle as asked and the data tiers allowed to it.
func maintainerRoute(t *testing.T, on bool, tiers ...modelsdomain.DataTier) modelsdomain.Settings {
	t.Helper()
	ref := modelsdomain.Ref{Instance: "local", Model: "synth"}
	s := modelsdomain.Settings{
		Switch: modelsdomain.Switch{Enabled: on, Toggles: map[modelsdomain.Feature]bool{modelsdomain.FeatureMaintainer: on}},
		Routes: map[modelsdomain.Role]modelsdomain.Route{modelsdomain.RoleMaintainer: {Role: modelsdomain.RoleMaintainer, Models: []modelsdomain.Ref{ref}}},
		Policy: modelsdomain.DataTierPolicy{},
	}
	for _, tier := range tiers {
		s.Policy[tier] = []modelsdomain.Ref{ref}
	}
	return s
}

func newRouterFor(settings modelsdomain.Settings, p *providerFake, log *callLog) *modelsapp.Router {
	return modelsapp.NewRouter(&modelsapp.RouterPorts{
		Settings: ownerlessSettings{s: settings}, Providers: p, Audit: log, Attention: noRaise{},
		CallIDs: func() (string, error) { return "00112233445566aa", nil },
		Now:     func() time.Time { return time.Date(2026, 11, 13, 6, 30, 0, 0, time.UTC) },
	})
}

// askedRouter is a router over a provider that answers, and its audit log.
func askedRouter(t *testing.T) (maintainerAsker, *providerFake, *callLog) {
	t.Helper()
	provider := &providerFake{content: `{"note":"ok","changes":[]}`, entry: modelsdomain.ModelEntry{ID: "synth", StructuredOut: true}}
	log := &callLog{}
	return maintainerAsker{router: newRouterFor(maintainerRoute(t, true, modelsdomain.DataTierT0, modelsdomain.DataTierT2), provider, log)}, provider, log
}

func TestMaintainerAskerSendsFiguresAsT0AndInstructionsAsT2(t *testing.T) {
	t.Parallel()
	asker, provider, _ := askedRouter(t)
	res, err := asker.Ask(t.Context(), &portfolioapp.AskRequest{Instructions: "Instructions.", Figures: "Figures.", Schema: `{"type":"object"}`})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if res.Content != provider.content || res.Model != "local/synth" {
		t.Errorf("result = %+v, want the provider's content from local/synth", res)
	}
	if len(provider.sent) != 1 {
		t.Fatalf("the provider was sent %d requests, want one", len(provider.sent))
	}
	got := provider.sent[0]
	want := []modelsdomain.Message{
		{Role: modelsdomain.MessageSystem, Content: "Instructions.", DataTier: modelsdomain.DataTierT2},
		{Role: modelsdomain.MessageUser, Content: "Figures.", DataTier: modelsdomain.DataTierT0},
	}
	if diff := cmp.Diff(want, got.Messages); diff != "" {
		t.Errorf("messages (-want +got):\n%s", diff)
	}
	if got.ResponseSchemaJSON != `{"type":"object"}` || len(got.Tools) != 0 {
		t.Errorf("request = %+v, want the schema as structured output and no tools: the maintainer has none", got)
	}
}

func TestMaintainerAskerNamesTheAskingSession(t *testing.T) {
	t.Parallel()
	asker, _, log := askedRouter(t)
	if _, err := asker.Ask(t.Context(), &portfolioapp.AskRequest{Instructions: "i", Figures: "f", Schema: "{}", Actor: "00000000000000a1"}); err != nil {
		t.Fatalf("Ask: %v", err)
	}
	want := modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: "00000000000000a1"}
	if len(log.sent) != 1 || log.sent[0].Role != modelsdomain.RoleMaintainer || log.sent[0].Actor != want {
		t.Errorf("audit sent = %+v, want one entry of the maintainer role for the asking session", log.sent)
	}
}

func TestMaintainerAskerNamesTheSystemWhenTheScheduleAsks(t *testing.T) {
	t.Parallel()
	asker, _, log := askedRouter(t)
	if _, err := asker.Ask(t.Context(), &portfolioapp.AskRequest{Instructions: "i", Figures: "f", Schema: "{}"}); err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if len(log.sent) != 1 || log.sent[0].Actor.Kind != modelsapp.ActorSystem {
		t.Errorf("audit sent = %+v, want the system as actor", log.sent)
	}
}

func TestMaintainerAskerFailuresAreClassesAndNoText(t *testing.T) {
	t.Parallel()
	structured := modelsdomain.ModelEntry{ID: "synth", StructuredOut: true}
	tests := []struct {
		name      string
		settings  modelsdomain.Settings
		provider  *providerFake
		wantClass string
		wantCalls int
	}{
		{
			name: "the switch off sends nothing", settings: maintainerRoute(t, false, modelsdomain.DataTierT0, modelsdomain.DataTierT2),
			provider: &providerFake{entry: structured}, wantClass: "denied",
		},
		{
			name: "no model allowed for the figures' tier", settings: maintainerRoute(t, true, modelsdomain.DataTierT2),
			provider: &providerFake{entry: structured}, wantClass: "denied",
		},
		{
			name: "no model allowed for the instructions' tier", settings: maintainerRoute(t, true, modelsdomain.DataTierT0),
			provider: &providerFake{entry: structured}, wantClass: "denied",
		},
		{
			name: "no tier allowed at all: deny by default", settings: maintainerRoute(t, true),
			provider: &providerFake{entry: structured}, wantClass: "denied",
		},
		{
			name: "a model that has no structured output", settings: maintainerRoute(t, true, modelsdomain.DataTierT0, modelsdomain.DataTierT2),
			provider: &providerFake{entry: modelsdomain.ModelEntry{ID: "synth"}}, wantClass: "denied",
		},
		{
			name: "no route", settings: func() modelsdomain.Settings {
				s := maintainerRoute(t, true, modelsdomain.DataTierT0, modelsdomain.DataTierT2)
				s.Routes = nil
				return s
			}(),
			provider: &providerFake{entry: structured}, wantClass: "denied",
		},
		{
			name: "every model down", settings: maintainerRoute(t, true, modelsdomain.DataTierT0, modelsdomain.DataTierT2),
			provider:  &providerFake{entry: structured, err: fmt.Errorf("%w: ETFA 63.1", modelsapp.ErrProviderUnavailable)},
			wantClass: "unavailable", wantCalls: 1,
		},
		{
			name: "a provider that rejects the request", settings: maintainerRoute(t, true, modelsdomain.DataTierT0, modelsdomain.DataTierT2),
			provider:  &providerFake{entry: structured, err: modelsapp.ErrProviderRejected},
			wantClass: "rejected", wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			asker := maintainerAsker{router: newRouterFor(tt.settings, tt.provider, &callLog{})}
			_, err := asker.Ask(t.Context(), &portfolioapp.AskRequest{Instructions: "i", Figures: "f", Schema: "{}"})
			ae, ok := errors.AsType[*portfolioapp.AskError](err)
			if !ok || ae.Class != tt.wantClass {
				t.Fatalf("Ask error = %v, want class %q", err, tt.wantClass)
			}
			if strings.Contains(err.Error(), "ETFA") || strings.Contains(err.Error(), "63.1") {
				t.Errorf("the error quotes what the provider or the figures said: %v", err)
			}
			if len(tt.provider.sent) != tt.wantCalls {
				t.Errorf("the provider was sent %d requests, want %d", len(tt.provider.sent), tt.wantCalls)
			}
		})
	}
}

func TestAskClass(t *testing.T) {
	t.Parallel()
	tests := []struct {
		err  error
		want string
	}{
		{modelsdomain.ErrSwitchOff, "denied"},
		{modelsdomain.ErrFeatureOff, "denied"},
		{modelsdomain.ErrNoRoute, "denied"},
		{modelsdomain.ErrDataTierDenied, "denied"},
		{modelsdomain.ErrNoCapableModel, "denied"},
		{modelsdomain.ErrUntagged, "denied"},
		{fmt.Errorf("%w: %w", modelsapp.ErrNoAnswer, modelsapp.ErrProviderUnavailable), "unavailable"},
		{fmt.Errorf("%w: %w", modelsapp.ErrNoAnswer, modelsapp.ErrProviderMissing), "misconfigured"},
		{modelsapp.ErrNotModelProvider, "misconfigured"},
		{modelsapp.ErrProviderRejected, "rejected"},
		{context.DeadlineExceeded, "unavailable"},
		{context.Canceled, "canceled"},
		{modelsapp.ErrAuditUnavailable, "internal"},
		{errBoom, "internal"},
	}
	for _, tt := range tests {
		if got := askClass(tt.err); got != tt.want {
			t.Errorf("askClass(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestMaintainerGateIsTheSwitchAndTheToggle(t *testing.T) {
	t.Parallel()
	feature := func(enabled bool, toggles ...modelsdomain.Feature) modelsdomain.Settings {
		sw := modelsdomain.Switch{Enabled: enabled, Toggles: map[modelsdomain.Feature]bool{}}
		for _, f := range toggles {
			sw.Toggles[f] = true
		}
		return modelsdomain.Settings{Switch: sw}
	}
	tests := []struct {
		name string
		s    modelsdomain.Settings
		want bool
	}{
		{"nothing set", modelsdomain.Settings{}, false},
		{"the switch on, the toggle off", feature(true), false},
		{"another feature's toggle on", feature(true, modelsdomain.FeatureAsk), false},
		{"the toggle on, the switch off", feature(false, modelsdomain.FeatureMaintainer), false},
		{"both on", feature(true, modelsdomain.FeatureMaintainer), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := maintainerGate{settings: ownerlessSettings{s: tt.s}}.MaintainerOn(t.Context())
			if err != nil || got != tt.want {
				t.Errorf("MaintainerOn = %v, %v; want %v", got, err, tt.want)
			}
		})
	}
}

// heldFake answers the views' held values, and keeps the caller it was asked
// as.
type heldFake struct {
	held   viewsapp.HeldValues
	err    error
	caller identity.Caller
}

func (h *heldFake) HeldValues(ctx context.Context) (viewsapp.HeldValues, error) {
	h.caller = identityapp.CallerFrom(ctx)
	return h.held, h.err
}

// heldCase is one answer of the views and what the maintainer makes of it.
type heldCase struct {
	name    string
	fake    *heldFake
	wantErr error
}

// checkHoldings fails the test unless the maintainer's holdings are what c wants.
func checkHoldings(t *testing.T, c *heldCase, day kernel.Date) {
	t.Helper()
	got, err := maintainerHoldings{views: c.fake}.Holdings(t.Context())
	if !errors.Is(err, c.wantErr) {
		t.Fatalf("Holdings error = %v, want %v", err, c.wantErr)
	}
	if c.fake.caller.Holder != identity.HolderOwner {
		t.Errorf("the views were read as %v, want the owner: the weights are of the whole portfolio", c.fake.caller.Holder)
	}
	if c.wantErr != nil {
		return
	}
	if !got.AsOf.Equal(day) || got.Currency.String() != "PLN" || got.Total.String() != "10000" || got.Values["IETEST000026/ETFA"].String() != "6300" {
		t.Errorf("Holdings = %+v, want the views' values in PLN", got)
	}
}

func TestMaintainerHoldingsReadAsTheOwnerAndRefuseAnIncompleteValuation(t *testing.T) {
	t.Parallel()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	day, err := kernel.ParseDate("2026-11-12")
	if err != nil {
		t.Fatal(err)
	}
	pl := func(s string) kernel.Money {
		m, parseErr := kernel.ParseMoney(s, pln)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		return m
	}
	complete := viewsapp.HeldValues{On: day, Total: pl("10000"), ByInstrument: map[string]kernel.Money{"IETEST000026/ETFA": pl("6300")}}
	incomplete := complete
	incomplete.Unvalued = 1
	tests := []heldCase{
		{name: "a complete valuation", fake: &heldFake{held: complete}},
		{name: "nothing recorded", fake: &heldFake{err: viewsapp.ErrNothingHeld}, wantErr: portfolioapp.ErrNoHoldings},
		{name: "a holding without a price skews every weight", fake: &heldFake{held: incomplete}, wantErr: portfolioapp.ErrNoHoldings},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkHoldings(t, &tt, day)
		})
	}
}

func TestMaintainerAuditWritesTheEntriesOfD211(t *testing.T) {
	t.Parallel()
	system := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	owner := identity.Actor{Kind: identity.ActorOwner, ID: "00000000000000a1"}
	run := identity.Ref{Kind: identity.RefMaintainerRun, ID: "00112233445566aa"}
	role := identity.Ref{Kind: identity.RefModelRole, ID: "maintainer"}
	at := time.Date(2026, 11, 13, 6, 30, 0, 0, time.UTC)
	tests := []struct {
		name string
		r    portfolioapp.AuditRecord
		want identity.AuditEntry
	}{
		{
			name: "a run with a note", r: portfolioapp.AuditRecord{Event: portfolioapp.AuditRun, Outcome: "ok", RunID: run.ID, At: at},
			want: identity.AuditEntry{At: at, Actor: system, Action: identity.ActionMaintainerRun, Outcome: identity.OutcomeOK, Refs: []identity.Ref{role, run}},
		},
		{
			name: "a run the router refused", r: portfolioapp.AuditRecord{Event: portfolioapp.AuditRun, Outcome: "refused", RunID: run.ID, Class: "denied", At: at},
			want: identity.AuditEntry{
				At: at, Actor: system, Action: identity.ActionMaintainerRun, Outcome: identity.OutcomeRefused,
				Refs: []identity.Ref{role, run, {Kind: identity.RefModelError, ID: "denied"}},
			},
		},
		{
			name: "a proposal dropped", r: portfolioapp.AuditRecord{Event: portfolioapp.AuditDropped, Outcome: "refused", RunID: run.ID, At: at},
			want: identity.AuditEntry{At: at, Actor: system, Action: identity.ActionMaintainerProposalDropped, Outcome: identity.OutcomeRefused, Refs: []identity.Ref{run}},
		},
		{
			name: "a proposal applied by the owner", r: portfolioapp.AuditRecord{Event: portfolioapp.AuditApplied, Outcome: "ok", RunID: run.ID, Actor: owner.ID, At: at},
			want: identity.AuditEntry{At: at, Actor: owner, Action: identity.ActionMaintainerProposalApplied, Outcome: identity.OutcomeOK, Refs: []identity.Ref{run}},
		},
		{
			name: "a proposal dismissed by the owner", r: portfolioapp.AuditRecord{Event: portfolioapp.AuditDismissed, Outcome: "ok", RunID: run.ID, Actor: owner.ID, At: at},
			want: identity.AuditEntry{At: at, Actor: owner, Action: identity.ActionMaintainerProposalDismissed, Outcome: identity.OutcomeOK, Refs: []identity.Ref{run}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := &recordingLog{}
			if err := (maintainerAudit{log: log}).Record(t.Context(), &tt.r); err != nil {
				t.Fatalf("Record: %v", err)
			}
			if diff := cmp.Diff([]identity.AuditEntry{tt.want}, log.entries); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
	t.Run("a log that is down fails the record", func(t *testing.T) {
		t.Parallel()
		err := (maintainerAudit{log: &recordingLog{err: errLogDown}}).Record(t.Context(), &portfolioapp.AuditRecord{Event: portfolioapp.AuditDismissed, Outcome: "ok", RunID: run.ID, Actor: owner.ID, At: at})
		if !errors.Is(err, errLogDown) {
			t.Errorf("Record error = %v, want the log's", err)
		}
	})
	t.Run("an invalid run ID is no entry", func(t *testing.T) {
		t.Parallel()
		log := &recordingLog{}
		err := (maintainerAudit{log: log}).Record(t.Context(), &portfolioapp.AuditRecord{Event: portfolioapp.AuditDismissed, Outcome: "ok", RunID: "trim ETFA", Actor: owner.ID, At: at})
		if !errors.Is(err, identity.ErrInvalidAuditEntry) || len(log.entries) != 0 {
			t.Errorf("Record = %v with %d entries, want ErrInvalidAuditEntry and none", err, len(log.entries))
		}
	})
}

// accountsFake lists accounts.
type accountsFake []ledger.Account

func (a accountsFake) Accounts(context.Context) ([]ledger.Account, error) { return a, nil }

// viewAuthz lets a caller view the accounts listed for its holder, and keeps
// the callers it was asked about.
type viewAuthz struct {
	allowed map[identity.Holder]map[string]bool
	asked   []identity.Caller
	err     error
}

func (a *viewAuthz) MayAct(_ context.Context, c identity.Caller, act identity.Action, account string) (bool, error) {
	a.asked = append(a.asked, c)
	if a.err != nil || act != identity.ActionView {
		return false, a.err
	}
	return a.allowed[c.Holder][account], nil
}

// sessionsByToken knows "owner", "demo" and "viewer".
type sessionsByToken struct{}

func (sessionsByToken) Session(_ context.Context, _ identity.Origin, token string, want identity.SessionScope) (identity.Session, error) {
	var s identity.Session
	switch token {
	case "owner":
		s = identity.Session{ID: "owner-id", Scope: identity.SessionScopeOwner}
	case "demo":
		s = identity.Session{ID: "demo-id", Scope: identity.SessionScopeRead, Authenticator: identity.AuthenticatorDemo}
	default:
		return identity.Session{}, fmt.Errorf("%w: unknown", identityapp.ErrInvalidSession)
	}
	if !s.Scope.Includes(want) {
		return identity.Session{}, fmt.Errorf("%w: scope", identityapp.ErrForbidden)
	}
	return s, nil
}

// readerCase is one session and who the maintainer says it is.
type readerCase struct {
	name, origin, token string
	authz               *viewAuthz
	want                portfolioapp.Reader
	wantErr             error
}

// checkReader fails the test unless the maintainer's sessions say what c wants.
func checkReader(t *testing.T, c *readerCase, s maintainerSessions) {
	t.Helper()
	got, err := s.Reader(t.Context(), c.origin, c.token)
	if !errors.Is(err, c.wantErr) {
		t.Fatalf("Reader error = %v, want %v", err, c.wantErr)
	}
	if err == nil && got != c.want {
		t.Errorf("Reader = %+v, want %+v", got, c.want)
	}
	for _, asked := range c.authz.asked {
		if asked.Holder != identity.HolderDemoSession || asked.ID != "demo-id" {
			t.Errorf("the authorizer was asked about %+v, want the demo session", asked)
		}
	}
	if c.token == "owner" && len(c.authz.asked) != 0 {
		t.Errorf("the owner's session was checked account by account: %v", c.authz.asked)
	}
}

func TestMaintainerSessionsTellWhoMayReadTheWholePortfolio(t *testing.T) {
	t.Parallel()
	origins, err := identity.NewOrigins([]string{"http://localhost:8080"})
	if err != nil {
		t.Fatal(err)
	}
	const local = "http://localhost:8080"
	accounts := accountsFake{{ID: "acc-1"}, {ID: "acc-2"}}
	both := map[string]bool{"acc-1": true, "acc-2": true}
	demoViews := func(allowed map[string]bool) *viewAuthz {
		return &viewAuthz{allowed: map[identity.Holder]map[string]bool{identity.HolderDemoSession: allowed}}
	}
	tests := []readerCase{
		{name: "the owner needs no account check", origin: local, token: "owner", authz: &viewAuthz{}, want: portfolioapp.Reader{SessionID: "owner-id", Owner: true, ViewsAll: true}},
		{name: "a demo session the household grant lets view every account", origin: local, token: "demo", authz: demoViews(both), want: portfolioapp.Reader{SessionID: "demo-id", ViewsAll: true}},
		{name: "a session that may view only one account", origin: local, token: "demo", authz: demoViews(map[string]bool{"acc-1": true}), want: portfolioapp.Reader{SessionID: "demo-id"}},
		{name: "an authorizer that can't tell is a refusal", origin: local, token: "demo", authz: &viewAuthz{err: errSpiceDown}, want: portfolioapp.Reader{SessionID: "demo-id"}},
		{name: "no session", origin: local, token: "nope", authz: &viewAuthz{}, wantErr: portfolioapp.ErrNoSession},
		{name: "an origin off the list", origin: "https://other.example.com", token: "owner", authz: &viewAuthz{}, wantErr: portfolioapp.ErrNoSession},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkReader(t, &tt, maintainerSessions{auth: sessionsByToken{}, origins: origins, ledger: accounts, authz: tt.authz})
		})
	}
}

func TestMaintainerServiceIsMountedForViewPluginsOnly(t *testing.T) {
	t.Parallel()
	m := portfolioapp.NewMaintainer(&portfolioapp.MaintainerPorts{Sessions: readerOnly{}})
	other := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := withMaintainerService(other, m, uiCallers{}, slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	call := func(token string) connect.Code {
		c := portfoliov1connect.NewMaintainerServiceClient(srv.Client(), srv.URL)
		req := connect.NewRequest(&portfoliov1.AskMaintainerRequest{Origin: "o", SessionToken: "s"})
		if token != "" {
			req.Header().Set("Authorization", "Bearer "+token)
		}
		_, err := c.AskMaintainer(t.Context(), req)
		return connect.CodeOf(err)
	}
	// A view plugin gets through to the use case, which refuses a session that is not the owner's.
	if code := call("ui"); code != connect.CodePermissionDenied {
		t.Errorf("a view plugin's call: code = %v, want PermissionDenied from the use case", code)
	}
	for _, token := range []string{"", "stranger"} {
		if code := call(token); code != connect.CodePermissionDenied {
			t.Errorf("a call with the token %q: code = %v, want PermissionDenied", token, code)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/anything", http.NoBody))
	if rec.Code != http.StatusTeapot {
		t.Errorf("another path = %d, want the handler it was mounted over", rec.Code)
	}
}

// readerOnly is a session that is no owner's.
type readerOnly struct{}

func (readerOnly) Reader(context.Context, string, string) (portfolioapp.Reader, error) {
	return portfolioapp.Reader{SessionID: "00000000000000b2", ViewsAll: true}, nil
}
