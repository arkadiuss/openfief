package ui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errThresholdTooHigh is the fake core's refusal of a threshold of 1 or
// more.
var errThresholdTooHigh = errors.New("invalid price check threshold: 1")

// otherInstrument is a second synthetic instrument, a thinly traded
// listing with its own 5% threshold.
func otherInstrument() *viewsv1.Instrument {
	return &viewsv1.Instrument{Isin: "XX0000000002", Symbol: "THIN"}
}

// ListPriceChecks answers, for any status, the synthetic instrument's
// flagged check of Sunday 2026-03-15 against Friday's price, and its
// matched check of 2026-03-12; and the thin listing's currency mismatch
// at its own 5%.
func (f *fakeRead) ListPriceChecks(_ context.Context, r *connect.Request[viewsv1.ListPriceChecksRequest]) (*connect.Response[viewsv1.ListPriceChecksResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&viewsv1.ListPriceChecksResponse{Threshold: "0.01", Checks: []*viewsv1.PriceCheck{
		{
			Date: "2026-03-12", Instrument: syntheticInstrument(), PriceDate: "2026-03-12",
			Price: &viewsv1.Money{Amount: "84", Currency: "GBP"}, PriceSource: "eodhd",
			CrossCheck: &viewsv1.Money{Amount: "84.2", Currency: "GBP"}, CrossCheckSource: "ibkr-main",
			RelativeDifference: "0.002381", Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_MATCHED, Threshold: "0.01",
		},
		{
			Date: "2026-03-15", Instrument: syntheticInstrument(), PriceDate: "2026-03-13",
			Price: &viewsv1.Money{Amount: "84.125", Currency: "GBP"}, PriceSource: "eodhd",
			CrossCheck: &viewsv1.Money{Amount: "86.5", Currency: "GBP"}, CrossCheckSource: "ibkr-main",
			RelativeDifference: "0.028232", Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_FLAGGED, Threshold: "0.01",
		},
		{
			Date: "2026-03-13", Instrument: otherInstrument(), PriceDate: "2026-03-13",
			Price: &viewsv1.Money{Amount: "12", Currency: "PLN"}, PriceSource: "eodhd",
			CrossCheck: &viewsv1.Money{Amount: "3.1", Currency: "USD"}, CrossCheckSource: "ibkr-main",
			Status: viewsv1.PriceCheckStatus_PRICE_CHECK_STATUS_CURRENCY_MISMATCH, Threshold: "0.05",
		},
	}}), nil
}

// fakePriceChecks is an in-memory PriceCheckService: it wants an
// owner-scoped session, refuses a threshold of 1 as the core does, and
// records what it took.
type fakePriceChecks struct {
	valuationv1connect.UnimplementedPriceCheckServiceHandler

	mu         sync.Mutex
	thresholds []*valuationv1.SetPriceCheckThresholdRequest
	acks       []*valuationv1.AcknowledgePriceCheckRequest
}

