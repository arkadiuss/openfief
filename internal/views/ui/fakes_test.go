package ui

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/kernel"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// Synthetic credentials and record for the fakes; nothing here is real.
const (
	pluginToken  = "plugin-token-1"
	ownerSession = "session-owner"
	readSession  = "session-read"
	demoSession  = "session-demo"
	newSession   = "session-new"
	accountID    = "acc-synthetic-1"
	testHost     = "ui.example.com"
	testOrigin   = "https://" + testHost
	localOrigin  = "http://localhost:8080"
)

// codeErr is a Connect error with code, as the fake core returns it.
func codeErr(code connect.Code, err error) error {
	return fmt.Errorf("fake core: %w", connect.NewError(code, err))
}

// errLastPasskey is the fake core's reason for refusing to remove the last
// passkey.
var errLastPasskey = errors.New("the last passkey can't be removed")

// fakeAuth is an in-memory AuthService. It accepts only the plugin's token
// in Authorization and the sessions it knows, and records what it was asked.
type fakeAuth struct {
	identityv1connect.UnimplementedAuthServiceHandler

	mu          sync.Mutex
	demo        bool
	down        bool
	origins     []string
	bearers     []string
	enrollCode  string
	regSession  string
	credentials []string
	removed     []string
	revoked     []string
	loggedOut   []string
	tokenTTL    time.Duration
	linkTarget  string
	lastPasskey bool
}

func (f *fakeAuth) GetLoginMethods(_ context.Context, r *connect.Request[identityv1.GetLoginMethodsRequest]) (*connect.Response[identityv1.GetLoginMethodsResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&identityv1.GetLoginMethodsResponse{PasskeyRegistered: true, DemoSessions: f.flag(&f.demo)}), nil
}

func (f *fakeAuth) BeginLogin(_ context.Context, r *connect.Request[identityv1.BeginLoginRequest]) (*connect.Response[identityv1.BeginLoginResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&identityv1.BeginLoginResponse{CeremonyId: "c-login", OptionsJson: `{"challenge":"AAAA"}`}), nil
}

func (f *fakeAuth) FinishLogin(_ context.Context, r *connect.Request[identityv1.FinishLoginRequest]) (*connect.Response[identityv1.FinishLoginResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if r.Msg.GetCeremonyId() != "c-login" {
		return nil, codeErr(connect.CodeInvalidArgument, nil)
	}
	f.mu.Lock()
	f.credentials = append(f.credentials, r.Msg.GetCredentialJson())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.FinishLoginResponse{SessionToken: newSession, Session: issued(identityv1.SessionScope_SESSION_SCOPE_OWNER)}), nil
}

