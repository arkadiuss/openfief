package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
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

func revaluer(t *testing.T, l *fakeLedger, s *fakeStore, today string) *app.Revaluer {
	t.Helper()
	return app.NewRevaluer(l, s, pl.NewPack(), todayFunc(t, today))
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
			got, err := revaluer(t, etfLedger(t), s, "2026-03-06").Revalue(t.Context(), acc1, date(t, "2026-03-02"))
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
	s.rates = []valuation.FXRate{rate(t, "USD", "2026-02-27", "4.0000"), rate(t, "USD", "2026-03-02", "4.0000")}
	if _, err := revaluer(t, etfLedger(t), s, "2026-03-03").Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	etf := instrument(t, etfISIN, "SYNTH")
	// The cost, 990 USD, converts at the rate of 2026-02-27 for tax, the last
	// before the purchase: 3960 PLN. No fee rule: liquidation equals net.
	row := func(on, value, pln, priceDate, net string) valuation.HoldingValuation {
		return valuation.HoldingValuation{
			Account: acc1, Instrument: etf, Date: date(t, on), Basis: valuation.Market, Quantity: kernel.NewQuantity(dec(t, "10")),
			Value: money(t, value, "USD"), PLN: money(t, pln, "PLN"), Source: "vendor",
			PriceDate: date(t, priceDate), RateDate: date(t, "2026-03-02"),
			Net: money(t, net, "PLN"), Liquidation: money(t, net, "PLN"),
		}
	}
	want := []valuation.HoldingValuation{
		row("2026-03-02", "1000", "4000", "2026-03-02", "3992.40"),
		row("2026-03-03", "1015", "4060", "2026-03-03", "4041.00"),
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
	r := revaluer(t, etfLedger(t), s, "2026-03-06")
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
	got, err := revaluer(t, etfLedger(t), s, "2026-03-03").Revalue(t.Context(), acc1, date(t, "2020-01-01"))
	if err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	want := app.RevalueResult{Start: date(t, "2026-03-02"), ValuedThrough: date(t, "2026-03-01"), Rows: 2, Missing: 2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	empty, err := revaluer(t, &fakeLedger{}, s, "2026-03-03").Revalue(t.Context(), acc2, date(t, "2026-03-01"))
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
	r := revaluer(t, etfLedger(t), s, "2026-03-06")
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
	pit := jurisdiction.Settings{MarginalRate: dec(t, "0.32"), HasMarginalRate: true}
	tests := []struct {
		name     string
		wrapper  jurisdiction.Wrapper // empty: none recorded
		settings jurisdiction.Settings
	}{
		{name: "taxable when no wrapper is recorded"},
		{name: "inside IKE", wrapper: "ike"},
		{name: "inside IKZE, without the owner's marginal rate", wrapper: "ikze"},
		{name: "inside IKZE, with the owner's marginal rate", wrapper: "ikze", settings: pit},
		{name: "inside OKI", wrapper: "oki"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newFakeStore()
			s.terms[bond.ISIN] = terms
			s.settings = tt.settings
			wrapper := jurisdiction.Wrapper("taxable")
			if tt.wrapper != "" {
				wrapper, s.wrappers[acc1] = tt.wrapper, tt.wrapper
			}
			got, err := revaluer(t, l, s, "2024-09-15").Revalue(t.Context(), acc1, date(t, "2024-09-14"))
			if err != nil {
				t.Fatalf("Revalue: %v", err)
			}
			if got.Missing != 0 || got.Rows != 2 {
				t.Fatalf("result = %+v, want 2 rows and nothing missing", got)
			}
			v := bondLotValues(t, &terms, wrapper, &tt.settings)
			checkBondRow(t, &s.rows[acc1][1], &v)
		})
	}
}

// checkBondRow checks a stored bond row against the values the pack gives.
func checkBondRow(t *testing.T, row *valuation.HoldingValuation, v *jurisdiction.Values) {
	t.Helper()
	if row.Basis != valuation.GrossAccrued || row.Source != valuation.StatutorySource || !row.PLN.Equal(v.Gross) || !row.Value.Equal(v.Gross) {
		t.Errorf("row = %+v, want gross %s", row, v.Gross)
	}
	if !row.Net.Equal(v.Net) {
		t.Errorf("net = %s, want %s", row.Net, v.Net)
	}
	unknown := row.LiquidationMissing == valuation.MissingMarginalRate
	if v.LiquidationUnknown != unknown || (!unknown && !row.Liquidation.Equal(v.Liquidation)) {
		t.Errorf("liquidation = %s (missing %s), want %s (unknown %v)", row.Liquidation, row.LiquidationMissing, v.Liquidation, v.LiquidationUnknown)
	}
}

// bondLotValues are pl.ValueLot's and the pack's three values on
// 2024-09-15 of the lot TestRevalueBonds values: the test checks wiring,
// not the bond formulas or the taxes.
func bondLotValues(t *testing.T, terms *pl.Terms, wrapper jurisdiction.Wrapper, s *jurisdiction.Settings) jurisdiction.Values {
	t.Helper()
	w, err := pl.ParseTaxWrapper(string(wrapper))
	if err != nil {
		t.Fatal(err)
	}
	v, err := pl.ValueLot(terms, nil, kernel.NewQuantity(dec(t, "10")),
		pl.Purchase{Date: date(t, "2024-03-15"), Price: money(t, "100", "PLN")}, w, date(t, "2024-09-15"))
	if err != nil {
		t.Fatal(err)
	}
	income, err := v.Interest.Add(v.Discount)
	if err != nil {
		t.Fatal(err)
	}
	values, err := pl.NewPack().ValueAccrued(&jurisdiction.Accrued{Wrapper: wrapper, Gross: v.Gross, Income: income, EarlyFee: v.Fee}, s)
	if err != nil {
		t.Fatal(err)
	}
	return values
}

// TestRevalueFeeRuleAndShortHoldings checks that the account's fee rule
// reaches its liquidation values, and that a holding sold short, which has
// no lots, is stored with its gross value and its net and liquidation
// values missing for want of lots.
func TestRevalueFeeRuleAndShortHoldings(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	s := marketStore(t, []string{"2026-03-02"}, []string{"2026-02-27", "2026-03-02"})
	s.fees[acc1] = jurisdiction.FeeRule{Rate: dec(t, "0.01")}
	if _, err := revaluer(t, etfLedger(t), s, "2026-03-02").Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue: %v", err)
	}
	// 10 x 100 USD x 4 = 4000 PLN on a cost of 990 USD x 4 = 3960. Net: Belka
	// on 40 is 7.60. Liquidation: 1% commission 40.00, proceeds 3960, no gain.
	row := s.rows[acc1][0]
	if !row.Net.Equal(money(t, "3992.40", "PLN")) || !row.Liquidation.Equal(money(t, "3960.00", "PLN")) {
		t.Errorf("net %s and liquidation %s, want 3992.40 and 3960.00 PLN", row.Net, row.Liquidation)
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "short"}, Account: acc2, Date: date(t, "2026-03-02")}
	sell, err := ledger.NewSell(h, etf, kernel.NewQuantity(dec(t, "5")), money(t, "500", "USD"), money(t, "0", "USD"))
	if err != nil {
		t.Fatal(err)
	}
	short := &fakeLedger{logs: map[ledger.AccountID][]ledger.VersionedTransaction{acc2: {{Transaction: sell, Version: 1}}}}
	if _, err := revaluer(t, short, s, "2026-03-02").Revalue(t.Context(), acc2, date(t, "2026-03-02")); err != nil {
		t.Fatalf("Revalue of a short holding: %v", err)
	}
	got := s.rows[acc2][0]
	if !got.Valued() || got.NetMissing != valuation.MissingLots || got.LiquidationMissing != valuation.MissingLots {
		t.Errorf("short holding = %+v, want gross with net and liquidation missing lots", got)
	}
}

func TestRevalueErrors(t *testing.T) {
	t.Parallel()
	broken := etfLedger(t)
	broken.err = errStoreDown
	if _, err := revaluer(t, broken, newFakeStore(), "2026-03-02").Revalue(t.Context(), acc1, date(t, "2026-03-02")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("failing log: error = %v", err)
	}
	s := newFakeStore()
	s.err = errStoreDown
	if _, err := revaluer(t, etfLedger(t), s, "2026-03-02").Revalue(t.Context(), acc1, date(t, "2026-03-02")); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("failing store: error = %v", err)
	}
}
