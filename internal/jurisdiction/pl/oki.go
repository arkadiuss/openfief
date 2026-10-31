package pl

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrTaxDay reports a day of an OKI account's tax year that the yearly tax
// can't count: an amount not in PLN, a negative payment, an exemption class the
// pack doesn't have or a day given twice.
var ErrTaxDay = errors.New("invalid OKI tax day")

// The exemption classes of OKI's yearly tax (D-127): the exemptions of art. 26
// ust. 1 of the act on OKI (Dz.U. 2026 poz. 1098) count savings assets
// (pkt 1) and investment assets (pkt 2) apart, and other assets get none.
const (
	// Savings are art. 26 ust. 1 pkt 1's assets: PLN cash on a bank
	// account or on the cash account of a treasury savings bond account,
	// treasury savings bonds, PLN Treasury bills and the lowest-risk funds.
	Savings jurisdiction.ExemptionClass = "savings"
	// Investment are pkt 2's assets: qualifying funds, PLN bonds and covered
	// bonds, shares of companies with capital in PLN, and PLN cash on any
	// other cash account, such as a brokerage account's.
	Investment jurisdiction.ExemptionClass = "investment"
	// Other is every other asset: none of it is exempt. An asset the pack
	// can't classify counts as Other until the owner classifies it, so the
	// estimate errs high (D-132).
	Other jurisdiction.ExemptionClass = "other"
)

// OKIClasses returns OKI's exemption classes, savings first.
func OKIClasses() []jurisdiction.ExemptionClass {
	return []jurisdiction.ExemptionClass{Savings, Investment, Other}
}

// OKIExemptions are the exempt amounts of one tax year (art. 26 ust. 1 and
// 2): up to Savings of the average of savings assets and up to Investment
// of the average of investment assets, at most Total together, savings
// assets first.
type OKIExemptions struct {
	Savings    kernel.Money
	Investment kernel.Money
	Total      kernel.Money
}

// OKIAct is what the act on OKI fixes for its yearly tax on asset value
// (D-127), held as data: the day it is in force from, the rate and the
// exempt amounts of each tax year, and the readings that set later years'.
// The zero act is not in force: no tax is computed under it.
type OKIAct struct {
	InForceFrom kernel.Date
	// Rates are the rates the act fixes or the minister announced, by tax
	// year, as fractions: 0.85% for 2027 (art. 43) is 0.0085.
	Rates map[int]kernel.Decimal
	// ReferenceRates are the NBP reference rate's changes: a year with no
	// rate in Rates takes OKIRate of the reference rate in force on
	// 31 October of the year before (art. 25 ust. 1).
	ReferenceRates []RateChange
	// Exemptions are the exempt amounts of every year before IndexedFrom.
	Exemptions OKIExemptions
	// IndexedFrom is the first tax year whose exempt amounts rise with CPI
	// (art. 26 ust. 5, art. 44); CPI holds the index for each year from it:
	// the price index of the first three quarters of the year before over
	// the same quarters a year earlier, as a factor (103.5 is 1.035).
	IndexedFrom int
	CPI         map[int]kernel.Decimal
}

// The act's numbers: in force from 2027-01-01 (art. 46), 0.85% for 2027
// (art. 43), exempt amounts in PLN (art. 26 ust. 1 and 2) rising with CPI
// from the 2030 tax year (art. 44).
const (
	okiInForceYear = 2027
	okiFirstRate   = 85 // in ten-thousandths: 0.0085 is 0.85%
	okiIndexedFrom = 2030
	okiSavings     = 25000
	okiInvestment  = 100000
	// rateDay is the day of October whose reference rate sets the next
	// year's rate (art. 25 ust. 1).
	rateDay = 31
	// hundredsExponent rounds exempt amounts to whole hundreds of złoty.
	hundredsExponent = 2
)

// CurrentOKIAct returns the act on OKI (ustawa z dnia 3 lipca 2026 r. o
// osobistych kontach inwestycyjnych, Dz.U. 2026 poz. 1098) as this code
// knows it: in force from 2027-01-01, 0.85% for 2027, 25,000 PLN of savings
// assets and 100,000 PLN of investment assets exempt, at most 100,000 PLN
// together, rising with CPI from 2030. No later rate, reference rate or CPI
// index is known to it yet, so a later year's tax waits for them
// (docs/implementation-questions.md).
func CurrentOKIAct() OKIAct {
	from, err := kernel.NewDate(okiInForceYear, time.January, 1)
	if err != nil { // a constant valid date
		return OKIAct{}
	}
	return OKIAct{
		InForceFrom: from,
		Rates:       map[int]kernel.Decimal{okiInForceYear: kernel.NewDecimal(okiFirstRate, -4)},
		Exemptions: OKIExemptions{
			Savings:    kernel.NewMoney(kernel.NewDecimal(okiSavings, 0), pln()),
			Investment: kernel.NewMoney(kernel.NewDecimal(okiInvestment, 0), pln()),
			Total:      kernel.NewMoney(kernel.NewDecimal(okiInvestment, 0), pln()),
		},
		IndexedFrom: okiIndexedFrom,
	}
}

