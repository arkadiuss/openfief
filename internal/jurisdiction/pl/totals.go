package pl

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// Totals returns the three values (D-77) of a total of accounts' rows, an
// account's or the portfolio's, counting each wrapper's tax on the total
// (D-128) instead of adding up the rows' own taxes:
//
//   - Taxable: the gains of every taxable account's holdings priced at
//     market offset each other and the tax year's realized gains and
//     losses, as PIT-38 counts them (PIT38Tax): the net total takes the
//     unrealized gain over the cost basis, the liquidation total the gain
//     of the proceeds after the sale cost. Bonds keep their own rows' tax,
//     which the issuer withholds outside PIT-38; losses carried from
//     earlier years are not counted.
//   - IKE: net is gross; liquidation is an early return of the whole
//     account, its proceeds after sale costs taxed at 19% on what they
//     exceed the payments in (PaidIn), rounded half up to grosze.
//   - IKZE: net is gross less the 10% exit tax on the account's gross,
//     liquidation the proceeds less the owner's marginal rate on them,
//     unknown while that rate is not set (D-78).
//   - OKI: the rows' sum: no tax inside it (D-127).
//
// Every amount must be in PLN; a zero amount without a currency counts as
// zero PLN.
func (*Pack) Totals(accounts []jurisdiction.AccountRows, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	out := values(kernel.Decimal{}, kernel.Decimal{}, kernel.Decimal{})
	var pit pit38
	for i := range accounts {
		v, err := pit.accountTotal(&accounts[i], s)
		if err != nil {
			return jurisdiction.Values{}, err
		}
		if out, err = sumValues(&out, &v); err != nil {
			return jurisdiction.Values{}, err
		}
	}
	tax, err := pit.taxes()
	if err != nil {
		return jurisdiction.Values{}, err
	}
	if out, err = sumValues(&out, &tax); err != nil {
		return jurisdiction.Values{}, err
	}
	return out, nil
}

// PIT38Tax returns the tax PIT-38 counts on base, the year's gains less
// losses: the base and the tax are rounded to whole złoty, under 50 groszy
// down and from 50 groszy up (art. 63 § 1 of the tax ordinance), and a base
// that is not positive is taxed nothing. A base of 1,234.50 PLN rounds to
// 1,235, and 19% of it, 234.65, to 235 PLN.
func PIT38Tax(base kernel.Decimal) (kernel.Decimal, error) {
	whole, err := base.Round(0, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("PIT-38 base: %w", err)
	}
	if whole.Sign() <= 0 {
		return kernel.Decimal{}, nil
	}
	tax, err := whole.Mul(BelkaRate()).Round(0, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("PIT-38 tax: %w", err)
	}
	return tax, nil
}

// pit38 sums the PIT-38 bases of a total's taxable accounts, for its net
// and liquidation values.
type pit38 struct {
	baseNet, baseLiquidation kernel.Decimal
}

// accountTotal returns a's part of a total under its wrapper's rules; a
// taxable account's part leaves its tax to p.
func (p *pit38) accountTotal(a *jurisdiction.AccountRows, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	w, err := ParseTaxWrapper(string(a.Wrapper))
	if err != nil {
		return jurisdiction.Values{}, err
	}
	if err := allPLN(a.Rows.Gross, a.Rows.Net, a.Rows.Liquidation, a.Listed.Gross, a.Listed.Net, a.Listed.Liquidation,
		a.ListedCost, a.ListedSaleCost, a.Realized, a.PaidIn); err != nil {
		return jurisdiction.Values{}, err
	}
	switch w {
	case Taxable:
		return p.add(a), nil
	case IKE:
		return ikeTotal(a)
	case IKZE:
		return ikzeTotal(a, s)
	case OKI:
		return a.Rows, nil
	}
	return jurisdiction.Values{}, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, w)
}

// add adds a taxable account and returns its values without its listed
// holdings' own taxes: their gross for net, their proceeds after the sale
// cost for liquidation.
func (p *pit38) add(a *jurisdiction.AccountRows) jurisdiction.Values {
	listedGross, saleCost := a.Listed.Gross.Amount(), a.ListedSaleCost.Amount()
	proceeds := listedGross.Sub(saleCost)
	net := a.Rows.Net.Amount().Sub(a.Listed.Net.Amount()).Add(listedGross)
	liquidation := a.Rows.Liquidation.Amount().Sub(a.Listed.Liquidation.Amount()).Add(proceeds)
	gain := listedGross.Sub(a.ListedCost.Amount()).Add(a.Realized.Amount())
	p.baseNet = p.baseNet.Add(gain)
	p.baseLiquidation = p.baseLiquidation.Add(gain.Sub(saleCost))
	out := values(a.Rows.Gross.Amount(), net, liquidation)
	out.LiquidationUnknown = a.Rows.LiquidationUnknown
	return out
}

// taxes returns the PIT-38 taxes of the summed accounts as negative values.
func (p *pit38) taxes() (jurisdiction.Values, error) {
	net, err := PIT38Tax(p.baseNet)
	if err != nil {
		return jurisdiction.Values{}, err
	}
	liquidation, err := PIT38Tax(p.baseLiquidation)
	if err != nil {
		return jurisdiction.Values{}, err
	}
	return values(kernel.Decimal{}, net.Neg(), liquidation.Neg()), nil
}

// ikeTotal returns an IKE account's total: gross at the exit, and an early
// return taxed on what its proceeds exceed the payments in.
func ikeTotal(a *jurisdiction.AccountRows) (jurisdiction.Values, error) {
	gross := a.Rows.Gross.Amount()
	proceeds := gross.Sub(a.ListedSaleCost.Amount())
	liquidation, err := afterTax(proceeds, proceeds.Sub(a.PaidIn.Amount()), BelkaRate())
	if err != nil {
		return jurisdiction.Values{}, err
	}
	return values(gross, a.Rows.Net.Amount(), liquidation), nil
}

// ikzeTotal returns an IKZE account's total: the exit tax on its gross, and
// the owner's marginal rate on its proceeds.
func ikzeTotal(a *jurisdiction.AccountRows, s *jurisdiction.Settings) (jurisdiction.Values, error) {
	gross := a.Rows.Gross.Amount()
	return ikzeValues(gross, gross.Sub(a.ListedSaleCost.Amount()), s)
}

// sumValues adds v to sum, keeping that some liquidation value is unknown.
// Every amount is in PLN (allPLN), so only the amounts are added.
func sumValues(sum, v *jurisdiction.Values) (jurisdiction.Values, error) {
	if err := allPLN(v.Gross, v.Net, v.Liquidation); err != nil {
		return jurisdiction.Values{}, fmt.Errorf("total: %w", err)
	}
	out := values(sum.Gross.Amount().Add(v.Gross.Amount()), sum.Net.Amount().Add(v.Net.Amount()),
		sum.Liquidation.Amount().Add(v.Liquidation.Amount()))
	out.LiquidationUnknown = sum.LiquidationUnknown || v.LiquidationUnknown
	if out.LiquidationUnknown {
		out.Liquidation = kernel.NewMoney(kernel.Decimal{}, pln())
	}
	return out, nil
}
