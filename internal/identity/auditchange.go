package identity

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// The audit actions of the owner's changes in the other contexts, and of the
// operator's plugin installs (D-211): what the owner changed, by the records
// it names, never an amount, a name, a secret or a question's text. Each is
// written once, as ok, in the transaction that keeps the change from being
// made when the log can't take its entry.
const (
	// ActionSettingsChanged is the owner changing a setting other than the
	// agents-and-models ones (ActionModelSettingsChanged): it names the
	// setting (RefSetting), and the instrument or account it is of, never
	// the value.
	ActionSettingsChanged AuditAction = "settings_changed"
	// ActionAccountCreated is the owner creating an account: the account,
	// the sleeve it was put in and the connector instance bound to it.
	ActionAccountCreated AuditAction = "account_created"
	// ActionAccountChanged is the owner changing an account's name, sleeve,
	// tax wrapper, fee rule, payout setting or earlier payments.
	ActionAccountChanged AuditAction = "account_changed"
	// ActionAccountBound is the owner binding a connector instance to an
	// account: the account and the instance.
	ActionAccountBound AuditAction = "account_bound"
	// ActionAccountDeleted is the owner deleting an account and its records.
	ActionAccountDeleted AuditAction = "account_deleted"
	// ActionBrokerAccountRemapped is the owner mapping a connector's broker
	// account to another account, or leaving it unmapped: the account it now
	// maps to, when there is one, and the instance, never the broker's
	// account number.
	ActionBrokerAccountRemapped AuditAction = "broker_account_remapped"
	// ActionSleeveAdded is a sleeve added.
	ActionSleeveAdded AuditAction = "sleeve_added"
	// ActionSleeveChanged is a sleeve renamed, moved to another kind or
	// given another role.
	ActionSleeveChanged AuditAction = "sleeve_changed"
	// ActionSleeveDeleted is a sleeve deleted.
	ActionSleeveDeleted AuditAction = "sleeve_deleted"
	// ActionSleeveKindAdded is a sleeve kind added.
	ActionSleeveKindAdded AuditAction = "sleeve_kind_added"
	// ActionSleeveKindRenamed is a sleeve kind renamed.
	ActionSleeveKindRenamed AuditAction = "sleeve_kind_renamed"
	// ActionSleeveKindDeleted is a sleeve kind deleted.
	ActionSleeveKindDeleted AuditAction = "sleeve_kind_deleted"
	// ActionTransferConfirmed is the owner confirming that two legs are one
	// transfer between their own accounts: the accounts of the legs.
	ActionTransferConfirmed AuditAction = "transfer_confirmed"
	// ActionTransferRejected is the owner rejecting a suggested transfer.
	ActionTransferRejected AuditAction = "transfer_rejected"
	// ActionCashCheckpointRecorded is the owner entering the cash an
	// institution showed for an account on a day, never the amount.
	ActionCashCheckpointRecorded AuditAction = "cash_checkpoint_recorded"
	// ActionCashDifferenceBooked is the owner booking the difference of a
	// cash checkpoint as a transaction of the account.
	ActionCashDifferenceBooked AuditAction = "cash_difference_booked"
	// ActionInstrumentRenamed is the owner giving an instrument a display
	// name, which the entry names the instrument of and not the name.
	ActionInstrumentRenamed AuditAction = "instrument_renamed"
	// ActionStatementUploaded is a statement the owner uploaded for an
	// account and the core imported.
	ActionStatementUploaded AuditAction = "statement_uploaded"
	// ActionTickerConfirmed is the owner confirming the ticker found for an
	// instrument.
	ActionTickerConfirmed AuditAction = "ticker_confirmed"
	// ActionTickerCorrected is the owner correcting an instrument's ticker.
	ActionTickerCorrected AuditAction = "ticker_corrected"
	// ActionBondLotAdded is the owner entering a bond lot by hand.
	ActionBondLotAdded AuditAction = "bond_lot_added"
	// ActionBondLotDeleted is the owner deleting a hand-entered bond lot.
	ActionBondLotDeleted AuditAction = "bond_lot_deleted"
	// ActionSyncRetried is the owner asking a connector's failed sync to run
	// again.
	ActionSyncRetried AuditAction = "sync_retried"
	// ActionAttentionAcknowledged is the owner acknowledging an attention
	// item.
	ActionAttentionAcknowledged AuditAction = "attention_acknowledged"
	// ActionAttentionResolved is the owner resolving an attention item.
	ActionAttentionResolved AuditAction = "attention_resolved"
	// ActionPriceCheckAcknowledged is the owner acknowledging a price check
	// of an instrument.
	ActionPriceCheckAcknowledged AuditAction = "price_check_acknowledged"
	// ActionPluginInstalled is a Plugin the operator installed, seen by the
	// core as the Plugin's first generation: the instance and the generation.
	ActionPluginInstalled AuditAction = "plugin_installed"
	// ActionPluginUpgraded is a Plugin whose generation changed after the
	// core had seen an earlier one: the operator applied a new spec.
	ActionPluginUpgraded AuditAction = "plugin_upgraded"
	// ActionPluginRemoved is a Plugin the operator removed, seen by the core
	// as the instance gone.
	ActionPluginRemoved AuditAction = "plugin_removed"
)

