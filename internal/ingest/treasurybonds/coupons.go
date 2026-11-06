package treasurybonds

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// DueCoupon is one coupon that fell due on an open lot of a held account:
// the account with its payout setting, the lot's instrument, series terms,
// purchase date and bonds, and the interest period it pays.
type DueCoupon struct {
	Account   ledger.AccountID
	Wrapper   string
	Payout    ingest.PayoutSetting
	Lot       ledger.Lot
	Terms     pl.Terms
	Period    int
	Due       kernel.Date
	Purchased kernel.Date
}

// DueCoupons returns every coupon (pl.CouponDates) that fell due on or
// before today on an open lot of a coupon series held in accounts (D-117),
// in account, lot and period order: all of them since the lot's purchase,
// on every sync, so a lot entered late gets its earlier coupons (D-143). An account whose lots are deposits and
// buys on their purchase dates has its whole history, so every coupon
// since purchase is due (D-98); in an account seeded with opening lots,
// only those after its go-live are, since what fell due on or before it is
// history the opening lots hold. A lot whose instrument is no
// bond series is not a bond and is left out; a bond series without terms
// is a problem, and so is a lot whose purchase date doesn't match its
// series; DueMaturities leaves such lots to these problems.
func DueCoupons(accounts []ingest.HeldAccount, terms []ingest.BondTerms, today kernel.Date) ([]DueCoupon, []ingest.Problem) {
	var (
		out      []DueCoupon
		problems []ingest.Problem
	)
	for i := range accounts {
		a := &accounts[i]
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
		p := ingest.Skipped(lotRef(a.Account, s, lot.Opened), "no series terms to compute its coupons and maturity")
		return nil, &p
	}
	dates, err := pl.CouponDates(s, lot.Opened)
	if err != nil {
		p := ingest.Skipped(lotRef(a.Account, s, lot.Opened), err.Error())
		return nil, &p
	}
	var out []DueCoupon
	for k, due := range dates {
		if (a.GoLive.IsZero() || due.After(a.GoLive)) && !due.After(today) {
			out = append(out, DueCoupon{
				Account: a.Account, Wrapper: a.Wrapper, Payout: a.Payout, Lot: *lot, Terms: t, Period: k + 1, Due: due, Purchased: lot.Opened,
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
// due, followed by a withdrawal of the cash it brought when the account's
// payout setting pays it out (PayoutWithdrawal, D-135). readings supply the
// period index readings. Each is keyed
// "lot:<series>:<purchase date>:coupon:<due date>" under the source
// SourcePrefix plus the account, as the lot itself, so it is idempotent per
// lot and coupon date (D-143): delivering it again on every sync changes
// nothing.
// A coupon whose reading is missing waits as a problem, and so does one in
// an account with a wrapper the pack doesn't know.
func BookCoupons(due []DueCoupon, readings []valuation.IndexReading) (ingest.Mapping, error) {
	return bookPayouts(due, func(c *DueCoupon) payout {
		t, err := bookCoupon(c, readings)
		return payout{tx: t, err: err, setting: c.Payout, ref: couponRef(c), waiting: "its period's index reading is not published yet"}
	})
}

// payout is one coupon or maturity booked: its transaction or the error
// that kept it from being booked, its account's payout setting, how a
// problem names it, and why it waits for a missing index reading.
type payout struct {
	tx      ledger.Transaction
	err     error
	setting ingest.PayoutSetting
	ref     string
	waiting string
}

// bookPayouts books each of due with book, followed by its payout
// withdrawal when its account pays it out (PayoutWithdrawal). One whose
// index reading is missing, or whose account's wrapper the pack doesn't
// know, waits as a problem; any other error fails.
func bookPayouts[T any](due []T, book func(*T) payout) (ingest.Mapping, error) {
	var out ingest.Mapping
	for i := range due {
		p := book(&due[i])
		_, missing := errors.AsType[*pl.MissingIndexError](p.err)
		switch {
		case missing:
			out.Problems = append(out.Problems, ingest.Skipped(p.ref, p.waiting))
			continue
		case errors.Is(p.err, pl.ErrUnknownTaxWrapper):
			out.Problems = append(out.Problems, ingest.Skipped(p.ref, "the account's tax wrapper is unknown"))
			continue
		case p.err != nil:
			return ingest.Mapping{}, p.err
		}
		out.Transactions = append(out.Transactions, p.tx)
		w, paid, err := PayoutWithdrawal(&p.tx, p.setting)
		if err != nil {
			return ingest.Mapping{}, err
		}
		if paid {
			out.Transactions = append(out.Transactions, w)
		}
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

// AccountsBookedIn returns the accounts of accounts whose coupons and
// maturities a sync batch of mode books: in a transactions batch those
// seeded with opening lots, which the core syncs in that mode (D-95), and
// in an opening batch the others, whose lots are deposits and buys and
// which are never seeded (D-98). So a sync that runs both batches books
// each account once, and never seeds an account twice.
func AccountsBookedIn(accounts []ingest.HeldAccount, mode ingest.Mode) []ingest.HeldAccount {
	var out []ingest.HeldAccount
	for i := range accounts {
		if seeded := !accounts[i].GoLive.IsZero(); seeded == (mode == ingest.ModeTransactions) {
			out = append(out, accounts[i])
		}
	}
	return out
}

// CouponSync books the coupons and maturities that fell due on the
// treasury bonds the core says each bound account holds (D-117, D-135,
// D-143), fetching the CPI and NBP reference-rate readings their periods
// need from GUS and NBP, since the connector stores nothing. It keeps in
// memory each year of CPI GUS served in full, whose twelve final readings
// never change, so a later sync asks only for the years it lacks. Use it by
// pointer.
type CouponSync struct {
	CPI   CPISource
	Rates ReferenceRateSource

	// mu guards kept.
	mu sync.Mutex
	// kept holds the CPI readings of each year GUS served in full.
	kept map[int][]valuation.IndexReading
}

// Coupons returns the mapping of every coupon due and every lot matured on
// accounts' open lots as of today (DueCoupons, BookCoupons, DueMaturities,
// BookMaturities), each paid out when its account's payout setting says
// so. It asks GUS once per CPI year they need that it doesn't keep from an
// earlier sync, at most maxCPIYearsPerSync years, the earliest first: a
// coupon or maturity whose year is left for a later sync waits as a
// problem, and since the years served in full are kept, each sync gets
// further until every coupon since each lot's purchase is booked (D-143).
// NBP's archive is asked only when a reference-rate period is due. A
// failed request fails the sync.
func (c *CouponSync) Coupons(ctx context.Context, accounts []ingest.HeldAccount, terms []ingest.BondTerms, today kernel.Date) (ingest.Mapping, error) {
	due, problems := DueCoupons(accounts, terms, today)
	matured := DueMaturities(accounts, terms, today)
	years, needRate := CPIYears(due)
	maturityYears, maturityRate := MaturityCPIYears(matured)
	years = slices.Compact(slices.Sorted(slices.Values(append(years, maturityYears...))))
	needRate = needRate || maturityRate
	readings, err := c.cpiReadings(ctx, years)
	if err != nil {
		return ingest.Mapping{}, err
	}
	if needRate {
		if readings, err = c.withReferenceRates(ctx, readings, earliestPurchase(due, matured), today); err != nil {
			return ingest.Mapping{}, err
		}
	}
	m, err := BookCoupons(due, readings)
	if err != nil {
		return ingest.Mapping{}, err
	}
	r, err := BookMaturities(matured, readings)
	if err != nil {
		return ingest.Mapping{}, err
	}
	m.Transactions = append(m.Transactions, r.Transactions...)
	m.Problems = slices.Concat(problems, m.Problems, r.Problems)
	return m, nil
}

// withReferenceRates returns readings with the NBP reference-rate readings
// of every month from the month of from up to today.
func (c *CouponSync) withReferenceRates(ctx context.Context, readings []valuation.IndexReading, from, today kernel.Date) ([]valuation.IndexReading, error) {
	changes, err := c.Rates.ReferenceRates(ctx)
	if err != nil {
		return nil, fmt.Errorf("reference rates: %w", err)
	}
	rates, err := ReferenceReadings(changes, from, today)
	if err != nil {
		return nil, err
	}
	return append(readings, rates...), nil
}

// cpiReadings returns the CPI readings of years, in order: those of a year
// kept from an earlier sync, and of at most maxCPIYearsPerSync more, asked
// of GUS, the earliest first. A year GUS serves in full, twelve months, is
// kept for later syncs; a later year is left out, to be asked next time.
func (c *CouponSync) cpiReadings(ctx context.Context, years []int) ([]valuation.IndexReading, error) {
	var (
		out   []valuation.IndexReading
		asked int
	)
	for _, y := range years {
		if kept, ok := c.keptYear(y); ok {
			out = append(out, kept...)
			continue
		}
		if asked == maxCPIYearsPerSync {
			continue
		}
		asked++
		cpi, err := c.CPI.CPI(ctx, y)
		if err != nil {
			return nil, fmt.Errorf("CPI for %d: %w", y, err)
		}
		c.keep(y, cpi)
		out = append(out, cpi...)
	}
	return out, nil
}

// keptYear returns the CPI readings of year kept from an earlier sync, and
// whether there are any.
func (c *CouponSync) keptYear(year int) ([]valuation.IndexReading, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.kept[year]
	return r, ok
}

// keep keeps the CPI readings of year when they are the whole year.
func (c *CouponSync) keep(year int, readings []valuation.IndexReading) {
	if len(readings) != monthsPerYear {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kept == nil {
		c.kept = map[int][]valuation.IndexReading{}
	}
	c.kept[year] = slices.Clone(readings)
}

// earliestPurchase returns the earliest purchase date of the lots of due
// and matured.
func earliestPurchase(due []DueCoupon, matured []DueMaturity) kernel.Date {
	var first kernel.Date
	for i := range due {
		if first.IsZero() || due[i].Purchased.Before(first) {
			first = due[i].Purchased
		}
	}
	for i := range matured {
		if first.IsZero() || matured[i].Lot.Opened.Before(first) {
			first = matured[i].Lot.Opened
		}
	}
	return first
}
