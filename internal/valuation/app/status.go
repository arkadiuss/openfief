package app

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ValuationStatus is how current one account's stored valuations are, for
// the read API (D-36): the date it is valued through and whether a
// revaluation is pending or running.
type ValuationStatus struct {
	Account ledger.AccountID
	// ValuedThrough is the last day whose holdings all have a valuation; zero
	// when the account was never revalued.
	ValuedThrough kernel.Date
	// Pending reports a revaluation waiting to run: a change not yet picked
	// up, or a failed run waiting to be retried.
	Pending bool
	// Running reports a revaluation running now.
	Running bool
}
