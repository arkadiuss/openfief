package ledger

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidAssertion reports a balance assertion without an account or
// date, or on a ledger account other than cash or a holding.
var ErrInvalidAssertion = errors.New("invalid balance assertion")

// BalanceAssertion is a broker-reported balance on a date that the derived
// balance must match (D-18): cash in one currency, or units of one
// instrument, in one account at the end of Date.
type BalanceAssertion struct {
	Account AccountID
	Date    kernel.Date
	// Book is CashBook of the currency or HoldingBook of the instrument.
	Book Book
	// Balance is the reported amount in the cash currency, or the reported
	// number of units of the instrument.
	Balance kernel.Decimal
}

// CashAssertion returns the assertion that account holds cash at the end of
// on.
func CashAssertion(account AccountID, on kernel.Date, cash kernel.Money) BalanceAssertion {
	return BalanceAssertion{Account: account, Date: on, Book: CashBook(cash.Currency()), Balance: cash.Amount()}
}

// HoldingAssertion returns the assertion that account holds units of in at
// the end of on.
func HoldingAssertion(account AccountID, on kernel.Date, in Instrument, units kernel.Quantity) BalanceAssertion {
	return BalanceAssertion{Account: account, Date: on, Book: HoldingBook(in), Balance: units.Decimal()}
}

// Validate reports whether a names an account, a date and a valid cash or
// holding ledger account.
func (a *BalanceAssertion) Validate() error {
	switch {
	case a.Account == "":
		return fmt.Errorf("%w: no account", ErrInvalidAssertion)
	case a.Date.IsZero():
		return fmt.Errorf("%w %s: no date", ErrInvalidAssertion, a.Account)
	case a.Book.Kind != BookCash && a.Book.Kind != BookHolding:
		return fmt.Errorf("%w %s: only cash and holdings are asserted, not %s", ErrInvalidAssertion, a.Account, a.Book)
	}
	if err := a.Book.Validate(); err != nil {
		return fmt.Errorf("%w %s: %w", ErrInvalidAssertion, a.Account, err)
	}
	return nil
}

// AssertionMismatchError reports a balance assertion that the derived
// balance does not match. It carries both values.
type AssertionMismatchError struct {
	Assertion BalanceAssertion
	// Derived is the balance derived from the log, in the assertion's unit.
	Derived kernel.Decimal
}

// Error names the account, ledger account, date and both balances.
func (e *AssertionMismatchError) Error() string {
	a := &e.Assertion
	return fmt.Sprintf("balance assertion failed for %s %s on %s: reported %s, derived %s",
		a.Account, a.Book, a.Date, a.Balance, e.Derived)
}

// CheckBalanceAssertion derives a's balance from the current versions of the
// log and returns nil if it matches, an *AssertionMismatchError with both
// values if it does not, or an error for an invalid assertion or log.
func CheckBalanceAssertion(log []VersionedTransaction, a *BalanceAssertion) error {
	if err := a.Validate(); err != nil {
		return err
	}
	derived, err := balanceOn(log, a.Account, a.Book, a.Date)
	if err != nil {
		return err
	}
	if !derived.Equal(a.Balance) {
		return &AssertionMismatchError{Assertion: *a, Derived: derived}
	}
	return nil
}

// CashOn returns account's cash in c at the end of on, derived from the
// current versions of the log.
func CashOn(log []VersionedTransaction, account AccountID, c kernel.Currency, on kernel.Date) (kernel.Money, error) {
	d, err := balanceOn(log, account, CashBook(c), on)
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(d, c), nil
}

// balanceOn sums book's postings in account up to the end of on: units for
// a holding, the amount for any other ledger account.
func balanceOn(log []VersionedTransaction, account AccountID, book Book, on kernel.Date) (kernel.Decimal, error) {
	txs, err := currentInAccount(log, account, on)
	if err != nil {
		return kernel.Decimal{}, err
	}
	var sum kernel.Decimal
	for i := range txs {
		for j := range txs[i].Postings {
			p := &txs[i].Postings[j]
			switch {
			case p.Book != book:
			case book.Kind == BookHolding:
				sum = sum.Add(p.Quantity.Decimal())
			default:
				sum = sum.Add(p.Amount.Amount())
			}
		}
	}
	return sum, nil
}
