package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// errAttentionDown is the fake attention items' failure.
var errAttentionDown = errors.New("attention down")

// attentionCall is one call a fakeAttention got: "raise kind subject:
// detail" or "clear kind subject".
type attentionCall string

// fakeAttention records the attention items the ledger raises and clears,
// or fails with err.
type fakeAttention struct {
	calls []attentionCall
	err   error
}

func (f *fakeAttention) Raise(_ context.Context, kind, subject, detail string) error {
	f.calls = append(f.calls, attentionCall("raise "+kind+" "+subject+": "+detail))
	return f.err
}

func (f *fakeAttention) Clear(_ context.Context, kind, subject string) error {
	f.calls = append(f.calls, attentionCall("clear "+kind+" "+subject))
	return f.err
}

// TestCashCheckpoints walks D-158 on the ledger: the owner enters the cash
// an institution showed, which is stored as a balance assertion marked
// entered and checked against the derived cash, which it never changes; a
// mismatch raises the checkpoint's attention item; booking the difference
// appends a deposit dated on the checkpoint, after which the checkpoint
// matches and the item clears; a second booking finds nothing to book.
func TestCashCheckpoints(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	l, _ := seeded(t)
	attn := &fakeAttention{}
	l.WithAttention(attn)
	if _, err := l.Import(ctx, []ledger.Transaction{must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1000", "PLN")))}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	on := date(t, "2026-03-31")
	check, err := l.RecordCashCheckpoint(ctx, acc1, on, money(t, "1012.50", "PLN"))
	if err != nil {
		t.Fatalf("RecordCashCheckpoint: %v", err)
	}
	want := app.AssertionCheck{Assertion: ledger.CashCheckpoint(acc1, on, money(t, "1012.50", "PLN")), Status: app.AssertionFlagged, Derived: dec(t, "1000")}
	if diff := cmp.Diff(want, check); diff != "" {
		t.Errorf("checkpoint (-want +got):\n%s", diff)
	}
	checkCash(t, l, "1000")
	checkBooking(t, l)
	wantCalls := []attentionCall{
		"raise cash-checkpoint-mismatch acc-1/2026-03-31/PLN: shown 1012.5 PLN, derived 1000 PLN, difference 12.5 PLN",
		"clear cash-checkpoint-mismatch acc-1/2026-03-31/PLN",
	}
	if diff := cmp.Diff(wantCalls, attn.calls); diff != "" {
		t.Errorf("attention calls (-want +got):\n%s", diff)
	}
}

// checkBooking books the difference of acc1's PLN checkpoint of
// 2026-03-31, 12.50 PLN over the derived cash: a deposit on that date,
// after which the checkpoint, still marked entered, matches and the derived
// cash is the cash shown; booking again finds nothing to book.
func checkBooking(t *testing.T, l *app.Ledger) {
	t.Helper()
	ctx, on, pln := t.Context(), date(t, "2026-03-31"), money(t, "0", "PLN").Currency()
	booked, err := l.BookCashDifference(ctx, acc1, on, pln, ledger.AsDepositOrWithdrawal)
	if err != nil {
		t.Fatalf("BookCashDifference: %v", err)
	}
	if booked.Kind != ledger.KindDeposit || !booked.Postings[0].Amount.Equal(money(t, "12.50", "PLN")) || !booked.Date.Equal(on) {
		t.Errorf("booked %s of %s on %s, want a deposit of 12.50 PLN on %s", booked.Kind, booked.Postings[0].Amount, booked.Date, on)
	}
	checkCash(t, l, "1012.50")
	stored, err := l.AssertionChecks(ctx, acc1)
	if err != nil || len(stored) != 1 || stored[0].Flagged() || !stored[0].Assertion.Entered {
		t.Errorf("stored checks = %+v, %v; want the entered checkpoint, matched", stored, err)
	}
	if _, err := l.BookCashDifference(ctx, acc1, on, pln, ledger.AsDepositOrWithdrawal); !errors.Is(err, ledger.ErrNoCashDifference) {
		t.Errorf("booking again = %v, want ErrNoCashDifference", err)
	}
}

// TestCashDifferenceAsCashIncome checks that the owner can book a cash
// checkpoint's difference as cash income (D-158, D-152): less cash shown
// than derived is booked as negative cash income, a fee, on the
// checkpoint's date, after which the checkpoint matches.
func TestCashDifferenceAsCashIncome(t *testing.T) {
	t.Parallel()
	ctx, on := t.Context(), date(t, "2026-03-31")
	l, _ := seeded(t)
	if _, err := l.Import(ctx, []ledger.Transaction{must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1000", "PLN")))}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if _, err := l.RecordCashCheckpoint(ctx, acc1, on, money(t, "990", "PLN")); err != nil {
		t.Fatalf("RecordCashCheckpoint: %v", err)
	}
	booked, err := l.BookCashDifference(ctx, acc1, on, money(t, "0", "PLN").Currency(), ledger.AsCashIncome)
	if err != nil {
		t.Fatalf("BookCashDifference: %v", err)
	}
	if booked.Kind != ledger.KindCashIncome || !booked.Date.Equal(on) || !booked.Postings[1].Amount.Equal(money(t, "-10", "PLN")) {
		t.Errorf("booked %s on %s, %v; want cash income of -10 PLN on %s", booked.Kind, booked.Date, booked.Postings, on)
	}
	checkCash(t, l, "990")
	stored, err := l.AssertionChecks(ctx, acc1)
	if err != nil || len(stored) != 1 || stored[0].Flagged() {
		t.Errorf("stored checks = %+v, %v; want the checkpoint, matched", stored, err)
	}
}

// checkCash fails t unless acc1's derived PLN cash at the end of
// 2026-03-31 is want.
func checkCash(t *testing.T, l *app.Ledger, want string) {
	t.Helper()
	got, err := l.CashOn(t.Context(), acc1, date(t, "2026-03-31"))
	if err != nil {
		t.Fatalf("CashOn: %v", err)
	}
	if len(got) != 1 || !got[0].Equal(money(t, want, "PLN")) {
		t.Errorf("derived cash = %v, want %s PLN", got, want)
	}
}

func TestCashCheckpointRefusals(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	l, _ := seeded(t)
	on := date(t, "2026-03-31")
	pln := money(t, "1", "PLN")
	if _, err := l.RecordCashCheckpoint(ctx, "acc-9", on, pln); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("a checkpoint of an unknown account = %v, want ErrUnknownAccount", err)
	}
	if _, err := l.BookCashDifference(ctx, acc1, on, pln.Currency(), ledger.AsDepositOrWithdrawal); !errors.Is(err, app.ErrUnknownCheckpoint) {
		t.Errorf("booking without a checkpoint = %v, want ErrUnknownCheckpoint", err)
	}
	if _, err := l.Import(ctx, nil, []ledger.BalanceAssertion{ledger.CashAssertion(acc1, on, pln)}); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if _, err := l.BookCashDifference(ctx, acc1, on, pln.Currency(), ledger.AsDepositOrWithdrawal); !errors.Is(err, app.ErrUnknownCheckpoint) {
		t.Errorf("booking from a reported assertion = %v, want ErrUnknownCheckpoint", err)
	}
	if _, err := l.BookCashDifference(ctx, "acc-9", on, pln.Currency(), ledger.AsDepositOrWithdrawal); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("booking in an unknown account = %v, want ErrUnknownAccount", err)
	}
	l.WithAttention(&fakeAttention{err: errAttentionDown})
	if _, err := l.RecordCashCheckpoint(ctx, acc1, on, money(t, "2", "PLN")); !errors.Is(err, errAttentionDown) {
		t.Errorf("a checkpoint whose item can't be raised = %v, want the attention error", err)
	}
}

// fakeBook answers the owner's checkpoint calls, or fails with err.
type fakeBook struct{ err error }

func (b fakeBook) RecordCashCheckpoint(_ context.Context, account ledger.AccountID, on kernel.Date, cash kernel.Money) (app.AssertionCheck, error) {
	if b.err != nil {
		return app.AssertionCheck{}, fmt.Errorf("recording: %w", b.err)
	}
	return app.AssertionCheck{Assertion: ledger.CashCheckpoint(account, on, cash), Status: app.AssertionMatched, Derived: cash.Amount()}, nil
}

func (b fakeBook) BookCashDifference(context.Context, ledger.AccountID, kernel.Date, kernel.Currency, ledger.DifferenceBooking) (ledger.Transaction, error) {
	if b.err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking: %w", b.err)
	}
	return ledger.Transaction{}, nil
}

