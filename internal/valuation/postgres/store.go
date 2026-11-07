package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

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

// NewStore returns the store over pool. Run Migrate first. Its River
// client, which only enqueues, installs plugins, such as the core's job
// tracing.
func NewStore(pool *pgxpool.Pool, plugins ...rivertype.Plugin) (*Store, error) {
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{Schema: RiverSchema, Plugins: plugins})
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	return &Store{pool: pool, insert: c}, nil
}

// Prices returns the prices of instrument dated on or before through from
// its one source (SourceOf), in date order: never another source's, which
// only cross-check them (D-94, valuation.PricesBySource). None when no
// source has a price of it.
func (s *Store) Prices(ctx context.Context, instrument string, through kernel.Date) ([]valuation.Price, error) {
	src, ok, err := s.SourceOf(ctx, valuation.PriceSeries(instrument))
	if err != nil || !ok {
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
	return valuation.PricesBySource(out, src.Source).Own, nil
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

// FXRates returns the rates of currency dated on or before through from
// its one source (SourceOf), in date order: never another source's (D-94).
// None when no source has a rate of it.
func (s *Store) FXRates(ctx context.Context, currency kernel.Currency, through kernel.Date) ([]valuation.FXRate, error) {
	src, ok, err := s.SourceOf(ctx, valuation.FXSeries(currency))
	if err != nil || !ok {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT currency, base, date::text, rate::text, source FROM valuation.fx_rates
		WHERE currency = $1 AND date <= $2 ORDER BY date, base, source`, currency.String(), through.String())
	if err != nil {
		return nil, fmt.Errorf("loading FX rates of %s: %w", currency, err)
	}
	out, err := pgx.CollectRows(rows, scanRate)
	if err != nil {
		return nil, fmt.Errorf("loading FX rates of %s: %w", currency, err)
	}
	return valuation.FXRatesBySource(out, src.Source).Own, nil
}

// scanRate reads one FX rate row, with the base it names (D-180).
func scanRate(row pgx.CollectableRow) (valuation.FXRate, error) {
	var currency, base, day, rate, source string
	if err := row.Scan(&currency, &base, &day, &rate, &source); err != nil {
		return valuation.FXRate{}, fmt.Errorf("reading an FX rate: %w", err)
	}
	c, err1 := kernel.ParseCurrency(currency)
	b, err4 := kernel.ParseCurrency(base)
	d, err2 := kernel.ParseDate(day)
	r, err3 := kernel.ParseDecimal(rate)
	if err := errors.Join(err1, err2, err3, err4); err != nil {
		return valuation.FXRate{}, fmt.Errorf("%w: FX rate: %w", ErrCorrupt, err)
	}
	return valuation.FXRate{Currency: c, Base: b, Date: d, Rate: r, Source: source}, nil
}

// IndexReadings returns the readings of each index kind from its one
// source (SourceOf), in kind, year and month order: never another
// source's (D-94).
func (s *Store) IndexReadings(ctx context.Context) ([]valuation.IndexReading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT kind, year, month, value::text, source FROM valuation.index_readings ORDER BY kind, year, month, source`)
	if err != nil {
		return nil, fmt.Errorf("loading index readings: %w", err)
	}
	out, err := pgx.CollectRows(rows, scanReading)
	if err != nil {
		return nil, fmt.Errorf("loading index readings: %w", err)
	}
	var own []valuation.IndexReading
	for _, kind := range []valuation.IndexKind{valuation.CPI, valuation.ReferenceRate} {
		src, ok, err := s.SourceOf(ctx, valuation.IndexSeries(kind))
		if err != nil {
			return nil, err
		}
		if ok {
			ofKind := slices.DeleteFunc(slices.Clone(out), func(r valuation.IndexReading) bool { return r.Kind != kind })
			own = append(own, valuation.IndexReadingsBySource(ofKind, src.Source).Own...)
		}
	}
	slices.SortStableFunc(own, func(a, b valuation.IndexReading) int {
		return cmp.Or(cmp.Compare(a.Kind.String(), b.Kind.String()), cmp.Compare(a.Year, b.Year), cmp.Compare(a.Month, b.Month))
	})
	return own, nil
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

// BondTerms returns the series terms of the bond with isin, with their
// attribution when they were read from published data (D-97), and false
// when none are stored.
func (s *Store) BondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	var (
		series, first, margin, fee string
		att                        attributionRow
	)
	err := s.pool.QueryRow(ctx, `
		SELECT series, first_period_rate::text, margin::text, early_redemption_fee::text,
		    source_publisher, source_dataset, source_license, source_url, source_made::text, source_fetched,
		    source_processing
		FROM valuation.bond_terms WHERE isin = $1`, isin.String()).Scan(&series, &first, &margin, &fee,
		&att.publisher, &att.dataset, &att.license, &att.url, &att.made, &att.fetched, &att.processing)
	if errors.Is(err, pgx.ErrNoRows) {
		return pl.Terms{}, false, nil
	}
	if err != nil {
		return pl.Terms{}, false, fmt.Errorf("loading bond terms of %s: %w", isin, err)
	}
	sr, err1 := pl.ParseSeries(series)
	f, err2 := kernel.ParseDecimal(first)
	m, err3 := kernel.ParseDecimal(margin)
	b, err4 := parseMoney(fee, pl.Nominal().Currency().String())
	var (
		a    *pl.Attribution
		err5 error
	)
	if att.publisher != nil {
		a, err5 = att.attribution()
	}
	if err := errors.Join(err1, err2, err3, err4, err5); err != nil {
		return pl.Terms{}, false, fmt.Errorf("%w: bond terms of %s: %w", ErrCorrupt, isin, err)
	}
	return pl.Terms{Series: sr, FirstPeriodRate: f, Margin: m, EarlyRedemptionFee: b, Attribution: a}, true, nil
}

// attribution returns the attribution the row holds, whose publisher is
// set.
func (r *attributionRow) attribution() (*pl.Attribution, error) {
	if r.publisher == nil || r.dataset == nil || r.license == nil || r.url == nil || r.made == nil || r.fetched == nil || r.processing == nil {
		return nil, fmt.Errorf("%w: a part of the attribution is missing", pl.ErrInvalidAttribution)
	}
	made, err := kernel.ParseDate(*r.made)
	if err != nil {
		return nil, fmt.Errorf("attribution date: %w", err)
	}
	a := &pl.Attribution{
		Publisher: *r.publisher, Dataset: *r.dataset, License: *r.license, URL: *r.url, Made: made,
		Fetched: r.fetched.UTC(), Processing: *r.processing,
	}
	if err := a.Validate(); err != nil {
		return nil, fmt.Errorf("attribution: %w", err)
	}
	return a, nil
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
    (account_id, isin, symbol, date, basis, quantity, value, value_currency, reporting, reporting_currency, source, price_date,
     rate_date, missing, detail, net, liquidation, net_missing, liquidation_missing, tax_cost, sale_cost)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
ON CONFLICT (account_id, isin, symbol, date, basis) DO UPDATE SET
    quantity = EXCLUDED.quantity, value = EXCLUDED.value, value_currency = EXCLUDED.value_currency,
    reporting = EXCLUDED.reporting, reporting_currency = EXCLUDED.reporting_currency, source = EXCLUDED.source, price_date = EXCLUDED.price_date,
    rate_date = EXCLUDED.rate_date, missing = EXCLUDED.missing, detail = EXCLUDED.detail,
    net = EXCLUDED.net, liquidation = EXCLUDED.liquidation, net_missing = EXCLUDED.net_missing,
    liquidation_missing = EXCLUDED.liquidation_missing, tax_cost = EXCLUDED.tax_cost, sale_cost = EXCLUDED.sale_cost`

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
