package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// holdingLog is a synthetic log: an opening lot of A, a buy of A corrected
// from 10 to 12 units, a sale of A, a dividend on A, a buy and full sale of
// B, and a buy of A in another account.
func holdingLog(t *testing.T) []ledger.VersionedTransaction {
	t.Helper()
	a, b := instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")
	none := kernel.Money{}
	otherAccount := header(t, "x-1", "2026-03-02")
	otherAccount.Account = "acc-2"
	v := func(tx ledger.Transaction, n int) ledger.VersionedTransaction {
		return ledger.VersionedTransaction{Transaction: tx, Version: n}
	}
	return []ledger.VersionedTransaction{
		v(must(t)(ledger.NewOpeningLot(header(t, "o-1", "2026-03-01"), a, qty(t, "3"), money(t, "100", "USD"), date(t, "2019-05-17"))), 1),
		v(must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-02"), a, qty(t, "10"), money(t, "1000", "USD"), none)), 1),
		v(must(t)(ledger.NewBuy(header(t, "b-1", "2026-03-02"), a, qty(t, "12"), money(t, "1200", "USD"), none)), 2),
		v(must(t)(ledger.NewBuy(header(t, "b-2", "2026-03-03"), b, qty(t, "5"), money(t, "50", "EUR"), none)), 1),
		v(must(t)(ledger.NewSell(header(t, "s-1", "2026-03-05"), a, qty(t, "4"), money(t, "480", "USD"), none)), 1),
		v(must(t)(ledger.NewDividend(header(t, "d-1", "2026-03-06"), a, money(t, "7", "USD"), none)), 1),
		v(must(t)(ledger.NewSell(header(t, "s-2", "2026-03-07"), b, qty(t, "5"), money(t, "55", "EUR"), none)), 1),
		v(must(t)(ledger.NewBuy(otherAccount, a, qty(t, "100"), money(t, "1", "USD"), none)), 1),
	}
}

func TestHoldingOn(t *testing.T) {
	t.Parallel()
	log := holdingLog(t)
	tests := []struct {
		name   string
		isin   string
		symbol string
		on     string
		want   string
	}{
		{name: "before anything", isin: isinA, symbol: "TSTA", on: "2026-02-28", want: "0"},
		{name: "opening lot on go-live", isin: isinA, symbol: "TSTA", on: "2026-03-01", want: "3"},
		{name: "corrected buy counts once, at its current version", isin: isinA, symbol: "TSTA", on: "2026-03-02", want: "15"},
		{name: "after a sale", isin: isinA, symbol: "TSTA", on: "2026-03-05", want: "11"},
		{name: "a dividend moves no units", isin: isinA, symbol: "TSTA", on: "2026-03-06", want: "11"},
		{name: "other instrument not yet bought", isin: isinB, symbol: "TSTB", on: "2026-03-02", want: "0"},
		{name: "other instrument", isin: isinB, symbol: "TSTB", on: "2026-03-03", want: "5"},
		{name: "other instrument sold out", isin: isinB, symbol: "TSTB", on: "2026-03-07", want: "0"},
		{name: "same ISIN on another venue", isin: isinA, symbol: "TSTA.XX", on: "2026-03-07", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := instrument(t, tt.isin, tt.symbol)
			got, err := ledger.HoldingOn(log, testAccount, in, date(t, tt.on))
			if err != nil {
				t.Fatalf("HoldingOn: %v", err)
			}
			want := ledger.Holding{Account: testAccount, Instrument: in, Date: date(t, tt.on), Quantity: qty(t, tt.want)}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("HoldingOn (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHoldingsOn(t *testing.T) {
	t.Parallel()
	log := holdingLog(t)
	a, b := instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")
	h := func(in ledger.Instrument, on, units string) ledger.Holding {
		return ledger.Holding{Account: testAccount, Instrument: in, Date: date(t, on), Quantity: qty(t, units)}
	}
	tests := []struct {
		name string
		on   string
		want []ledger.Holding
	}{
		{name: "nothing yet", on: "2026-02-28", want: nil},
		{name: "two instruments in ISIN order", on: "2026-03-05", want: []ledger.Holding{h(b, "2026-03-05", "5"), h(a, "2026-03-05", "11")}},
		{name: "sold-out holding left out", on: "2026-03-07", want: []ledger.Holding{h(a, "2026-03-07", "11")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.HoldingsOn(log, testAccount, date(t, tt.on))
			if err != nil {
				t.Fatalf("HoldingsOn: %v", err)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("HoldingsOn (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHoldingsOnInvalidLog(t *testing.T) {
	t.Parallel()
	log := holdingLog(t)
	log[1].Version = 2 // now two versions numbered 2 for b-1
	in := instrument(t, isinA, "TSTA")
	on := date(t, "2026-03-07")
	_, errOne := ledger.HoldingOn(log, testAccount, in, on)
	_, errAll := ledger.HoldingsOn(log, testAccount, on)
	_, errLots := ledger.LotsOn(log, testAccount, in, on)
	for _, err := range []error{errOne, errAll, errLots} {
		if diff := cmp.Diff(ledger.ErrInvalidLog, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("error (-want +got):\n%s", diff)
		}
	}
}
