package valuation_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Synthetic ISINs with valid check digits; nothing real is implied.
const (
	etfISIN  = "IETEST000026"
	bondISIN = "PL0000000014"
)

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

func holding(t *testing.T, in ledger.Instrument, on, quantity string) ledger.Holding {
	t.Helper()
	return ledger.Holding{Account: "acc-1", Instrument: in, Date: date(t, on), Quantity: kernel.NewQuantity(dec(t, quantity))}
}

func TestMissingReasonNames(t *testing.T) {
	t.Parallel()
	for _, r := range []valuation.MissingReason{valuation.MissingPrice, valuation.MissingFXRate, valuation.MissingIndexReading, valuation.MissingMarginalRate, valuation.MissingLots} {
		got, err := valuation.ParseMissingReason(r.String())
		if err != nil || got != r {
			t.Errorf("ParseMissingReason(%q) = %v, %v", r.String(), got, err)
		}
	}
	if _, err := valuation.ParseMissingReason("none"); !cmp.Equal(valuation.ErrUnknownMissingReason, err, cmpopts.EquateErrors()) {
		t.Errorf("ParseMissingReason(none): error = %v", err)
	}
	if got := valuation.MissingReason(9).String(); got != "MissingReason(9)" {
		t.Errorf("MissingReason(9).String() = %q", got)
	}
}

func TestPriceOnOrBefore(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH").String()
	prices := []valuation.Price{
		{Instrument: etf, Date: date(t, "2026-03-06"), Close: money(t, "100", "USD"), Source: "a"},
		{Instrument: etf, Date: date(t, "2026-03-09"), Close: money(t, "101", "USD"), Source: "a"},
		{Instrument: "OTHER", Date: date(t, "2026-03-08"), Close: money(t, "1", "USD"), Source: "a"},
	}
	tests := []struct {
		name    string
		cutoff  string
		want    int
		wantErr error
	}{
		{name: "same day", cutoff: "2026-03-09", want: 1},
		{name: "weekend takes Friday", cutoff: "2026-03-08", want: 0},
		{name: "before the first price", cutoff: "2026-03-05", wantErr: valuation.ErrNoPrice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.PriceOnOrBefore(prices, etf, date(t, tt.cutoff))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if err == nil {
				if diff := cmp.Diff(prices[tt.want], got); diff != "" {
					t.Errorf("(-want +got):\n%s", diff)
				}
			}
		})
	}
}

// taxableTerms are the terms of a taxable account with no fee rule set.
func taxableTerms() *valuation.AccountTerms {
	return &valuation.AccountTerms{Pack: pl.NewPack(), Wrapper: "taxable"}
}

// marketLot is an open lot of in bought on opened at cost.
func marketLot(t *testing.T, in ledger.Instrument, opened, quantity, cost, currency string) ledger.Lot {
	t.Helper()
	return ledger.Lot{
		Instrument: in, Opened: date(t, opened), Quantity: kernel.NewQuantity(dec(t, quantity)),
		Cost: money(t, cost, currency), Origin: ledger.TransactionKey{Source: "test", ExternalID: "buy-" + opened},
	}
}

