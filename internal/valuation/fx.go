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

// plnPlaces is the number of decimal places of a PLN amount (grosze).
const plnPlaces = 2

// PLN returns the currency every FX rate converts to.
func PLN() kernel.Currency {
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		return kernel.Currency{}
	}
	return c
}

// FXRate is the value in PLN of one unit of Currency on Date, from Source,
// such as NBP table A in M0. A provider that quotes per 100 units (NBP does
// for some currencies) is normalized to one unit by its adapter.
type FXRate struct {
	Currency kernel.Currency
	Date     kernel.Date
	Rate     kernel.Decimal
	Source   string
}

// RateOnOrBefore returns the rate for currency c that applies on cutoff: the
// latest rate in rates dated on or before cutoff. A provider publishes only on
// business days, so a weekend or holiday takes the last published rate before
// it. For valuation on a date, cutoff is that date; Polish tax uses an earlier
// cutoff (pl.TaxRateCutoff). Two different rates for c on the chosen date are
// an error rather than a guess.
func RateOnOrBefore(rates []FXRate, c kernel.Currency, cutoff kernel.Date) (FXRate, error) {
	var best FXRate
	found := false
	for _, r := range rates {
		if !r.Currency.Equal(c) || r.Date.After(cutoff) {
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
		return FXRate{}, fmt.Errorf("%w: %s on or before %s", ErrNoRate, c, cutoff)
	}
	return best, nil
}

// ToPLN converts m to PLN with rate r, rounding the result to grosze with
// mode: amount × rate, rounded once. r must be a rate for m's currency.
func ToPLN(m kernel.Money, r FXRate, mode kernel.RoundingMode) (kernel.Money, error) {
	if !r.Currency.Equal(m.Currency()) {
		return kernel.Money{}, &kernel.CurrencyMismatchError{Left: m.Currency(), Right: r.Currency}
	}
	if r.Rate.Sign() <= 0 {
		return kernel.Money{}, fmt.Errorf("%w: %s %s on %s", ErrInvalidRate, r.Currency, r.Rate, r.Date)
	}
	return roundPLN(m.Amount().Mul(r.Rate), mode)
}

// ConvertToPLN converts m to PLN with the rate that applies on cutoff
// (RateOnOrBefore). An amount already in PLN needs no rate and is only
// rounded to grosze.
func ConvertToPLN(m kernel.Money, rates []FXRate, cutoff kernel.Date, mode kernel.RoundingMode) (kernel.Money, error) {
	if m.Currency().Equal(PLN()) {
		return roundPLN(m.Amount(), mode)
	}
	r, err := RateOnOrBefore(rates, m.Currency(), cutoff)
	if err != nil {
		return kernel.Money{}, err
	}
	return ToPLN(m, r, mode)
}

func roundPLN(amount kernel.Decimal, mode kernel.RoundingMode) (kernel.Money, error) {
	d, err := amount.Round(plnPlaces, mode)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("convert to PLN: %w", err)
	}
	return kernel.NewMoney(d, PLN()), nil
}
