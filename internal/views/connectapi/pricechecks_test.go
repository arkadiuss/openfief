package connectapi_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// fakePriceChecks answers fixed price checks at a 1% threshold.
type fakePriceChecks []valuation.PriceCheck

func (f fakePriceChecks) PriceChecks(context.Context) ([]valuation.PriceCheck, error) { return f, nil }

func (fakePriceChecks) Threshold() kernel.Decimal { return kernel.NewDecimal(1, -2) }

// priceCheckAPI is the read API serving checks, or comparing no prices when
// checks is nil: a matched and a flagged check of the synthetic ETF1, from
// the eodhd and ibkr-main instances.
func priceCheckAPI(t *testing.T, checks app.PriceChecks) http.Handler {
	t.Helper()
	var opts []app.Option
	if checks != nil {
		opts = append(opts, app.WithPriceChecks(checks))
	}
	h, err := connectapi.New(newViews(t, newLedger(t), opts...), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))),
		connectapi.WithRequestIDs(func() string { return "req-1" }))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return h
}

// syntheticChecks are a matched check on 2026-03-02, a flagged one the
// owner acknowledged on Sunday 2026-03-08 against Friday's price, and a
// currency mismatch on 2026-03-09.
func syntheticChecks(t *testing.T) fakePriceChecks {
	t.Helper()
	in := instrument(t, isinIE, "ETF1").String()
	check := func(priced, on, closing, mark, code, diff string, status valuation.PriceCheckStatus) valuation.PriceCheck {
		return valuation.PriceCheck{
			Price:      valuation.Price{Instrument: in, Date: date(t, priced), Close: money(t, closing, "USD"), Source: "eodhd"},
			CrossCheck: valuation.CrossCheckPrice{Instrument: in, Date: date(t, on), Close: money(t, mark, code), Source: "ibkr-main"},
			Difference: dec(t, diff), Threshold: dec(t, "0.01"), Status: status,
		}
	}
	flagged := check("2026-03-06", "2026-03-08", "10.50", "10.8", "USD", "0.028571", valuation.PriceFlagged)
	flagged.Acknowledged = true
	return fakePriceChecks{
		check("2026-03-02", "2026-03-02", "10.50", "10.52", "USD", "0.001905", valuation.PriceMatched),
		flagged,
		check("2026-03-09", "2026-03-09", "10.50", "9.70", "EUR", "0", valuation.PriceCurrencyMismatch),
	}
}

func TestConnectListPriceChecks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(priceCheckAPI(t, syntheticChecks(t)))
	t.Cleanup(srv.Close)
	c := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL)
	got, err := c.ListPriceChecks(context.Background(), authed(goodToken, &viewsv1.ListPriceChecksRequest{
		Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED,
	}))
	if err != nil {
		t.Fatalf("ListPriceChecks: %v", err)
	}
	want := &viewsv1.ListPriceChecksResponse{Threshold: "0.01", Checks: []*viewsv1.PriceCheck{{
		Date: "2026-03-08", Instrument: &viewsv1.Instrument{Isin: isinIE, Symbol: "ETF1"},
		Price: &viewsv1.Money{Amount: "10.5", Currency: "USD"}, PriceSource: "eodhd", PriceDate: "2026-03-06",
		CrossCheck: &viewsv1.Money{Amount: "10.8", Currency: "USD"}, CrossCheckSource: "ibkr-main",
		RelativeDifference: "0.028571", Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED,
		Threshold: "0.01", Acknowledged: true,
	}}}
	if diff := cmp.Diff(want, got.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("ListPriceChecks(flagged) (-want +got):\n%s", diff)
	}
	all, err := c.ListPriceChecks(context.Background(), authed(goodToken, &viewsv1.ListPriceChecksRequest{}))
	if err != nil || len(all.Msg.GetChecks()) != 3 {
		t.Errorf("ListPriceChecks(all) = %v, %v; want every check", all, err)
	}
	_, bad := c.ListPriceChecks(context.Background(), authed(goodToken, &viewsv1.ListPriceChecksRequest{Status: viewsv1.PriceCheckStatus(9)}))
	if connect.CodeOf(bad) != connect.CodeInvalidArgument {
		t.Errorf("ListPriceChecks with an unknown status = %v, want invalid argument", bad)
	}
	_, unauthenticated := c.ListPriceChecks(context.Background(), authed("", &viewsv1.ListPriceChecksRequest{}))
	if connect.CodeOf(unauthenticated) != connect.CodeUnauthenticated {
		t.Errorf("ListPriceChecks without a token = %v, want unauthenticated", unauthenticated)
	}

	none := httptest.NewServer(priceCheckAPI(t, nil))
	t.Cleanup(none.Close)
	empty, err := viewsv1connect.NewReadServiceClient(none.Client(), none.URL).
		ListPriceChecks(context.Background(), authed(goodToken, &viewsv1.ListPriceChecksRequest{}))
	if err != nil || empty.Msg.GetThreshold() != "" || len(empty.Msg.GetChecks()) != 0 {
		t.Errorf("ListPriceChecks in a core comparing no prices = %v, %v; want no threshold and no checks", empty, err)
	}
}

// TestConnectListCurrencyMismatches checks that a pair in two currencies
// is listed with its own status and no difference (D-111).
func TestConnectListCurrencyMismatches(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(priceCheckAPI(t, syntheticChecks(t)))
	t.Cleanup(srv.Close)
	mismatch, err := viewsv1connect.NewReadServiceClient(srv.Client(), srv.URL).ListPriceChecks(context.Background(),
		authed(goodToken, &viewsv1.ListPriceChecksRequest{Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_CURRENCY_MISMATCH}))
	if err != nil || len(mismatch.Msg.GetChecks()) != 1 || mismatch.Msg.GetChecks()[0].GetRelativeDifference() != "" {
		t.Errorf("ListPriceChecks(currency mismatch) = %v, %v; want one check with no difference (D-111)", mismatch, err)
	}
}

func TestRESTPriceChecks(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: priceCheckAPI(t, syntheticChecks(t))}
	req, rec := a.serve(t, http.MethodGet, "/v1/price-checks?status=PRICE_CHECK_STATUS_FLAGGED", goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/price-checks = %d\n%s", rec.Code, rec.Body)
	}
	c.check(t, req, rec)
	for _, want := range []string{
		`"threshold":"0.01"`, `"relativeDifference":"0.028571"`, `"status":"PRICE_CHECK_STATUS_FLAGGED"`,
		`"priceDate":"2026-03-06"`, `"acknowledged":true`,
	} {
		if !bytes.Contains(rec.Body.Bytes(), []byte(want)) {
			t.Errorf("GET /v1/price-checks = %s, want %s", rec.Body, want)
		}
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`"date":"2026-03-02"`)) {
		t.Errorf("GET /v1/price-checks?status=flagged = %s, want no matched check", rec.Body)
	}
}

func TestPriceCheckEnumsMatch(t *testing.T) {
	t.Parallel()
	for _, s := range []valuation.PriceCheckStatus{valuation.PriceMatched, valuation.PriceFlagged, valuation.PriceCurrencyMismatch} {
		checkEnum(t, viewsv1.PriceCheckStatus(0).Descriptor(), "PRICE_CHECK_STATUS_", int(s), s.String())
	}
}
