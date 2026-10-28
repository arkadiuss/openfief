package demo

import (
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The XTB account's plan: on which business day of a month each operation
// happens, and its amounts, in grosze or in units.
const (
	xtbDepositDay, xtbBuyDay, xtbDividendDay = 3, 5, 8

	xtbDepositBase, xtbDepositStep, xtbDepositSteps = 80_000, 10_000, 4
	xtbEnergyMin, xtbEnergyMax                      = 10, 20
	xtbEmergingMin, xtbEmergingMax                  = 5, 12
	xtbBankDividend, belkaPct                       = 210, 19
	xtbOpeningCash                                  = 250_000
)

// simulateXTB simulates the XTB account in PLN and returns its seeding and
// its history: opening lots in the bank stock and the emerging markets ETF,
// then a monthly deposit, a monthly purchase that alternates between the
// energy stock and the ETF, and the bank's yearly dividend with Belka tax
// withheld. XTB charges no commission on these.
func simulateXTB(m *market, r *rng, goLive, today kernel.Date) (history, error) {
	s := newSim(r, m.paths, "x1")
	lots := m.openingLots(s, goLive, []lotSpec{{symBank, 60, 500, 75}, {symEmerging, 25, 90, 98}})
	cash := map[kernel.Currency]int64{m.pln: xtbOpeningCash}
	s.open(lots, cash)
	w := direct{account: XTBAccount, source: xtbSource}
	opening, err := w.opening(goLive, lots, cash)
	if err != nil {
		return history{}, err
	}
	s.idPrefix, s.seq = "x2", 0
	for _, p := range periods(goLive, today) {
		pl := xtbPlan(m, s, &p)
		pl.run()
	}
	activity, err := w.activity(s.ops, s.snapshot(today))
	if err != nil {
		return history{}, err
	}
	return history{opening: opening, activity: activity}, nil
}

// xtbPlan plans one month of the XTB account.
func xtbPlan(m *market, s *sim, p *period) plan {
	var pl plan
	deposit := xtbDepositBase + xtbDepositStep*s.r.intn(xtbDepositSteps)
	pl.on(p, xtbDepositDay, func(d kernel.Date) { s.deposit(d, m.pln, deposit) })
	if p.index%2 == 0 {
		units := s.r.between(xtbEnergyMin, xtbEnergyMax)
		pl.on(p, xtbBuyDay, func(d kernel.Date) { s.buy(d, m.in(symEnergy), units, 0) })
	} else {
		units := s.r.between(xtbEmergingMin, xtbEmergingMax)
		pl.on(p, xtbBuyDay, func(d kernel.Date) { s.buy(d, m.in(symEmerging), units, 0) })
	}
	if p.month == time.July {
		pl.on(p, xtbDividendDay, func(d kernel.Date) { s.dividend(d, m.in(symBank), xtbBankDividend, belkaPct) })
	}
	return pl
}
