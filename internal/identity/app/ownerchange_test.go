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

// fakeOutbox is a ChangeOutbox over a list of pending entries: it runs the
// change, and keeps the entry only when the change succeeds, as the real one
// does in the one transaction; with enqueueFails set it cannot enqueue, so the
// change is not made.
type fakeOutbox struct {
	pending      []identity.AuditEntry
	enqueueFails error
}

func (o *fakeOutbox) Within(ctx context.Context, e *identity.AuditEntry, change func(context.Context) error) error {
	if o.enqueueFails != nil {
		return fmt.Errorf("%w: %w", kernel.ErrAuditUnavailable, o.enqueueFails)
	}
	if err := change(ctx); err != nil {
		return err
	}
	o.pending = append(o.pending, *e)
	return nil
}

// TestRecordOwnerChangeEntersTheEntryInTheOutbox checks the seam the other
// contexts' owner changes go through (D-211): the entry, with the time of the
// change and the owner's session, goes to the outbox together with the change,
// a failure of the change leaves no entry, and an outbox that can't take the
// entry means the change is not made.
func TestRecordOwnerChangeEntersTheEntryInTheOutbox(t *testing.T) {
	t.Parallel()
	const session = identity.SessionID("0123456789abcdef")
	refs := []identity.Ref{{Kind: identity.RefAccount, ID: "xtb-main"}}
	out := &fakeOutbox{}
	f := newAuthFixture(t, demoConfig{}, app.WithChangeOutbox(out))
	called := false
	err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, refs, func(context.Context) error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("RecordOwnerChange = %v, change called %v", err, called)
	}
	want := []identity.AuditEntry{{
		At: f.clock.now(), Actor: identity.Actor{Kind: identity.ActorOwner, ID: string(session)},
		Action: identity.ActionAccountDeleted, Outcome: identity.OutcomeOK, Refs: refs,
	}}
	if diff := cmp.Diff(want, out.pending); diff != "" {
		t.Errorf("pending entries (-want +got):\n%s", diff)
	}
	if len(f.store.audit) != 0 {
		t.Errorf("the log holds %d entries before the worker ran, want none: the outbox holds them", len(f.store.audit))
	}
}

// TestRecordOwnerChangeFailsClosed checks the cases in which no entry is
// enqueued: a change that fails leaves none and its error comes through, and
// an outbox that can't take the entry, none at all, or an entry the domain
// refuses, means the change is not made.
func TestRecordOwnerChangeFailsClosed(t *testing.T) {
	t.Parallel()
	const session = identity.SessionID("0123456789abcdef")
	good := []identity.Ref{{Kind: identity.RefAccount, ID: "xtb-main"}}
	bad := []identity.Ref{{Kind: identity.RefAccount, ID: "an account number: 12 3456"}}
	tests := []struct {
		name       string
		out        *fakeOutbox // nil: no outbox
		refs       []identity.Ref
		changeErr  error
		wantErr    error
		wantCalled bool
	}{
		{name: "a change that fails", out: &fakeOutbox{}, refs: good, changeErr: errChangeBroke, wantErr: errChangeBroke, wantCalled: true},
		{name: "an enqueue that fails", out: &fakeOutbox{enqueueFails: errLogDown}, refs: good, wantErr: kernel.ErrAuditUnavailable},
		{name: "no outbox", refs: good, wantErr: kernel.ErrAuditUnavailable},
		{name: "an entry the domain refuses", out: &fakeOutbox{}, refs: bad, wantErr: identity.ErrInvalidAuditEntry},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var opts []app.AuthOption
			if tt.out != nil {
				opts = append(opts, app.WithChangeOutbox(tt.out))
			}
			f := newAuthFixture(t, demoConfig{}, opts...)
			called := false
			err := f.auth.RecordOwnerChange(t.Context(), session, identity.ActionAccountDeleted, tt.refs, func(context.Context) error {
				called = true
				return tt.changeErr
			})
			if !errors.Is(err, tt.wantErr) || called != tt.wantCalled {
				t.Errorf("RecordOwnerChange = %v, change called %v; want %v and called %v", err, called, tt.wantErr, tt.wantCalled)
			}
			if tt.out != nil && len(tt.out.pending) != 0 {
				t.Errorf("%d entries enqueued, want none", len(tt.out.pending))
			}
		})
	}
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
