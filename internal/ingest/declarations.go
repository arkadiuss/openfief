package ingest

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Declarations are what a connector's manifest declares about what it
// sends and takes, beyond its capabilities: whether it may send cross-check
// prices (D-112), whether it takes statement files the owner uploads
// (D-109), whether it is a market-data connector that its grant approves
// as one (D-123), the institutions it books for (D-182), and whether it
// may deliver bond terms and index levels (D-179) and the index readings
// bond terms depend on (D-186). The manifest is signed with the plugin's
// image (D-42).
type Declarations struct {
	CrossCheckPrices bool
	StatementFiles   bool
	// MarketData: the manifest declares a market-data connector and the
	// grant approves it, so each sync tells it the instruments the owner
	// holds and the model portfolio's targets (D-93, D-123).
	MarketData bool
	// BondTerms and IndexLevels: the manifest declares that the connector
	// delivers treasury bond series terms, or market index levels, and the
	// grant approves that, so the core stores them (D-179).
	BondTerms   bool
	IndexLevels bool
	// IndexReadings: the manifest declares that the connector delivers
	// index readings, such as CPI and the NBP reference rate, and the grant
	// approves that, so the core stores them (D-186).
	IndexReadings bool
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

// KindUndeclaredDeliveries is the kind of the attention item (D-130) a
// connector instance raises when it delivers bond terms or index levels
// (D-179), or index readings (D-186), it may not deliver, so they are
// dropped, as cross-check prices are (D-141). Its subject is the instance:
// one item per instance, open while it keeps happening, resolved by the
// instance's next delivery that carries none it may not deliver. The kind
// keeps the name it had before index readings joined it, so an item
// raised earlier still resolves.
const KindUndeclaredDeliveries = "bond-terms-or-index-levels-undeclared"

// UndeclaredDrop is what dropping a delivery's undeclared bond terms,
// index levels (D-179) and index readings (D-186) leaves: the ingest problems, one per kind dropped,
// and the detail of the instance's attention item, empty when nothing is
// dropped.
type UndeclaredDrop struct {
	Problems []Problem
	Detail   string
}

// UndeclaredDeliveries returns what dropping a delivery's terms bond terms,
// levels index levels (D-179) and readings index readings (D-186) that its
// connector may not deliver, as decl says, leaves. The detail names the
// kinds, never a count or a batch, so a later delivery that drops the same
// kinds doesn't reopen an item the owner acknowledged.
func UndeclaredDeliveries(decl Declarations, terms, levels, readings int) UndeclaredDrop {
	var problems []Problem
	var kinds []string
	if terms > 0 && !decl.BondTerms {
		problems = append(problems, Skipped("bond_terms", strconv.Itoa(terms)+
			" bond series terms not stored: the connector's signed manifest doesn't declare them or its grant doesn't approve them"))
		kinds = append(kinds, "bond series terms")
	}
	if levels > 0 && !decl.IndexLevels {
		problems = append(problems, Skipped("index_levels", strconv.Itoa(levels)+
			" index levels not stored: the connector's signed manifest doesn't declare them or its grant doesn't approve them"))
		kinds = append(kinds, "market index levels")
	}
	if readings > 0 && !decl.IndexReadings {
		problems = append(problems, Skipped("index_readings", strconv.Itoa(readings)+
			" index readings not stored: the connector's signed manifest doesn't declare them or its grant doesn't approve them"))
		kinds = append(kinds, "index readings")
	}
	if len(kinds) == 0 {
		return UndeclaredDrop{}
	}
	return UndeclaredDrop{Problems: problems, Detail: "the connector delivers " + joinKinds(kinds) +
		" that its signed manifest doesn't declare or its grant doesn't approve, so they are dropped and the rest of " +
		"its deliveries is imported; install a version that declares them, and approve it"}
}

// joinKinds names one or more kinds in a sentence: "a", "a and b",
// "a, b and c".
func joinKinds(kinds []string) string {
	last := len(kinds) - 1
	if last == 0 {
		return kinds[0]
	}
	return strings.Join(kinds[:last], ", ") + " and " + kinds[last]
}

// KindUndeclaredCrossCheckPrices is the kind of the attention item (D-130)
// a connector instance raises when its manifest doesn't declare the
// cross-check prices it sends, so they are dropped (D-141). Its subject is
// the instance; it resolves on the instance's next delivery that carries
// no undeclared ones.
const KindUndeclaredCrossCheckPrices = "cross-check-prices-undeclared"

// UndeclaredCrossCheckPricesDetail is the detail of that attention item. It
// names no count or batch, so a later delivery that drops prices again
// doesn't reopen an item the owner acknowledged.
const UndeclaredCrossCheckPricesDetail = "the connector sends cross-check prices its signed manifest doesn't declare, " +
	"so they are dropped and the rest of its deliveries is imported; install a version that declares crossCheckPrices"

// UndeclaredCrossCheckPrices is the ingest problem a delivery's n
// cross-check prices become when its connector's manifest doesn't declare
// them (D-112, D-141): none of them is stored, and the rest of the delivery
// is imported.
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
