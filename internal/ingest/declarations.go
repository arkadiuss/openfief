package ingest

import (
	"strconv"
)

// Declarations are what a connector's manifest declares about what it
// sends and takes, beyond its capabilities: whether it may send cross-check
// prices (D-112) and whether it takes statement files the owner uploads
// (D-109). The manifest is signed with the plugin's image (D-42).
type Declarations struct {
	CrossCheckPrices bool
	StatementFiles   bool
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
