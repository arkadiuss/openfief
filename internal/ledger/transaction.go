package ledger

import (
	"errors"
	"fmt"
	"sort"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidTransaction reports a transaction or posting that is malformed:
// a missing key, account, kind or date, or a posting that doesn't fit its
// ledger account. An unbalanced transaction is an *UnbalancedError instead.
var ErrInvalidTransaction = errors.New("invalid transaction")

// minPostings is the fewest postings a double-entry transaction can have.
const minPostings = 2

// TransactionKey identifies a transaction across re-imports: the source it
// came from plus the institution's own ID for it (D-18). All versions of a
// transaction share one key.
type TransactionKey struct {
	// Source names where the transaction came from, such as a connector's
	// import path for one account.
	Source string
	// ExternalID is the institution's own ID for the transaction.
	ExternalID string
}

// String returns "source/external ID".
func (k TransactionKey) String() string { return k.Source + "/" + k.ExternalID }

// TransactionKind says what economic event a transaction records.
type TransactionKind int

// The transaction kinds. The zero kind is not a kind.
const (
	KindBuy TransactionKind = iota + 1
	KindSell
	KindDividend
	KindFee
	KindFXConversion
	KindDeposit
	KindWithdrawal
	// KindOpeningLot seeds a lot from a broker's current lot-level positions
	// at go-live (D-23).
	KindOpeningLot
)

// String returns the kind's name, such as "buy", or a placeholder for an
// invalid kind.
func (k TransactionKind) String() string {
	switch k {
	case KindBuy:
		return "buy"
	case KindSell:
		return "sell"
	case KindDividend:
		return "dividend"
	case KindFee:
		return "fee"
	case KindFXConversion:
		return "fx-conversion"
	case KindDeposit:
		return "deposit"
	case KindWithdrawal:
		return "withdrawal"
	case KindOpeningLot:
		return "opening-lot"
	}
	return fmt.Sprintf("TransactionKind(%d)", int(k))
}

// Posting is one leg of a transaction: an amount, and for a holding a
// quantity, booked to one ledger account. Amount always carries a currency;
// for every ledger account but a holding it is the ledger account's currency
// and Quantity is zero.
type Posting struct {
	Book     Book
	Amount   kernel.Money
	Quantity kernel.Quantity
}

// Validate reports whether p fits its ledger account and moves something.
func (p *Posting) Validate() error {
	if err := p.Book.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTransaction, err)
	}
	if p.Amount.Currency().IsZero() {
		return fmt.Errorf("%w: posting to %s has no currency", ErrInvalidTransaction, p.Book)
	}
	if p.Book.Kind != BookHolding {
		if !p.Amount.Currency().Equal(p.Book.Currency) {
			return fmt.Errorf("%w: posting of %s to %s", ErrInvalidTransaction, p.Amount.Currency(), p.Book)
		}
		if !p.Quantity.IsZero() {
			return fmt.Errorf("%w: posting to %s has a quantity", ErrInvalidTransaction, p.Book)
		}
	}
	if p.Amount.IsZero() && p.Quantity.IsZero() {
		return fmt.Errorf("%w: empty posting to %s", ErrInvalidTransaction, p.Book)
	}
	return nil
}

// Transaction is one economic event from a source, recorded in one account as
// postings that sum to zero per currency (D-19). It is one version of the
// transaction its Key names.
type Transaction struct {
	Key     TransactionKey
	Account AccountID
	Kind    TransactionKind
	// Date is the date the event happened, such as the trade date.
	Date kernel.Date
	// LotOpened is the date an opening lot was first acquired, on or before
	// Date. Only KindOpeningLot has one.
	LotOpened   kernel.Date
	Description string
	Postings    []Posting
}

// UnbalancedError reports a transaction whose postings don't sum to zero in
// one currency. Sum is what they sum to.
type UnbalancedError struct {
	Key TransactionKey
	Sum kernel.Money
}

// Error names the transaction and the currency that doesn't balance.
func (e *UnbalancedError) Error() string {
	return fmt.Sprintf("transaction %s: postings sum to %s %s, not zero", e.Key, e.Sum.Amount(), e.Sum.Currency())
}

// Validate reports whether t is well formed and balanced: it returns an error
// wrapping ErrInvalidTransaction for a malformed transaction, and an
// *UnbalancedError, naming the first currency in code order, when its
// postings don't sum to zero per currency.
func (t *Transaction) Validate() error {
	if err := t.validateHeader(); err != nil {
		return err
	}
	if len(t.Postings) < minPostings {
		return fmt.Errorf("%w %s: %d postings, need at least %d", ErrInvalidTransaction, t.Key, len(t.Postings), minPostings)
	}
	for i := range t.Postings {
		if err := t.Postings[i].Validate(); err != nil {
			return fmt.Errorf("transaction %s: %w", t.Key, err)
		}
	}
	for _, sum := range sumsByCurrency(t.Postings) {
		if !sum.IsZero() {
			return &UnbalancedError{Key: t.Key, Sum: sum}
		}
	}
	return nil
}

// validateHeader checks everything about t except its postings.
func (t *Transaction) validateHeader() error {
	switch {
	case t.Key.Source == "" || t.Key.ExternalID == "":
		return fmt.Errorf("%w %s: needs a source and an external ID", ErrInvalidTransaction, t.Key)
	case t.Account == "":
		return fmt.Errorf("%w %s: no account", ErrInvalidTransaction, t.Key)
	case t.Kind < KindBuy || t.Kind > KindOpeningLot:
		return fmt.Errorf("%w %s: %s", ErrInvalidTransaction, t.Key, t.Kind)
	case t.Date.IsZero():
		return fmt.Errorf("%w %s: no date", ErrInvalidTransaction, t.Key)
	case (t.Kind == KindOpeningLot) == t.LotOpened.IsZero():
		return fmt.Errorf("%w %s: only an opening lot has, and must have, a lot open date", ErrInvalidTransaction, t.Key)
	case t.LotOpened.After(t.Date):
		return fmt.Errorf("%w %s: lot opened %s after %s", ErrInvalidTransaction, t.Key, t.LotOpened, t.Date)
	}
	return nil
}

// sumsByCurrency sums the postings' amounts per currency, in currency code
// order.
func sumsByCurrency(ps []Posting) []kernel.Money {
	sums := map[kernel.Currency]kernel.Decimal{}
	for _, p := range ps {
		c := p.Amount.Currency()
		sums[c] = sums[c].Add(p.Amount.Amount())
	}
	out := make([]kernel.Money, 0, len(sums))
	for c, d := range sums {
		out = append(out, kernel.NewMoney(d, c))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Currency().String() < out[j].Currency().String() })
	return out
}
