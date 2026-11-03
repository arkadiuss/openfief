package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNoRate reports that no FX rate for a currency is dated on or before the
// date asked for.
var ErrNoRate = errors.New("no FX rate")

// ErrConflictingRates reports two different rates for one currency and date.
var ErrConflictingRates = errors.New("conflicting FX rates")

// ErrInvalidRate reports an FX rate that is not positive.
var ErrInvalidRate = errors.New("invalid FX rate")

// amountPlaces is the number of decimal places a converted amount is
// rounded to: the minor unit of the currencies OpenFief reports in, such
// as grosze or cents.
const amountPlaces = 2

// FXRate is the value in Base of one unit of Currency on Date, from Source,
// such as NBP table A in M0, whose base is PLN. Every stored rate names its
// base, never an assumed one (D-180). A provider that quotes per 100 units
// (NBP does for some currencies) is normalized to one unit by its adapter.
type FXRate struct {
	Currency kernel.Currency
	Base     kernel.Currency
	Date     kernel.Date
	Rate     kernel.Decimal
	Source   string
}

// RateOnOrBefore returns the rate for currency c into base that applies on
// cutoff: the latest such rate in rates dated on or before cutoff; rates
// into another base are ignored. A provider publishes only on
// business days, so a weekend or holiday takes the last published rate before
// it. For valuation on a date, cutoff is that date; Polish tax uses an earlier
// cutoff (pl.TaxRateCutoff). Two different rates for c on the chosen date are
// an error rather than a guess.
func RateOnOrBefore(rates []FXRate, c, base kernel.Currency, cutoff kernel.Date) (FXRate, error) {
	var best FXRate
	found := false
	for _, r := range rates {
		if !r.Currency.Equal(c) || !r.Base.Equal(base) || r.Date.After(cutoff) {
			continue
		}
		switch {
		case !found || r.Date.After(best.Date):
			best, found = r, true
		case r.Date.Equal(best.Date) && !r.Rate.Equal(best.Rate):
			return FXRate{}, fmt.Errorf("%w: %s on %s: %s and %s", ErrConflictingRates, c, r.Date, best.Rate, r.Rate)
		}
	}
	if !found {
		return FXRate{}, fmt.Errorf("%w: %s to %s on or before %s", ErrNoRate, c, base, cutoff)
	}
	return best, nil
}

// Convert converts m into r's base with rate r, rounding the result to the
// minor unit with mode: amount × rate, rounded once. r must be a rate for
// m's currency.
func Convert(m kernel.Money, r *FXRate, mode kernel.RoundingMode) (kernel.Money, error) {
	if !r.Currency.Equal(m.Currency()) {
		return kernel.Money{}, &kernel.CurrencyMismatchError{Left: m.Currency(), Right: r.Currency}
	}
	if r.Rate.Sign() <= 0 {
		return kernel.Money{}, fmt.Errorf("%w: %s %s on %s", ErrInvalidRate, r.Currency, r.Rate, r.Date)
	}
	return roundIn(m.Amount().Mul(r.Rate), r.Base, mode)
}

// ConvertAt converts m into currency to with the rate that applies on
// cutoff (RateOnOrBefore). An amount already in to needs no rate and is
// only rounded to the minor unit. The caller names to, from the
// jurisdiction pack, never an assumed currency (D-180).
func ConvertAt(m kernel.Money, to kernel.Currency, rates []FXRate, cutoff kernel.Date, mode kernel.RoundingMode) (kernel.Money, error) {
	if m.Currency().Equal(to) {
		return roundIn(m.Amount(), to, mode)
	}
	r, err := RateOnOrBefore(rates, m.Currency(), to, cutoff)
	if err != nil {
		return kernel.Money{}, err
	}
	return Convert(m, &r, mode)
}

// roundIn returns amount in currency c, rounded to the minor unit.
func roundIn(amount kernel.Decimal, c kernel.Currency, mode kernel.RoundingMode) (kernel.Money, error) {
	d, err := amount.Round(amountPlaces, mode)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("convert to %s: %w", c, err)
	}
	return kernel.NewMoney(d, c), nil
}
