package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// AttentionItems reads the attention items (D-130);
// *attention/app.Items implements it.
type AttentionItems interface {
	// List returns every item, in ID order, with its history.
	List(ctx context.Context) ([]attention.Item, error)
}

// WithAttention lists the attention items from items; without it the read
// API lists none.
func WithAttention(items AttentionItems) Option {
	return func(v *Views) { v.attention = items }
}

// AttentionView is which items to list, as the Inbox's views split them
// (D-196). The zero view is every item.
type AttentionView int

// The views.
const (
	// AttentionOpen lists the open items.
	AttentionOpen AttentionView = iota + 1
	// AttentionAcknowledged lists the acknowledged items.
	AttentionAcknowledged
	// AttentionDone lists the resolved items.
	AttentionDone
)

// AttentionQuery selects attention items: those in View, and, when set,
// those tied to Account (any of an item's accounts), and those whose scope
// (attention.ScopeOf) names Instrument (an ISIN) or Instance.
type AttentionQuery struct {
	View       AttentionView
	Account    string
	Instrument string
	Instance   string
}

// ScopedItem is an attention item with its scope and group, and every
// account it is tied to: its scope's account, or both legs' accounts of a
// broken transfer (D-131), whose scope then names the withdrawal's.
type ScopedItem struct {
	attention.Item

	Scope    attention.Scope
	Group    attention.Group
	Accounts []string
}

// AttentionList is the attention items a query selects, in ID order, and
// how many of those the query's scope selects are open, acknowledged and
// done, whatever its view.
type AttentionList struct {
	Items        []ScopedItem
	Open         int
	Acknowledged int
	Done         int
}

// Attention returns the attention items q selects with their history, in
// ID order. A broken transfer's item is tied to both its legs' accounts
// (views.PairAccounts). An item tied to an account the caller may not view
// is left out (D-163), so a broken transfer's is listed only when the caller
// may view both, as Transfers lists the pair; an item tied to no account
// names only its connector instance, as attention.ScopeOf reads it (D-165).
func (v *Views) Attention(ctx context.Context, q *AttentionQuery) (AttentionList, error) {
	if v.attention == nil {
		return AttentionList{}, nil
	}
	items, err := v.attention.List(ctx)
	if err != nil {
		return AttentionList{}, fmt.Errorf("loading attention items: %w", err)
	}
	pairs, err := v.brokenPairAccounts(ctx, items)
	if err != nil {
		return AttentionList{}, err
	}
	var out AttentionList
	for i := range items {
		it := scopedItem(&items[i], pairs)
		if !q.scopes(&it) {
			continue
		}
		visible, authzErr := v.mayViewAll(ctx, it.Accounts)
		if authzErr != nil {
			return AttentionList{}, authzErr
		}
		if visible {
			out.add(&it, q.View)
		}
	}
	return out, nil
}

// scopedItem returns it with its scope, group and accounts; pairs are the
// accounts of each confirmed pair's legs, by pair.
func scopedItem(it *attention.Item, pairs map[string][]ledger.AccountID) ScopedItem {
	out := ScopedItem{Item: *it, Scope: attention.ScopeOf(it), Group: attention.GroupOf(it.Kind)}
	if it.Kind == attention.KindTransferBroken {
		for _, a := range pairs[it.Subject] {
			out.Accounts = append(out.Accounts, string(a))
		}
		if len(out.Accounts) > 0 {
			out.Scope.Account = out.Accounts[0]
		}
		return out
	}
	if out.Scope.Account != "" {
		out.Accounts = []string{out.Scope.Account}
	}
	return out
}

// brokenPairAccounts returns the accounts of each confirmed pair's legs
// (views.PairAccounts), read only when items hold a broken transfer.
func (v *Views) brokenPairAccounts(ctx context.Context, items []attention.Item) (map[string][]ledger.AccountID, error) {
	if !slices.ContainsFunc(items, func(it attention.Item) bool { return it.Kind == attention.KindTransferBroken }) {
		return map[string][]ledger.AccountID{}, nil
	}
	accounts, err := v.ledger.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	var txs []ledger.Transaction
	for i := range accounts {
		got, txErr := v.ledger.CurrentTransactions(ctx, accounts[i].ID)
		if txErr != nil {
			return nil, fmt.Errorf("loading transactions of %s: %w", accounts[i].ID, txErr)
		}
		txs = append(txs, got...)
	}
	links, err := v.ledger.TransferLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading transfer links: %w", err)
	}
	return views.PairAccounts(links, txs), nil
}

// mayViewAll reports whether the caller in ctx may view an item tied to
// accounts: one tied to some only when it may view every one of them
// (D-163); one tied to none, as every holder of read access may today
// (D-165).
func (v *Views) mayViewAll(ctx context.Context, accounts []string) (bool, error) {
	caller := identityapp.CallerFrom(ctx)
	for _, a := range accounts {
		ok, err := v.authz.MayAct(ctx, caller, identity.ActionView, a)
		if err != nil {
			return false, fmt.Errorf("authorizing a view of %s: %w", a, err)
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// add counts it in l by its state, and lists it when it is in view.
func (l *AttentionList) add(it *ScopedItem, view AttentionView) {
	state := it.State()
	switch state {
	case attention.StateOpen:
		l.Open++
	case attention.StateAcknowledged:
		l.Acknowledged++
	case attention.StateResolved:
		l.Done++
	}
	if view == 0 || view.holds(state) {
		l.Items = append(l.Items, *it)
	}
}

// scopes reports whether q's account, instrument and instance, those set,
// match it: the account any of its accounts.
func (q *AttentionQuery) scopes(it *ScopedItem) bool {
	return (q.Account == "" || slices.Contains(it.Accounts, q.Account)) &&
		(q.Instrument == "" || q.Instrument == it.Scope.Instrument) &&
		(q.Instance == "" || q.Instance == it.Scope.Instance)
}

// holds reports whether an item in state belongs in view v.
func (v AttentionView) holds(state attention.State) bool {
	switch v {
	case AttentionOpen:
		return state == attention.StateOpen
	case AttentionAcknowledged:
		return state == attention.StateAcknowledged
	case AttentionDone:
		return state == attention.StateResolved
	}
	return false
}
