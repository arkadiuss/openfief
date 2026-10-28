package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// etfLedger is acc-1 buying 10 units of a synthetic USD-quoted ETF on
// Monday 2026-03-02.
func etfLedger(t *testing.T) *fakeLedger {
	t.Helper()
	etf := instrument(t, etfISIN, "SYNTH")
	return &fakeLedger{logs: map[ledger.AccountID][]ledger.VersionedTransaction{
		acc1: {buy(t, acc1, "b1", "2026-03-02", etf, "10", "990")},
	}}
}

func todayFunc(t *testing.T, s string) func() kernel.Date {
	t.Helper()
	d := date(t, s)
	return func() kernel.Date { return d }
}

func revaluer(t *testing.T, l *fakeLedger, s *fakeStore, basis valuation.Basis, today string) *app.Revaluer {
	t.Helper()
	r, err := app.NewRevaluer(l, s, basis, todayFunc(t, today))
	if err != nil {
		t.Fatalf("NewRevaluer: %v", err)
	}
	return r
}

func usdPrice(t *testing.T, on, closing string) valuation.Price {
	t.Helper()
	return valuation.Price{Instrument: instrument(t, etfISIN, "SYNTH").String(), Date: date(t, on), Close: money(t, closing, "USD"), Source: "vendor"}
}

func TestRevalueMissingInputs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		prices      []string
		rates       []string
		wantThrough string
		wantMissing int
	}{
		{name: "every input: valued through today", prices: []string{"2026-03-02"}, rates: []string{"2026-03-02"}, wantThrough: "2026-03-06"},
		{name: "no price on the first day", prices: []string{"2026-03-03"}, rates: []string{"2026-03-02"}, wantThrough: "2026-03-01", wantMissing: 1},
		{name: "no FX rate at all", prices: []string{"2026-03-02"}, wantThrough: "2026-03-01", wantMissing: 5},
		{name: "FX rates from Wednesday", prices: []string{"2026-03-02"}, rates: []string{"2026-03-04"}, wantThrough: "2026-03-01", wantMissing: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := marketStore(t, tt.prices, tt.rates)
			got, err := revaluer(t, etfLedger(t), s, 0, "2026-03-06").Revalue(t.Context(), acc1, date(t, "2026-03-02"))
			if err != nil {
				t.Fatalf("Revalue: %v", err)
			}
			want := app.RevalueResult{Start: date(t, "2026-03-02"), ValuedThrough: date(t, tt.wantThrough), Rows: 5, Missing: tt.wantMissing}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("result (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(date(t, tt.wantThrough), s.through[acc1]); diff != "" {
				t.Errorf("stored valued-through date (-want +got):\n%s", diff)
			}
			if len(s.rows[acc1]) != 5 {
				t.Errorf("stored %d rows, want 5", len(s.rows[acc1]))
			}
		})
	}
}

// marketStore returns a store with a USD close of 100 for the ETF on each of
// prices and a USD rate of 4 on each of rates.
func marketStore(t *testing.T, prices, rates []string) *fakeStore {
	t.Helper()
	s := newFakeStore()
	for _, d := range prices {
		s.prices = append(s.prices, usdPrice(t, d, "100"))
	}
	for _, d := range rates {
		s.rates = append(s.rates, rate(t, "USD", d, "4.0000"))
	}
	return s
}

func TestRevalueStoresMarketValues(t *testing.T) {
	t.Parallel()
	s := newFakeStore()
	s.prices = []valuation.Price{usdPrice(t, "2026-03-02", "100"), usdPrice(t, "2026-03-03", "101.5")}
	s.rates = []valuation.FXRate{rate(t, "USD", "2026-03-02", "4.0000")}
	if _, err := revaluer(t, etfLedger(t), s, 0, "2026-03-03").Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	etf := instrument(t, etfISIN, "SYNTH")
	row := func(on, value, pln, priceDate string) valuation.HoldingValuation {
		return valuation.HoldingValuation{
			Account: acc1, Instrument: etf, Date: date(t, on), Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, "10")),
			Value: money(t, value, "USD"), PLN: money(t, pln, "PLN"), Source: "vendor",
			PriceDate: date(t, priceDate), RateDate: date(t, "2026-03-02"),
		}
	}
	want := []valuation.HoldingValuation{
		row("2026-03-02", "1000", "4000", "2026-03-02"),
		row("2026-03-03", "1015", "4060", "2026-03-03"),
	}
	if diff := cmp.Diff(want, s.rows[acc1]); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestRevalueIsIdempotent(t *testing.T) {
	t.Parallel()
	s := newFakeStore()
	s.prices = []valuation.Price{usdPrice(t, "2026-03-02", "100"), usdPrice(t, "2026-03-04", "102")}
	s.rates = []valuation.FXRate{rate(t, "USD", "2026-03-02", "4.0000")}
	r := revaluer(t, etfLedger(t), s, 0, "2026-03-06")
	if _, err := r.Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	first := s.rows[acc1]
	for _, from := range []string{"2026-03-02", "2026-03-04", "2020-01-01"} {
		if _, err := r.Revalue(t.Context(), acc1, date(t, from)); err != nil {
			t.Fatalf("Revalue from %s: %v", from, err)
		}
		if diff := cmp.Diff(first, s.rows[acc1]); diff != "" {
			t.Errorf("rows after revaluing again from %s (-first +again):\n%s", from, diff)
		}
	}
}