func (f *fakePriceChecks) SetPriceCheckThreshold(_ context.Context, r *connect.Request[valuationv1.SetPriceCheckThresholdRequest]) (*connect.Response[valuationv1.SetPriceCheckThresholdResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if r.Msg.GetThreshold() == "1" {
		return nil, codeErr(connect.CodeInvalidArgument, errThresholdTooHigh)
	}
	f.locked(func() { f.thresholds = append(f.thresholds, r.Msg) })
	return connect.NewResponse(&valuationv1.SetPriceCheckThresholdResponse{}), nil
}

func (f *fakePriceChecks) AcknowledgePriceCheck(_ context.Context, r *connect.Request[valuationv1.AcknowledgePriceCheckRequest]) (*connect.Response[valuationv1.AcknowledgePriceCheckResponse], error) {
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.locked(func() { f.acks = append(f.acks, r.Msg) })
	return connect.NewResponse(&valuationv1.AcknowledgePriceCheckResponse{}), nil
}

// locked runs fn with the fake's lock held.
func (f *fakePriceChecks) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// TestPriceChecksPage checks the price checks page (D-112): the latest
// check of each instrument, with both prices, their sources, the price's
// own date when earlier, the difference in percent and the status, its
// history behind it, and each instrument's threshold (D-110), for any
// session, with the owner's forms only for the owner; without a session,
// the login page.
func TestPriceChecksPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/price-checks", ""), demoSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<title>Price checks", "2026-03-15", "XX0000000001", "84,125 GBP", "(2026-03-13)", "86,5 GBP", "ibkr-main",
		"2.8232%", "by more than 1%", `<span class="pill pill-warn">flagged</span>`, "History: 1 earlier check(s)", "0.2381%",
		"XX0000000002", `<span class="pill pill-warn">currency mismatch</span>`, "Flagged beyond 5%, this instrument's own threshold",
		"1 instrument(s) have a flagged price check")
	wantNoBody(t, rec, "<script>", "style=", "Acknowledge</button>", "/price-checks/threshold")
	rec = h.serve(t, withSession(request(http.MethodGet, "/price-checks", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `action="/price-checks/acknowledge"`, `name="date" value="2026-03-15"`, "Acknowledge</button>",
		`action="/price-checks/threshold"`, `name="threshold" inputmode="decimal" maxlength="12" value="5"`)
	rec = h.serve(t, request(http.MethodGet, "/price-checks", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
}

// TestPriceCheckForms checks the owner's forms: an acknowledgement and a
// threshold in percent reach the core as the core takes them, a cleared
// threshold as empty, and the page answers a malformed percent, the core's
// refusal and a read session.
func TestPriceCheckForms(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/price-checks/acknowledge", "isin=XX0000000001&symbol=SYN&date=2026-03-15&source=ibkr-main"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	for _, in := range []string{"2.5", url.QueryEscape(" 3 %"), ""} {
		rec = h.serve(t, withSession(form("/price-checks/threshold", "isin=XX0000000002&symbol=THIN&threshold="+in), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
	}
	var got []string
	h.checks.locked(func() {
		for _, a := range h.checks.acks {
			got = append(got, a.GetIsin()+" "+a.GetSymbol()+" "+a.GetDate()+" "+a.GetSource())
		}
		for _, s := range h.checks.thresholds {
			got = append(got, s.GetIsin()+" "+s.GetThreshold())
		}
	})
	want := []string{"XX0000000001 SYN 2026-03-15 ibkr-main", "XX0000000002 0.025", "XX0000000002 0.03", "XX0000000002 "}
	if len(got) != len(want) {
		t.Fatalf("handed to the core %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("handed to the core %q, want %q", got[i], want[i])
		}
	}
	for _, bad := range []string{"0", "100", "-1", "two", "1e2"} {
		rec = h.serve(t, withSession(form("/price-checks/threshold", "isin=XX0000000002&symbol=THIN&threshold="+bad), ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
		wantBody(t, rec, "a percent greater than 0 and less than 100")
	}
	rec = h.serve(t, withSession(form("/price-checks/threshold", "isin=XX0000000002&symbol=THIN&threshold=100%25"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	rec = h.serve(t, withSession(form("/price-checks/threshold", "isin=XX0000000002&symbol=THIN&threshold=99.99999999"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	rec = h.serve(t, withSession(form("/price-checks/acknowledge", "isin=XX0000000001&symbol=SYN&date=2026-03-15&source=ibkr-main"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
}

// TestFractionOf checks how a threshold typed in percent becomes the
// fraction the core takes.
func TestFractionOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "1", want: "0.01"},
		{in: "2.5", want: "0.025"},
		{in: "0.0001", want: "0.000001"},
		{in: " 3 % ", want: "0.03"},
		{in: "", want: ""},
		{in: "0.00001", wantErr: true},
		{in: "0", wantErr: true},
		{in: "100", wantErr: true},
		{in: "abc", wantErr: true},
		{in: "1234567890123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := fractionOf(tt.in)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("fractionOf(%q) = %q, %v; want %q, error %v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// TestOverviewFlagsPriceChecks checks that the overview counts the
// instruments whose latest price check is flagged and not acknowledged,
// and links the page, next to the flagged balance assertions.
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
