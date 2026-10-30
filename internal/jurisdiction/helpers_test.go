package jurisdiction_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// money parses "amount CUR"-style input given apart: an empty currency
// gives money without one.
func money(t *testing.T, amount, currency string) kernel.Money {
	t.Helper()
	if currency == "" {
		return kernel.NewMoney(dec(t, amount), kernel.Currency{})
	}
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", currency, err)
	}
	return kernel.NewMoney(dec(t, amount), c)
}

// Property tests draw synthetic cases from a seeded generator, so every run
// checks the same cases and a failure names the case that broke.
const (
	propertySeed  = 20261029
	propertyCases = 500
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
