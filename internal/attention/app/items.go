package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/attention"
)

// Items are the attention item use cases (D-130): the checks raise and
// clear items, the owner acknowledges and resolves them, and every change
// lands in the store with its outbox event.
type Items struct {
	store  Store
	owners Owners
	now    func() time.Time
	checks []Check
}

// New returns the use cases over store, with the owner's sessions checked
// by owners, changes timed by now, and checks swept by Sweep.
func New(store Store, owners Owners, now func() time.Time, checks ...Check) *Items {
	return &Items{store: store, owners: owners, now: now, checks: checks}
}

// List returns every item, in ID order, with its history.
func (s *Items) List(ctx context.Context) ([]attention.Item, error) {
	items, err := s.store.Items(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("loading attention items: %w", err)
	}
	return items, nil
}

// Sweep runs every check and reconciles its kind's items with what it
// found (attention.Reconcile), each kind in its own database transaction.
// A failing check doesn't stop the others; the errors are joined.
func (s *Items) Sweep(ctx context.Context) error {
	var errs []error
	for _, c := range s.checks {
		conds, err := c.Find(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("checking %s: %w", c.Kind, err))
			continue
		}
		if err := s.reconcile(ctx, c.Kind, conds); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Raise records that a check found cond now: it raises the condition's
// item, opens it again or records its new detail (attention.Observe). It
// is the port for a check that reports one condition at a time, such as
// one the jurisdiction pack finds while it estimates a tax.
func (s *Items) Raise(ctx context.Context, cond attention.Condition) error {
	return s.observe(ctx, &cond, true)
}

// Clear records that a check no longer finds the condition of kind and
// subject: its open or acknowledged item is resolved.
func (s *Items) Clear(ctx context.Context, kind attention.Kind, subject string) error {
	return s.observe(ctx, &attention.Condition{Kind: kind, Subject: subject}, false)
}

// Acknowledge records, for the owner, that the owner has seen the open
// item id (attention.Acknowledge), and returns the item. An unknown item
// fails with ErrUnknownItem, a resolved one with
// attention.ErrInvalidTransition.
func (s *Items) Acknowledge(ctx context.Context, origin, sessionToken string, id int64) (attention.Item, error) {
	return s.ownerAction(ctx, origin, sessionToken, id, attention.Acknowledge)
}

// Resolve records, for the owner, that the owner resolved item id
// (attention.Resolve), and returns the item. An unknown item fails with
// ErrUnknownItem; an item of a kind that resolves only when its check
// clears with attention.ErrNotOwnerResolvable.
func (s *Items) Resolve(ctx context.Context, origin, sessionToken string, id int64) (attention.Item, error) {
	return s.ownerAction(ctx, origin, sessionToken, id, attention.Resolve)
}

// ownerAction checks the owner, then applies action to item id in one
// transaction and returns the item as it stands after it.
func (s *Items) ownerAction(ctx context.Context, origin, sessionToken string, id int64,
	action func(*attention.Item, time.Time) (attention.Change, bool, error),
) (attention.Item, error) {
	if err := s.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return attention.Item{}, fmt.Errorf("attention item %d: %w", id, err)
	}
	var out attention.Item
	err := s.store.InTx(ctx, func(tx Store) error {
		if err := tx.Lock(ctx); err != nil {
			return fmt.Errorf("locking: %w", err)
		}
		it, found, err := tx.Item(ctx, id)
		switch {
		case err != nil:
			return fmt.Errorf("loading: %w", err)
		case !found:
			return ErrUnknownItem
		}
		ch, ok, err := action(&it, s.now().UTC())
		if err != nil {
			return err
		}
		if ok {
			step := attention.Step{Kind: it.Kind, Subject: it.Subject, ItemID: it.ID, Change: ch}
			if err := appendAll(ctx, tx, []attention.Step{step}); err != nil {
				return err
			}
			it.Changes = append(it.Changes, ch)
		}
		out = it
		return nil
	})
	if err != nil {
		return attention.Item{}, fmt.Errorf("attention item %d: %w", id, err)
	}
	return out, nil
}

// reconcile brings kind's items in line with conds in one transaction.
func (s *Items) reconcile(ctx context.Context, kind attention.Kind, conds []attention.Condition) error {
	err := s.store.InTx(ctx, func(tx Store) error {
		if err := tx.Lock(ctx); err != nil {
			return fmt.Errorf("locking: %w", err)
		}
		items, err := tx.Items(ctx, kind)
		if err != nil {
			return fmt.Errorf("loading: %w", err)
		}
		steps, err := attention.Reconcile(kind, items, conds, s.now().UTC())
		if err != nil {
			return fmt.Errorf("reconciling: %w", err)
		}
		return appendAll(ctx, tx, steps)
	})
	if err != nil {
		return fmt.Errorf("attention items of %s: %w", kind, err)
	}
	return nil
}

// observe applies one finding to its item in one transaction.
func (s *Items) observe(ctx context.Context, cond *attention.Condition, present bool) error {
	if err := cond.Validate(); err != nil {
		return fmt.Errorf("attention item: %w", err)
	}
	err := s.store.InTx(ctx, func(tx Store) error {
		if err := tx.Lock(ctx); err != nil {
			return fmt.Errorf("locking: %w", err)
		}
		items, err := tx.Items(ctx, cond.Kind)
		if err != nil {
			return fmt.Errorf("loading: %w", err)
		}
		var it *attention.Item
		for i := range items {
			if items[i].Subject == cond.Subject {
				it = &items[i]
			}
		}
		ch, ok := attention.Observe(it, cond, present, s.now().UTC())
		if !ok {
			return nil
		}
		step := attention.Step{Kind: cond.Kind, Subject: cond.Subject, Change: ch}
		if it != nil {
			step.ItemID = it.ID
		}
		return appendAll(ctx, tx, []attention.Step{step})
	})
	if err != nil {
		return fmt.Errorf("attention item %s %q: %w", cond.Kind, cond.Subject, err)
	}
	return nil
}

// appendAll appends steps on tx, in order.
func appendAll(ctx context.Context, tx Store, steps []attention.Step) error {
	for i := range steps {
		if _, err := tx.Append(ctx, &steps[i]); err != nil {
			return fmt.Errorf("storing: %w", err)
		}
	}
	return nil
}
