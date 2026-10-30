package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

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

// CrossCheckedPrices returns the prices to compare and what to compare
// them with (D-22, D-94, D-111): the winning price of every date of each
// instrument that has a cross-check price or more than one source's price
// on some date, in instrument then date order; and every cross-check price
// with every lower-ranked source's price as one
// (valuation.LowerRankedChecks), in instrument, date and source order.
func (s *Store) CrossCheckedPrices(ctx context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error) {
	rank, err := s.rankOf(ctx, valuation.KindPrices)
	if err != nil {
		return nil, nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.instrument, p.date::text, p.close::text, p.currency, p.source FROM valuation.prices p
		WHERE p.instrument IN (
			SELECT instrument FROM valuation.cross_check_prices
			UNION
			SELECT instrument FROM valuation.prices GROUP BY instrument, date HAVING count(*) > 1)
		ORDER BY p.instrument, p.date, p.source`)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	all, err := pgx.CollectRows(rows, scanPrice)
	if err != nil {
		return nil, nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	ranked := valuation.RankedPrices(all, rank)
	prices := ranked.Winners
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
	checks = append(checks, valuation.LowerRankedChecks(ranked.Lower)...)
	slices.SortStableFunc(checks, func(a, b valuation.CrossCheckPrice) int {
		return cmp.Or(cmp.Compare(a.Instrument, b.Instrument), a.Date.Compare(b.Date), cmp.Compare(a.Source, b.Source))
	})
	return prices, checks, nil
}

// PriceCheckThresholds returns the owner's threshold of each instrument
// that has one (D-110).
func (s *Store) PriceCheckThresholds(ctx context.Context) (valuation.Thresholds, error) {
	rows, err := s.pool.Query(ctx, `SELECT instrument, threshold::text FROM valuation.price_check_thresholds`)
	if err != nil {
		return valuation.Thresholds{}, fmt.Errorf("loading price check thresholds: %w", err)
	}
	out := valuation.Thresholds{ByInstrument: map[string]kernel.Decimal{}}
	var instrument, threshold string
	_, err = pgx.ForEachRow(rows, []any{&instrument, &threshold}, func() error {
		d, parseErr := kernel.ParseDecimal(threshold)
		if parseErr != nil {
			return fmt.Errorf("%w: price check threshold of %s: %w", ErrCorrupt, instrument, parseErr)
		}
		out.ByInstrument[instrument] = d
		return nil
	})
	if err != nil {
		return valuation.Thresholds{}, fmt.Errorf("loading price check thresholds: %w", err)
	}
	return out, nil
}

// SavePriceCheckThreshold sets instrument's threshold, or removes it when
// threshold is nil.
func (s *Store) SavePriceCheckThreshold(ctx context.Context, instrument string, threshold *kernel.Decimal) error {
	var err error
	if threshold == nil {
		_, err = s.pool.Exec(ctx, `DELETE FROM valuation.price_check_thresholds WHERE instrument = $1`, instrument)
	} else {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO valuation.price_check_thresholds (instrument, threshold) VALUES ($1, $2)
			ON CONFLICT (instrument) DO UPDATE SET threshold = EXCLUDED.threshold, set_at = now()`,
			instrument, threshold.String())
	}
	if err != nil {
		return fmt.Errorf("saving the price check threshold of %s: %w", instrument, err)
	}
	return nil
}

// PriceCheckAcknowledgements returns every acknowledgement recorded, the
// oldest first.
func (s *Store) PriceCheckAcknowledgements(ctx context.Context) ([]valuation.PriceCheckAcknowledgement, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT instrument, date::text, source, cross_check::text, cross_check_currency, price_date::text, price::text, price_currency
		FROM valuation.price_check_acknowledgements ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("loading price check acknowledgements: %w", err)
	}
	acks, err := pgx.CollectRows(rows, scanAcknowledgement)
	if err != nil {
		return nil, fmt.Errorf("loading price check acknowledgements: %w", err)
	}
	return acks, nil
}

// AppendPriceCheckAcknowledgement records an acknowledgement; the table
// refuses any change to one recorded (D-112).
func (s *Store) AppendPriceCheckAcknowledgement(ctx context.Context, ack *valuation.PriceCheckAcknowledgement) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO valuation.price_check_acknowledgements
			(instrument, date, source, cross_check, cross_check_currency, price_date, price, price_currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ack.Instrument, ack.Date.String(), ack.Source, ack.CrossCheck.Amount().String(), ack.CrossCheck.Currency().String(),
		ack.PriceDate.String(), ack.Price.Amount().String(), ack.Price.Currency().String())
	if err != nil {
		return fmt.Errorf("recording a price check acknowledgement: %w", err)
	}
	return nil
}

// scanAcknowledgement reads one acknowledgement row.
func scanAcknowledgement(row pgx.CollectableRow) (valuation.PriceCheckAcknowledgement, error) {
	var instrument, day, source, cross, crossCurrency, priceDay, price, priceCurrency string
	if err := row.Scan(&instrument, &day, &source, &cross, &crossCurrency, &priceDay, &price, &priceCurrency); err != nil {
		return valuation.PriceCheckAcknowledgement{}, fmt.Errorf("reading an acknowledgement: %w", err)
	}
	d, err1 := kernel.ParseDate(day)
	c, err2 := parseMoney(cross, crossCurrency)
	pd, err3 := kernel.ParseDate(priceDay)
	p, err4 := parseMoney(price, priceCurrency)
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		return valuation.PriceCheckAcknowledgement{}, fmt.Errorf("%w: acknowledgement of %s: %w", ErrCorrupt, instrument, err)
	}
	return valuation.PriceCheckAcknowledgement{
		Instrument: instrument, Date: d, Source: source, CrossCheck: c, PriceDate: pd, Price: p,
	}, nil
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
