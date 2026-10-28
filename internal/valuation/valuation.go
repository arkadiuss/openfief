package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrUnknownBasis reports a Basis outside the declared constants, or text
// that names none of them.
var ErrUnknownBasis = errors.New("unknown valuation basis")

// ErrBasisNotForBonds reports a basis that cannot value a retail treasury
// bond: such bonds are not traded, so they have no market price.
var ErrBasisNotForBonds = errors.New("valuation basis does not apply to retail treasury bonds")

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

// Basis says which value of a holding a valuation is.
type Basis int

const (
	// Market is quantity times an end-of-day price.
	Market Basis = iota + 1
	// GrossAccrued is a bond's nominal plus accrued interest, before any fee or tax.
	GrossAccrued
	// NetIfRedeemed is what redeeming a bond today would pay, after the
	// early-redemption fee and Belka tax (no tax inside IKE or IKZE).
	NetIfRedeemed
)

// DefaultBondBasis is the basis bonds count at in totals unless configured
// otherwise (D-25).
const DefaultBondBasis = NetIfRedeemed

// String returns the basis's name: market, gross_accrued or net_if_redeemed.
func (b Basis) String() string {
	switch b {
	case Market:
		return "market"
	case GrossAccrued:
		return "gross_accrued"
	case NetIfRedeemed:
		return "net_if_redeemed"
	}
	return fmt.Sprintf("Basis(%d)", int(b))
}

// ParseBasis parses a basis name as String writes it.
func ParseBasis(s string) (Basis, error) {
	for _, b := range []Basis{Market, GrossAccrued, NetIfRedeemed} {
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

// BondBasis returns the basis bonds count at in totals for a configured
// basis: the zero value (not configured) gives DefaultBondBasis, and Market
// is rejected because retail treasury bonds have no market price.
func BondBasis(configured Basis) (Basis, error) {
	switch configured {
	case 0:
		return DefaultBondBasis, nil
	case GrossAccrued, NetIfRedeemed:
		return configured, nil
	case Market:
		return 0, fmt.Errorf("%w: %s", ErrBasisNotForBonds, configured)
	}
	return 0, fmt.Errorf("%w: %d", ErrUnknownBasis, configured)
}

// BondValuation picks the value of a bond lot that counts in totals on the
// configured basis (see BondBasis): the gross accrued value or the
// net-if-redeemed value that pl.ValueLot computed for date on.
func BondValuation(instrument string, on kernel.Date, v *pl.LotValue, configured Basis, source string) (Valuation, error) {
	basis, err := BondBasis(configured)
	if err != nil {
		return Valuation{}, err
	}
	value := v.Net
	if basis == GrossAccrued {
		value = v.Gross
	}
	return Valuation{Instrument: instrument, Date: on, Value: value, Basis: basis, Source: source}, nil
}
