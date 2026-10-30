package connectapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// client is a Connect client of the read API served over HTTP by httptest.
type client struct {
	viewsv1connect.ReadServiceClient
}

func newClient(t *testing.T, l *fakeLedger) client {
	t.Helper()
	srv := httptest.NewServer(newAPI(t, l).handler)
	t.Cleanup(srv.Close)
	return client{viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)}
}

// authed wraps msg in a request that sends token as bearer token ("" for
// none).
func authed[T any](token string, msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if token != "" {
		req.Header().Set("Authorization", "Bearer "+token)
	}
	return req
}

func TestConnectQueries(t *testing.T) {
	t.Parallel()
	c := newClient(t, newLedger(t))
	ctx := context.Background()
	accounts, err := c.ListAccounts(ctx, authed(goodToken, &viewsv1.ListAccountsRequest{}))
	if err != nil || len(accounts.Msg.GetAccounts()) != 1 || accounts.Msg.GetAccounts()[0].GetInstitution() != "broker-a" {
		t.Errorf("ListAccounts() = %v, %v, want acc-1 at broker-a", accounts, err)
	}
	holdings, err := c.ListHoldings(ctx, authed(goodToken, &viewsv1.ListHoldingsRequest{AccountId: "acc-1", Date: "2026-03-02"}))
	if err != nil || len(holdings.Msg.GetHoldings()) != 2 || holdings.Msg.GetHoldings()[1].GetQuantity() != "3" {
		t.Errorf("ListHoldings() = %v, %v, want the bond and 3 ETF1", holdings, err)
	}
}

func TestConnectValuation(t *testing.T) {
	t.Parallel()
	c := newClient(t, newLedger(t))
	ctx := context.Background()
	status, err := c.GetValuationStatus(ctx, authed(goodToken, &viewsv1.GetValuationStatusRequest{AccountId: "acc-1"}))
	if err != nil || status.Msg.GetStatus().GetValuedThrough() != "2026-03-02" || !status.Msg.GetStatus().GetRunning() {
		t.Errorf("GetValuationStatus() = %v, %v, want valued through 2026-03-02, running", status, err)
	}
	alloc, err := c.GetAllocation(ctx, authed(goodToken, &viewsv1.GetAllocationRequest{Date: "2026-03-02", By: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY}))
	// 1043.21 PLN bond + 126.00 PLN for the USD ETF + 400.00 PLN for 100 USD cash.
	if err != nil || alloc.Msg.GetTotalReporting().GetAmount() != "1569.21" || alloc.Msg.GetTotalReporting().GetCurrency() != "PLN" {
		t.Errorf("GetAllocation() = %v, %v, want 1569.21 PLN", alloc, err)
	}
}

func TestConnectPerformance(t *testing.T) {
	t.Parallel()
	c := newClient(t, newLedger(t))
	got, err := c.GetPerformance(context.Background(), authed(goodToken, &viewsv1.GetPerformanceRequest{}))
	if err != nil {
		t.Fatalf("GetPerformance() error = %v", err)
	}
	days := got.Msg.GetDays()
	if len(days) != 1 || days[0].GetDate() != "2026-03-02" || days[0].GetIndex() != "100" || got.Msg.GetGap() != viewsv1.PerformanceGap_PERFORMANCE_GAP_UNSPECIFIED {
		t.Errorf("GetPerformance() = %v, want one day, 2026-03-02 at 100, no gap", got.Msg)
	}
}

func TestConnectPages(t *testing.T) {
	t.Parallel()
	c := newClient(t, newLedger(t))
	ctx := context.Background()
	page, err := c.ListTransactions(ctx, authed(goodToken, &viewsv1.ListTransactionsRequest{AccountId: "acc-1", PageSize: 1}))
	if err != nil || len(page.Msg.GetTransactions()) != 1 || page.Msg.GetNextPageToken() == "" {
		t.Fatalf("ListTransactions(page 1) = %v, %v, want one and a token", page, err)
	}
	next, err := c.ListTransactions(ctx, authed(goodToken, &viewsv1.ListTransactionsRequest{AccountId: "acc-1", PageSize: 1, PageToken: page.Msg.GetNextPageToken()}))
	if err != nil || len(next.Msg.GetTransactions()) != 1 || next.Msg.GetNextPageToken() != "" ||
		next.Msg.GetTransactions()[0].GetExternalId() == page.Msg.GetTransactions()[0].GetExternalId() {
		t.Errorf("ListTransactions(page 2) = %v, %v, want the other transaction and no token", next, err)
	}
}

