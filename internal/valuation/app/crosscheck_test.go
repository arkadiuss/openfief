package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// fakeCrossChecks keeps prices and cross-check prices in memory.
type fakeCrossChecks struct {
	prices []valuation.Price
	saved  []valuation.CrossCheckPrice
	err    error
}

func (f *fakeCrossChecks) SaveCrossCheckPrices(_ context.Context, prices []valuation.CrossCheckPrice) error {
	if f.err != nil {
		return f.err
	}
	f.saved = append(f.saved, prices...)
	return nil
}

func (f *fakeCrossChecks) CrossCheckedPrices(context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error) {
	return f.prices, f.saved, f.err
}

// ibkrMark is a synthetic cross-check price of the ETF from the ibkr
// instance.
func ibkrMark(t *testing.T, on, closing string) valuation.CrossCheckPrice {
	t.Helper()
	return valuation.CrossCheckPrice{
		Instrument: instrument(t, etfISIN, "TSTB").String(), Date: date(t, on),
		Close: money(t, closing, "USD"), Source: "ibkr",
	}
}

func TestNewCrossChecks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, threshold, want string
		wantErr               error
	}{
		{name: "unset takes the default", threshold: "0", want: "0.01"},
		{name: "a configured threshold", threshold: "0.005", want: "0.005"},
		{name: "a threshold of one", threshold: "1", wantErr: valuation.ErrInvalidPriceCheckThreshold},
		{name: "a negative threshold", threshold: "-0.01", wantErr: valuation.ErrInvalidPriceCheckThreshold},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := app.NewCrossChecks(&fakeCrossChecks{}, dec(t, tt.threshold))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NewCrossChecks() error (-want +got):\n%s", diff)
			}
			if err == nil && !c.Threshold().Equal(dec(t, tt.want)) {
				t.Errorf("Threshold() = %s, want %s", c.Threshold(), tt.want)
			}
		})
	}
}

func TestRecordCrossCheckPrices(t *testing.T) {
	t.Parallel()
	store := &fakeCrossChecks{}
	c, err := app.NewCrossChecks(store, kernel.Decimal{})
	if err != nil {
		t.Fatalf("NewCrossChecks: %v", err)
	}
	ok := []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-05", "99"), ibkrMark(t, "2026-03-06", "103")}
	if recErr := c.RecordCrossCheckPrices(t.Context(), ok); recErr != nil {
		t.Fatalf("RecordCrossCheckPrices: %v", recErr)
	}
	if diff := cmp.Diff(ok, store.saved); diff != "" {
		t.Errorf("stored (-want +got):\n%s", diff)
	}

	bad := ibkrMark(t, "2026-03-06", "103")
	bad.Source = ""
	err = c.RecordCrossCheckPrices(t.Context(), []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-07", "104"), bad})
	if !cmp.Equal(err, app.ErrInvalidInput, cmpopts.EquateErrors()) || !cmp.Equal(err, valuation.ErrInvalidCrossCheckPrice, cmpopts.EquateErrors()) {
		t.Errorf("RecordCrossCheckPrices(invalid) error = %v, want ErrInvalidInput and ErrInvalidCrossCheckPrice", err)
	}
	if len(store.saved) != len(ok) {
		t.Errorf("an invalid batch stored %d prices, want none of it", len(store.saved)-len(ok))
	}

	down, err := app.NewCrossChecks(&fakeCrossChecks{err: errStoreDown}, kernel.Decimal{})
	if err != nil {
		t.Fatalf("NewCrossChecks: %v", err)
	}
	if err := down.RecordCrossCheckPrices(t.Context(), ok); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("RecordCrossCheckPrices() with the store down error = %v, want %v", err, errStoreDown)
	}
}

func TestPriceChecks(t *testing.T) {
	t.Parallel()
	in := instrument(t, etfISIN, "TSTB").String()
	store := &fakeCrossChecks{
		prices: []valuation.Price{
			{Instrument: in, Date: date(t, "2026-03-05"), Close: money(t, "99", "USD"), Source: "eodhd"},
			{Instrument: in, Date: date(t, "2026-03-06"), Close: money(t, "100", "USD"), Source: "eodhd"},
		},
		saved: []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-06", "100.4"), ibkrMark(t, "2026-03-05", "99")},
	}
	c, err := app.NewCrossChecks(store, dec(t, "0.003"))
	if err != nil {
		t.Fatalf("NewCrossChecks: %v", err)
	}
	got, err := c.PriceChecks(t.Context())
	if err != nil {
		t.Fatalf("PriceChecks: %v", err)
	}
	want := []valuation.PriceCheck{
		{Price: store.prices[0], CrossCheck: store.saved[1], Difference: dec(t, "0"), Status: valuation.PriceMatched},
		{Price: store.prices[1], CrossCheck: store.saved[0], Difference: dec(t, "0.004"), Status: valuation.PriceFlagged},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PriceChecks() (-want +got):\n%s", diff)
	}

	down, err := app.NewCrossChecks(&fakeCrossChecks{err: errStoreDown}, kernel.Decimal{})
	if err != nil {
		t.Fatalf("NewCrossChecks: %v", err)
	}
	if _, err := down.PriceChecks(t.Context()); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("PriceChecks() with the store down error = %v, want %v", err, errStoreDown)
	}
}
