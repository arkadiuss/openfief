package identity_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Synthetic records of the owner's changes: IDs of the shapes the entries
// hold, none an amount, a name or a secret.
const (
	changeAccount    = "xtb-main"
	changeOther      = "bonds-main"
	changeSleeve     = "growth"
	changeISIN       = "PLTEST000010"
	changeInstance   = "xtb-flex"
	changeGeneration = "3"
)

// TestChangeEntriesNameWhatChangedByID checks the entries of the owner's
// changes in the other contexts and of the operator's installs (D-211): each
// action names the records its rule says, once ok, by IDs of their shapes.
func TestChangeEntriesNameWhatChangedByID(t *testing.T) {
	t.Parallel()
	owner := identity.Actor{Kind: identity.ActorOwner, ID: auditSession}
	operator := identity.Actor{Kind: identity.ActorOperator, ID: identity.OperatorActorID}
	ref := func(k identity.RefKind, id string) identity.Ref { return identity.Ref{Kind: k, ID: id} }
	tests := []struct {
		name    string
		actor   identity.Actor
		action  identity.AuditAction
		outcome identity.AuditOutcome
		refs    []identity.Ref
		wantErr bool
	}{
		{
			name: "a setting", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "tax:marginal_rate")},
		},
		{
			name: "an instrument's setting names the instrument", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "tax:exemption_class"), ref(identity.RefInstrument, changeISIN)},
		},
		{
			name: "an instrument's setting without the instrument", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "tax:exemption_class")}, wantErr: true,
		},
		{
			name: "a global setting naming an instrument", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "tax:marginal_rate"), ref(identity.RefInstrument, changeISIN)}, wantErr: true,
		},
		{
			name: "an account's chart start", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "chart:start"), ref(identity.RefAccount, changeAccount)},
		},
		{
			name: "the portfolio's chart start", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "chart:portfolio_start")},
		},
		{
			name: "a setting nobody knows", actor: owner, action: identity.ActionSettingsChanged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSetting, "tax:everything")}, wantErr: true,
		},
		{
			name: "an account created, with its sleeve and connector", actor: owner, action: identity.ActionAccountCreated, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount), ref(identity.RefSleeve, changeSleeve), ref(identity.RefPlugin, changeInstance)},
		},
		{
			name: "an account bound", actor: owner, action: identity.ActionAccountBound, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount), ref(identity.RefPlugin, changeInstance)},
		},
		{
			name: "an account bound to no instance", actor: owner, action: identity.ActionAccountBound, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount)}, wantErr: true,
		},
		{
			name: "a broker account left unmapped", actor: owner, action: identity.ActionBrokerAccountRemapped, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, changeInstance)},
		},
		{
			name: "a sleeve kind renamed", actor: owner, action: identity.ActionSleeveKindRenamed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefSleeveKind, "stocks")},
		},
		{
			name: "a transfer confirmed between two accounts", actor: owner, action: identity.ActionTransferConfirmed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount), ref(identity.RefAccount, changeOther)},
		},
		{name: "a transfer rejected with no leg stored", actor: owner, action: identity.ActionTransferRejected, outcome: identity.OutcomeOK},
		{
			name: "a transfer of three accounts", actor: owner, action: identity.ActionTransferConfirmed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, "a"), ref(identity.RefAccount, "b"), ref(identity.RefAccount, "c")}, wantErr: true,
		},
		{
			name: "an instrument renamed", actor: owner, action: identity.ActionInstrumentRenamed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefInstrument, changeISIN)},
		},
		{
			name: "an instrument by a name", actor: owner, action: identity.ActionInstrumentRenamed, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefInstrument, "Some Fund")}, wantErr: true,
		},
		{
			name: "an inbox item acknowledged", actor: owner, action: identity.ActionAttentionAcknowledged, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAttentionItem, "42")},
		},
		{
			name: "an inbox item by a sentence", actor: owner, action: identity.ActionAttentionResolved, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAttentionItem, "the fund needs a ticker")}, wantErr: true,
		},
		{
			name: "an inbox item numbered zero", actor: owner, action: identity.ActionAttentionResolved, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAttentionItem, "0")}, wantErr: true,
		},
		{
			name: "a bond lot deleted", actor: owner, action: identity.ActionBondLotDeleted, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount), ref(identity.RefPlugin, changeInstance)},
		},
		{
			name: "a sync retried", actor: owner, action: identity.ActionSyncRetried, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, changeInstance)},
		},
		{
			name: "a plugin installed by the operator", actor: operator, action: identity.ActionPluginInstalled, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, changeInstance), ref(identity.RefPluginGeneration, changeGeneration)},
		},
		{
			name: "a plugin removed by the operator", actor: operator, action: identity.ActionPluginRemoved, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, changeInstance), ref(identity.RefPluginGeneration, changeGeneration)},
		},
		{
			name: "a plugin upgraded without its generation", actor: operator, action: identity.ActionPluginUpgraded, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, changeInstance)}, wantErr: true,
		},
		{
			name: "a plugin named like a secret", actor: operator, action: identity.ActionPluginInstalled, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefPlugin, "Not A Plugin"), ref(identity.RefPluginGeneration, changeGeneration)}, wantErr: true,
		},
		{
			name: "a change that was refused is no entry", actor: owner, action: identity.ActionAccountDeleted, outcome: identity.OutcomeRefused,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount)}, wantErr: true,
		},
		{
			name: "a change that failed is no entry", actor: owner, action: identity.ActionAccountDeleted, outcome: identity.OutcomeFailed,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount)}, wantErr: true,
		},
		{
			name: "a record the action does not name", actor: owner, action: identity.ActionAccountDeleted, outcome: identity.OutcomeOK,
			refs: []identity.Ref{ref(identity.RefAccount, changeAccount), ref(identity.RefSleeve, changeSleeve)}, wantErr: true,
		},
		{
			name: "the operator's ID is its own", actor: identity.Actor{Kind: identity.ActorOperator, ID: "core"}, action: identity.ActionPluginRemoved,
			outcome: identity.OutcomeOK, refs: []identity.Ref{ref(identity.RefPlugin, changeInstance), ref(identity.RefPluginGeneration, changeGeneration)}, wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e, err := identity.NewAuditEntry(auditAt(), tt.actor, tt.action, tt.outcome, tt.refs...)
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, identity.ErrInvalidAuditEntry)) {
				t.Fatalf("NewAuditEntry = %+v, %v; wantErr %v", e, err, tt.wantErr)
			}
		})
	}
}

