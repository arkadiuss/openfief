package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// queryFixture is a ledger with a synthetic history in acc-1: a deposit, two
// buys of A, a sale of 12 A, an FX conversion, and a deposit first booked to
// acc-2 and then corrected to acc-1.
type queryFixture struct {
	ledger                               *app.Ledger
	a                                    ledger.Instrument
	deposit, buy1, buy2, moved, sell, fx ledger.Transaction
}

func newQueryFixture(t *testing.T) queryFixture {
	t.Helper()
	l, _ := seeded(t)
	a := instrument(t, isinA, "TSTA")
	zero := kernel.Money{}
	f := queryFixture{
		ledger:  l,
		a:       a,
		deposit: must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "2000", "USD"))),
		buy1:    must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), a, qty(t, "10"), money(t, "500", "USD"), zero)),
		buy2:    must(t)(ledger.NewBuy(header(t, acc1, "b2", "2026-03-04"), a, qty(t, "5"), money(t, "300", "USD"), zero)),
		moved:   must(t)(ledger.NewDeposit(header(t, acc1, "m1", "2026-03-04"), money(t, "50", "USD"))),
		sell:    must(t)(ledger.NewSell(header(t, acc1, "s1", "2026-03-05"), a, qty(t, "12"), money(t, "900", "USD"), zero)),
		fx:      must(t)(ledger.NewFXConversion(header(t, acc1, "fx", "2026-03-05"), money(t, "100", "USD"), money(t, "90", "EUR"), zero)),
	}
	misbooked := must(t)(ledger.NewDeposit(header(t, acc2, "m1", "2026-03-04"), money(t, "50", "USD")))
	for _, txs := range [][]ledger.Transaction{{misbooked}, {f.deposit, f.buy1, f.buy2, f.moved, f.sell, f.fx}} {
		if _, err := l.Import(t.Context(), txs, nil); err != nil {
			t.Fatalf("Import: %v", err)
		}
	}
	return f
}

func TestCurrentTransactions(t *testing.T) {
	t.Parallel()
	f := newQueryFixture(t)
	tests := []struct {
		name    string
		account ledger.AccountID
		want    []ledger.Transaction
	}{
		{name: "in date then key order", account: acc1, want: []ledger.Transaction{f.deposit, f.buy1, f.buy2, f.moved, f.fx, f.sell}},
		{name: "a transaction corrected to another account leaves", account: acc2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := f.ledger.CurrentTransactions(t.Context(), tt.account)
			checkResult(t, "CurrentTransactions", tt.want, got, err)
		})
	}
}

func TestPositionsOnADate(t *testing.T) {
	t.Parallel()
	f := newQueryFixture(t)
	holding := func(on, units string) ledger.Holding {
		return ledger.Holding{Account: acc1, Instrument: f.a, Date: date(t, on), Quantity: qty(t, units)}
	}
	lot := func(opened, units, cost string, origin ledger.Transaction) ledger.Lot {
		return ledger.Lot{Instrument: f.a, Opened: date(t, opened), Quantity: qty(t, units), Cost: money(t, cost, "USD"), Origin: origin.Key}
	}
	tests := []struct {
		name         string
		on           string
		wantHoldings []ledger.Holding
		wantLots     []ledger.Lot
		wantCash     []kernel.Money
	}{
		{name: "before any transaction", on: "2026-03-01"},
		{
			name:         "after both buys",
			on:           "2026-03-04",
			wantHoldings: []ledger.Holding{holding("2026-03-04", "15")},
			wantLots:     []ledger.Lot{lot("2026-03-03", "10", "500", f.buy1), lot("2026-03-04", "5", "300", f.buy2)},
			wantCash:     []kernel.Money{money(t, "1250", "USD")},
		},
		{
			name:         "after the sale, FIFO",
			on:           "2026-03-05",
			wantHoldings: []ledger.Holding{holding("2026-03-05", "3")},
			wantLots:     []ledger.Lot{lot("2026-03-04", "3", "180", f.buy2)},
			wantCash:     []kernel.Money{money(t, "90", "EUR"), money(t, "2050", "USD")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			on := date(t, tt.on)
			holdings, err := f.ledger.HoldingsOn(t.Context(), acc1, on)
			checkResult(t, "HoldingsOn", tt.wantHoldings, holdings, err)
			lots, err := f.ledger.LotsOn(t.Context(), acc1, on)
			checkResult(t, "LotsOn", tt.wantLots, lots, err)
			cash, err := f.ledger.CashOn(t.Context(), acc1, on)
			checkResult(t, "CashOn", tt.wantCash, cash, err)
		})
	}
}

func TestQueriesOfAnUnknownAccount(t *testing.T) {
	t.Parallel()
	l, _ := seeded(t)
	ctx, on := t.Context(), date(t, "2026-03-05")
	const unknown ledger.AccountID = "acc-9"
	_, accountErr := l.Account(ctx, unknown)
	_, txsErr := l.CurrentTransactions(ctx, unknown)
	_, holdingsErr := l.HoldingsOn(ctx, unknown, on)
	_, lotsErr := l.LotsOn(ctx, unknown, on)
	_, cashErr := l.CashOn(ctx, unknown, on)
	_, checksErr := l.AssertionChecks(ctx, unknown)
	for name, err := range map[string]error{
		"Account": accountErr, "CurrentTransactions": txsErr, "HoldingsOn": holdingsErr,
		"LotsOn": lotsErr, "CashOn": cashErr, "AssertionChecks": checksErr,
	} {
		if diff := cmp.Diff(app.ErrUnknownAccount, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("%s error (-want +got):\n%s", name, diff)
		}
	}
}
