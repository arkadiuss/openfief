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

// fakeCrossChecks keeps prices, cross-check prices, thresholds and
// acknowledgements in memory.
type fakeCrossChecks struct {
	prices     []valuation.Price
	saved      []valuation.CrossCheckPrice
	thresholds map[string]kernel.Decimal
	acks       []valuation.PriceCheckAcknowledgement
	err        error
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

func (f *fakeCrossChecks) PriceCheckThresholds(context.Context) (valuation.Thresholds, error) {
	return valuation.Thresholds{ByInstrument: f.thresholds}, f.err
}

func (f *fakeCrossChecks) SavePriceCheckThreshold(_ context.Context, instrument string, threshold *kernel.Decimal) error {
	if f.err != nil {
		return f.err
	}
	if f.thresholds == nil {
		f.thresholds = map[string]kernel.Decimal{}
	}
	if threshold == nil {
		delete(f.thresholds, instrument)
		return nil
	}
	f.thresholds[instrument] = *threshold
	return nil
}

func (f *fakeCrossChecks) PriceCheckAcknowledgements(context.Context) ([]valuation.PriceCheckAcknowledgement, error) {
	return f.acks, f.err
}

func (f *fakeCrossChecks) AppendPriceCheckAcknowledgement(_ context.Context, ack *valuation.PriceCheckAcknowledgement) error {
	if f.err != nil {
		return f.err
	}
	f.acks = append(f.acks, *ack)
	return nil
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

// ownerCrossChecks returns the cross-check use cases over store, with
// ownerToken the owner's session.
func ownerCrossChecks(store app.CrossCheckStore) *app.CrossChecks {
	return app.NewCrossChecks(store, fakeOwners{err: app.ErrNotOwner})
}

func TestCrossChecksThreshold(t *testing.T) {
	t.Parallel()
	if got := app.NewCrossChecks(&fakeCrossChecks{}, nil).Threshold(); !got.Equal(dec(t, "0.01")) {
		t.Errorf("Threshold() = %s, want 0.01 (D-110)", got)
	}
}

func TestRecordCrossCheckPrices(t *testing.T) {
	t.Parallel()
	store := &fakeCrossChecks{}
	c := app.NewCrossChecks(store, nil)
	ok := []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-05", "99"), ibkrMark(t, "2026-03-06", "103")}
	if recErr := c.RecordCrossCheckPrices(t.Context(), ok); recErr != nil {
		t.Fatalf("RecordCrossCheckPrices: %v", recErr)
	}
	if diff := cmp.Diff(ok, store.saved); diff != "" {
		t.Errorf("stored (-want +got):\n%s", diff)
	}

	bad := ibkrMark(t, "2026-03-06", "103")
	bad.Source = ""
	err := c.RecordCrossCheckPrices(t.Context(), []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-07", "104"), bad})
	if !cmp.Equal(err, app.ErrInvalidInput, cmpopts.EquateErrors()) || !cmp.Equal(err, valuation.ErrInvalidCrossCheckPrice, cmpopts.EquateErrors()) {
		t.Errorf("RecordCrossCheckPrices(invalid) error = %v, want ErrInvalidInput and ErrInvalidCrossCheckPrice", err)
	}
	if len(store.saved) != len(ok) {
		t.Errorf("an invalid batch stored %d prices, want none of it", len(store.saved)-len(ok))
	}

	down := app.NewCrossChecks(&fakeCrossChecks{err: errStoreDown}, nil)
	if err := down.RecordCrossCheckPrices(t.Context(), ok); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("RecordCrossCheckPrices() with the store down error = %v, want %v", err, errStoreDown)
	}
}

// checkedStore holds two days of the ETF's prices and two IBKR marks: one
// matched, one 3% above Friday's close on the Sunday after it.
func checkedStore(t *testing.T) *fakeCrossChecks {
	t.Helper()
	in := instrument(t, etfISIN, "TSTB").String()
	return &fakeCrossChecks{
		prices: []valuation.Price{
			{Instrument: in, Date: date(t, "2026-03-05"), Close: money(t, "99", "USD"), Source: "eodhd"},
			{Instrument: in, Date: date(t, "2026-03-06"), Close: money(t, "100", "USD"), Source: "eodhd"},
		},
		saved: []valuation.CrossCheckPrice{ibkrMark(t, "2026-03-08", "103"), ibkrMark(t, "2026-03-05", "99")},
	}
}

func TestPriceChecks(t *testing.T) {
	t.Parallel()
	store := checkedStore(t)
	got, err := app.NewCrossChecks(store, nil).PriceChecks(t.Context())
	if err != nil {
		t.Fatalf("PriceChecks: %v", err)
	}
	one := dec(t, "0.01")
	want := []valuation.PriceCheck{
		{Price: store.prices[0], CrossCheck: store.saved[1], Difference: dec(t, "0"), Threshold: one, Status: valuation.PriceMatched},
		{Price: store.prices[1], CrossCheck: store.saved[0], Difference: dec(t, "0.03"), Threshold: one, Status: valuation.PriceFlagged},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("PriceChecks() (-want +got):\n%s", diff)
	}

	// The owner's threshold of the ETF keeps the 3% gap matched (D-110).
	store.thresholds = map[string]kernel.Decimal{store.prices[0].Instrument: dec(t, "0.05")}
	loose, err := app.NewCrossChecks(store, nil).PriceChecks(t.Context())
	if err != nil || len(loose) != 2 || loose[1].Status != valuation.PriceMatched {
		t.Errorf("PriceChecks() at the owner's 5%% = %v, %v; want the Sunday check matched", loose, err)
	}

	down := app.NewCrossChecks(&fakeCrossChecks{err: errStoreDown}, nil)
	if _, err := down.PriceChecks(t.Context()); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("PriceChecks() with the store down error = %v, want %v", err, errStoreDown)
	}
}

func TestSetThreshold(t *testing.T) {
	t.Parallel()
	in := instrument(t, etfISIN, "TSTB").String()
	five, one := dec(t, "0.05"), dec(t, "1")
	tests := []struct {
		name, token, instrument string
		threshold               *kernel.Decimal
		storeErr                error
		want                    map[string]kernel.Decimal
		wantErr                 error
	}{
		{name: "the owner sets one instrument's", token: ownerToken, instrument: in, threshold: &five, want: map[string]kernel.Decimal{in: five}},
		{name: "the owner clears it", token: ownerToken, instrument: in, want: map[string]kernel.Decimal{}},
		{name: "not the owner", token: "ofsess_read", instrument: in, threshold: &five, wantErr: app.ErrNotOwner},
		{name: "no instrument", token: ownerToken, threshold: &five, wantErr: app.ErrInvalidInput},
		{name: "a threshold of one", token: ownerToken, instrument: in, threshold: &one, wantErr: valuation.ErrInvalidPriceCheckThreshold},
		{name: "the store fails", token: ownerToken, instrument: in, threshold: &five, storeErr: errStoreDown, wantErr: errStoreDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeCrossChecks{thresholds: map[string]kernel.Decimal{in: dec(t, "0.02")}, err: tt.storeErr}
			err := ownerCrossChecks(store).SetThreshold(t.Context(), "https://openfief.example", tt.token, tt.instrument, tt.threshold)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SetThreshold() error (-want +got):\n%s", diff)
			}
			if err == nil {
				if diff := cmp.Diff(tt.want, store.thresholds); diff != "" {
					t.Errorf("thresholds (-want +got):\n%s", diff)
				}
			}
		})
	}
	if err := app.NewCrossChecks(&fakeCrossChecks{}, nil).SetThreshold(t.Context(), "", ownerToken, in, &five); !cmp.Equal(err, app.ErrNotOwner, cmpopts.EquateErrors()) {
		t.Errorf("SetThreshold() with no owners = %v, want ErrNotOwner", err)
	}
}

