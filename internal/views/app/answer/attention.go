package answer

import (
	"fmt"
	"strings"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The core's kinds of attention items the views tell apart (D-130, D-131,
// D-132, D-94, D-105, D-158, D-179, D-196, D-204). The Inbox and the Ask
// box's "what needs me" name an item in the same words, so they share this
// table.
const (
	KindAssertion         = "balance-assertion-flagged"
	KindPriceCheck        = "price-check-flagged"
	KindImportProblems    = "import-problems"
	KindBrokerUnmapped    = "broker-account-unmapped"
	KindTicker            = "ticker-unconfirmed"
	KindSettingMissing    = "setting-missing"
	KindTransferBroken    = "transfer-broken"
	KindTransferSuggested = "transfer-suggested"
	KindUnclassified      = "instrument-unclassified"
	KindSyncFailing       = "connector-sync-failing"
	KindCashCheckpoint    = "cash-checkpoint-mismatch"
	KindBondFeeAssumed    = "bond-fee-assumed"
	KindBondSeries        = "bond-series-incomplete"
	// KindViewPluginUngranted is a view plugin that can't read anything until
	// the owner grants it accounts (D-196, D-209).
	KindViewPluginUngranted = "view-plugin-ungranted"
	// KindAuditLogBehind is the audit entries waiting in the outbox through
	// several failed attempts to enter the log (D-196, D-211).
	KindAuditLogBehind = "audit-log-behind"
	// KindTargetProposal and KindMaintainerFailing are the kinds the
	// model-portfolio maintainer raises (D-196, D-204): a proposal of target
	// changes for the owner to apply or dismiss, and a weekly run that keeps
	// failing.
	KindTargetProposal    = "model-target-proposal"
	KindMaintainerFailing = "maintainer-failing"
)

// KindLabel names kind, one of the core's kinds of attention items; a kind it
// doesn't know shows its own name.
func KindLabel(kind string) string {
	labels := map[string]string{
		KindAssertion:                           "Flagged balance assertion",
		KindPriceCheck:                          "Flagged price check",
		KindImportProblems:                      "Rows an import left out",
		KindBrokerUnmapped:                      "Unmapped broker account",
		KindTicker:                              "Ticker to confirm",
		KindSettingMissing:                      "Missing setting",
		KindTransferBroken:                      "Transfer broken by a correction",
		KindTransferSuggested:                   "Transfer to confirm",
		KindUnclassified:                        "Instrument to classify",
		"market-data-gap":                       "Market data stopped arriving",
		"bond-terms-or-index-levels-undeclared": "Bond terms, index levels or readings dropped",
		KindSyncFailing:                         "Sync keeps failing",
		KindCashCheckpoint:                      "Cash balance differs",
		KindBondFeeAssumed:                      "Bond fee to confirm",
		KindBondSeries:                          "Bond series still without its ISIN or terms",
		"model-route-broken":                    "Model route names a missing provider instance",
		KindTargetProposal:                      "Target weights proposed by the maintainer",
		KindMaintainerFailing:                   "Maintainer's weekly run failing",
		KindViewPluginUngranted:                 "View plugin without access to any account",
		"cross-check-prices-undeclared":         "Undeclared cross-check prices dropped",
		KindAuditLogBehind:                      "Audit log is behind",
	}
	if l, ok := labels[kind]; ok {
		return l
	}
	return kind
}

// fixedTitles are the titles of the kinds whose title names nothing of
// their item.
func fixedTitles() map[string]string {
	return map[string]string{
		KindImportProblems:    "An import left rows out",
		KindTransferBroken:    "A confirmed transfer no longer matches",
		KindTransferSuggested: "Confirm a transfer between your accounts",
		KindTargetProposal:    "The maintainer proposes new target weights",
		KindMaintainerFailing: "The maintainer's weekly run keeps failing",
	}
}

// subjectTitles are the titles of the kinds whose subject says what they are
// about, an instrument's ISIN or a view plugin's instance, as formats of it.
func subjectTitles() map[string]string {
	return map[string]string{
		KindUnclassified:        "Classify %s for the yearly tax",
		KindBondFeeAssumed:      "Confirm the early-redemption fee of %s",
		KindViewPluginUngranted: "Grant %s access to your accounts",
	}
}

// ItemTitle says in a few words what an item asks of the owner.
func ItemTitle(it *viewsv1.AttentionItem) string {
	if title, ok := fixedTitles()[it.GetKind()]; ok {
		return title
	}
	subject := it.GetSubject()
	if format, ok := subjectTitles()[it.GetKind()]; ok {
		return fmt.Sprintf(format, subject)
	}
	switch it.GetKind() {
	case KindTicker:
		_, symbol, _ := strings.Cut(subject, "/")
		return "Confirm the ISIN of " + symbol
	case KindBrokerUnmapped:
		return "Map a broker account " + it.GetConnectorInstance() + " reported"
	case KindSyncFailing:
		return "The " + it.GetConnectorInstance() + " sync keeps failing"
	case KindPriceCheck:
		return "The price of " + subject + " differs"
	case KindAssertion:
		return "A balance in " + it.GetAccountId() + " doesn't match"
	case KindCashCheckpoint:
		return "The cash balance you entered for " + it.GetAccountId() + " differs"
	case KindSettingMissing:
		return settingTitle(subject)
	}
	return KindLabel(it.GetKind())
}

// settingTitle titles a missing setting's item by its subject: an IKE's
// missing earlier payments, "earlier-payments/<account>", by its account
// (D-149), any other by the setting.
func settingTitle(subject string) string {
	if account, ok := strings.CutPrefix(subject, "earlier-payments/"); ok {
		return "Enter the earlier payments into " + account
	}
	return "Set " + subject
}

// Tone is how an item's status pill is toned: the page maps it to a style.
type Tone int

// The tones of a status pill.
const (
	ToneNeutral Tone = iota
	ToneOK
	ToneWarn
	ToneDanger
)

// LabelDone names the done view and a done item's pill.
const LabelDone = "Done"

// ItemStatus returns an item's status pill: its word and its tone.
func ItemStatus(it *viewsv1.AttentionItem) (string, Tone) {
	switch it.GetState() {
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_ACKNOWLEDGED:
		return "Acknowledged", ToneNeutral
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED:
		return LabelDone, ToneOK
	case viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_UNSPECIFIED:
	}
	if it.GetKind() == KindSyncFailing {
		return "Failing", ToneDanger
	}
	return "Needs you", ToneWarn
}

// itemWord is the word of an item's status pill.
func itemWord(it *viewsv1.AttentionItem) string {
	word, _ := ItemStatus(it)
	return word
}

// The Inbox's three groups (D-196), in the order it shows them.
const (
	GroupDataComingIn = iota
	GroupNumbersToCheck
	GroupDecisions
)

// GroupIndex returns the index of the group items of g are shown in; an item
// without a group is a decision, as a kind the core doesn't know is.
func GroupIndex(g viewsv1.AttentionItemGroup) int {
	switch g {
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN:
		return GroupDataComingIn
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK:
		return GroupNumbersToCheck
	case viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS, viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_UNSPECIFIED:
	}
	return GroupDecisions
}

// GroupTitles are the titles of the Inbox's groups, by GroupIndex.
func GroupTitles() []string {
	return []string{"Data coming in", "Numbers to check", "Decisions"}
}

// GroupTitle names an attention item's group as the Inbox does.
func GroupTitle(g viewsv1.AttentionItemGroup) string {
	return GroupTitles()[GroupIndex(g)]
}
