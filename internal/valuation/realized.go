package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// RealizedGain is one sale's gain, or loss when negative, in the pack's
// currency for tax.
type RealizedGain struct {
	Date kernel.Date
	Gain kernel.Money
}

// RealizedGains returns each sale's gain for tax (D-128): its proceeds less
// its fees, each converted at the rate the pack takes for the sale's date,
// less the cost of the lots it consumed with their buy fees (D-150), each
// converted at the rate the pack takes for its open date (TaxCost); every
// conversion is rounded with ValueRounding. It returns MissingFXRate when a
// currency has no such rate. Only sales count: a dividend is a separate
// source of income that capital losses don't offset (D-150).
func (a *AccountTerms) RealizedGains(sales []ledger.Sale, rates []FXRate) ([]RealizedGain, MissingReason, error) {
	out := make([]RealizedGain, 0, len(sales))
	for i := range sales {
		s := &sales[i]
		gain, missing, err := a.saleGain(s, rates)
		if err != nil || missing != 0 {
			return nil, missing, err
		}
		out = append(out, RealizedGain{Date: s.Date, Gain: gain})
	}
	return out, 0, nil
}

// saleGain returns one sale's gain in the pack's currency.
func (a *AccountTerms) saleGain(s *ledger.Sale, rates []FXRate) (kernel.Money, MissingReason, error) {
	cost, missing, err := a.TaxCost(s.Consumed, false, rates)
	if err != nil || missing != 0 {
		return kernel.Money{}, missing, err
	}
	day := a.Pack.TaxRateDate(s.Date)
	// The proceeds count positive, the fees and the cost negative.
	amounts := append([]kernel.Money{s.Proceeds}, s.Fees...)
	gain := cost.Neg()
	for i, m := range amounts {
		pln, err := ConvertAt(m, a.Pack.ReportingCurrency(), rates, day, ValueRounding)
		if errors.Is(err, ErrNoRate) {
			return kernel.Money{}, MissingFXRate, nil
		}
		if err != nil {
			return kernel.Money{}, 0, fmt.Errorf("sale %s: %w", s.Key, err)
		}
		if i > 0 {
			pln = pln.Neg()
		}
		if gain, err = gain.Add(pln); err != nil {
			return kernel.Money{}, 0, fmt.Errorf("sale %s: %w", s.Key, err)
		}
	}
	return gain, 0, nil
}

// RealizedInYear returns the sum, in currency in, of gains dated in day's
// year on or before day: the tax year's realized gains and losses through
// day (D-128).
func RealizedInYear(gains []RealizedGain, day kernel.Date, in kernel.Currency) (kernel.Money, error) {
	sum := kernel.NewMoney(kernel.Decimal{}, in)
	for i := range gains {
		g := &gains[i]
		if g.Date.Year() != day.Year() || g.Date.After(day) {
			continue
		}
		var err error
		if sum, err = sum.Add(g.Gain); err != nil {
			return kernel.Money{}, fmt.Errorf("realized gains: %w", err)
		}
	}
	return sum, nil
}
