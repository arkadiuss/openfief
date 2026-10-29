package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// ErrCorrupt reports a stored row that does not decode.
var ErrCorrupt = errors.New("corrupt valuation data")

// Store is the valuation context's store in PostgreSQL, in the valuation
// schema that Migrate creates: inputs, stored valuations and pending
// revaluations. Every write of an input enqueues its revaluations in the same
// database transaction, through River (D-36).
type Store struct {
	pool *pgxpool.Pool
	// insert is an insert-only River client: it enqueues jobs, never works
	// them.
	insert *river.Client[pgx.Tx]
}

var (
	_ app.Store    = (*Store)(nil)
	_ app.Recorder = (*Store)(nil)
)

// NewStore returns the store over pool. Run Migrate first.
func NewStore(pool *pgxpool.Pool) (*Store, error) {
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{Schema: RiverSchema})
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	return &Store{pool: pool, insert: c}, nil
}

// Prices returns the price of instrument on each date on or before
// through that has one, in date order: of every source's, the
// highest-ranked source's (D-94, valuation.RankedPrices).
func (s *Store) Prices(ctx context.Context, instrument string, through kernel.Date) ([]valuation.Price, error) {
	rank, err := s.rankOf(ctx, valuation.KindPrices)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT instrument, date::text, close::text, currency, source FROM valuation.prices
		WHERE instrument = $1 AND date <= $2 ORDER BY date, source`, instrument, through.String())
	if err != nil {
		return nil, fmt.Errorf("loading prices of %s: %w", instrument, err)
	}
	out, err := pgx.CollectRows(rows, scanPrice)
	if err != nil {
		return nil, fmt.Errorf("loading prices of %s: %w", instrument, err)
	}
	return valuation.RankedPrices(out, rank).Winners, nil
}

// scanPrice reads one price row.
func scanPrice(row pgx.CollectableRow) (valuation.Price, error) {
	var instrument, day, closing, currency, source string
	if err := row.Scan(&instrument, &day, &closing, &currency, &source); err != nil {
		return valuation.Price{}, fmt.Errorf("reading a price: %w", err)
	}
	d, err1 := kernel.ParseDate(day)
	m, err2 := parseMoney(closing, currency)
	if err := errors.Join(err1, err2); err != nil {
		return valuation.Price{}, fmt.Errorf("%w: price of %s: %w", ErrCorrupt, instrument, err)
	}
	return valuation.Price{Instrument: instrument, Date: d, Close: m, Source: source}, nil
}

// IndexLevels returns every level of the market index named index dated on
// or before through, in date order.
func (s *Store) IndexLevels(ctx context.Context, index string, through kernel.Date) ([]valuation.IndexLevel, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT index, date::text, level::text, source FROM valuation.index_levels
		WHERE index = $1 AND date <= $2 ORDER BY date`, index, through.String())
	if err != nil {
		return nil, fmt.Errorf("loading levels of %s: %w", index, err)
	}
	out, err := pgx.CollectRows(rows, scanIndexLevel)
	if err != nil {
		return nil, fmt.Errorf("loading levels of %s: %w", index, err)
	}
	return out, nil
}

// scanIndexLevel reads one index level row.
func scanIndexLevel(row pgx.CollectableRow) (valuation.IndexLevel, error) {
	var index, day, level, source string
	if err := row.Scan(&index, &day, &level, &source); err != nil {
		return valuation.IndexLevel{}, fmt.Errorf("reading an index level: %w", err)
	}
	d, err1 := kernel.ParseDate(day)
	v, err2 := kernel.ParseDecimal(level)
	if err := errors.Join(err1, err2); err != nil {
		return valuation.IndexLevel{}, fmt.Errorf("%w: level of %s: %w", ErrCorrupt, index, err)
	}
	return valuation.IndexLevel{Index: index, Date: d, Level: v, Source: source}, nil
}

