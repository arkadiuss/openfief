package app

import (
	"cmp"
	"context"
	"fmt"
	"strconv"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// MarginalRateSetting is the subject of the missing-setting item for the
// owner's marginal income tax rate (D-78).
const MarginalRateSetting = "marginal-rate"

// Ledger is what the ledger's checks read; *ledger/app.Ledger implements
// it.
type Ledger interface {
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
	// AssertionChecks returns account's balance assertions with their
	// latest check.
	AssertionChecks(ctx context.Context, account ledger.AccountID) ([]ledgerapp.AssertionCheck, error)
	// CurrentTransactions returns the current version of each of
	// account's transactions.
	CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error)
	// TransferLinks returns the current link of each pair the owner
	// decided on (D-105).
	TransferLinks(ctx context.Context) ([]ledger.TransferLink, error)
}

// PriceChecks reads the price checks (D-110..D-112);
// *valuation/app.CrossChecks implements it.
type PriceChecks interface {
	PriceChecks(ctx context.Context) ([]valuation.PriceCheck, error)
}

// Imports reads what the connectors' imports left for the owner;
// *ingest/postgres.Store implements it.
type Imports interface {
	// AllBrokerAccounts returns every instance's broker accounts (D-72).
	AllBrokerAccounts(ctx context.Context) ([]ingest.BrokerAccount, error)
	// PendingTickers returns every ticker without a confirmed ISIN
	// (D-116).
	PendingTickers(ctx context.Context) ([]ingest.PendingTicker, error)
	// UnsupportedBatches returns each completed batch with rows OpenFief
	// could not book yet.
	UnsupportedBatches(ctx context.Context) ([]ingest.BatchProblems, error)
}

// Values reads what the values of D-77 need from the owner;
// *valuation/postgres.Store implements it.
type Values interface {
	// TaxWrapper returns account's tax wrapper; ok is false when none is
	// set.
	TaxWrapper(ctx context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error)
	// ValueSettings returns the owner's settings the values need.
	ValueSettings(ctx context.Context) (jurisdiction.Settings, error)
}

// FlaggedAssertions is the check of balance assertions (D-18): each
// flagged one is a condition, its subject the account, date and ledger
// account, its detail the derived and the reported balance.
func FlaggedAssertions(l Ledger) Check {
	return Check{Kind: attention.KindAssertionFlagged, Find: func(ctx context.Context) ([]attention.Condition, error) {
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading accounts: %w", err)
		}
		var out []attention.Condition
		for i := range accounts {
			checks, err := l.AssertionChecks(ctx, accounts[i].ID)
			if err != nil {
				return nil, fmt.Errorf("loading balance assertions of %s: %w", accounts[i].ID, err)
			}
			for j := range checks {
				if c := &checks[j]; c.Flagged() {
					a := &c.Assertion
					out = append(out, attention.Condition{
						Kind:    attention.KindAssertionFlagged,
						Subject: string(a.Account) + "/" + a.Date.String() + "/" + bookName(a.Book),
						Detail:  "derived " + c.Derived.String() + ", reported " + a.Balance.String(),
					})
				}
			}
		}
		return out, nil
	}}
}

// bookName names a ledger account: "cash:PLN", or the instrument.
func bookName(b ledger.Book) string {
	if b.Kind == ledger.BookCash {
		return "cash:" + b.Currency.String()
	}
	return b.Instrument.String()
}

// BrokenTransfers is the check of confirmed transfers (D-131): each one
// whose legs no longer match after a correction is a condition
// (ledger.BrokenTransferLinks), its subject the pair.
func BrokenTransfers(l Ledger) Check {
	return Check{Kind: attention.KindTransferBroken, Find: func(ctx context.Context) ([]attention.Condition, error) {
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading accounts: %w", err)
		}
		var txs []ledger.Transaction
		for i := range accounts {
			got, txErr := l.CurrentTransactions(ctx, accounts[i].ID)
			if txErr != nil {
				return nil, fmt.Errorf("loading transactions of %s: %w", accounts[i].ID, txErr)
			}
			txs = append(txs, got...)
		}
		links, err := l.TransferLinks(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading transfer links: %w", err)
		}
		broken := ledger.BrokenTransferLinks(links, txs)
		out := make([]attention.Condition, 0, len(broken))
		for _, b := range broken {
			out = append(out, attention.Condition{
				Kind: attention.KindTransferBroken, Subject: b.Pair.String(),
				Detail: "confirmed as version " + strconv.Itoa(b.Version) + "; its legs no longer match",
			})
		}
		return out, nil
	}}
}

