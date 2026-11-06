package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// PriceHistory returns every stored price of the instruments with ISIN isin,
// each listing written as ledger.Instrument.String() ("<ISIN>/<symbol>"),
// dated in the range from through to, both inclusive, from every source (D-94), in
// instrument, date and source order: the instrument page's price history
// with its sources (D-195). Which source values each listing is its
// series' (SourceOf).
func (s *Store) PriceHistory(ctx context.Context, isin string, from, to kernel.Date) ([]valuation.Price, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT instrument, date::text, close::text, currency, source FROM valuation.prices
		WHERE (instrument = $1 OR starts_with(instrument, $1 || '/')) AND date BETWEEN $2 AND $3
		ORDER BY instrument COLLATE "C", date, source COLLATE "C"`, isin, from.String(), to.String())
	if err != nil {
		return nil, fmt.Errorf("loading the price history of %s: %w", isin, err)
	}
	out, err := pgx.CollectRows(rows, scanPrice)
	if err != nil {
		return nil, fmt.Errorf("loading the price history of %s: %w", isin, err)
	}
	return out, nil
}

// LatestPrices returns the latest stored price of each listing of the
// instrument with ISIN isin from each source, in listing then source order:
// what the instrument page shows next to each listing's price source
// (D-195).
func (s *Store) LatestPrices(ctx context.Context, isin string) ([]valuation.Price, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT instrument, day, closing, currency, source FROM (
			SELECT DISTINCT ON (instrument, source) instrument, date::text AS day, close::text AS closing, currency, source
			FROM valuation.prices
			WHERE instrument = $1 OR starts_with(instrument, $1 || '/')
			ORDER BY instrument, source, date DESC) latest
		ORDER BY instrument COLLATE "C", source COLLATE "C"`, isin)
	if err != nil {
		return nil, fmt.Errorf("loading the latest prices of %s: %w", isin, err)
	}
	out, err := pgx.CollectRows(rows, scanPrice)
	if err != nil {
		return nil, fmt.Errorf("loading the latest prices of %s: %w", isin, err)
	}
	return out, nil
}
