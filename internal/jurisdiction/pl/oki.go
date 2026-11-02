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

// ErrOKIAct reports OKI act data the pack refuses: exempt amounts announced
// for a year before the act announces any, not in PLN, or lower than the
// year before's, since they never fall (D-145).
var ErrOKIAct = errors.New("invalid OKI act data")

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

// The forms of an OKI account the owner sets when creating it (D-147). The
// form decides the class of its PLN cash (art. 26 ust. 1).
const (
	// FormBank is a bank account: its PLN cash is a savings asset.
	FormBank jurisdiction.AccountForm = "bank-account"
	// FormTreasuryBonds is a treasury savings bond account: the PLN cash
	// on its cash account is a savings asset.
	FormTreasuryBonds jurisdiction.AccountForm = "treasury-bond-account"
	// FormBrokerage is a brokerage account: its PLN cash is an investment
	// asset.
	FormBrokerage jurisdiction.AccountForm = "brokerage-account"
	// FormFundRegister is a fund register: the PLN cash on its cash
	// account is an investment asset.
	FormFundRegister jurisdiction.AccountForm = "fund-register"
)

// OKIClasses returns OKI's exemption classes, savings first.
func OKIClasses() []jurisdiction.ExemptionClass {
	return []jurisdiction.ExemptionClass{Savings, Investment, Other}
}

// OKIForms returns the forms of an OKI account (D-147), in the order the
// act names them.
func OKIForms() []jurisdiction.AccountForm {
	return []jurisdiction.AccountForm{FormBank, FormTreasuryBonds, FormBrokerage, FormFundRegister}
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
// (D-127), held as data: the day it is in force from, and the rate and the
// exempt amounts of each tax year. The rates from 2028 and the exempt
// amounts from 2030 are the finance minister's announcements, added to this
// data with each release: OpenFief never computes them (D-145). The zero
// act is not in force: no tax is computed under it.
type OKIAct struct {
	InForceFrom kernel.Date
	// Rates are the rates by tax year, as fractions: 2027's fixed by the
	// act (art. 43), 0.85% is 0.0085, and each later year's as the
	// minister announced it (art. 25 ust. 2). A year without one has no
	// tax yet.
	Rates map[int]kernel.Decimal
	// Exemptions are the act's exempt amounts (art. 26 ust. 1 and 2), for
	// every year before AnnouncedFrom.
	Exemptions OKIExemptions
	// AnnouncedFrom is the first tax year whose exempt amounts rise with
	// CPI (art. 26 ust. 5, art. 44), which the minister announces;
	// Announced holds each such year's as announced. A year from
	// AnnouncedFrom without them has no exemptions yet.
	AnnouncedFrom int
	Announced     map[int]OKIExemptions
}

// The act's numbers: in force from 2027-01-01 (art. 46), 0.85% for 2027
// (art. 43), exempt amounts in PLN (art. 26 ust. 1 and 2), announced raised
// with CPI from the 2030 tax year (art. 44).
const (
	okiInForceYear   = 2027
	okiFirstRate     = 85 // in ten-thousandths: 0.0085 is 0.85%
	okiAnnouncedFrom = 2030
	okiSavings       = 25000
	okiInvestment    = 100000
)

// CurrentOKIAct returns the act on OKI (ustawa z dnia 3 lipca 2026 r. o
// osobistych kontach inwestycyjnych, Dz.U. 2026 poz. 1098) with the
// minister's announcements as this release knows them: in force from
// 2027-01-01, 0.85% for 2027, 25,000 PLN of savings assets and 100,000 PLN
// of investment assets exempt, at most 100,000 PLN together, announced from
// 2030. No rate for 2028 or later and no exempt amounts for 2030 or later
// are announced yet, so such a year's tax waits for a release that adds
// them (D-145).
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
		AnnouncedFrom: okiAnnouncedFrom,
	}
}

