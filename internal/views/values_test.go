package views_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

func TestValueKindNames(t *testing.T) {
	t.Parallel()
	got := []string{views.Gross.String(), views.Net.String(), views.Liquidation.String(), views.ValueKind(9).String()}
	if diff := cmp.Diff([]string{"gross", "net", "liquidation", "ValueKind(9)"}, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

// ikzeRow is a valued IKZE holding valuation of 1000 PLN gross, 900 net,
// whose liquidation value lacks the owner's marginal rate.
func ikzeRow(t *testing.T) valuation.HoldingValuation {
	t.Helper()
	return valuation.HoldingValuation{
		Account: "acc-1", Instrument: instrument(t, isinIE, "ETF1"), Date: date(t, "2026-03-02"), Basis: valuation.Market,
		Value: money(t, "250", "USD"), PLN: money(t, "1000", "PLN"), Net: money(t, "900", "PLN"),
		LiquidationMissing: valuation.MissingMarginalRate,
	}
}

func TestHoldingValue(t *testing.T) {
	t.Parallel()
	row := ikzeRow(t)
	gap := valuation.HoldingValuation{Missing: valuation.MissingPrice}
	tests := []struct {
		name        string
		row         *valuation.HoldingValuation
		kind        views.ValueKind
		want        string
		wantMissing valuation.MissingReason
	}{
		{name: "gross", row: &row, kind: views.Gross, want: "1000"},
		{name: "net", row: &row, kind: views.Net, want: "900"},
		{name: "liquidation without the marginal rate", row: &row, kind: views.Liquidation, wantMissing: valuation.MissingMarginalRate},
		{name: "a row with no price has no value of any kind", row: &gap, kind: views.Net, wantMissing: valuation.MissingPrice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, missing, err := views.HoldingValue(tt.row, tt.kind)
			if err != nil {
				t.Fatalf("HoldingValue: %v", err)
			}
			if missing != tt.wantMissing || (tt.want != "" && !got.Equal(money(t, tt.want, "PLN"))) {
				t.Errorf("HoldingValue = %s, missing %s; want %s, missing %s", got, missing, tt.want, tt.wantMissing)
			}
		})
	}
	if _, _, err := views.HoldingValue(&row, 0); !errors.Is(err, views.ErrUnknownValueKind) {
		t.Errorf("an unknown kind: %v", err)
	}
}

func TestCashValue(t *testing.T) {
	t.Parallel()
	v := jurisdiction.Values{Gross: money(t, "500", "PLN"), Net: money(t, "450", "PLN"), LiquidationUnknown: true}
	gross, okGross, err1 := views.CashValue(&v, views.Gross)
	net, okNet, err2 := views.CashValue(&v, views.Net)
	_, okLiq, err3 := views.CashValue(&v, views.Liquidation)
	_, _, err4 := views.CashValue(&v, 0)
	if err := errors.Join(err1, err2, err3); err != nil {
		t.Fatal(err)
	}
	if !gross.Equal(v.Gross) || !net.Equal(v.Net) || !okGross || !okNet || okLiq || !errors.Is(err4, views.ErrUnknownValueKind) {
		t.Errorf("CashValue = %s %v, %s %v, liquidation %v, unknown kind %v", gross, okGross, net, okNet, okLiq, err4)
	}
}

// TestTotalsByDateWithCash sums an IKZE holding and cash per date: gross
// and net with the cash, liquidation missing for want of the marginal rate,
// cash without an FX rate in no total but per currency.
func TestTotalsByDateWithCash(t *testing.T) {
	t.Parallel()
	on := date(t, "2026-03-02")
	cash := []views.DayCash{
		{Date: on, Amount: money(t, "100", "PLN"), Values: jurisdiction.Values{Gross: money(t, "100", "PLN"), Net: money(t, "90", "PLN"), LiquidationUnknown: true}},
		{Date: on, Amount: money(t, "5", "CHF"), Unconverted: true},
		{Date: date(t, "2026-03-03"), Amount: money(t, "100", "PLN"), Values: jurisdiction.Values{Gross: money(t, "100", "PLN"), Net: money(t, "100", "PLN"), Liquidation: money(t, "100", "PLN")}},
	}
	got, err := views.TotalsByDate([]valuation.HoldingValuation{ikzeRow(t)}, cash, pln(t))
	if err != nil {
		t.Fatalf("TotalsByDate: %v", err)
	}
	want := []views.DateTotal{
		{
			Date: on, Reporting: money(t, "1100", "PLN"), ByCurrency: []kernel.Money{money(t, "5", "CHF"), money(t, "100", "PLN"), money(t, "250", "USD")},
			Net: money(t, "990", "PLN"), Liquidation: money(t, "0", "PLN"), LiquidationMissing: true, Cash: money(t, "100", "PLN"), CashUnconverted: true,
		},
		{
			Date: date(t, "2026-03-03"), Reporting: money(t, "100", "PLN"), ByCurrency: []kernel.Money{money(t, "100", "PLN")},
			Net: money(t, "100", "PLN"), Liquidation: money(t, "100", "PLN"), Cash: money(t, "100", "PLN"),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestCashOnDays(t *testing.T) {
	t.Parallel()
	usd := money(t, "1000", "USD").Currency()
	h := func(id, day string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: "acc-1", Date: date(t, day)}
	}
	deposit, err1 := ledger.NewDeposit(h("d1", "2026-03-02"), money(t, "1000", "USD"))
	buy, err2 := ledger.NewBuy(h("b1", "2026-03-04"), instrument(t, isinIE, "ETF1"), kernel.NewQuantity(kernel.NewDecimal(10, 0)),
		money(t, "1000", "USD"), kernel.NewMoney(kernel.Decimal{}, usd))
	plnIn, err3 := ledger.NewDeposit(h("d2", "2026-03-03"), money(t, "50", "PLN"))
	if err := errors.Join(err1, err2, err3); err != nil {
		t.Fatal(err)
	}
	got := views.CashOnDays([]ledger.Transaction{buy, deposit, plnIn}, date(t, "2026-03-01"), date(t, "2026-03-04"))
	want := []views.CashBalance{
		{Date: date(t, "2026-03-02"), Amount: money(t, "1000", "USD")},
		{Date: date(t, "2026-03-03"), Amount: money(t, "50", "PLN")},
		{Date: date(t, "2026-03-03"), Amount: money(t, "1000", "USD")},
		{Date: date(t, "2026-03-04"), Amount: money(t, "50", "PLN")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

// TestValueHoldingsOfAKind values the IKZE holding at its net value, and
// leaves it out of a liquidation allocation with its reason.
func TestValueHoldingsOfAKind(t *testing.T) {
	t.Parallel()
	acc := ledger.Account{ID: "acc-1", Institution: "broker-a", Name: "Synthetic", Sleeve: "core"}
	row := ikzeRow(t)
	holdings := []ledger.Holding{{Account: acc.ID, Instrument: row.Instrument, Date: row.Date, Quantity: kernel.NewQuantity(kernel.NewDecimal(1, 0))}}
	items, unvalued, err := views.ValueHoldings(acc, holdings, []valuation.HoldingValuation{row}, row.Date, views.Net)
	if err != nil || len(unvalued) != 0 || len(items) != 1 || !items[0].Reporting.Equal(money(t, "900", "PLN")) || !items[0].Value.Equal(row.Value) {
		t.Errorf("net: items %v, unvalued %v, %v", items, unvalued, err)
	}
	items, unvalued, err = views.ValueHoldings(acc, holdings, []valuation.HoldingValuation{row}, row.Date, views.Liquidation)
	want := []views.Unvalued{{Account: acc.ID, Holding: holdings[0], Missing: "marginal_rate"}}
	if diff := cmp.Diff(want, unvalued); err != nil || len(items) != 0 || diff != "" {
		t.Errorf("liquidation: items %v, %v (-want +got):\n%s", items, err, diff)
	}
}
