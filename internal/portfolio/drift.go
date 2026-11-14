package portfolio

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNoPortfolioValue reports weights asked of a portfolio worth nothing: a
// share of zero is undefined.
var ErrNoPortfolioValue = errors.New("the portfolio has no value to weigh targets against")

// Figure is one target's standing against the real portfolio (D-204): its
// target weight, the weight the owner actually holds and the drift between
// them. OpenFief computes it; no model does.
type Figure struct {
	Target Target
	// Held is false for an index, which can't be held: it has no current
	// weight and no drift.
	Held bool
	// Current is the share of the portfolio's total value held in the
	// target's instrument, in percent, rounded to WeightPlaces. Zero when
	// the target is not held.
	Current kernel.Decimal
	// Drift is Current minus the target's weight, in percentage points:
	// positive when the portfolio holds more than the model asks. Zero when
	// the target is not held.
	Drift kernel.Decimal
}

// Drift is the figures of every target of a model portfolio and the share of
// the portfolio outside them.
type Drift struct {
	// Figures are the model's targets in its order.
	Figures []Figure
	// Other is the share of the portfolio in percent that no instrument
	// target holds: cash and every holding outside the model. It is what is
	// left of 100% after the held targets, so it may be negative only by
	// rounding.
	Other kernel.Decimal
}

// DriftOf returns the figures of model's targets for a portfolio worth
// total in the reporting currency, where values holds the value in that
// currency of each instrument held, by the name a target gives it ("ISIN/
// symbol"). Every weight is a percent of total, rounded half-even to
// WeightPlaces. An instrument target the portfolio holds none of has a
// current weight of zero; an index target is not held. A total that is not
// above zero is an error wrapping ErrNoPortfolioValue, and a value below
// zero ErrInvalidModel.
func DriftOf(model *Model, values map[string]kernel.Decimal, total kernel.Decimal) (Drift, error) {
	if total.Sign() <= 0 {
		return Drift{}, ErrNoPortfolioValue
	}
	out := Drift{Figures: make([]Figure, 0, len(model.Targets))}
	held := kernel.Decimal{}
	for i := range model.Targets {
		t := &model.Targets[i]
		f := Figure{Target: *t}
		if t.Kind == TargetInstrument {
			f.Held = true
			v := values[t.Name]
			if v.Sign() < 0 {
				return Drift{}, fmt.Errorf("%w: a negative value held in target %s", ErrInvalidModel, t.Key())
			}
			share, err := percentOf(v, total)
			if err != nil {
				return Drift{}, err
			}
			f.Current, f.Drift = share, share.Sub(t.Weight)
			held = held.Add(share)
		}
		out.Figures = append(out.Figures, f)
	}
	out.Other = hundredPercent().Sub(held)
	return out, nil
}

// percentOf returns v as a percent of total, rounded half-even to
// WeightPlaces.
func percentOf(v, total kernel.Decimal) (kernel.Decimal, error) {
	share, err := v.Mul(hundredPercent()).Quo(total, WeightPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("weighing a holding: %w", err)
	}
	return share, nil
}
