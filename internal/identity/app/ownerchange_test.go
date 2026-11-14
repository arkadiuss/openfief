package app_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/kernel"
)

var (
	errChangeBroke = errors.New("the store broke")
	errLogDown     = errors.New("the log is down")
)

// TestRecordOwnerChangeKeepsTheEntryWithTheChange checks the seam the other
// contexts' owner changes go through (D-211): the entry is written before the
// change and kept only when the change succeeds, a failure of the change
// leaves no entry, and a log that can't take the entry means the change is not
// called at all.
func TestRecordOwnerChangeKeepsTheEntryWithTheChange(t *testing.T) {
	t.Parallel()
	const session = identity.SessionID("0123456789abcdef")
	refs := []identity.Ref{{Kind: identity.RefAccount, ID: "xtb-main"}}

	t.Run("the change is made and its entry kept", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture(t, demoConfig{})
		var order []string
		err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, refs, func(context.Context) error {
			order = append(order, fmt.Sprintf("change after %d", len(f.store.audit)))
			return nil
		})
		if err != nil {
			t.Fatalf("RecordOwnerChange: %v", err)
		}
		want := []identity.AuditEntry{{
			Seq: 1, At: f.clock.now(), Actor: identity.Actor{Kind: identity.ActorOwner, ID: string(session)},
			Action: identity.ActionAccountDeleted, Outcome: identity.OutcomeOK, Refs: refs,
		}}
		if diff := cmp.Diff(want, f.store.audit); diff != "" {
			t.Errorf("entries (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff([]string{"change after 1"}, order); diff != "" {
			t.Errorf("the change ran (-want +got):\n%s", diff)
		}
	})
	t.Run("a change that fails leaves no entry and its error comes through", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture(t, demoConfig{})
		err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, refs, func(context.Context) error { return errChangeBroke })
		if !errors.Is(err, errChangeBroke) || len(f.store.audit) != 0 {
			t.Errorf("RecordOwnerChange = %v with %d entries, want the change's error and none", err, len(f.store.audit))
		}
	})
	t.Run("a log that is down means no change", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture(t, demoConfig{})
		f.store.auditFail = errLogDown
		called := false
		err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, refs, func(context.Context) error {
			called = true
			return nil
		})
		if !errors.Is(err, kernel.ErrAuditUnavailable) || called {
			t.Errorf("RecordOwnerChange = %v, change called %v; want ErrAuditUnavailable and no change", err, called)
		}
	})
	t.Run("an entry the domain refuses means no change", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture(t, demoConfig{})
		called := false
		bad := []identity.Ref{{Kind: identity.RefAccount, ID: "an account number: 12 3456"}}
		err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, bad, func(context.Context) error {
			called = true
			return nil
		})
		if !errors.Is(err, identity.ErrInvalidAuditEntry) || called {
			t.Errorf("RecordOwnerChange = %v, change called %v; want ErrInvalidAuditEntry and no change", err, called)
		}
	})
}

// plugins is a PluginGenerationStore over the in-memory auth store, with the
// generations rolled back, like the entries, when a transaction fails.
type plugins struct {
	*memAuthStore

	generations map[string]int64
}

func newPlugins() *plugins {
	return &plugins{memAuthStore: newMemAuthStore(), generations: map[string]int64{}}
}

func (p *plugins) Atomically(ctx context.Context, fn func(ctx context.Context) error) error {
	before := maps.Clone(p.generations)
	err := p.memAuthStore.Atomically(ctx, fn)
	if err != nil {
		p.generations = before
	}
	return err
}

func (p *plugins) LockPluginGenerations(context.Context) error { return nil }

func (p *plugins) PluginGenerations(context.Context) (map[string]int64, error) {
	return maps.Clone(p.generations), nil
}

func (p *plugins) SavePluginGeneration(_ context.Context, instance string, generation int64, _ time.Time) error {
	p.generations[instance] = generation
	return nil
}

func (p *plugins) DeletePluginGeneration(_ context.Context, instance string) error {
	delete(p.generations, instance)
	return nil
}

