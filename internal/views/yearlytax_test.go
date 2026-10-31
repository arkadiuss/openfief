package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// taxYearAccount is an OKI account on 2027-03-01 and 2027-03-02 holding a
// treasury bond, a fund priced in USD that no one classified, a fund in
// PLN the owner classified as an investment asset, 200 PLN of cash and, on
// the second day, 10 USD; 300 PLN paid in and 100 PLN withdrawn that day.
func taxYearAccount(t *testing.T) views.TaxYearAccount {
	t.Helper()
	row := func(isin, day, value, currency string, basis valuation.Basis) valuation.HoldingValuation {
		return valuation.HoldingValuation{
			Account: "acc-1", Instrument: instrument(t, isin, "T"), Date: date(t, day), Basis: basis,
			Value: money(t, value, currency), PLN: money(t, "1", "PLN"), Net: money(t, "1", "PLN"), Liquidation: money(t, "1", "PLN"),
		}
	}
	bond := "PLEDO0536000"
	usd := money(t, "0", "USD").Currency()
	return views.TaxYearAccount{
		Account: "acc-1", Wrapper: "oki", From: date(t, "2027-03-01"), To: date(t, "2027-03-02"),
		Vals: []valuation.HoldingValuation{
			row(bond, "2027-03-01", "1000.00", "PLN", valuation.GrossAccrued),
			row(isinPL, "2027-03-01", "100", "USD", valuation.Market),
			row(isinIE, "2027-03-01", "50", "PLN", valuation.Market),
			row(bond, "2027-03-02", "1001.00", "PLN", valuation.GrossAccrued),
			row(isinPL, "2027-03-02", "110", "USD", valuation.Market),
			row(isinIE, "2027-03-02", "55", "PLN", valuation.Market),
		},
		Cash: []views.CashBalance{
			{Date: date(t, "2027-03-01"), Amount: money(t, "200", "PLN")},
			{Date: date(t, "2027-03-02"), Amount: money(t, "200", "PLN")},
			{Date: date(t, "2027-03-02"), Amount: money(t, "10", "USD")},
		},
		Flows: []views.CashFlow{
			{Account: "acc-1", Date: date(t, "2027-03-01"), Amount: money(t, "5000", "PLN"), Seeding: true},
			{Account: "acc-1", Date: date(t, "2027-03-02"), Amount: money(t, "300", "PLN")},
			{Account: "acc-1", Date: date(t, "2027-03-02"), Amount: money(t, "-100", "PLN")},
		},
		Rates: []valuation.FXRate{
			{Currency: usd, Date: date(t, "2027-02-26"), Rate: kernel.NewDecimal(400, -2), Source: "test"},
			{Currency: usd, Date: date(t, "2027-03-01"), Rate: kernel.NewDecimal(410, -2), Source: "test"},
			{Currency: usd, Date: date(t, "2027-03-02"), Rate: kernel.NewDecimal(450, -2), Source: "test"},
		},
		Classes: map[string]jurisdiction.ExemptionClass{isinIE: pl.Investment},
	}
}

// TestTaxDays checks the days of a yearly tax (D-127, D-132). Each day's
// foreign amounts convert at the rate of the day before: on 2027-03-01 the
// last rate on or before 2027-02-28 (4.00), so the USD fund is 400.00; on
// 2027-03-02 the rate of 2027-03-01 (4.10), not that day's 4.50: 451.00 and
// 41.00 of cash. The bond and the PLN cash of an account holding bonds are
// savings assets; the owner's fund is an investment asset; the USD fund,
// which no one classified, and USD cash count as other. Opening cash is no
// payment.
func TestTaxDays(t *testing.T) {
	t.Parallel()
	a := taxYearAccount(t)
	got, unclassified, err := views.TaxDays(&a, pl.NewPack())
	if err != nil {
		t.Fatalf("TaxDays: %v", err)
	}
	day := func(d, savings, investment, other, in, out string) jurisdiction.TaxDay {
		return jurisdiction.TaxDay{Date: date(t, d), Values: []jurisdiction.ClassValue{
			{Class: pl.Savings, Value: money(t, savings, "PLN")},
			{Class: pl.Investment, Value: money(t, investment, "PLN")},
			{Class: pl.Other, Value: money(t, other, "PLN")},
		}, In: money(t, in, "PLN"), Out: money(t, out, "PLN")}
	}
	want := jurisdiction.TaxAccount{Account: "acc-1", Days: []jurisdiction.TaxDay{
		day("2027-03-01", "1200.00", "50.00", "400.00", "0", "0"),
		day("2027-03-02", "1201.00", "55.00", "492.00", "300.00", "100.00"),
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("days (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{isinPL}, unclassified); diff != "" {
		t.Errorf("unclassified (-want +got):\n%s", diff)
	}
	a.Rates = nil
	if _, _, err := views.TaxDays(&a, pl.NewPack()); !errors.Is(err, views.ErrTaxConversion) {
		t.Errorf("without USD rates: %v, want ErrTaxConversion", err)
	}
}
