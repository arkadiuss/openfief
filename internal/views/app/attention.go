package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
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
// those whose scope (attention.ScopeOf) names Account, Instrument (an
// ISIN) or Instance.
type AttentionQuery struct {
	View       AttentionView
	Account    string
	Instrument string
	Instance   string
}

// ScopedItem is an attention item with its scope and group.
type ScopedItem struct {
	attention.Item

	Scope attention.Scope
	Group attention.Group
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
// ID order. An item tied to an account the caller may not view is left out
// (D-163); an item tied to no account names only its connector instance,
// as attention.ScopeOf reads it (D-165).
func (v *Views) Attention(ctx context.Context, q *AttentionQuery) (AttentionList, error) {
	if v.attention == nil {
		return AttentionList{}, nil
	}
	items, err := v.attention.List(ctx)
	if err != nil {
		return AttentionList{}, fmt.Errorf("loading attention items: %w", err)
	}
	var out AttentionList
	for i := range items {
		it := ScopedItem{Item: items[i], Scope: attention.ScopeOf(&items[i]), Group: attention.GroupOf(items[i].Kind)}
		if !q.scopes(&it.Scope) {
			continue
		}
		visible, authzErr := v.mayViewScope(ctx, &it.Scope)
		if authzErr != nil {
			return AttentionList{}, authzErr
		}
		if visible {
			out.add(&it, q.View)
		}
	}
	return out, nil
}

// mayViewScope reports whether the caller in ctx may view an item of scope
// s: one tied to an account only when it may view the account (D-163); any
// other, as every holder of read access may today (D-165).
func (v *Views) mayViewScope(ctx context.Context, s *attention.Scope) (bool, error) {
	if s.Account == "" {
		return true, nil
	}
	ok, err := v.authz.MayAct(ctx, identityapp.CallerFrom(ctx), identity.ActionView, s.Account)
	if err != nil {
		return false, fmt.Errorf("authorizing a view of %s: %w", s.Account, err)
	}
	return ok, nil
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
// match s.
func (q *AttentionQuery) scopes(s *attention.Scope) bool {
	return (q.Account == "" || q.Account == s.Account) &&
		(q.Instrument == "" || q.Instrument == s.Instrument) &&
		(q.Instance == "" || q.Instance == s.Instance)
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
