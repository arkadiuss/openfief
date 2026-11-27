package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// DefaultQueueWait is how long a call of the Ask box waits for a free one of
// its calls at once before the router refuses it as busy.
const DefaultQueueWait = 5 * time.Second

// Guards are D-208's limits on the router's calls to models, each off in the
// zero value: an installation that sets none is limited by nothing but its
// routes, and the demo sets all of them.
type Guards struct {
	// Budget caps what the day's calls cost, counted from the model entries'
	// prices. Usage keeps the day's total, so it needs one whenever the budget
	// is enabled.
	Budget modelrouter.Budget
	Usage  UsageStore
	// Zone is the time zone a day is counted in: the core's own, so the budget
	// starts over at the same midnight "today" does. UTC when nil.
	Zone *time.Location
	// AskCalls is how many calls of the Ask box's role may be in flight at
	// once, in this process; zero is no limit. A call past it waits up to
	// QueueWait for one to finish (DefaultQueueWait when zero) and is then
	// refused with ErrBusy. Each replica of the core has its own.
	AskCalls  int
	QueueWait time.Duration
}

// BudgetState reports whether the day's budget is spent at the router's
// clock, and, when it is, when it starts over. It reads the day's total and
// writes nothing, so the Ask box can say the budget is used up before it
// counts a question. An error wraps ErrUsageUnavailable. Without a budget it
// is never spent.
func (r *Router) BudgetState(ctx context.Context) (BudgetStatus, error) {
	if !r.guards.Budget.Enabled() {
		return BudgetStatus{}, nil
	}
	day, ends := modelrouter.BudgetDay(r.now(), r.guards.Zone)
	total, err := r.guards.Usage.Spent(ctx, day, r.guards.Budget.Currency)
	if err != nil {
		return BudgetStatus{}, fmt.Errorf("%w: reading the day's total: %w", ErrUsageUnavailable, err)
	}
	return BudgetStatus{Spent: r.guards.Budget.Exceeded(total), ResetsAt: ends}, nil
}

// BudgetStatus is whether the day's budget is spent and, when a budget is set,
// when the day ends and it starts over.
type BudgetStatus struct {
	Spent    bool
	ResetsAt time.Time
}

// checkBudget refuses a call when the day's budget is spent, or when one of
// the candidates can't be counted against it, which would let that call spend
// without being counted: it is priced in another currency than the budget's
// (modelrouter.ErrUnpriceable), the catalog has no entry for it, which is
// every model after a restart until its provider lists them, or a third
// party runs it and its entry has no price (modelrouter.ErrNoPrice). Only a
// model on the owner's hardware is free without a price.
func (r *Router) checkBudget(ctx context.Context, candidates []modelrouter.Ref) error {
	if !r.guards.Budget.Enabled() {
		return nil
	}
	state, err := r.BudgetState(ctx)
	if err != nil {
		return err
	}
	if state.Spent {
		return &modelrouter.BudgetSpentError{ResetsAt: state.ResetsAt}
	}
	for _, m := range candidates {
		e, ok := r.entryOf(ctx, m)
		if !ok {
			return fmt.Errorf("model %s: %w: the catalog has no entry for it", m, modelrouter.ErrNoPrice)
		}
		if _, priceErr := modelrouter.CallCost(&e, modelrouter.Usage{}, r.guards.Budget.Currency); priceErr != nil {
			return fmt.Errorf("model %s: %w", m, priceErr)
		}
		if !e.HasPrice {
			if err := r.checkFree(ctx, m); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFree refuses the unpriced model m unless it runs on the owner's
// hardware: a model a third party runs costs money whatever its entry says,
// and where m runs, when its provider can't say, is not known to be free.
func (r *Router) checkFree(ctx context.Context, m modelrouter.Ref) error {
	inst, err := r.providers.Describe(ctx, m.Instance)
	if err != nil {
		return fmt.Errorf("model %s: %w: where it runs is unknown: %w", m, modelrouter.ErrNoPrice, err)
	}
	if inst.Location != modelrouter.LocationOwnerHardware {
		return fmt.Errorf("model %s: %w: its entry has no price and it doesn't run on the owner's hardware", m, modelrouter.ErrNoPrice)
	}
	return nil
}

// recordUsage adds what the answered call cost to its day. Without a budget
// it records nothing. A call that can't be priced, because its entry is gone
// or priced in another currency than the cap's, is not recorded at no cost:
// checkBudget refused both before the call, so one that reaches here lost its
// entry while it ran, and the error withholds the answer like any usage that
// can't be recorded.
func (r *Router) recordUsage(ctx context.Context, m modelrouter.Ref, res *modelrouter.Result) error {
	if !r.guards.Budget.Enabled() {
		return nil
	}
	entry, ok := r.entryOf(ctx, m)
	if !ok {
		return fmt.Errorf("%w: pricing the call: %w", ErrUsageUnavailable, modelrouter.ErrNoPrice)
	}
	price, err := modelrouter.CallCost(&entry, res.Usage, r.guards.Budget.Currency)
	if err != nil {
		return fmt.Errorf("%w: pricing the call: %w", ErrUsageUnavailable, err)
	}
	day, _ := modelrouter.BudgetDay(r.now(), r.guards.Zone)
	rec := UsageRecord{Day: day, Model: m, Usage: res.Usage, Cost: price, Currency: r.guards.Budget.Currency}
	// Recorded even when the caller gave up: the call was made and paid for.
	if err := r.guards.Usage.Record(context.WithoutCancel(ctx), &rec); err != nil {
		return fmt.Errorf("%w: recording the call: %w", ErrUsageUnavailable, err)
	}
	return nil
}

// entryOf returns the model entry m names, from the cache or its provider.
func (r *Router) entryOf(ctx context.Context, m modelrouter.Ref) (modelrouter.ModelEntry, bool) {
	e, ok := r.entriesOf(ctx, m.Instance)[m.Model]
	return e, ok
}

// acquire takes one of the Ask box's calls at once for a call of role, waiting
// up to the queue wait for one, and returns what frees it. Another role, or no
// limit, takes nothing. A call that waited in vain gets ErrBusy; one whose
// caller gave up first gets that context's error.
func (r *Router) acquire(ctx context.Context, role modelrouter.Role) (func(), error) {
	if role != modelrouter.RoleAsk || r.gate == nil {
		return func() {}, nil
	}
	free := func() { <-r.gate }
	select {
	case r.gate <- struct{}{}:
		return free, nil
	default:
	}
	wait := r.guards.QueueWait
	if wait <= 0 {
		wait = DefaultQueueWait
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case r.gate <- struct{}{}:
		return free, nil
	case <-timer.C:
		return nil, ErrBusy
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for a free model call: %w", ctx.Err())
	}
}