// The kinds of record the entries above name, beyond those of audit.go.
const (
	// RefInstrument is an instrument, by its ISIN.
	RefInstrument RefKind = "instrument"
	// RefSleeveKind is a sleeve kind, by its key.
	RefSleeveKind RefKind = "sleeve_kind"
	// RefAttentionItem is an attention item, by its number.
	RefAttentionItem RefKind = "attention_item"
	// RefBondLot is a hand-entered bond lot, by its ID.
	RefBondLot RefKind = "bond_lot"
	// RefPlugin is a plugin instance of any kind, by its name.
	RefPlugin RefKind = "plugin"
	// RefPluginGeneration is the generation of a Plugin resource, a decimal
	// number the operator's API server counts: with the instance it says
	// which spec was applied.
	RefPluginGeneration RefKind = "plugin_generation"
	// RefSetting is a setting the owner changed, one of Settings.
	RefSetting RefKind = "setting"
)

// maxGenerationDigits is the most digits of a generation: an int64's.
const maxGenerationDigits = 19

// maxAttentionItemDigits is the most digits of an attention item's number.
const maxAttentionItemDigits = 19

// refRule says how many records of a kind an entry of an action names.
type refRule struct {
	kind     RefKind
	min, max int
}

// one and maybe are the usual rules: a record named once, and one named at
// most once.
func one(k RefKind) refRule   { return refRule{kind: k, min: 1, max: 1} }
func maybe(k RefKind) refRule { return refRule{kind: k, min: 0, max: 1} }

// transferLegs is how many legs a transfer has, and so the most accounts its
// entry names.
const transferLegs = 2

// upToTwo is the rule of a transfer's accounts: those of its legs that are
// stored, none, one or both.
func upToTwo(k RefKind) refRule { return refRule{kind: k, min: 0, max: transferLegs} }

