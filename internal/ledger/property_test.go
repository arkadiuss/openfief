package ledger_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Property tests draw synthetic cases from a seeded generator, so every run
// checks the same cases and a failure names the seed and case that broke.
const (
	propertySeed  = 20260302
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

// decimal returns a positive decimal with up to maxUnits whole units and the
// given number of decimal places.
func (r *rng) decimal(maxUnits int, places int32) kernel.Decimal {
	scale := int64(1)
	for range places {
		scale *= 10
	}
	return kernel.NewDecimal(int64(r.intn(maxUnits))*scale+int64(r.intn(int(scale)))+1, -places)
}

// randomTransaction builds one of the event constructors' transactions with
// random amounts, currencies and fees.
func randomTransaction(t *testing.T, r *rng, i int) ledger.Transaction {
	t.Helper()
	currencies := []kernel.Currency{cur(t, "PLN"), cur(t, "USD"), cur(t, "EUR")}
	instruments := []ledger.Instrument{instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")}
	pick := func() kernel.Currency { return currencies[r.intn(len(currencies))] }
	amount := func(c kernel.Currency) kernel.Money { return kernel.NewMoney(r.decimal(10000, 2), c) }
	fee := kernel.Money{}
	if r.intn(2) == 0 {
		fee = amount(pick())
	}
	h := header(t, fmt.Sprintf("p-%d", i), "2026-03-02")
	in := instruments[r.intn(len(instruments))]
	units := kernel.NewQuantity(r.decimal(500, 3))
	opened := date(t, "2020-01-15")
	builders := []func() (ledger.Transaction, error){
		func() (ledger.Transaction, error) { return ledger.NewBuy(h, in, units, amount(pick()), fee) },
		func() (ledger.Transaction, error) { return ledger.NewSell(h, in, units, amount(pick()), fee) },
		func() (ledger.Transaction, error) {
			gross := amount(pick())
			return ledger.NewDividend(h, in, gross, gross.Mul(kernel.NewDecimal(int64(r.intn(31)), -2)))
		},
		func() (ledger.Transaction, error) { return ledger.NewFee(h, amount(pick())) },
		func() (ledger.Transaction, error) {
			from := r.intn(len(currencies))
			to := (from + 1 + r.intn(len(currencies)-1)) % len(currencies)
			return ledger.NewFXConversion(h, amount(currencies[from]), amount(currencies[to]), fee)
		},
		func() (ledger.Transaction, error) { return ledger.NewDeposit(h, amount(pick())) },
		func() (ledger.Transaction, error) { return ledger.NewWithdrawal(h, amount(pick())) },
		func() (ledger.Transaction, error) { return ledger.NewOpeningLot(h, in, units, amount(pick()), opened) },
	}
	tx, err := builders[r.intn(len(builders))]()
	if err != nil {
		t.Fatalf("case %d: %v", i, err)
	}
	return tx
}

func TestPropertyPostingsSumToZeroPerCurrency(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	for i := range propertyCases {
		tx := randomTransaction(t, r, i)
		sums := map[kernel.Currency]kernel.Decimal{}
		for _, p := range tx.Postings {
			sums[p.Amount.Currency()] = sums[p.Amount.Currency()].Add(p.Amount.Amount())
		}
		for c, sum := range sums {
			if !sum.IsZero() {
				t.Fatalf("seed %d case %d (%s): postings sum to %s %s", propertySeed, i, tx.Kind, sum, c)
			}
		}
		if err := tx.Validate(); err != nil {
			t.Fatalf("seed %d case %d (%s): Validate() = %v", propertySeed, i, tx.Kind, err)
		}
		// Any change to one posting's amount must unbalance the transaction.
		broken := tx
		broken.Postings = append([]ledger.Posting(nil), tx.Postings...)
		p := &broken.Postings[r.intn(len(broken.Postings))]
		p.Amount = kernel.NewMoney(p.Amount.Amount().Add(kernel.NewDecimal(1, -3)), p.Amount.Currency())
		var unbalanced *ledger.UnbalancedError
		if !errors.As(broken.Validate(), &unbalanced) {
			t.Fatalf("seed %d case %d (%s): a posting off by 0.001 still validates", propertySeed, i, tx.Kind)
		}
	}
}
