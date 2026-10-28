package ledger_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// lot builds a synthetic lot of instrument A.
func lot(t *testing.T, id, opened, units, cost string, opening bool) ledger.Lot {
	t.Helper()
	return ledger.Lot{
		Instrument: instrument(t, isinA, "TSTA"),
		Opened:     date(t, opened),
		Quantity:   qty(t, units),
		Cost:       money(t, cost, "USD"),
		Opening:    opening,
		Origin:     ledger.TransactionKey{Source: "test", ExternalID: id},
	}
}

func TestConsumeFIFO(t *testing.T) {
	t.Parallel()
	opening := lot(t, "o-1", "2019-05-17", "3", "100.00", true)
	first := lot(t, "b-1", "2026-03-02", "10", "1500.00", false)
	second := lot(t, "b-2", "2026-03-02", "5", "800.00", false) // same day, later key
	in := []ledger.Lot{second, opening, first}                  // not in FIFO order
	tests := []struct {
		name          string
		units         string
		wantConsumed  []ledger.Lot
		wantRemaining []ledger.Lot
		wantErr       error
		wantOversell  *ledger.OversellError
	}{
		{
			name:          "part of the oldest lot, cost split and rounded",
			units:         "1",
			wantConsumed:  []ledger.Lot{lot(t, "o-1", "2019-05-17", "1", "33.33", true)},
			wantRemaining: []ledger.Lot{lot(t, "o-1", "2019-05-17", "2", "66.67", true), first, second},
		},
		{
			name:          "exactly the oldest lot",
			units:         "3",
			wantConsumed:  []ledger.Lot{opening},
			wantRemaining: []ledger.Lot{first, second},
		},
		{
			name:          "across lots, same-day lots in key order",
			units:         "5",
			wantConsumed:  []ledger.Lot{opening, lot(t, "b-1", "2026-03-02", "2", "300.00", false)},
			wantRemaining: []ledger.Lot{lot(t, "b-1", "2026-03-02", "8", "1200.00", false), second},
		},
		{
			name:          "fractional units",
			units:         "3.5",
			wantConsumed:  []ledger.Lot{opening, lot(t, "b-1", "2026-03-02", "0.5", "75.00", false)},
			wantRemaining: []ledger.Lot{lot(t, "b-1", "2026-03-02", "9.5", "1425.00", false), second},
		},
		{name: "everything", units: "18", wantConsumed: []ledger.Lot{opening, first, second}},
		{
			name:         "one unit too many",
			units:        "18.001",
			wantOversell: &ledger.OversellError{Instrument: opening.Instrument, Requested: qty(t, "18.001"), Available: qty(t, "18")},
		},
		{name: "nothing", units: "0", wantErr: ledger.ErrInvalidTransaction},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.ConsumeFIFO(in, qty(t, tt.units))
			checkOversell(t, err, tt.wantErr, tt.wantOversell)
			if diff := cmp.Diff(tt.wantConsumed, got.Consumed, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("consumed (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantRemaining, got.Remaining, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("remaining (-want +got):\n%s", diff)
			}
		})
	}
}

// checkOversell compares err with wantErr, or, when wantOversell is set,
// expects that *OversellError.
func checkOversell(t *testing.T, err, wantErr error, wantOversell *ledger.OversellError) {
	t.Helper()
	if wantOversell == nil {
		if diff := cmp.Diff(wantErr, err, cmpopts.EquateErrors()); diff != "" {
			t.Fatalf("error (-want +got):\n%s", diff)
		}
		return
	}
	var got *ledger.OversellError
	if !errors.As(err, &got) {
		t.Fatalf("error = %v, want an *OversellError", err)
	}
	if diff := cmp.Diff(wantOversell, got); diff != "" {
		t.Errorf("OversellError (-want +got):\n%s", diff)
	}
	if !strings.Contains(err.Error(), wantOversell.Requested.String()) || !strings.Contains(err.Error(), wantOversell.Available.String()) {
		t.Errorf("Error() = %q, want both quantities", err)
	}
}

