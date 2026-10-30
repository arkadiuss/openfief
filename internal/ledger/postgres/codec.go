package postgres

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// parseKind returns the transaction kind whose String is s.
func parseKind(s string) (ledger.TransactionKind, error) {
	for k := ledger.KindBuy; k <= ledger.KindDeletion; k++ {
		if k.String() == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: transaction kind %q", ErrCorruptLog, s)
}

// parseBookKind returns the ledger account kind whose String is s.
func parseBookKind(s string) (ledger.BookKind, error) {
	for k := ledger.BookCash; k <= ledger.BookEquity; k++ {
		if k.String() == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: ledger account kind %q", ErrCorruptLog, s)
}

// parseStatus returns the assertion status whose String is s.
func parseStatus(s string) (app.AssertionStatus, error) {
	for st := app.AssertionMatched; st <= app.AssertionFlagged; st++ {
		if st.String() == s {
			return st, nil
		}
	}
	return 0, fmt.Errorf("%w: assertion status %q", ErrCorruptLog, s)
}

// parseDate parses a stored YYYY-MM-DD date.
func parseDate(s string) (kernel.Date, error) {
	d, err := kernel.ParseDate(s)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: %w", ErrCorruptLog, err)
	}
	return d, nil
}

// parseDecimal parses a stored numeric's text.
func parseDecimal(s string) (kernel.Decimal, error) {
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: %w", ErrCorruptLog, err)
	}
	return d, nil
}

// decodeBook returns the ledger account of kind with currency, or, for a
// holding, the instrument isin and symbol.
func decodeBook(kind, currency, isin, symbol string) (ledger.Book, error) {
	k, err := parseBookKind(kind)
	if err != nil {
		return ledger.Book{}, err
	}
	if k == ledger.BookHolding {
		return decodeHolding(isin, symbol)
	}
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		return ledger.Book{}, fmt.Errorf("%w: %w", ErrCorruptLog, err)
	}
	return ledger.Book{Kind: k, Currency: c}, nil
}

// decodeHolding returns the holding ledger account of the instrument isin
// and symbol.
func decodeHolding(isin, symbol string) (ledger.Book, error) {
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		return ledger.Book{}, fmt.Errorf("%w: %w", ErrCorruptLog, err)
	}
	return ledger.HoldingBook(ledger.Instrument{ISIN: i, Symbol: symbol}), nil
}

// decodePosting returns the posting stored as these columns.
func decodePosting(kind, isin, symbol, amount, currency, quantity string) (ledger.Posting, error) {
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		return ledger.Posting{}, fmt.Errorf("%w: %w", ErrCorruptLog, err)
	}
	book, err := decodeBook(kind, currency, isin, symbol)
	if err != nil {
		return ledger.Posting{}, err
	}
	a, err := parseDecimal(amount)
	if err != nil {
		return ledger.Posting{}, err
	}
	q, err := parseDecimal(quantity)
	if err != nil {
		return ledger.Posting{}, err
	}
	return ledger.Posting{Book: book, Amount: kernel.NewMoney(a, c), Quantity: kernel.NewQuantity(q)}, nil
}
