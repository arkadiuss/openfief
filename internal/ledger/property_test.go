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
		if _, ok := errors.AsType[*ledger.UnbalancedError](broken.Validate()); !ok {
			t.Fatalf("seed %d case %d (%s): a posting off by 0.001 still validates", propertySeed, i, tx.Kind)
		}
	}
}

// randomLots returns 1 to 6 lots of instrument A with random open dates
// (some on the same day), units and USD costs, in random order.
func randomLots(t *testing.T, r *rng, base kernel.Date) []ledger.Lot {
	t.Helper()
	in := instrument(t, isinA, "TSTA")
	usd := cur(t, "USD")
	lots := make([]ledger.Lot, 1+r.intn(6))
	for i := range lots {
		lots[i] = ledger.Lot{
			Instrument: in,
			Opened:     base.AddDays(r.intn(10)),
			Quantity:   kernel.NewQuantity(r.decimal(100, 3)),
			Cost:       kernel.NewMoney(r.decimal(10000, 4), usd),
			Opening:    r.intn(2) == 0,
			Origin:     ledger.TransactionKey{Source: "test", ExternalID: fmt.Sprintf("l-%d", i)},
		}
	}
	return lots
}

// totalUnits sums the lots' units.
func totalUnits(lots []ledger.Lot) kernel.Quantity {
	var q kernel.Quantity
	for i := range lots {
		q = q.Add(lots[i].Quantity)
	}
	return q
}

// costByOrigin sums the lots' costs per origin.
func costByOrigin(lots ...[]ledger.Lot) map[ledger.TransactionKey]kernel.Decimal {
	out := map[ledger.TransactionKey]kernel.Decimal{}
	for _, ls := range lots {
		for i := range ls {
			out[ls[i].Origin] = out[ls[i].Origin].Add(ls[i].Cost.Amount())
		}
	}
	return out
}

// fifoBefore reports whether a is consumed before b: earlier open date, then
// smaller origin key.
func fifoBefore(a, b *ledger.Lot) bool {
	if c := a.Opened.Compare(b.Opened); c != 0 {
		return c < 0
	}
	return a.Origin.ExternalID < b.Origin.ExternalID
}

func TestPropertyFIFOConsumesOldestFirstAndConservesCost(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 1}
	base := date(t, "2025-01-01")
	for i := range propertyCases {
		lots := randomLots(t, r, base)
		total := totalUnits(lots)
		// A share of the total in thousandths, at least 0.001 unit.
		take, err := total.Decimal().Mul(kernel.NewDecimal(int64(1+r.intn(1000)), -3)).Round(3, kernel.RoundDown)
		if err != nil {
			t.Fatal(err)
		}
		q := kernel.NewQuantity(take.Add(kernel.NewDecimal(1, -3)))
		if q.Cmp(total) > 0 {
			q = total
		}
		got, err := ledger.ConsumeFIFO(lots, q)
		if err != nil {
			t.Fatalf("seed %d case %d: ConsumeFIFO(%s of %s): %v", propertySeed+1, i, q, total, err)
		}
		checkFIFOCase(t, i, lots, got.Consumed, got.Remaining, q)
	}
}

// checkFIFOCase checks one ConsumeFIFO result against its invariants.
func checkFIFOCase(t *testing.T, i int, lots, consumed, remaining []ledger.Lot, q kernel.Quantity) {
	t.Helper()
	if got := totalUnits(consumed); !got.Equal(q) {
		t.Fatalf("case %d: consumed %s units, want %s", i, got, q)
	}
	if got := totalUnits(consumed).Add(totalUnits(remaining)); !got.Equal(totalUnits(lots)) {
		t.Fatalf("case %d: consumed plus remaining is %s units, want %s", i, got, totalUnits(lots))
	}
	// Cost conservation: per original lot, consumed cost + remaining cost == original cost.
	want, got := costByOrigin(lots), costByOrigin(consumed, remaining)
	for origin, cost := range want {
		if !got[origin].Equal(cost) {
			t.Fatalf("case %d: lot %s cost %s became %s", i, origin, cost, got[origin])
		}
	}
	checkFIFOOrder(t, i, consumed, remaining)
}

