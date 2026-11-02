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

// ErrTaxConversion reports an amount a yearly tax needs in PLN with no FX
// rate on or before the date the pack's tax rules take.
var ErrTaxConversion = errors.New("no FX rate for the yearly tax")

// TaxYearAccount is one account's record for a yearly tax (D-127): the
// days of the year it was held that are counted, From through To, its
// form (D-147), and its stored valuations, cash, cash flows and the FX
// rates of their currencies, with the owner's classes of instruments
// (D-132).
type TaxYearAccount struct {
	Account  ledger.AccountID
	Wrapper  jurisdiction.Wrapper
	Form     jurisdiction.AccountForm
	From, To kernel.Date
	Vals     []valuation.HoldingValuation
	Cash     []CashBalance
	Flows    []CashFlow
	Rates    []valuation.FXRate
	Classes  map[string]jurisdiction.ExemptionClass
}

// TaxDays returns a's days for pack's yearly tax, every amount converted to
// PLN at the latest FX rate on or before the date the pack takes
// (Pack.TaxRateDate: for the Polish pack, the NBP rate of the last business
// day before), rounded with valuation.ValueRounding, and the instruments
// counted in the class pack gives them without knowing it because the
// owner hasn't classified them (D-132), in ISIN order. Each day has its
// valued holdings and cash per exemption class, and its deposits as payments
// in and withdrawals as payments out; opening cash is no payment. An
// amount with no rate is an error wrapping ErrTaxConversion.
func TaxDays(a *TaxYearAccount, pack jurisdiction.Pack) (jurisdiction.TaxAccount, []string, error) {
	out := jurisdiction.TaxAccount{Account: string(a.Account)}
	unclassified := map[string]bool{}
	for day := a.From; !day.After(a.To); day = day.AddDays(1) {
		td, err := a.taxDay(pack, day, unclassified)
		if err != nil {
			return jurisdiction.TaxAccount{}, nil, err
		}
		out.Days = append(out.Days, td)
	}
	missing := make([]string, 0, len(unclassified))
	for isin := range unclassified {
		missing = append(missing, isin)
	}
	slices.Sort(missing)
	return out, missing, nil
}

// taxDay returns a's day for the yearly tax.
func (a *TaxYearAccount) taxDay(pack jurisdiction.Pack, day kernel.Date, unclassified map[string]bool) (jurisdiction.TaxDay, error) {
	cutoff := pack.TaxRateDate(day)
	sums := map[jurisdiction.ExemptionClass]kernel.Decimal{}
	for _, held := range a.assetsOn(day) {
		pln, err := a.toPLN(held.amount, cutoff)
		if err != nil {
			return jurisdiction.TaxDay{}, err
		}
		class := a.classOf(pack, &held.asset, unclassified)
		sums[class] = sums[class].Add(pln.Amount())
	}
	out := jurisdiction.TaxDay{Date: day}
	for _, class := range pack.ExemptionClasses(a.Wrapper) {
		if s, ok := sums[class]; ok {
			out.Values = append(out.Values, jurisdiction.ClassValue{Class: class, Value: kernel.NewMoney(s, valuation.PLN())})
		}
	}
	inOut, err := a.payments(day, cutoff)
	out.In, out.Out = inOut[0], inOut[1]
	return out, err
}

// heldAsset is an asset a's account held at the end of a day, with its
// amount in its own currency.
type heldAsset struct {
	asset  jurisdiction.Asset
	amount kernel.Money
}

// assetsOn returns a's valued holdings and cash on day, its cash with the
// account's form.
func (a *TaxYearAccount) assetsOn(day kernel.Date) []heldAsset {
	var out []heldAsset
	for i := range a.Vals {
		v := &a.Vals[i]
		if !v.Date.Equal(day) || !v.Valued() {
			continue
		}
		asset := jurisdiction.Asset{Kind: jurisdiction.AssetListed, Instrument: v.Instrument.ISIN.String(), Currency: v.Value.Currency()}
		if v.Basis == valuation.GrossAccrued {
			asset.Kind = jurisdiction.AssetAccrued
		}
		out = append(out, heldAsset{asset: asset, amount: v.Value})
	}
	for i := range a.Cash {
		if c := &a.Cash[i]; c.Date.Equal(day) {
			asset := jurisdiction.Asset{Kind: jurisdiction.AssetCash, Currency: c.Amount.Currency(), Form: a.Form}
			out = append(out, heldAsset{asset: asset, amount: c.Amount})
		}
	}
	return out
}

// classOf returns asset's class: the pack's when it can tell, the owner's
// when the owner classified the instrument, and otherwise the class the
// pack counts it in, noting the instrument in unclassified.
func (a *TaxYearAccount) classOf(pack jurisdiction.Pack, asset *jurisdiction.Asset, unclassified map[string]bool) jurisdiction.ExemptionClass {
	class, known := pack.ExemptionClass(a.Wrapper, asset)
	if known {
		return class
	}
	if owner, ok := a.Classes[asset.Instrument]; ok && asset.Instrument != "" {
		return owner
	}
	unclassified[asset.Instrument] = true
	return class
}

// payments returns a's deposits and withdrawals on day in PLN, in that
// order, both not negative.
func (a *TaxYearAccount) payments(day, cutoff kernel.Date) ([2]kernel.Money, error) {
	paid, withdrawn := kernel.Decimal{}, kernel.Decimal{}
	for i := range a.Flows {
		f := &a.Flows[i]
		if !f.Date.Equal(day) || f.Seeding {
			continue
		}
		pln, err := a.toPLN(f.Amount, cutoff)
		if err != nil {
			return [2]kernel.Money{}, err
		}
		if pln.Sign() > 0 {
			paid = paid.Add(pln.Amount())
		} else {
			withdrawn = withdrawn.Sub(pln.Amount())
		}
	}
	return [2]kernel.Money{kernel.NewMoney(paid, valuation.PLN()), kernel.NewMoney(withdrawn, valuation.PLN())}, nil
}

// toPLN converts m at the latest rate on or before cutoff.
func (a *TaxYearAccount) toPLN(m kernel.Money, cutoff kernel.Date) (kernel.Money, error) {
	pln, err := valuation.ConvertToPLN(m, a.Rates, cutoff, valuation.ValueRounding)
	if errors.Is(err, valuation.ErrNoRate) {
		return kernel.Money{}, fmt.Errorf("%w: %s of %s on or before %s", ErrTaxConversion, m.Currency(), a.Account, cutoff)
	}
	if err != nil {
		return kernel.Money{}, fmt.Errorf("converting for the yearly tax: %w", err)
	}
	return pln, nil
}
