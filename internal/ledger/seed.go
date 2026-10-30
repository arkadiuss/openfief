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
	case KindBuy, KindSell, KindDividend, KindFee, KindFXConversion, KindInterest, KindDeletion:
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

// Errors adding opening lots one at a time.
var (
	// ErrGoLiveMismatch reports opening lots added one at a time on another
	// date than the account's go-live date, or on two dates at once.
	ErrGoLiveMismatch = errors.New("opening lots not on the account's go-live date")
	// ErrBookedBeforeGoLive reports opening lots for an account that is not
	// seeded but has a current transaction dated on or before them: the lot
	// would count what that transaction holds twice.
	ErrBookedBeforeGoLive = errors.New("the account has transactions on or before go-live")
	// ErrNotOpeningLot reports a seeding transaction other than an opening
	// lot, such as opening cash, among opening lots added one at a time.
	ErrNotOpeningLot = errors.New("only opening lots are added one at a time")
)

// CheckOpeningLots checks adding the seeding transactions of account in
// incoming to it one at a time, as a manually entered account receives its
// opening lots (D-98), given log, every stored version of every key in
// account. Unlike a seeding (DecideSeed), the account may already be
// seeded. Every incoming seeding transaction must be an opening lot
// (ErrNotOpeningLot), all on one date, and that date must be the account's
// go-live date (GoLive) when it is seeded, so its go-live never moves
// (ErrGoLiveMismatch). An account not seeded takes that date as its go-live
// and must have no current transaction dated on or before it
// (ErrBookedBeforeGoLive). Each lot is then imported like any transaction,
// keyed per lot, so entering one again unchanged changes nothing
// (DecideImport). Transactions of other accounts and those that don't seed
// are ignored.
func CheckOpeningLots(log []VersionedTransaction, account AccountID, incoming []Transaction) error {
	day, err := openingLotsDate(account, incoming)
	if err != nil || day.IsZero() {
		return err
	}
	goLive, seeded, err := GoLive(log, account)
	if err != nil {
		return err
	}
	if seeded {
		if !goLive.Equal(day) {
			return fmt.Errorf("%w: %s went live on %s, not %s", ErrGoLiveMismatch, account, goLive, day)
		}
		return nil
	}
	current, err := CurrentTransactions(log)
	if err != nil {
		return err
	}
	for i := range current {
		if t := &current[i]; t.Account == account && !t.Date.After(day) {
			return fmt.Errorf("%w: %s has %s on %s, go-live %s", ErrBookedBeforeGoLive, account, t.Key, t.Date, day)
		}
	}
	return nil
}

// openingLotsDate returns the one date of the seeding transactions of
// account in incoming, all opening lots, or the zero date when there are
// none.
func openingLotsDate(account AccountID, incoming []Transaction) (kernel.Date, error) {
	var day kernel.Date
	for i := range incoming {
		t := &incoming[i]
		switch {
		case t.Account != account || !IsSeeding(t):
			continue
		case t.Kind != KindOpeningLot:
			return kernel.Date{}, fmt.Errorf("%w: %s", ErrNotOpeningLot, t.Key)
		case !day.IsZero() && !t.Date.Equal(day):
			return kernel.Date{}, fmt.Errorf("%w: %s on %s and %s", ErrGoLiveMismatch, account, day, t.Date)
		}
		day = t.Date
	}
	return day, nil
}
