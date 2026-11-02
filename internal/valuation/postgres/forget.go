package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ForgetAccount removes what the valuation schema holds of a deleted
// account (D-120), in one transaction: its tax wrapper and fee rule, its
// charts' start (D-167), its stored valuations, its valuation status and
// any pending revaluation. An
// account the ledger deleted held no transactions, so its valuations are
// derived from nothing (D-36). Forgetting it again changes nothing.
func (s *Store) ForgetAccount(ctx context.Context, account ledger.AccountID) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, q := range []string{
			"DELETE FROM valuation.tax_wrappers WHERE account_id = $1", "DELETE FROM valuation.fee_rules WHERE account_id = $1",
			"DELETE FROM valuation.holding_valuations WHERE account_id = $1",
			"DELETE FROM valuation.account_status WHERE account_id = $1", "DELETE FROM valuation.dirty WHERE account_id = $1",
			"DELETE FROM valuation.settings WHERE key = '" + chartStartKey + ":' || $1",
		} {
			if _, err := tx.Exec(ctx, q, string(account)); err != nil {
				return fmt.Errorf("%s: %w", q, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("forgetting account %s: %w", account, err)
	}
	return nil
}
