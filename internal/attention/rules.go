package attention

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

// Step is a change to append to the item of a kind and subject: ItemID is
// the stored item's, zero when the change raises a new one.
type Step struct {
	Kind    Kind
	Subject string
	ItemID  int64
	Change  Change
}

// Observe returns the change a check's finding makes to it, the item of
// the condition's kind and subject, nil or with no history when there is
// none yet; ok is false when nothing changes. present reports whether the
// check found the condition now, with cond's detail; absent, only cond's
// kind and subject count.
//
//   - Found with no item, or with a resolved item whose condition had
//     cleared: the item is raised or reopened.
//   - Found with an open or acknowledged item and another detail: it is
//     open again with the new detail, so the owner sees what changed.
//   - Found with an item the owner resolved: it stays resolved while the
//     detail is the same, and reopens when it changes.
//   - Not found with an open or acknowledged item: it is resolved as
//     cleared. Not found otherwise: nothing changes.
func Observe(it *Item, cond *Condition, present bool, at time.Time) (Change, bool) {
	if it == nil {
		it = &Item{}
	}
	cur := it.Current()
	switch cur.State {
	case StateOpen, StateAcknowledged:
		if !present {
			return it.next(StateResolved, ReasonCleared, cur.Detail, at), true
		}
		if cond.Detail != cur.Detail {
			return it.next(StateOpen, ReasonChanged, cond.Detail, at), true
		}
		return Change{}, false
	case StateResolved:
		if !present || (cur.Reason == ReasonResolved && cond.Detail == cur.Detail) {
			return Change{}, false
		}
		return it.next(StateOpen, ReasonReopened, cond.Detail, at), true
	}
	if !present {
		return Change{}, false
	}
	return it.next(StateOpen, ReasonRaised, cond.Detail, at), true
}

// Reconcile returns the steps that bring items, every stored item of kind,
// in line with conds, everything a check of kind finds now (Observe): one
// item per subject, new conditions raised, cleared ones resolved. Steps
// come in subject order. A condition of another kind, an invalid one, or
// two with one subject fail with an error wrapping ErrInvalidCondition.
func Reconcile(kind Kind, items []Item, conds []Condition, at time.Time) ([]Step, error) {
	found, err := bySubject(kind, conds)
	if err != nil {
		return nil, err
	}
	stored := make(map[string]*Item, len(items))
	subjects := make([]string, 0, len(found)+len(items))
	for s := range found {
		subjects = append(subjects, s)
	}
	for i := range items {
		stored[items[i].Subject] = &items[i]
		if found[items[i].Subject] == nil {
			subjects = append(subjects, items[i].Subject)
		}
	}
	slices.SortFunc(subjects, cmp.Compare)
	var out []Step
	for _, s := range subjects {
		cond, present := found[s]
		if !present {
			cond = &Condition{Kind: kind, Subject: s}
		}
		it := stored[s]
		if ch, ok := Observe(it, cond, present, at); ok {
			step := Step{Kind: kind, Subject: s, Change: ch}
			if it != nil {
				step.ItemID = it.ID
			}
			out = append(out, step)
		}
	}
	return out, nil
}

// bySubject returns conds by subject, each valid and of kind, with no
// subject twice; otherwise an error wrapping ErrInvalidCondition.
func bySubject(kind Kind, conds []Condition) (map[string]*Condition, error) {
	if err := kind.Validate(); err != nil {
		return nil, err
	}
	found := make(map[string]*Condition, len(conds))
	for i := range conds {
		c := &conds[i]
		if err := c.Validate(); err != nil {
			return nil, err
		}
		if c.Kind != kind || found[c.Subject] != nil {
			return nil, fmt.Errorf("%w: %s %q: another kind or a repeated subject", ErrInvalidCondition, c.Kind, c.Subject)
		}
		found[c.Subject] = c
	}
	return found, nil
}

// Acknowledge returns the change that records the owner's acknowledgement
// of it: an open item becomes acknowledged, with its detail. ok is false
// for an item already acknowledged; a resolved item, or one with no
// history, fails with an error wrapping ErrInvalidTransition.
func Acknowledge(it *Item, at time.Time) (Change, bool, error) {
	cur := it.Current()
	switch cur.State {
	case StateOpen:
		return it.next(StateAcknowledged, ReasonAcknowledged, cur.Detail, at), true, nil
	case StateAcknowledged:
		return Change{}, false, nil
	case StateResolved:
	}
	return Change{}, false, fmt.Errorf("%w: acknowledging a %s item", ErrInvalidTransition, cur.State)
}

// Resolve returns the change that records the owner resolving it, an open
// or acknowledged item of a kind the owner may resolve (OwnerResolves). ok
// is false for an item already resolved. Another kind fails with
// ErrNotOwnerResolvable, an item with no history with ErrInvalidTransition.
func Resolve(it *Item, at time.Time) (Change, bool, error) {
	if !OwnerResolves(it.Kind) {
		return Change{}, false, fmt.Errorf("%w: %s", ErrNotOwnerResolvable, it.Kind)
	}
	cur := it.Current()
	switch cur.State {
	case StateOpen, StateAcknowledged:
		return it.next(StateResolved, ReasonResolved, cur.Detail, at), true, nil
	case StateResolved:
		return Change{}, false, nil
	}
	return Change{}, false, fmt.Errorf("%w: resolving an item with no history", ErrInvalidTransition)
}
