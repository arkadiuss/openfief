package pl

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNotPLN reports an amount handed to the Polish pack in a currency other
// than PLN: its tax rules count in PLN (D-20, D-62).
var ErrNotPLN = errors.New("the Polish pack values amounts in PLN")

// Pack is the Polish jurisdiction pack (D-70): the rules of this package
// behind jurisdiction.Pack. Its zero value has no OKI act in force; NewPack
// takes the act as this code knows it.
type Pack struct {
	// OKI is the act that taxes OKI's balance (D-76).
	OKI OKIAct
}

var _ jurisdiction.Pack = (*Pack)(nil)

// NewPack returns the Polish pack with the OKI act as CurrentOKIAct gives it.
func NewPack() *Pack { return &Pack{OKI: CurrentOKIAct()} }

// RulesVersion is the version of the pack's rules: raise it whenever a
// change to them changes a stored value, so every account is revalued.
const RulesVersion = "1"

// Version returns "pl/" and RulesVersion.
func (*Pack) Version() string { return "pl/" + RulesVersion }

// TaxRateDate returns TaxRateCutoff(event): the NBP rate of the last
// business day before the event converts it for tax (D-20).
func (*Pack) TaxRateDate(event kernel.Date) kernel.Date { return TaxRateCutoff(event) }

// ReportingCurrency returns PLN, the default reporting currency (D-62).
func (*Pack) ReportingCurrency() kernel.Currency { return pln() }

// NumberFormat returns the Polish default (D-113): thousands grouped with a
// space and a decimal comma, as in 5 000,00 PLN.
func (*Pack) NumberFormat() jurisdiction.NumberFormat {
	return jurisdiction.NumberFormat{Group: jurisdiction.SeparatorSpace, Decimal: jurisdiction.SeparatorComma}
}

// Wrappers returns the keys of TaxWrappers, taxable first (D-76).
func (*Pack) Wrappers() []jurisdiction.Wrapper {
	out := make([]jurisdiction.Wrapper, 0, len(TaxWrappers()))
	for _, w := range TaxWrappers() {
		out = append(out, jurisdiction.Wrapper(w.String()))
	}
	return out
}

// IKZEExitRate returns the 10% flat tax on a withdrawal from IKZE once its
// conditions are met, levied on the whole amount withdrawn (art. 30 ust. 1
// pkt 13 of the PIT act: "w wysokości 10% przychodu").
func IKZEExitRate() kernel.Decimal {
	const percent = 10
	return kernel.NewDecimal(percent, -2)
}

// ValueListed returns the three values (D-77) of a holding priced at
// market, from its gross value, its tax cost basis and the cost of selling
// it today, all in PLN:
//
//   - Taxable: net is gross less Belka tax on the gain over the cost basis;
//     liquidation is the proceeds after the sale cost less Belka tax on
//     their gain over the cost basis (the sale cost is deductible).
//   - IKE: net is gross, untaxed once the withdrawal conditions are met;
//     liquidation is a return (zwrot) before them, taxed like a taxable
//     sale on its gain.
//   - IKZE: net is gross less the 10% exit tax (IKZEExitRate); liquidation
//     is the proceeds less tax at the owner's marginal rate on all of them,
//     unknown while that rate is not set (D-78).
//   - OKI: net is gross and liquidation the proceeds: no Belka tax is due
//     inside (D-76).
//
// Taxes are rounded half up to whole grosze and are zero on a base that is
// not positive.
func (*Pack) ValueListed(in *jurisdiction.Listed, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	w, err := ParseTaxWrapper(string(in.Wrapper))
	if err != nil {
		return jurisdiction.Values{}, err
	}
	if err := allPLN(in.Gross, in.Cost, in.SaleCost); err != nil {
		return jurisdiction.Values{}, err
	}
	gross, cost := in.Gross.Amount(), in.Cost.Amount()
	proceeds := gross.Sub(in.SaleCost.Amount())
	switch w {
	case Taxable, IKE:
		liquidation, err := afterTax(proceeds, proceeds.Sub(cost), BelkaRate())
		if err != nil {
			return jurisdiction.Values{}, err
		}
		net := gross
		if w == Taxable {
			if net, err = afterTax(gross, gross.Sub(cost), BelkaRate()); err != nil {
				return jurisdiction.Values{}, err
			}
		}
		return values(gross, net, liquidation), nil
	case IKZE:
		return ikzeValues(gross, proceeds, s)
	case OKI:
		return values(gross, gross, proceeds), nil
	}
	return jurisdiction.Values{}, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, w)
}

