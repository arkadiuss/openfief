package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// The bench's stand-ins for what the core's valuation worker computes. They
// are fixed and public, so the numbers a question reads never change: flat FX
// rates to the reporting currency (PLN), a bond at its nominal of 100 PLN, and
// the net and liquidation values as fixed shares of the gross one.
const (
	reporting     = "PLN"
	bondNominal   = 100
	usdRatePLN    = "4.00"
	eurRatePLN    = "4.30"
	netShare      = "0.97"
	liquidShare   = "0.995"
	indexStart    = 100
	indexPlaces   = 4
	moneyPlaces   = 2
	cashAssetName = "cash"
	assetBond     = "bond"
	// quotientPlaces is how many places a day's return is kept to.
	quotientPlaces = 8
)

// holdingValue is one holding's value on a day.
type holdingValue struct {
	Account  ledger.AccountID
	ISIN     string
	Quantity kernel.Decimal
	// Native is the value in the instrument's currency and Reporting in PLN.
	Native    kernel.Decimal
	Currency  string
	Reporting kernel.Decimal
}

// dayValues is the portfolio on one day: every account's holdings and cash
// valued, and the money that flowed in or out that day.
type dayValues struct {
	Date     kernel.Date
	Holdings []holdingValue
	// Cash is each account's cash in PLN, by account.
	Cash map[ledger.AccountID]kernel.Decimal
	// CashByCurrency is each account's cash by currency code, unconverted.
	CashByCurrency map[ledger.AccountID]map[string]kernel.Decimal
	// Flow is the deposits less the withdrawals of the day, in PLN.
	Flow kernel.Decimal
	// Index is the time-weighted return index at the end of the day.
	Index kernel.Decimal
}

// String names a valued day by its date and gross value.
func (d *dayValues) String() string { return d.Date.String() + " " + d.Gross().String() }

// Gross returns the day's gross value: holdings plus cash, in PLN.
func (d *dayValues) Gross() kernel.Decimal {
	var sum kernel.Decimal
	for i := range d.Holdings {
		sum = sum.Add(d.Holdings[i].Reporting)
	}
	for _, c := range d.Cash {
		sum = sum.Add(c)
	}
	return sum
}

// AccountGross returns one account's gross value on the day.
func (d *dayValues) AccountGross(id ledger.AccountID) kernel.Decimal {
	sum := d.Cash[id]
	for i := range d.Holdings {
		if d.Holdings[i].Account == id {
			sum = sum.Add(d.Holdings[i].Reporting)
		}
	}
	return sum
}

// Portfolio is the seed valued on every day from go-live to the bench's day.
type Portfolio struct {
	Seed *Seed
	// Days are the valued days, one per calendar day, oldest first.
	Days []dayValues
}

// priceKey is an instrument's ISIN and a day.
type priceTable map[string][]pricePoint

type pricePoint struct {
	Date  kernel.Date
	Close kernel.Decimal
}

// ledgerBooks is the cash and holdings the replayed transactions leave.
type ledgerBooks struct {
	cash  map[ledger.AccountID]map[string]kernel.Decimal
	units map[ledger.AccountID]map[string]kernel.Decimal
}

func newBooks() *ledgerBooks {
	return &ledgerBooks{cash: map[ledger.AccountID]map[string]kernel.Decimal{}, units: map[ledger.AccountID]map[string]kernel.Decimal{}}
}

// apply books t's cash and holding postings.
func (b *ledgerBooks) apply(t *ledger.Transaction) {
	for i := range t.Postings {
		p := &t.Postings[i]
		switch p.Book.Kind {
		case ledger.BookCash:
			addTo(b.cash, t.Account, p.Book.Currency.String(), p.Amount.Amount())
		case ledger.BookHolding:
			addTo(b.units, t.Account, p.Book.Instrument.ISIN.String(), p.Quantity.Decimal())
		case ledger.BookFee, ledger.BookWithholdingTax, ledger.BookFX, ledger.BookEquity, ledger.BookIncome:
		}
	}
}

// addTo adds v to books[account][key].
func addTo(books map[ledger.AccountID]map[string]kernel.Decimal, account ledger.AccountID, key string, v kernel.Decimal) {
	if books[account] == nil {
		books[account] = map[string]kernel.Decimal{}
	}
	books[account][key] = books[account][key].Add(v)
}

// fxRate returns the PLN value of one unit of currency.
func fxRate(currency string) (kernel.Decimal, error) {
	switch currency {
	case reporting:
		return kernel.NewDecimal(1, 0), nil
	case "USD":
		return parseConstant(usdRatePLN)
	case "EUR":
		return parseConstant(eurRatePLN)
	}
	return kernel.Decimal{}, fmt.Errorf("%w: no rate for %q", errSeed, currency)
}