func TestCashCheckpointsNeedTheOwner(t *testing.T) {
	t.Parallel()
	on := date(t, "2026-03-31")
	pln := money(t, "1", "PLN")
	for _, tt := range []struct {
		name, session string
		book          fakeBook
		denied        []ledger.AccountID
		want          error
	}{
		{name: "the owner", session: "owner"},
		{name: "not the owner", session: "read", want: app.ErrNotOwner},
		{name: "the owner may not change the account", session: "owner", denied: []ledger.AccountID{acc1}, want: app.ErrNotOwner},
		{name: "the ledger refuses", session: "owner", book: fakeBook{err: ledger.ErrNoCashDifference}, want: ledger.ErrNoCashDifference},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &app.CashCheckpoints{Owners: fakeOwners{denied: tt.denied}, Ledger: tt.book}
			if _, err := c.Record(t.Context(), "o", tt.session, acc1, on, pln); !errors.Is(err, tt.want) {
				t.Errorf("Record = %v, want %v", err, tt.want)
			}
			if _, err := c.BookDifference(t.Context(), "o", tt.session, acc1, on, pln.Currency(), ledger.AsCashIncome); !errors.Is(err, tt.want) {
				t.Errorf("BookDifference = %v, want %v", err, tt.want)
			}
		})
	}
}