// OKIRate returns the rate of a tax year from the NBP reference rate in
// force on 31 October of the year before (art. 25 ust. 1): 19% of it, at
// least 0.1%, rounded down to two decimal places of a percent. With a
// reference rate of 5.75%, 19% is 1.0925%, so the rate is 1.09% (0.0109).
func OKIRate(reference kernel.Decimal) (kernel.Decimal, error) {
	const percentPlaces = 4 // two decimal places of a percent, as a fraction
	r, err := reference.Mul(BelkaRate()).Round(percentPlaces, kernel.RoundDown)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("OKI rate: %w", err)
	}
	if floor := kernel.NewDecimal(1, -3); r.Cmp(floor) < 0 {
		return floor, nil
	}
	return r, nil
}

// IndexOKIExemption returns an exempt amount raised by a CPI factor and
// rounded down to whole hundreds of złoty (art. 26 ust. 5): 25,000 PLN
// raised by 1.035 is 25,875, so 25,800 PLN.
func IndexOKIExemption(amount kernel.Money, cpi kernel.Decimal) (kernel.Money, error) {
	// Round takes no negative places with RoundDown: count in hundreds.
	hundreds, err := amount.Amount().Mul(cpi).Mul(kernel.NewDecimal(1, -hundredsExponent)).Round(0, kernel.RoundDown)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("OKI exemption: %w", err)
	}
	return kernel.NewMoney(hundreds.Mul(kernel.NewDecimal(1, hundredsExponent)), amount.Currency()), nil
}

// rate returns the rate of year, and false when it is not known.
func (a *OKIAct) rate(year int) (kernel.Decimal, bool, error) {
	if r, ok := a.Rates[year]; ok {
		return r, true, nil
	}
	if year <= a.InForceFrom.Year() {
		return kernel.Decimal{}, false, nil
	}
	day, err := kernel.NewDate(year-1, time.October, rateDay)
	if err != nil {
		return kernel.Decimal{}, false, fmt.Errorf("OKI rate of %d: %w", year, err)
	}
	ref, ok := RateInForce(a.ReferenceRates, day)
	if !ok {
		return kernel.Decimal{}, false, nil
	}
	r, err := OKIRate(ref)
	return r, err == nil, err
}

// exemptions returns the exempt amounts of year, and false when a CPI
// index it needs is not known.
func (a *OKIAct) exemptions(year int) (OKIExemptions, bool, error) {
	e := a.Exemptions
	if a.IndexedFrom == 0 {
		return e, true, nil
	}
	for y := a.IndexedFrom; y <= year; y++ {
		cpi, ok := a.CPI[y]
		if !ok {
			return OKIExemptions{}, false, nil
		}
		s, err1 := IndexOKIExemption(e.Savings, cpi)
		i, err2 := IndexOKIExemption(e.Investment, cpi)
		t, err3 := IndexOKIExemption(e.Total, cpi)
		if err := errors.Join(err1, err2, err3); err != nil {
			return OKIExemptions{}, false, err
		}
		e = OKIExemptions{Savings: s, Investment: i, Total: t}
	}
	return e, true, nil
}

// ExemptionClasses returns OKIClasses for OKI and none for the other wrappers,
// which have no yearly tax.
func (*Pack) ExemptionClasses(w jurisdiction.Wrapper) []jurisdiction.ExemptionClass {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return nil
	}
	return OKIClasses()
}

// ExemptionClass returns the class of a in OKI (art. 26 ust. 1): a treasury
// savings bond, the only holding valued at its accrued value, is a savings
// asset; PLN cash is a savings asset on an account that held such bonds in
// the year, the cash account of a bond account, and an investment asset on
// any other, a brokerage account's; cash in another currency is Other. The
// pack holds no data on what a holding priced at market is (a company's
// share capital, a fund's investment policy), so it can't tell for any of
// them: it counts as Other, so the estimate errs high, and ExemptionClass
// reports false (D-132). In the other wrappers no class is needed: the
// zero class and true.
func (*Pack) ExemptionClass(w jurisdiction.Wrapper, a *jurisdiction.Asset) (jurisdiction.ExemptionClass, bool) {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return "", true
	}
	switch a.Kind {
	case jurisdiction.AssetAccrued:
		return Savings, true
	case jurisdiction.AssetCash:
		switch {
		case !a.Currency.Equal(pln()):
			return Other, true
		case a.AccountHoldsAccrued:
			return Savings, true
		default:
			return Investment, true
		}
	case jurisdiction.AssetListed:
		return Other, false
	}
	return Other, false
}

