package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The synthetic instrument the price check test stores prices of.
const (
	checkISIN   = "IETEST000026"
	checkSymbol = "TSTB"
)

// priceCheckAPI is a generated PriceCheckService client.
type priceCheckAPI struct {
	valuationv1connect.PriceCheckServiceClient
}

// priceCheckClient returns a PriceCheckService client of srv that presents
// token as the view plugin's.
func priceCheckClient(srv *httptest.Server, token string) priceCheckAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return priceCheckAPI{valuationv1connect.NewPriceCheckServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// storeCheckedPrices stores Friday's close of the instrument and IBKR's
// mark of the Sunday after, 3% above it.
func storeCheckedPrices(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	store, err := valuationpg.NewStore(pool)
	if err != nil {
		t.Fatalf("valuation store: %v", err)
	}
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	friday, err1 := kernel.ParseDate("2026-03-06")
	sunday, err2 := kernel.ParseDate("2026-03-08")
	if err1 != nil || err2 != nil {
		t.Fatalf("dates: %v %v", err1, err2)
	}
	in := checkISIN + "/" + checkSymbol
	if err := store.SavePrices(t.Context(), []valuation.Price{
		{Instrument: in, Date: friday, Close: kernel.NewMoney(kernel.NewDecimal(100, 0), usd), Source: "eodhd"},
	}, nil); err != nil {
		t.Fatalf("SavePrices: %v", err)
	}
	if err := store.SaveCrossCheckPrices(t.Context(), []valuation.CrossCheckPrice{
		{Instrument: in, Date: sunday, Close: kernel.NewMoney(kernel.NewDecimal(103, 0), usd), Source: "ibkr-main"},
	}); err != nil {
		t.Fatalf("SaveCrossCheckPrices: %v", err)
	}
}

// listedCheck returns the one price check the read API lists.
func listedCheck(t *testing.T, srv *httptest.Server, session string) *viewsv1.PriceCheck {
	t.Helper()
	var got viewsv1.ListPriceChecksResponse
	getJSON(t, srv, "/v1/price-checks", session, &got)
	if len(got.GetChecks()) != 1 {
		t.Fatalf("GET /v1/price-checks = %v, want one check", got.GetChecks())
	}
	return got.GetChecks()[0]
}

// TestPriceCheckService checks, through the real core and PostgreSQL, that
// a cross-check price is compared with the latest price on or before its
// date (D-111), that the owner's threshold for the instrument replaces 1%
// (D-110), and that the owner acknowledges a flagged check, recorded
// append-only (D-112), while other callers are refused.
func TestPriceCheckService(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	code := enrollCode(t, enrollCmd(t, dsn))
	_, srv, pool := startCoreForUI(t, dsn, func(*serveConfig) {})
	session := passkeyLogin(t, authClient(srv.URL, srv.Client(), uiCaller), passkeyEnroll(t, authClient(srv.URL, srv.Client(), uiCaller), code))
	storeCheckedPrices(t, pool)
	ctx := t.Context()

	got := listedCheck(t, srv, session)
	if got.GetStatus() != viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED || got.GetPriceDate() != "2026-03-06" ||
		got.GetThreshold() != "0.01" || got.GetAcknowledged() {
		t.Fatalf("the Sunday mark = %v; want flagged at 1%% against Friday's price, unacknowledged", got)
	}
	c := priceCheckClient(srv, uiCaller)
	ack := &valuationv1.AcknowledgePriceCheckRequest{
		Origin: localUI, SessionToken: session, Isin: checkISIN, Symbol: checkSymbol, Date: "2026-03-08", Source: "ibkr-main",
	}
	if _, err := c.AcknowledgePriceCheck(ctx, connect.NewRequest(ack)); err != nil {
		t.Fatalf("AcknowledgePriceCheck: %v", err)
	}
	if !listedCheck(t, srv, session).GetAcknowledged() {
		t.Error("the check is not acknowledged after AcknowledgePriceCheck")
	}
	checkThresholds(t, srv, session)
	// Matched now, at 5%: it can't be acknowledged; an unknown one is not found.
	if _, err := c.AcknowledgePriceCheck(ctx, connect.NewRequest(ack)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("acknowledging a matched check = %v, want InvalidArgument", err)
	}
	ack.Date = "2026-03-09"
	if _, err := c.AcknowledgePriceCheck(ctx, connect.NewRequest(ack)); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("acknowledging no check = %v, want NotFound", err)
	}
}

// checkThresholds sets the instrument's threshold to 5%, which matches the
// 3% gap, and checks that malformed thresholds and other callers are
// refused.
func checkThresholds(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	ctx := t.Context()
	set := &valuationv1.SetPriceCheckThresholdRequest{Origin: localUI, SessionToken: session, Isin: checkISIN, Symbol: checkSymbol, Threshold: "0.05"}
	if _, err := priceCheckClient(srv, uiCaller).SetPriceCheckThreshold(ctx, connect.NewRequest(set)); err != nil {
		t.Fatalf("SetPriceCheckThreshold: %v", err)
	}
	if got := listedCheck(t, srv, session); got.GetStatus() != viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_MATCHED || got.GetThreshold() != "0.05" {
		t.Errorf("at the owner's 5%% the check = %v, want matched", got)
	}
	refusals := []struct {
		name, caller, session, threshold, date string
		want                                   connect.Code
	}{
		{name: "a threshold of 100%", caller: uiCaller, session: session, threshold: "1", want: connect.CodeInvalidArgument},
		{name: "a threshold in percent", caller: uiCaller, session: session, threshold: "5%", want: connect.CodeInvalidArgument},
		{name: "not a view plugin", caller: "wrong", session: session, threshold: "0.02", want: connect.CodePermissionDenied},
		{name: "an unknown session", caller: uiCaller, session: "ofsess_nope_nope", threshold: "0.02", want: connect.CodeUnauthenticated},
	}
	for _, r := range refusals {
		_, err := priceCheckClient(srv, r.caller).SetPriceCheckThreshold(ctx, connect.NewRequest(&valuationv1.SetPriceCheckThresholdRequest{
			Origin: localUI, SessionToken: r.session, Isin: checkISIN, Symbol: checkSymbol, Threshold: r.threshold,
		}))
		if connect.CodeOf(err) != r.want {
			t.Errorf("%s: SetPriceCheckThreshold = %v, want %s", r.name, err, r.want)
		}
	}
}
