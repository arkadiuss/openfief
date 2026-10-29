package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Synthetic instruments, written as ledger.Instrument.String() writes them.
const (
	etf   = "PLTEST000010/TSTA"
	stock = "IETEST000026/TSTB"
)

// price builds a synthetic market-data price from the eodhd instance.
func price(t *testing.T, instrument, on, closing, code string) valuation.Price {
	t.Helper()
	return valuation.Price{Instrument: instrument, Date: date(t, on), Close: money(t, closing, code), Source: "eodhd"}
}

// crossCheck builds a synthetic cross-check price from source.
func crossCheck(t *testing.T, instrument, on, closing, code, source string) valuation.CrossCheckPrice {
	t.Helper()
	return valuation.CrossCheckPrice{Instrument: instrument, Date: date(t, on), Close: money(t, closing, code), Source: source}
}

// orDefault returns s, or def when s is empty.
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func TestCrossCheckPriceValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, instrument, date, closing, source string
		wantErr                                 error
	}{
		{name: "an IBKR mark price", instrument: etf, date: "2026-03-06", closing: "101.25", source: "ibkr"},
		{name: "no instrument", date: "2026-03-06", closing: "1", source: "ibkr", wantErr: valuation.ErrInvalidCrossCheckPrice},
		{name: "no date", instrument: etf, closing: "1", source: "ibkr", wantErr: valuation.ErrInvalidCrossCheckPrice},
		{name: "a zero close", instrument: etf, date: "2026-03-06", closing: "0", source: "ibkr", wantErr: valuation.ErrInvalidCrossCheckPrice},
		{name: "a negative close", instrument: etf, date: "2026-03-06", closing: "-1", source: "ibkr", wantErr: valuation.ErrInvalidCrossCheckPrice},
		{name: "no source", instrument: etf, date: "2026-03-06", closing: "1", wantErr: valuation.ErrInvalidCrossCheckPrice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := valuation.CrossCheckPrice{Instrument: tt.instrument, Close: money(t, tt.closing, "USD"), Source: tt.source}
			if tt.date != "" {
				c.Date = date(t, tt.date)
			}
			if diff := cmp.Diff(tt.wantErr, c.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckThreshold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, threshold string
		wantErr         error
	}{
		{name: "the default, one percent", threshold: "0.01"},
		{name: "a tenth of a percent", threshold: "0.001"},
		{name: "just below one", threshold: "0.999999"},
		{name: "zero", threshold: "0", wantErr: valuation.ErrInvalidPriceCheckThreshold},
		{name: "negative", threshold: "-0.01", wantErr: valuation.ErrInvalidPriceCheckThreshold},
		{name: "one", threshold: "1", wantErr: valuation.ErrInvalidPriceCheckThreshold},
		{name: "a percent written as a whole number", threshold: "5", wantErr: valuation.ErrInvalidPriceCheckThreshold},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := valuation.CheckThreshold(dec(t, tt.threshold))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CheckThreshold(%s) error (-want +got):\n%s", tt.threshold, diff)
			}
			if err == nil && !got.Equal(dec(t, tt.threshold)) {
				t.Errorf("CheckThreshold(%s) = %s", tt.threshold, got)
			}
		})
	}
	if got := valuation.DefaultPriceCheckThreshold(); !got.Equal(dec(t, "0.01")) {
		t.Errorf("DefaultPriceCheckThreshold() = %s, want 0.01", got)
	}
}

func TestPriceCheckStatusNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status valuation.PriceCheckStatus
		want   string
	}{
		{valuation.PriceMatched, "matched"},
		{valuation.PriceFlagged, "flagged"},
		{0, "PriceCheckStatus(0)"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", int(tt.status), got, tt.want)
		}
	}
}

func TestCheckPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		price, cross, threshold    string
		crossCode                  string
		crossInstrument, crossDate string
		wantDiff                   string
		wantStatus                 valuation.PriceCheckStatus
		wantErr                    error
	}{
		{name: "equal prices", price: "100", cross: "100", threshold: "0.01", wantDiff: "0", wantStatus: valuation.PriceMatched},
		{name: "an aggregated close a cent off", price: "100.00", cross: "100.01", threshold: "0.01", wantDiff: "0.0001", wantStatus: valuation.PriceMatched},
		{name: "exactly at the threshold is matched", price: "100", cross: "101", threshold: "0.01", wantDiff: "0.01", wantStatus: valuation.PriceMatched},
		{name: "just beyond the threshold is flagged", price: "100", cross: "101.0001", threshold: "0.01", wantDiff: "0.010001", wantStatus: valuation.PriceFlagged},
		{name: "beyond the threshold downward is flagged", price: "100", cross: "98.5", threshold: "0.01", wantDiff: "-0.015", wantStatus: valuation.PriceFlagged},
		{name: "pence read as pounds", price: "84.125", cross: "8412.5", threshold: "0.01", wantDiff: "99", wantStatus: valuation.PriceFlagged},
		{name: "a stricter threshold flags a small gap", price: "100", cross: "100.2", threshold: "0.001", wantDiff: "0.002", wantStatus: valuation.PriceFlagged},
		{
			name: "a difference that rounds to zero is still flagged exactly", price: "3", cross: "3.0000006",
			threshold: "0.0000001", wantDiff: "0", wantStatus: valuation.PriceFlagged,
		},
		{name: "the difference is rounded half even", price: "3", cross: "4", threshold: "0.5", wantDiff: "0.333333", wantStatus: valuation.PriceMatched},
		{name: "another currency", price: "100", cross: "100", crossCode: "EUR", threshold: "0.01", wantErr: valuation.ErrNotComparable},
		{name: "another instrument", price: "100", cross: "100", crossInstrument: stock, threshold: "0.01", wantErr: valuation.ErrNotComparable},
		{name: "another date", price: "100", cross: "100", crossDate: "2026-03-05", threshold: "0.01", wantErr: valuation.ErrNotComparable},
		{name: "a zero price", price: "0", cross: "100", threshold: "0.01", wantErr: valuation.ErrNotComparable},
		{name: "a zero threshold", price: "100", cross: "100", threshold: "0", wantErr: valuation.ErrInvalidPriceCheckThreshold},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := price(t, etf, "2026-03-06", tt.price, "USD")
			c := crossCheck(t, orDefault(tt.crossInstrument, etf), orDefault(tt.crossDate, "2026-03-06"), tt.cross, orDefault(tt.crossCode, "USD"), "ibkr")
			got, err := valuation.CheckPrice(&p, &c, dec(t, tt.threshold))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CheckPrice() error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			want := valuation.PriceCheck{Price: p, CrossCheck: c, Difference: dec(t, tt.wantDiff), Status: tt.wantStatus}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("CheckPrice() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckPrices(t *testing.T) {
	t.Parallel()
	prices := []valuation.Price{
		price(t, stock, "2026-03-06", "50", "EUR"),
		price(t, etf, "2026-03-06", "100", "USD"),
		price(t, etf, "2026-03-05", "99", "USD"),
		price(t, etf, "2026-03-04", "98", "USD"),
	}
	crossChecks := []valuation.CrossCheckPrice{
		crossCheck(t, etf, "2026-03-06", "103", "USD", "ibkr-b"),
		crossCheck(t, etf, "2026-03-06", "100.5", "USD", "ibkr-a"),
		crossCheck(t, stock, "2026-03-06", "50", "EUR", "ibkr-a"),
		// Another currency than the price's: not compared.
		crossCheck(t, etf, "2026-03-05", "99", "EUR", "ibkr-a"),
		// No price on that date: nothing to compare with.
		crossCheck(t, etf, "2026-03-03", "97", "USD", "ibkr-a"),
		crossCheck(t, etf, "2026-03-04", "98", "USD", "ibkr-a"),
	}
	got, err := valuation.CheckPrices(prices, crossChecks, valuation.DefaultPriceCheckThreshold())
	if err != nil {
		t.Fatalf("CheckPrices() error: %v", err)
	}
	want := []valuation.PriceCheck{
		{Price: prices[3], CrossCheck: crossChecks[5], Difference: dec(t, "0"), Status: valuation.PriceMatched},
		{Price: prices[0], CrossCheck: crossChecks[2], Difference: dec(t, "0"), Status: valuation.PriceMatched},
		{Price: prices[1], CrossCheck: crossChecks[1], Difference: dec(t, "0.005"), Status: valuation.PriceMatched},
		{Price: prices[1], CrossCheck: crossChecks[0], Difference: dec(t, "0.03"), Status: valuation.PriceFlagged},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CheckPrices() (-want +got):\n%s", diff)
	}

	none, err := valuation.CheckPrices(nil, crossChecks, valuation.DefaultPriceCheckThreshold())
	if err != nil || len(none) != 0 {
		t.Errorf("CheckPrices() without prices = %v, %v; want none", none, err)
	}
	if _, err := valuation.CheckPrices(prices, crossChecks, kernel.Decimal{}); !cmp.Equal(err, valuation.ErrInvalidPriceCheckThreshold, cmpopts.EquateErrors()) {
		t.Errorf("CheckPrices() with a zero threshold error = %v, want %v", err, valuation.ErrInvalidPriceCheckThreshold)
	}
}
