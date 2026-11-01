package xtb

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// The Closed Positions and Open Positions columns the mapper reads.
const colPositionID = "Position ID"

// closedPositions flags closed CFD positions. A closed stock or ETF needs
// nothing: its purchase and sale are in Cash Operations. A closed CFD's
// result isn't booked yet, so each is an ingest problem (D-24).
func (m *mapping) closedPositions() {
	if m.x.closed == nil {
		return
	}
	for _, rec := range m.x.closed.records() {
		if c := rec.text(colCategory); !stockCategory(c) {
			m.problem(ingest.Unsupported(m.positionRef(sheetClosed, rec),
				fmt.Sprintf("closed CFD position (category %q): its result is not booked yet (D-24)", shown(c))))
		}
	}
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