// ValueAccrued returns the three values (D-77) of a bond lot, or of bond
// lots summed, from their gross accrued value, the income their exit taxes
// (interest plus purchase discount) and the early-redemption fee redeeming
// them today would charge in their wrapper (LotValue.Fee), all in PLN:
//
//   - Taxable: net is gross less Belka tax on the income, at maturity
//     without the fee; liquidation is gross less the fee less Belka tax on
//     the income minus the fee, what an early redemption pays (D-37).
//   - IKE: net is gross; liquidation is a return before the withdrawal
//     conditions, with no fee (LotValue.Fee is zero in IKE) and Belka tax
//     on the income.
//   - IKZE: net is gross less the 10% exit tax; liquidation is gross less
//     tax at the owner's marginal rate, unknown while it is not set (D-78).
//   - OKI: net is gross; liquidation is gross less the fee.
func (*Pack) ValueAccrued(in *jurisdiction.Accrued, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	w, err := ParseTaxWrapper(string(in.Wrapper))
	if err != nil {
		return jurisdiction.Values{}, err
	}
	if err := allPLN(in.Gross, in.Income, in.EarlyFee); err != nil {
		return jurisdiction.Values{}, err
	}
	gross, income, fee := in.Gross.Amount(), in.Income.Amount(), in.EarlyFee.Amount()
	afterFee := gross.Sub(fee)
	switch w {
	case Taxable, IKE:
		liquidation, err := afterTax(afterFee, income.Sub(fee), BelkaRate())
		if err != nil {
			return jurisdiction.Values{}, err
		}
		net := gross
		if w == Taxable {
			if net, err = afterTax(gross, income, BelkaRate()); err != nil {
				return jurisdiction.Values{}, err
			}
		}
		return values(gross, net, liquidation), nil
	case IKZE:
		return ikzeValues(gross, afterFee, s)
	case OKI:
		return values(gross, gross, afterFee), nil
	}
	return jurisdiction.Values{}, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, w)
}

// ValueCash returns the three values (D-77) of cash in PLN: its amount in
// every wrapper but IKZE, where a withdrawal is taxed like any other
// (IKZEExitRate at the planned exit, the owner's marginal rate before it).
func (*Pack) ValueCash(in *jurisdiction.Cash, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	w, err := ParseTaxWrapper(string(in.Wrapper))
	if err != nil {
		return jurisdiction.Values{}, err
	}
	if err := allPLN(in.Amount); err != nil {
		return jurisdiction.Values{}, err
	}
	amount := in.Amount.Amount()
	switch w {
	case Taxable, IKE, OKI:
		return values(amount, amount, amount), nil
	case IKZE:
		return ikzeValues(amount, amount, s)
	}
	return jurisdiction.Values{}, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, w)
}

// ikzeValues returns the values of something worth gross inside IKZE that
// would pay out cash today: net after the exit tax on gross, liquidation
// after the owner's marginal rate on cash, or unknown without that rate.
func ikzeValues(gross, cash kernel.Decimal, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	net, err := afterTax(gross, gross, IKZEExitRate())
	if err != nil {
		return jurisdiction.Values{}, err
	}
	out := values(gross, net, kernel.Decimal{})
	if s == nil || !s.HasMarginalRate {
		out.LiquidationUnknown = true
		return out, nil
	}
	if rateErr := jurisdiction.ValidateMarginalRate(s.MarginalRate); rateErr != nil {
		return jurisdiction.Values{}, fmt.Errorf("IKZE liquidation: %w", rateErr)
	}
	liquidation, err := afterTax(cash, cash, s.MarginalRate)
	if err != nil {
		return jurisdiction.Values{}, err
	}
	out.Liquidation = kernel.NewMoney(liquidation, pln())
	return out, nil
}

// afterTax returns amount less rate of base (taxAt).
func afterTax(amount, base, rate kernel.Decimal) (kernel.Decimal, error) {
	tax, err := taxAt(base, rate)
	if err != nil {
		return kernel.Decimal{}, err
	}
	return amount.Sub(tax), nil
}

// values returns the three values in PLN.
func values(gross, net, liquidation kernel.Decimal) jurisdiction.Values {
	c := pln()
	return jurisdiction.Values{
		Gross: kernel.NewMoney(gross, c), Net: kernel.NewMoney(net, c), Liquidation: kernel.NewMoney(liquidation, c),
	}
}

