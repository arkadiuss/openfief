package valuation

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of cross-check prices and price checks.
var (
	// ErrInvalidCrossCheckPrice reports a cross-check price without an
	// instrument, a date, a positive close or a source.
	ErrInvalidCrossCheckPrice = errors.New("invalid cross-check price")
	// ErrInvalidPriceCheckThreshold reports a price check threshold that is
	// not a fraction greater than 0 and less than 1.
	ErrInvalidPriceCheckThreshold = errors.New("invalid price check threshold")
	// ErrNotComparable reports a price and a cross-check price of different
	// instruments, dates or currencies, or a price that isn't positive.
	ErrNotComparable = errors.New("prices are not comparable")
)

// DifferencePlaces is how many decimal places a price check's relative
// difference keeps: 0.000001 is a ten-thousandth of a percent.
const DifferencePlaces = 6

// DifferenceRounding is how a price check's relative difference is rounded
// to DifferencePlaces. Whether a check is flagged never depends on it: the
// threshold is compared with the exact difference.
const DifferenceRounding = kernel.RoundHalfEven

// CrossCheckPrice is an institution's own end-of-day price of one unit of an
// instrument (glossary: Cross-check price), such as the mark price of a
// position in an IBKR Flex statement, in the currency the institution
// reports it in. D-22 makes such prices a cross-check of the market-data
// connector's prices: they are compared with them and never value a
// holding. Instrument is written like Price's (ledger.Instrument.String());
// Source names who delivered it, such as a connector instance.
type CrossCheckPrice struct {
	Instrument string
	Date       kernel.Date
	Close      kernel.Money
	Source     string
}

// Validate reports whether c names an instrument and a source, and has a
// date and a positive close.
func (c *CrossCheckPrice) Validate() error {
	switch {
	case c.Instrument == "":
		return fmt.Errorf("%w: no instrument", ErrInvalidCrossCheckPrice)
	case c.Date.IsZero():
		return fmt.Errorf("%w: %s has no date", ErrInvalidCrossCheckPrice, c.Instrument)
	case c.Close.Sign() <= 0:
		return fmt.Errorf("%w: %s on %s: close %s is not positive", ErrInvalidCrossCheckPrice, c.Instrument, c.Date, c.Close)
	case c.Source == "":
		return fmt.Errorf("%w: %s on %s has no source", ErrInvalidCrossCheckPrice, c.Instrument, c.Date)
	}
	return nil
}

// DefaultPriceCheckThreshold is the relative difference beyond which a price
// check is flagged unless the owner configures another: 0.01, one percent.
// It is a setting because no decision fixes it.
func DefaultPriceCheckThreshold() kernel.Decimal { return kernel.NewDecimal(1, -2) }

// CheckThreshold returns threshold when it is a fraction greater than 0 and
// less than 1, and ErrInvalidPriceCheckThreshold otherwise.
func CheckThreshold(threshold kernel.Decimal) (kernel.Decimal, error) {
	if threshold.Sign() <= 0 || threshold.Cmp(kernel.NewDecimal(1, 0)) >= 0 {
		return kernel.Decimal{}, fmt.Errorf("%w: %s", ErrInvalidPriceCheckThreshold, threshold)
	}
	return threshold, nil
}

// PriceCheckStatus is the result of comparing a price with a cross-check
// price.
type PriceCheckStatus int

const (
	// PriceMatched is a cross-check price within the threshold of the price.
	PriceMatched PriceCheckStatus = iota + 1
	// PriceFlagged is a cross-check price that differs from the price by
	// more than the threshold: the owner should look at which one is wrong.
	PriceFlagged
)

// String returns matched or flagged.
func (s PriceCheckStatus) String() string {
	switch s {
	case PriceMatched:
		return "matched"
	case PriceFlagged:
		return "flagged"
	}
	return fmt.Sprintf("PriceCheckStatus(%d)", int(s))
}

// PriceCheck is a price compared with a cross-check price of the same
// instrument, date and currency (glossary: Price check). Difference is the
// cross-check price's relative difference from the price, (cross-check -
// price) / price, rounded to DifferencePlaces with DifferenceRounding.
type PriceCheck struct {
	Price      Price
	CrossCheck CrossCheckPrice
	Difference kernel.Decimal
	Status     PriceCheckStatus
}

// CheckPrice compares p with c: it is flagged when |c - p| exceeds threshold
// times p, exactly, and matched otherwise, the threshold itself included.
// Prices of different instruments, dates or currencies, or a price that
// isn't positive, are ErrNotComparable; a threshold outside (0, 1) is
// ErrInvalidPriceCheckThreshold.
func CheckPrice(p *Price, c *CrossCheckPrice, threshold kernel.Decimal) (PriceCheck, error) {
	if _, err := CheckThreshold(threshold); err != nil {
		return PriceCheck{}, err
	}
	if p.Instrument != c.Instrument || !p.Date.Equal(c.Date) || p.Close.Sign() <= 0 {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s and %s on %s", ErrNotComparable, p.Instrument, p.Date, c.Instrument, c.Date)
	}
	diff, err := c.Close.Sub(p.Close)
	if err != nil {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s: %w", ErrNotComparable, p.Instrument, p.Date, err)
	}
	relative, err := diff.Amount().Quo(p.Close.Amount(), DifferencePlaces, DifferenceRounding)
	if err != nil {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s: %w", ErrNotComparable, p.Instrument, p.Date, err)
	}
	absolute := diff.Amount()
	if absolute.Sign() < 0 {
		absolute = absolute.Neg()
	}
	status := PriceMatched
	if absolute.Cmp(threshold.Mul(p.Close.Amount())) > 0 {
		status = PriceFlagged
	}
	return PriceCheck{Price: *p, CrossCheck: *c, Difference: relative, Status: status}, nil
}

// CheckPrices compares every cross-check price that has a price of the same
// instrument, date and currency with that price (D-22), in date, instrument
// and cross-check source order. A cross-check price with no such price is
// left out: there is nothing to compare it with. A threshold outside (0, 1)
// is ErrInvalidPriceCheckThreshold.
func CheckPrices(prices []Price, crossChecks []CrossCheckPrice, threshold kernel.Decimal) ([]PriceCheck, error) {
	if _, err := CheckThreshold(threshold); err != nil {
		return nil, err
	}
	// Prices are keyed by instrument and date, which a space separates: an
	// instrument is ISIN/symbol and a date YYYY-MM-DD, so neither has one.
	byKey := make(map[string][]*Price, len(prices))
	for i := range prices {
		k := prices[i].Instrument + " " + prices[i].Date.String()
		byKey[k] = append(byKey[k], &prices[i])
	}
	var out []PriceCheck
	for i := range crossChecks {
		c := &crossChecks[i]
		for _, p := range byKey[c.Instrument+" "+c.Date.String()] {
			if !p.Close.Currency().Equal(c.Close.Currency()) {
				continue
			}
			check, err := CheckPrice(p, c, threshold)
			if err != nil {
				continue // a price that isn't positive: nothing to compare with
			}
			out = append(out, check)
		}
	}
	slices.SortStableFunc(out, func(a, b PriceCheck) int {
		return cmp.Or(a.CrossCheck.Date.Compare(b.CrossCheck.Date),
			cmp.Compare(a.CrossCheck.Instrument, b.CrossCheck.Instrument),
			cmp.Compare(a.CrossCheck.Source, b.CrossCheck.Source))
	})
	return out, nil
}
