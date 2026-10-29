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

// ErrInvalidInput reports a price, FX rate, index reading, tax wrapper, fee
// rule or setting that cannot be stored.
var ErrInvalidInput = errors.New("invalid valuation input")

// InputRecorder records the inputs of revaluations, each with the revaluations it
// causes (D-36): a Planner decides which, and a Recorder stores the input
// and enqueues them in one database transaction.
type InputRecorder struct {
	planner  *Planner
	recorder Recorder
	pack     jurisdiction.Pack
}

// NewInputRecorder returns the input use cases over planner and recorder,
// with the tax wrappers and rules of pack (D-70).
func NewInputRecorder(planner *Planner, recorder Recorder, pack jurisdiction.Pack) *InputRecorder {
	return &InputRecorder{planner: planner, recorder: recorder, pack: pack}
}

// RecordPrices stores end-of-day prices (D-21) and enqueues a revaluation of
// every account that held each instrument, from the price's date.
func (r *InputRecorder) RecordPrices(ctx context.Context, prices []valuation.Price) error {
	earliest := map[string]kernel.Date{}
	for i := range prices {
		p := &prices[i]
		if p.Instrument == "" || p.Close.Sign() <= 0 || p.Date.IsZero() {
			return fmt.Errorf("%w: price of %q on %s", ErrInvalidInput, p.Instrument, p.Date)
		}
		if d, ok := earliest[p.Instrument]; !ok || p.Date.Before(d) {
			earliest[p.Instrument] = p.Date
		}
	}
	var revals []valuation.Revaluation
	for instrument, from := range earliest {
		got, err := r.planner.ForInstrument(ctx, instrument, from)
		if err != nil {
			return fmt.Errorf("recording prices: %w", err)
		}
		revals = append(revals, got...)
	}
	if err := r.recorder.SavePrices(ctx, prices, valuation.Coalesce(revals)); err != nil {
		return fmt.Errorf("recording prices: %w", err)
	}
	return nil
}

// RecordFXRates stores FX rates to PLN (D-20) and enqueues a revaluation of
// every account from the earliest rate's date.
func (r *InputRecorder) RecordFXRates(ctx context.Context, rates []valuation.FXRate) error {
	var dates []kernel.Date
	for i := range rates {
		x := &rates[i]
		if x.Rate.Sign() <= 0 || x.Currency.Equal(valuation.PLN()) || x.Date.IsZero() {
			return fmt.Errorf("%w: %s rate on %s", ErrInvalidInput, x.Currency, x.Date)
		}
		dates = append(dates, x.Date)
	}
	revals, err := r.forAllFrom(ctx, dates)
	if err != nil {
		return fmt.Errorf("recording FX rates: %w", err)
	}
	if err := r.recorder.SaveFXRates(ctx, rates, revals); err != nil {
		return fmt.Errorf("recording FX rates: %w", err)
	}
	return nil
}

// RecordIndexReadings stores period index readings and enqueues a
// revaluation of every account from the first day of the earliest
// reading's month.
func (r *InputRecorder) RecordIndexReadings(ctx context.Context, readings []valuation.IndexReading) error {
	var dates []kernel.Date
	for i := range readings {
		x := &readings[i]
		switch x.Kind {
		case valuation.CPI, valuation.ReferenceRate:
		default:
			return fmt.Errorf("%w: %w: %s", ErrInvalidInput, valuation.ErrUnknownIndexKind, x.Kind)
		}
		first, err := kernel.NewDate(x.Year, x.Month, 1)
		if err != nil {
			return fmt.Errorf("%w: %s reading for %d-%02d: %w", ErrInvalidInput, x.Kind, x.Year, int(x.Month), err)
		}
		dates = append(dates, first)
	}
	revals, err := r.forAllFrom(ctx, dates)
	if err != nil {
		return fmt.Errorf("recording index readings: %w", err)
	}
	if err := r.recorder.SaveIndexReadings(ctx, readings, revals); err != nil {
		return fmt.Errorf("recording index readings: %w", err)
	}
	return nil
}

// RecordIndexLevels stores market index levels (D-51) for the model
// portfolio's index targets (D-53). No holding is valued from an index
// level, so they enqueue no revaluation. Every level must be valid, or
// nothing is stored.
func (r *InputRecorder) RecordIndexLevels(ctx context.Context, levels []valuation.IndexLevel) error {
	for i := range levels {
		if err := levels[i].Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
	}
	if err := r.recorder.SaveIndexLevels(ctx, levels); err != nil {
		return fmt.Errorf("recording index levels: %w", err)
	}
	return nil
}

// RecordBondTerms stores the series terms of the retail treasury bond with
// isin, which makes holdings of it valued as bonds, and enqueues a
// revaluation of every account that held it, from its first transaction.
// Polish retail treasury bonds have an ISIN per series (PL0000...), and the
// instrument's venue symbol is the series code, such as EDO0536.
func (r *InputRecorder) RecordBondTerms(ctx context.Context, isin ledger.ISIN, terms *pl.Terms) error {
	if isin.IsZero() {
		return fmt.Errorf("%w: bond terms without an ISIN", ErrInvalidInput)
	}
	if _, err := terms.Series.Type.Structure(); err != nil {
		return fmt.Errorf("%w: terms of %s: %w", ErrInvalidInput, isin, err)
	}
	revals, err := r.planner.ForISIN(ctx, isin)
	if err != nil {
		return fmt.Errorf("recording bond terms: %w", err)
	}
	if err := r.recorder.SaveBondTerms(ctx, isin, terms, revals); err != nil {
		return fmt.Errorf("recording bond terms of %s: %w", isin, err)
	}
	return nil
}

