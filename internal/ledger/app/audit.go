package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ChangeAuditor keeps the owner's changes together with their entries in the audit
// log (D-211). cmd/core implements it over the identity context's log, for
// the owner's session that asked.
type ChangeAuditor interface {
	// Audited has change made as the owner's change of kind action, naming
	// refs ("<kind>:<ID>"), entered in the audit log: the entry is kept with
	// the change or neither is, and change is not called when the log can't
	// take the entry. An error of change is returned as it is, wrapped.
	Audited(ctx context.Context, origin, sessionToken, action string, refs []string, change func(context.Context) error) error
}

// The audit actions of the owner's changes in the ledger, as the identity
// context spells them.
const (
	auditTransferConfirmed      = "transfer_confirmed"
	auditTransferRejected       = "transfer_rejected"
	auditCashCheckpointRecorded = "cash_checkpoint_recorded"
	auditCashDifferenceBooked   = "cash_difference_booked"
	auditInstrumentRenamed      = "instrument_renamed"
)

// audited runs change as the owner's change of kind action, entered in the
// audit log when a is set. Without a ChangeAuditor, as in the tools that build the
// use cases without a core, nothing is entered.
func audited(ctx context.Context, a ChangeAuditor, origin, sessionToken, action string, refs []string, change func(context.Context) error) error {
	if a == nil {
		return change(ctx)
	}
	if err := a.Audited(ctx, origin, sessionToken, action, refs, change); err != nil {
		return fmt.Errorf("entering %s in the audit log: %w", action, err)
	}
	return nil
}

// accountRefs names accounts as audit references.
func accountRefs(accounts ...ledger.AccountID) []string {
	refs := make([]string, 0, len(accounts))
	for _, a := range accounts {
		refs = append(refs, "account:"+string(a))
	}
	return refs
}
