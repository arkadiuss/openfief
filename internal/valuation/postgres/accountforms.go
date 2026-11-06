package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// AccountForm returns account's form (D-147), and false when none is
// stored.
func (s *Store) AccountForm(ctx context.Context, account ledger.AccountID) (jurisdiction.AccountForm, bool, error) {
	var form string
	err := s.pool.QueryRow(ctx, `SELECT form FROM valuation.account_forms WHERE account_id = $1`, string(account)).Scan(&form)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("loading the form of %s: %w", account, err)
	}
	return jurisdiction.AccountForm(form), true, nil
}

// SaveAccountForm records account's form, replacing any earlier one. No
// stored valuation depends on it, so nothing is revalued.
func (s *Store) SaveAccountForm(ctx context.Context, account ledger.AccountID, form jurisdiction.AccountForm) error {
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.account_forms (account_id, form) VALUES ($1, $2)
		ON CONFLICT (account_id) DO UPDATE SET form = EXCLUDED.form, set_at = now()`, string(account), string(form)); err != nil {
		return fmt.Errorf("saving the form of %s: %w", account, err)
	}
	return nil
}

// EarlierPayments returns the total paid into account before its history
// is known, as the owner entered it (D-149), and false when none is stored.
func (s *Store) EarlierPayments(ctx context.Context, account ledger.AccountID) (kernel.Money, bool, error) {
	var amount, currency string
	err := s.pool.QueryRow(ctx, `SELECT amount::text, currency FROM valuation.earlier_payments WHERE account_id = $1`,
		string(account)).Scan(&amount, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return kernel.Money{}, false, nil
	}
	if err != nil {
		return kernel.Money{}, false, fmt.Errorf("loading the earlier payments of %s: %w", account, err)
	}
	a, err1 := kernel.ParseDecimal(amount)
	c, err2 := kernel.ParseCurrency(currency)
	if err := errors.Join(err1, err2); err != nil {
		return kernel.Money{}, false, fmt.Errorf("%w: earlier payments of %s: %w", ErrCorrupt, account, err)
	}
	return kernel.NewMoney(a, c), true, nil
}

// SaveEarlierPayments records the total paid into account before its
// history is known, replacing any earlier one. Totals are taxed when read,
// so nothing is revalued.
func (s *Store) SaveEarlierPayments(ctx context.Context, account ledger.AccountID, paid kernel.Money) error {
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.earlier_payments (account_id, amount, currency) VALUES ($1, $2, $3)
		ON CONFLICT (account_id) DO UPDATE SET amount = EXCLUDED.amount, currency = EXCLUDED.currency, set_at = now()`,
		string(account), paid.Amount().String(), paid.Currency().String()); err != nil {
		return fmt.Errorf("saving the earlier payments of %s: %w", account, err)
	}
	return nil
}
