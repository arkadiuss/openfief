package treasurybonds

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// DueCoupon is one coupon that fell due on an open lot of a held account:
// the account, the lot's instrument, series terms, purchase date and bonds,
// and the interest period it pays.
type DueCoupon struct {
	Account   ledger.AccountID
	Wrapper   string
	Lot       ledger.Lot
	Terms     pl.Terms
	Period    int
	Due       kernel.Date
	Purchased kernel.Date
}

// DueCoupons returns every coupon (pl.CouponDates) that fell due after its
// account's go-live and on or before today on an open lot of a coupon
// series held in accounts (D-117), in account, lot and period order. What
// fell due on or before go-live is history the opening lots hold (D-98),
// and an account not seeded yet has no lots. A lot whose instrument is no
// bond series is not a bond and is left out; a bond series without terms
// is a problem, and so is a lot whose purchase date doesn't match its
// series.
func DueCoupons(accounts []ingest.HeldAccount, terms []ingest.BondTerms, today kernel.Date) ([]DueCoupon, []ingest.Problem) {
	var (
		out      []DueCoupon
		problems []ingest.Problem
	)
	for i := range accounts {
		a := &accounts[i]
		if a.GoLive.IsZero() {
			continue
		}
		for j := range a.Lots {
			due, problem := dueOnLot(a, &a.Lots[j], terms, today)
			out = append(out, due...)
			if problem != nil {
				problems = append(problems, *problem)
			}
		}
	}
	return out, problems
}

// dueOnLot returns the coupons due on lot of account a (DueCoupons), or the
// problem that keeps them from being computed.
func dueOnLot(a *ingest.HeldAccount, lot *ledger.Lot, terms []ingest.BondTerms, today kernel.Date) ([]DueCoupon, *ingest.Problem) {
	s, err := pl.ParseSeries(lot.Instrument.Symbol)
	if err != nil {
		return nil, nil
	}
	t, ok := termsOf(terms, lot.Instrument)
	if !ok {
		p := ingest.Skipped(lotRef(a.Account, s, lot.Opened), "no series terms to compute its coupons")
		return nil, &p
	}
	dates, err := pl.CouponDates(s, lot.Opened)
	if err != nil {
		p := ingest.Skipped(lotRef(a.Account, s, lot.Opened), err.Error())
		return nil, &p
	}
	var out []DueCoupon
	for k, due := range dates {
		if due.After(a.GoLive) && !due.After(today) {
			out = append(out, DueCoupon{
				Account: a.Account, Wrapper: a.Wrapper, Lot: *lot, Terms: t, Period: k + 1, Due: due, Purchased: lot.Opened,
			})
		}
	}
	return out, nil
}

// termsOf returns the terms of the series in, by ISIN.
func termsOf(terms []ingest.BondTerms, in ledger.Instrument) (pl.Terms, bool) {
	i := slices.IndexFunc(terms, func(t ingest.BondTerms) bool { return t.Instrument.ISIN == in.ISIN })
	if i < 0 {
		return pl.Terms{}, false
	}
	return terms[i].Terms, true
}

// lotRef names a lot of account in a problem.
func lotRef(account ledger.AccountID, s pl.Series, purchased kernel.Date) string {
	return string(account) + " " + s.String() + " bought " + purchased.String()
}

// CPIYears returns the years of CPI the coupons' periods need, in order,
// and whether any needs the NBP reference rate: the reading of each
// indexed period after the first (valuation.LotIndices).
func CPIYears(due []DueCoupon) ([]int, bool) {
	var (
		years []int
		rate  bool
	)
	for i := range due {
		c := &due[i]
		st, err := c.Terms.Series.Type.Structure()
		if err != nil || c.Period == 1 {
			continue
		}
		switch st.Indexation {
		case pl.Inflation:
			start, err := pl.PeriodStart(c.Terms.Series.Type, c.Purchased, c.Period)
			if err != nil {
				continue
			}
			y, _ := pl.CPIMonth(start)
			years = append(years, y)
		case pl.ReferenceRate:
			rate = true
		case pl.Fixed:
		}
	}
	slices.Sort(years)
	return slices.Compact(years), rate
}