func TestRevalueStartsAtTheFirstTransaction(t *testing.T) {
	t.Parallel()
	s := newFakeStore()
	got, err := revaluer(t, etfLedger(t), s, 0, "2026-03-03").Revalue(t.Context(), acc1, date(t, "2020-01-01"))
	if err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	want := app.RevalueResult{Start: date(t, "2026-03-02"), ValuedThrough: date(t, "2026-03-01"), Rows: 2, Missing: 2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	empty, err := revaluer(t, &fakeLedger{}, s, 0, "2026-03-03").Revalue(t.Context(), acc2, date(t, "2026-03-01"))
	if err != nil {
		t.Fatalf("Revalue of an empty account: %v", err)
	}
	if diff := cmp.Diff(app.RevalueResult{Start: date(t, "2026-03-01"), ValuedThrough: date(t, "2026-03-03")}, empty); diff != "" {
		t.Errorf("empty account (-want +got):\n%s", diff)
	}
}

func TestRevalueKeepsAnEarlierGap(t *testing.T) {
	t.Parallel()
	s := newFakeStore()
	s.prices = []valuation.Price{usdPrice(t, "2026-03-04", "100")}
	s.rates = []valuation.FXRate{rate(t, "USD", "2026-03-02", "4.0000")}
	r := revaluer(t, etfLedger(t), s, 0, "2026-03-06")
	steps := []struct {
		from        string
		price       string
		wantThrough string
	}{
		{from: "2026-03-02", wantThrough: "2026-03-01"},
		{from: "2026-03-05", price: "2026-03-05", wantThrough: "2026-03-01"},
		{from: "2026-03-02", price: "2026-03-02", wantThrough: "2026-03-06"},
	}
	for _, st := range steps {
		if st.price != "" {
			s.prices = append(s.prices, usdPrice(t, st.price, "100"))
		}
		got, err := r.Revalue(t.Context(), acc1, date(t, st.from))
		if err != nil {
			t.Fatalf("Revalue from %s: %v", st.from, err)
		}
		if diff := cmp.Diff(date(t, st.wantThrough), got.ValuedThrough); diff != "" {
			t.Errorf("from %s (-want +got):\n%s", st.from, diff)
		}
	}
}

func TestRevalueBonds(t *testing.T) {
	t.Parallel()
	bond := instrument(t, bondISIN, "EDO0334")
	series, err := pl.ParseSeries("EDO0334")
	if err != nil {
		t.Fatal(err)
	}
	terms := pl.Terms{Series: series, FirstPeriodRate: dec(t, "0.06"), Margin: dec(t, "0.02"), EarlyRedemptionFee: money(t, "2.00", "PLN")}
	l := &fakeLedger{logs: map[ledger.AccountID][]ledger.VersionedTransaction{
		acc1: {buy(t, acc1, "b1", "2024-03-15", bond, "10", "1000")},
	}}
	tests := []struct {
		name      string
		basis     valuation.Basis
		wrapper   pl.TaxWrapper
		wantBasis valuation.Basis
	}{
		{name: "net if redeemed by default", wantBasis: valuation.NetIfRedeemed, wrapper: pl.Taxable},
		{name: "inside IKE no fee or tax", wantBasis: valuation.NetIfRedeemed, wrapper: pl.IKE},
		{name: "gross accrued when configured", basis: valuation.GrossAccrued, wantBasis: valuation.GrossAccrued, wrapper: pl.Taxable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newFakeStore()
			s.terms[bond.ISIN] = terms
			s.wrappers[acc1] = tt.wrapper
			got, err := revaluer(t, l, s, tt.basis, "2024-09-15").Revalue(t.Context(), acc1, date(t, "2024-09-14"))
			if err != nil {
				t.Fatalf("Revalue: %v", err)
			}
			if got.Missing != 0 || got.Rows != 2 {
				t.Fatalf("result = %+v, want 2 rows and nothing missing", got)
			}
			last := s.rows[acc1][1]
			value := bondLotValue(t, &terms, tt.wrapper, tt.wantBasis == valuation.GrossAccrued)
			if last.Basis != tt.wantBasis || last.Source != valuation.StatutorySource || !last.PLN.Equal(value) || !last.Value.Equal(value) {
				t.Errorf("row on 2024-09-15 = %+v, want %s on %s", last, value, tt.wantBasis)
			}
		})
	}
}

// bondLotValue is pl.ValueLot's gross or net value on 2024-09-15 of the lot
// TestRevalueBonds values: the test checks wiring, not the bond formulas.
func bondLotValue(t *testing.T, terms *pl.Terms, wrapper pl.TaxWrapper, gross bool) kernel.Money {
	t.Helper()
	v, err := pl.ValueLot(terms, nil, kernel.NewQuantity(dec(t, "10")),
		pl.Purchase{Date: date(t, "2024-03-15"), Price: money(t, "100", "PLN")}, wrapper, date(t, "2024-09-15"))
	if err != nil {
		t.Fatal(err)
	}
	if gross {
		return v.Gross
	}
	return v.Net
}

func TestRevalueErrors(t *testing.T) {
	t.Parallel()
	if _, err := app.NewRevaluer(&fakeLedger{}, newFakeStore(), valuation.Market, todayFunc(t, "2026-03-02")); !cmp.Equal(valuation.ErrBasisNotForBonds, err, cmpopts.EquateErrors()) {
		t.Errorf("NewRevaluer(market): error = %v", err)
	}
	broken := etfLedger(t)
	broken.err = errStoreDown
	if _, err := revaluer(t, broken, newFakeStore(), 0, "2026-03-02").Revalue(t.Context(), acc1, date(t, "2026-03-02")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("failing log: error = %v", err)
	}
	s := newFakeStore()
	s.err = errStoreDown
	if _, err := revaluer(t, etfLedger(t), s, 0, "2026-03-02").Revalue(t.Context(), acc1, date(t, "2026-03-02")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("failing store: error = %v", err)
	}
}