// FXRates returns the rate of currency on each date on or before through
// that has one, in date order: of every source's, the highest-ranked
// source's (D-94).
func (s *Store) FXRates(ctx context.Context, currency kernel.Currency, through kernel.Date) ([]valuation.FXRate, error) {
	rank, err := s.rankOf(ctx, valuation.KindFXRates)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT currency, date::text, rate::text, source FROM valuation.fx_rates
		WHERE currency = $1 AND date <= $2 ORDER BY date, source`, currency.String(), through.String())
	if err != nil {
		return nil, fmt.Errorf("loading FX rates of %s: %w", currency, err)
	}
	out, err := pgx.CollectRows(rows, scanRate)
	if err != nil {
		return nil, fmt.Errorf("loading FX rates of %s: %w", currency, err)
	}
	return valuation.RankedFXRates(out, rank), nil
}

// scanRate reads one FX rate row.
func scanRate(row pgx.CollectableRow) (valuation.FXRate, error) {
	var currency, day, rate, source string
	if err := row.Scan(&currency, &day, &rate, &source); err != nil {
		return valuation.FXRate{}, fmt.Errorf("reading an FX rate: %w", err)
	}
	c, err1 := kernel.ParseCurrency(currency)
	d, err2 := kernel.ParseDate(day)
	r, err3 := kernel.ParseDecimal(rate)
	if err := errors.Join(err1, err2, err3); err != nil {
		return valuation.FXRate{}, fmt.Errorf("%w: FX rate: %w", ErrCorrupt, err)
	}
	return valuation.FXRate{Currency: c, Date: d, Rate: r, Source: source}, nil
}

// IndexReadings returns the reading of each index kind and month, in kind,
// year and month order: of every source's, the highest-ranked source's
// (D-94).
func (s *Store) IndexReadings(ctx context.Context) ([]valuation.IndexReading, error) {
	rank, err := s.rankOf(ctx, valuation.KindIndexReadings)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT kind, year, month, value::text, source FROM valuation.index_readings ORDER BY kind, year, month, source`)
	if err != nil {
		return nil, fmt.Errorf("loading index readings: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanReading)
	if err != nil {
		return nil, fmt.Errorf("loading index readings: %w", err)
	}
	return valuation.RankedIndexReadings(out, rank), nil
}

// scanReading reads one index reading row.
func scanReading(row pgx.CollectableRow) (valuation.IndexReading, error) {
	var kind, value, source string
	var year, month int32
	if err := row.Scan(&kind, &year, &month, &value, &source); err != nil {
		return valuation.IndexReading{}, fmt.Errorf("reading an index reading: %w", err)
	}
	k, err1 := valuation.ParseIndexKind(kind)
	v, err2 := kernel.ParseDecimal(value)
	if err := errors.Join(err1, err2); err != nil {
		return valuation.IndexReading{}, fmt.Errorf("%w: index reading: %w", ErrCorrupt, err)
	}
	return valuation.IndexReading{Kind: k, Year: int(year), Month: time.Month(month), Value: v, Source: source}, nil
}

// BondTerms returns the series terms of the bond with isin, and false when
// none are stored.
func (s *Store) BondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	var series, first, margin, fee string
	err := s.pool.QueryRow(ctx, `
		SELECT series, first_period_rate::text, margin::text, early_redemption_fee::text
		FROM valuation.bond_terms WHERE isin = $1`, isin.String()).Scan(&series, &first, &margin, &fee)
	if errors.Is(err, pgx.ErrNoRows) {
		return pl.Terms{}, false, nil
	}
	if err != nil {
		return pl.Terms{}, false, fmt.Errorf("loading bond terms of %s: %w", isin, err)
	}
	sr, err1 := pl.ParseSeries(series)
	f, err2 := kernel.ParseDecimal(first)
	m, err3 := kernel.ParseDecimal(margin)
	b, err4 := parseMoney(fee, "PLN")
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		return pl.Terms{}, false, fmt.Errorf("%w: bond terms of %s: %w", ErrCorrupt, isin, err)
	}
	return pl.Terms{Series: sr, FirstPeriodRate: f, Margin: m, EarlyRedemptionFee: b}, true, nil
}

// TaxWrapper returns account's tax wrapper, pl.Taxable unless one is stored.
func (s *Store) TaxWrapper(ctx context.Context, account ledger.AccountID) (pl.TaxWrapper, error) {
	var name string
	err := s.pool.QueryRow(ctx, `SELECT wrapper FROM valuation.tax_wrappers WHERE account_id = $1`, string(account)).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return pl.Taxable, nil
	}
	if err != nil {
		return 0, fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	return parseWrapper(name)
}

// ValuedThrough returns account's valued-through date, and false when the
// account was never revalued.
func (s *Store) ValuedThrough(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	var day string
	err := s.pool.QueryRow(ctx, `SELECT valued_through::text FROM valuation.account_status WHERE account_id = $1`, string(account)).Scan(&day)
	if errors.Is(err, pgx.ErrNoRows) {
		return kernel.Date{}, false, nil
	}
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("loading the valued-through date of %s: %w", account, err)
	}
	d, err := kernel.ParseDate(day)
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("%w: valued-through date of %s: %w", ErrCorrupt, account, err)
	}
	return d, true, nil
}

