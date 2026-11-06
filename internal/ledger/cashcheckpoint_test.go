package ledger_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestCashDifference books the difference to a cash checkpoint on the
// synthetic log, whose derived USD cash on 2026-03-06 is 1287 (D-158): more
// cash shown is a deposit, less a withdrawal, each dated on the checkpoint
// and keyed by it; once booked, the checkpoint matches, and the derived cash
// changes only by the booked transaction. Booked as cash income (D-152),
// more cash shown is income and less is a fee. A matching checkpoint, an
// assertion a connector reported and a holding are refused.
func TestCashDifference(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		shown    string
		reported bool
		as       ledger.DifferenceBooking
		wantKind ledger.TransactionKind
		wantCash string
		wantID   string
		wantErr  error
	}{
		{name: "more cash shown is a deposit", shown: "1300", wantKind: ledger.KindDeposit, wantCash: "13", wantID: "2026-03-06/USD/1300/1287"},
		{name: "less cash shown is a withdrawal", shown: "1280.50", wantKind: ledger.KindWithdrawal, wantCash: "-6.5", wantID: "2026-03-06/USD/1280.5/1287"},
		{
			name: "more cash shown booked as cash income is income", shown: "1300", as: ledger.AsCashIncome,
			wantKind: ledger.KindCashIncome, wantCash: "13", wantID: "2026-03-06/USD/1300/1287",
		},
		{
			name: "less cash shown booked as cash income is a fee", shown: "1280.50", as: ledger.AsCashIncome,
			wantKind: ledger.KindCashIncome, wantCash: "-6.5", wantID: "2026-03-06/USD/1280.5/1287",
		},
		{name: "an unknown way to book it", shown: "1300", as: ledger.AsCashIncome + 1, wantErr: ledger.ErrUnknownBooking},
		{name: "a matching checkpoint has nothing to book", shown: "1287.00", wantErr: ledger.ErrNoCashDifference},
		{name: "a reported assertion is no checkpoint", shown: "1300", reported: true, wantErr: ledger.ErrNotCheckpoint},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := holdingLog(t)
			on := date(t, "2026-03-06")
			a := ledger.CashCheckpoint(testAccount, on, money(t, tt.shown, "USD"))
			a.Entered = !tt.reported
			derived, err := ledger.CashOn(log, testAccount, cur(t, "USD"), on)
			if err != nil {
				t.Fatalf("CashOn: %v", err)
			}
			got, err := ledger.CashDifference(&a, derived.Amount(), tt.as)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CashDifference = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			checkBooked(t, log, &a, &got, tt.wantKind, tt.wantCash, tt.wantID)
		})
	}
	holding := ledger.HoldingAssertion(testAccount, date(t, "2026-03-06"), instrument(t, isinA, "AAA"), qty(t, "1"))
	holding.Entered = true
	if _, err := ledger.CashDifference(&holding, dec(t, "0"), ledger.AsDepositOrWithdrawal); !errors.Is(err, ledger.ErrInvalidAssertion) {
		t.Errorf("CashDifference of a holding = %v, want ErrInvalidAssertion", err)
	}
}

// checkBooked checks the booked difference got against the checkpoint a:
// its kind, key, date and cash posting, and that appending it to log makes
// the checkpoint match while the earlier cash stays as derived.
func checkBooked(t *testing.T, log []ledger.VersionedTransaction, a *ledger.BalanceAssertion, got *ledger.Transaction,
	kind ledger.TransactionKind, cash, id string,
) {
	t.Helper()
	wantKey := ledger.TransactionKey{Source: ledger.CashCheckpointSource, ExternalID: id}
	if got.Kind != kind || got.Key != wantKey || !got.Date.Equal(a.Date) || got.Account != a.Account {
		t.Errorf("booked %s %s on %s in %s, want %s %s on %s", got.Kind, got.Key, got.Date, got.Account, kind, wantKey, a.Date)
	}
	if diff := cmp.Diff([]kernel.Money{money(t, cash, "USD")}, cashPostings(got)); diff != "" {
		t.Errorf("cash posting (-want +got):\n%s", diff)
	}
	booked := append(append([]ledger.VersionedTransaction(nil), log...), ledger.VersionedTransaction{Transaction: *got, Version: 1})
	if err := ledger.CheckBalanceAssertion(booked, a); err != nil {
		t.Errorf("the checkpoint after booking: %v", err)
	}
	day := date(t, "2026-03-05")
	want, err := ledger.CashOn(log, a.Account, a.Book.Currency, day)
	if err != nil {
		t.Fatalf("CashOn: %v", err)
	}
	if before, err := ledger.CashOn(booked, a.Account, a.Book.Currency, day); err != nil || !before.Equal(want) {
		t.Errorf("cash the day before = %v, %v; want %s unchanged", before, err, want)
	}
}

// cashPostings returns the amounts t posts to cash.
func cashPostings(t *ledger.Transaction) []kernel.Money {
	var out []kernel.Money
	for _, p := range t.Postings {
		if p.Book.Kind == ledger.BookCash {
			out = append(out, p.Amount)
		}
	}
	return out
}

func TestCheckpointItem(t *testing.T) {
	t.Parallel()
	a := ledger.CashCheckpoint("acc/1", date(t, "2026-03-31"), money(t, "105.00", "PLN"))
	subject, detail := ledger.CheckpointSubject(&a), ledger.CheckpointDetail(&a, dec(t, "100"))
	if subject != "acc/1/2026-03-31/PLN" {
		t.Errorf("subject = %q", subject)
	}
	if want := "shown 105 PLN, derived 100 PLN, difference 5 PLN"; detail != want {
		t.Errorf("detail = %q, want %q", detail, want)
	}
}

func TestCashCheckpointIsEnteredCash(t *testing.T) {
	t.Parallel()
	a := ledger.CashCheckpoint(testAccount, date(t, "2026-03-31"), money(t, "10", "PLN"))
	if !a.Entered || a.Book != ledger.CashBook(cur(t, "PLN")) || a.Validate() != nil {
		t.Errorf("CashCheckpoint = %+v, want a valid entered PLN cash assertion", a)
	}
}
