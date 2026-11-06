package views

import (
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TaxedAccount is one account in a total whose tax its wrapper's rules
// count (D-128): its stored valuations and cash, the gains of its sales,
// and what was paid into it, every amount in the reporting currency.
type TaxedAccount struct {
	Wrapper jurisdiction.Wrapper
	Vals    []valuation.HoldingValuation
	Cash    []DayCash
	// Realized are its sales' gains for tax; RealizedMissing reports that
	// some lacked an FX rate, so its net and liquidation totals are not
	// given.
	Realized        []valuation.RealizedGain
	RealizedMissing bool
	// PaidIn are the payments into it (positive) and withdrawals (negative)
	// for tax (PaidIn); PaidInMissing reports that some lacked an FX rate.
	PaidIn        []Payment
	PaidInMissing bool
}

// Payment is an amount paid into an account, or withdrawn when negative,
// in the pack's currency for tax.
type Payment struct {
	Date   kernel.Date
	Amount kernel.Money
}

// PaidIn returns what was paid into one account and withdrawn from it,
// from its current transactions txs, in date order, each converted at the
// latest FX rate on or before the date pack takes for it (Pack.TaxRateDate
// of its date): each deposit and withdrawal. The earlier history of an
// account seeded from a starting balance (D-23) is unknown, so what was
// paid in before is earlier, the total the owner entered (D-149), dated on
// its go-live, the date of its latest seeding transaction; its seeding, the
// opening cash and opening lots, is no payment. It reports true, missing,
// for such an account with earlier nil, and when an amount has no rate.
func PaidIn(txs []ledger.Transaction, rates []valuation.FXRate, pack jurisdiction.Pack, earlier *kernel.Money) ([]Payment, bool, error) {
	var out []Payment
	if goLive, seeded := goLiveOf(txs); seeded {
		if earlier == nil {
			return nil, true, nil
		}
		out = append(out, Payment{Date: goLive, Amount: *earlier})
	}
	for _, f := range ExternalFlows(txs) {
		if f.Seeding {
			continue
		}
		pln, err := valuation.ConvertAt(f.Amount, pack.ReportingCurrency(), rates, pack.TaxRateDate(f.Date), valuation.ValueRounding)
		if errors.Is(err, valuation.ErrNoRate) {
			return nil, true, nil
		}
		if err != nil {
			return nil, false, fmt.Errorf("paid in: %w", err)
		}
		out = append(out, Payment{Date: f.Date, Amount: pln})
	}
	slices.SortStableFunc(out, func(a, b Payment) int { return a.Date.Compare(b.Date) })
	return out, false, nil
}

// goLiveOf returns the date of the latest seeding transaction among txs,
// and false when there is none: the account has its whole history.
func goLiveOf(txs []ledger.Transaction) (kernel.Date, bool) {
	var out kernel.Date
	seeded := false
	for i := range txs {
		if t := &txs[i]; ledger.IsSeeding(t) && (!seeded || t.Date.After(out)) {
			out, seeded = t.Date, true
		}
	}
	return out, seeded
}

// TaxTotals sets the net and liquidation values of each total in totals,
// which must sum exactly the rows of accounts, to what pack counts on the
// accounts' rows of its date (jurisdiction.Pack.Totals), and its losses
// offset to the difference from the rows: for one account, its holdings'
// and cash's rows; for several, the accounts' own totals, each counted the
// same way. A value some row or input lacks stays missing.
func TaxTotals(totals []DateTotal, accounts []TaxedAccount, pack jurisdiction.Pack, settings *jurisdiction.Settings) error {
	for i := range totals {
		t := &totals[i]
		rows := make([]jurisdiction.AccountRows, 0, len(accounts))
		for j := range accounts {
			r, err := accountRows(&accounts[j], t.Date, pack.ReportingCurrency())
			if err != nil {
				return fmt.Errorf("total on %s: %w", t.Date, err)
			}
			rows = append(rows, r)
			t.NetMissing = t.NetMissing || accounts[j].RealizedMissing
			t.LiquidationMissing = t.LiquidationMissing || accounts[j].RealizedMissing || accounts[j].PaidInMissing
		}
		if err := t.tax(rows, pack, settings); err != nil {
			return fmt.Errorf("total on %s: %w", t.Date, err)
		}
	}
	return nil
}

// tax sets t's taxed values and their offset from the rows.
func (t *DateTotal) tax(rows []jurisdiction.AccountRows, pack jurisdiction.Pack, settings *jurisdiction.Settings) error {
	taxed, err := pack.Totals(rows, settings)
	if err != nil {
		return fmt.Errorf("taxing the total: %w", err)
	}
	sum := rowSum(rows)
	if len(rows) > 1 {
		sum = jurisdiction.Values{}
		for i := range rows {
			own, err := pack.Totals(rows[i:i+1], settings)
			if err != nil {
				return fmt.Errorf("taxing an account's total: %w", err)
			}
			sum = addTaxValues(&sum, &own)
		}
	}
	zero := kernel.NewMoney(kernel.Decimal{}, taxed.Gross.Currency())
	t.NetOffset, t.LiquidationOffset = zero, zero
	if !t.NetMissing {
		t.Net, t.NetOffset = taxed.Net, minus(taxed.Net, sum.Net)
	}
	if taxed.LiquidationUnknown || sum.LiquidationUnknown {
		t.LiquidationMissing = true
	}
	if !t.LiquidationMissing {
		t.Liquidation, t.LiquidationOffset = taxed.Liquidation, minus(taxed.Liquidation, sum.Liquidation)
	}
	return nil
}

// rowSum returns the sum of the accounts' rows.
func rowSum(rows []jurisdiction.AccountRows) jurisdiction.Values {
	var sum jurisdiction.Values
	for i := range rows {
		sum = addTaxValues(&sum, &rows[i].Rows)
	}
	return sum
}

// accountRows returns a's rows on day for its total: the sums of its valued
// holdings' and its converted cash's values, of its holdings priced at
// market apart with their cost basis and sale cost, its realized gains in
// day's year, in currency in, and what was paid in through day.
func accountRows(a *TaxedAccount, day kernel.Date, in kernel.Currency) (jurisdiction.AccountRows, error) {
	out := jurisdiction.AccountRows{Wrapper: a.Wrapper}
	for i := range a.Vals {
		if v := &a.Vals[i]; v.Date.Equal(day) && v.Valued() {
			addHolding(&out, v)
		}
	}
	for i := range a.Cash {
		if c := &a.Cash[i]; c.Date.Equal(day) && !c.Unconverted {
			out.Rows = addTaxValues(&out.Rows, &c.Values)
		}
	}
	realized, err := valuation.RealizedInYear(a.Realized, day, in)
	if err != nil {
		return jurisdiction.AccountRows{}, fmt.Errorf("account rows: %w", err)
	}
	out.Realized = realized
	for i := range a.PaidIn {
		if p := &a.PaidIn[i]; !p.Date.After(day) {
			out.PaidIn = plus(out.PaidIn, p.Amount)
		}
	}
	return out, nil
}

// rowsOf returns a valued holding's values as a row.
func rowsOf(v *valuation.HoldingValuation) jurisdiction.Values {
	return jurisdiction.Values{Gross: v.Reporting, Net: v.Net, Liquidation: v.Liquidation, LiquidationUnknown: !v.LiquidationValued()}
}

// addHolding adds a valued holding to out's rows and, priced at market with
// its net value, to its listed holdings with its cost basis and sale cost.
func addHolding(out *jurisdiction.AccountRows, v *valuation.HoldingValuation) {
	row := rowsOf(v)
	out.Rows = addTaxValues(&out.Rows, &row)
	if v.Basis != valuation.Market || !v.NetValued() {
		return
	}
	out.Listed = addTaxValues(&out.Listed, &row)
	out.ListedCost, out.ListedSaleCost = plus(out.ListedCost, v.TaxCost), plus(out.ListedSaleCost, v.SaleCost)
}

// addTaxValues adds v to sum by amount: every value of a total is in the
// reporting currency, and a sum with no currency yet is nothing.
func addTaxValues(sum, v *jurisdiction.Values) jurisdiction.Values {
	return jurisdiction.Values{
		Gross: plus(sum.Gross, v.Gross), Net: plus(sum.Net, v.Net), Liquidation: plus(sum.Liquidation, v.Liquidation),
		LiquidationUnknown: sum.LiquidationUnknown || v.LiquidationUnknown,
	}
}

// plus returns a + b in b's currency, or a's when b has none.
func plus(a, b kernel.Money) kernel.Money {
	c := b.Currency()
	if c.IsZero() {
		c = a.Currency()
	}
	return kernel.NewMoney(a.Amount().Add(b.Amount()), c)
}

// minus returns a - b in a's currency.
func minus(a, b kernel.Money) kernel.Money {
	return kernel.NewMoney(a.Amount().Sub(b.Amount()), a.Currency())
}