func TestValueMarketHolding(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	plnETF := instrument(t, "PLTEST000010", "SYNTHPL")
	prices := []valuation.Price{
		{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "101.37", "USD"), Source: "vendor"},
		{Instrument: plnETF.String(), Date: date(t, "2026-03-06"), Close: money(t, "10.005", "PLN"), Source: "vendor"},
	}
	rates := []valuation.FXRate{rate(t, "USD", "2026-02-27", "4.0000"), rate(t, "USD", "2026-03-06", "4.0123")}
	lots := map[string][]ledger.Lot{
		etf.String():    {marketLot(t, etf, "2026-03-02", "12.5", "1200", "USD")},
		plnETF.String(): {marketLot(t, plnETF, "2026-03-02", "1", "9", "PLN")},
	}
	tests := []struct {
		name string
		h    ledger.Holding
		want valuation.HoldingValuation
	}{
		{
			// Gross: 12.5 x 101.37 USD x 4.0123 = 5084.0856375 -> 5084.09. Cost: 1200 USD at
			// the rate of 2026-02-27 (the day before 2026-03-02 is a Sunday), 4800.00.
			// Belka on 284.09 is 53.9771 -> 53.98: net 5030.11; no fee rule, so liquidation too.
			name: "Sunday uses Friday's close and rate: 12.5 x 101.37 USD x 4.0123 = 5084.0856375",
			h:    holding(t, etf, "2026-03-08", "12.5"),
			want: valuation.HoldingValuation{
				Value: money(t, "1267.125", "USD"), PLN: money(t, "5084.09", "PLN"), Source: "vendor",
				PriceDate: date(t, "2026-03-06"), RateDate: date(t, "2026-03-06"),
				Net: money(t, "5030.11", "PLN"), Liquidation: money(t, "5030.11", "PLN"),
			},
		},
		{
			// 10.01 - 19% x 1.01 (0.1919 -> 0.19) = 9.82.
			name: "a PLN price needs no rate",
			h:    holding(t, plnETF, "2026-03-06", "1"),
			want: valuation.HoldingValuation{
				Value: money(t, "10.005", "PLN"), PLN: money(t, "10.01", "PLN"), Source: "vendor", PriceDate: date(t, "2026-03-06"),
				Net: money(t, "9.82", "PLN"), Liquidation: money(t, "9.82", "PLN"),
			},
		},
		{
			name: "no price yet is a missing input",
			h:    holding(t, etf, "2026-03-05", "1"),
			want: valuation.HoldingValuation{Missing: valuation.MissingPrice},
		},
		{
			name: "no rate yet is a missing input",
			h:    holding(t, etf, "2026-03-06", "1"),
			want: valuation.HoldingValuation{Missing: valuation.MissingFXRate, Source: "vendor", PriceDate: date(t, "2026-03-06")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := rates
			if tt.want.Missing == valuation.MissingFXRate {
				r = nil
			}
			m := valuation.MarketHolding{Prices: prices, Rates: r, Lots: lots[tt.h.Instrument.String()], Terms: taxableTerms()}
			got, err := valuation.ValueMarketHolding(&tt.h, &m)
			if err != nil {
				t.Fatalf("ValueMarketHolding: %v", err)
			}
			want := tt.want
			want.Account, want.Instrument, want.Date, want.Quantity, want.Basis = tt.h.Account, tt.h.Instrument, tt.h.Date, tt.h.Quantity, valuation.Market
			if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(valuation.HoldingValuation{}, "Detail")); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
			if got.Valued() == (got.Detail != "") {
				t.Errorf("Valued() = %v with detail %q", got.Valued(), got.Detail)
			}
		})
	}
}

