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
	// instruments, a price dated after the cross-check price, or a price
	// that isn't positive.
	ErrNotComparable = errors.New("prices are not comparable")
	// ErrNotFlagged reports an acknowledgement of a price check that isn't
	// flagged: only a flagged check can be acknowledged (D-112).
	ErrNotFlagged = errors.New("the price check is not flagged")
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
// check is flagged unless the owner set another for its instrument: 0.01,
// one percent (D-110).
func DefaultPriceCheckThreshold() kernel.Decimal { return kernel.NewDecimal(1, -2) }

// CheckThreshold returns threshold when it is a fraction greater than 0 and
// less than 1, and ErrInvalidPriceCheckThreshold otherwise.
func CheckThreshold(threshold kernel.Decimal) (kernel.Decimal, error) {
	if threshold.Sign() <= 0 || threshold.Cmp(kernel.NewDecimal(1, 0)) >= 0 {
		return kernel.Decimal{}, fmt.Errorf("%w: %s", ErrInvalidPriceCheckThreshold, threshold)
	}
	return threshold, nil
}

// Thresholds are the price check thresholds: DefaultPriceCheckThreshold,
// and those the owner set for single instruments, such as a thinly traded
// Warsaw listing (D-110).
type Thresholds struct {
	// ByInstrument holds the owner's thresholds, keyed by instrument as
	// Price writes it.
	ByInstrument map[string]kernel.Decimal
}

// For returns the threshold of instrument: the owner's, or the default.
func (t Thresholds) For(instrument string) kernel.Decimal {
	if d, ok := t.ByInstrument[instrument]; ok {
		return d
	}
	return DefaultPriceCheckThreshold()
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
	// PriceCurrencyMismatch is a cross-check price in another currency than
	// the price: the pair is shown with this status rather than skipped,
	// because the market-data connector may price another listing than the
	// one held (D-111). It has no difference.
	PriceCurrencyMismatch
)

// String returns matched, flagged or currency_mismatch.
func (s PriceCheckStatus) String() string {
	switch s {
	case PriceMatched:
		return "matched"
	case PriceFlagged:
		return "flagged"
	case PriceCurrencyMismatch:
		return "currency_mismatch"
	}
	return fmt.Sprintf("PriceCheckStatus(%d)", int(s))
}

// PriceCheck is a cross-check price compared with the latest stored price
// of its instrument on or before its date (glossary: Price check, D-111).
// Difference is the cross-check price's relative difference from the
// price, (cross-check - price) / price, rounded to DifferencePlaces with
// DifferenceRounding; zero for PriceCurrencyMismatch. Threshold is the one
// it was flagged at, and Acknowledged tells whether the owner acknowledged
// it as flagged with these prices (D-112).
type PriceCheck struct {
	Price        Price
	CrossCheck   CrossCheckPrice
	Difference   kernel.Decimal
	Threshold    kernel.Decimal
	Status       PriceCheckStatus
	Acknowledged bool
}

// CheckPrice compares p with c: in another currency it is
// PriceCurrencyMismatch; otherwise it is flagged when |c - p| exceeds
// threshold times p, exactly, and matched otherwise, the threshold itself
// included. A price of another instrument, one dated after c, or one that
// isn't positive is ErrNotComparable; a threshold outside (0, 1) is
// ErrInvalidPriceCheckThreshold.
func CheckPrice(p *Price, c *CrossCheckPrice, threshold kernel.Decimal) (PriceCheck, error) {
	if _, err := CheckThreshold(threshold); err != nil {
		return PriceCheck{}, err
	}
	if p.Instrument != c.Instrument || p.Date.After(c.Date) || p.Close.Sign() <= 0 {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s and %s on %s", ErrNotComparable, p.Instrument, p.Date, c.Instrument, c.Date)
	}
	out := PriceCheck{Price: *p, CrossCheck: *c, Threshold: threshold, Status: PriceCurrencyMismatch}
	if !p.Close.Currency().Equal(c.Close.Currency()) {
		return out, nil
	}
	diff, err := c.Close.Sub(p.Close)
	if err != nil {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s: %w", ErrNotComparable, p.Instrument, p.Date, err)
	}
	if out.Difference, err = diff.Amount().Quo(p.Close.Amount(), DifferencePlaces, DifferenceRounding); err != nil {
		return PriceCheck{}, fmt.Errorf("%w: %s on %s: %w", ErrNotComparable, p.Instrument, p.Date, err)
	}
	absolute := diff.Amount()
	if absolute.Sign() < 0 {
		absolute = absolute.Neg()
	}
	out.Status = PriceMatched
	if absolute.Cmp(threshold.Mul(p.Close.Amount())) > 0 {
		out.Status = PriceFlagged
	}
	return out, nil
}

