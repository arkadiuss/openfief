package views_test

import (
	"fmt"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// Property tests draw synthetic cases from a seeded generator, so every run
// checks the same cases and a failure names the seed and case that broke.
const (
	propertySeed  = 20261028
	propertyCases = 200
	propertyDays  = 30
)

// rng is splitmix64: a tiny deterministic generator, so the tests need no
// math/rand and no wall clock.
type rng struct{ state uint64 }

func (r *rng) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns a number in [0, n), for n > 0.
func (r *rng) intn(n int) int {
	v := int(r.next() >> 1) // fits a 64-bit int
	return v % n
}

// price returns a PLN price in grosze between 50.00 and 149.99.
func (r *rng) price() int64 { return int64(5000 + r.intn(10000)) }

// pricePath is a random daily price of ETF1 in grosze for days 0 to days-1.
func pricePath(r *rng, days int) []int64 {
	out := make([]int64, days)
	for i := range out {
		out[i] = r.price()
	}
	return out
}

// investedCase is a portfolio fully invested in ETF1 at the prices of path:
// acc-1 deposits the cost of units on day 0 and buys them; on day k, acc-2
// deposits the cost of more units at that day's price and buys them. The
// valuations are exact: units times price.
func investedCase(t *testing.T, path []int64, units, k, more int64) *views.PerformanceInput {
	t.Helper()
	etf := instrument(t, isinIE, "ETF1")
	grosze := func(n int64) kernel.Money { return kernel.NewMoney(kernel.NewDecimal(n, -2), pln(t)) }
	in := &views.PerformanceInput{From: perfDay(t, 0), To: perfDay(t, len(path)-1), ValuedThrough: perfDay(t, len(path)-1), Reporting: pln(t)}
	buy := func(acc ledger.AccountID, day int, qty int64) {
		h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: fmt.Sprintf("%s-%d", acc, day)}, Account: acc, Date: perfDay(t, day)}
		cost := grosze(qty * path[day])
		h.Key.ExternalID += "-deposit"
		dep, err := ledger.NewDeposit(h, cost)
		if err != nil {
			t.Fatalf("NewDeposit: %v", err)
		}
		h.Key.ExternalID += "-buy"
		b, err := ledger.NewBuy(h, etf, kernel.NewQuantity(kernel.NewDecimal(qty, 0)), cost, kernel.Money{})
		if err != nil {
			t.Fatalf("NewBuy: %v", err)
		}
		in.Transactions = append(in.Transactions, dep, b)
	}
	buy("acc-1", 0, units)
	if more > 0 {
		buy("acc-2", int(k), more)
	}
	for day, p := range path {
		in.Valuations = append(in.Valuations, valuation.HoldingValuation{
			Account: "acc-1", Instrument: etf, Date: perfDay(t, day), Basis: valuation.Market,
			Quantity: kernel.NewQuantity(kernel.NewDecimal(units, 0)), Value: grosze(units * p), PLN: grosze(units * p),
		})
		if more > 0 && int64(day) >= k {
			in.Valuations = append(in.Valuations, valuation.HoldingValuation{
				Account: "acc-2", Instrument: etf, Date: perfDay(t, day), Basis: valuation.Market,
				Quantity: kernel.NewQuantity(kernel.NewDecimal(more, 0)), Value: grosze(more * p), PLN: grosze(more * p),
			})
		}
	}
	return in
}

func indexes(t *testing.T, in *views.PerformanceInput) []kernel.Decimal {
	t.Helper()
	got, err := views.Performance(in)
	if err != nil {
		t.Fatalf("Performance() error = %v", err)
	}
	if got.Gap != 0 {
		t.Fatalf("Performance() gap = %v, want none", got.Gap)
	}
	out := make([]kernel.Decimal, 0, len(got.Days))
	for i := range got.Days {
		out = append(out, got.Days[i].Index)
	}
	return out
}

// TestPerformanceIgnoresDepositSize checks that the time-weighted return does
// not depend on the size of a deposit invested at the day's price: the index
// is the same, day by day, with no deposit and with a deposit of any size on
// any day.
func TestPerformanceIgnoresDepositSize(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	for c := range propertyCases {
		path := pricePath(r, propertyDays)
		units, k, more := int64(1+r.intn(50)), int64(1+r.intn(propertyDays-1)), int64(1+r.intn(5000))
		without := indexes(t, investedCase(t, path, units, k, 0))
		with := indexes(t, investedCase(t, path, units, k, more))
		for day := range without {
			if !without[day].Equal(with[day]) {
				t.Fatalf("seed %d case %d: day %d index %s with a deposit of %d units on day %d, %s without", propertySeed, c, day, with[day], more, k, without[day])
			}
		}
	}
}

// cashOnly is a portfolio of PLN cash only: acc-1 deposits or withdraws a
// random amount every day, never more than it holds.
func cashOnly(t *testing.T, r *rng) *views.PerformanceInput {
	t.Helper()
	in := &views.PerformanceInput{From: perfDay(t, 0), To: perfDay(t, propertyDays-1), ValuedThrough: perfDay(t, propertyDays-1), Reporting: pln(t)}
	balance := int64(0)
	for day := range propertyDays {
		h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: fmt.Sprintf("flow-%d", day)}, Account: "acc-1", Date: perfDay(t, day)}
		amount := int64(1 + r.intn(1_000_000))
		build, sign := ledger.NewDeposit, int64(1)
		if balance > amount && r.intn(2) == 0 {
			build, sign = ledger.NewWithdrawal, -1
		}
		tx, err := build(h, kernel.NewMoney(kernel.NewDecimal(amount, -2), pln(t)))
		if err != nil {
			t.Fatalf("building flow: %v", err)
		}
		balance += sign * amount
		in.Transactions = append(in.Transactions, tx)
	}
	return in
}

// TestPerformanceOfCashIsFlat checks that cash in the reporting currency
// earns nothing, whatever deposits and withdrawals move it: the index stays
// at 100 every day.
func TestPerformanceOfCashIsFlat(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 1}
	for c := range propertyCases {
		for day, idx := range indexes(t, cashOnly(t, r)) {
			if !idx.Equal(kernel.NewDecimal(100, 0)) {
				t.Fatalf("seed %d case %d: day %d index %s, want 100", propertySeed+1, c, day, idx)
			}
		}
	}
}

// TestPerformanceDrift checks that rounding each day's factor to TWRPlaces
// stays within the documented bound over the longest range: with no cash
// flows the exact index is 100 × last value / first value, and the chained
// one may differ from it by at most 2 × 10^-12 of the growth per day.
func TestPerformanceDrift(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 2}
	path := pricePath(r, views.MaxRangeDays+1)
	in := investedCase(t, path, 7, 0, 0)
	got := indexes(t, in)
	exact, err := kernel.NewDecimal(100*path[len(path)-1], 0).Quo(kernel.NewDecimal(path[0], 0), 20, kernel.RoundHalfEven)
	if err != nil {
		t.Fatalf("exact index: %v", err)
	}
	drift := got[len(got)-1].Sub(exact)
	// 100 × 2e-12 × days × the largest growth on the path (at most 3).
	bound := kernel.NewDecimal(6*int64(len(path)), -10)
	if drift.Cmp(bound) > 0 || drift.Neg().Cmp(bound) > 0 {
		t.Errorf("index after %d days = %s, exact %s: drift %s exceeds %s", len(path)-1, got[len(got)-1], exact, drift, bound)
	}
}
