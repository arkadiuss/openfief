package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
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
	pack  jurisdiction.Pack
	today func() kernel.Date
}

// NewRevaluer returns a Revaluer that reads the log from log, inputs and
// valuations from store, computes each holding's three values (D-77) with
// pack, and values through today().
func NewRevaluer(log TransactionLog, store Store, pack jurisdiction.Pack, today func() kernel.Date) *Revaluer {
	return &Revaluer{log: log, store: store, pack: pack, today: today}
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

// Revalue rebuilds account's valuations from the date from to today: for
// each day it derives the holdings from the log and values each one, at
// market price converted to PLN, or, for a bond with stored series terms,
// at its lots' statutory value, with its net and liquidation values in the
// account's tax wrapper under its fee rule and the owner's settings. It
// then replaces the stored valuations dated on or after from with the new
// rows and records the valued-through date (valuation.ValuedThrough). A
// missing price, FX rate or index reading is stored as a row with the gap,
// not an error. The rows depend only on the log, the stored inputs and
// today, so running it again gives the same result.
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
	if len(log) == 0 {
		return nil, nil
	}
	terms, err := in.accountTerms(ctx, r.pack)
	if err != nil {
		return nil, err
	}
	var rows []valuation.HoldingValuation
	for day := start; !day.After(today); day = day.AddDays(1) {
		holdings, err := ledger.HoldingsOn(log, account, day)
		if err != nil {
			return nil, fmt.Errorf("revaluing %s on %s: %w", account, day, err)
		}
		for i := range holdings {
			row, err := r.value(ctx, in, log, &holdings[i], terms)
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
func (r *Revaluer) value(ctx context.Context, in *inputCache, log []ledger.VersionedTransaction, h *ledger.Holding, terms *valuation.AccountTerms) (valuation.HoldingValuation, error) {
	bondTerms, isBond, err := in.bondTerms(ctx, h.Instrument.ISIN)
	if err != nil {
		return valuation.HoldingValuation{}, err
	}
	lots, err := ledger.LotsOn(log, h.Account, h.Instrument, h.Date)
	if _, oversold := errors.AsType[*ledger.OversellError](err); oversold && !isBond {
		return r.valueMarket(ctx, in, h, &valuation.MarketHolding{LotsUnknown: true, Terms: terms})
	}
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("lots of %s: %w", h.Instrument, err)
	}
	if !isBond {
		return r.valueMarket(ctx, in, h, &valuation.MarketHolding{Lots: lots, Terms: terms})
	}
	readings, err := in.indexReadings(ctx)
	if err != nil {
		return valuation.HoldingValuation{}, err
	}
	row, err := valuation.ValueBondHolding(h, &valuation.BondHolding{Lots: lots, Terms: &bondTerms, Readings: readings, Account: terms})
	if err != nil {
		return valuation.HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, err)
	}
	return row, nil
}

// valueMarket values h at market with m's lots and terms, loading the
// prices of its instrument and the FX rates of every currency they, its
// lots' costs and the fee rule's minimum are in.
func (r *Revaluer) valueMarket(ctx context.Context, in *inputCache, h *ledger.Holding, m *valuation.MarketHolding) (valuation.HoldingValuation, error) {
	prices, err := in.prices(ctx, h.Instrument.String())
	if err != nil {
		return valuation.HoldingValuation{}, err
	}
	currencies := []kernel.Currency{m.Terms.Fee.Minimum.Currency()}
	for i := range prices {
		currencies = append(currencies, prices[i].Close.Currency())
	}
	for i := range m.Lots {
		currencies = append(currencies, m.Lots[i].Cost.Currency())
	}
	if m.Rates, err = in.rates(ctx, currencies); err != nil {
		return valuation.HoldingValuation{}, err
	}
	m.Prices = prices
	row, err := valuation.ValueMarketHolding(h, m)
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
	store      Inputs
	account    ledger.AccountID
	through    kernel.Date
	terms      map[ledger.ISIN]*pl.Terms // nil value: not a bond
	pricesOf   map[string][]valuation.Price
	ratesOf    map[kernel.Currency][]valuation.FXRate
	readings   []valuation.IndexReading
	haveIndexR bool
}

func newInputCache(store Inputs, account ledger.AccountID, through kernel.Date) *inputCache {
	return &inputCache{
		store: store, account: account, through: through,
		terms: map[ledger.ISIN]*pl.Terms{}, pricesOf: map[string][]valuation.Price{}, ratesOf: map[kernel.Currency][]valuation.FXRate{},
	}
}

// accountTerms loads the account's tax wrapper, the pack's default unless
// one was recorded, its fee rule and the owner's settings.
func (c *inputCache) accountTerms(ctx context.Context, pack jurisdiction.Pack) (*valuation.AccountTerms, error) {
	w, recorded, err := c.store.TaxWrapper(ctx, c.account)
	if err != nil {
		return nil, fmt.Errorf("loading the tax wrapper of %s: %w", c.account, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(pack)
	}
	fee, err := c.store.FeeRule(ctx, c.account)
	if err != nil {
		return nil, fmt.Errorf("loading the fee rule of %s: %w", c.account, err)
	}
	settings, err := c.store.ValueSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the owner's settings: %w", err)
	}
	return &valuation.AccountTerms{Pack: pack, Wrapper: w, Fee: fee, Settings: settings}, nil
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

// prices returns the prices of instrument.
func (c *inputCache) prices(ctx context.Context, instrument string) ([]valuation.Price, error) {
	prices, ok := c.pricesOf[instrument]
	if !ok {
		var err error
		if prices, err = c.store.Prices(ctx, instrument, c.through); err != nil {
			return nil, fmt.Errorf("loading prices of %s: %w", instrument, err)
		}
		c.pricesOf[instrument] = prices
	}
	return prices, nil
}

// rates returns the FX rates of currencies, each once; PLN and a zero
// currency need none.
func (c *inputCache) rates(ctx context.Context, currencies []kernel.Currency) ([]valuation.FXRate, error) {
	var out []valuation.FXRate
	seen := map[kernel.Currency]bool{}
	for _, cur := range currencies {
		if seen[cur] || cur.IsZero() || cur.Equal(valuation.PLN()) {
			continue
		}
		seen[cur] = true
		got, ok := c.ratesOf[cur]
		if !ok {
			var err error
			if got, err = c.store.FXRates(ctx, cur, c.through); err != nil {
				return nil, fmt.Errorf("loading FX rates of %s: %w", cur, err)
			}
			c.ratesOf[cur] = got
		}
		out = append(out, got...)
	}
	return out, nil
}

// indexReadings returns every period index reading.
func (c *inputCache) indexReadings(ctx context.Context) ([]valuation.IndexReading, error) {
	if !c.haveIndexR {
		readings, err := c.store.IndexReadings(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading index readings: %w", err)
		}
		c.readings, c.haveIndexR = readings, true
	}
	return c.readings, nil
}
