package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// AppendInstrumentNames stores names in their order (D-197): the owner's
// always, a connector's only while no connector's name of its ISIN is
// stored, so the first delivered name stays; a later one changes nothing.
func (s *Store) AppendInstrumentNames(ctx context.Context, names []ledger.InstrumentName) error {
	for _, n := range names {
		if _, err := s.db.Exec(ctx, `
			INSERT INTO ledger.instrument_names (isin, name, by_owner, source)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (isin) WHERE NOT by_owner DO NOTHING`,
			n.ISIN.String(), n.Name, n.ByOwner, n.Source); err != nil {
			return fmt.Errorf("storing a name of %s: %w", n.ISIN, err)
		}
	}
	return nil
}

// InstrumentNames returns every stored name, in the order stored.
func (s *Store) InstrumentNames(ctx context.Context) ([]ledger.InstrumentName, error) {
	rows, err := s.db.Query(ctx, `SELECT isin, name, by_owner, source FROM ledger.instrument_names ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("loading instrument names: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanName)
	if err != nil {
		return nil, fmt.Errorf("loading instrument names: %w", err)
	}
	return out, nil
}

// scanName reads one instrument name row.
func scanName(row pgx.CollectableRow) (ledger.InstrumentName, error) {
	var n ledger.InstrumentName
	var isin string
	if err := row.Scan(&isin, &n.Name, &n.ByOwner, &n.Source); err != nil {
		return ledger.InstrumentName{}, fmt.Errorf("reading an instrument name: %w", err)
	}
	var err error
	if n.ISIN, err = ledger.ParseISIN(isin); err != nil {
		return ledger.InstrumentName{}, fmt.Errorf("reading an instrument name: %w", err)
	}
	return n, nil
}
