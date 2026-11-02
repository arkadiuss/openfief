package xtb

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// The Closed Positions and Open Positions columns the mapper reads.
const (
	colPositionID = "Position ID"
	colCloseTime  = "Close Time (UTC)"
	colProfit     = "Profit/Loss"
)

// closedPositionPrefix starts the external ID of every closed CFD's
// result the mapper books, so it never meets an operation ID of Cash
// Operations.
const closedPositionPrefix = "closed-position:"

// closedPositions books each closed CFD position's result as cash income
// (D-24, D-152), so the account's cash reconciles. A closed stock or ETF
// needs nothing: its purchase and sale are in Cash Operations.
func (m *mapping) closedPositions() {
	if m.x.closed == nil {
		return
	}
	for _, rec := range m.x.closed.records() {
		if c := rec.text(colCategory); !stockCategory(c) {
			ref := m.positionRef(sheetClosed, rec)
			if err := m.closedCFD(rec); err != nil {
				m.problem(problemOf(ref, err))
			}
		}
	}
}

// closedCFD books a closed CFD position's result, its Profit/Loss in the
// account's currency, as cash income on the UTC date it closed (D-152): a
// profit adds cash, a loss takes it out, and one of zero books nothing. It
// is keyed by the position's ID, which it must have.
func (m *mapping) closedCFD(rec record) error {
	id, err := rec.id(colPositionID)
	if err != nil {
		return err
	}
	if id == "" {
		return unsupported("a closed CFD position without a position ID is not supported")
	}
	if !rec.t.has(colProfit) || !rec.t.has(colCloseTime) {
		return unsupported("a closed CFD position without %q and %q is not supported", colProfit, colCloseTime)
	}
	result, err := m.money(rec, colProfit)
	if err != nil {
		return err
	}
	if result.IsZero() {
		return skipped("a closed CFD position of no result")
	}
	on, err := rec.date(colCloseTime, m.x.date1904)
	if err != nil {
		return err
	}
	h := ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: BrokerSource, ExternalID: closedPositionPrefix + id},
		Account: standIn, Date: on,
	}
	t, err := ledger.NewCashIncome(h, result, kernel.Money{})
	if err != nil {
		return fmt.Errorf("%w: %w", errLedger, err)
	}
	t.Account, t.Description = "", "closed CFD position "+shown(rec.text(colTicker))
	m.entries.Transactions = append(m.entries.Transactions, t)
	return nil
}

// openPositions flags open CFD positions as unsupported (D-24). An open
// stock or ETF needs nothing: its purchase is in Cash Operations.
func (m *mapping) openPositions() {
	if m.x.open == nil {
		return
	}
	for _, rec := range m.x.open.records() {
		if c := rec.text(colCategory); !stockCategory(c) {
			m.problem(ingest.Unsupported(m.positionRef(sheetOpen, rec),
				fmt.Sprintf("open CFD position (category %q) is not supported (D-24)", shown(c))))
		}
	}
}

// positionRef names a position row in a problem, with its position ID when
// it has one.
func (m *mapping) positionRef(sheetName string, rec record) string {
	ref := fmt.Sprintf("%s#%d", sheetName, rec.n)
	if rec.t.has(colPositionID) {
		if id, err := rec.id(colPositionID); err == nil && id != "" {
			ref += "[Position ID=" + shown(id) + "]"
		}
	}
	return ref
}