// allPLN reports whether every amount is in PLN; a zero amount without a
// currency counts as zero PLN.
func allPLN(ms ...kernel.Money) error {
	for _, m := range ms {
		if m.Currency().Equal(pln()) || (m.IsZero() && m.Currency().IsZero()) {
			continue
		}
		return fmt.Errorf("%w: %s", ErrNotPLN, m)
	}
	return nil
}

// OKIAct is what the act on OKI fixes for its yearly tax (D-76): the day it
// is in force from, the balance up to which nothing is taxed, and the rate
// on the part of the average balance above it. Rates are fractions. The
// zero act is not in force: no tax is computed under it.
type OKIAct struct {
	InForceFrom kernel.Date
	Limit       kernel.Money
	Rate        kernel.Decimal
}

// CurrentOKIAct returns the OKI act as this code knows it: not in force,
// with no date, limit or rate, because no act in force fixing them is known
// to it yet (docs/implementation-questions.md). Setting them here, from the
// act's text, is what turns the yearly tax on.
func CurrentOKIAct() OKIAct { return OKIAct{} }

// YearlyTax returns OKI's tax for year (D-76) under p.OKI, from the daily
// balances of every OKI account summed: the average of the balances dated
// in the year on or after the act's in-force date, rounded half up to
// grosze; the part of it above the limit; and the rate of that part,
// rounded half up to grosze. A year the act isn't in force in gets no tax
// and InForce false. Any wrapper but OKI is an error wrapping
// jurisdiction.ErrNoYearlyTax.
func (p *Pack) YearlyTax(w jurisdiction.Wrapper, year int, days []jurisdiction.DayBalance) (jurisdiction.YearlyTax, error) {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return jurisdiction.YearlyTax{}, fmt.Errorf("%w: %q", jurisdiction.ErrNoYearlyTax, w)
	}
	out := jurisdiction.YearlyTax{Wrapper: w, Year: year}
	from, ok, err := p.OKI.inForceIn(year)
	if err != nil || !ok {
		return out, err
	}
	out.InForce, out.Limit, out.Rate = true, p.OKI.Limit, p.OKI.Rate
	zero := kernel.NewMoney(kernel.Decimal{}, pln())
	out.Average, out.Excess, out.Tax = zero, zero, zero
	avg, n, err := averageBalance(year, from, days)
	if err != nil || n == 0 {
		return out, err
	}
	out.Days, out.Average = n, kernel.NewMoney(avg, pln())
	if excess := avg.Sub(p.OKI.Limit.Amount()); excess.Sign() > 0 {
		out.Excess = kernel.NewMoney(excess, pln())
		tax, err := taxAt(excess, p.OKI.Rate)
		if err != nil {
			return jurisdiction.YearlyTax{}, err
		}
		out.Tax = kernel.NewMoney(tax, pln())
	}
	return out, nil
}

// averageBalance returns the average of the balances dated in year on or
// after from, rounded half up to grosze, and how many there were.
func averageBalance(year int, from kernel.Date, days []jurisdiction.DayBalance) (kernel.Decimal, int, error) {
	sum, n := kernel.Decimal{}, 0
	for _, d := range days {
		if d.Date.Year() != year || d.Date.Before(from) {
			continue
		}
		if err := allPLN(d.Balance); err != nil {
			return kernel.Decimal{}, 0, err
		}
		sum, n = sum.Add(d.Balance.Amount()), n+1
	}
	if n == 0 {
		return kernel.Decimal{}, 0, nil
	}
	avg, err := sum.Quo(kernel.NewDecimal(int64(n), 0), grosze, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Decimal{}, 0, fmt.Errorf("OKI average balance: %w", err)
	}
	return avg, n, nil
}

// inForceIn returns the first day of year the act is in force on, and
// false when it is not in force in year at all: it has no date, or it
// starts after the year.
func (a *OKIAct) inForceIn(year int) (kernel.Date, bool, error) {
	if a.InForceFrom.IsZero() || a.InForceFrom.Year() > year {
		return kernel.Date{}, false, nil
	}
	if err := allPLN(a.Limit); err != nil {
		return kernel.Date{}, false, fmt.Errorf("OKI act: %w", err)
	}
	if a.InForceFrom.Year() == year {
		return a.InForceFrom, true, nil
	}
	first, err := kernel.NewDate(year, time.January, 1)
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("OKI year %d: %w", year, err)
	}
	return first, true, nil
}