// YearlyTax returns OKI's tax on asset value for year (D-127) under p.OKI,
// from the days each of the owner's OKI accounts was held in the year on
// or after the act's in-force date:
//
//   - Each account's average value (art. 23 ust. 2 and 4) is the sum of its
//     end-of-day values plus, for each day, its payments in up to that
//     day's withdrawals, divided by the days it was held. Split by asset
//     class (art. 26 ust. 3), each class's average is its own values' sum
//     plus the payments in the proportion of that sum to the account's,
//     over the same days; an account with no value all year puts its
//     payments in Other. Each class's average is rounded half up to grosze.
//   - The base is the sum of the accounts' averages. Savings assets are
//     exempt up to the year's savings amount, then investment assets up to
//     their amount and to what is left of the total (art. 26 ust. 2); the
//     rest is taxed at the year's rate, rounded half up to grosze.
//
// A year the act isn't in force in gets no tax and InForce false; a year
// whose exempt amounts or rate are not known yet gets the averages only.
// Any wrapper but OKI is an error wrapping jurisdiction.ErrNoYearlyTax.
func (p *Pack) YearlyTax(w jurisdiction.Wrapper, year int, accounts []jurisdiction.TaxAccount) (jurisdiction.YearlyTax, error) {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return jurisdiction.YearlyTax{}, fmt.Errorf("%w: %q", jurisdiction.ErrNoYearlyTax, w)
	}
	out := jurisdiction.YearlyTax{Wrapper: w, Year: year}
	act := &p.OKI
	if act.InForceFrom.IsZero() || act.InForceFrom.Year() > year {
		return out, nil
	}
	out.InForce = true
	sums, err := sumAverages(year, act.InForceFrom, accounts, &out)
	if err != nil {
		return jurisdiction.YearlyTax{}, err
	}
	total := kernel.Decimal{}
	for _, c := range OKIClasses() {
		out.Averages = append(out.Averages, classValue(c, sums[c]))
		total = total.Add(sums[c])
	}
	out.Average = kernel.NewMoney(total, pln())
	if exemptErr := p.exempt(year, sums, total, &out); exemptErr != nil || !out.ExemptionsKnown {
		return out, exemptErr
	}
	if taxErr := act.tax(year, &out); taxErr != nil {
		return jurisdiction.YearlyTax{}, taxErr
	}
	return out, nil
}

// tax sets out's rate and tax on its taxed part when year's rate is known.
func (a *OKIAct) tax(year int, out *jurisdiction.YearlyTax) error {
	r, known, err := a.rate(year)
	if err != nil || !known {
		return err
	}
	tax, err := taxAt(out.Taxed.Amount(), r)
	if err != nil {
		return err
	}
	out.Rate, out.RateKnown, out.Tax = r, true, kernel.NewMoney(tax, pln())
	return nil
}

// exempt sets out's exempt part of each class's average and the part of
// total left to tax, under year's exempt amounts when they are known.
func (p *Pack) exempt(year int, sums map[jurisdiction.ExemptionClass]kernel.Decimal, total kernel.Decimal, out *jurisdiction.YearlyTax) error {
	e, known, err := p.OKI.exemptions(year)
	if err != nil || !known {
		return err
	}
	savings := clamp(sums[Savings], e.Savings.Amount())
	investment := clamp(sums[Investment], minDecimal(e.Investment.Amount(), e.Total.Amount().Sub(savings)))
	taxed := total.Sub(savings).Sub(investment)
	if taxed.Sign() < 0 {
		taxed = kernel.Decimal{}
	}
	out.ExemptionsKnown = true
	out.Exempt = []jurisdiction.ClassValue{
		classValue(Savings, savings), classValue(Investment, investment), classValue(Other, kernel.Decimal{}),
	}
	out.Taxed = kernel.NewMoney(taxed, pln())
	return nil
}

// sumAverages returns the sum over accounts of each one's average value
// per class.
func sumAverages(year int, from kernel.Date, accounts []jurisdiction.TaxAccount, out *jurisdiction.YearlyTax) (map[jurisdiction.ExemptionClass]kernel.Decimal, error) {
	sums := classSums()
	for i := range accounts {
		avg, err := accountAverages(year, from, &accounts[i], out)
		if err != nil {
			return nil, err
		}
		for c, v := range avg {
			sums[c] = sums[c].Add(v)
		}
	}
	return sums, nil
}