// Check returns an error wrapping ErrOKIAct unless a's announced exempt
// amounts are for years from AnnouncedFrom, in PLN, and never lower than
// those of the latest year before them, the act's own amounts first: the
// exempt amounts never fall (D-145).
func (a *OKIAct) Check() error {
	years := make([]int, 0, len(a.Announced))
	for y := range a.Announced {
		years = append(years, y)
	}
	slices.Sort(years)
	prev := a.Exemptions
	for _, y := range years {
		e := a.Announced[y]
		if y < a.AnnouncedFrom {
			return fmt.Errorf("%w: exempt amounts announced for %d, before %d", ErrOKIAct, y, a.AnnouncedFrom)
		}
		if err := allPLN(e.Savings, e.Investment, e.Total); err != nil {
			return fmt.Errorf("%w: %d: %w", ErrOKIAct, y, err)
		}
		if e.Savings.Amount().Cmp(prev.Savings.Amount()) < 0 || e.Investment.Amount().Cmp(prev.Investment.Amount()) < 0 ||
			e.Total.Amount().Cmp(prev.Total.Amount()) < 0 {
			return fmt.Errorf("%w: the exempt amounts of %d fall below the year before's", ErrOKIAct, y)
		}
		prev = e
	}
	return nil
}

// rate returns the rate of year, and false when it is not known.
func (a *OKIAct) rate(year int) (kernel.Decimal, bool) {
	r, ok := a.Rates[year]
	return r, ok
}

// exemptions returns the exempt amounts of year, and false when the
// minister's announcement for it is not known.
func (a *OKIAct) exemptions(year int) (OKIExemptions, bool) {
	if a.AnnouncedFrom == 0 || year < a.AnnouncedFrom {
		return a.Exemptions, true
	}
	e, ok := a.Announced[year]
	return e, ok
}

// ExemptionClasses returns OKIClasses for OKI and none for the other wrappers,
// which have no yearly tax.
func (*Pack) ExemptionClasses(w jurisdiction.Wrapper) []jurisdiction.ExemptionClass {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return nil
	}
	return OKIClasses()
}

// AccountForms returns OKIForms for OKI and none for the other wrappers,
// whose rules don't tell forms apart.
func (*Pack) AccountForms(w jurisdiction.Wrapper) []jurisdiction.AccountForm {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return nil
	}
	return OKIForms()
}

// ExemptionClass returns the class of a in OKI (art. 26 ust. 1): a treasury
// savings bond, the only holding valued at its accrued value, is a savings
// asset; the account's form decides its PLN cash (D-147): a savings asset on
// a bank account or a treasury bond account, an investment asset on a
// brokerage account or a fund register, and Other on an account whose form
// isn't set, so the estimate errs high; cash in another currency is Other.
// The pack holds no data on what a holding priced at market is (a company's
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
		if !a.Currency.Equal(pln()) {
			return Other, true
		}
		return cashClass(a.Form), true
	case jurisdiction.AssetListed:
		return Other, false
	}
	return Other, false
}