func (f *fakeAuth) CreateDemoSession(_ context.Context, r *connect.Request[identityv1.CreateDemoSessionRequest]) (*connect.Response[identityv1.CreateDemoSessionResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if !f.flag(&f.demo) {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	return connect.NewResponse(&identityv1.CreateDemoSessionResponse{SessionToken: demoSession, Session: issued(identityv1.SessionScope_SESSION_SCOPE_READ)}), nil
}

func (f *fakeAuth) BeginRegistration(_ context.Context, r *connect.Request[identityv1.BeginRegistrationRequest]) (*connect.Response[identityv1.BeginRegistrationResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.enrollCode, f.regSession = r.Msg.GetEnrollmentCode(), r.Msg.GetSessionToken()
	f.mu.Unlock()
	if r.Msg.GetEnrollmentCode() == "" {
		if err := owner(r.Msg.GetSessionToken()); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&identityv1.BeginRegistrationResponse{CeremonyId: "c-reg", OptionsJson: `{"challenge":"BBBB"}`}), nil
}

func (f *fakeAuth) FinishRegistration(_ context.Context, r *connect.Request[identityv1.FinishRegistrationRequest]) (*connect.Response[identityv1.FinishRegistrationResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	resp := &identityv1.FinishRegistrationResponse{Passkey: &identityv1.Passkey{Id: "pk2", Name: r.Msg.GetName()}}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.credentials = append(f.credentials, r.Msg.GetCredentialJson())
	if f.enrollCode != "" {
		resp.SessionToken, resp.Session = newSession, issued(identityv1.SessionScope_SESSION_SCOPE_OWNER)
	}
	return connect.NewResponse(resp), nil
}

func (f *fakeAuth) Logout(_ context.Context, r *connect.Request[identityv1.LogoutRequest]) (*connect.Response[identityv1.LogoutResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.loggedOut = append(f.loggedOut, r.Msg.GetSessionToken())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.LogoutResponse{}), nil
}

func (f *fakeAuth) WhoAmI(_ context.Context, r *connect.Request[identityv1.WhoAmIRequest]) (*connect.Response[identityv1.WhoAmIResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	scope, err := scopeOf(r.Msg.GetSessionToken())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&identityv1.WhoAmIResponse{Owner: &identityv1.Owner{}, Session: issued(scope)}), nil
}

func (f *fakeAuth) ListPasskeys(_ context.Context, r *connect.Request[identityv1.ListPasskeysRequest]) (*connect.Response[identityv1.ListPasskeysResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	created := timestamppb.New(time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC))
	return connect.NewResponse(&identityv1.ListPasskeysResponse{Passkeys: []*identityv1.Passkey{
		{Id: "pk1", Name: "laptop", RpId: testHost, CreatedAt: created},
	}}), nil
}

func (f *fakeAuth) RemovePasskey(_ context.Context, r *connect.Request[identityv1.RemovePasskeyRequest]) (*connect.Response[identityv1.RemovePasskeyResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if f.flag(&f.lastPasskey) {
		return nil, codeErr(connect.CodeFailedPrecondition, errLastPasskey)
	}
	f.mu.Lock()
	f.removed = append(f.removed, r.Msg.GetPasskeyId())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.RemovePasskeyResponse{}), nil
}

func (f *fakeAuth) CreateScriptToken(_ context.Context, r *connect.Request[identityv1.CreateScriptTokenRequest]) (*connect.Response[identityv1.CreateScriptTokenResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.tokenTTL = r.Msg.GetTtl().AsDuration()
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.CreateScriptTokenResponse{
		Token:  &identityv1.ScriptToken{Id: "0123456789abcdef", Name: r.Msg.GetName()},
		Secret: "ofs_0123456789abcdef_synthetic-secret",
	}), nil
}

func (f *fakeAuth) ListScriptTokens(_ context.Context, r *connect.Request[identityv1.ListScriptTokensRequest]) (*connect.Response[identityv1.ListScriptTokensResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	ts := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	return connect.NewResponse(&identityv1.ListScriptTokensResponse{Tokens: []*identityv1.ScriptToken{
		{Id: "fedcba9876543210", Name: "backup-script", Status: identityv1.ScriptTokenStatus_SCRIPT_TOKEN_STATUS_ACTIVE, CreatedAt: ts, ExpiresAt: ts},
	}}), nil
}

func (f *fakeAuth) RevokeScriptToken(_ context.Context, r *connect.Request[identityv1.RevokeScriptTokenRequest]) (*connect.Response[identityv1.RevokeScriptTokenResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.revoked = append(f.revoked, r.Msg.GetTokenId())
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.RevokeScriptTokenResponse{}), nil
}

func (f *fakeAuth) CreateEnrollmentLink(_ context.Context, r *connect.Request[identityv1.CreateEnrollmentLinkRequest]) (*connect.Response[identityv1.CreateEnrollmentLinkResponse], error) {
	if err := f.record(r.Header(), r.Msg.GetOrigin()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.linkTarget = r.Msg.GetTargetOrigin()
	f.mu.Unlock()
	return connect.NewResponse(&identityv1.CreateEnrollmentLinkResponse{
		Url:       r.Msg.GetTargetOrigin() + "/enroll#code=synthetic-code",
		ExpiresAt: timestamppb.New(time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)),
	}), nil
}

// locked runs fn with the fake's lock held: tests set its switches and read
// what it recorded through it, since the fake runs on the server's
// goroutines.
func (f *fakeAuth) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// flag reads one of the fake's switches under its lock.
func (f *fakeAuth) flag(p *bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return *p
}

// record notes a call's origin and bearer token and checks the bearer.
func (f *fakeAuth) record(h http.Header, origin string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.origins = append(f.origins, origin)
	f.bearers = append(f.bearers, h.Get("Authorization"))
	if f.down {
		return codeErr(connect.CodeUnavailable, nil)
	}
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	return nil
}

// scopeOf returns the scope of a known session token.
func scopeOf(token string) (identityv1.SessionScope, error) {
	switch token {
	case ownerSession, newSession:
		return identityv1.SessionScope_SESSION_SCOPE_OWNER, nil
	case readSession, demoSession:
		return identityv1.SessionScope_SESSION_SCOPE_READ, nil
	}
	return identityv1.SessionScope_SESSION_SCOPE_UNSPECIFIED, codeErr(connect.CodeUnauthenticated, nil)
}

// owner checks that token is a known owner-scoped session.
func owner(token string) error {
	scope, err := scopeOf(token)
	if err != nil {
		return err
	}
	if scope != identityv1.SessionScope_SESSION_SCOPE_OWNER {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	return nil
}

// issued is a session the fake hands out, expiring on a fixed date.
func issued(scope identityv1.SessionScope) *identityv1.Session {
	return &identityv1.Session{Id: "s1", Scope: scope, ExpiresAt: timestamppb.New(time.Date(2030, 1, 2, 0, 0, 0, 0, time.UTC))}
}

// fakeRead is an in-memory read API over a small synthetic record. It
// accepts only known session tokens as bearer tokens and records what it
// was asked.
type fakeRead struct {
	viewsv1connect.UnimplementedReadServiceHandler

	mu         sync.Mutex
	bearers    []string
	dimensions []viewsv1.AllocationDimension
	pageTokens []string
	dates      []string
	// modelOn turns the model portfolio on; off, GetModelPortfolio is
	// not found, as the core answers (D-52).
	modelOn bool
	// institution is the account's institution; empty is a synthetic
	// broker.
	institution string
	// pluginsUnavailable makes ListPlugins answer UNAVAILABLE, as a core
	// that can't read its cluster does.
	pluginsUnavailable bool
}

func pln(amount string) *viewsv1.Money { return &viewsv1.Money{Amount: amount, Currency: "PLN"} }

func syntheticInstrument() *viewsv1.Instrument {
	return &viewsv1.Instrument{Isin: "XX0000000001", Symbol: "SYN"}
}

func (f *fakeRead) ListAccounts(_ context.Context, r *connect.Request[viewsv1.ListAccountsRequest]) (*connect.Response[viewsv1.ListAccountsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListAccountsResponse{Accounts: []*viewsv1.Account{
		{Id: accountID, Institution: "synthetic-broker", Name: "Synthetic brokerage", Sleeve: "core"},
	}}), nil
}

func (f *fakeRead) GetAccount(_ context.Context, r *connect.Request[viewsv1.GetAccountRequest]) (*connect.Response[viewsv1.GetAccountResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	if r.Msg.GetAccountId() != accountID {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	var institution string
	f.locked(func() { institution = f.institution })
	return connect.NewResponse(&viewsv1.GetAccountResponse{Account: &viewsv1.Account{
		Id: accountID, Institution: cmp.Or(institution, "synthetic-broker"), Name: "Synthetic brokerage", Sleeve: "core",
	}}), nil
}

func (f *fakeRead) ListValuationStatuses(_ context.Context, r *connect.Request[viewsv1.ListValuationStatusesRequest]) (*connect.Response[viewsv1.ListValuationStatusesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListValuationStatusesResponse{Statuses: []*viewsv1.ValuationStatus{f.status()}}), nil
}

func (f *fakeRead) GetValuationStatus(_ context.Context, r *connect.Request[viewsv1.GetValuationStatusRequest]) (*connect.Response[viewsv1.GetValuationStatusResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.GetValuationStatusResponse{Status: f.status()}), nil
}

func (f *fakeRead) GetAllocation(_ context.Context, r *connect.Request[viewsv1.GetAllocationRequest]) (*connect.Response[viewsv1.GetAllocationResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.dimensions = append(f.dimensions, r.Msg.GetBy())
	f.dates = append(f.dates, r.Msg.GetDate())
	f.mu.Unlock()
	return connect.NewResponse(&viewsv1.GetAllocationResponse{
		Date: r.Msg.GetDate(), By: r.Msg.GetBy(), TotalReporting: pln("1500.00"),
		Groups: []*viewsv1.AllocationGroup{
			{Key: "listed", ValueReporting: pln("1000.00")},
			{Key: "cash", ValueReporting: pln("500.00")},
		},
	}), nil
}

func (f *fakeRead) ListValuationTotals(_ context.Context, r *connect.Request[viewsv1.ListValuationTotalsRequest]) (*connect.Response[viewsv1.ListValuationTotalsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListValuationTotalsResponse{
		From: r.Msg.GetFrom(), To: r.Msg.GetTo(), ReportingCurrency: "PLN",
		Totals: []*viewsv1.DateTotal{
			{Date: "2026-03-30", ValueReporting: pln("990.00")},
			{Date: "2026-03-31", ValueReporting: pln("1000.00")},
		},
	}), nil
}

func (f *fakeRead) ListAccountValuations(_ context.Context, r *connect.Request[viewsv1.ListAccountValuationsRequest]) (*connect.Response[viewsv1.ListAccountValuationsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListAccountValuationsResponse{
		AccountId: accountID, From: r.Msg.GetFrom(), To: r.Msg.GetTo(), ReportingCurrency: "PLN",
		Valuations: []*viewsv1.Valuation{{
			Date: r.Msg.GetTo(), Instrument: syntheticInstrument(), Basis: viewsv1.ValuationBasis_VALUATION_BASIS_MARKET,
			Value: &viewsv1.Money{Amount: "250.00", Currency: "EUR"}, ValueReporting: pln("1000.00"),
		}},
		Totals: []*viewsv1.DateTotal{{Date: r.Msg.GetTo(), ValueReporting: pln("1000.00")}},
	}), nil
}

func (f *fakeRead) ListAssertionChecks(_ context.Context, r *connect.Request[viewsv1.ListAssertionChecksRequest]) (*connect.Response[viewsv1.ListAssertionChecksResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListAssertionChecksResponse{AccountId: accountID, Checks: []*viewsv1.AssertionCheck{{
		Date: "2026-03-15", Book: &viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_CASH, Currency: "EUR"},
		Reported: "100.00", Derived: "90.00", Status: viewsv1.AssertionStatus_ASSERTION_STATUS_FLAGGED,
	}}}), nil
}

func (f *fakeRead) ListHoldings(_ context.Context, r *connect.Request[viewsv1.ListHoldingsRequest]) (*connect.Response[viewsv1.ListHoldingsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListHoldingsResponse{AccountId: accountID, Date: r.Msg.GetDate(), Holdings: []*viewsv1.Holding{
		{Instrument: syntheticInstrument(), Quantity: "10"},
	}}), nil
}

func (f *fakeRead) ListLots(_ context.Context, r *connect.Request[viewsv1.ListLotsRequest]) (*connect.Response[viewsv1.ListLotsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListLotsResponse{AccountId: accountID, Date: r.Msg.GetDate(), Lots: []*viewsv1.Lot{
		{Instrument: syntheticInstrument(), Opened: "2026-01-02", Quantity: "10", Cost: &viewsv1.Money{Amount: "2000.00", Currency: "EUR"}, Opening: true},
	}}), nil
}

func (f *fakeRead) ListCash(_ context.Context, r *connect.Request[viewsv1.ListCashRequest]) (*connect.Response[viewsv1.ListCashResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListCashResponse{AccountId: accountID, Date: r.Msg.GetDate(), Balances: []*viewsv1.Money{
		{Amount: "12345.67", Currency: "EUR"},
	}}), nil
}

func (f *fakeRead) ListTransactions(_ context.Context, r *connect.Request[viewsv1.ListTransactionsRequest]) (*connect.Response[viewsv1.ListTransactionsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.pageTokens = append(f.pageTokens, r.Msg.GetPageToken())
	f.mu.Unlock()
	resp := &viewsv1.ListTransactionsResponse{Transactions: []*viewsv1.Transaction{{
		Source: "synthetic", ExternalId: "t1", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-01-02",
		Description: "Synthetic deposit",
		Postings: []*viewsv1.Posting{{
			Book: &viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_CASH, Currency: "EUR"}, Amount: &viewsv1.Money{Amount: "100.00", Currency: "EUR"},
		}},
	}, {
		Source: "plugin:ibkr-old:ibkr", ExternalId: "t2", Kind: viewsv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-01-03",
		Description: "Synthetic deposit from a removed instance", ConnectorInstance: "ibkr-old", ConnectorInstanceRemoved: true,
		Postings: []*viewsv1.Posting{{
			Book: &viewsv1.Book{Kind: viewsv1.BookKind_BOOK_KIND_CASH, Currency: "EUR"}, Amount: &viewsv1.Money{Amount: "50.00", Currency: "EUR"},
		}},
	}}}
	if r.Msg.GetPageToken() == "" {
		resp.NextPageToken = "page-2"
	}
	return connect.NewResponse(resp), nil
}

// errPluginsUnavailable is the core's answer while it can't read its
// cluster.
var errPluginsUnavailable = errors.New("the installed plugins can't be read now")

// ListPlugins lists two synthetic plugins: ibkr-main, signed by OpenFief's
// CI and running, and ui-local, signed by the owner's own signer; or it
// answers UNAVAILABLE when pluginsUnavailable is set.
func (f *fakeRead) ListPlugins(_ context.Context, r *connect.Request[viewsv1.ListPluginsRequest]) (*connect.Response[viewsv1.ListPluginsResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	unavailable := f.pluginsUnavailable
	f.mu.Unlock()
	if unavailable {
		return nil, codeErr(connect.CodeUnavailable, errPluginsUnavailable)
	}
	return connect.NewResponse(&viewsv1.ListPluginsResponse{Plugins: []*viewsv1.Plugin{
		{
			Instance: "ibkr-main", Image: "registry.example.com/openfief/ibkr-flex@sha256:a1a1", ImageDigest: "sha256:a1a1",
			Capabilities: []string{"connector"}, RequestedCapabilities: []string{"connector"},
			Trust: viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI, Signer: "openfief-ci", TrustReason: "SignedByTrustedSigner",
			Phase: "Running", Ready: true,
		},
		{
			Instance: "ui-local", Image: "registry.example.com/owner/ui@sha256:b2b2", ImageDigest: "sha256:b2b2",
			RequestedCapabilities: []string{"view"},
			Trust:                 viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED, TrustReason: "UntrustedSigner",
			Phase: "SignatureNotVerified",
		},
	}}), nil
}

// locked runs fn with the fake's lock held, for reading what it recorded.
func (f *fakeRead) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// check notes a call's bearer token and refuses unknown sessions.
func (f *fakeRead) check(h http.Header) error {
	f.mu.Lock()
	f.bearers = append(f.bearers, h.Get("Authorization"))
	f.mu.Unlock()
	for _, s := range []string{ownerSession, readSession, demoSession, newSession} {
		if h.Get("Authorization") == "Bearer "+s {
			return nil
		}
	}
	return codeErr(connect.CodeUnauthenticated, nil)
}

func (f *fakeRead) status() *viewsv1.ValuationStatus {
	return &viewsv1.ValuationStatus{AccountId: accountID, ValuedThrough: "2026-03-31", Pending: true}
}

// harness is the UI under test with its fakes behind real Connect
// servers, and the plugin token file the AuthService,
// ModelPortfolioService and BondLotService clients read.
type harness struct {
	auth      *fakeAuth
	model     *fakeModel
	bondLots  *fakeBondLots
	read      *fakeRead
	handler   http.Handler
	tokenFile string
}

// newHarness starts both fakes and the UI. The UI's AuthService client is
// the SDK's CoreHTTPClient over a token file, as the plugin wires it.
func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{auth: &fakeAuth{}, model: newFakeModel(), bondLots: &fakeBondLots{}, read: &fakeRead{}}
	mux := http.NewServeMux()
	mux.Handle(identityv1connect.NewAuthServiceHandler(h.auth))
	mux.Handle(portfoliov1connect.NewModelPortfolioServiceHandler(h.model))
	mux.Handle(ingestv1connect.NewBondLotServiceHandler(h.bondLots))
	mux.Handle(viewsv1connect.NewReadServiceHandler(h.read))
	core := httptest.NewServer(mux)
	t.Cleanup(core.Close)
	h.tokenFile = filepath.Join(t.TempDir(), "token")
	h.writeToken(t, pluginToken)
	handler, err := New(&Config{
		Auth:     identityv1connect.NewAuthServiceClient(openfief.CoreHTTPClient(openfief.TokenFile{Path: h.tokenFile}), core.URL),
		Model:    portfoliov1connect.NewModelPortfolioServiceClient(openfief.CoreHTTPClient(openfief.TokenFile{Path: h.tokenFile}), core.URL),
		BondLots: ingestv1connect.NewBondLotServiceClient(openfief.CoreHTTPClient(openfief.TokenFile{Path: h.tokenFile}), core.URL),
		Read:     viewsv1connect.NewReadServiceClient(core.Client(), core.URL),
		Today:    func() kernel.Date { return mustDate(t, "2026-04-01") },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h.handler = handler
	return h
}

// writeToken replaces the plugin's token file, as the kubelet does when it
// rotates the projected token.
func (h *harness) writeToken(t *testing.T, token string) {
	t.Helper()
	if err := os.WriteFile(h.tokenFile, []byte(token+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// mustDate parses a YYYY-MM-DD date.
func mustDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