// checkFIFOOrder checks that nothing left is older than anything consumed,
// except the one lot consumed in part, which is the newest consumed and the
// oldest left.
func checkFIFOOrder(t *testing.T, i int, consumed, remaining []ledger.Lot) {
	t.Helper()
	for c := range consumed {
		for k := range remaining {
			if remaining[k].Origin != consumed[c].Origin && fifoBefore(&remaining[k], &consumed[c]) {
				t.Fatalf("case %d: lot %s left while newer lot %s was consumed", i, remaining[k].Origin, consumed[c].Origin)
			}
		}
	}
	if len(consumed) > 0 && len(remaining) > 0 && consumed[len(consumed)-1].Origin == remaining[0].Origin {
		if remaining[0].Quantity.Sign() <= 0 {
			t.Fatalf("case %d: lot %s split with nothing left", i, remaining[0].Origin)
		}
	}
}

func TestPropertyOversellConsumesNothing(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 2}
	base := date(t, "2025-01-01")
	for i := range propertyCases {
		lots := randomLots(t, r, base)
		total := totalUnits(lots)
		q := total.Add(kernel.NewQuantity(r.decimal(5, 3)))
		got, err := ledger.ConsumeFIFO(lots, q)
		var oversell *ledger.OversellError
		if !errors.As(err, &oversell) || !oversell.Requested.Equal(q) || !oversell.Available.Equal(total) {
			t.Fatalf("case %d: ConsumeFIFO(%s of %s) = %v, want an *OversellError with both", i, q, total, err)
		}
		if got.Consumed != nil || got.Remaining != nil {
			t.Fatalf("case %d: an oversell consumed or returned lots", i)
		}
	}
}

func TestPropertyLotsMatchHolding(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed + 3}
	in := instrument(t, isinA, "TSTA")
	usd := cur(t, "USD")
	day := date(t, "2026-03-02")
	for c := range propertyCases / 10 {
		log := make([]ledger.VersionedTransaction, 0, 20)
		var held kernel.Quantity
		for i := range 20 {
			h := ledger.TransactionHeader{
				Key:     ledger.TransactionKey{Source: "test", ExternalID: fmt.Sprintf("c%d-%02d", c, i)},
				Account: testAccount, Date: day.AddDays(r.intn(15)),
			}
			units := kernel.NewQuantity(r.decimal(20, 2))
			var tx ledger.Transaction
			var err error
			if r.intn(3) == 0 && held.Cmp(units) >= 0 && i > 0 {
				tx, err = ledger.NewSell(h, in, units, kernel.NewMoney(r.decimal(1000, 2), usd), kernel.Money{})
				held = held.Sub(units)
			} else {
				h.Date = day // acquisitions first, so every sale is covered
				tx, err = ledger.NewBuy(h, in, units, kernel.NewMoney(r.decimal(1000, 2), usd), kernel.Money{})
				held = held.Add(units)
			}
			if err != nil {
				t.Fatalf("case %d: %v", c, err)
			}
			log = append(log, ledger.VersionedTransaction{Transaction: tx, Version: 1})
		}
		checkLotsMatchHolding(t, c, log, in, day.AddDays(20))
	}
}

// checkLotsMatchHolding checks that the open lots hold exactly the holding.
func checkLotsMatchHolding(t *testing.T, c int, log []ledger.VersionedTransaction, in ledger.Instrument, on kernel.Date) {
	t.Helper()
	lots, err := ledger.LotsOn(log, testAccount, in, on)
	if err != nil {
		t.Fatalf("case %d: LotsOn: %v", c, err)
	}
	holding, err := ledger.HoldingOn(log, testAccount, in, on)
	if err != nil {
		t.Fatalf("case %d: HoldingOn: %v", c, err)
	}
	if got := totalUnits(lots); !got.Equal(holding.Quantity) {
		t.Fatalf("case %d: lots hold %s units, holding is %s", c, got, holding.Quantity)
	}
}