// TransferSuggestions reads the transfers suggested between the owner's
// own accounts that the owner has neither confirmed nor rejected (D-105);
// the core computes them from the ledger by the read side's rule, over
// every account.
type TransferSuggestions interface {
	SuggestedTransfers(ctx context.Context) ([]ledger.TransferPair, error)
}

// SuggestedTransferDetail is the detail of every transfer-suggested item:
// the same while the pair stays suggested, so an acknowledged item stays
// acknowledged.
const SuggestedTransferDetail = "suggested; until you confirm it, its legs count as money in and out"

// SuggestedTransfers is the check of suggested transfers (D-105, D-130):
// each pair suggested and not decided yet is a condition, its subject the
// pair as ledger.TransferPair.String writes it, like a broken transfer's.
// Confirming or rejecting the pair ends the suggestion, and the item
// clears with the next sweep.
func SuggestedTransfers(s TransferSuggestions) Check {
	return Check{Kind: attention.KindTransferSuggested, Find: func(ctx context.Context) ([]attention.Condition, error) {
		pairs, err := s.SuggestedTransfers(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading suggested transfers: %w", err)
		}
		out := make([]attention.Condition, 0, len(pairs))
		for _, p := range pairs {
			out = append(out, attention.Condition{
				Kind: attention.KindTransferSuggested, Subject: p.String(), Detail: SuggestedTransferDetail,
			})
		}
		return out, nil
	}}
}

// FlaggedPriceChecks is the check of price checks (D-110..D-112): each
// instrument whose latest check, by date then source, is flagged and not
// acknowledged is a condition, its subject the instrument, its detail the
// check's date, source and difference.
func FlaggedPriceChecks(p PriceChecks) Check {
	return Check{Kind: attention.KindPriceCheckFlagged, Find: func(ctx context.Context) ([]attention.Condition, error) {
		checks, err := p.PriceChecks(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading price checks: %w", err)
		}
		var out []attention.Condition
		for _, c := range latestChecks(checks) {
			if c.Status == valuation.PriceFlagged && !c.Acknowledged {
				out = append(out, attention.Condition{
					Kind: attention.KindPriceCheckFlagged, Subject: c.CrossCheck.Instrument,
					Detail: c.CrossCheck.Date.String() + " " + c.CrossCheck.Source + ": differs by " + c.Difference.String(),
				})
			}
		}
		return out, nil
	}}
}

// latestChecks returns each instrument's latest check among checks, by
// date then source, in the order the instruments first appear.
func latestChecks(checks []valuation.PriceCheck) []*valuation.PriceCheck {
	latest := map[string]int{}
	var out []*valuation.PriceCheck
	for i := range checks {
		c := &checks[i]
		j, seen := latest[c.CrossCheck.Instrument]
		switch {
		case !seen:
			latest[c.CrossCheck.Instrument] = len(out)
			out = append(out, c)
		case cmp.Or(c.CrossCheck.Date.Compare(out[j].CrossCheck.Date), cmp.Compare(c.CrossCheck.Source, out[j].CrossCheck.Source)) > 0:
			out[j] = c
		}
	}
	return out
}

// UnmappedBrokerAccounts is the check of broker accounts (D-72): each one a
// connector reported that no OpenFief account is mapped to is a condition,
// its subject the instance and the broker account.
func UnmappedBrokerAccounts(im Imports) Check {
	return Check{Kind: attention.KindBrokerAccountUnmapped, Find: func(ctx context.Context) ([]attention.Condition, error) {
		all, err := im.AllBrokerAccounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading broker accounts: %w", err)
		}
		var out []attention.Condition
		for i := range all {
			if b := &all[i]; !b.Mapped() {
				out = append(out, attention.Condition{
					Kind: attention.KindBrokerAccountUnmapped, Subject: b.Instance + "/" + string(b.ID),
					Detail: "last seen " + b.LastSeen.String(),
				})
			}
		}
		return out, nil
	}}
}

// UnconfirmedTickers is the check of tickers (D-116): each ticker whose
// ISIN the owner hasn't confirmed is a condition, its subject the instance
// and the ticker, its detail how many rows wait.
func UnconfirmedTickers(im Imports) Check {
	return Check{Kind: attention.KindTickerUnconfirmed, Find: func(ctx context.Context) ([]attention.Condition, error) {
		pending, err := im.PendingTickers(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading pending tickers: %w", err)
		}
		out := make([]attention.Condition, 0, len(pending))
		for i := range pending {
			p := &pending[i]
			out = append(out, attention.Condition{
				Kind: attention.KindTickerUnconfirmed, Subject: p.Instance + "/" + p.Ticker.Symbol,
				Detail: strconv.Itoa(p.Waiting) + " rows wait",
			})
		}
		return out, nil
	}}
}

