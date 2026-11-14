package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// CashCheckpointKind is the kind of attention item a mismatched cash
// checkpoint raises (D-158, D-130). Its subject and detail come from
// ledger.CheckpointSubject and ledger.CheckpointDetail.
const CashCheckpointKind = "cash-checkpoint-mismatch"

// ErrUnknownCheckpoint reports a cash checkpoint the owner never entered:
// no difference is booked from it.
var ErrUnknownCheckpoint = errors.New("unknown cash checkpoint")

// AttentionRaiser raises and clears attention items for the owner (D-130),
// the shape of the attention store's port; cmd/core adapts the attention
// items to it. Both are idempotent: raising an item that is open, or
// clearing one that isn't, changes nothing.
type AttentionRaiser interface {
	Raise(ctx context.Context, kind, subject, detail string) error
	Clear(ctx context.Context, kind, subject string) error
}

// WithAttention makes l raise an attention item through a for a cash
// checkpoint that doesn't match the derived cash when the owner enters it,
// and clear it once the owner books the difference (D-158), and returns l.
// The attention context's own check keeps the items in step after later
// imports.
func (l *Ledger) WithAttention(a AttentionRaiser) *Ledger {
	l.attention = a
	return l
}

// RecordCashCheckpoint stores the cash an institution showed for account at
// the end of on, as the owner entered it (D-158): a balance assertion
// (D-18), checked against the derived cash like any other, which it never
// changes. A checkpoint for the same date and currency replaces the
// earlier one. A mismatch raises an attention item; a match clears it. It
// returns the checkpoint's check.
func (l *Ledger) RecordCashCheckpoint(ctx context.Context, account ledger.AccountID, on kernel.Date, cash kernel.Money) (AssertionCheck, error) {
	a := ledger.CashCheckpoint(account, on, cash)
	if err := a.Validate(); err != nil {
		return AssertionCheck{}, fmt.Errorf("recording a cash checkpoint: %w", err)
	}
	sum, err := l.Import(ctx, nil, []ledger.BalanceAssertion{a})
	if err != nil {
		return AssertionCheck{}, fmt.Errorf("recording a cash checkpoint: %w", err)
	}
	return l.afterCheckpoint(ctx, &sum, &a)
}

// BookCashDifference books the difference between account's cash
// checkpoint of on in c and the cash derived on that date as an ordinary
// manual transaction (ledger.CashDifference), as the owner chose: by
// default a deposit when the institution showed more, a withdrawal when it
// showed less, or cash income (D-152). It returns the transaction. A checkpoint the owner never entered fails with
// ErrUnknownCheckpoint, a matching one with ledger.ErrNoCashDifference.
// Once booked, the checkpoint matches and its attention item clears.
func (l *Ledger) BookCashDifference(ctx context.Context, account ledger.AccountID, on kernel.Date, c kernel.Currency, as ledger.DifferenceBooking) (ledger.Transaction, error) {
	a, err := l.checkpoint(ctx, account, on, c)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	log, err := l.store.Log(ctx, account)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: loading the log of %s: %w", account, err)
	}
	derived, err := ledger.CashOn(log, account, c, on)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	t, err := ledger.CashDifference(&a, derived.Amount(), as)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	sum, err := l.Import(ctx, []ledger.Transaction{t}, nil)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	if _, err := l.afterCheckpoint(ctx, &sum, &a); err != nil {
		return ledger.Transaction{}, err
	}
	return t, nil
}

// checkpoint returns account's stored cash checkpoint of on in c, or an
// error wrapping ErrUnknownCheckpoint.
func (l *Ledger) checkpoint(ctx context.Context, account ledger.AccountID, on kernel.Date, c kernel.Currency) (ledger.BalanceAssertion, error) {
	if _, err := lookupAccount(ctx, l.store, account); err != nil {
		return ledger.BalanceAssertion{}, err
	}
	checks, err := l.store.AssertionChecks(ctx, account)
	if err != nil {
		return ledger.BalanceAssertion{}, fmt.Errorf("loading balance assertions of %s: %w", account, err)
	}
	for i := range checks {
		if a := &checks[i].Assertion; a.Entered && a.Date.Equal(on) && a.Book == ledger.CashBook(c) {
			return *a, nil
		}
	}
	return ledger.BalanceAssertion{}, fmt.Errorf("%w: %s %s on %s", ErrUnknownCheckpoint, account, c, on)
}

