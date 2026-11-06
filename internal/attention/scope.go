package attention

import (
	"fmt"
	"strings"
)

// KindSyncFailing is a connector instance whose syncs keep failing
// (D-196, ingest.SyncStatus.KeepsFailing). Subject: the instance.
const KindSyncFailing Kind = "connector-sync-failing"

// Kinds other contexts raise through their own ports, named here so the
// read API can group and scope their items; a test checks them against
// the raising contexts' constants.
const (
	// kindUnclassified is valuation's instrument to classify (D-132).
	// Subject: the ISIN.
	kindUnclassified Kind = "instrument-unclassified"
	// kindMarketDataGap is valuation's series that stopped arriving
	// (D-94). Subject: the account and the series, such as
	// "acc-1/prices/XX0000000001/SYM" or "acc-1/fx_rates/USD".
	kindMarketDataGap Kind = "market-data-gap"
	// kindUndeclared is ingest's deliveries dropped as undeclared (D-179,
	// D-186). Subject: the instance.
	kindUndeclared Kind = "bond-terms-or-index-levels-undeclared"
)

// Group is what the owner does about an item, as the Inbox groups items
// (D-196).
type Group int

// The groups. The zero Group is none.
const (
	// GroupDataComingIn: tickers to confirm, unmapped broker accounts,
	// failing syncs, import problems and market data that stopped or was
	// dropped.
	GroupDataComingIn Group = iota + 1
	// GroupNumbersToCheck: balance mismatches and price differences.
	GroupNumbersToCheck
	// GroupDecisions: transfers to confirm, missing settings, instruments to
	// classify, and any kind the core doesn't know.
	GroupDecisions
)

// String returns "data-coming-in", "numbers-to-check" or "decisions", or a
// placeholder for an invalid group.
func (g Group) String() string {
	switch g {
	case GroupDataComingIn:
		return "data-coming-in"
	case GroupNumbersToCheck:
		return "numbers-to-check"
	case GroupDecisions:
		return "decisions"
	}
	return fmt.Sprintf("Group(%d)", int(g))
}

// GroupOf returns the group items of kind k belong to (D-196). A kind the
// core doesn't know is a decision: the owner decides what it means.
func GroupOf(k Kind) Group {
	switch k {
	case KindTickerUnconfirmed, KindBrokerAccountUnmapped, KindSyncFailing, KindImportProblems,
		kindUndeclared, kindMarketDataGap:
		return GroupDataComingIn
	case KindAssertionFlagged, KindCashCheckpointMismatch, KindPriceCheckFlagged:
		return GroupNumbersToCheck
	case KindTransferBroken, KindTransferSuggested, KindSettingMissing, kindUnclassified:
	}
	return GroupDecisions
}

// Scope is what an item is about, as far as its kind and subject say: the
// account it is tied to, the instrument's ISIN, and the connector instance.
// Each is empty when the item names none. An item tied to no account, such
// as an unmapped broker account, names only its instance (D-165).
type Scope struct {
	Account    string
	Instrument string
	Instance   string
}

// ScopeOf returns the scope of it, read from its kind, its subject and, for
// an import batch's problems, its current detail ("<instance>: ..."), which
// names the batch's instance:
//
//   - balance-assertion-flagged "<account>/<date>/<ISIN>/<symbol>" or
//     ".../cash:<currency>": the account and the instrument;
//   - cash-checkpoint-mismatch "<account>/<date>/<currency>": the account;
//   - price-check-flagged "<ISIN>/<symbol>", instrument-unclassified
//     "<ISIN>": the instrument;
//   - market-data-gap "<account>/<kind>/<key>": the account, and the
//     instrument of a prices series;
//   - broker-account-unmapped and ticker-unconfirmed "<instance>/...",
//     connector-sync-failing and bond-terms-or-index-levels-undeclared
//     "<instance>": the instance;
//   - import-problems: the instance its detail names;
//   - setting-missing, transfer-broken and transfer-suggested (a pair of
//     transaction keys) and any other kind: nothing.
func ScopeOf(it *Item) Scope {
	parts := strings.Split(it.Subject, "/")
	switch it.Kind {
	case KindAssertionFlagged, KindCashCheckpointMismatch, kindMarketDataGap:
		return accountScope(it.Kind, parts)
	case KindPriceCheckFlagged, kindUnclassified:
		return Scope{Instrument: parts[0]}
	case KindBrokerAccountUnmapped, KindTickerUnconfirmed, KindSyncFailing, kindUndeclared:
		return Scope{Instance: parts[0]}
	case KindImportProblems:
		if instance, _, found := strings.Cut(it.Current().Detail, ": "); found {
			return Scope{Instance: instance}
		}
	case KindSettingMissing, KindTransferBroken, KindTransferSuggested:
	}
	return Scope{}
}

// accountScopeParts is how many parts a subject starting with its account
// has at least when it also names an instrument: the account, a date or a
// series kind, then the instrument.
const accountScopeParts = 3

// accountScope returns the scope of an item of kind k whose subject, split
// at "/" into parts, starts with its account: a balance assertion's names
// the instrument of a holding, and a market-data gap's that of a prices
// series.
func accountScope(k Kind, parts []string) Scope {
	s := Scope{Account: parts[0]}
	if len(parts) < accountScopeParts {
		return s
	}
	holding := k == KindAssertionFlagged && !strings.HasPrefix(parts[2], "cash:")
	if holding || (k == kindMarketDataGap && parts[1] == "prices") {
		s.Instrument = parts[2]
	}
	return s
}