// parseConstant reads one of this file's decimal constants.
func parseConstant(text string) (kernel.Decimal, error) {
	d, err := kernel.ParseDecimal(text)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: the constant %q: %w", errSeed, text, err)
	}
	return d, nil
}

// NewPortfolio replays the seed's transactions and values every day from
// go-live to today with the dataset's closes: a holding at the latest close on
// or before the day, a bond at its nominal, cash as it stands.
func NewPortfolio(s *Seed) (*Portfolio, error) {
	prices := priceTableOf(s.Dataset)
	books := newBooks()
	p := &Portfolio{Seed: s}
	next := 0
	prev := kernel.NewDecimal(indexStart, 0)
	var prevGross kernel.Decimal
	for day := s.Dataset.GoLive; !day.After(s.Today); day = day.AddDays(1) {
		var flow kernel.Decimal
		for ; next < len(s.Transactions) && !s.Transactions[next].Date.After(day); next++ {
			t := &s.Transactions[next]
			books.apply(t)
			f, err := flowOf(t)
			if err != nil {
				return nil, err
			}
			flow = flow.Add(f)
		}
		dv, err := valueDay(s, books, prices, day)
		if err != nil {
			return nil, err
		}
		dv.Flow = flow
		dv.Index = prev
		if gross := dv.Gross(); !prevGross.IsZero() {
			if dv.Index, err = chain(prev, gross.Sub(flow), prevGross); err != nil {
				return nil, err
			}
		}
		prev, prevGross = dv.Index, dv.Gross()
		p.Days = append(p.Days, dv)
	}
	return p, nil
}

// chain returns the index prev moved by the day's return, (value less the
// day's flow) over the previous day's value, to indexPlaces.
func chain(prev, valueLessFlow, previous kernel.Decimal) (kernel.Decimal, error) {
	r, err := valueLessFlow.Quo(previous, quotientPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: a return: %w", errSeed, err)
	}
	out, err := prev.Mul(r).Round(indexPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: an index: %w", errSeed, err)
	}
	return out, nil
}

// flowOf returns the money t moved into or out of the portfolio, in PLN: a
// deposit or a withdrawal that is not an account's seeding.
func flowOf(t *ledger.Transaction) (kernel.Decimal, error) {
	if (t.Kind != ledger.KindDeposit && t.Kind != ledger.KindWithdrawal) || ledger.IsSeeding(t) {
		return kernel.Decimal{}, nil
	}
	var sum kernel.Decimal
	for i := range t.Postings {
		p := &t.Postings[i]
		if p.Book.Kind != ledger.BookCash {
			continue
		}
		rate, err := fxRate(p.Book.Currency.String())
		if err != nil {
			return kernel.Decimal{}, err
		}
		sum = sum.Add(p.Amount.Amount().Mul(rate))
	}
	return sum, nil
}

// priceTableOf returns the dataset's closes by ISIN, oldest first.
func priceTableOf(ds *demo.Dataset) priceTable {
	out := priceTable{}
	for i := range ds.Prices {
		pr := &ds.Prices[i]
		isin, _, _ := strings.Cut(pr.Instrument, "/")
		out[isin] = append(out[isin], pricePoint{Date: pr.Date, Close: pr.Close.Amount()})
	}
	return out
}

// closeOn returns the latest close of isin on or before day.
func (t priceTable) closeOn(isin string, day kernel.Date) (kernel.Decimal, bool) {
	pts := t[isin]
	i := sort.Search(len(pts), func(i int) bool { return pts[i].Date.After(day) })
	if i == 0 {
		return kernel.Decimal{}, false
	}
	return pts[i-1].Close, true
}

// valueDay values the books on day.
func valueDay(s *Seed, books *ledgerBooks, prices priceTable, day kernel.Date) (dayValues, error) {
	dv := dayValues{Date: day, Cash: map[ledger.AccountID]kernel.Decimal{}, CashByCurrency: map[ledger.AccountID]map[string]kernel.Decimal{}}
	for _, acc := range sortedAccountIDs(books) {
		for _, isin := range sortedKeys(books.units[acc]) {
			q := books.units[acc][isin]
			if q.IsZero() {
				continue
			}
			hv, err := valueHolding(s, prices, acc, isin, q, day)
			if err != nil {
				return dayValues{}, err
			}
			dv.Holdings = append(dv.Holdings, hv)
		}
		cash := map[string]kernel.Decimal{}
		var total kernel.Decimal
		for _, cur := range sortedKeys(books.cash[acc]) {
			amount := books.cash[acc][cur]
			rate, err := fxRate(cur)
			if err != nil {
				return dayValues{}, err
			}
			cash[cur] = amount
			total = total.Add(amount.Mul(rate))
		}
		dv.CashByCurrency[acc], dv.Cash[acc] = cash, round2(total)
	}
	return dv, nil
}

