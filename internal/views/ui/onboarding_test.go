package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

var (
	errRoleNeedsOverride = errors.New("sleeve play is offensive but kind core is defensive; a role other than the kind's needs the owner's explicit choice")
	errBrokerMapped      = errors.New("the broker account is mapped to another account")
)

// fakeSleeves is an in-memory SleeveService. Like the core, it wants the
// plugin's token and an owner-scoped session; it refuses an offensive
// sleeve of kind core without the override, and records what it took.
type fakeSleeves struct {
	portfoliov1connect.UnimplementedSleeveServiceHandler

	mu    sync.Mutex
	calls []string
}

func (f *fakeSleeves) AddSleeveKind(_ context.Context, r *connect.Request[portfoliov1.AddSleeveKindRequest]) (*connect.Response[portfoliov1.AddSleeveKindResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken(), "kind "+r.Msg.GetName()+" "+r.Msg.GetDefaultRole().String()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&portfoliov1.AddSleeveKindResponse{Kind: &portfoliov1.SleeveKind{Key: "gold", Name: r.Msg.GetName()}}), nil
}

func (f *fakeSleeves) RenameSleeveKind(_ context.Context, r *connect.Request[portfoliov1.RenameSleeveKindRequest]) (*connect.Response[portfoliov1.RenameSleeveKindResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken(), "rename "+r.Msg.GetKey()+" "+r.Msg.GetName()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&portfoliov1.RenameSleeveKindResponse{Kind: &portfoliov1.SleeveKind{Key: r.Msg.GetKey(), Name: r.Msg.GetName()}}), nil
}

func (f *fakeSleeves) AddSleeve(_ context.Context, r *connect.Request[portfoliov1.AddSleeveRequest]) (*connect.Response[portfoliov1.AddSleeveResponse], error) {
	m := r.Msg
	if m.GetKind() == "core" && m.GetRole() == portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE && !m.GetRoleOverride() {
		return nil, codeErr(connect.CodeInvalidArgument, errRoleNeedsOverride)
	}
	call := "add " + m.GetName() + " " + m.GetKind() + " " + m.GetRole().String()
	if m.GetRoleOverride() {
		call += " override"
	}
	if err := f.admit(r.Header(), m.GetSessionToken(), call); err != nil {
		return nil, err
	}
	return connect.NewResponse(&portfoliov1.AddSleeveResponse{Sleeve: &portfoliov1.Sleeve{Id: "x", Name: m.GetName()}}), nil
}

func (f *fakeSleeves) UpdateSleeve(_ context.Context, r *connect.Request[portfoliov1.UpdateSleeveRequest]) (*connect.Response[portfoliov1.UpdateSleeveResponse], error) {
	m := r.Msg
	if err := f.admit(r.Header(), m.GetSessionToken(), "update "+m.GetId()+" "+m.GetName()+" "+m.GetKind()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&portfoliov1.UpdateSleeveResponse{Sleeve: &portfoliov1.Sleeve{Id: m.GetId(), Name: m.GetName()}}), nil
}

// admit checks a call's caller and session and records it as call.
func (f *fakeSleeves) admit(h http.Header, session, call string) error {
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(session); err != nil {
		return err
	}
	f.locked(func() { f.calls = append(f.calls, call) })
	return nil
}

// locked runs fn with the fake's lock held.
func (f *fakeSleeves) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// fakeAccounts is an in-memory AccountService, like fakeSleeves; it
// refuses the synthetic broker account U0000002 as mapped to another
// account.
type fakeAccounts struct {
	ingestv1connect.UnimplementedAccountServiceHandler

	mu    sync.Mutex
	calls []string
}

func (f *fakeAccounts) CreateAccount(_ context.Context, r *connect.Request[ingestv1.CreateAccountRequest]) (*connect.Response[ingestv1.CreateAccountResponse], error) {
	m := r.Msg
	if m.GetBrokerAccount() == "U0000002" {
		return nil, codeErr(connect.CodeFailedPrecondition, errBrokerMapped)
	}
	call := "create " + m.GetName() + "|" + m.GetInstitution() + "|" + m.GetSleeve() + "|" + m.GetConnectorInstance() + "|" + m.GetBrokerAccount()
	if err := f.admit(r.Header(), m.GetSessionToken(), call); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.CreateAccountResponse{Account: &ingestv1.Account{Id: "x", Name: m.GetName()}}), nil
}

func (f *fakeAccounts) UpdateAccount(_ context.Context, r *connect.Request[ingestv1.UpdateAccountRequest]) (*connect.Response[ingestv1.UpdateAccountResponse], error) {
	m := r.Msg
	if err := f.admit(r.Header(), m.GetSessionToken(), "update "+m.GetAccountId()+"|"+m.GetName()+"|"+m.GetSleeve()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.UpdateAccountResponse{Account: &ingestv1.Account{Id: m.GetAccountId()}}), nil
}

