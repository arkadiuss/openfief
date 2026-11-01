package ui

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ListYearlyTaxes answers OKI's tax for the year asked as if the act were
// in force: 30 000 of savings assets and 90 000 of investment assets,
// 100 000 exempt, 20 000 taxed at 0.85%, through 2026-03-31, with one fund
// no one classified.
func (f *fakeRead) ListYearlyTaxes(_ context.Context, r *connect.Request[viewsv1.ListYearlyTaxesRequest]) (*connect.Response[viewsv1.ListYearlyTaxesResponse], error) {
	if err := f.check(r.Header()); err != nil {
		return nil, err
	}
	class := func(c, amount string) *viewsv1.ClassAmount {
		return &viewsv1.ClassAmount{ExemptionClass: c, Amount: pln(amount)}
	}
	return connect.NewResponse(&viewsv1.ListYearlyTaxesResponse{Taxes: []*viewsv1.YearlyTax{{
		TaxWrapper: "oki", Year: r.Msg.GetYear(), InForce: true, Days: 90, Through: "2026-03-31", Average: pln("120000.00"),
		Averages: []*viewsv1.ClassAmount{class("savings", "30000.00"), class("investment", "90000.00"), class("other", "0")},
		Exempt:   []*viewsv1.ClassAmount{class("savings", "25000.00"), class("investment", "75000.00"), class("other", "0")},
		Taxed:    pln("20000.00"), Rate: "0.0085", Tax: pln("170.00"), UnclassifiedInstruments: []string{"PLTEST000010"},
	}}}), nil
}

func (f *fakeSettings) SetExemptionClass(_ context.Context, r *connect.Request[valuationv1.SetExemptionClassRequest]) (*connect.Response[valuationv1.SetExemptionClassResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	if r.Msg.GetExemptionClass() == "bonds" {
		return nil, codeErr(connect.CodeInvalidArgument, errBadRateFake)
	}
	f.locked(func() {
		if f.classes == nil {
			f.classes = map[string]string{}
		}
		f.classes[r.Msg.GetIsin()] = r.Msg.GetExemptionClass()
	})
	return connect.NewResponse(&valuationv1.SetExemptionClassResponse{}), nil
}

// TestYearlyTaxSettings reads the settings page's yearly taxes (D-127):
// the averages and exemptions per class, the rate and tax, and the fund to
// classify (D-132); classifying it saves the class and lists it as the
// owner's.
func TestYearlyTaxSettings(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `id="taxes-title">Yearly taxes in 2026`, "Savings assets", "30 000,00 PLN", "25 000,00 PLN", "taxed 20 000,00 PLN",
		"0.85%", "170,00 PLN", "Counted through 2026-03-31", "PLTEST000010", "not classified", `action="/settings/exemption-class"`)
	post := h.serve(t, withSession(form("/settings/exemption-class", "isin=PLTEST000010&exemption_class=investment"), ownerSession))
	wantStatus(t, post, http.StatusSeeOther)
	var got map[string]string
	h.settings.locked(func() { got = map[string]string{"PLTEST000010": h.settings.classes["PLTEST000010"]} })
	if got["PLTEST000010"] != "investment" {
		t.Errorf("classes = %v", got)
	}
	again := h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantBody(t, again, `<option value="investment" selected>Investment assets</option>`)
	refused := h.serve(t, withSession(form("/settings/exemption-class", "isin=PLTEST000010&exemption_class=bonds"), ownerSession))
	wantStatus(t, refused, http.StatusBadRequest)
}

// TestLossesOffsetLine reads the losses-offset line (D-128) under the
// overview's accounts and on the account page.
func TestLossesOffsetLine(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<th scope="row" colspan="3">Losses offset</th>`, "25,00 PLN", "Portfolio total", "The total counts tax as PIT-38 does")
	acc := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, acc, http.StatusOK)
	wantBody(t, acc, "Of which losses offset", "10,00 PLN")
}
