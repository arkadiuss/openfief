package ledger

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrAlreadySeeded reports a seeding of an account that is already seeded.
// An *AlreadySeededError wraps it.
var ErrAlreadySeeded = errors.New("account already seeded")

// OpeningCashPrefix starts the external ID of every opening cash
// transaction. Opening cash is an ordinary deposit, or a withdrawal for a
// negative balance, against the equity ledger account (D-23, D-33); this
// prefix is what tells it apart from other deposits and withdrawals, so a
// connector must key its opening cash with OpeningCashID.
const OpeningCashPrefix = "opening-cash:"

// OpeningCashID returns the external ID of the opening cash in c on the
// go-live date on: "opening-cash:<currency>:<date>". One account has at most
// one opening cash transaction per currency and date.
func OpeningCashID(c kernel.Currency, on kernel.Date) string {
	return OpeningCashPrefix + c.String() + ":" + on.String()
}

// NewOpeningCash books the cash an account holds at go-live (D-23): a
// deposit of a positive balance or a withdrawal of a negative one, against
// the equity ledger account (D-33). The header's key must have an external
// ID from OpeningCashID, and cash must not be zero.
func NewOpeningCash(h TransactionHeader, cash kernel.Money) (Transaction, error) {
	if !strings.HasPrefix(h.Key.ExternalID, OpeningCashPrefix) {
		return Transaction{}, fmt.Errorf("%w %s: opening cash needs an external ID starting %q", ErrInvalidTransaction, h.Key, OpeningCashPrefix)
	}
	if cash.Sign() < 0 {
		return NewWithdrawal(h, cash.Neg())
	}
	return NewDeposit(h, cash)
}

// IsSeeding reports whether t seeds an account at go-live: an opening lot,
// or opening cash, a deposit or withdrawal keyed by OpeningCashID.
func IsSeeding(t *Transaction) bool {
	switch t.Kind {
	case KindOpeningLot:
		return true
	case KindDeposit, KindWithdrawal:
		return strings.HasPrefix(t.Key.ExternalID, OpeningCashPrefix)
	case KindBuy, KindSell, KindDividend, KindFee, KindFXConversion:
		return false
	}
	return false
}

// IsSeeded reports whether account is already seeded: whether the current
// transactions of account in log include a seeding one (IsSeeding).
// Superseded versions don't count.
func IsSeeded(log []VersionedTransaction, account AccountID) (bool, error) {
	current, err := CurrentTransactions(log)
	if err != nil {
		return false, err
	}
	for i := range current {
		if current[i].Account == account && IsSeeding(&current[i]) {
			return true, nil
		}
	}
	return false, nil
}

// GoLive returns account's go-live date: the latest date of a current
// seeding transaction of account in log (IsSeeding), which is the day its
// opening lots and opening cash were taken (D-23); and false when account
// is not seeded. Superseded versions don't count.
func GoLive(log []VersionedTransaction, account AccountID) (kernel.Date, bool, error) {
	current, err := CurrentTransactions(log)
	if err != nil {
		return kernel.Date{}, false, err
	}
	var day kernel.Date
	seeded := false
	for i := range current {
		t := &current[i]
		if t.Account != account || !IsSeeding(t) {
			continue
		}
		if !seeded || t.Date.After(day) {
			day, seeded = t.Date, true
		}
	}
	return day, seeded, nil
}

// SeedDecision says what to do with a seeding of one account.
type SeedDecision int

// The seed decisions. The zero decision is not a decision.
const (
	// SeedFirst means the account is not seeded yet: import the seeding.
	SeedFirst SeedDecision = iota + 1
	// SeedUnchanged means the account is already seeded and every incoming
	// transaction of it equals its current version: the seeding is a no-op.
	SeedUnchanged
)

// AlreadySeededError reports a seeding of Account, which is already seeded,
// that would change its log. It wraps ErrAlreadySeeded.
type AlreadySeededError struct {
	Account AccountID
}

// Error names the account.
func (e *AlreadySeededError) Error() string {
	return fmt.Sprintf("account %s is already seeded: seeding it again is refused", e.Account)
}

// Unwrap returns ErrAlreadySeeded.
func (e *AlreadySeededError) Unwrap() error { return ErrAlreadySeeded }

// DecideSeed decides a seeding of account at go-live (D-23), given log, every
// stored version of every key in account, and incoming, the seeding's
// transactions (those of other accounts are ignored). An account that is not
// seeded (IsSeeded) is SeedFirst. An already seeded account is refused with
// an *AlreadySeededError, unless every incoming transaction of it is
// unchanged (DecideImport): re-running the same seeding is SeedUnchanged.
func DecideSeed(log []VersionedTransaction, account AccountID, incoming []Transaction) (SeedDecision, error) {
	seeded, err := IsSeeded(log, account)
	if err != nil {
		return 0, err
	}
	if !seeded {
		return SeedFirst, nil
	}
	byKey := map[TransactionKey][]VersionedTransaction{}
	for i := range log {
		byKey[log[i].Transaction.Key] = append(byKey[log[i].Transaction.Key], log[i])
	}
	for i := range incoming {
		t := &incoming[i]
		if t.Account != account {
			continue
		}
		res, err := DecideImport(byKey[t.Key], t)
		if err != nil {
			return 0, err
		}
		if res.Decision != ImportUnchanged {
			return 0, &AlreadySeededError{Account: account}
		}
	}
	return SeedUnchanged, nil
}