// BookCoupons books each due coupon as interest (D-101) in its account: its
// lot's bonds times what one bond is paid (pl.Terms.Coupon), with the Belka
// tax the account's wrapper withholds (pl.BelkaTax), on the day it fell
// due. readings supply the period index readings. Each is keyed
// "lot:<series>:<purchase date>:coupon:<due date>" under the source
// SourcePrefix plus the account, as the lot itself, so it is idempotent per
// lot and coupon date: delivering it again on every sync changes nothing.
// A coupon whose reading is missing waits as a problem, and so does one in
// an account with a wrapper the pack doesn't know.
func BookCoupons(due []DueCoupon, readings []valuation.IndexReading) (ingest.Mapping, error) {
	var out ingest.Mapping
	for i := range due {
		c := &due[i]
		t, err := bookCoupon(c, readings)
		_, missing := errors.AsType[*pl.MissingIndexError](err)
		switch {
		case missing:
			out.Problems = append(out.Problems, ingest.Skipped(couponRef(c), "its period's index reading is not published yet"))
			continue
		case errors.Is(err, pl.ErrUnknownTaxWrapper):
			out.Problems = append(out.Problems, ingest.Skipped(couponRef(c), "the account's tax wrapper is unknown"))
			continue
		case err != nil:
			return ingest.Mapping{}, err
		}
		out.Transactions = append(out.Transactions, t)
	}
	return out, nil
}

// couponRef names a coupon in a problem.
func couponRef(c *DueCoupon) string {
	return lotRef(c.Account, c.Terms.Series, c.Purchased) + " coupon " + c.Due.String()
}

// bookCoupon returns the interest transaction of c.
func bookCoupon(c *DueCoupon, readings []valuation.IndexReading) (ledger.Transaction, error) {
	idx, err := valuation.LotIndices(c.Terms.Series, c.Purchased, c.Due, readings)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("coupon of %s: %w", c.Terms.Series, err)
	}
	coupon, err := c.Terms.Coupon(idx, c.Purchased, c.Period)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("coupon of %s: %w", c.Terms.Series, err)
	}
	wrapper, err := pl.ParseTaxWrapper(c.Wrapper)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("coupon of %s: %w", c.Terms.Series, err)
	}
	gross := kernel.NewMoney(coupon.PerBond.Amount().Mul(c.Lot.Quantity.Decimal()), coupon.PerBond.Currency())
	tax, err := pl.BelkaTax(gross, wrapper)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("coupon of %s: %w", c.Terms.Series, err)
	}
	m := Mapper{account: c.Account}
	key := ingest.BondLotKey{Series: c.Terms.Series, Purchased: c.Purchased}
	t, err := ledger.NewInterest(m.header(lotID(key)+":coupon:"+c.Due.String(), c.Due), c.Lot.Instrument, gross, tax)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("coupon of %s: %w", c.Terms.Series, err)
	}
	t.Description = fmt.Sprintf("Treasury bonds %s coupon, period %d", c.Terms.Series, c.Period)
	return t, nil
}

// CouponSync books the coupons that fell due on the treasury bonds the
// core says each bound account holds (D-117), fetching the CPI and NBP
// reference-rate readings their periods need from GUS and NBP, since the
// connector stores nothing.
type CouponSync struct {
	CPI   CPISource
	Rates ReferenceRateSource
}

// Coupons returns the mapping of every coupon due on accounts' open lots
// as of today (DueCoupons, BookCoupons). It asks GUS once per CPI year the
// coupons need, at most maxCPIYearsPerSync years, the earliest first: a
// coupon whose year is left for a later sync waits as a problem. NBP's
// archive is asked only when a reference-rate coupon is due. A failed
// request fails the sync.
func (c *CouponSync) Coupons(ctx context.Context, accounts []ingest.HeldAccount, terms []ingest.BondTerms, today kernel.Date) (ingest.Mapping, error) {
	due, problems := DueCoupons(accounts, terms, today)
	years, needRate := CPIYears(due)
	var readings []valuation.IndexReading
	for i, y := range years {
		if i == maxCPIYearsPerSync {
			break
		}
		cpi, err := c.CPI.CPI(ctx, y)
		if err != nil {
			return ingest.Mapping{}, fmt.Errorf("CPI for %d: %w", y, err)
		}
		readings = append(readings, cpi...)
	}
	if needRate {
		changes, err := c.Rates.ReferenceRates(ctx)
		if err != nil {
			return ingest.Mapping{}, fmt.Errorf("reference rates: %w", err)
		}
		rates, err := ReferenceReadings(changes, earliestPurchase(due), today)
		if err != nil {
			return ingest.Mapping{}, err
		}
		readings = append(readings, rates...)
	}
	m, err := BookCoupons(due, readings)
	if err != nil {
		return ingest.Mapping{}, err
	}
	m.Problems = append(problems, m.Problems...)
	return m, nil
}

// earliestPurchase returns the earliest purchase date of due's lots.
func earliestPurchase(due []DueCoupon) kernel.Date {
	var first kernel.Date
	for i := range due {
		if first.IsZero() || due[i].Purchased.Before(first) {
			first = due[i].Purchased
		}
	}
	return first
}
