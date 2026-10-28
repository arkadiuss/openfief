package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Revaluer runs revaluations (D-36): it derives an account's holdings and
// lots for each day from the log and stores their valuations.
type Revaluer struct {
	log   TransactionLog
	store Store
	basis valuation.Basis
	today func() kernel.Date
}

// NewRevaluer returns a Revaluer that reads the log from log, inputs and
// valuations from store, values bonds on the configured basis (D-25; zero
// means the default, net if redeemed) and values through today().
func NewRevaluer(log TransactionLog, store Store, basis valuation.Basis, today func() kernel.Date) (*Revaluer, error) {
	b, err := valuation.BondBasis(basis)
	if err != nil {
		return nil, fmt.Errorf("revaluer: %w", err)
	}
	return &Revaluer{log: log, store: store, basis: b, today: today}, nil
}

// RevalueResult is what one revaluation did.
type RevalueResult struct {
	// Start is the first day valued: from, or the account's first
	// transaction date when that is later.
	Start kernel.Date
	// ValuedThrough is the account's valued-through date afterwards.
	ValuedThrough kernel.Date
	// Rows counts the stored valuations, and Missing those lacking an input.
	Rows    int
	Missing int
}

// Revalue rebuilds account's valuations from the date from to today: for each day it
// derives the holdings from the log and values each one, at market price
// converted to PLN, or, for a bond with stored series terms, at its lots'
// statutory value on the configured basis. It then replaces the stored
// valuations dated on or after from with the new rows and records the
// valued-through date (valuation.ValuedThrough). A missing price, FX rate or
// index reading is stored as a row with the gap, not an error. The rows
// depend only on the log, the stored inputs, the basis and today, so running
// it again gives the same result.
func (r *Revaluer) Revalue(ctx context.Context, account ledger.AccountID, from kernel.Date) (RevalueResult, error) {
	today := r.today()
	log, err := r.log.Log(ctx, account)
	if err != nil {
		return RevalueResult{}, fmt.Errorf("revaluing %s: loading the log: %w", account, err)
	}
	start := from
	if first, ok := firstDate(log); ok && first.After(start) {
		start = first
	}
	rows, err := r.rows(ctx, log, account, start, today)
	if err != nil {
		return RevalueResult{}, err
	}
	prior, hasPrior, err := r.store.ValuedThrough(ctx, account)
	if err != nil {
		return RevalueResult{}, fmt.Errorf("revaluing %s: %w", account, err)
	}
	through := valuation.ValuedThrough(prior, hasPrior, start, today, rows)
	if err := r.store.ReplaceValuations(ctx, account, from, rows, through); err != nil {
		return RevalueResult{}, fmt.Errorf("revaluing %s: storing: %w", account, err)
	}
	res := RevalueResult{Start: start, ValuedThrough: through, Rows: len(rows)}
	for i := range rows {
		if !rows[i].Valued() {
			res.Missing++
		}
	}
	return res, nil
}

// rows values account's holdings on each day from start to today, in date
// then instrument order; none when log is empty.
func (r *Revaluer) rows(ctx context.Context, log []ledger.VersionedTransaction, account ledger.AccountID, start, today kernel.Date) ([]valuation.HoldingValuation, error) {
	in := newInputCache(r.store, account, today)
	var rows []valuation.HoldingValuation
	for day := start; len(log) > 0 && !day.After(today); day = day.AddDays(1) {
		holdings, err := ledger.HoldingsOn(log, account, day)
		if err != nil {
			return nil, fmt.Errorf("revaluing %s on %s: %w", account, day, err)
		}
		for i := range holdings {
			row, err := r.value(ctx, in, log, &holdings[i])
			if err != nil {
				return nil, fmt.Errorf("revaluing %s on %s: %w", account, day, err)
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

// value values one holding: as a bond when its ISIN has series terms, at
// market otherwise.
func (r *Revaluer) value(ctx context.Context, in *inputCache, log []ledger.VersionedTransaction, h *ledger.Holding) (valuation.HoldingValuation, error) {
	terms, isBond, err := in.bondTerms(ctx, h.Instrument.ISIN)
	if err != nil {
		return valuation.HoldingValuation{}, err
	}
	if !isBond {
		prices, rates, loadErr := in.market(ctx, h.Instrument.String())
		if loadErr != nil {
			return valuation.HoldingValuation{}, loadErr
		}
		row, valueErr := valuation.ValueMarketHolding(h, prices, rates)
		if valueErr != nil {
			return valuation.HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, valueErr)
		}
		return row, nil
	}
	lots, err := ledger.LotsOn(log, h.Account, h.Instrument, h.Date)
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("lots of %s: %w", h.Instrument, err)
	}
	readings, wrapper, err := in.bondInputs(ctx)
	if err != nil {
		return valuation.HoldingValuation{}, err
	}
	row, err := valuation.ValueBondHolding(h, &valuation.BondHolding{
		Lots: lots, Terms: &terms, Readings: readings, Wrapper: wrapper, Basis: r.basis,
	})
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, err)
	}
	return row, nil
}

