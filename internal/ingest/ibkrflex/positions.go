package ibkrflex

import (
	"fmt"
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

// heldUnits is what OpenPositions reports for one instrument: its units,
// and IBKR's mark price from its SUMMARY row, or else from its LOT rows when
// they all agree.
type heldUnits struct {
	summary     kernel.Quantity
	hasSummary  bool
	lots        kernel.Quantity
	summaryMark *ingest.CrossCheckPrice
	lotMark     *ingest.CrossCheckPrice
	lotsDiffer  bool
}

// markOf returns the instrument's cross-check price: its SUMMARY row's mark,
// or else its LOT rows' when they agree. The bool reports LOT rows whose
// marks disagree, with no SUMMARY mark to settle it.
func (h *heldUnits) markOf() (*ingest.CrossCheckPrice, bool) {
	if h.summaryMark != nil {
		return h.summaryMark, false
	}
	if h.lotsDiffer {
		return nil, true
	}
	return h.lotMark, false
}

// addMark records a row's mark price: a SUMMARY row's replaces any other,
// and LOT rows' must all be equal.
func (h *heldUnits) addMark(m *ingest.CrossCheckPrice, summary bool) {
	switch {
	case summary:
		h.summaryMark = m
	case h.lotMark == nil:
		h.lotMark = m
	case !h.lotMark.Date.Equal(m.Date) || !h.lotMark.Close.Equal(m.Close):
		h.lotsDiffer = true
	}
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
// SUMMARY row or else the sum of its LOT rows, and IBKR's mark price as a
// cross-check price (D-22).
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
		mark, differ := h.markOf()
		switch {
		case differ:
			s.problem(fmt.Sprintf("FlexStatement#%d/OpenPositions[isin=%s]", s.index, in.ISIN),
				skipped("the mark prices of the instrument's LOT rows disagree; no cross-check price"))
		case mark != nil:
			s.out.CrossCheckPrices = append(s.out.CrossCheckPrices, *mark)
		}
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
	switch mark, ok, err := p.s.mark(r, in); {
	case err != nil:
		p.s.problem(ref, skipped("no cross-check price: %v", err))
	case ok:
		h.addMark(&mark, level != levelLot)
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

// mark reads an OpenPosition row's markPrice, IBKR's closing price of the
// instrument, as a cross-check price in the row's currency (D-22), dated by
// the row's reportDate or else the statement's last day. The bool is false
// for a row without a markPrice; an unreadable or non-positive one is an error.
func (s *stmt) mark(r *row, in ledger.Instrument) (ingest.CrossCheckPrice, bool, error) {
	if r.text("markPrice") == "" {
		return ingest.CrossCheckPrice{}, false, nil
	}
	closing, err := r.money("markPrice")
	if err != nil {
		return ingest.CrossCheckPrice{}, false, err
	}
	if closing.Sign() <= 0 {
		return ingest.CrossCheckPrice{}, false, fmt.Errorf("%w: markPrice is not positive", errField)
	}
	on := s.to
	if r.text("reportDate") != "" {
		if on, err = r.date("reportDate"); err != nil {
			return ingest.CrossCheckPrice{}, false, err
		}
	}
	return ingest.CrossCheckPrice{Instrument: in, Date: on, Close: closing}, true, nil
}
