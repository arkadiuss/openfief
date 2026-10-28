package demo

import (
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// market holds the demo's instruments by symbol, their price paths and the
// currencies they trade in.
type market struct {
	bySymbol map[string]*Instrument
	paths    map[*Instrument]pricePath
	usd      kernel.Currency
	eur      kernel.Currency
	pln      kernel.Currency
}

// in returns the instrument with symbol.
func (m *market) in(symbol string) *Instrument { return m.bySymbol[symbol] }

// lotSpec is an opening lot: units of an instrument opened daysBefore go-live
// at costPct percent of the instrument's first close.
type lotSpec struct {
	symbol     string
	units      int64
	daysBefore int
	costPct    int64
}

// openingLots returns the lots of specs, with IDs from s.
func (m *market) openingLots(s *sim, goLive kernel.Date, specs []lotSpec) []lot {
	const pct = 100
	out := make([]lot, 0, len(specs))
	for _, sp := range specs {
		in := m.in(sp.symbol)
		out = append(out, lot{
			in: in, units: sp.units, costCents: sp.units * in.startCents * sp.costPct / pct,
			opened: goLive.AddDays(-sp.daysBefore), id: s.nextID(),
		})
	}
	return out
}

// statement is one IBKR activity statement's content: its period's
// operations and the account at the period's end.
type statement struct {
	from, to kernel.Date
	ops      []op
	end      snapshot
}

// ibkrHistory is the IBKR account's opening state and its monthly
// statements.
type ibkrHistory struct {
	goLive      kernel.Date
	lots        []lot
	openingCash map[kernel.Currency]int64
	statements  []statement
}

// The IBKR account's plan: on which business day of a month each operation
// happens, and its amounts, in hundredths of the currency or in units.
const (
	ibkrDepositDay, ibkrBuyDay, ibkrFXDay, ibkrEuroBuyDay = 2, 4, 7, 9
	ibkrTechDividendDay, ibkrSaleDay, ibkrEuroDividendDay = 10, 11, 12
	ibkrFeeDay                                            = 18

	ibkrDepositBase, ibkrDepositStep, ibkrDepositSteps  = 50_000, 5_000, 5
	ibkrBuyMin, ibkrBuyMax                              = 3, 5
	ibkrCommission, ibkrFXCommission, ibkrDataFee       = 100, 200, 150
	ibkrTechDividend, ibkrEuroDividend, ibkrWithholdPct = 30, 45, 15
	ibkrFXUnits, ibkrFXRateBase, ibkrFXRateSpread       = 500, 10_800, 100
	ibkrEuroBuyUnits, ibkrSaleUnits                     = 8, 5
	ibkrOpeningUSD, ibkrOpeningEUR                      = 180_000, 15_000
	// The months, counted from go-live's, of the one-off operations.
	ibkrFXMonth, ibkrEuroBuyMonth, ibkrSaleMonth = 4, 5, 7
	// quarter makes the tech stock pay in March, June, September and
	// December.
	quarter = 3
)

// simulateIBKR simulates the IBKR account: opening lots in the world ETF,
// the tech stock and the euro dividend stock, then a monthly deposit and ETF
// purchase, a monthly data fee, quarterly dividends with tax withheld, one
// conversion of dollars to euros, one euro stock purchase and one partial
// sale.
func simulateIBKR(m *market, r *rng, goLive, today kernel.Date) ibkrHistory {
	s := newSim(r, m.paths, "61")
	lots := m.openingLots(s, goLive, []lotSpec{
		{symWorld, 25, 420, 80}, {symWorld, 15, 150, 95}, {symTech, 12, 300, 70}, {symEuroDiv, 40, 200, 90},
	})
	h := ibkrHistory{goLive: goLive, lots: lots, openingCash: map[kernel.Currency]int64{m.usd: ibkrOpeningUSD, m.eur: ibkrOpeningEUR}}
	s.open(lots, h.openingCash)
	s.idPrefix, s.seq = "91", 0
	for _, p := range periods(goLive, today) {
		start := len(s.ops)
		pl := ibkrPlan(m, s, &p)
		pl.run()
		h.statements = append(h.statements, statement{from: p.from, to: p.to, ops: s.ops[start:], end: s.snapshot(p.to)})
	}
	return h
}

// ibkrPlan plans one month of the IBKR account.
func ibkrPlan(m *market, s *sim, p *period) plan {
	var pl plan
	deposit := ibkrDepositBase + ibkrDepositStep*s.r.intn(ibkrDepositSteps)
	units := s.r.between(ibkrBuyMin, ibkrBuyMax)
	pl.on(p, ibkrDepositDay, func(d kernel.Date) { s.deposit(d, m.usd, deposit) })
	pl.on(p, ibkrBuyDay, func(d kernel.Date) { s.buy(d, m.in(symWorld), units, ibkrCommission) })
	if p.month%quarter == 0 {
		pl.on(p, ibkrTechDividendDay, func(d kernel.Date) { s.dividend(d, m.in(symTech), ibkrTechDividend, ibkrWithholdPct) })
	}
	if p.month == time.April || p.month == time.October {
		pl.on(p, ibkrEuroDividendDay, func(d kernel.Date) { s.dividend(d, m.in(symEuroDiv), ibkrEuroDividend, ibkrWithholdPct) })
	}
	switch p.index {
	case ibkrFXMonth:
		rate := ibkrFXRateBase + s.r.intn(ibkrFXRateSpread)
		pl.on(p, ibkrFXDay, func(d kernel.Date) { s.fx(d, m.eur, m.usd, ibkrFXUnits, rate, ibkrFXCommission) })
	case ibkrEuroBuyMonth:
		pl.on(p, ibkrEuroBuyDay, func(d kernel.Date) { s.buy(d, m.in(symEuroDiv), ibkrEuroBuyUnits, ibkrCommission) })
	case ibkrSaleMonth:
		pl.on(p, ibkrSaleDay, func(d kernel.Date) { s.sell(d, m.in(symTech), ibkrSaleUnits, ibkrCommission) })
	}
	pl.on(p, ibkrFeeDay, func(d kernel.Date) { s.fee(d, m.usd, ibkrDataFee, "DEMO MARKET DATA FEE") })
	return pl
}