// valueHolding values q units of isin held in acc on day.
func valueHolding(s *Seed, prices priceTable, acc ledger.AccountID, isin string, q kernel.Decimal, day kernel.Date) (holdingValue, error) {
	info, ok := s.Instruments[isin]
	if !ok {
		return holdingValue{}, fmt.Errorf("%w: a holding of an unknown instrument", errSeed)
	}
	closePrice := kernel.NewDecimal(bondNominal, 0)
	if info.AssetClass != assetBond {
		c, found := prices.closeOn(isin, day)
		if !found {
			return holdingValue{}, fmt.Errorf("%w: no close of %s on %s", errSeed, info.Symbol, day)
		}
		closePrice = c
	}
	rate, err := fxRate(info.Currency)
	if err != nil {
		return holdingValue{}, err
	}
	native := round2(q.Mul(closePrice))
	return holdingValue{
		Account: acc, ISIN: isin, Quantity: q, Native: native, Currency: info.Currency, Reporting: round2(native.Mul(rate)),
	}, nil
}

// round2 rounds d to hundredths, half to even. The mode is a known one, so
// it can't fail; d is returned unrounded if it somehow does.
func round2(d kernel.Decimal) kernel.Decimal {
	r, err := d.Round(moneyPlaces, kernel.RoundHalfEven)
	if err != nil {
		return d
	}
	return r
}

// sortedAccountIDs returns the accounts the books hold anything for.
func sortedAccountIDs(b *ledgerBooks) []ledger.AccountID {
	seen := map[ledger.AccountID]bool{}
	for a := range b.cash {
		seen[a] = true
	}
	for a := range b.units {
		seen[a] = true
	}
	out := make([]ledger.AccountID, 0, len(seen))
	for a := range seen {
		out = append(out, a)
	}
	slices.Sort(out)
	return out
}

// sortedKeys returns the keys of m in order.
func sortedKeys(m map[string]kernel.Decimal) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// On returns the valued day for d: the first day for one before go-live, the
// last for one after today; ok is false before go-live.
func (p *Portfolio) On(d kernel.Date) (*dayValues, bool) {
	if len(p.Days) == 0 || d.Before(p.Days[0].Date) || d.After(p.Days[len(p.Days)-1].Date) {
		return nil, false
	}
	return &p.Days[p.Days[0].Date.DaysUntil(d)], true
}

// Net returns a day's net value: a fixed share of its gross one.
func Net(gross kernel.Decimal) kernel.Decimal { return share2(gross, netShare) }

// Liquidation returns a day's liquidation value: a fixed share of its gross.
func Liquidation(gross kernel.Decimal) kernel.Decimal { return share2(gross, liquidShare) }

// share2 returns gross times the decimal share, to hundredths.
func share2(gross kernel.Decimal, share string) kernel.Decimal {
	f, err := kernel.ParseDecimal(share)
	if err != nil { // the shares are constants
		return gross
	}
	return round2(gross.Mul(f))
}

// DropMonth returns the calendar month, as "YYYY-MM", in which the
// time-weighted return index fell most from the last day of the month before
// to the last day of the month, and by how many index points. Deposits keep the
// gross value rising through a fall, so the index, which leaves them out, is
// what shows a month the portfolio lost. ok is false when no month fell.
func (p *Portfolio) DropMonth() (string, kernel.Decimal, bool) {
	ends := map[string]kernel.Decimal{}
	var order []string
	for i := range p.Days {
		d := &p.Days[i]
		if d.Date.AddDays(1).Month() != d.Date.Month() { // the last day of a month
			key := monthKey(d.Date)
			ends[key] = d.Index
			order = append(order, key)
		}
	}
	var month string
	var drop kernel.Decimal
	found := false
	for i := 1; i < len(order); i++ {
		change := ends[order[i]].Sub(ends[order[i-1]])
		if change.Sign() < 0 && (!found || change.Cmp(drop) < 0) {
			month, drop, found = order[i], change, true
		}
	}
	return month, drop, found
}

// monthKey returns d's month as "YYYY-MM".
func monthKey(d kernel.Date) string {
	return fmt.Sprintf("%04d-%02d", d.Year(), int(d.Month()))
}
