package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// listedRow is a valued holding priced at market on 2026-03-02 with its
// gross, net and liquidation values and its cost basis, in PLN.
func listedRow(t *testing.T, account ledger.AccountID, isin, gross, net, cost string) valuation.HoldingValuation {
	t.Helper()
	return valuation.HoldingValuation{
		Account: account, Instrument: instrument(t, isin, "T"), Date: date(t, "2026-03-02"), Basis: valuation.Market,
		Value: money(t, gross, "PLN"), PLN: money(t, gross, "PLN"), Net: money(t, net, "PLN"), Liquidation: money(t, net, "PLN"),
		TaxCost: money(t, cost, "PLN"), SaleCost: money(t, "0", "PLN"),
	}
}

// TestTaxTotals checks the losses-offset line (D-128) on 2026-03-02.
// acc-1 holds A, worth 10 000 at a cost of 8 000 (own tax 380, net 9 620),
// and realized a 500 loss this year (a 1 000 gain in 2025 doesn't count):
// base 1 500, tax 285, total 9 715, offset +95. acc-2 holds B, worth 3 000
// at a cost of 4 000, and 1 000 of cash: no tax, total 4 000, offset 0.
// The portfolio offsets both: base 500, tax 95, total 13 905, which is
// 190 more than the accounts' totals, 13 715.
func TestTaxTotals(t *testing.T) {
	t.Parallel()
	acc1 := views.TaxedAccount{
		Wrapper: "taxable",
		Vals:    []valuation.HoldingValuation{listedRow(t, "acc-1", isinPL, "10000", "9620", "8000")},
		Realized: []valuation.RealizedGain{
			{Date: date(t, "2025-06-01"), Gain: money(t, "1000", "PLN")}, {Date: date(t, "2026-02-01"), Gain: money(t, "-500", "PLN")},
		},
	}
	cash := views.DayCash{
		Date: date(t, "2026-03-02"), Amount: money(t, "1000", "PLN"),
		Values: jurisdiction.Values{Gross: money(t, "1000", "PLN"), Net: money(t, "1000", "PLN"), Liquidation: money(t, "1000", "PLN")},
	}
	acc2 := views.TaxedAccount{
		Wrapper: "taxable", Vals: []valuation.HoldingValuation{listedRow(t, "acc-2", isinIE, "3000", "3000", "4000")},
		Cash: []views.DayCash{cash},
	}
	tests := []struct {
		name           string
		accounts       []views.TaxedAccount
		net, netOffset string
		liq, liqOffset string
		netMissing     bool
	}{
		{name: "one account: its rows", accounts: []views.TaxedAccount{acc1}, net: "9715", netOffset: "95", liq: "9715", liqOffset: "95"},
		{name: "a loss alone", accounts: []views.TaxedAccount{acc2}, net: "4000", netOffset: "0", liq: "4000", liqOffset: "0"},
		{name: "the portfolio: the accounts' totals", accounts: []views.TaxedAccount{acc1, acc2}, net: "13905", netOffset: "190", liq: "13905", liqOffset: "190"},
		{
			name: "a realized gain without a rate leaves both missing", netMissing: true,
			accounts: []views.TaxedAccount{{Wrapper: "taxable", Vals: acc1.Vals, RealizedMissing: true}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := taxedTotal(t, tt.accounts)
			if tt.netMissing {
				if !got.NetMissing || !got.LiquidationMissing {
					t.Errorf("total = %+v, want net and liquidation missing", got)
				}
				return
			}
			want := []string{tt.net, tt.netOffset, tt.liq, tt.liqOffset}
			have := []string{got.Net.Amount().String(), got.NetOffset.Amount().String(), got.Liquidation.Amount().String(), got.LiquidationOffset.Amount().String()}
			if diff := cmp.Diff(want, have); diff != "" {
				t.Errorf("net, offset, liquidation, offset (-want +got):\n%s", diff)
			}
		})
	}
}

// taxedTotal returns the accounts' total on their one date, taxed.
func taxedTotal(t *testing.T, accounts []views.TaxedAccount) views.DateTotal {
	t.Helper()
	var vals []valuation.HoldingValuation
	var cash []views.DayCash
	for i := range accounts {
		vals, cash = append(vals, accounts[i].Vals...), append(cash, accounts[i].Cash...)
	}
	totals, err := views.TotalsByDate(vals, cash, pln(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := views.TaxTotals(totals, accounts, pl.NewPack(), &jurisdiction.Settings{}); err != nil {
		t.Fatalf("TaxTotals: %v", err)
	}
	return totals[0]
}

// TestPaidIn checks what an IKE's early return is taxed above (D-128): an
// opening lot of 100 USD opened 2025-12-30, at the rate of 2025-12-29
// (4.00), 400; a 1 000 PLN deposit; a 200 PLN withdrawal. A deposit in a
// currency without a rate is reported.
func TestPaidIn(t *testing.T) {
	t.Parallel()
	h := func(id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: "acc-1", Date: date(t, day)}
	}
	deposit, err1 := ledger.NewDeposit(h("d", "2026-01-05"), money(t, "1000", "PLN"))
	withdrawal, err2 := ledger.NewWithdrawal(h("w", "2026-02-01"), money(t, "200", "PLN"))
	opening, err3 := ledger.NewOpeningLot(h("o", "2026-01-02"), instrument(t, isinIE, "T"), kernel.NewQuantity(kernel.NewDecimal(1, 0)),
		money(t, "100", "USD"), date(t, "2025-12-30"))
	euros, err4 := ledger.NewDeposit(h("e", "2026-01-06"), money(t, "5", "EUR"))
	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			t.Fatal(err)
		}
	}
	rates := []valuation.FXRate{{Currency: money(t, "0", "USD").Currency(), Date: date(t, "2025-12-29"), Rate: kernel.NewDecimal(4, 0), Source: "test"}}
	got, missing, err := views.PaidIn([]ledger.Transaction{withdrawal, deposit, opening}, rates, pl.NewPack())
	if err != nil || missing {
		t.Fatalf("PaidIn: %v, %v", missing, err)
	}
	want := []views.Payment{
		{Date: date(t, "2026-01-02"), Amount: money(t, "400.00", "PLN")},
		{Date: date(t, "2026-01-05"), Amount: money(t, "1000.00", "PLN")},
		{Date: date(t, "2026-02-01"), Amount: money(t, "-200.00", "PLN")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if _, missing, err := views.PaidIn([]ledger.Transaction{euros}, rates, pl.NewPack()); err != nil || !missing {
		t.Errorf("a EUR deposit without a rate: %v, %v; want missing", missing, err)
	}
}
