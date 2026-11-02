package jurisdiction

import "github.com/arkadiuss/openfief/internal/kernel"

// ExemptionClass is a class a pack's yearly tax tells assets apart by, such as
// the Polish pack's savings and investment assets for OKI's exemptions
// (D-127). Packs define their own; the zero value is none.
type ExemptionClass string

// AccountForm is what kind of account an account is, for a wrapper whose
// rules tell forms apart, such as an OKI account's form, which decides
// whether its cash is a savings or an investment asset (D-147). Packs
// define their own; the zero value is none set.
type AccountForm string

// AssetKind is how an asset is valued: at market, at its accrued value, or
// as cash.
type AssetKind int

// The asset kinds.
const (
	// AssetListed is a holding priced at market, such as a stock or a fund.
	AssetListed AssetKind = iota + 1
	// AssetAccrued is a holding valued at its accrued value, such as a
	// treasury savings bond.
	AssetAccrued
	// AssetCash is cash.
	AssetCash
)

// Asset is what a pack classifies for a yearly tax (D-127, D-132).
type Asset struct {
	Kind AssetKind
	// Instrument is the holding's ISIN; empty for cash.
	Instrument string
	// Currency is cash's currency, or the currency a holding is priced in.
	Currency kernel.Currency
	// Form is the form of the asset's account (D-147), which decides the
	// class of its cash; empty when none is set.
	Form AccountForm
}

// ClassValue is an amount of one exemption class.
type ClassValue struct {
	Class ExemptionClass
	Value kernel.Money
}

// TaxDay is one day an account was held in a yearly tax's year: its value
// at the end of the day per exemption class, and the payments into it and
// withdrawals from it that day, every amount in the pack's currency
// converted at the rate the pack's tax rules take (Pack.TaxRateDate).
type TaxDay struct {
	Date   kernel.Date
	Values []ClassValue
	In     kernel.Money
	Out    kernel.Money
}

// TaxAccount is one account's days in a yearly tax's year: one TaxDay for
// each day of the year it was held, through the last day counted.
type TaxAccount struct {
	Account string
	Days    []TaxDay
}

// YearlyTax is a wrapper's tax for one year on the value of its accounts,
// such as OKI's (D-127).
type YearlyTax struct {
	Wrapper Wrapper
	Year    int
	// InForce reports that the law that levies it was in force in the year.
	// When it is false, nothing else but Wrapper and Year is set: no tax is
	// computed for a year the law didn't cover.
	InForce bool
	// Days is the most days any account was held in the year, through
	// Through, the last day counted; zero when none was.
	Days    int
	Through kernel.Date
	// Estimate reports a year counted through a day before its last, such
	// as the current year: the tax is an estimate over the days so far
	// (D-146).
	Estimate bool
	// Average is the base before exemptions: the sum over the accounts of
	// each one's average value. Averages splits it by exemption class. A
	// pack may keep them exact and show them rounded, as the Polish pack
	// does (D-144).
	Average  kernel.Money
	Averages []ClassValue
	// Exempt is the exempt part of each class's average, and Taxed the
	// part of Average left to tax, the tax base, which the Polish pack
	// rounds to whole złoty (D-144). ExemptionsKnown is false when the law's
	// exempt amounts for the year are not known: Exempt, Taxed and Tax are
	// then unset.
	Exempt          []ClassValue
	ExemptionsKnown bool
	Taxed           kernel.Money
	// Rate is the year's rate as a fraction; RateKnown is false when it is
	// not known yet, and Tax is then unset.
	Rate      kernel.Decimal
	RateKnown bool
	Tax       kernel.Money
}

// AccountRows is one account's rows on one day, what its part of a total
// is computed from (D-128). Every amount is in the pack's currency.
type AccountRows struct {
	Wrapper Wrapper
	// Rows sums the account's rows: each holding's values carrying its own
	// tax as if it alone were sold, and its cash.
	Rows Values
	// Listed sums the rows of its holdings priced at market, ListedCost
	// their cost basis for tax and ListedSaleCost what selling them today
	// would cost under the account's fee rule.
	Listed         Values
	ListedCost     kernel.Money
	ListedSaleCost kernel.Money
	// Realized is the gains less the losses of the account's sales in the
	// tax year through the day.
	Realized kernel.Money
	// PaidIn is what was paid into the account less what was withdrawn
	// from it, through the day.
	PaidIn kernel.Money
}