// accountAverages returns one account's average value per class over its
// days in year on or after from, and counts its days and last day in out.
func accountAverages(year int, from kernel.Date, a *jurisdiction.TaxAccount, out *jurisdiction.YearlyTax) (map[jurisdiction.ExemptionClass]kernel.Decimal, error) {
	sums := classSums()
	total, paid, days := kernel.Decimal{}, kernel.Decimal{}, 0
	seen := map[kernel.Date]bool{}
	for i := range a.Days {
		d := &a.Days[i]
		if d.Date.Year() != year || d.Date.Before(from) {
			continue
		}
		if seen[d.Date] {
			return nil, fmt.Errorf("%w: %s: %s given twice", ErrTaxDay, a.Account, d.Date)
		}
		seen[d.Date] = true
		added, err := addDay(sums, d)
		if err != nil {
			return nil, fmt.Errorf("%s on %s: %w", a.Account, d.Date, err)
		}
		total, paid, days = total.Add(added.value), paid.Add(added.counted), days+1
		if d.Date.After(out.Through) {
			out.Through = d.Date
		}
	}
	out.Days = max(out.Days, days)
	if days == 0 {
		return sums, nil
	}
	return averages(sums, total, paid, days)
}

// dayAdded is what one day adds to an account's average: its counted
// payments and its value.
type dayAdded struct {
	counted, value kernel.Decimal
}

// addDay adds d's values to sums by class and returns what d adds.
func addDay(sums map[jurisdiction.ExemptionClass]kernel.Decimal, d *jurisdiction.TaxDay) (dayAdded, error) {
	counted, err := countedPayments(d)
	if err != nil {
		return dayAdded{}, err
	}
	out := dayAdded{counted: counted}
	for _, v := range d.Values {
		if _, ok := sums[v.Class]; !ok {
			return dayAdded{}, fmt.Errorf("%w: class %q", ErrTaxDay, v.Class)
		}
		if pErr := allPLN(v.Value); pErr != nil {
			return dayAdded{}, fmt.Errorf("%w: %w", ErrTaxDay, pErr)
		}
		sums[v.Class] = sums[v.Class].Add(v.Value.Amount())
		out.value = out.value.Add(v.Value.Amount())
	}
	return out, nil
}

// averages divides each class's sum, plus its share of the payments, by
// days: the payments go to the classes in proportion to their sums, or all
// to Other when the sums add up to zero.
func averages(sums map[jurisdiction.ExemptionClass]kernel.Decimal, total, paid kernel.Decimal, days int) (map[jurisdiction.ExemptionClass]kernel.Decimal, error) {
	n := kernel.NewDecimal(int64(days), 0)
	out := classSums()
	for c, s := range sums {
		num, den := s.Mul(total.Add(paid)), total.Mul(n)
		if total.IsZero() {
			num, den = s, n
			if c == Other {
				num = s.Add(paid)
			}
		}
		avg, err := num.Quo(den, grosze, kernel.RoundHalfUp)
		if err != nil {
			return nil, fmt.Errorf("OKI average of %s: %w", c, err)
		}
		out[c] = avg
	}
	return out, nil
}

// countedPayments returns the payments of d that count (art. 23 ust. 4):
// those up to the day's withdrawals, none on a day without withdrawals.
func countedPayments(d *jurisdiction.TaxDay) (kernel.Decimal, error) {
	if err := allPLN(d.In, d.Out); err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: %w", ErrTaxDay, err)
	}
	in, out := d.In.Amount(), d.Out.Amount()
	if in.Sign() < 0 || out.Sign() < 0 {
		return kernel.Decimal{}, fmt.Errorf("%w: negative payments %s in, %s out", ErrTaxDay, in, out)
	}
	return minDecimal(in, out), nil
}

// classSums returns a zero sum for each of OKI's classes.
func classSums() map[jurisdiction.ExemptionClass]kernel.Decimal {
	out := map[jurisdiction.ExemptionClass]kernel.Decimal{}
	for _, c := range OKIClasses() {
		out[c] = kernel.Decimal{}
	}
	return out
}

// classValue returns amount of class c in PLN.
func classValue(c jurisdiction.ExemptionClass, amount kernel.Decimal) jurisdiction.ClassValue {
	return jurisdiction.ClassValue{Class: c, Value: kernel.NewMoney(amount, pln())}
}

// clamp returns v limited to 0..limit.
func clamp(v, limit kernel.Decimal) kernel.Decimal {
	if v.Sign() <= 0 || limit.Sign() <= 0 {
		return kernel.Decimal{}
	}
	return minDecimal(v, limit)
}

// minDecimal returns the smaller of a and b.
func minDecimal(a, b kernel.Decimal) kernel.Decimal {
	return slices.MinFunc([]kernel.Decimal{a, b}, kernel.Decimal.Cmp)
}