// RecordTaxWrapper records account's tax wrapper (D-76), one of the pack's,
// which sets the taxes and fees of its net and liquidation values, and
// enqueues a revaluation of the account from its first transaction.
func (r *InputRecorder) RecordTaxWrapper(ctx context.Context, account ledger.AccountID, wrapper jurisdiction.Wrapper) error {
	if !jurisdiction.HasWrapper(r.pack, wrapper) {
		return fmt.Errorf("%w: %w: %q", ErrInvalidInput, jurisdiction.ErrUnknownWrapper, wrapper)
	}
	revals, err := r.planner.FromStart(ctx, account)
	if err != nil {
		return fmt.Errorf("recording the tax wrapper of %s: %w", account, err)
	}
	if err := r.recorder.SaveTaxWrapper(ctx, account, wrapper, revals); err != nil {
		return fmt.Errorf("recording the tax wrapper of %s: %w", account, err)
	}
	return nil
}

// RecordFeeRule records account's fee rule (D-78), which sets the sale
// cost in its liquidation value, and enqueues a revaluation of the account
// from its first transaction. A rule breaking FeeRule.Validate is an error
// wrapping ErrInvalidInput.
func (r *InputRecorder) RecordFeeRule(ctx context.Context, account ledger.AccountID, rule *jurisdiction.FeeRule) error {
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	revals, err := r.planner.FromStart(ctx, account)
	if err != nil {
		return fmt.Errorf("recording the fee rule of %s: %w", account, err)
	}
	if err := r.recorder.SaveFeeRule(ctx, account, rule, revals); err != nil {
		return fmt.Errorf("recording the fee rule of %s: %w", account, err)
	}
	return nil
}

// RecordMarginalRate records the owner's marginal income tax rate (D-78),
// which an IKZE's liquidation value needs, and enqueues a revaluation of
// every account from its first transaction. A rate that is not above 0 and
// below 1 is an error wrapping ErrInvalidInput.
func (r *InputRecorder) RecordMarginalRate(ctx context.Context, rate kernel.Decimal) error {
	if err := jurisdiction.ValidateMarginalRate(rate); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	revals, err := r.planner.FromStart(ctx, "")
	if err != nil {
		return fmt.Errorf("recording the marginal rate: %w", err)
	}
	if err := r.recorder.SaveMarginalRate(ctx, rate, revals); err != nil {
		return fmt.Errorf("recording the marginal rate: %w", err)
	}
	return nil
}

// RecordNumberFormat records the owner's number format (D-113). No value
// depends on it, so nothing is revalued. A format breaking
// NumberFormat.Validate is an error wrapping ErrInvalidInput.
func (r *InputRecorder) RecordNumberFormat(ctx context.Context, format jurisdiction.NumberFormat) error {
	if err := format.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := r.recorder.SaveNumberFormat(ctx, format); err != nil {
		return fmt.Errorf("recording the number format: %w", err)
	}
	return nil
}

// RulesVersion returns the version of the rules the values are computed
// under: the pack's, such as "pl/1".
func (r *InputRecorder) RulesVersion() string { return r.pack.Version() }

// ApplyRules records the version of the jurisdiction pack's rules
// (RulesVersion) and, when it differs from the recorded one, revalues every
// account from its first transaction, since a change of rules can change
// any stored value (D-36). It reports whether it did. Run it at start-up.
func (r *InputRecorder) ApplyRules(ctx context.Context) (bool, error) {
	revals, err := r.planner.FromStart(ctx, "")
	if err != nil {
		return false, fmt.Errorf("applying the rules: %w", err)
	}
	changed, err := r.recorder.SaveRulesVersion(ctx, r.RulesVersion(), revals)
	if err != nil {
		return false, fmt.Errorf("applying the rules %s: %w", r.RulesVersion(), err)
	}
	return changed, nil
}

// EndOfDay enqueues the end-of-day revaluation of every account from its
// valued-through date (Planner.EndOfDay).
func (r *InputRecorder) EndOfDay(ctx context.Context, store ValuationStore) error {
	revals, err := r.planner.EndOfDay(ctx, store)
	if err != nil {
		return fmt.Errorf("end of day: %w", err)
	}
	if err := r.recorder.Enqueue(ctx, revals); err != nil {
		return fmt.Errorf("end of day: %w", err)
	}
	return nil
}

// forAllFrom returns a revaluation of every account from the earliest of
// dates, or none when dates is empty.
func (r *InputRecorder) forAllFrom(ctx context.Context, dates []kernel.Date) ([]valuation.Revaluation, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	from := dates[0]
	for _, d := range dates[1:] {
		if d.Before(from) {
			from = d
		}
	}
	return r.planner.ForAll(ctx, from)
}