// ImportProblems is the check of import batches: each completed batch with
// rows OpenFief could not book yet is a condition, its subject the batch,
// its detail the instance and the number of rows. The condition never
// clears, so the owner resolves the item by hand
// (attention.OwnerResolves).
func ImportProblems(im Imports) Check {
	return Check{Kind: attention.KindImportProblems, Find: func(ctx context.Context) ([]attention.Condition, error) {
		batches, err := im.UnsupportedBatches(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading import batches: %w", err)
		}
		out := make([]attention.Condition, 0, len(batches))
		for i := range batches {
			b := &batches[i]
			out = append(out, attention.Condition{
				Kind: attention.KindImportProblems, Subject: string(b.Batch),
				Detail: b.Instance + ": " + strconv.Itoa(b.Unsupported) + " rows left out",
			})
		}
		return out, nil
	}}
}

// Syncs reads where each connector instance's syncs stand;
// *ingest/postgres.Store implements it.
type Syncs interface {
	// SyncStatuses returns every instance's sync status, in instance
	// order.
	SyncStatuses(ctx context.Context) ([]ingest.SyncStatus, error)
}

// FailingSyncs is the check of connector syncs (D-196): each instance whose
// sync keeps failing (ingest.SyncStatus.KeepsFailing) is a condition, its
// subject the instance, its detail when the failures began, to the minute,
// in UTC. The detail stays the same while the same run of failures goes
// on, so an acknowledged item stays acknowledged; the item clears with the
// next success, and a new run of failures raises it again. What happened
// in each attempt is the read API's sync history, not the detail.
func FailingSyncs(s Syncs) Check {
	return Check{Kind: attention.KindSyncFailing, Find: func(ctx context.Context) ([]attention.Condition, error) {
		statuses, err := s.SyncStatuses(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading sync statuses: %w", err)
		}
		var out []attention.Condition
		for i := range statuses {
			if st := &statuses[i]; st.KeepsFailing() {
				out = append(out, attention.Condition{
					Kind: attention.KindSyncFailing, Subject: st.Instance,
					Detail: "failing since " + st.FailingSince.UTC().Format("2006-01-02T15:04Z"),
				})
			}
		}
		return out, nil
	}}
}

// MissingSettings is the check of the owner's settings (D-78): while some
// account's tax wrapper can't have a liquidation value without a setting
// the owner hasn't set, such as the marginal income tax rate for IKZE,
// the setting is a condition, its subject MarginalRateSetting. pack values
// the accounts' cash.
func MissingSettings(l Ledger, v Values, pack jurisdiction.Pack) Check {
	return Check{Kind: attention.KindSettingMissing, Find: func(ctx context.Context) ([]attention.Condition, error) {
		settings, err := v.ValueSettings(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading the owner's settings: %w", err)
		}
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading accounts: %w", err)
		}
		for i := range accounts {
			unknown, valueErr := liquidationUnknown(ctx, v, pack, accounts[i].ID, &settings)
			if valueErr != nil {
				return nil, valueErr
			}
			if unknown {
				return []attention.Condition{{
					Kind: attention.KindSettingMissing, Subject: MarginalRateSetting,
					Detail: "the liquidation value of an account's tax wrapper needs it",
				}}, nil
			}
		}
		return nil, nil
	}}
}

// liquidationUnknown reports whether pack can't give account's cash a
// liquidation value under settings: its tax wrapper, or the pack's default
// when none is set, needs a setting the owner hasn't set.
func liquidationUnknown(ctx context.Context, v Values, pack jurisdiction.Pack, account ledger.AccountID, settings *jurisdiction.Settings) (bool, error) {
	w, recorded, err := v.TaxWrapper(ctx, account)
	if err != nil {
		return false, fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(pack)
	}
	probe := kernel.NewMoney(kernel.NewDecimal(1, 0), pack.ReportingCurrency())
	values, err := pack.ValueCash(&jurisdiction.Cash{Wrapper: w, Amount: probe}, settings)
	if err != nil {
		return false, fmt.Errorf("valuing cash of %s: %w", account, err)
	}
	return values.LiquidationUnknown, nil
}
