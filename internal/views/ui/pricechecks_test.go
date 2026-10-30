package ui

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errUnexpectedStatus reports a price check filter the UI never asks for.
var errUnexpectedStatus = errors.New("the UI asks for flagged price checks only")

// ListPriceChecks answers one flagged check of the synthetic instrument at
// a 1% threshold, for any status asked; the UI asks for flagged ones.
func (f *fakeRead) ListPriceChecks(_ context.Context, r *connect.Request[viewsv1.ListPriceChecksRequest]) (*connect.Response[viewsv1.ListPriceChecksResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	if r.Msg.GetStatus() != viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED {
		return nil, codeErr(connect.CodeInvalidArgument, errUnexpectedStatus)
	}
	return connect.NewResponse(&viewsv1.ListPriceChecksResponse{Threshold: "0.01", Checks: []*viewsv1.PriceCheck{{
		Date: "2026-03-13", Instrument: syntheticInstrument(),
		Price: &viewsv1.Money{Amount: "84.125", Currency: "GBP"}, PriceSource: "eodhd",
		CrossCheck: &viewsv1.Money{Amount: "86.5", Currency: "GBP"}, CrossCheckSource: "ibkr-main",
		RelativeDifference: "0.028232", Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED,
	}}}), nil
}

// TestPriceChecksPage checks the flagged price checks page (D-22): each
// check with both prices, their sources and the difference in percent,
// and the threshold, for any session; without one, the login page.
func TestPriceChecksPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/price-checks", ""), demoSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Flagged price checks", "2026-03-13", "XX0000000001", "84,125 GBP", "eodhd", "86,5 GBP", "ibkr-main",
		"2.8232%", "by more than 1%")
	wantNoBody(t, rec, "<script>", "style=")
	rec = h.serve(t, request(http.MethodGet, "/price-checks", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
}

// TestOverviewFlagsPriceChecks checks that the overview counts the flagged
// price checks and links them, next to the flagged balance assertions.
func TestOverviewFlagsPriceChecks(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<a href="/price-checks">1 flagged price check(s)</a>`, `<a href="/price-checks">Price checks</a>`)
}

func TestPercent(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, in, want string }{
		{name: "the default threshold", in: "0.01", want: "1%"},
		{name: "a relative difference, every digit kept", in: "0.028232", want: "2.8232%"},
		{name: "below the price", in: "-0.015", want: "-1.5%"},
		{name: "no difference", in: "0", want: "0%"},
		{name: "none", in: "", want: ""},
		{name: "not a decimal", in: "1%", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := percent(tt.in); got != tt.want {
				t.Errorf("percent(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
