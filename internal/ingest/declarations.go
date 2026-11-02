package ingest

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// Declarations are what a connector's manifest declares about what it
// sends and takes, beyond its capabilities: whether it may send cross-check
// prices (D-112), whether it takes statement files the owner uploads
// (D-109), and whether it is a market-data connector that its grant
// approves as one (D-123), and the institutions it books for (D-182). The
// manifest is signed with the plugin's image (D-42).
type Declarations struct {
	CrossCheckPrices bool
	StatementFiles   bool
	// MarketData: the manifest declares a market-data connector and the
	// grant approves it, so each sync tells it the instruments the owner
	// holds and the model portfolio's targets (D-93, D-123).
	MarketData bool
	// Institutions are the institutions the connector books for (D-182).
	Institutions []InstitutionID
}

// ErrUndeclaredInstitution reports a delivery that names an institution its
// connector's manifest doesn't declare (D-182): nothing of it is imported.
var ErrUndeclaredInstitution = errors.New("the connector's manifest doesn't declare the institution its delivery names")

// CheckInstitution refuses a delivery naming institution unless d declares
// it (D-182), so a connector can't book, or rank its market data, as
// another institution's.
func (d *Declarations) CheckInstitution(institution InstitutionID) error {
	if slices.Contains(d.Institutions, institution) {
		return nil
	}
	return fmt.Errorf("%w: %q", ErrUndeclaredInstitution, institution)
}

// UndeclaredCrossCheckPrices is the ingest problem a delivery's n
// cross-check prices become when its connector's manifest doesn't declare
// them (D-112): none of them is stored, and the rest of the delivery is
// imported.
func UndeclaredCrossCheckPrices(n int) Problem {
	return Skipped("cross_check_prices", strconv.Itoa(n)+
		" cross-check prices not stored: the connector's manifest doesn't declare that it sends them")
}

// UploadMode returns the ingest mode of a statement upload for an account
// (D-109, D-95): the owner's override when it is a mode, otherwise opening
// for an account not yet seeded and transactions for one already seeded.
func UploadMode(seeded bool, override Mode) Mode {
	switch override {
	case ModeOpening, ModeTransactions:
		return override
	}
	if seeded {
		return ModeTransactions
	}
	return ModeOpening
}
