package ibkrflex

import (
	"strconv"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// IBKR's asset categories and levels of detail that this adapter reads.
const (
	categoryStock = "STK"
	categoryCash  = "CASH"
	levelLot      = "LOT"
	levelSummary  = "SUMMARY"
)

// heldUnits is what OpenPositions reports for one instrument.
type heldUnits struct {
	summary    kernel.Quantity
	hasSummary bool
	lots       kernel.Quantity
}

// positions maps one OpenPositions section.
type positions struct {
	s     *stmt
	held  map[ledger.Instrument]*heldUnits
	order []ledger.Instrument
	// opened counts the lots per instrument and open date, for their keys.
	opened map[string]int
}

// openPositions maps the OpenPositions section: in opening mode each LOT row
// becomes an opening lot (D-23), and in every mode each instrument gets a
// holding balance assertion on the statement's last day (D-18), from its
// SUMMARY row or else the sum of its LOT rows.
func (s *stmt) openPositions(rows []row) {
	p := positions{s: s, held: map[ledger.Instrument]*heldUnits{}, opened: map[string]int{}}
	for i := range rows {
		p.row(&rows[i])
	}
	for _, in := range p.order {
		h := p.held[in]
		units := h.lots
		if h.hasSummary {
			units = h.summary
		}
		s.out.Assertions = append(s.out.Assertions, ledger.HoldingAssertion(s.account, s.to, in, units))
	}
}

// row maps one OpenPositions row.
func (p *positions) row(r *row) {
	ref := p.s.ref("OpenPositions", r, "")
	level := r.text("levelOfDetail")
	if level != levelSummary && level != levelLot && level != "" {
		p.s.problem(ref, skipped("level of detail %q is not read", level))
		return
	}
	in, units, err := position(r)
	if err != nil {
		p.s.problem(ref, err)
		return
	}
	h, ok := p.held[in]
	if !ok {
		h = &heldUnits{}
		p.held[in] = h
		p.order = append(p.order, in)
	}
	if level != levelLot {
		h.summary, h.hasSummary = units, true
		return
	}
	h.lots = h.lots.Add(units)
	if p.s.mode == ingest.ModeOpening {
		t, err := p.openingLot(r, in, units)
		p.s.add(ref, &t, err)
	}
}

// position reads an OpenPosition row's instrument and units: a long position
// in a stock or ETF.
func position(r *row) (ledger.Instrument, kernel.Quantity, error) {
	if r.elem != "OpenPosition" {
		return ledger.Instrument{}, kernel.Quantity{}, skipped("<%s> is not an open position", r.elem)
	}
	if err := needStock(r); err != nil {
		return ledger.Instrument{}, kernel.Quantity{}, err
	}
	in, err := r.instrument()
	if err != nil {
		return ledger.Instrument{}, kernel.Quantity{}, err
	}
	d, err := r.decimal("position")
	if err != nil {
		return ledger.Instrument{}, kernel.Quantity{}, err
	}
	if d.Sign() <= 0 {
		return ledger.Instrument{}, kernel.Quantity{}, unsupported("short or empty positions are not supported")
	}
	return in, kernel.NewQuantity(d), nil
}

// needStock fails unless r is a stock or ETF with no contract multiplier.
func needStock(r *row) error {
	if c := r.text("assetCategory"); c != categoryStock {
		return unsupported("asset category %q is not supported", c)
	}
	if m := r.text("multiplier"); m != "" && m != "1" {
		return unsupported("contract multiplier %q is not supported", m)
	}
	return nil
}

// openingLot returns the opening lot of a LOT row, recorded on the go-live
// date. Its key names the instrument, the open date and IBKR's originating
// transaction ID, or when the row has none, its ordinal among the lots of
// that instrument opened that day.
func (p *positions) openingLot(r *row, in ledger.Instrument, units kernel.Quantity) (ledger.Transaction, error) {
	on, err := r.date("openDateTime")
	if err != nil {
		return ledger.Transaction{}, err
	}
	cost, err := r.money("costBasisMoney")
	if err != nil {
		return ledger.Transaction{}, err
	}
	key := "opening-lot:" + in.ISIN.String() + ":" + on.String()
	p.opened[key]++
	id := r.text("originatingTransactionID")
	if id == "" {
		id = "n" + strconv.Itoa(p.opened[key])
	}
	t, err := ledger.NewOpeningLot(p.s.header(key+":"+id, p.s.to), in, units, cost, on)
	return booked(&t, err, "opening lot")
}
