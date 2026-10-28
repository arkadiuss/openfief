package valuation_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
	for _, r := range []valuation.MissingReason{valuation.MissingPrice, valuation.MissingFXRate, valuation.MissingIndexReading} {
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

func TestValueMarketHolding(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	plnETF := instrument(t, "PLTEST000010", "SYNTHPL")
	prices := []valuation.Price{
		{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "101.37", "USD"), Source: "vendor"},
		{Instrument: plnETF.String(), Date: date(t, "2026-03-06"), Close: money(t, "10.005", "PLN"), Source: "vendor"},
	}
	rates := []valuation.FXRate{rate(t, "USD", "2026-03-06", "4.0123")}
	tests := []struct {
		name string
		h    ledger.Holding
		want valuation.HoldingValuation
	}{
		{
			name: "Sunday uses Friday's close and rate: 12.5 x 101.37 USD x 4.0123 = 5084.0856375",
			h:    holding(t, etf, "2026-03-08", "12.5"),
			want: valuation.HoldingValuation{
				Value: money(t, "1267.125", "USD"), PLN: money(t, "5084.09", "PLN"), Source: "vendor",
				PriceDate: date(t, "2026-03-06"), RateDate: date(t, "2026-03-06"),
			},
		},
		{
			name: "a PLN price needs no rate",
			h:    holding(t, plnETF, "2026-03-06", "1"),
			want: valuation.HoldingValuation{
				Value: money(t, "10.005", "PLN"), PLN: money(t, "10.01", "PLN"), Source: "vendor", PriceDate: date(t, "2026-03-06"),
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
			got, err := valuation.ValueMarketHolding(&tt.h, prices, r)
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

func TestValueMarketHoldingConflictingRates(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	prices := []valuation.Price{{Instrument: etf.String(), Date: date(t, "2026-03-06"), Close: money(t, "1", "USD"), Source: "vendor"}}
	rates := []valuation.FXRate{rate(t, "USD", "2026-03-06", "4"), rate(t, "USD", "2026-03-06", "5")}
	h := holding(t, etf, "2026-03-06", "1")
	if _, err := valuation.ValueMarketHolding(&h, prices, rates); !cmp.Equal(valuation.ErrConflictingRates, err, cmpopts.EquateErrors()) {
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
		basis    valuation.Basis
		readings []valuation.IndexReading
		idx      pl.Indices // the readings ValueLot should get; nil when missing
		wantBase valuation.Basis
		missing  valuation.MissingReason
	}{
		{name: "default basis sums the lots' net values", on: "2024-09-15", idx: pl.Indices{}, wantBase: valuation.NetIfRedeemed},
		{name: "gross accrued when configured", on: "2024-09-15", basis: valuation.GrossAccrued, idx: pl.Indices{}, wantBase: valuation.GrossAccrued},
		{name: "second year with its CPI reading", on: "2025-06-01", readings: cpi, idx: pl.Indices{2: dec(t, "0.05")}, wantBase: valuation.NetIfRedeemed},
		{name: "second year without its CPI reading is a missing input", on: "2025-06-01", wantBase: valuation.NetIfRedeemed, missing: valuation.MissingIndexReading},
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
				Terms: edoTerms(t), Readings: tt.readings, Wrapper: pl.Taxable, Basis: tt.basis,
			}
			got, err := valuation.ValueBondHolding(&h, &b)
			if err != nil {
				t.Fatalf("ValueBondHolding: %v", err)
			}
			want := valuation.HoldingValuation{
				Account: h.Account, Instrument: bond, Date: h.Date, Basis: tt.wantBase, Quantity: h.Quantity,
				Source: valuation.StatutorySource, Missing: tt.missing,
			}
			if tt.idx != nil {
				want.Value = lotSum(t, tt.idx, tt.on, tt.wantBase == valuation.GrossAccrued)
				want.PLN = want.Value
			}
			if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(valuation.HoldingValuation{}, "Detail")); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

// lotSum returns the gross or net value, on the date on, of the two lots
// TestValueBondHolding values, 10 bonds at 100 PLN and 5 at 99.90 PLN, from
// pl.ValueLot itself: the test checks summing and wiring, not the bond
// formulas, which internal/jurisdiction/pl tests.
func lotSum(t *testing.T, idx pl.Indices, on string, gross bool) kernel.Money {
	t.Helper()
	sum := money(t, "0", "PLN")
	for _, lot := range []struct{ quantity, price string }{{"10", "100"}, {"5", "99.90"}} {
		v, err := pl.ValueLot(edoTerms(t), idx, kernel.NewQuantity(dec(t, lot.quantity)),
			pl.Purchase{Date: date(t, "2024-03-15"), Price: money(t, lot.price, "PLN")}, pl.Taxable, date(t, on))
		if err != nil {
			t.Fatalf("ValueLot: %v", err)
		}
		value := v.Net
		if gross {
			value = v.Gross
		}
		if sum, err = sum.Add(value); err != nil {
			t.Fatal(err)
		}
	}
	return sum
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
		basis   valuation.Basis
		wantErr error
	}{
		{name: "market basis", lot: bondLot(t, bond, "2024-03-15", "10", "1000", "b1"), basis: valuation.Market, wantErr: valuation.ErrBasisNotForBonds},
		{name: "cost not in PLN", lot: usdLot, wantErr: valuation.ErrBondLot},
		{name: "lot outside its series' sale month", lot: bondLot(t, bond, "2024-04-15", "10", "1000", "b1"), wantErr: pl.ErrSeriesMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := valuation.BondHolding{Lots: []ledger.Lot{tt.lot}, Terms: edoTerms(t), Wrapper: pl.Taxable, Basis: tt.basis}
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