// TestValueMarketHoldingValues values the USD ETF of TestValueMarketHolding,
// gross 5084.09 PLN on cost 4800.00 PLN, in each wrapper and with fee
// rules: the net and liquidation values (D-77), and what they lack.
func TestValueMarketHoldingValues(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	prices := []valuation.Price{{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "101.37", "USD"), Source: "vendor"}}
	rates := []valuation.FXRate{rate(t, "USD", "2026-02-27", "4.0000"), rate(t, "USD", "2026-03-06", "4.0123")}
	bought := []ledger.Lot{marketLot(t, etf, "2026-03-02", "12.5", "1200", "USD")}
	brokerFee := jurisdiction.FeeRule{Rate: dec(t, "0.0029"), Minimum: money(t, "14", "PLN"), ConversionRate: dec(t, "0.005")}
	pit := jurisdiction.Settings{MarginalRate: dec(t, "0.32"), HasMarginalRate: true}
	tests := []struct {
		name        string
		wrapper     jurisdiction.Wrapper
		fee         jurisdiction.FeeRule
		settings    jurisdiction.Settings
		lots        []ledger.Lot
		lotsUnknown bool
		net, liq    string
		netMissing  valuation.MissingReason
		liqMissing  valuation.MissingReason
	}{
		{name: "taxable without a fee rule", wrapper: "taxable", lots: bought, net: "5030.11", liq: "5030.11"},
		{
			// 0.29% x 5084.09 = 14.74 > 14, plus 0.5% conversion: 40.164311 -> 40.16.
			// Proceeds 5043.93, Belka on 243.93 is 46.3467 -> 46.35: 4997.58.
			name: "taxable: commission and conversion fee", wrapper: "taxable", fee: brokerFee, lots: bought, net: "5030.11", liq: "4997.58",
		},
		{
			// A 5 USD minimum at 4.0123 is 20.06 PLN, above 0.1% of the value.
			// Proceeds 5064.03, Belka on 264.03 is 50.1657 -> 50.17: 5013.86.
			name: "a minimum in USD converted on the day", wrapper: "taxable", lots: bought, net: "5030.11", liq: "5013.86",
			fee: jurisdiction.FeeRule{Rate: dec(t, "0.001"), Minimum: money(t, "5", "USD")},
		},
		{
			name: "a minimum in a currency with no rate leaves liquidation missing", wrapper: "taxable", lots: bought, net: "5030.11",
			fee: jurisdiction.FeeRule{Minimum: money(t, "5", "EUR")}, liqMissing: valuation.MissingFXRate,
		},
		{name: "IKE: untaxed net, a return taxed on its gain", wrapper: "ike", lots: bought, net: "5084.09", liq: "5030.11"},
		{name: "IKZE without the owner's rate", wrapper: "ikze", lots: bought, net: "4575.68", liqMissing: valuation.MissingMarginalRate},
		{name: "IKZE at a 32% marginal rate", wrapper: "ikze", settings: pit, lots: bought, net: "4575.68", liq: "3457.18"},
		{name: "OKI: no Belka tax", wrapper: "oki", lots: bought, net: "5084.09", liq: "5084.09"},
		{
			name: "a lot cost with no rate before its purchase", wrapper: "taxable",
			lots:       []ledger.Lot{marketLot(t, etf, "2026-02-27", "12.5", "1200", "USD")},
			netMissing: valuation.MissingFXRate, liqMissing: valuation.MissingFXRate,
		},
		{name: "unknown lots", wrapper: "taxable", lotsUnknown: true, netMissing: valuation.MissingLots, liqMissing: valuation.MissingLots},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := holding(t, etf, "2026-03-08", "12.5")
			terms := valuation.AccountTerms{Pack: pl.NewPack(), Wrapper: tt.wrapper, Fee: tt.fee, Settings: tt.settings}
			got, err := valuation.ValueMarketHolding(&h, &valuation.MarketHolding{
				Prices: prices, Rates: rates, Lots: tt.lots, LotsUnknown: tt.lotsUnknown, Terms: &terms,
			})
			if err != nil {
				t.Fatalf("ValueMarketHolding: %v", err)
			}
			want := []any{got.PLN.String(), tt.netMissing, tt.liqMissing, got.NetValued(), got.LiquidationValued()}
			gotParts := []any{"5084.09 PLN", got.NetMissing, got.LiquidationMissing, tt.netMissing == 0, tt.liqMissing == 0}
			if diff := cmp.Diff(want, gotParts); diff != "" {
				t.Errorf("gross and what is missing (-want +got):\n%s", diff)
			}
			if tt.netMissing == 0 && !got.Net.Equal(money(t, tt.net, "PLN")) {
				t.Errorf("net = %s, want %s", got.Net, tt.net)
			}
			if tt.liqMissing == 0 && !got.Liquidation.Equal(money(t, tt.liq, "PLN")) {
				t.Errorf("liquidation = %s, want %s", got.Liquidation, tt.liq)
			}
		})
	}
}

func TestValueMarketHoldingConflictingRates(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	prices := []valuation.Price{{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "1", "USD"), Source: "vendor"}}
	rates := []valuation.FXRate{rate(t, "USD", "2026-03-06", "4"), rate(t, "USD", "2026-03-06", "5")}
	h := holding(t, etf, "2026-03-06", "1")
	m := valuation.MarketHolding{Prices: prices, Rates: rates, Terms: taxableTerms()}
	if _, err := valuation.ValueMarketHolding(&h, &m); !cmp.Equal(valuation.ErrConflictingRates, err, cmpopts.EquateErrors()) {
		t.Errorf("error = %v, want ErrConflictingRates", err)
	}
}

// edoTerms are synthetic EDO terms: 6% in the first year, CPI + 2% after,
// 2.00 PLN early-redemption fee.
func edoTerms(t *testing.T) *pl.Terms {
	t.Helper()
	s, err := pl.ParseSeries("EDO0334")
	if err != nil {
		t.Fatal(err)
	}
	return &pl.Terms{Series: s, FirstPeriodRate: dec(t, "0.06"), Margin: dec(t, "0.02"), EarlyRedemptionFee: money(t, "2.00", "PLN")}
}