// changeRules maps each action of the owner's changes and the operator's
// installs to the records its entries name: each kind between min and max
// times, and no other. A function, not a variable: the domain holds no
// package-level state.
func changeRules() map[AuditAction][]refRule {
	return map[AuditAction][]refRule{
		ActionSettingsChanged:        {one(RefSetting), maybe(RefInstrument), maybe(RefAccount)},
		ActionAccountCreated:         {one(RefAccount), maybe(RefSleeve), maybe(RefPlugin)},
		ActionAccountChanged:         {one(RefAccount), maybe(RefSleeve)},
		ActionAccountBound:           {one(RefAccount), one(RefPlugin)},
		ActionAccountDeleted:         {one(RefAccount)},
		ActionBrokerAccountRemapped:  {maybe(RefAccount), one(RefPlugin)},
		ActionSleeveAdded:            {one(RefSleeve), maybe(RefSleeveKind)},
		ActionSleeveChanged:          {one(RefSleeve), maybe(RefSleeveKind)},
		ActionSleeveDeleted:          {one(RefSleeve)},
		ActionSleeveKindAdded:        {one(RefSleeveKind)},
		ActionSleeveKindRenamed:      {one(RefSleeveKind)},
		ActionSleeveKindDeleted:      {one(RefSleeveKind)},
		ActionTransferConfirmed:      {upToTwo(RefAccount)},
		ActionTransferRejected:       {upToTwo(RefAccount)},
		ActionCashCheckpointRecorded: {one(RefAccount)},
		ActionCashDifferenceBooked:   {one(RefAccount)},
		ActionInstrumentRenamed:      {one(RefInstrument)},
		ActionStatementUploaded:      {one(RefAccount), maybe(RefPlugin)},
		ActionTickerConfirmed:        {one(RefInstrument)},
		ActionTickerCorrected:        {one(RefInstrument)},
		ActionBondLotAdded:           {one(RefAccount), maybe(RefPlugin)},
		ActionBondLotDeleted:         {one(RefAccount), maybe(RefPlugin), maybe(RefBondLot)},
		ActionSyncRetried:            {one(RefPlugin), maybe(RefAccount)},
		ActionAttentionAcknowledged:  {one(RefAttentionItem)},
		ActionAttentionResolved:      {one(RefAttentionItem)},
		ActionPriceCheckAcknowledged: {one(RefInstrument)},
		ActionPluginInstalled:        {one(RefPlugin), one(RefPluginGeneration)},
		ActionPluginUpgraded:         {one(RefPlugin), one(RefPluginGeneration)},
		ActionPluginRemoved:          {one(RefPlugin), one(RefPluginGeneration)},
	}
}

// changeActionKinds returns, for each action of changeRules, the kinds of
// record its entries may name, for actionKinds.
func changeActionKinds() map[AuditAction][]RefKind {
	rules := changeRules()
	out := make(map[AuditAction][]RefKind, len(rules))
	for a, rs := range rules {
		kinds := make([]RefKind, 0, len(rs))
		for _, r := range rs {
			kinds = append(kinds, r.kind)
		}
		out[a] = kinds
	}
	return out
}

// isChangeAction reports whether a is an action of changeRules.
func isChangeAction(a AuditAction) bool {
	_, ok := changeRules()[a]
	return ok
}

// validateChange checks that e, of an action of changeRules, names its
// records as many times as the action's rule says, and is ok: a change is
// only entered once it was made.
func validateChange(e *AuditEntry) error {
	if e.Outcome != OutcomeOK {
		return fmt.Errorf("%w: a %s entry is ok", ErrInvalidAuditEntry, e.Action)
	}
	rules := changeRules()[e.Action]
	for _, r := range rules {
		n := 0
		for _, ref := range e.Refs {
			if ref.Kind == r.kind {
				n++
			}
		}
		if n < r.min || n > r.max {
			return fmt.Errorf("%w: a %s entry names a %s %d to %d times", ErrInvalidAuditEntry, e.Action, r.kind, r.min, r.max)
		}
	}
	if e.Action == ActionSettingsChanged {
		return validateSettingRefs(e)
	}
	return nil
}

// settingRefs maps each setting to the record kind its entry must also name,
// where the setting is of one: an instrument's exemption class, early
// redemption fee and price check threshold, and an account's chart start.
func settingRefs() map[string]RefKind {
	return map[string]RefKind{
		"tax:exemption_class":      RefInstrument,
		"tax:early_redemption_fee": RefInstrument,
		"pricecheck:threshold":     RefInstrument,
		"chart:start":              RefAccount,
	}
}

// validateSettingRefs checks that a settings entry names the instrument or
// account its setting is of, and none for a setting that is of neither.
func validateSettingRefs(e *AuditEntry) error {
	var setting string
	for _, r := range e.Refs {
		if r.Kind == RefSetting {
			setting = r.ID
		}
	}
	want, of := settingRefs()[setting]
	for _, r := range e.Refs {
		if r.Kind == RefSetting {
			continue
		}
		if !of || r.Kind != want {
			return fmt.Errorf("%w: a %s entry names only the record its setting is of", ErrInvalidAuditEntry, e.Action)
		}
	}
	if of && !slices.ContainsFunc(e.Refs, func(r Ref) bool { return r.Kind == want }) {
		return fmt.Errorf("%w: a %s entry names the %s its setting is of", ErrInvalidAuditEntry, e.Action, want)
	}
	return nil
}

