package valuation

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// AccountTerms are what an account's net and liquidation values (D-77)
// depend on besides its holdings: the jurisdiction pack that computes them
// (D-70), the account's tax wrapper (D-76) and fee rule, and the owner's
// settings, such as the marginal income tax rate (D-78).
type AccountTerms struct {
	Pack     jurisdiction.Pack
	Wrapper  jurisdiction.Wrapper
	Fee      jurisdiction.FeeRule
	Settings jurisdiction.Settings
}

// TaxCost returns the cost basis of lots in PLN for tax: each lot's cost
// converted at the latest rate on or before the date the pack takes for an
// event on its open date (Pack.TaxRateDate), rounded with ValueRounding.
// It returns MissingLots when lotsUnknown, and MissingFXRate when a lot's
// currency has no such rate.
func (a *AccountTerms) TaxCost(lots []ledger.Lot, lotsUnknown bool, rates []FXRate) (kernel.Money, MissingReason, error) {
	total := kernel.NewMoney(kernel.Decimal{}, PLN())
	if lotsUnknown {
		return kernel.Money{}, MissingLots, nil
	}
	for i := range lots {
		c, err := ConvertToPLN(lots[i].Cost, rates, a.Pack.TaxRateDate(lots[i].Opened), ValueRounding)
		if errors.Is(err, ErrNoRate) {
			return kernel.Money{}, MissingFXRate, nil
		}
		if err != nil {
			return kernel.Money{}, 0, fmt.Errorf("lot %s: %w", lots[i].Origin, err)
		}
		if total, err = total.Add(c); err != nil {
			return kernel.Money{}, 0, fmt.Errorf("lot %s: %w", lots[i].Origin, err)
		}
	}
	return total, 0, nil
}

// listed sets the net and liquidation values of row, a holding priced at
// market whose gross value in PLN is set, with the pack's ValueListed: its
// cost basis is its open lots' costs, each converted at the rate the pack
// takes for tax (TaxCost), and selling it costs what the fee rule charges
// (FeeRule.SaleCost), with its conversion fee when the price is in another
// currency than the pack's. Lots or rates it lacks leave the values it
// needs them for missing, with the reason.
func (a *AccountTerms) listed(row *HoldingValuation, m *MarketHolding) error {
	cost, costMissing, err := a.TaxCost(m.Lots, m.LotsUnknown, m.Rates)
	if err != nil {
		return fmt.Errorf("cost basis of %s: %w", row.Instrument, err)
	}
	if costMissing != 0 {
		row.NetMissing, row.LiquidationMissing = costMissing, costMissing
		return nil
	}
	minimum, minimumMissing, err := a.minimum(m.Rates, row.Date)
	if err != nil {
		return err
	}
	foreign := !row.Value.Currency().Equal(a.Pack.ReportingCurrency())
	saleCost, err := a.Fee.SaleCost(row.PLN, minimum, foreign)
	if err != nil {
		return fmt.Errorf("sale cost of %s: %w", row.Instrument, err)
	}
	if minimumMissing != 0 {
		saleCost = kernel.NewMoney(kernel.Decimal{}, PLN())
	}
	v, err := a.Pack.ValueListed(&jurisdiction.Listed{Wrapper: a.Wrapper, Gross: row.PLN, Cost: cost, SaleCost: saleCost}, &a.Settings)
	if err != nil {
		return fmt.Errorf("values of %s: %w", row.Instrument, err)
	}
	row.Net, row.Liquidation = v.Net, v.Liquidation
	switch {
	case minimumMissing != 0:
		row.Liquidation, row.LiquidationMissing = kernel.Money{}, minimumMissing
	case v.LiquidationUnknown:
		row.Liquidation, row.LiquidationMissing = kernel.Money{}, MissingMarginalRate
	}
	return nil
}

// minimum returns the fee rule's minimum in PLN on day, converted at the
// latest rate on or before it, or MissingFXRate when there is none.
func (a *AccountTerms) minimum(rates []FXRate, day kernel.Date) (kernel.Money, MissingReason, error) {
	if a.Fee.Minimum.IsZero() {
		return kernel.NewMoney(kernel.Decimal{}, PLN()), 0, nil
	}
	m, err := ConvertToPLN(a.Fee.Minimum, rates, day, ValueRounding)
	if errors.Is(err, ErrNoRate) {
		return kernel.Money{}, MissingFXRate, nil
	}
	if err != nil {
		return kernel.Money{}, 0, fmt.Errorf("the fee rule's minimum: %w", err)
	}
	return m, 0, nil
}

// addValues adds v to sum, keeping that some liquidation value is unknown;
// a sum with no currency yet is nothing.
func addValues(sum, v *jurisdiction.Values) (jurisdiction.Values, error) {
	if sum.Gross.Currency().IsZero() {
		return *v, nil
	}
	gross, err1 := sum.Gross.Add(v.Gross)
	net, err2 := sum.Net.Add(v.Net)
	liquidation, err3 := sum.Liquidation.Add(v.Liquidation)
	if err := errors.Join(err1, err2, err3); err != nil {
		return jurisdiction.Values{}, fmt.Errorf("adding values: %w", err)
	}
	return jurisdiction.Values{
		Gross: gross, Net: net, Liquidation: liquidation, LiquidationUnknown: sum.LiquidationUnknown || v.LiquidationUnknown,
	}, nil
}
