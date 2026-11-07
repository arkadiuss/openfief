package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// RekeyInstrument moves the model portfolio's target of from, named
// "ISIN/symbol", to to, in one database transaction, when the registry's
// ISIN replaces a provisional one (D-214). When the model already targets
// to, from's weight joins to's, since both name the same series, so the
// weights still sum to 100. With no target of from, it changes nothing.
func (s *Store) RekeyInstrument(ctx context.Context, from, to ledger.Instrument) error {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			UPDATE portfolio.model_targets t SET weight = t.weight + f.weight
			FROM portfolio.model_targets f
			WHERE t.kind = 'instrument' AND t.name = $2 AND f.kind = 'instrument' AND f.name = $1`,
			from.String(), to.String()); err != nil {
			return fmt.Errorf("joining the target's weight: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			DELETE FROM portfolio.model_targets f WHERE f.kind = 'instrument' AND f.name = $1
			AND EXISTS (SELECT 1 FROM portfolio.model_targets t WHERE t.kind = 'instrument' AND t.name = $2)`,
			from.String(), to.String()); err != nil {
			return fmt.Errorf("dropping the joined target: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE portfolio.model_targets SET name = $2 WHERE kind = 'instrument' AND name = $1`,
			from.String(), to.String()); err != nil {
			return fmt.Errorf("renaming the target: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("re-keying the model's target %s to %s: %w", from, to, err)
	}
	return nil
}
