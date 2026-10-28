package portfolio_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Property tests draw synthetic cases from a seeded generator, so every run
// checks the same cases and a failure names the case that broke.
const (
	propertySeed  = 20261028
	propertyCases = 150
	// propertyDays spans two year ends, so rebalancing is exercised.
	propertyDays = 800
)

// rng is splitmix64: a tiny deterministic generator, so the tests need no
// math/rand and no wall clock.
type rng struct{ state uint64 }

// next returns the next 64 random bits.
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

// path returns days random prices in hundredths, starting in [50, 150) and
// moving at most 5% a day, so they stay positive.
func (r *rng) path(days int) []int64 {
	out := make([]int64, days)
	p := int64(5000 + r.intn(10000))
	for i := range out {
		p += p * int64(r.intn(11)-5) / 100
		out[i] = max(p, 1)
	}
	return out
}

// propertyStart is the first day of every property case: late in a year.
func propertyStart(t *testing.T) kernel.Date {
	t.Helper()
	return mustDate(t, "2022-11-15")
}

// closesFrom turns prices in hundredths, one per day from start, scaled by
// k, into PLN closes.
func closesFrom(t *testing.T, start kernel.Date, path []int64, k int64) []portfolio.Close {
	t.Helper()
	out := make([]portfolio.Close, 0, len(path))
	for i, p := range path {
		out = append(out, portfolio.Close{Date: start.AddDays(i), Value: kernel.NewMoney(kernel.NewDecimal(p*k, -2), mustCurrency(t, "PLN"))})
	}
	return out
}

// within reports whether a and b differ by at most tol.
func within(a, b, tol kernel.Decimal) bool {
	d := a.Sub(b)
	if d.Sign() < 0 {
		d = d.Neg()
	}
	return d.Cmp(tol) <= 0
}

// TestBacktestOfOneTargetIsItsPrice: a model of one target at 100% is
// 100 × price / first price, to within 10^-10, on every day.
func TestBacktestOfOneTargetIsItsPrice(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	start, tol := propertyStart(t), mustDec(t, "1e-10")
	for c := range propertyCases {
		path := r.path(propertyDays / 10)
		tg := portfolio.Target{Kind: portfolio.TargetInstrument, Name: etfA, Weight: mustDec(t, "100")}
		got, err := portfolio.Backtest(&portfolio.BacktestInput{
			Targets: []portfolio.Target{tg}, Closes: map[string][]portfolio.Close{tg.Key(): closesFrom(t, start, path, 1)},
			Reporting: mustCurrency(t, "PLN"), Through: start.AddDays(len(path)),
		})
		if err != nil || len(got.Days) != len(path) {
			t.Fatalf("case %d: %d days, %v", c, len(got.Days), err)
		}
		for i, d := range got.Days {
			want, qErr := kernel.NewDecimal(path[i]*100, 0).Quo(kernel.NewDecimal(path[0], 0), 14, kernel.RoundHalfEven)
			if qErr != nil {
				t.Fatal(qErr)
			}
			if !within(d.Value, want, tol) {
				t.Fatalf("case %d day %d: index %s, want %s", c, i, d.Value, want)
			}
		}
	}
}

// TestBacktestIgnoresPriceScale: multiplying one target's prices by any
// factor, as a share split would, leaves the model's index unchanged to
// within 10^-9, across year-end rebalances.
func TestBacktestIgnoresPriceScale(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 1}
	start, tol := propertyStart(t), mustDec(t, "1e-9")
	a := portfolio.Target{Kind: portfolio.TargetInstrument, Name: etfA, Weight: mustDec(t, "70")}
	b := portfolio.Target{Kind: portfolio.TargetInstrument, Name: etfB, Weight: mustDec(t, "30")}
	for c := range propertyCases / 10 {
		pa, pb, k := r.path(propertyDays), r.path(propertyDays), int64(2+r.intn(999))
		run := func(scale int64) []portfolio.Point {
			got, err := portfolio.Backtest(&portfolio.BacktestInput{
				Targets: []portfolio.Target{a, b},
				Closes: map[string][]portfolio.Close{
					a.Key(): closesFrom(t, start, pa, 1), b.Key(): closesFrom(t, start, pb, scale),
				},
				Reporting: mustCurrency(t, "PLN"), Through: start.AddDays(propertyDays),
			})
			if err != nil {
				t.Fatalf("case %d: %v", c, err)
			}
			return got.Days
		}
		plain, scaled := run(1), run(k)
		for i := range plain {
			if !within(plain[i].Value, scaled[i].Value, tol) {
				t.Fatalf("case %d (x%d) day %s: %s vs %s", c, k, plain[i].Date, plain[i].Value, scaled[i].Value)
			}
		}
	}
}

// TestMaxDrawdownIsTheWorstPair: MaxDrawdown's depth equals the largest
// fall over every pair of a day and a later one, found by brute force, and
// its peak and trough are such a pair.
func TestMaxDrawdownIsTheWorstPair(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 2}
	start := propertyStart(t)
	for c := range propertyCases {
		path := r.path(2 + r.intn(60))
		points := make([]portfolio.Point, len(path))
		for i, p := range path {
			points[i] = portfolio.Point{Date: start.AddDays(i), Value: kernel.NewDecimal(p, -2)}
		}
		got, err := portfolio.MaxDrawdown(points)
		if err != nil {
			t.Fatalf("case %d: %v", c, err)
		}
		worst := worstFall(t, points)
		if !got.Depth.Equal(worst) {
			t.Fatalf("case %d: depth %s, brute force %s", c, got.Depth, worst)
		}
		if got.Trough.Before(got.Peak) {
			t.Fatalf("case %d: trough %s before peak %s", c, got.Trough, got.Peak)
		}
	}
}

// worstFall is the largest fall over every pair of a point and a later
// one, as a fraction of the first, rounded like MaxDrawdown's depth.
func worstFall(t *testing.T, points []portfolio.Point) kernel.Decimal {
	t.Helper()
	worst := kernel.Decimal{}
	for i := range points {
		for j := i + 1; j < len(points); j++ {
			fall, err := points[i].Value.Sub(points[j].Value).Quo(points[i].Value, portfolio.BacktestPlaces, portfolio.BacktestRounding)
			if err != nil {
				t.Fatal(err)
			}
			if fall.Cmp(worst) > 0 {
				worst = fall
			}
		}
	}
	return worst
}