// cashClass returns the class of PLN cash on an OKI account of form f.
func cashClass(f jurisdiction.AccountForm) jurisdiction.ExemptionClass {
	switch f {
	case FormBank, FormTreasuryBonds:
		return Savings
	case FormBrokerage, FormFundRegister:
		return Investment
	}
	return Other
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
//     payments in Other. Every average is kept exact (D-144).
//   - The base is the sum of the accounts' averages. Savings assets are
//     exempt up to the year's savings amount, then investment assets up to
//     their amount and to what is left of the total (art. 26 ust. 2); the
//     rest is the tax base, rounded to whole złoty, and the tax is the
//     year's rate of it, rounded to whole złoty too, both under 50 groszy
//     down and from 50 groszy up, as the tax ordinance rounds bases and
//     taxes (art. 63 § 1, D-144).
//
// The averages and exempt parts it reports are the exact ones shown to the
// grosz, rounded half up; Taxed is the whole-złoty base. A year counted
// through a day before its last, such as the current year, is an estimate
// over the days so far (Estimate, D-146). A year the act isn't in force in
// gets no tax and InForce false; a year whose exempt amounts or rate the
// minister hasn't announced yet gets the averages only (D-145). Act data
// that breaks OKIAct.Check is an error wrapping ErrOKIAct, and any wrapper
// but OKI an error wrapping jurisdiction.ErrNoYearlyTax.
func (p *Pack) YearlyTax(w jurisdiction.Wrapper, year int, accounts []jurisdiction.TaxAccount) (jurisdiction.YearlyTax, error) {
	if w != jurisdiction.Wrapper(OKI.String()) {
		return jurisdiction.YearlyTax{}, fmt.Errorf("%w: %q", jurisdiction.ErrNoYearlyTax, w)
	}
	out := jurisdiction.YearlyTax{Wrapper: w, Year: year}
	act := &p.OKI
	if act.InForceFrom.IsZero() || act.InForceFrom.Year() > year {
		return out, nil
	}
	if err := act.Check(); err != nil {
		return jurisdiction.YearlyTax{}, err
	}
	out.InForce = true
	if err := act.yearlyTax(year, accounts, &out); err != nil {
		return jurisdiction.YearlyTax{}, err
	}
	return out, nil
}

// yearlyTax sets out's days, averages, exempt parts, base, rate and tax for
// year under a, as far as a knows the year's exempt amounts and rate.
func (a *OKIAct) yearlyTax(year int, accounts []jurisdiction.TaxAccount, out *jurisdiction.YearlyTax) error {
	sums, err := sumAverages(year, a.InForceFrom, accounts, out)
	if err != nil {
		return err
	}
	if out.Estimate, err = estimate(year, out); err != nil {
		return err
	}
	total, err := averaged(sums, out)
	if err != nil {
		return err
	}
	e, known := a.exemptions(year)
	if !known {
		return nil
	}
	base, err := exempt(&e, sums, total, out)
	if err != nil {
		return err
	}
	if r, rateKnown := a.rate(year); rateKnown {
		return taxWhole(base, r, out)
	}
	return nil
}

// averaged sets out's averages per class and their sum from sums, shown to
// the grosz, and returns the exact sum.
func averaged(sums map[jurisdiction.ExemptionClass]ratio, out *jurisdiction.YearlyTax) (ratio, error) {
	total := zeroRatio()
	for _, c := range OKIClasses() {
		cv, err := classValue(c, sums[c])
		if err != nil {
			return ratio{}, err
		}
		out.Averages = append(out.Averages, cv)
		total = total.add(sums[c])
	}
	avg, err := shown(total)
	if err != nil {
		return ratio{}, err
	}
	out.Average = avg
	return total, nil
}

// estimate reports whether out, counted through out.Through, ends before
// year's last day: the tax of a year counted over its days so far (D-146).
func estimate(year int, out *jurisdiction.YearlyTax) (bool, error) {
	if out.Days == 0 {
		return false, nil
	}
	last, err := kernel.NewDate(year, time.December, lastDayOfDecember)
	if err != nil {
		return false, fmt.Errorf("the last day of %d: %w", year, err)
	}
	return out.Through.Before(last), nil
}

// lastDayOfDecember is the day of the year's last day.
const lastDayOfDecember = 31

// taxWhole sets out's rate and tax: rate of base, both whole złoty,
// rounded under 50 groszy down and from 50 up (D-144).
func taxWhole(base, rate kernel.Decimal, out *jurisdiction.YearlyTax) error {
	tax := kernel.Decimal{}
	if base.Sign() > 0 {
		var err error
		if tax, err = base.Mul(rate).Round(0, kernel.RoundHalfUp); err != nil {
			return fmt.Errorf("OKI tax: %w", err)
		}
	}
	out.Rate, out.RateKnown, out.Tax = rate, true, kernel.NewMoney(tax, pln())
	return nil
}

// exempt sets out's exempt part of each class's average and its tax base,
// the part of total left to tax rounded to whole złoty (D-144), under the
// year's exempt amounts e, and returns the base.
func exempt(e *OKIExemptions, sums map[jurisdiction.ExemptionClass]ratio, total ratio, out *jurisdiction.YearlyTax) (kernel.Decimal, error) {
	savings := sums[Savings].clamp(ratioOf(e.Savings.Amount()))
	investmentLimit := ratioOf(e.Investment.Amount()).min(ratioOf(e.Total.Amount()).sub(savings))
	investment := sums[Investment].clamp(investmentLimit)
	taxed := total.sub(savings).sub(investment)
	if taxed.sign() < 0 {
		taxed = zeroRatio()
	}
	base, err := taxed.round(0, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("OKI tax base: %w", err)
	}
	s, err1 := classValue(Savings, savings)
	i, err2 := classValue(Investment, investment)
	o, err3 := classValue(Other, zeroRatio())
	if err := errors.Join(err1, err2, err3); err != nil {
		return kernel.Decimal{}, err
	}
	out.ExemptionsKnown = true
	out.Exempt = []jurisdiction.ClassValue{s, i, o}
	out.Taxed = kernel.NewMoney(base, pln())
	return base, nil
}

// sumAverages returns the sum over accounts of each one's exact average
// value per class.
func sumAverages(year int, from kernel.Date, accounts []jurisdiction.TaxAccount, out *jurisdiction.YearlyTax) (map[jurisdiction.ExemptionClass]ratio, error) {
	sums := classRatios()
	for i := range accounts {
		avg, err := accountAverages(year, from, &accounts[i], out)
		if err != nil {
			return nil, err
		}
		for c, v := range avg {
			sums[c] = sums[c].add(v)
		}
	}
	return sums, nil
}

// accountAverages returns one account's exact average value per class over
// its days in year on or after from, and counts its days and last day in
// out.
func accountAverages(year int, from kernel.Date, a *jurisdiction.TaxAccount, out *jurisdiction.YearlyTax) (map[jurisdiction.ExemptionClass]ratio, error) {
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
		return classRatios(), nil
	}
	return averages(sums, total, paid, days), nil
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
// days, exactly: the payments go to the classes in proportion to their
// sums, or all to Other when the sums add up to zero.
func averages(sums map[jurisdiction.ExemptionClass]kernel.Decimal, total, paid kernel.Decimal, days int) map[jurisdiction.ExemptionClass]ratio {
	n := kernel.NewDecimal(int64(days), 0)
	out := classRatios()
	for c, s := range sums {
		num, den := s.Mul(total.Add(paid)), total.Mul(n)
		if total.IsZero() {
			num, den = s, n
			if c == Other {
				num = s.Add(paid)
			}
		}
		out[c] = ratio{num: num, den: den}.normal()
	}
	return out
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

// classRatios returns an exact zero for each of OKI's classes.
func classRatios() map[jurisdiction.ExemptionClass]ratio {
	out := map[jurisdiction.ExemptionClass]ratio{}
	for _, c := range OKIClasses() {
		out[c] = zeroRatio()
	}
	return out
}

// classValue returns the exact amount r of class c in PLN, shown to the
// grosz.
func classValue(c jurisdiction.ExemptionClass, r ratio) (jurisdiction.ClassValue, error) {
	v, err := shown(r)
	if err != nil {
		return jurisdiction.ClassValue{}, fmt.Errorf("OKI %s: %w", c, err)
	}
	return jurisdiction.ClassValue{Class: c, Value: v}, nil
}

// shown returns the exact amount r in PLN rounded half up to the grosz, as
// the yearly tax reports it.
func shown(r ratio) (kernel.Money, error) {
	d, err := r.round(grosze, kernel.RoundHalfUp)
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(d, pln()), nil
}

// minDecimal returns the smaller of a and b.
func minDecimal(a, b kernel.Decimal) kernel.Decimal {
	return slices.MinFunc([]kernel.Decimal{a, b}, kernel.Decimal.Cmp)
}