// afterCheckpoint returns the check of checkpoint a among sum's, after
// raising its attention item when it is flagged or clearing it when it
// matches.
func (l *Ledger) afterCheckpoint(ctx context.Context, sum *ImportSummary, a *ledger.BalanceAssertion) (AssertionCheck, error) {
	for i := range sum.Checks {
		c := &sum.Checks[i]
		if c.Assertion.Date.Equal(a.Date) && c.Assertion.Book == a.Book {
			return *c, l.notifyCheckpoint(ctx, c)
		}
	}
	return AssertionCheck{}, fmt.Errorf("%w: %s %s on %s", ErrUnknownCheckpoint, a.Account, a.Book, a.Date)
}

// notifyCheckpoint raises c's attention item when it is flagged, and
// clears it otherwise.
func (l *Ledger) notifyCheckpoint(ctx context.Context, c *AssertionCheck) error {
	if l.attention == nil {
		return nil
	}
	subject := ledger.CheckpointSubject(&c.Assertion)
	if c.Flagged() {
		if err := l.attention.Raise(ctx, CashCheckpointKind, subject, ledger.CheckpointDetail(&c.Assertion, c.Derived)); err != nil {
			return fmt.Errorf("raising the cash checkpoint's attention item: %w", err)
		}
		return nil
	}
	if err := l.attention.Clear(ctx, CashCheckpointKind, subject); err != nil {
		return fmt.Errorf("clearing the cash checkpoint's attention item: %w", err)
	}
	return nil
}

// CashCheckpoints are the owner's cash checkpoints (D-158), for an
// owner-scoped session only (D-79).
type CashCheckpoints struct {
	Owners Owners
	Ledger CashCheckpointBook
	// Audit keeps each checkpoint and booking with its audit entry; nil
	// enters none.
	Audit ChangeAuditor
}

// CashCheckpointBook records cash checkpoints and books their
// differences; *Ledger implements it.
type CashCheckpointBook interface {
	RecordCashCheckpoint(ctx context.Context, account ledger.AccountID, on kernel.Date, cash kernel.Money) (AssertionCheck, error)
	BookCashDifference(ctx context.Context, account ledger.AccountID, on kernel.Date, c kernel.Currency, as ledger.DifferenceBooking) (ledger.Transaction, error)
}

// Record records the cash the institution showed for account on a date, as
// the owner entered it (Ledger.RecordCashCheckpoint). The caller must be
// the owner and may change account (D-163); otherwise it fails with
// ErrNotOwner and stores nothing.
func (c *CashCheckpoints) Record(ctx context.Context, origin, sessionToken string, account ledger.AccountID, on kernel.Date, cash kernel.Money) (AssertionCheck, error) {
	if err := c.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return AssertionCheck{}, fmt.Errorf("recording a cash checkpoint: %w", err)
	}
	if err := mayChange(ctx, c.Owners, origin, sessionToken, []ledger.AccountID{account}); err != nil {
		return AssertionCheck{}, fmt.Errorf("recording a cash checkpoint: %w", err)
	}
	var check AssertionCheck
	err := audited(ctx, c.Audit, origin, sessionToken, auditCashCheckpointRecorded, accountRefs(account), func(ctx context.Context) error {
		var recordErr error
		check, recordErr = c.Ledger.RecordCashCheckpoint(ctx, account, on, cash)
		if recordErr != nil {
			return fmt.Errorf("recording: %w", recordErr)
		}
		return nil
	})
	if err != nil {
		return AssertionCheck{}, fmt.Errorf("recording a cash checkpoint: %w", err)
	}
	return check, nil
}

// BookDifference books the difference to account's checkpoint of on in cur
// as the owner chose, a deposit or withdrawal by default, or cash income
// (Ledger.BookCashDifference). The caller must
// be the owner and may change account (D-163); otherwise it fails with
// ErrNotOwner and books nothing.
func (c *CashCheckpoints) BookDifference(ctx context.Context, origin, sessionToken string, account ledger.AccountID, on kernel.Date, cur kernel.Currency, as ledger.DifferenceBooking) (ledger.Transaction, error) {
	if err := c.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	if err := mayChange(ctx, c.Owners, origin, sessionToken, []ledger.AccountID{account}); err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	var t ledger.Transaction
	err := audited(ctx, c.Audit, origin, sessionToken, auditCashDifferenceBooked, accountRefs(account), func(ctx context.Context) error {
		var bookErr error
		t, bookErr = c.Ledger.BookCashDifference(ctx, account, on, cur, as)
		if bookErr != nil {
			return fmt.Errorf("booking: %w", bookErr)
		}
		return nil
	})
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking a cash difference: %w", err)
	}
	return t, nil
}