// operatorEntry is the entry the operator's change of instance at generation
// has.
func operatorEntry(at time.Time, action identity.AuditAction, instance, generation string) identity.AuditEntry {
	return identity.AuditEntry{
		At: at, Actor: identity.Actor{Kind: identity.ActorOperator, ID: identity.OperatorActorID}, Action: action, Outcome: identity.OutcomeOK,
		Refs: []identity.Ref{{Kind: identity.RefPlugin, ID: instance}, {Kind: identity.RefPluginGeneration, ID: generation}},
	}
}

// observedStep is one observation of the plugins and how many entries it
// writes.
type observedStep struct {
	name     string
	observed []app.ObservedPlugin
	want     int
}

// TestPluginObservationsEnterWhatTheOperatorDid checks that the core enters
// an install, an upgrade and a removal of a plugin as the operator's, once per
// instance and generation: seeing the same again, a resource without a
// generation and one being deleted enter nothing.
func TestPluginObservationsEnterWhatTheOperatorDid(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.November, 14, 9, 0, 0, 0, time.UTC)
	store := newPlugins()
	o := app.NewPluginObservations(store, func() time.Time { return at })
	xtb := func(generation int64) app.ObservedPlugin {
		return app.ObservedPlugin{Instance: "xtb-main", Generation: generation}
	}
	ghost := app.ObservedPlugin{Instance: "ghostfolio", Generation: 4}
	steps := []observedStep{
		{"two installs", []app.ObservedPlugin{xtb(1), ghost}, 2},
		{"the same again", []app.ObservedPlugin{xtb(1), ghost}, 0},
		{"an upgrade", []app.ObservedPlugin{xtb(2), ghost}, 1},
		{"one being deleted and one with no generation", []app.ObservedPlugin{
			{Instance: "xtb-main", Generation: 3, Deleting: true}, ghost, {Instance: "new"},
		}, 0},
		{"a removal", []app.ObservedPlugin{ghost}, 1},
	}
	for _, s := range steps {
		n, err := o.Observe(t.Context(), s.observed)
		if err != nil || n != s.want {
			t.Fatalf("%s: Observe = %d, %v; want %d", s.name, n, err, s.want)
		}
	}
	want := []identity.AuditEntry{
		operatorEntry(at, identity.ActionPluginInstalled, "ghostfolio", "4"),
		operatorEntry(at, identity.ActionPluginInstalled, "xtb-main", "1"),
		operatorEntry(at, identity.ActionPluginUpgraded, "xtb-main", "2"),
		operatorEntry(at, identity.ActionPluginRemoved, "xtb-main", "2"),
	}
	got := make([]identity.AuditEntry, len(store.audit))
	for i, e := range store.audit {
		e.Seq = 0
		got[i] = e
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("entries (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(map[string]int64{"ghostfolio": 4}, store.generations); diff != "" {
		t.Errorf("generations kept (-want +got):\n%s", diff)
	}
}

// TestPluginObservationsKeepNothingWhenTheLogIsDown checks that an entry that
// can't be written leaves the generations as they were, so the next
// observation, with the log back, enters it.
func TestPluginObservationsKeepNothingWhenTheLogIsDown(t *testing.T) {
	t.Parallel()
	store := newPlugins()
	o := app.NewPluginObservations(store, func() time.Time { return time.Date(2026, time.November, 14, 9, 0, 0, 0, time.UTC) })
	ghost := func(generation int64) []app.ObservedPlugin {
		return []app.ObservedPlugin{{Instance: "ghostfolio", Generation: generation}}
	}
	if _, err := o.Observe(t.Context(), ghost(4)); err != nil {
		t.Fatalf("Observe: %v", err)
	}
	store.auditFail = errLogDown
	if _, err := o.Observe(t.Context(), ghost(5)); err == nil {
		t.Fatal("Observe with the log down = nil, want an error")
	}
	if store.generations["ghostfolio"] != 4 || len(store.audit) != 1 {
		t.Errorf("a failed observation left generations %v and %d entries, want generation 4 and the first entry", store.generations, len(store.audit))
	}
	store.auditFail = nil
	if n, err := o.Observe(t.Context(), ghost(5)); err != nil || n != 1 || store.generations["ghostfolio"] != 5 {
		t.Errorf("the retry = %d, %v with generations %v; want one entry and generation 5", n, err, store.generations)
	}
}
