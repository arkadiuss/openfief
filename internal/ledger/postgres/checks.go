package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// SaveAssertionCheck creates c or replaces the stored check of the same
// account, date and ledger account.
func (s *Store) SaveAssertionCheck(ctx context.Context, c *app.AssertionCheck) error {
	a := &c.Assertion
	var currency *string
	if a.Book.Kind != ledger.BookHolding {
		code := a.Book.Currency.String()
		currency = &code
	}
	in := columnsOf(a.Book)
	_, err := s.db.Exec(ctx, `
		INSERT INTO ledger.assertion_checks
		    (account_id, date, book, book_kind, currency, isin, symbol, balance, status, derived)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (account_id, date, book) DO UPDATE SET
		    balance = EXCLUDED.balance, status = EXCLUDED.status, derived = EXCLUDED.derived, checked_at = now()`,
		string(a.Account), a.Date.String(), a.Book.String(), a.Book.Kind.String(), currency, in.isin, in.symbol,
		a.Balance.String(), c.Status.String(), c.Derived.String())
	if err != nil {
		return fmt.Errorf("saving the balance assertion check of %s on %s: %w", a.Account, a.Date, err)
	}
	return nil
}

// AssertionChecks returns account's checks, in date then ledger account
// order.
func (s *Store) AssertionChecks(ctx context.Context, account ledger.AccountID) ([]app.AssertionCheck, error) {
	rows, err := s.db.Query(ctx, `
		SELECT to_char(date, 'YYYY-MM-DD'), book_kind, coalesce(currency, ''), coalesce(isin, ''), coalesce(symbol, ''),
		       balance::text, status, derived::text
		FROM ledger.assertion_checks
		WHERE account_id = $1
		ORDER BY date, book COLLATE "C"`, string(account))
	if err != nil {
		return nil, fmt.Errorf("loading balance assertion checks of %s: %w", account, err)
	}
	out, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (app.AssertionCheck, error) {
		return scanCheck(row, account)
	})
	if err != nil {
		return nil, fmt.Errorf("loading balance assertion checks of %s: %w", account, err)
	}
	return out, nil
}

// scanCheck reads one row of AssertionChecks' query.
func scanCheck(row pgx.CollectableRow, account ledger.AccountID) (app.AssertionCheck, error) {
	var date, kind, currency, isin, symbol, balance, status, derived string
	if err := row.Scan(&date, &kind, &currency, &isin, &symbol, &balance, &status, &derived); err != nil {
		return app.AssertionCheck{}, fmt.Errorf("reading a check: %w", err)
	}
	c := app.AssertionCheck{Assertion: ledger.BalanceAssertion{Account: account}}
	var err error
	if c.Assertion.Date, err = parseDate(date); err != nil {
		return app.AssertionCheck{}, err
	}
	if c.Assertion.Book, err = decodeBook(kind, currency, isin, symbol); err != nil {
		return app.AssertionCheck{}, err
	}
	if c.Assertion.Balance, err = parseDecimal(balance); err != nil {
		return app.AssertionCheck{}, err
	}
	if c.Status, err = parseStatus(status); err != nil {
		return app.AssertionCheck{}, err
	}
	if c.Derived, err = parseDecimal(derived); err != nil {
		return app.AssertionCheck{}, err
	}
	return c, nil
}