func bondLot(t *testing.T, in ledger.Instrument, opened, quantity, cost, id string) ledger.Lot {
	t.Helper()
	return ledger.Lot{
		Instrument: in, Opened: date(t, opened), Quantity: kernel.NewQuantity(dec(t, quantity)),
		Cost: money(t, cost, "PLN"), Origin: ledger.TransactionKey{Source: "manual", ExternalID: id},
	}
}

func TestValueBondHolding(t *testing.T) {
	t.Parallel()
	bond := instrument(t, bondISIN, "EDO0334")
	cpi := []valuation.IndexReading{{Kind: valuation.CPI, Year: 2025, Month: time.January, Value: dec(t, "0.05"), Source: "gus"}}
	tests := []struct {
		name     string
		on       string
		wrapper  jurisdiction.Wrapper
		readings []valuation.IndexReading
		idx      pl.Indices // the readings ValueLot should get; nil when missing
		missing  valuation.MissingReason
	}{
		{name: "taxable: sums the lots' three values", on: "2024-09-15", wrapper: "taxable", idx: pl.Indices{}},
		{name: "IKE: no fee", on: "2024-09-15", wrapper: "ike", idx: pl.Indices{}},
		{name: "IKZE without the owner's rate: liquidation missing", on: "2024-09-15", wrapper: "ikze", idx: pl.Indices{}},
		{name: "OKI: the fee, no Belka tax", on: "2024-09-15", wrapper: "oki", idx: pl.Indices{}},
		{name: "second year with its CPI reading", on: "2025-06-01", wrapper: "taxable", readings: cpi, idx: pl.Indices{2: dec(t, "0.05")}},
		{name: "second year without its CPI reading is a missing input", on: "2025-06-01", wrapper: "taxable", missing: valuation.MissingIndexReading},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := holding(t, bond, tt.on, "15")
			b := valuation.BondHolding{
				Lots: []ledger.Lot{
					bondLot(t, bond, "2024-03-15", "10", "1000", "b1"),
					bondLot(t, bond, "2024-03-15", "5", "499.50", "b2"),
				},
				Terms: edoTerms(t), Readings: tt.readings, Account: &valuation.AccountTerms{Pack: pl.NewPack(), Wrapper: tt.wrapper},
			}
			got, err := valuation.ValueBondHolding(&h, &b)
			if err != nil {
				t.Fatalf("ValueBondHolding: %v", err)
			}
			want := valuation.HoldingValuation{
				Account: h.Account, Instrument: bond, Date: h.Date, Basis: valuation.GrossAccrued, Quantity: h.Quantity,
				Source: valuation.StatutorySource, Missing: tt.missing,
			}
			if tt.idx != nil {
				v := lotSum(t, tt.idx, tt.on, tt.wrapper)
				want.Value, want.PLN, want.Net, want.Liquidation = v.Gross, v.Gross, v.Net, v.Liquidation
				if v.LiquidationUnknown {
					want.Liquidation, want.LiquidationMissing = kernel.Money{}, valuation.MissingMarginalRate
				}
			}
			if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(valuation.HoldingValuation{}, "Detail")); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

// lotSum returns the three values, on the date on in wrapper, of the two
// lots TestValueBondHolding values, 10 bonds at 100 PLN and 5 at 99.90 PLN,
// each from pl.ValueLot and the pack's ValueAccrued themselves: the test
// checks summing and wiring, not the bond formulas or the taxes, which
// internal/jurisdiction/pl tests.
func lotSum(t *testing.T, idx pl.Indices, on string, wrapper jurisdiction.Wrapper) jurisdiction.Values {
	t.Helper()
	w, err := pl.ParseTaxWrapper(string(wrapper))
	if err != nil {
		t.Fatal(err)
	}
	var sum [3]kernel.Decimal
	unknown := false
	for _, lot := range []struct{ quantity, price string }{{"10", "100"}, {"5", "99.90"}} {
		v, err := pl.ValueLot(edoTerms(t), idx, kernel.NewQuantity(dec(t, lot.quantity)),
			pl.Purchase{Date: date(t, "2024-03-15"), Price: money(t, lot.price, "PLN")}, w, date(t, on))
		if err != nil {
			t.Fatalf("ValueLot: %v", err)
		}
		income, err := v.Interest.Add(v.Discount)
		if err != nil {
			t.Fatal(err)
		}
		vals, err := pl.NewPack().ValueAccrued(&jurisdiction.Accrued{Wrapper: wrapper, Gross: v.Gross, Income: income, EarlyFee: v.Fee}, &jurisdiction.Settings{})
		if err != nil {
			t.Fatal(err)
		}
		sum[0], sum[1], sum[2] = sum[0].Add(vals.Gross.Amount()), sum[1].Add(vals.Net.Amount()), sum[2].Add(vals.Liquidation.Amount())
		unknown = unknown || vals.LiquidationUnknown
	}
	return jurisdiction.Values{
		Gross: money(t, sum[0].String(), "PLN"), Net: money(t, sum[1].String(), "PLN"), Liquidation: money(t, sum[2].String(), "PLN"),
		LiquidationUnknown: unknown,
	}
}

func TestValueBondHoldingRejects(t *testing.T) {
	t.Parallel()
	bond := instrument(t, bondISIN, "EDO0334")
	h := holding(t, bond, "2024-09-15", "10")
	usdLot := bondLot(t, bond, "2024-03-15", "10", "1000", "b1")
	usdLot.Cost = money(t, "1000", "USD")
	tests := []struct {
		name    string
		lot     ledger.Lot
		wrapper jurisdiction.Wrapper
		wantErr error
	}{
		{name: "an unknown wrapper", lot: bondLot(t, bond, "2024-03-15", "10", "1000", "b1"), wrapper: "ppk", wantErr: pl.ErrUnknownTaxWrapper},
		{name: "cost not in PLN", lot: usdLot, wrapper: "taxable", wantErr: valuation.ErrBondLot},
		{name: "lot outside its series' sale month", lot: bondLot(t, bond, "2024-04-15", "10", "1000", "b1"), wrapper: "taxable", wantErr: pl.ErrSeriesMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := valuation.BondHolding{Lots: []ledger.Lot{tt.lot}, Terms: edoTerms(t), Account: &valuation.AccountTerms{Pack: pl.NewPack(), Wrapper: tt.wrapper}}
			if _, err := valuation.ValueBondHolding(&h, &b); !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValuedThrough(t *testing.T) {
	t.Parallel()
	row := func(on string, reason valuation.MissingReason) valuation.HoldingValuation {
		return valuation.HoldingValuation{Date: date(t, on), Missing: reason}
	}
	tests := []struct {
		name     string
		prior    string
		from     string
		rows     []valuation.HoldingValuation
		want     string
		hasPrior bool
	}{
		{name: "no gaps: today", from: "2026-03-01", rows: []valuation.HoldingValuation{row("2026-03-01", 0)}, want: "2026-03-10"},
		{
			name: "stops the day before the first gap", from: "2026-03-01", want: "2026-03-04",
			rows: []valuation.HoldingValuation{row("2026-03-01", 0), row("2026-03-07", valuation.MissingFXRate), row("2026-03-05", valuation.MissingPrice)},
		},
		{name: "a gap on the first day", from: "2026-03-01", rows: []valuation.HoldingValuation{row("2026-03-01", valuation.MissingPrice)}, want: "2026-02-28"},
		{name: "an earlier gap before from stays", prior: "2026-02-20", hasPrior: true, from: "2026-03-01", want: "2026-02-20"},
		{name: "a prior date up to the day before from is superseded", prior: "2026-02-28", hasPrior: true, from: "2026-03-01", want: "2026-03-10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var prior kernel.Date
			if tt.hasPrior {
				prior = date(t, tt.prior)
			}
			got := valuation.ValuedThrough(prior, tt.hasPrior, date(t, tt.from), date(t, "2026-03-10"), tt.rows)
			if diff := cmp.Diff(date(t, tt.want), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestCoalesce(t *testing.T) {
	t.Parallel()
	got := valuation.Coalesce([]valuation.Revaluation{
		{Account: "acc-2", From: date(t, "2026-03-05")},
		{Account: "acc-1", From: date(t, "2026-03-10")},
		{Account: "acc-1", From: date(t, "2026-03-02")},
		{Account: "acc-1", From: date(t, "2026-03-04")},
	})
	want := []valuation.Revaluation{
		{Account: "acc-1", From: date(t, "2026-03-02")},
		{Account: "acc-2", From: date(t, "2026-03-05")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if got := valuation.Coalesce(nil); len(got) != 0 {
		t.Errorf("Coalesce(nil) = %v", got)
	}
}
