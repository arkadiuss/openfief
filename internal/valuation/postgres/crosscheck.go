package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

var _ app.CrossCheckStore = (*Store)(nil)

// SaveCrossCheckPrices upserts cross-check prices in one transaction, keyed
// by instrument, date and source. They cause no revaluation.
func (s *Store) SaveCrossCheckPrices(ctx context.Context, prices []valuation.CrossCheckPrice) error {
	err := s.inTx(ctx, nil, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for i := range prices {
			p := &prices[i]
			batch.Queue(`
				INSERT INTO valuation.cross_check_prices (instrument, date, close, currency, source) VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (instrument, date, source) DO UPDATE SET close = EXCLUDED.close, currency = EXCLUDED.currency`,
				p.Instrument, p.Date.String(), p.Close.Amount().String(), p.Close.Currency().String(), p.Source)
		}
		return tx.SendBatch(ctx, batch).Close()
	})
	if err != nil {
		return fmt.Errorf("saving cross-check prices: %w", err)
	}
	return nil
}

// CrossCheckedPrices returns every cross-check price, and every price of an
// instrument and date that has one, both in instrument then date order.
func (s *Store) CrossCheckedPrices(ctx context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.instrument, p.date::text, p.close::text, p.currency, p.source FROM valuation.prices p
		WHERE EXISTS (SELECT 1 FROM valuation.cross_check_prices c WHERE c.instrument = p.instrument AND c.date = p.date)
		ORDER BY p.instrument, p.date`)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	prices, err := pgx.CollectRows(rows, scanPrice)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	rows, err = s.pool.Query(ctx, `
		SELECT instrument, date::text, close::text, currency, source FROM valuation.cross_check_prices
		ORDER BY instrument, date, source`)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-check prices: %w", err)
	}
	checks, err := pgx.CollectRows(rows, scanCrossCheckPrice)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-check prices: %w", err)
	}
	return prices, checks, nil
}

// scanCrossCheckPrice reads one cross-check price row.
func scanCrossCheckPrice(row pgx.CollectableRow) (valuation.CrossCheckPrice, error) {
	var instrument, day, closing, currency, source string
	if err := row.Scan(&instrument, &day, &closing, &currency, &source); err != nil {
		return valuation.CrossCheckPrice{}, fmt.Errorf("reading a cross-check price: %w", err)
	}
	d, err1 := kernel.ParseDate(day)
	m, err2 := parseMoney(closing, currency)
	if err := errors.Join(err1, err2); err != nil {
		return valuation.CrossCheckPrice{}, fmt.Errorf("%w: cross-check price of %s: %w", ErrCorrupt, instrument, err)
	}
	return valuation.CrossCheckPrice{Instrument: instrument, Date: d, Close: m, Source: source}, nil
}