func TestConsumeFIFOCostSplitAddsUp(t *testing.T) {
	t.Parallel()
	// 100.00 over 3 units: 33.333... rounds to 33.33, then 66.67 / 2 = 33.335
	// rounds half-even to 33.34, and the last unit takes the 33.33 left.
	lots := []ledger.Lot{lot(t, "o-1", "2019-05-17", "3", "100.00", true)}
	costs := make([]kernel.Money, 0, 3)
	for range 3 {
		got, err := ledger.ConsumeFIFO(lots, qty(t, "1"))
		if err != nil {
			t.Fatalf("ConsumeFIFO: %v", err)
		}
		costs = append(costs, got.Consumed[0].Cost)
		lots = got.Remaining
	}
	want := []kernel.Money{money(t, "33.33", "USD"), money(t, "33.34", "USD"), money(t, "33.33", "USD")}
	if diff := cmp.Diff(want, costs); diff != "" {
		t.Errorf("consumed costs (-want +got):\n%s", diff)
	}
	if len(lots) != 0 {
		t.Errorf("lots left = %v, want none", lots)
	}
}

// lotLog is a synthetic log of instrument A: an opening lot, a buy corrected
// from 1400 to 1500 USD, a sale across both, and on one day a sale whose key
// sorts before the buy that covers it.
func lotLog(t *testing.T) []ledger.VersionedTransaction {
	t.Helper()
	a := instrument(t, isinA, "TSTA")
	none := kernel.Money{}
	v := func(tx ledger.Transaction, n int) ledger.VersionedTransaction {
		return ledger.VersionedTransaction{Transaction: tx, Version: n}
	}
	return []ledger.VersionedTransaction{
		v(must(t)(ledger.NewOpeningLot(header(t, "o-1", "2026-03-02"), a, qty(t, "3"), money(t, "100.00", "USD"), date(t, "2019-05-17"))), 1),
		v(must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-03"), a, qty(t, "10"), money(t, "1400.00", "USD"), none)), 1),
		v(must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-03"), a, qty(t, "10"), money(t, "1500.00", "USD"), none)), 2),
		v(must(t)(ledger.NewSell(header(t, "s-1", "2026-03-04"), a, qty(t, "5"), money(t, "800", "USD"), none)), 1),
		v(must(t)(ledger.NewSell(header(t, "a-sell", "2026-03-05"), a, qty(t, "10"), money(t, "1600", "USD"), none)), 1),
		v(must(t)(ledger.NewBuy(header(t, "b-2", "2026-03-05"), a, qty(t, "4"), money(t, "600.00", "USD"), none)), 1),
	}
}

func TestLotsOn(t *testing.T) {
	t.Parallel()
	log := lotLog(t)
	tests := []struct {
		name string
		on   string
		want []ledger.Lot
	}{
		{name: "before go-live", on: "2026-03-01"},
		{name: "opening lot keeps its open date and cost", on: "2026-03-02", want: []ledger.Lot{lot(t, "o-1", "2019-05-17", "3", "100.00", true)}},
		{name: "buy at its current version", on: "2026-03-03", want: []ledger.Lot{
			lot(t, "o-1", "2019-05-17", "3", "100.00", true), lot(t, "b-1", "2026-03-03", "10", "1500.00", false),
		}},
		{name: "sale takes the opening lot first", on: "2026-03-04", want: []ledger.Lot{lot(t, "b-1", "2026-03-03", "8", "1200.00", false)}},
		{name: "same-day buy applied before the sale", on: "2026-03-05", want: []ledger.Lot{lot(t, "b-2", "2026-03-05", "2", "300.00", false)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.LotsOn(log, testAccount, instrument(t, isinA, "TSTA"), date(t, tt.on))
			if err != nil {
				t.Fatalf("LotsOn: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("LotsOn (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLotsOnOversell(t *testing.T) {
	t.Parallel()
	a := instrument(t, isinA, "TSTA")
	log := []ledger.VersionedTransaction{
		{Transaction: must(t)(ledger.NewOpeningLot(header(t, "o-1", "2026-03-02"), a, qty(t, "3"), money(t, "100", "USD"), date(t, "2019-05-17"))), Version: 1},
		{Transaction: must(t)(ledger.NewSell(header(t, "s-1", "2026-03-04"), a, qty(t, "4"), money(t, "500", "USD"), kernel.Money{})), Version: 1},
	}
	_, err := ledger.LotsOn(log, testAccount, a, date(t, "2026-03-04"))
	want := &ledger.OversellError{
		Account: testAccount, Instrument: a, Date: date(t, "2026-03-04"),
		Disposal:  ledger.TransactionKey{Source: "test", ExternalID: "s-1"},
		Requested: qty(t, "4"), Available: qty(t, "3"),
	}
	checkOversell(t, err, nil, want)
}
