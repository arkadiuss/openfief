package demo

import (
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// opKind is what one simulated operation in an account does.
type opKind int

const (
	opDeposit opKind = iota + 1
	opBuy
	opSell
	opDividend
	opFee
	opFX
)

// op is one operation of the simulated history, before it is written as a
// Flex row or a ledger transaction. Amounts are in hundredths.
type op struct {
	kind opKind
	id   string
	date kernel.Date
	in   *Instrument
	// units is the quantity bought or sold.
	units int64
	// cents is the deposit, the trade's value, the gross dividend, the fee,
	// or the FX quote amount paid, in currency.
	cents    int64
	currency kernel.Currency
	// withheld is the dividend's tax withheld, and fee a trade's or FX
	// conversion's commission, both in currency.
	withheld int64
	fee      int64
	// base and baseUnits are the currency an FX conversion buys and how
	// many whole units of it.
	base      kernel.Currency
	baseUnits int64
	desc      string
}

// lot is an opening lot of the simulated history.
type lot struct {
	in        *Instrument
	units     int64
	costCents int64
	opened    kernel.Date
	id        string
}

// snapshot is an account's cash and holdings at the end of a day.
type snapshot struct {
	date  kernel.Date
	cash  map[kernel.Currency]int64
	units map[*Instrument]int64
}

// sim simulates one account's history: it applies operations to its cash
// and holdings, skipping any it cannot afford, and records them.
type sim struct {
	r     *rng
	paths map[*Instrument]pricePath
	// idPrefix and seq make each operation's institution ID.
	idPrefix string
	seq      int
	cash     map[kernel.Currency]int64
	units    map[*Instrument]int64
	ops      []op
}

func newSim(r *rng, paths map[*Instrument]pricePath, idPrefix string) *sim {
	return &sim{r: r, paths: paths, idPrefix: idPrefix, cash: map[kernel.Currency]int64{}, units: map[*Instrument]int64{}}
}

// nextID returns the next institution ID of this account.
func (s *sim) nextID() string {
	const idDigits = 6
	s.seq++
	n := strconv.Itoa(s.seq)
	return s.idPrefix + strings.Repeat("0", max(idDigits-len(n), 0)) + n
}

// open records an opening lot and opening cash: what the account held at
// go-live.
func (s *sim) open(lots []lot, cash map[kernel.Currency]int64) {
	for i := range lots {
		s.units[lots[i].in] += lots[i].units
	}
	for c, n := range cash {
		s.cash[c] += n
	}
}

// add records o.
func (s *sim) add(o *op) {
	o.id = s.nextID()
	s.ops = append(s.ops, *o)
}

// deposit brings cents of c into the account.
func (s *sim) deposit(d kernel.Date, c kernel.Currency, n int64) {
	s.cash[c] += n
	s.add(&op{kind: opDeposit, date: d, cents: n, currency: c})
}

// fee charges cents of c.
func (s *sim) fee(d kernel.Date, c kernel.Currency, n int64, desc string) {
	s.cash[c] -= n
	s.add(&op{kind: opFee, date: d, cents: n, currency: c, desc: desc})
}

// buy buys units of in at the day's close plus a commission, unless the
// cash in its currency doesn't cover it.
func (s *sim) buy(d kernel.Date, in *Instrument, units, commission int64) {
	value := units * s.paths[in][d]
	if value <= 0 || s.cash[in.Currency] < value+commission {
		return
	}
	s.cash[in.Currency] -= value + commission
	s.units[in] += units
	s.add(&op{kind: opBuy, date: d, in: in, units: units, cents: value, currency: in.Currency, fee: commission})
}

// sell sells units of in at the day's close, less a commission, if the
// account holds them.
func (s *sim) sell(d kernel.Date, in *Instrument, units, commission int64) {
	value := units * s.paths[in][d]
	if value <= 0 || s.units[in] < units {
		return
	}
	s.cash[in.Currency] += value - commission
	s.units[in] -= units
	s.add(&op{kind: opSell, date: d, in: in, units: units, cents: value, currency: in.Currency, fee: commission})
}

// dividend pays perUnit hundredths for every unit of in held, with
// withholdPct percent of it withheld, rounded half up to the hundredth.
func (s *sim) dividend(d kernel.Date, in *Instrument, perUnit, withholdPct int64) {
	const pct, half = 100, 50
	gross := perUnit * s.units[in]
	if gross <= 0 {
		return
	}
	withheld := (gross*withholdPct + half) / pct
	s.cash[in.Currency] += gross - withheld
	s.add(&op{kind: opDividend, date: d, in: in, cents: gross, currency: in.Currency, withheld: withheld})
}

// fx buys baseUnits whole units of base with quote at rate (base in quote,
// in ten-thousandths), plus a commission in quote, if the cash covers it.
func (s *sim) fx(d kernel.Date, base, quote kernel.Currency, baseUnits, rate, commission int64) {
	const tenThousandthsPerHundredth = 100
	paid := baseUnits * rate / tenThousandthsPerHundredth
	if s.cash[quote] < paid+commission {
		return
	}
	s.cash[quote] -= paid + commission
	s.cash[base] += baseUnits * tenThousandthsPerHundredth
	s.add(&op{kind: opFX, date: d, cents: paid, currency: quote, fee: commission, base: base, baseUnits: baseUnits})
}

// snapshot returns the account's cash and holdings now, as of d.
func (s *sim) snapshot(d kernel.Date) snapshot {
	snap := snapshot{date: d, cash: map[kernel.Currency]int64{}, units: map[*Instrument]int64{}}
	maps.Copy(snap.cash, s.cash)
	for in, n := range s.units {
		if n != 0 {
			snap.units[in] = n
		}
	}
	return snap
}

// scheduled is an operation planned for a day, applied in date order.
type scheduled struct {
	date  kernel.Date
	apply func(d kernel.Date)
}

// plan holds the operations planned for one period.
type plan []scheduled

// on plans apply for the n-th business day of p, if that day is in p.
func (pl *plan) on(p *period, n int, apply func(d kernel.Date)) {
	if d, ok := p.businessDay(n); ok {
		*pl = append(*pl, scheduled{date: d, apply: apply})
	}
}

// run applies the planned operations in date order, keeping the planned
// order within one day.
func (pl *plan) run() {
	sort.SliceStable(*pl, func(i, j int) bool { return (*pl)[i].date.Before((*pl)[j].date) })
	for _, s := range *pl {
		s.apply(s.date)
	}
}
