package ui

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// errBadRateFake is the fake SettingsService's refusal of a rate.
var errBadRateFake = errors.New("the marginal rate must be a decimal above 0 and below 1")

// fakeSettings is an in-memory SettingsService offering two number
// formats; choosing one changes what the fake read API's GetSettings
// answers, as the core's settings do.
type fakeSettings struct {
	valuationv1connect.UnimplementedSettingsServiceHandler

	read *fakeRead
	mu   sync.Mutex
	rate string
}

// fakeFormats are the formats the fake offers, the Polish default first.
func fakeFormats() []*valuationv1.NumberFormat {
	return []*valuationv1.NumberFormat{
		{Key: "space-comma", GroupSeparator: " ", DecimalSeparator: ","},
		{Key: "comma-dot", GroupSeparator: ",", DecimalSeparator: "."},
	}
}

func (f *fakeSettings) GetOwnerSettings(_ context.Context, r *connect.Request[valuationv1.GetOwnerSettingsRequest]) (*connect.Response[valuationv1.GetOwnerSettingsResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return connect.NewResponse(&valuationv1.GetOwnerSettingsResponse{
		ReportingCurrency: "PLN", NumberFormat: fakeFormats()[0], NumberFormatDefault: true,
		NumberFormats: fakeFormats(), TaxWrappers: []string{"taxable", "ike", "ikze", "oki"}, MarginalRate: f.rate,
	}), nil
}

func (f *fakeSettings) SetNumberFormat(_ context.Context, r *connect.Request[valuationv1.SetNumberFormatRequest]) (*connect.Response[valuationv1.SetNumberFormatResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	for _, nf := range fakeFormats() {
		if nf.GetKey() == r.Msg.GetKey() {
			f.read.locked(func() {
				f.read.format = &viewsv1.NumberFormat{Key: nf.GetKey(), GroupSeparator: nf.GetGroupSeparator(), DecimalSeparator: nf.GetDecimalSeparator()}
			})
			return connect.NewResponse(&valuationv1.SetNumberFormatResponse{NumberFormat: nf}), nil
		}
	}
	return nil, codeErr(connect.CodeInvalidArgument, errBadRateFake)
}

func (f *fakeSettings) SetMarginalRate(_ context.Context, r *connect.Request[valuationv1.SetMarginalRateRequest]) (*connect.Response[valuationv1.SetMarginalRateResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	if r.Msg.GetRate() != "0.32" && r.Msg.GetRate() != "0.12" {
		return nil, codeErr(connect.CodeInvalidArgument, errBadRateFake)
	}
	f.locked(func() { f.rate = r.Msg.GetRate() })
	return connect.NewResponse(&valuationv1.SetMarginalRateResponse{}), nil
}

func (f *fakeSettings) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// TestOverviewThreeValues reads the overview: the total's gross, net and
// a missing liquidation value, the account's three values, the value
// chart from go-live with its switch starting on gross, and the
// performance chart apart from it, in the Polish default format.
func TestOverviewThreeValues(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "1 500,00 PLN", "1 400,00 PLN", "<dt>Liquidation</dt><dd><span class=\"warn-text\">missing</span>",
		"1 000,00 PLN", "950,00 PLN", `id="value-title">Value over time`, `id="history-title">Performance`,
		`aria-current="true">Gross</a>`, `/value-chart?chart=net`, `&#34;group&#34;:&#34; &#34;,&#34;decimal&#34;:&#34;,&#34;`,
		"Gross value with cash")
	if got := seen(h.read, func() []string { return valueNames(h.read.values) }); len(got) != 1 || got[0] != "VALUE_KIND_GROSS" {
		t.Errorf("allocation values asked = %q, want gross", got)
	}
}

// valueNames names value kinds.
func valueNames(ks []viewsv1.ValueKind) []string {
	out := make([]string, len(ks))
	for i, k := range ks {
		out[i] = k.String()
	}
	return out
}

// TestValueSwitches picks the liquidation allocation and the net value
// chart, as whole pages and as htmx fragments.
func TestValueSwitches(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/?value=liquidation&chart=net", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Liquidation value</th>", "Net value with cash")
	frag := h.serve(t, htmx(withSession(request(http.MethodGet, "/value-chart?chart=liquidation&from=2026-03-30&date=2026-03-31", ""), readSession)))
	wantStatus(t, frag, http.StatusOK)
	wantBody(t, frag, `id="value-chart"`, "The line skips 2 day(s) without a Liquidation value")
	wantNoBody(t, frag, "<html")
	all := h.serve(t, htmx(withSession(request(http.MethodGet, "/value-chart?chart=all&from=2026-03-30&date=2026-03-31", ""), readSession)))
	wantStatus(t, all, http.StatusOK)
	wantBody(t, all, `aria-current="true">All three</a>`, "Net value with cash", "Liquidation value with cash",
		"The liquidation line skips 2 day(s)")
	alloc := h.serve(t, htmx(withSession(request(http.MethodGet, "/allocation?by=sleeve&value=net&date=2026-03-31", ""), readSession)))
	wantStatus(t, alloc, http.StatusOK)
	wantBody(t, alloc, "Net value</th>", `aria-current="true">Net</a>`)
	if got := seen(h.read, func() []string { return valueNames(h.read.values) }); len(got) != 2 || got[1] != "VALUE_KIND_NET" {
		t.Errorf("allocation values asked = %q", got)
	}
	plain := h.serve(t, withSession(request(http.MethodGet, "/value-chart?chart=net&date=2026-03-31", ""), readSession))
	wantStatus(t, plain, http.StatusSeeOther)
}

// TestAccountThreeValuesAndMarks reads the account page: the account's
// wrapper and fee rule, its value with net and a missing liquidation value
// with the reason, and the holding no longer updated from ibkr-old.
func TestAccountThreeValuesAndMarks(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "no tax wrapper", "none set, so sales cost nothing", "950,00 PLN",
		"missing: needs your marginal PIT rate", "no longer updated", `<span class="mono">ibkr-old</span>`, "12 345,67 EUR")
}

