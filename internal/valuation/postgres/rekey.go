package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// rekeyed is one table's re-key: it moves the rows keyed by from to to,
// dropping first those whose key to already has, so to's own stay. The
// arguments are $1 from and $2 to, by ISIN or by instrument
// ("ISIN/symbol"), as the table is keyed.
type rekeyed struct {
	byInstrument bool
	drop, move   string
}

// rekeys returns the re-keys of every valuation table keyed by an
// instrument that RekeyInstrument moves.
func rekeys() []rekeyed {
	return []rekeyed{
		{
			drop: `DELETE FROM valuation.bond_terms f WHERE f.isin = $1
		AND EXISTS (SELECT 1 FROM valuation.bond_terms t WHERE t.isin = $2)`,
			move: `UPDATE valuation.bond_terms SET isin = $2 WHERE isin = $1`,
		},
		{
			drop: `DELETE FROM valuation.exemption_classes f WHERE f.instrument = $1
		AND EXISTS (SELECT 1 FROM valuation.exemption_classes t WHERE t.instrument = $2)`,
			move: `UPDATE valuation.exemption_classes SET instrument = $2 WHERE instrument = $1`,
		},
		{
			byInstrument: true, drop: `DELETE FROM valuation.prices f WHERE f.instrument = $1
		AND EXISTS (SELECT 1 FROM valuation.prices t WHERE t.instrument = $2 AND t.date = f.date AND t.source = f.source)`,
			move: `UPDATE valuation.prices SET instrument = $2 WHERE instrument = $1`,
		},
		{
			byInstrument: true, drop: `DELETE FROM valuation.cross_check_prices f WHERE f.instrument = $1
		AND EXISTS (SELECT 1 FROM valuation.cross_check_prices t WHERE t.instrument = $2 AND t.date = f.date AND t.source = f.source)`,
			move: `UPDATE valuation.cross_check_prices SET instrument = $2 WHERE instrument = $1`,
		},
		{
			byInstrument: true, drop: `DELETE FROM valuation.price_check_thresholds f WHERE f.instrument = $1
		AND EXISTS (SELECT 1 FROM valuation.price_check_thresholds t WHERE t.instrument = $2)`,
			move: `UPDATE valuation.price_check_thresholds SET instrument = $2 WHERE instrument = $1`,
		},
		{
			byInstrument: true, drop: `DELETE FROM valuation.series_sources f WHERE f.kind = 'prices' AND f.key = $1
		AND EXISTS (SELECT 1 FROM valuation.series_sources t WHERE t.kind = 'prices' AND t.key = $2)`,
			move: `UPDATE valuation.series_sources SET key = $2 WHERE kind = 'prices' AND key = $1`,
		},
	}
}

// RekeyInstrument moves what the valuation schema keeps of from to to, in
// one database transaction, when the registry's ISIN replaces a provisional
// one (D-214): a treasury bond series' terms, with their source and
// attribution, the owner's exemption class (D-132), and its prices,
// cross-check prices, price check threshold and prices source, should it
// have any. Where to already has its own, to's stay and from's go. The
// stored valuations stay: the ledger's re-key enqueues the revaluation
// that replaces them from the first day re-keyed (D-36). The owner's
// acknowledgements of price checks are append-only and stay with from;
// treasury bonds have no price checks. With nothing under from, it changes
// nothing.
func (s *Store) RekeyInstrument(ctx context.Context, from, to ledger.Instrument) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, r := range rekeys() {
			f, t := from.ISIN.String(), to.ISIN.String()
			if r.byInstrument {
				f, t = from.String(), to.String()
			}
			if _, err := tx.Exec(ctx, r.drop, f, t); err != nil {
				return fmt.Errorf("%s: %w", r.drop, err)
			}
			if _, err := tx.Exec(ctx, r.move, f, t); err != nil {
				return fmt.Errorf("%s: %w", r.move, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("re-keying %s to %s: %w", from, to, err)
	}
	return nil
}
