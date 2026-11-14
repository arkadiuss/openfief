package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ChangeAuditor keeps the owner's changes together with their entries in the
// audit log (D-211). cmd/core implements it over the identity context's log,
// for the owner's session that asked.
type ChangeAuditor interface {
	// Audited has change made as the owner's change of kind action, naming
	// refs ("<kind>:<ID>"), entered in the audit log: the entry is kept with
	// the change or neither is, and change is not called when the log can't
	// take the entry. An error of change is returned as it is, wrapped.
	Audited(ctx context.Context, origin, sessionToken, action string, refs []string, change func(context.Context) error) error
}

// The audit actions of the owner's changes in the ingest context, as the
// identity context spells them.
const (
	auditAccountCreated        = "account_created"
	auditAccountChanged        = "account_changed"
	auditAccountBound          = "account_bound"
	auditAccountDeleted        = "account_deleted"
	auditBrokerAccountRemapped = "broker_account_remapped"
	auditStatementUploaded     = "statement_uploaded"
	auditTickerConfirmed       = "ticker_confirmed"
	auditTickerCorrected       = "ticker_corrected"
	auditBondLotAdded          = "bond_lot_added"
	auditBondLotDeleted        = "bond_lot_deleted"
	auditSyncRetried           = "sync_retried"
	auditAccountRefPrefix      = "account:"
	auditPluginRefPrefix       = "plugin:"
	auditInstrumentRefPrefix   = "instrument:"
)

// audited runs change as the owner's change of kind action, entered in the
// audit log when a is set. Without a ChangeAuditor, as in the tools that
// build the use cases without a core, nothing is entered.
func audited(ctx context.Context, a ChangeAuditor, origin, sessionToken, action string, refs []string, change func(context.Context) error) error {
	if a == nil {
		return change(ctx)
	}
	if err := a.Audited(ctx, origin, sessionToken, action, refs, change); err != nil {
		return fmt.Errorf("entering %s in the audit log: %w", action, err)
	}
	return nil
}

// accountRef names an account as an audit reference.
func accountRef(id ledger.AccountID) string { return auditAccountRefPrefix + string(id) }

// createdRefs names what a created account is: the account, and the connector
// instance bound to it when it was bound.
func createdRefs(id ledger.AccountID, instance string) []string {
	if instance == "" {
		return []string{accountRef(id)}
	}
	return []string{accountRef(id), instanceRef(instance)}
}

// instanceRef names a connector instance as an audit reference.
func instanceRef(instance string) string { return auditPluginRefPrefix + instance }