// Settings are the settings a ActionSettingsChanged entry may name, a closed
// set: the number format, the marginal rate, an instrument's exemption class
// and early redemption fee, an account's chart start, an instrument's price
// check threshold, the ranking of each kind of market data and its gap
// threshold, the source of a series, and the model portfolio.
func Settings() []string {
	base := []string{
		"display:number_format", "tax:marginal_rate", "tax:exemption_class", "tax:early_redemption_fee",
		"chart:start", "chart:portfolio_start", "pricecheck:threshold", "marketdata:series_source", "model_portfolio",
	}
	out := slices.Clone(base)
	for _, kind := range []string{"prices", "fx_rates", "index_readings"} {
		out = append(out, "marketdata:ranking:"+kind, "marketdata:gap:"+kind)
	}
	return out
}

// validateChangeRef checks r's ID when r is of a kind of audit changes: an
// ID of the shape its kind has, never a name, an amount or a secret.
func (r Ref) validateChangeRef() error {
	var ok bool
	switch r.Kind {
	case RefInstrument:
		ok = isISINShape(r.ID)
	case RefSleeveKind, RefBondLot:
		ok = isRecordID(r.ID) && !looksLikeSecret(r.ID)
	case RefAttentionItem:
		ok = isPositiveDecimal(r.ID, maxAttentionItemDigits)
	case RefPlugin:
		ok = isPluginInstanceName(r.ID)
	case RefPluginGeneration:
		ok = isPositiveDecimal(r.ID, maxGenerationDigits)
	case RefSetting:
		ok = slices.Contains(Settings(), r.ID)
	case RefSession, RefPasskey, RefEnrollmentCode, RefScriptToken, RefPerson, RefViewPlugin, RefAccount, RefSleeve,
		RefHousehold, RefModelRole, RefModel, RefModelCall, RefModelError, RefMaintainerRun, RefModelSetting:
	}
	if !ok {
		return fmt.Errorf("%w: a %s reference is the ID of its shape", ErrInvalidAuditEntry, r.Kind)
	}
	return nil
}

// isISINShape reports whether s has an ISIN's shape: two upper-case letters,
// nine upper-case letters or digits and a check digit. The check digit is not
// verified here, which is the ledger's.
func isISINShape(s string) bool {
	const isinLen = 12
	if len(s) != isinLen {
		return false
	}
	for i := range len(s) {
		c := s[i]
		upper := c >= 'A' && c <= 'Z'
		digit := c >= '0' && c <= '9'
		switch {
		case i < 2 && !upper, i == isinLen-1 && !digit, !upper && !digit:
			return false
		}
	}
	return true
}

// isPositiveDecimal reports whether s is a decimal number of 1 to maxDigits
// digits with no sign and no leading zero, above zero.
func isPositiveDecimal(s string, maxDigits int) bool {
	if s == "" || len(s) > maxDigits || strings.HasPrefix(s, "0") {
		return false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return err == nil && n > 0 && strconv.FormatInt(n, 10) == s
}

// ParseRefKind parses a kind of record as its constant spells it.
func ParseRefKind(s string) (RefKind, error) {
	k := RefKind(s)
	switch k {
	case RefSession, RefPasskey, RefEnrollmentCode, RefScriptToken, RefPerson, RefViewPlugin, RefAccount, RefSleeve,
		RefHousehold, RefModelRole, RefModel, RefModelCall, RefModelError, RefMaintainerRun, RefModelSetting,
		RefInstrument, RefSleeveKind, RefAttentionItem, RefBondLot, RefPlugin, RefPluginGeneration, RefSetting:
		return k, nil
	}
	return "", fmt.Errorf("%w: unknown reference kind", ErrInvalidAuditEntry)
}
