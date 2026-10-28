package ledger

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// ErrInvalidAccount reports an account without an ID, an institution, a
// name or a sleeve.
var ErrInvalidAccount = errors.New("invalid account")

// ErrInvalidBook reports a ledger account whose kind, currency and
// instrument don't fit together.
var ErrInvalidBook = errors.New("invalid ledger account")

// AccountID identifies an account.
type AccountID string

// Account is one account at an institution, such as one IBKR linked account.
// It belongs to exactly one sleeve.
type Account struct {
	ID AccountID
	// Institution is the ingest context's ID of the institution that holds the
	// account.
	Institution string
	Name        string
	Sleeve      portfolio.SleeveID
}

// Validate reports whether a has an ID, an institution, a name and a sleeve.
func (a Account) Validate() error {
	switch {
	case a.ID == "":
		return fmt.Errorf("%w: empty ID", ErrInvalidAccount)
	case a.Institution == "":
		return fmt.Errorf("%w %s: no institution", ErrInvalidAccount, a.ID)
	case a.Name == "":
		return fmt.Errorf("%w %s: empty name", ErrInvalidAccount, a.ID)
	case a.Sleeve == "":
		return fmt.Errorf("%w %s: no sleeve", ErrInvalidAccount, a.ID)
	}
	return nil
}

// BookKind says what a ledger account holds.
type BookKind int

// The ledger account kinds. The zero kind is not a kind.
const (
	// BookCash holds cash in one currency.
	BookCash BookKind = iota + 1
	// BookHolding holds units of one instrument. Its postings carry the
	// quantity and the money value the transaction exchanged for it.
	BookHolding
	// BookFee holds fees paid, in one currency.
	BookFee
	// BookWithholdingTax holds tax withheld at source, in one currency.
	BookWithholdingTax
	// BookFX is the conversion side of currency exchange, one per currency,
	// so that each currency balances on its own.
	BookFX
	// BookEquity is the counterpart of what enters or leaves the account from
	// outside: deposits, withdrawals and opening lots, in one currency.
	BookEquity
)

// String returns the kind's name, such as "cash", or a placeholder for an
// invalid kind.
func (k BookKind) String() string {
	switch k {
	case BookCash:
		return "cash"
	case BookHolding:
		return "holding"
	case BookFee:
		return "fee"
	case BookWithholdingTax:
		return "withholding-tax"
	case BookFX:
		return "fx"
	case BookEquity:
		return "equity"
	}
	return fmt.Sprintf("BookKind(%d)", int(k))
}

// Book is a ledger account: a double-entry account inside an account, cash
// per currency, a holding per instrument, or fees, withholding tax, FX or
// equity per currency. (Its Go name avoids ledger.LedgerAccount, which the
// linter rejects as a stutter.) A holding has an instrument and no currency;
// every other kind has a currency and no instrument. Books are comparable and
// can be map keys.
type Book struct {
	Kind       BookKind
	Currency   kernel.Currency
	Instrument Instrument
}

// CashBook returns the cash ledger account for c.
func CashBook(c kernel.Currency) Book {
	return Book{Kind: BookCash, Currency: c}
}

// HoldingBook returns the holding ledger account for in.
func HoldingBook(in Instrument) Book {
	return Book{Kind: BookHolding, Instrument: in}
}

// FeeBook returns the fee ledger account for c.
func FeeBook(c kernel.Currency) Book {
	return Book{Kind: BookFee, Currency: c}
}

// WithholdingTaxBook returns the withholding tax ledger account for c.
func WithholdingTaxBook(c kernel.Currency) Book {
	return Book{Kind: BookWithholdingTax, Currency: c}
}

// FXBook returns the FX ledger account for c.
func FXBook(c kernel.Currency) Book {
	return Book{Kind: BookFX, Currency: c}
}

// EquityBook returns the equity ledger account for c.
func EquityBook(c kernel.Currency) Book {
	return Book{Kind: BookEquity, Currency: c}
}

// Validate reports whether l's kind is known and it has exactly what its kind
// needs: an instrument for a holding, a currency for every other kind.
func (l Book) Validate() error {
	switch l.Kind {
	case BookHolding:
		if !l.Currency.IsZero() {
			return fmt.Errorf("%w %s: a holding has no currency", ErrInvalidBook, l)
		}
		if err := l.Instrument.Validate(); err != nil {
			return fmt.Errorf("%w %s: %w", ErrInvalidBook, l, err)
		}
		return nil
	case BookCash, BookFee, BookWithholdingTax, BookFX, BookEquity:
		if l.Currency.IsZero() || l.Instrument != (Instrument{}) {
			return fmt.Errorf("%w %s: needs a currency and no instrument", ErrInvalidBook, l)
		}
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalidBook, l.Kind)
}

// String returns "kind:currency" or "holding:ISIN/symbol".
func (l Book) String() string {
	if l.Kind == BookHolding {
		return l.Kind.String() + ":" + l.Instrument.String()
	}
	return l.Kind.String() + ":" + l.Currency.String()
}