// TestEveryChangeActionIsKnownAndParses checks that each action of the
// owner's changes is among AuditActions and parses back.
func TestEveryChangeActionIsKnownAndParses(t *testing.T) {
	t.Parallel()
	actions := identity.AuditActions()
	for _, a := range []identity.AuditAction{
		identity.ActionSettingsChanged, identity.ActionAccountCreated, identity.ActionAccountChanged, identity.ActionAccountBound,
		identity.ActionAccountDeleted, identity.ActionBrokerAccountRemapped, identity.ActionSleeveAdded, identity.ActionSleeveChanged,
		identity.ActionSleeveDeleted, identity.ActionSleeveKindAdded, identity.ActionSleeveKindRenamed, identity.ActionSleeveKindDeleted,
		identity.ActionTransferConfirmed, identity.ActionTransferRejected, identity.ActionCashCheckpointRecorded,
		identity.ActionCashDifferenceBooked, identity.ActionInstrumentRenamed, identity.ActionStatementUploaded,
		identity.ActionTickerConfirmed, identity.ActionTickerCorrected, identity.ActionBondLotAdded, identity.ActionBondLotDeleted,
		identity.ActionSyncRetried, identity.ActionAttentionAcknowledged, identity.ActionAttentionResolved,
		identity.ActionPriceCheckAcknowledged, identity.ActionPluginInstalled, identity.ActionPluginUpgraded, identity.ActionPluginRemoved,
	} {
		if !slices.Contains(actions, a) {
			t.Errorf("%s is not among AuditActions", a)
		}
		if got, err := identity.ParseAuditAction(string(a)); err != nil || got != a {
			t.Errorf("ParseAuditAction(%q) = %q, %v", a, got, err)
		}
	}
}

// TestSettingsAndReferenceKindsAreClosedSets checks that each setting is a
// valid reference, and that the kinds of record parse and nothing else does.
func TestSettingsAndReferenceKindsAreClosedSets(t *testing.T) {
	t.Parallel()
	for _, s := range identity.Settings() {
		if err := (identity.Ref{Kind: identity.RefSetting, ID: s}).Validate(); err != nil {
			t.Errorf("setting %q is no valid reference: %v", s, err)
		}
	}
	if _, err := identity.ParseRefKind("amount"); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("ParseRefKind(amount) = %v, want ErrInvalidAuditEntry", err)
	}
	for _, k := range []identity.RefKind{identity.RefInstrument, identity.RefSleeveKind, identity.RefAttentionItem, identity.RefBondLot, identity.RefPlugin, identity.RefPluginGeneration, identity.RefSetting} {
		if got, err := identity.ParseRefKind(string(k)); err != nil || got != k {
			t.Errorf("ParseRefKind(%q) = %q, %v", k, got, err)
		}
	}
}
