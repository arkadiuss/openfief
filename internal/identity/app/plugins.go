package app

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// PluginGenerationStore keeps the generation of each plugin instance the core
// last entered in the audit log, so that seeing the same generation again
// writes nothing. *identity/postgres.Store implements it, with Auditor and
// Atomic.
type PluginGenerationStore interface {
	Auditor
	Atomic
	// LockPluginGenerations makes every other observer wait until the
	// transaction ctx carries ends, so that two cores observing the same
	// plugins don't both enter them.
	LockPluginGenerations(ctx context.Context) error
	// PluginGenerations returns the generation entered for each instance.
	PluginGenerations(ctx context.Context) (map[string]int64, error)
	// SavePluginGeneration records that generation of instance was entered at
	// at, replacing the one before.
	SavePluginGeneration(ctx context.Context, instance string, generation int64, at time.Time) error
	// DeletePluginGeneration forgets instance, which was removed.
	DeletePluginGeneration(ctx context.Context, instance string) error
}

// ObservedPlugin is a Plugin resource as the core sees it: the instance it
// names, the generation of its spec, and whether it is being deleted. The
// operator applies a new spec as a new generation (D-211).
type ObservedPlugin struct {
	Instance   string
	Generation int64
	// Deleting reports that the resource is being deleted: the operator
	// still holds it, and its generation may change with the deletion, so it
	// is entered as removed once it is gone and not before.
	Deleting bool
}

// PluginObservations enter in the audit log what the operator did to the
// plugins, as the core observes it (D-211): an instance never seen before is
// installed, one whose generation changed is upgraded, and one that is gone
// is removed, each by the actor "operator", naming the instance and the
// generation. The operator is a separate process the core can't make wait for
// an entry, so this records after the fact, once per instance and generation:
// the generation recorded with the entry is what makes it idempotent, and the
// caller repeats it until it succeeds.
type PluginObservations struct {
	store PluginGenerationStore
	now   func() time.Time
}

// NewPluginObservations returns the observation of plugins over store, reading
// the time from now.
func NewPluginObservations(store PluginGenerationStore, now func() time.Time) *PluginObservations {
	return &PluginObservations{store: store, now: now}
}

// Observe enters, in one transaction, an entry for each instance of observed
// that is new or changed, and one for each instance entered before that
// observed lacks, and records the generations, so the same observation made
// again changes nothing. A resource without a generation (zero) can't be
// keyed and is left out. It returns how many entries it wrote; on an error
// nothing was written and the caller tries again.
func (o *PluginObservations) Observe(ctx context.Context, observed []ObservedPlugin) (int, error) {
	seen := slices.Clone(observed)
	slices.SortFunc(seen, func(a, b ObservedPlugin) int { return cmp.Compare(a.Instance, b.Instance) })
	var written int
	err := within(ctx, o.store, func(ctx context.Context) error {
		if err := o.store.LockPluginGenerations(ctx); err != nil {
			return fmt.Errorf("locking the plugin generations: %w", err)
		}
		known, err := o.store.PluginGenerations(ctx)
		if err != nil {
			return fmt.Errorf("loading the plugin generations: %w", err)
		}
		var observeErr error
		written, observeErr = o.observe(ctx, known, seen, o.now().UTC())
		return observeErr
	})
	if err != nil {
		return 0, fmt.Errorf("observing the plugins: %w", err)
	}
	return written, nil
}

// observe enters what seen, sorted by instance, and the instances of known
// that seen lacks, call for at at, and returns how many entries it wrote.
func (o *PluginObservations) observe(ctx context.Context, known map[string]int64, seen []ObservedPlugin, at time.Time) (int, error) {
	written := 0
	for _, p := range seen {
		wrote, err := o.enter(ctx, known, p, at)
		if err != nil {
			return 0, err
		}
		written += wrote
	}
	for _, name := range slices.Sorted(maps.Keys(known)) {
		if slices.ContainsFunc(seen, func(p ObservedPlugin) bool { return p.Instance == name }) {
			continue
		}
		if err := o.remove(ctx, name, known[name], at); err != nil {
			return 0, err
		}
		written++
	}
	return written, nil
}

// enter writes the entry p's state calls for, if any, and returns how many it
// wrote: none for a resource being deleted, one without a generation, or one
// whose generation was entered already.
func (o *PluginObservations) enter(ctx context.Context, known map[string]int64, p ObservedPlugin, at time.Time) (int, error) {
	if p.Deleting || p.Generation <= 0 {
		return 0, nil
	}
	before, entered := known[p.Instance]
	if entered && before == p.Generation {
		return 0, nil
	}
	action := identity.ActionPluginInstalled
	if entered {
		action = identity.ActionPluginUpgraded
	}
	if err := o.record(ctx, action, p.Instance, p.Generation, at); err != nil {
		return 0, err
	}
	if err := o.store.SavePluginGeneration(ctx, p.Instance, p.Generation, at); err != nil {
		return 0, fmt.Errorf("recording the generation of %s: %w", p.Instance, err)
	}
	return 1, nil
}

// remove writes the entry of the removal of instance, whose last generation
// entered was generation, and forgets it.
func (o *PluginObservations) remove(ctx context.Context, instance string, generation int64, at time.Time) error {
	if err := o.record(ctx, identity.ActionPluginRemoved, instance, generation, at); err != nil {
		return err
	}
	if err := o.store.DeletePluginGeneration(ctx, instance); err != nil {
		return fmt.Errorf("forgetting %s: %w", instance, err)
	}
	return nil
}

// record appends the entry of action by the operator, naming the instance and
// its generation.
func (o *PluginObservations) record(ctx context.Context, action identity.AuditAction, instance string, generation int64, at time.Time) error {
	e, err := identity.NewAuditEntry(at, identity.Actor{Kind: identity.ActorOperator, ID: identity.OperatorActorID}, action, identity.OutcomeOK,
		identity.Ref{Kind: identity.RefPlugin, ID: instance},
		identity.Ref{Kind: identity.RefPluginGeneration, ID: strconv.FormatInt(generation, 10)})
	if err != nil {
		return fmt.Errorf("building the audit entry of %s: %w", instance, err)
	}
	if err := o.store.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("appending the audit entry of %s: %w", instance, err)
	}
	return nil
}