func (f *fakeAccounts) BindAccount(_ context.Context, r *connect.Request[ingestv1.BindAccountRequest]) (*connect.Response[ingestv1.BindAccountResponse], error) {
	m := r.Msg
	if err := f.admit(r.Header(), m.GetSessionToken(), "bind "+m.GetAccountId()+"|"+m.GetConnectorInstance()+"|"+m.GetBrokerAccount()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.BindAccountResponse{}), nil
}

func (f *fakeAccounts) admit(h http.Header, session, call string) error {
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(session); err != nil {
		return err
	}
	f.locked(func() { f.calls = append(f.calls, call) })
	return nil
}

// locked runs fn with the fake's lock held.
func (f *fakeAccounts) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// ListSleeves lists the three default sleeves, satellite made defensive
// by the owner's explicit choice.
func (f *fakeRead) ListSleeves(_ context.Context, r *connect.Request[viewsv1.ListSleevesRequest]) (*connect.Response[viewsv1.ListSleevesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListSleevesResponse{Sleeves: []*viewsv1.Sleeve{
		{Id: "bonds", Name: "Bonds", Kind: "bonds", Role: viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE},
		{Id: "core", Name: "Core", Kind: "core", Role: viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE},
		{Id: "satellite", Name: "Satellite", Kind: "satellite", Role: viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE, RoleOverride: true},
	}}), nil
}

// ListSleeveKinds lists three built-in kinds.
func (f *fakeRead) ListSleeveKinds(_ context.Context, r *connect.Request[viewsv1.ListSleeveKindsRequest]) (*connect.Response[viewsv1.ListSleeveKindsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListSleeveKindsResponse{Kinds: []*viewsv1.SleeveKind{
		{Key: "bonds", Name: "Bonds", DefaultRole: viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE},
		{Key: "core", Name: "Core", DefaultRole: viewsv1.SleeveRole_SLEEVE_ROLE_DEFENSIVE},
		{Key: "satellite", Name: "Satellite", DefaultRole: viewsv1.SleeveRole_SLEEVE_ROLE_OFFENSIVE},
	}}), nil
}

// ListConnectorBindings binds ibkr-main to the synthetic account.
func (f *fakeRead) ListConnectorBindings(_ context.Context, r *connect.Request[viewsv1.ListConnectorBindingsRequest]) (*connect.Response[viewsv1.ListConnectorBindingsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListConnectorBindingsResponse{Bindings: []*viewsv1.ConnectorBinding{
		{ConnectorInstance: "ibkr-main", AccountId: accountID},
	}}), nil
}

// ListBrokerAccounts lists the synthetic broker accounts ibkr-main
// reported: U0000001 unmapped, U0000002 mapped to the synthetic account.
func (f *fakeRead) ListBrokerAccounts(_ context.Context, r *connect.Request[viewsv1.ListBrokerAccountsRequest]) (*connect.Response[viewsv1.ListBrokerAccountsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListBrokerAccountsResponse{BrokerAccounts: []*viewsv1.BrokerAccount{
		{ConnectorInstance: "ibkr-main", Id: "U0000001", Institution: "ibkr", LastSeen: "2026-03-31", InLastDelivery: true},
		{ConnectorInstance: "ibkr-main", Id: "U0000002", Institution: "ibkr", AccountId: accountID, LastSeen: "2026-03-31"},
	}}), nil
}

func TestSleevesPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Satellite", "differs from its kind", "Add a sleeve", "Add kind", `action="/sleeves/core"`,
		`action="/sleeve-kinds/satellite"`, "never changes", `name="next" value="/sleeves"`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Satellite", "Sleeve kinds")
	wantNoBody(t, rec, "Add a sleeve", "Add kind", "<form method=\"post\" action=\"/sleeves")
}