// TestNumberFormatSetting picks comma-dot on the settings page: the page
// then writes amounts with it; the marginal rate is typed in percent.
func TestNumberFormatSetting(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, page, http.StatusOK)
	wantBody(t, page, `id="values-title"`, "5 000,00 PLN", "5,000.00 PLN", "The jurisdiction pack's default",
		"an IKZE's liquidation value stays missing until it is set, as it is now")
	wantStatus(t, h.serve(t, withSession(form("/settings/number-format", "number_format=comma-dot"), ownerSession)), http.StatusSeeOther)
	acc := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantBody(t, acc, "12,345.67 EUR", "1,000.00 PLN")
	wantStatus(t, h.serve(t, withSession(form("/settings/marginal-rate", "marginal_rate=32"), ownerSession)), http.StatusSeeOther)
	if got := seenOne(h.settings, func() string { return h.settings.rate }); got != "0.32" {
		t.Errorf("marginal rate = %q, want 0.32", got)
	}
	bad := h.serve(t, withSession(form("/settings/marginal-rate", "marginal_rate=150"), ownerSession))
	wantStatus(t, bad, http.StatusBadRequest)
	wantBody(t, bad, "marginal rate must be a decimal")
	wantStatus(t, h.serve(t, withSession(form("/settings/marginal-rate", "marginal_rate=32"), readSession)), http.StatusForbidden)
}

// TestAccountTermsForms posts a new account's wrapper and fee rule in
// percent, and a change that leaves the fee rule out.
func TestAccountTermsForms(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), ownerSession))
	wantBody(t, page, `name="tax_wrapper"`, `<option value="ikze">IKZE</option>`, `name="fee_rate"`, "Tax wrapper")
	rec := h.serve(t, withSession(form("/accounts", "name=Retirement&sleeve=core&tax_wrapper=ikze&fee_rate=0,29&fee_minimum=5&fee_minimum_currency=pln&fee_conversion=0.5"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	created := h.accounts.lastCreate()
	rule := created.GetFeeRule()
	if created.GetTaxWrapper() != "ikze" || rule.GetRate() != "0.0029" || rule.GetMinimum() != "5" || rule.GetMinimumCurrency() != "PLN" || rule.GetConversionRate() != "0.005" {
		t.Errorf("CreateAccount = %v", created)
	}
	wantStatus(t, h.serve(t, withSession(form("/accounts/"+accountID+"/edit", "name=Renamed&sleeve=core&tax_wrapper=ike"), ownerSession)), http.StatusSeeOther)
	if updated := h.accounts.lastUpdate(); updated.GetTaxWrapper() != "ike" || updated.GetFeeRule() != nil {
		t.Errorf("UpdateAccount = %v, want ike with the fee rule kept", updated)
	}
}

// TestNumberFormats writes amounts in each format: only the separators
// change, never the digits, the sign or the currency.
func TestNumberFormats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		group, decimal, in, want string
	}{
		{" ", ",", "-1234567.80", "-1 234 567,80"},
		{",", ".", "1234567.8", "1,234,567.8"},
		{".", ",", "1234.5", "1.234,5"},
		{"'", ".", "12345", "12'345"},
		{"", ",", "1234.50", "1234,50"},
		{" ", ",", "not a number", "not a number"},
		{"", "", "1234.5", "1 234.5"},
	}
	for _, tt := range tests {
		t.Run(tt.group+tt.decimal+tt.in, func(t *testing.T) {
			t.Parallel()
			if got := separators(tt.group, tt.decimal).amount(tt.in); got != tt.want {
				t.Errorf("amount(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
	f := separators(" ", ",")
	if got := f.money(&viewsv1.Money{Amount: "5000.00", Currency: "PLN"}); got != "5 000,00 PLN" {
		t.Errorf("money = %q", got)
	}
	if got := f.money(nil); got != "" {
		t.Errorf("money(nil) = %q", got)
	}
	if got := f.feeRule(&viewsv1.FeeRule{Rate: "0.0029", Minimum: &viewsv1.Money{Amount: "5.00", Currency: "PLN"}, ConversionRate: "0.005"}); got != "0.29%, at least 5,00 PLN, plus 0.5% to convert a foreign currency" {
		t.Errorf("feeRule = %q", got)
	}
}

// TestValueLabels names value kinds, missing reasons and wrappers.
func TestValueLabels(t *testing.T) {
	t.Parallel()
	got := []string{
		parseValueKind("").Key, parseValueKind("net").Key, parseValueKind("bogus").Key,
		missingLabel("marginal_rate"), missingLabel("other"), wrapperLabel("ikze"), wrapperLabel("x"),
		fromPercent("0,29"), fromPercent("abc"), toPercent("0.32"), toPercent("0"),
	}
	want := []string{"gross", "net", "gross", "needs your marginal PIT rate", "other", "IKZE", "x", "0.0029", "abc", "32", ""}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("label %d = %q, want %q", i, got[i], want[i])
		}
	}
	total := &viewsv1.DateTotal{ValueReporting: pln("3"), NetReporting: pln("2")}
	if totalOf(total, parseValueKind("net")).GetAmount() != "2" || totalOf(total, parseValueKind("liquidation")) != nil ||
		totalOf(total, parseValueKind("gross")).GetAmount() != "3" {
		t.Errorf("totalOf picks the wrong value")
	}
}