func TestConnectErrors(t *testing.T) {
	t.Parallel()
	c := newClient(t, newLedger(t))
	ctx := context.Background()
	_, noToken := c.ListSleeves(ctx, authed("", &viewsv1.ListSleevesRequest{}))
	_, unknown := c.GetAccount(ctx, authed(goodToken, &viewsv1.GetAccountRequest{AccountId: "nope"}))
	_, badDate := c.ListCash(ctx, authed(goodToken, &viewsv1.ListCashRequest{AccountId: "acc-1", Date: "03/02/2026"}))
	_, noDimension := c.GetAllocation(ctx, authed(goodToken, &viewsv1.GetAllocationRequest{Date: "2026-03-02"}))
	_, badStatus := c.ListAssertionChecks(ctx, authed(goodToken, &viewsv1.ListAssertionChecksRequest{AccountId: "acc-1", Status: viewsv1.AssertionStatus(9)}))
	tests := []struct {
		name string
		err  error
		want connect.Code
	}{
		{"no token", noToken, connect.CodeUnauthenticated},
		{"unknown account", unknown, connect.CodeNotFound},
		{"bad date", badDate, connect.CodeInvalidArgument},
		{"unspecified dimension", noDimension, connect.CodeInvalidArgument},
		{"unknown assertion status", badStatus, connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		if got := connect.CodeOf(tt.err); got != tt.want {
			t.Errorf("%s: error = %v (%s), want %s", tt.name, tt.err, got, tt.want)
		}
	}
}

func TestConnectGET(t *testing.T) {
	t.Parallel()
	// Read-only methods are also reachable with Connect's GET form.
	_, rec := newAPI(t, newLedger(t)).serve(t, http.MethodGet,
		"/openfief.views.v1.ReadService/ListSleeves?connect=v1&encoding=json&message=%7B%7D", goodToken)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "SLEEVE_ROLE_DEFENSIVE") || !strings.Contains(rec.Body.String(), `"kind":"core"`) {
		t.Errorf("Connect GET ListSleeves = %d %s", rec.Code, rec.Body)
	}
}

// checkReadOnly fails unless m is a unary method with no side effects,
// mapped to a GET under /v1 without a body.
func checkReadOnly(t *testing.T, m protoreflect.MethodDescriptor) {
	t.Helper()
	opts, ok := m.Options().(*descriptorpb.MethodOptions)
	if !ok || opts.GetIdempotencyLevel() != descriptorpb.MethodOptions_NO_SIDE_EFFECTS {
		t.Errorf("%s does not declare no side effects", m.Name())
	}
	rule, ok := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)
	if !ok || !strings.HasPrefix(rule.GetGet(), "/v1/") || rule.GetBody() != "" {
		t.Errorf("%s is not a GET under /v1 without a body: %v", m.Name(), rule)
	}
	if m.IsStreamingClient() || m.IsStreamingServer() {
		t.Errorf("%s streams", m.Name())
	}
}

func TestServiceIsReadOnly(t *testing.T) {
	t.Parallel()
	svc := viewsv1.File_openfief_views_v1_read_service_proto.Services().ByName("ReadService")
	if svc == nil || svc.Methods().Len() == 0 {
		t.Fatal("ReadService has no methods")
	}
	for i := range svc.Methods().Len() {
		checkReadOnly(t, svc.Methods().Get(i))
	}
}

// enumName returns the protobuf name a domain enum value must have: the
// protobuf prefix plus the domain name in upper snake case.
func enumName(prefix, domain string) protoreflect.Name {
	return protoreflect.Name(prefix + strings.ToUpper(strings.ReplaceAll(domain, "-", "_")))
}

// checkEnum fails unless the protobuf enum e names the domain value domain,
// numbered number, with prefix plus its name.
func checkEnum(t *testing.T, e protoreflect.EnumDescriptor, prefix string, number int, domain string) {
	t.Helper()
	v := e.Values().ByName(enumName(prefix, domain))
	if v == nil || int(v.Number()) != number {
		t.Errorf("%s has no %s numbered %d", e.Name(), enumName(prefix, domain), number)
	}
}

func TestEnumsMatch(t *testing.T) {
	t.Parallel()
	for _, r := range []portfolio.Role{portfolio.RoleDefensive, portfolio.RoleOffensive} {
		checkEnum(t, viewsv1.SleeveRole(0).Descriptor(), "SLEEVE_ROLE_", int(r), r.String())
	}
	for k := ledger.BookCash; k <= ledger.BookEquity; k++ {
		checkEnum(t, viewsv1.BookKind(0).Descriptor(), "BOOK_KIND_", int(k), k.String())
	}
	for k := ledger.KindBuy; k <= ledger.KindOpeningLot; k++ {
		checkEnum(t, viewsv1.TransactionKind(0).Descriptor(), "TRANSACTION_KIND_", int(k), k.String())
	}
	for _, s := range []ledgerapp.AssertionStatus{ledgerapp.AssertionMatched, ledgerapp.AssertionFlagged} {
		checkEnum(t, viewsv1.AssertionStatus(0).Descriptor(), "ASSERTION_STATUS_", int(s), s.String())
	}
	for _, b := range []valuation.Basis{valuation.Market, valuation.GrossAccrued} {
		checkEnum(t, viewsv1.ValuationBasis(0).Descriptor(), "VALUATION_BASIS_", int(b), b.String())
	}
	for _, d := range []views.Dimension{views.ByAssetClass, views.ByCurrency, views.ByInstitution, views.BySleeve} {
		checkEnum(t, viewsv1.AllocationDimension(0).Descriptor(), "ALLOCATION_DIMENSION_", int(d), d.String())
	}
	for _, g := range []views.PerformanceGap{views.GapUnvalued, views.GapNoFXRate} {
		checkEnum(t, viewsv1.PerformanceGap(0).Descriptor(), "PERFORMANCE_GAP_", int(g), g.String())
	}
}
