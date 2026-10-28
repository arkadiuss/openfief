package demo

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrTooEarly reports a today whose go-live comes before a demo bond's
// purchase: the bonds are opening lots, bought before go-live.
var ErrTooEarly = errors.New("demo date too early")

const (
	// bondsOpeningCash is the PLN cash in the bond account at go-live, in
	// hundredths.
	bondsOpeningCash = 25_000
	// nominalCents is one bond's nominal, 100 PLN, in grosze.
	nominalCents = 10_000
)

// simulateBonds returns the bond account's seeding at go-live, its bond
// lots and some cash, and its history since: the coupons of the COI bonds,
// which pay their interest out yearly. The ledger has no interest kind yet,
// so a coupon is booked as a dividend of its series, with no tax withheld
// inside IKE.
func simulateBonds(goLive, today kernel.Date, readings []valuation.IndexReading) (history, error) {
	lots, err := bondLots()
	if err != nil {
		return history{}, err
	}
	w := direct{account: BondsAccount, source: bondsSource}
	held := make([]lot, 0, len(lots))
	units := map[*Instrument]int64{}
	for i := range lots {
		l := &lots[i]
		if l.purchased.After(goLive) {
			return history{}, fmt.Errorf("%w: go-live %s is before the purchase of %s", ErrTooEarly, goLive, l.series.Instrument.Symbol)
		}
		in := &Instrument{Instrument: l.series.Instrument, Name: l.series.Instrument.Symbol, Currency: pl.Nominal().Currency()}
		held = append(held, lot{in: in, units: l.units, costCents: l.units * nominalCents, opened: l.purchased, id: "b" + strconv.Itoa(i+1)})
		units[in] = l.units
	}
	cash := map[kernel.Currency]int64{pl.Nominal().Currency(): bondsOpeningCash}
	opening, err := w.opening(goLive, held, cash)
	if err != nil {
		return history{}, err
	}
	activity, err := w.coupons(lots, goLive, today, readings)
	if err != nil {
		return history{}, err
	}
	activity.Assertions = w.assertions(snapshot{date: today, units: units})
	return history{opening: opening, activity: activity}, nil
}

// coupons returns the coupons paid after goLive and up to today on the lots
// whose series pays coupons.
func (w direct) coupons(lots []bondLot, goLive, today kernel.Date, readings []valuation.IndexReading) (Batch, error) {
	var b Batch
	for i := range lots {
		l := &lots[i]
		st, err := l.series.Terms.Series.Type.Structure()
		if err != nil {
			return Batch{}, fmt.Errorf("demo coupons: %w", err)
		}
		if st.Payout != pl.Coupon {
			continue
		}
		for n := 1; n*st.PeriodMonths < st.TenorMonths; n++ {
			t, paid, err := w.coupon(l, n, goLive, today, readings)
			if err != nil {
				return Batch{}, err
			}
			if paid {
				b.Transactions = append(b.Transactions, t)
			}
		}
	}
	return b, nil
}

// coupon returns the coupon of interest period n of l, paid on the day the
// next period starts, and false when that day is not after goLive and up to
// today.
func (w direct) coupon(l *bondLot, n int, goLive, today kernel.Date, readings []valuation.IndexReading) (ledger.Transaction, bool, error) {
	terms := &l.series.Terms
	start, err1 := pl.PeriodStart(terms.Series.Type, l.purchased, n)
	paid, err2 := pl.PeriodStart(terms.Series.Type, l.purchased, n+1)
	if err := firstError(err1, err2); err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("demo coupon of %s: %w", terms.Series, err)
	}
	if !paid.After(goLive) || paid.After(today) {
		return ledger.Transaction{}, false, nil
	}
	idx, err := valuation.LotIndices(terms.Series, l.purchased, paid, readings)
	if err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("demo coupon of %s: %w", terms.Series, err)
	}
	rate, err := terms.PeriodRate(idx, n, start)
	if err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("demo coupon of %s: %w", terms.Series, err)
	}
	const grosze = 2
	perBond, err := pl.Nominal().Mul(rate).Round(grosze, kernel.RoundHalfUp)
	if err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("demo coupon of %s: %w", terms.Series, err)
	}
	in := l.series.Instrument
	h := w.header("coupon:"+in.ISIN.String()+":"+paid.String(), paid)
	none := kernel.NewMoney(kernel.Decimal{}, perBond.Currency())
	t, err := ledger.NewDividend(h, in, perBond.Mul(kernel.NewDecimal(l.units, 0)), none)
	if err != nil {
		return ledger.Transaction{}, false, fmt.Errorf("demo coupon of %s: %w", terms.Series, err)
	}
	t.Description = in.Symbol + " coupon, period " + strconv.Itoa(n) + " (booked as a dividend)"
	return t, true, nil
}
