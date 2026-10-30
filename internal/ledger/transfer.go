package ledger

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// TransferWindowDays is how many calendar days apart a withdrawal and a
// deposit may be dated and still be a transfer between the owner's own
// accounts (D-105).
const TransferWindowDays = 5

// Errors of transfer links.
var (
	// ErrInvalidTransfer reports a withdrawal and a deposit that can't be
	// one transfer (CheckTransfer), or a decision that is not one.
	ErrInvalidTransfer = errors.New("invalid transfer")
	// ErrTransferTaken reports a confirmation of a withdrawal or deposit
	// that is already a leg of another confirmed transfer.
	ErrTransferTaken = errors.New("already a leg of a confirmed transfer")
)

// TransferDecision is the owner's answer to a suggested transfer (D-105).
type TransferDecision int

// The decisions. The zero decision is not a decision.
const (
	// TransferConfirmed means the two legs are one transfer between the
	// owner's own accounts: neither is a cash flow of the time-weighted
	// return.
	TransferConfirmed TransferDecision = iota + 1
	// TransferRejected means they are not: both stay cash flows, and the
	// pair is not suggested again.
	TransferRejected
)

// String returns the decision's name, "confirmed" or "rejected", or a
// placeholder for an invalid decision.
func (d TransferDecision) String() string {
	switch d {
	case TransferConfirmed:
		return "confirmed"
	case TransferRejected:
		return "rejected"
	}
	return fmt.Sprintf("TransferDecision(%d)", int(d))
}

// TransferPair names a withdrawal and a deposit by their transaction keys.
type TransferPair struct {
	Withdrawal TransactionKey
	Deposit    TransactionKey
}

// String returns "withdrawal -> deposit".
func (p TransferPair) String() string { return p.Withdrawal.String() + " -> " + p.Deposit.String() }

// TransferLink is one version of the owner's decision on a pair (D-105).
// Versions of a pair are numbered from 1; the highest is current and
// supersedes the earlier ones, which stay stored (D-18).
type TransferLink struct {
	Pair     TransferPair
	Decision TransferDecision
	Version  int
}

// TransferAmount returns the cash t moves as a leg of a transfer: what a
// deposit adds to its cash, or minus what a withdrawal takes. ok is false
// for any other transaction, for opening cash (IsSeeding) and for a
// transaction with other than one cash posting.
func TransferAmount(t *Transaction) (kernel.Money, bool) {
	if (t.Kind != KindDeposit && t.Kind != KindWithdrawal) || IsSeeding(t) {
		return kernel.Money{}, false
	}
	var amount kernel.Money
	n := 0
	for _, p := range t.Postings {
		if p.Book.Kind == BookCash {
			amount = p.Amount
			n++
		}
	}
	return amount, n == 1
}

// CheckTransfer reports whether w and d may be one transfer between the
// owner's own accounts (D-105): w a withdrawal and d a deposit, neither
// opening cash, in different accounts, of exactly the same amount and
// currency, dated at most TransferWindowDays apart in either order. The
// error wraps ErrInvalidTransfer and names the rule.
func CheckTransfer(w, d *Transaction) error {
	out, okW := TransferAmount(w)
	in, okD := TransferAmount(d)
	switch {
	case !okW || w.Kind != KindWithdrawal:
		return fmt.Errorf("%w: %s is not a withdrawal", ErrInvalidTransfer, w.Key)
	case !okD || d.Kind != KindDeposit:
		return fmt.Errorf("%w: %s is not a deposit", ErrInvalidTransfer, d.Key)
	case w.Account == d.Account:
		return fmt.Errorf("%w: %s and %s are in the same account", ErrInvalidTransfer, w.Key, d.Key)
	case !in.Equal(out.Neg()):
		return fmt.Errorf("%w: %s and %s differ in amount or currency", ErrInvalidTransfer, w.Key, d.Key)
	case max(w.Date.DaysUntil(d.Date), d.Date.DaysUntil(w.Date)) > TransferWindowDays:
		return fmt.Errorf("%w: %s and %s are more than %d days apart", ErrInvalidTransfer, w.Key, d.Key, TransferWindowDays)
	}
	return nil
}

// CurrentTransferLinks returns the current version of each pair's link, in
// withdrawal then deposit key order. Each pair's versions, in any order,
// must be numbered 1 to n with none repeated; otherwise it fails with an
// error wrapping ErrInvalidLog.
func CurrentTransferLinks(all []TransferLink) ([]TransferLink, error) {
	sorted := slices.Clone(all)
	slices.SortFunc(sorted, func(a, b TransferLink) int {
		return cmp.Or(comparePairs(a.Pair, b.Pair), cmp.Compare(a.Version, b.Version))
	})
	var out []TransferLink
	for i := range sorted {
		l := sorted[i]
		first := i == 0 || sorted[i-1].Pair != l.Pair
		switch {
		case first && l.Version != 1, !first && l.Version != sorted[i-1].Version+1:
			return nil, fmt.Errorf("%w: transfer %s has version %d out of sequence", ErrInvalidLog, l.Pair, l.Version)
		case first:
			out = append(out, l)
		default:
			out[len(out)-1] = l
		}
	}
	return out, nil
}

// DecideTransferLink returns the link that records decision on pair, given
// the current links (CurrentTransferLinks): the pair's next version. ok is
// false when the pair's current decision already is decision, so there is
// nothing to append and the link returned is the current one. Confirming fails with an
// error wrapping ErrTransferTaken when the withdrawal or the deposit is a leg
// of another pair's confirmed link, so each leg is in at most one transfer;
// a decision that is neither fails with ErrInvalidTransfer.
func DecideTransferLink(current []TransferLink, pair TransferPair, decision TransferDecision) (TransferLink, bool, error) {
	if decision != TransferConfirmed && decision != TransferRejected {
		return TransferLink{}, false, fmt.Errorf("%w: %s", ErrInvalidTransfer, decision)
	}
	next := TransferLink{Pair: pair, Decision: decision, Version: 1}
	for _, l := range current {
		if l.Pair == pair {
			if l.Decision == decision {
				return l, false, nil
			}
			next.Version = l.Version + 1
			continue
		}
		if decision == TransferConfirmed && l.Decision == TransferConfirmed &&
			(l.Pair.Withdrawal == pair.Withdrawal || l.Pair.Deposit == pair.Deposit) {
			return TransferLink{}, false, fmt.Errorf("transfer %s: %w %s", pair, ErrTransferTaken, l.Pair)
		}
	}
	return next, true, nil
}

// comparePairs orders pairs by withdrawal, then deposit key.
func comparePairs(a, b TransferPair) int {
	return cmp.Or(compareKeys(a.Withdrawal, b.Withdrawal), compareKeys(a.Deposit, b.Deposit))
}

// compareKeys orders transaction keys by source, then external ID.
func compareKeys(a, b TransactionKey) int {
	return cmp.Or(cmp.Compare(a.Source, b.Source), cmp.Compare(a.ExternalID, b.ExternalID))
}