func TestAcknowledgePriceCheck(t *testing.T) {
	t.Parallel()
	in := instrument(t, etfISIN, "TSTB").String()
	tests := []struct {
		name, token, day string
		wantErr          error
	}{
		{name: "the owner acknowledges a flagged check", token: ownerToken, day: "2026-03-08"},
		{name: "a matched check", token: ownerToken, day: "2026-03-05", wantErr: valuation.ErrNotFlagged},
		{name: "no such check", token: ownerToken, day: "2026-03-09", wantErr: app.ErrUnknownPriceCheck},
		{name: "not the owner", token: "ofsess_read", day: "2026-03-08", wantErr: app.ErrNotOwner},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := checkedStore(t)
			c := ownerCrossChecks(store)
			got, err := c.Acknowledge(t.Context(), "https://openfief.example", tt.token, in, date(t, tt.day), "ibkr")
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Acknowledge() error (-want +got):\n%s", diff)
			}
			if err != nil {
				if len(store.acks) != 0 {
					t.Errorf("a refused acknowledgement recorded %v", store.acks)
				}
				return
			}
			checkAcknowledged(t, c, store, &got)
		})
	}
}

// checkAcknowledged checks that got, the check Acknowledge returned, is
// acknowledged and recorded once, that only it lists as acknowledged, and
// that a new mark of its day is flagged anew.
func checkAcknowledged(t *testing.T, c *app.CrossChecks, store *fakeCrossChecks, got *valuation.PriceCheck) {
	t.Helper()
	if !got.Acknowledged || len(store.acks) != 1 {
		t.Fatalf("Acknowledge() = %+v with %d recorded, want it acknowledged and recorded once", got, len(store.acks))
	}
	checks, err := c.PriceChecks(t.Context())
	if err != nil || !checks[1].Acknowledged || checks[0].Acknowledged {
		t.Errorf("PriceChecks() after acknowledging = %v, %v; want only the flagged one acknowledged", checks, err)
	}
	store.saved[0].Close = money(t, "104", "USD")
	again, err := c.PriceChecks(t.Context())
	if err != nil || again[1].Acknowledged {
		t.Errorf("PriceChecks() after a new mark = %v, %v; want it no longer acknowledged", again, err)
	}
}