func TestSleeveForms(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	posts := []struct {
		target, body, wantCall, wantLocation string
	}{
		{"/sleeves", "name=+Gold+&kind=core&role=&next=/setup?step%3Dsleeves", "add Gold core SLEEVE_ROLE_UNSPECIFIED", setupSleeves},
		{"/sleeves", "name=Cautious&kind=satellite&role=defensive&role_override=on&next=https://elsewhere.example", "add Cautious satellite SLEEVE_ROLE_DEFENSIVE override", sleevesPath},
		{"/sleeves/core", "name=Long+term&kind=core&role=&next=/sleeves", "update core Long term core", sleevesPath},
		{"/sleeve-kinds", "name=Gold&default_role=defensive&next=/sleeves", "kind Gold SLEEVE_ROLE_DEFENSIVE", sleevesPath},
		{"/sleeve-kinds/satellite", "name=Play&next=/sleeves", "rename satellite Play", sleevesPath},
	}
	for _, p := range posts {
		rec := h.serve(t, withSession(form(p.target, p.body), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != p.wantLocation {
			t.Errorf("POST %s went to %q, want %q", p.target, got, p.wantLocation)
		}
	}
	calls := seen(h.sleeves, func() []string { return h.sleeves.calls })
	for i, p := range posts {
		if i >= len(calls) || calls[i] != p.wantCall {
			t.Errorf("call %d = %q, want %q (all: %q)", i, calls, p.wantCall, calls)
		}
	}
	rec := h.serve(t, withSession(form("/sleeves", "name=Play&kind=core&role=offensive&next=/sleeves"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "explicit choice", "Add a sleeve")
	rec = h.serve(t, withSession(form("/sleeves", "name=Gold&kind=core"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	if n := len(seen(h.sleeves, func() []string { return h.sleeves.calls })); n != len(posts) {
		t.Errorf("%d calls reached the core, want %d: a refusal and a read session store nothing", n, len(posts))
	}
}

func TestAccountsPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Unmapped broker accounts", "U0000001", "not booked", "Synthetic brokerage", "ibkr-main",
		`value="ibkr-main/U0000001"`, "Create its account", "Map to an account", "Add an account", "Booked in")
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "U0000001", "Unmapped")
	wantNoBody(t, rec, "Add an account", "Create its account", `<form method="post" action="/accounts`)
}

func TestAccountForms(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	posts := []struct {
		target, body, wantCall string
	}{
		{
			"/accounts", "name=+IBKR+main+&sleeve=satellite&instance=&broker=ibkr-main/U0000001&institution=&next=/accounts",
			"create IBKR main||satellite|ibkr-main|U0000001",
		},
		{
			"/accounts", "name=Bonds&sleeve=bonds&instance=bonds-main&institution=treasury-bonds&next=/accounts",
			"create Bonds|treasury-bonds|bonds|bonds-main|",
		},
		{"/accounts/" + accountID + "/edit", "name=Renamed&sleeve=core&next=/accounts", "update " + accountID + "|Renamed|core"},
		{"/accounts/bind", "account=" + accountID + "&broker=ibkr-main/U0000001&next=/accounts", "bind " + accountID + "|ibkr-main|U0000001"},
		{"/accounts/bind", "account=" + accountID + "&instance=xtb-main&next=/accounts", "bind " + accountID + "|xtb-main|"},
	}
	for _, p := range posts {
		rec := h.serve(t, withSession(form(p.target, p.body), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != accountsPath {
			t.Errorf("POST %s went to %q, want %q", p.target, got, accountsPath)
		}
	}
	calls := seen(h.accounts, func() []string { return h.accounts.calls })
	for i, p := range posts {
		if i >= len(calls) || calls[i] != p.wantCall {
			t.Errorf("call %d: got %q, want %q", i, calls, p.wantCall)
		}
	}
	rec := h.serve(t, withSession(form("/accounts", "name=Other&sleeve=core&broker=ibkr-main/U0000002&next=/setup?step%3Daccounts"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "mapped to another account", "Accounts and connectors")
	rec = h.serve(t, withSession(form("/accounts", "name=X&sleeve=core"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
}

// TestFirstRunWizard walks the wizard (D-73): an owner whose install has
// no account lands on it from the overview, the sleeves step shows the
// defaults with their forms, the accounts step the broker accounts to
// pick from, and the last step where to go next; a read session neither
// lands on it nor may open it.
func TestFirstRunWizard(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.noAccounts = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != setupSleeves {
		t.Fatalf("an owner with no account went to %q, want the wizard", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "No accounts yet.")

	steps := []struct {
		step string
		want []string
	}{
		{stepSleeves, []string{`aria-current="step"`, "three sleeves", "Bonds", "Add a sleeve", `name="next" value="/setup?step=sleeves"`, "Continue to accounts"}},
		{stepAccounts, []string{"U0000001", "Add an account", `name="next" value="/setup?step=accounts"`, "Finish setup"}},
		{stepDone, []string{"OpenFief is set up.", "kasztan-seal", "Go to the overview"}},
		{"bogus", []string{"three sleeves"}},
	}
	for _, s := range steps {
		rec = h.serve(t, withSession(request(http.MethodGet, "/setup?step="+url.QueryEscape(s.step), ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, s.want...)
	}
	rec = h.serve(t, withSession(form("/accounts", "name=IBKR&sleeve=satellite&broker=ibkr-main/U0000001&next=/setup?step%3Daccounts"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != setupAccounts {
		t.Errorf("a wizard form went to %q, want back to its step", got)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/setup", ""), readSession))
	wantStatus(t, rec, http.StatusForbidden)
}

func TestReturnTo(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		sleevesPath: sleevesPath, accountsPath: accountsPath, setupSleeves: setupSleeves, setupAccounts: setupAccounts,
		"": sleevesPath, "https://elsewhere.example/": sleevesPath, "//elsewhere.example": sleevesPath, "/settings": sleevesPath,
	} {
		if got := returnTo(in, sleevesPath); got != want {
			t.Errorf("returnTo(%q) = %q, want %q", in, got, want)
		}
	}
}
