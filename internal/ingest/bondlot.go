package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidBondLot reports a bond lot that cannot be booked: no purchase
// date, a purchase date after today, a negative margin, or anything
// pl.ValueLot refuses for a lot bought for cash on that date (an unknown
// bond type, a purchase date that doesn't match the series, a quantity that
// is not a positive whole number, a negative rate or fee).
var ErrInvalidBondLot = errors.New("invalid bond lot")

// BondLot is a lot of Polish retail treasury bonds the owner entered by hand
// (D-9): its series, how many bonds and when they were bought, plus the
// series terms from the issue letter, which valuing the lot needs. The bonds
// were bought for cash at the nominal (pl.Nominal), so the lot cost
// Quantity times 100 PLN.
type BondLot struct {
	// Terms are the series and its terms.
	Terms     pl.Terms
	Quantity  kernel.Quantity
	Purchased kernel.Date
}

// Series returns the lot's bond series.
func (l *BondLot) Series() pl.Series { return l.Terms.Series }

// Cost returns what the lot cost: the nominal times the quantity, in PLN.
func (l *BondLot) Cost() kernel.Money { return pl.Nominal().Mul(l.Quantity.Decimal()) }

// Validate reports whether l can be booked as of today: it was bought on a
// date no later than today, its margin is not negative, and pl.ValueLot
// values it on its purchase date. Otherwise it returns an error wrapping
// ErrInvalidBondLot that names the rule broken; the error carries the
// series and dates, never an amount.
func (l *BondLot) Validate(today kernel.Date) error {
	switch {
	case l.Purchased.IsZero():
		return fmt.Errorf("%w: %s has no purchase date", ErrInvalidBondLot, l.Series())
	case l.Purchased.After(today):
		return fmt.Errorf("%w: %s bought on %s, after today %s", ErrInvalidBondLot, l.Series(), l.Purchased, today)
	case l.Terms.Margin.Sign() < 0:
		return fmt.Errorf("%w: %s has a negative margin", ErrInvalidBondLot, l.Series())
	}
	p := pl.Purchase{Date: l.Purchased, Price: pl.Nominal()}
	if _, err := pl.ValueLot(&l.Terms, nil, l.Quantity, p, pl.Taxable, l.Purchased); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBondLot, err)
	}
	return nil
}

// BondTerms are a treasury bond series' terms as a connector delivers them:
// the instrument they belong to and the terms. Recording them makes the
// core value holdings of the instrument as bonds (valuation's
// RecordBondTerms).
type BondTerms struct {
	Instrument ledger.Instrument
	Terms      pl.Terms
}
