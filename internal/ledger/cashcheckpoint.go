package ledger

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNoCashDifference reports a cash checkpoint the derived cash matches:
// there is no difference to book.
var ErrNoCashDifference = errors.New("the cash checkpoint matches the derived cash")

// ErrNotCheckpoint reports a balance assertion that is no cash checkpoint
// the owner entered, so no difference is booked from it.
var ErrNotCheckpoint = errors.New("not a cash checkpoint")

// CashCheckpointSource is the key source of the transactions the owner
// books from a cash checkpoint (D-158). Connectors' sources start with
// "plugin:", so no connector can write or supersede one.
const CashCheckpointSource = "owner:cash-checkpoint"

// CashCheckpoint returns the cash checkpoint that account held cash at the
// end of on, as the institution showed it and the owner entered it
// (D-158): a cash balance assertion marked Entered.
func CashCheckpoint(account AccountID, on kernel.Date, cash kernel.Money) BalanceAssertion {
	a := CashAssertion(account, on, cash)
	a.Entered = true
	return a
}

// DifferenceBooking is how the owner books a cash checkpoint's difference
// (D-158). The zero value is the default.
type DifferenceBooking int

// The ways to book a difference.
const (
	// AsDepositOrWithdrawal, the default, books a deposit when the
	// institution shows more cash and a withdrawal when it shows less, so
	// an unexplained difference never shows as return.
	AsDepositOrWithdrawal DifferenceBooking = iota
	// AsCashIncome books cash income (D-152), for a difference that is
	// income, such as interest, when the institution shows more cash, or a
	// fee when it shows less: performance counts it as return.
	AsCashIncome
)

// ErrUnknownBooking reports a DifferenceBooking outside the declared
// ones.
var ErrUnknownBooking = errors.New("unknown way to book a cash difference")

// CashDifference returns the ordinary manual transaction that books the
// difference between a cash checkpoint and the cash derived on its date
// (D-158), as the owner chose: by default a deposit when the institution
// shows more cash, a withdrawal when it shows less, so the difference
// never shows as return; or cash income of the difference, negative when
// the institution shows less (D-152). It is dated on the checkpoint's
// date, so the checkpoint matches once it is booked; nothing overwrites
// the derived cash. Its key names the checkpoint and both balances, so
// booking the same difference twice books it once. A matching checkpoint
// fails with ErrNoCashDifference, an assertion the owner didn't enter with
// ErrNotCheckpoint.
func CashDifference(a *BalanceAssertion, derived kernel.Decimal, as DifferenceBooking) (Transaction, error) {
	if err := a.Validate(); err != nil {
		return Transaction{}, err
	}
	if !a.Entered {
		return Transaction{}, fmt.Errorf("%w: %s %s on %s", ErrNotCheckpoint, a.Account, a.Book, a.Date)
	}
	diff := kernel.NewMoney(a.Balance.Sub(derived), a.Book.Currency)
	h := TransactionHeader{
		Key: TransactionKey{
			Source:     CashCheckpointSource,
			ExternalID: a.Date.String() + "/" + a.Book.Currency.String() + "/" + a.Balance.String() + "/" + derived.String(),
		},
		Account: a.Account, Date: a.Date,
	}
	if diff.Sign() == 0 {
		return Transaction{}, fmt.Errorf("%w: %s %s on %s", ErrNoCashDifference, a.Account, a.Book, a.Date)
	}
	t, err := differenceTransaction(h, diff, as)
	if err != nil {
		return Transaction{}, err
	}
	t.Description = fmt.Sprintf("Difference to the cash checkpoint of %s: %s shown, %s derived",
		a.Date, kernel.NewMoney(a.Balance, a.Book.Currency), kernel.NewMoney(derived, a.Book.Currency))
	return t, nil
}

// differenceTransaction books diff, not zero, as the owner chose.
func differenceTransaction(h TransactionHeader, diff kernel.Money, as DifferenceBooking) (Transaction, error) {
	switch as {
	case AsDepositOrWithdrawal:
		if diff.Sign() > 0 {
			return NewDeposit(h, diff)
		}
		return NewWithdrawal(h, diff.Neg())
	case AsCashIncome:
		return NewCashIncome(h, diff, kernel.Money{})
	}
	return Transaction{}, fmt.Errorf("%w: %d", ErrUnknownBooking, as)
}

// CheckpointSubject returns the subject of the attention item a
// mismatched cash checkpoint raises (D-158, D-130): the account, the date
// and the currency, as "acc-1/2026-03-31/PLN".
func CheckpointSubject(a *BalanceAssertion) string {
	return string(a.Account) + "/" + a.Date.String() + "/" + a.Book.Currency.String()
}

// CheckpointDetail returns the detail of that item: the cash shown, the
// cash derived and the difference.
func CheckpointDetail(a *BalanceAssertion, derived kernel.Decimal) string {
	c := a.Book.Currency
	return fmt.Sprintf("shown %s, derived %s, difference %s",
		kernel.NewMoney(a.Balance, c), kernel.NewMoney(derived, c), kernel.NewMoney(a.Balance.Sub(derived), c))
}
