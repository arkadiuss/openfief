package ingest

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// HeldAccount is what one account bound to a connector instance holds on
// the day of a sync (D-117): its non-zero holdings and open lots, with its
// go-live date and tax wrapper, so that a connector booking what the
// institution pays on a holding, such as a treasury bond's coupons, needs
// no storage of its own. The core sends it; the connector computes.
type HeldAccount struct {
	Account ledger.AccountID
	// GoLive is the account's go-live date, zero while it is not seeded:
	// what fell due on or before it is history its opening lots hold.
	GoLive kernel.Date
	// Wrapper is the account's tax wrapper key, such as "taxable" (D-76).
	Wrapper string
	// Payout is the account's payout setting (D-135): whether what its
	// bonds pay stays as its cash or is paid out.
	Payout   PayoutSetting
	Holdings []ledger.Holding
	// Lots are the account's open lots, oldest first.
	Lots []ledger.Lot
}