// insertValuation stores one row of valuation.holding_valuations.
const insertValuation = `
INSERT INTO valuation.holding_valuations
    (account_id, isin, symbol, date, basis, quantity, value, value_currency, pln, source, price_date, rate_date, missing, detail)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (account_id, isin, symbol, date, basis) DO UPDATE SET
    quantity = EXCLUDED.quantity, value = EXCLUDED.value, value_currency = EXCLUDED.value_currency,
    pln = EXCLUDED.pln, source = EXCLUDED.source, price_date = EXCLUDED.price_date,
    rate_date = EXCLUDED.rate_date, missing = EXCLUDED.missing, detail = EXCLUDED.detail`

// ReplaceValuations removes account's valuations dated on or after from,
// upserts rows, keyed by account, instrument, date and basis, and records
// through as the account's valued-through date, in one database transaction.
func (s *Store) ReplaceValuations(ctx context.Context, account ledger.AccountID, from kernel.Date, rows []valuation.HoldingValuation, through kernel.Date) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM valuation.holding_valuations WHERE account_id = $1 AND date >= $2`,
			string(account), from.String()); err != nil {
			return fmt.Errorf("removing valuations: %w", err)
		}
		batch := &pgx.Batch{}
		for i := range rows {
			batch.Queue(insertValuation, valuationArgs(&rows[i])...)
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return fmt.Errorf("storing valuations: %w", err)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO valuation.account_status (account_id, valued_through) VALUES ($1, $2)
			ON CONFLICT (account_id) DO UPDATE SET valued_through = EXCLUDED.valued_through, revalued_at = now()`,
			string(account), through.String())
		if err != nil {
			return fmt.Errorf("storing the valued-through date: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("replacing the valuations of %s: %w", account, err)
	}
	return nil
}

// valuationArgs returns the insertValuation arguments of r.
func valuationArgs(r *valuation.HoldingValuation) []any {
	var value, currency, pln, missing *string
	if r.Valued() {
		v, c, p := r.Value.Amount().String(), r.Value.Currency().String(), r.PLN.Amount().String()
		value, currency, pln = &v, &c, &p
	} else {
		m := r.Missing.String()
		missing = &m
	}
	return []any{
		string(r.Account), r.Instrument.ISIN.String(), r.Instrument.Symbol, r.Date.String(), r.Basis.String(),
		r.Quantity.String(), value, currency, pln, r.Source, optionalDate(r.PriceDate), optionalDate(r.RateDate),
		missing, r.Detail,
	}
}

// Valuations returns account's stored valuations dated on or after from
// and on or before to, in date, instrument and basis order.
func (s *Store) Valuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) ([]valuation.HoldingValuation, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT account_id, isin, symbol, date::text, basis, quantity::text, coalesce(value::text, ''),
		       coalesce(value_currency, ''), coalesce(pln::text, ''), source, coalesce(price_date::text, ''),
		       coalesce(rate_date::text, ''), coalesce(missing, ''), detail
		FROM valuation.holding_valuations
		WHERE account_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date, isin, symbol COLLATE "C", basis`, string(account), from.String(), to.String())
	if err != nil {
		return nil, fmt.Errorf("loading the valuations of %s: %w", account, err)
	}
	out, err := pgx.CollectRows(rows, scanValuation)
	if err != nil {
		return nil, fmt.Errorf("loading the valuations of %s: %w", account, err)
	}
	return out, nil
}

// scanValuation reads one row of Valuations.
func scanValuation(row pgx.CollectableRow) (valuation.HoldingValuation, error) {
	var account, isin, symbol, day, basis, quantity, value, currency, pln, source, priceDate, rateDate, missing, detail string
	if err := row.Scan(&account, &isin, &symbol, &day, &basis, &quantity, &value, &currency, &pln,
		&source, &priceDate, &rateDate, &missing, &detail); err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("reading a valuation: %w", err)
	}
	v, err := decodeValuation(&valuationText{
		isin: isin, symbol: symbol, day: day, basis: basis, quantity: quantity, value: value, currency: currency,
		pln: pln, priceDate: priceDate, rateDate: rateDate, missing: missing,
	})
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("%w: valuation of %s: %w", ErrCorrupt, account, err)
	}
	v.Account, v.Source, v.Detail = ledger.AccountID(account), source, detail
	return v, nil
}
