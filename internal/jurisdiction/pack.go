package jurisdiction

import (
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of the pack contract.
var (
	// ErrUnknownWrapper reports a tax wrapper the pack does not have.
	ErrUnknownWrapper = errors.New("unknown tax wrapper")
	// ErrInvalidMarginalRate reports an owner's marginal income tax rate
	// that is not a fraction above 0 and below 1.
	ErrInvalidMarginalRate = errors.New("marginal income tax rate must be above 0 and below 1")
	// ErrNoYearlyTax reports a yearly tax asked for a wrapper that has none.
	ErrNoYearlyTax = errors.New("the tax wrapper has no yearly tax")
)

// Wrapper names one of a pack's tax wrappers (D-76), such as "ike": the key
// an account stores. Packs define their own; the zero value is none.
type Wrapper string

// Values are the three values of D-77 of a holding, an account's cash or a
// total, in the pack's tax currency.
type Values struct {
	// Gross is the value before any tax or fee: the market value, or the
	// accrued value of a bond or deposit.
	Gross kernel.Money
	// Net is what the owner keeps at the planned exit (maturity, or a
	// withdrawal from a wrapper once its conditions are met), after the
	// taxes due then and without early-exit fees.
	Net kernel.Money
	// Liquidation is the cash in hand if it were sold or redeemed and
	// withdrawn today, after every tax and fee due today.
	Liquidation kernel.Money
	// LiquidationUnknown reports that the liquidation value needs an owner
	// setting that is not set, such as the marginal income tax rate for an
	// early withdrawal from IKZE (D-78): Liquidation is then zero and must
	// be shown as missing, never guessed.
	LiquidationUnknown bool
}

// Settings are the owner's settings that the values need (D-78).
type Settings struct {
	// MarginalRate is the owner's marginal income tax rate as a fraction,
	// such as 0.32; it counts only when HasMarginalRate.
	MarginalRate    kernel.Decimal
	HasMarginalRate bool
}

// ValidateMarginalRate reports whether r is a marginal income tax rate: a
// fraction above 0 and below 1.
func ValidateMarginalRate(r kernel.Decimal) error {
	if r.Sign() <= 0 || r.Cmp(kernel.NewDecimal(1, 0)) >= 0 {
		return fmt.Errorf("%w: %s", ErrInvalidMarginalRate, r)
	}
	return nil
}

// Listed is a holding priced at market, such as a stock or an ETF, with
// every amount in the pack's tax currency.
type Listed struct {
	Wrapper Wrapper
	// Gross is its market value.
	Gross kernel.Money
	// Cost is its cost basis for tax: the cost of its open lots, each
	// converted at the rate the pack's tax rules take.
	Cost kernel.Money
	// SaleCost is what selling it today would cost: the commission and the
	// currency conversion fee of the account's fee rule (FeeRule.SaleCost).
	SaleCost kernel.Money
}

// Accrued is a bond or deposit valued at its accrued value, with every
// amount in the pack's tax currency.
type Accrued struct {
	Wrapper Wrapper
	// Gross is its accrued value: nominal plus interest accrued and not yet
	// paid out.
	Gross kernel.Money
	// Income is what its exit taxes as income: the accrued interest plus
	// any discount of a purchase below the nominal.
	Income kernel.Money
	// EarlyFee is the fee redeeming it today would charge in its wrapper;
	// zero at maturity.
	EarlyFee kernel.Money
}

// Cash is an account's cash in the pack's tax currency.
type Cash struct {
	Wrapper Wrapper
	Amount  kernel.Money
}

// DayBalance is the value of the accounts in one wrapper at the end of a
// day, summed, for a yearly tax on the balance.
type DayBalance struct {
	Date    kernel.Date
	Balance kernel.Money
}

// YearlyTax is a wrapper's tax for one year on its balance, such as OKI's
// on the part of the average balance above a limit (D-76).
type YearlyTax struct {
	Wrapper Wrapper
	Year    int
	// InForce reports that the law that levies it was in force in the year.
	// When it is false, nothing else but Wrapper and Year is set: no tax is
	// computed for a year the law didn't cover.
	InForce bool
	// Days counts the days the average was taken over.
	Days int
	// Average is the average balance over those days.
	Average kernel.Money
	// Limit and Rate are the law's: the balance up to Limit is not taxed,
	// the part above it at Rate.
	Limit kernel.Money
	Rate  kernel.Decimal
	// Excess is the part of Average above Limit, and Tax the tax on it.
	Excess kernel.Money
	Tax    kernel.Money
}

// Pack is a jurisdiction pack (D-70): the pure rules of one country.
type Pack interface {
	// Version names the pack and the version of its rules, such as "pl/1".
	// Values computed under one version are rebuilt when it changes.
	Version() string
	// ReportingCurrency is the default reporting currency (D-62), and the
	// currency every amount the pack values is in.
	ReportingCurrency() kernel.Currency
	// NumberFormat is the default number format (D-113).
	NumberFormat() NumberFormat
	// Wrappers are the pack's tax wrappers (D-76), the default first.
	Wrappers() []Wrapper
	// TaxRateDate returns the latest date whose exchange rate converts an
	// amount of an event on date event for tax, such as a lot's cost.
	TaxRateDate(event kernel.Date) kernel.Date
	// ValueListed returns the three values of a holding priced at market.
	ValueListed(in *Listed, s *Settings) (Values, error)
	// ValueAccrued returns the three values of a bond or deposit.
	ValueAccrued(in *Accrued, s *Settings) (Values, error)
	// ValueCash returns the three values of cash.
	ValueCash(in *Cash, s *Settings) (Values, error)
	// YearlyTax returns w's tax for year on the daily balances summed over
	// every account in w, or an error wrapping ErrNoYearlyTax for a wrapper
	// without one.
	YearlyTax(w Wrapper, year int, days []DayBalance) (YearlyTax, error)
}

// HasWrapper reports whether p has the tax wrapper w.
func HasWrapper(p Pack, w Wrapper) bool {
	return slices.Contains(p.Wrappers(), w)
}

// DefaultWrapper returns p's default tax wrapper, its first.
func DefaultWrapper(p Pack) Wrapper {
	all := p.Wrappers()
	if len(all) == 0 {
		return ""
	}
	return all[0]
}
