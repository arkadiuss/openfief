package ingest

import (
	"strconv"
)

// Declarations are what a connector's manifest declares about what it
// sends and takes, beyond its capabilities: whether it may send cross-check
// prices (D-112), whether it takes statement files the owner uploads
// (D-109), and whether it is a market-data connector that its grant
// approves as one (D-123). The manifest is signed with the plugin's image
// (D-42).
type Declarations struct {
	CrossCheckPrices bool
	StatementFiles   bool
	// MarketData: the manifest declares a market-data connector and the
	// grant approves it, so each sync tells it the instruments the owner
	// holds and the model portfolio's targets (D-93, D-123).
	MarketData bool
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
