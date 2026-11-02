package ingest

import (
	"strconv"
	"strings"
)

// Declarations are what a connector's manifest declares about what it
// sends and takes, beyond its capabilities: whether it may send cross-check
// prices (D-112), whether it takes statement files the owner uploads
// (D-109), whether it is a market-data connector that its grant approves
// as one (D-123), and whether it may deliver bond terms and index levels
// (D-179). The manifest is signed with the plugin's image (D-42).
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
}

// KindUndeclaredDeliveries is the kind of the attention item (D-130) a
// connector instance raises when it delivers bond terms or index levels
// it may not deliver, so they are dropped (D-179), as cross-check prices
// are (D-141). Its subject is the instance: one item per instance, open
// while it keeps happening, resolved by the instance's next delivery that
// carries none it may not deliver.
const KindUndeclaredDeliveries = "bond-terms-or-index-levels-undeclared"

// UndeclaredDrop is what dropping a delivery's undeclared bond terms and
// index levels (D-179) leaves: the ingest problems, one per kind dropped,
// and the detail of the instance's attention item, empty when nothing is
// dropped.
type UndeclaredDrop struct {
	Problems []Problem
	Detail   string
}

// UndeclaredDeliveries returns what dropping a delivery's terms bond terms
// and levels index levels that its connector may not deliver (D-179), as
// decl says, leaves. The detail names the kinds, never a count or a batch,
// so a later delivery that drops the same kinds doesn't reopen an item the
// owner acknowledged.
func UndeclaredDeliveries(decl Declarations, terms, levels int) UndeclaredDrop {
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
	if len(kinds) == 0 {
		return UndeclaredDrop{}
	}
	return UndeclaredDrop{Problems: problems, Detail: "the connector delivers " + strings.Join(kinds, " and ") +
		" that its signed manifest doesn't declare or its grant doesn't approve, so they are dropped and the rest of " +
		"its deliveries is imported; install a version that declares them, and approve it"}
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