// firstDate returns the earliest date of a transaction version in log.
func firstDate(log []ledger.VersionedTransaction) (kernel.Date, bool) {
	var first kernel.Date
	for i := range log {
		if d := log[i].Transaction.Date; i == 0 || d.Before(first) {
			first = d
		}
	}
	return first, len(log) > 0
}

// inputCache loads each input of one revaluation from the store once.
type inputCache struct {
	store    Inputs
	account  ledger.AccountID
	through  kernel.Date
	terms    map[ledger.ISIN]*pl.Terms // nil value: not a bond
	prices   map[string][]valuation.Price
	rates    map[kernel.Currency][]valuation.FXRate
	readings []valuation.IndexReading
	wrapper  pl.TaxWrapper // zero until loaded
}

func newInputCache(store Inputs, account ledger.AccountID, through kernel.Date) *inputCache {
	return &inputCache{
		store: store, account: account, through: through,
		terms: map[ledger.ISIN]*pl.Terms{}, prices: map[string][]valuation.Price{}, rates: map[kernel.Currency][]valuation.FXRate{},
	}
}

// bondTerms returns the series terms of isin, and false when it is no bond.
func (c *inputCache) bondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	if t, ok := c.terms[isin]; ok {
		if t == nil {
			return pl.Terms{}, false, nil
		}
		return *t, true, nil
	}
	t, ok, err := c.store.BondTerms(ctx, isin)
	if err != nil {
		return pl.Terms{}, false, fmt.Errorf("loading bond terms of %s: %w", isin, err)
	}
	c.terms[isin] = nil
	if ok {
		c.terms[isin] = &t
	}
	return t, ok, nil
}

// market returns the prices of instrument and the FX rates of their
// currencies.
func (c *inputCache) market(ctx context.Context, instrument string) ([]valuation.Price, []valuation.FXRate, error) {
	prices, ok := c.prices[instrument]
	if !ok {
		var err error
		if prices, err = c.store.Prices(ctx, instrument, c.through); err != nil {
			return nil, nil, fmt.Errorf("loading prices of %s: %w", instrument, err)
		}
		c.prices[instrument] = prices
	}
	var rates []valuation.FXRate
	seen := map[kernel.Currency]bool{}
	for i := range prices {
		cur := prices[i].Close.Currency()
		if seen[cur] || cur.Equal(valuation.PLN()) {
			continue
		}
		seen[cur] = true
		got, ok := c.rates[cur]
		if !ok {
			var err error
			if got, err = c.store.FXRates(ctx, cur, c.through); err != nil {
				return nil, nil, fmt.Errorf("loading FX rates of %s: %w", cur, err)
			}
			c.rates[cur] = got
		}
		rates = append(rates, got...)
	}
	return prices, rates, nil
}

// bondInputs returns the index readings and the account's tax wrapper.
func (c *inputCache) bondInputs(ctx context.Context) ([]valuation.IndexReading, pl.TaxWrapper, error) {
	if c.wrapper == 0 {
		w, err := c.store.TaxWrapper(ctx, c.account)
		if err != nil {
			return nil, 0, fmt.Errorf("loading the tax wrapper of %s: %w", c.account, err)
		}
		readings, err := c.store.IndexReadings(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("loading index readings: %w", err)
		}
		c.wrapper, c.readings = w, readings
	}
	return c.readings, c.wrapper, nil
}
