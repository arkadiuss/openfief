package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrUnknownBasis reports a Basis outside the declared constants, or text
// that names none of them.
var ErrUnknownBasis = errors.New("unknown valuation basis")

// ErrNoPrice reports a valuation asked for with a price of another instrument.
var ErrNoPrice = errors.New("no price for the instrument")

// Price is an end-of-day price of one unit of an instrument (D-21), in the
// currency it is quoted in, from a market-data source.
type Price struct {
	Instrument string
	Date       kernel.Date
	Close      kernel.Money
	Source     string
}

// Basis says how a holding's gross value (D-77) was computed.
type Basis int

const (
	// Market is quantity times an end-of-day price.
	Market Basis = iota + 1
	// GrossAccrued is a bond's nominal plus accrued interest, before any fee
	// or tax.
	GrossAccrued
)

// String returns the basis's name: market or gross_accrued.
func (b Basis) String() string {
	switch b {
	case Market:
		return "market"
	case GrossAccrued:
		return "gross_accrued"
	}
	return fmt.Sprintf("Basis(%d)", int(b))
}

// ParseBasis parses a basis name as String writes it.
func ParseBasis(s string) (Basis, error) {
	for _, b := range []Basis{Market, GrossAccrued} {
		if b.String() == s {
			return b, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownBasis, s)
}

// Valuation is the value of a holding of one instrument on a date, with the
// basis it was computed on and its source.
type Valuation struct {
	Instrument string
	Date       kernel.Date
	Value      kernel.Money
	Basis      Basis
	Source     string
}

// MarketValuation values quantity units of p's instrument at p's close, in
// the price's own currency and exactly; converting and rounding are separate
// steps.
func MarketValuation(instrument string, quantity kernel.Quantity, p *Price) (Valuation, error) {
	if p.Instrument != instrument {
		return Valuation{}, fmt.Errorf("%w: %q priced as %q", ErrNoPrice, instrument, p.Instrument)
	}
	return Valuation{
		Instrument: instrument, Date: p.Date, Value: p.Close.Mul(quantity.Decimal()),
		Basis: Market, Source: p.Source,
	}, nil
}
