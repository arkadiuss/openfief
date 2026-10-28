package portfolio

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidModel reports a model portfolio or target that breaks one of
// its rules: a known kind, a name, an index's currency, a positive weight
// with at most WeightPlaces decimal places, no target twice, at most
// MaxTargets targets, and weights summing to 100% (D-52).
var ErrInvalidModel = errors.New("invalid model portfolio")

const (
	// MaxTargets is the most targets a model portfolio may have.
	MaxTargets = 50
	// WeightPlaces is the most decimal places a target's weight may have,
	// in percent: 0.01% is the finest weight.
	WeightPlaces = 2
	// maxNameLen bounds a target's name.
	maxNameLen = 64
)

// hundredPercent is what a model portfolio's weights sum to.
func hundredPercent() kernel.Decimal { return kernel.NewDecimal(indexBase, 0) }

// TargetKind says what a target is: an instrument, priced by the
// market-data connector, or an index, with end-of-day levels from it.
type TargetKind int

// The target kinds. The zero TargetKind is not a kind.
const (
	// TargetInstrument is something that can be held, such as an index ETF
	// or a stock, named by its ISIN and venue symbol.
	TargetInstrument TargetKind = iota + 1
	// TargetIndex is a market index, such as WIG20, which can't be held.
	TargetIndex
)

// String returns "instrument" or "index", or a placeholder for an invalid
// kind.
func (k TargetKind) String() string {
	switch k {
	case TargetInstrument:
		return "instrument"
	case TargetIndex:
		return "index"
	}
	return "TargetKind(" + strconv.Itoa(int(k)) + ")"
}

// ParseTargetKind parses "instrument" or "index".
func ParseTargetKind(s string) (TargetKind, error) {
	for _, k := range []TargetKind{TargetInstrument, TargetIndex} {
		if k.String() == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: unknown target kind %q", ErrInvalidModel, s)
}

// Target is one instrument or index in the model portfolio, with its
// weight (glossary: Target).
type Target struct {
	Kind TargetKind
	// Name identifies the target: an instrument's "ISIN/symbol", as the
	// ledger writes an instrument and prices are keyed, or an index's name
	// as the market-data connector names it, such as "WIG20".
	Name string
	// Currency is an index's currency, the one its levels are in. An
	// instrument's comes with its prices, so it has none here.
	Currency kernel.Currency
	// Weight is the target's share of the model in percent: 60 is 60%.
	Weight kernel.Decimal
}

// Key identifies the target within a model: its kind and name.
func (t *Target) Key() string { return t.Kind.String() + ":" + t.Name }

// Validate reports whether t has a known kind, a name of at most 64
// characters, a currency exactly when it is an index, and a weight above
// zero and at most 100% with at most WeightPlaces decimal places.
func (t *Target) Validate() error {
	switch t.Kind {
	case TargetInstrument, TargetIndex:
	default:
		return fmt.Errorf("%w: target %q: %s", ErrInvalidModel, t.Name, t.Kind)
	}
	if t.Name == "" || len(t.Name) > maxNameLen {
		return fmt.Errorf("%w: a target needs a name of 1 to %d characters", ErrInvalidModel, maxNameLen)
	}
	if (t.Kind == TargetIndex) == t.Currency.IsZero() {
		return fmt.Errorf("%w: target %s: an index needs a currency and an instrument has none", ErrInvalidModel, t.Key())
	}
	if t.Weight.Sign() <= 0 || t.Weight.Cmp(hundredPercent()) > 0 {
		return fmt.Errorf("%w: target %s: weight %s%% is not above 0%% and at most 100%%", ErrInvalidModel, t.Key(), t.Weight)
	}
	rounded, err := t.Weight.Round(WeightPlaces, kernel.RoundDown)
	if err != nil || !rounded.Equal(t.Weight) {
		return fmt.Errorf("%w: target %s: weight %s%% has more than %d decimal places", ErrInvalidModel, t.Key(), t.Weight, WeightPlaces)
	}
	return nil
}

// Model is the model portfolio (D-52): an optional weighted basket of
// targets for the whole portfolio. It is off by default, the zero Model.
type Model struct {
	// Enabled turns the model on: only then do the read API and the UI
	// show it.
	Enabled bool
	// Targets in the owner's order.
	Targets []Target
}

// Validate reports whether m's targets are valid, none twice, at most
// MaxTargets, and, when there are any, their weights sum to exactly 100%.
// A model without targets may be stored, but not turned on.
func (m *Model) Validate() error {
	if len(m.Targets) == 0 {
		if m.Enabled {
			return fmt.Errorf("%w: a model portfolio needs targets to be on", ErrInvalidModel)
		}
		return nil
	}
	if len(m.Targets) > MaxTargets {
		return fmt.Errorf("%w: %d targets, at most %d", ErrInvalidModel, len(m.Targets), MaxTargets)
	}
	seen := map[string]bool{}
	sum := kernel.Decimal{}
	for i := range m.Targets {
		t := &m.Targets[i]
		if err := t.Validate(); err != nil {
			return err
		}
		if seen[t.Key()] {
			return fmt.Errorf("%w: target %s twice", ErrInvalidModel, t.Key())
		}
		seen[t.Key()] = true
		sum = sum.Add(t.Weight)
	}
	if !sum.Equal(hundredPercent()) {
		return fmt.Errorf("%w: the weights sum to %s%%, not 100%%", ErrInvalidModel, sum)
	}
	return nil
}
