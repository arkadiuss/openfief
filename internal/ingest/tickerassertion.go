package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidTickerAssertion reports a ticker assertion without a valid
// symbol or date, or of a negative quantity.
var ErrInvalidTickerAssertion = errors.New("invalid ticker assertion")

// TickerAssertion is a holding an institution that names instruments by
// ticker only reports for an account on a date, such as the summed volume
// of a ticker's open positions in an xStation export's Open Positions
// sheet (D-155). The core asserts it as a balance assertion (D-18) of the
// instrument the owner confirmed for the ticker (Assertion), and until
// then holds it, to assert once the owner confirms the ticker. Ref names
// its statement rows, as Problem.Ref does.
type TickerAssertion struct {
	Symbol   string
	Account  ledger.AccountID
	Date     kernel.Date
	Quantity kernel.Quantity
	Ref      string
}

// Validate reports whether a has a valid symbol, a date and a quantity
// that is not negative. Its account may still be empty when a broker
// account names it.
func (a *TickerAssertion) Validate() error {
	if err := validTickerSymbol(a.Symbol); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTickerAssertion, err)
	}
	switch {
	case a.Date.IsZero():
		return fmt.Errorf("%w %s: no date", ErrInvalidTickerAssertion, a.Symbol)
	case a.Quantity.Sign() < 0:
		return fmt.Errorf("%w %s: a negative quantity", ErrInvalidTickerAssertion, a.Symbol)
	}
	return nil
}

// Assertion returns the balance assertion a makes once its ticker's
// listing l is confirmed (D-155, D-116): the account holds a's quantity
// of the instrument l's ISIN with the ticker's symbol at the end of its
// date.
func (a *TickerAssertion) Assertion(l Listing) ledger.BalanceAssertion {
	return ledger.HoldingAssertion(a.Account, a.Date, ledger.Instrument{ISIN: l.ISIN, Symbol: a.Symbol}, a.Quantity)
}

// BookTickerAssertions returns m with each ticker assertion whose symbol
// confirmed maps to a listing added to its balance assertions
// (TickerAssertion.Assertion), and the ticker assertions that wait for the
// owner to confirm their ticker's ISIN, which are asserted then (D-155).
// m's ticker assertions must have their account.
func BookTickerAssertions(m *Mapping, confirmed map[string]Listing) (Mapping, []TickerAssertion) {
	out := *m
	out.Assertions = append([]ledger.BalanceAssertion(nil), m.Assertions...)
	out.TickerAssertions = nil
	var waiting []TickerAssertion
	for i := range m.TickerAssertions {
		a := &m.TickerAssertions[i]
		listing, ok := confirmed[a.Symbol]
		if !ok {
			waiting = append(waiting, *a)
			continue
		}
		out.Assertions = append(out.Assertions, a.Assertion(listing))
	}
	return out, waiting
}