// CheckPrices compares every cross-check price with the latest of prices
// of its instrument dated on or before it (D-22, D-111), at the threshold
// thresholds give its instrument (D-110), and marks a flagged check that
// one of acks acknowledges (D-112). Prices hold one price per instrument
// and date, the winning one (D-94); of two, the first is taken. A
// cross-check price with no such price, or whose price isn't positive, is
// left out: there is nothing to compare it with. The checks are in date,
// instrument and cross-check source order. A threshold outside (0, 1) is
// ErrInvalidPriceCheckThreshold.
func CheckPrices(prices []Price, crossChecks []CrossCheckPrice, thresholds Thresholds, acks []PriceCheckAcknowledgement) ([]PriceCheck, error) {
	byInstrument := make(map[string][]*Price)
	for i := range prices {
		byInstrument[prices[i].Instrument] = append(byInstrument[prices[i].Instrument], &prices[i])
	}
	for _, ps := range byInstrument {
		slices.SortStableFunc(ps, func(a, b *Price) int { return a.Date.Compare(b.Date) })
	}
	var out []PriceCheck
	for i := range crossChecks {
		c := &crossChecks[i]
		p := latestOnOrBefore(byInstrument[c.Instrument], c.Date)
		if p == nil {
			continue
		}
		check, err := CheckPrice(p, c, thresholds.For(c.Instrument))
		switch {
		case errors.Is(err, ErrNotComparable):
			continue // a price that isn't positive: nothing to compare with
		case err != nil:
			return nil, err
		}
		check.Acknowledged = check.Status == PriceFlagged && slices.ContainsFunc(acks, func(a PriceCheckAcknowledgement) bool {
			return a.Acknowledges(&check)
		})
		out = append(out, check)
	}
	slices.SortStableFunc(out, func(a, b PriceCheck) int {
		return cmp.Or(a.CrossCheck.Date.Compare(b.CrossCheck.Date),
			cmp.Compare(a.CrossCheck.Instrument, b.CrossCheck.Instrument),
			cmp.Compare(a.CrossCheck.Source, b.CrossCheck.Source))
	})
	return out, nil
}

// latestOnOrBefore returns the latest of prices, sorted by date, dated on
// or before day, the first of two on one date; nil when there is none.
func latestOnOrBefore(prices []*Price, day kernel.Date) *Price {
	n, _ := slices.BinarySearchFunc(prices, day, func(p *Price, d kernel.Date) int {
		if p.Date.After(d) {
			return 1
		}
		return -1
	})
	if n == 0 {
		return nil
	}
	first := n - 1
	for first > 0 && prices[first-1].Date.Equal(prices[n-1].Date) {
		first--
	}
	return prices[first]
}

// PriceCheckAcknowledgement is the owner's acknowledgement of a flagged
// price check (D-112): the check's instrument, the cross-check price's
// date, source and close, and the date and close of the price it was
// compared with. It acknowledges the check only while both prices stay the
// same, so a cross-check price or a price delivered again with another
// close is flagged anew. Acknowledgements are recorded append-only.
type PriceCheckAcknowledgement struct {
	Instrument string
	Date       kernel.Date
	Source     string
	CrossCheck kernel.Money
	PriceDate  kernel.Date
	Price      kernel.Money
}

// Acknowledge returns the owner's acknowledgement of c, which must be
// flagged: any other status is ErrNotFlagged.
func Acknowledge(c *PriceCheck) (PriceCheckAcknowledgement, error) {
	if c.Status != PriceFlagged {
		return PriceCheckAcknowledgement{}, fmt.Errorf("%w: %s on %s from %s is %s", ErrNotFlagged,
			c.CrossCheck.Instrument, c.CrossCheck.Date, c.CrossCheck.Source, c.Status)
	}
	return PriceCheckAcknowledgement{
		Instrument: c.CrossCheck.Instrument, Date: c.CrossCheck.Date, Source: c.CrossCheck.Source,
		CrossCheck: c.CrossCheck.Close, PriceDate: c.Price.Date, Price: c.Price.Close,
	}, nil
}

// Acknowledges reports whether a acknowledges c: the same instrument,
// cross-check date, source and close, and the same price date and close.
func (a *PriceCheckAcknowledgement) Acknowledges(c *PriceCheck) bool {
	return a.Instrument == c.CrossCheck.Instrument && a.Date.Equal(c.CrossCheck.Date) && a.Source == c.CrossCheck.Source &&
		a.CrossCheck.Equal(c.CrossCheck.Close) && a.PriceDate.Equal(c.Price.Date) && a.Price.Equal(c.Price.Close)
}
