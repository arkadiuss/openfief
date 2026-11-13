package spicedb_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
)

// The IDs of the grants' holders.
const (
	tokenID  = "0011223344556677"
	personID = "8899aabbccddeeff"
)

// grants are the owner's grants (D-209) in force, as the identity context's
// table holds them: the Syncer writes each as a relationship.
func grants() []identity.Grant {
	at := time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC)
	return []identity.Grant{
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: tokenID}, Target: identity.Household(), CreatedAt: at},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: personID}, Target: identity.Household(), CreatedAt: at},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: personID}, Target: identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "bonds"}, CreatedAt: at},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ui"}, Target: identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "core"}, CreatedAt: at},
	}
}

// grantTuples are the relationships of grants().
func grantTuples() []spicedb.Tuple {
	return []spicedb.Tuple{
		{Resource: "household:home", Relation: "reader", Subject: "script_token:" + tokenID},
		{Resource: "household:home", Relation: "reader", Subject: "person:" + personID},
		{Resource: "account:bonds", Relation: "viewer", Subject: "person:" + personID},
		{Resource: "sleeve:core", Relation: "viewer", Subject: "view_plugin:ui"},
	}
}

// fixed is a loader of d.
func fixed(d spicedb.Desired) func(context.Context) (spicedb.Desired, error) {
	return func(context.Context) (spicedb.Desired, error) { return d, nil }
}

func TestPlan(t *testing.T) {
	t.Parallel()
	a := spicedb.Tuple{Resource: "account:a", Relation: "household", Subject: "household:home"}
	b := spicedb.Tuple{Resource: "account:a", Relation: "sleeve", Subject: "sleeve:core"}
	c := spicedb.Tuple{Resource: "account:a", Relation: "sleeve", Subject: "sleeve:play"}
	tests := []struct {
		name             string
		have, want       []spicedb.Tuple
		wantTouch, want2 []spicedb.Tuple
	}{
		{name: "nothing to do", have: []spicedb.Tuple{a, b}, want: []spicedb.Tuple{b, a}},
		{name: "everything is missing", want: []spicedb.Tuple{a, b}, wantTouch: []spicedb.Tuple{a, b}},
		{name: "a stale one goes", have: []spicedb.Tuple{a, b}, want: []spicedb.Tuple{a}, want2: []spicedb.Tuple{b}},
		{name: "a move touches and deletes", have: []spicedb.Tuple{a, b}, want: []spicedb.Tuple{a, c}, wantTouch: []spicedb.Tuple{c}, want2: []spicedb.Tuple{b}},
		{name: "duplicates count once", have: []spicedb.Tuple{b, b}, want: []spicedb.Tuple{a, a}, wantTouch: []spicedb.Tuple{a}, want2: []spicedb.Tuple{b}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := spicedb.Plan(tt.have, tt.want)
			if diff := cmp.Diff(tt.wantTouch, got.Touch, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("touch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want2, got.Delete, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("delete (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSyncerPlacesAnAccount checks that a placement writes the account's
// household and sleeve, moves it in one atomic write, and leaves grants.
func TestSyncerPlacesAnAccount(t *testing.T) {
	t.Parallel()
	f := &fake{tuples: grantTuples()}
	s := spicedb.NewSyncer(f)
	ctx := t.Context()

	if err := s.AccountPlaced(ctx, "brokerage", "core"); err != nil {
		t.Fatalf("AccountPlaced: %v", err)
	}
	if err := s.AccountPlaced(ctx, "brokerage", "core"); err != nil {
		t.Fatalf("AccountPlaced again: %v", err)
	}
	writes := len(f.writes)
	if err := s.AccountPlaced(ctx, "brokerage", "play"); err != nil {
		t.Fatalf("AccountPlaced move: %v", err)
	}
	if got := len(f.writes) - writes; got != 1 {
		t.Errorf("a move took %d writes, want 1 atomic write", got)
	}
	for _, want := range []spicedb.Tuple{
		{Resource: "account:brokerage", Relation: "household", Subject: "household:home"},
		{Resource: "account:brokerage", Relation: "sleeve", Subject: "sleeve:play"},
	} {
		if !f.has(want) {
			t.Errorf("missing %s; held %v", want, f.held())
		}
	}
	if f.has(spicedb.Tuple{Resource: "account:brokerage", Relation: "sleeve", Subject: "sleeve:core"}) {
		t.Error("the old sleeve's relationship is still held")
	}
	for _, g := range grantTuples() {
		if !f.has(g) {
			t.Errorf("grant %s was removed", g)
		}
	}
}

func TestSyncerUnplacesAnAccount(t *testing.T) {
	t.Parallel()
	f := &fake{tuples: grantTuples()}
	s := spicedb.NewSyncer(f)
	ctx := t.Context()
	if err := s.AccountPlaced(ctx, "brokerage", "core"); err != nil {
		t.Fatalf("AccountPlaced: %v", err)
	}
	for range 2 { // idempotent
		if err := s.AccountUnplaced(ctx, "brokerage"); err != nil {
			t.Fatalf("AccountUnplaced: %v", err)
		}
	}
	if f.has(spicedb.Tuple{Resource: "account:brokerage", Relation: "sleeve", Subject: "sleeve:core"}) {
		t.Error("the sleeve's relationship is still held")
	}
	// The owner keeps the account the ledger still holds.
	if !f.has(spicedb.Tuple{Resource: "account:brokerage", Relation: "household", Subject: "household:home"}) {
		t.Error("the account lost its household")
	}
	for _, g := range grantTuples() {
		if !f.has(g) {
			t.Errorf("grant %s was removed", g)
		}
	}
}

// TestSyncerReconciles checks the repair at start: missing relationships
// are written, those of a deleted account or a sleeve left are removed,
// demo sessions' grant follows the flag, the grants in force are written
// and a relationship of a grant that is not, or was never, in force is
// removed, and a second run writes nothing.
func TestSyncerReconciles(t *testing.T) {
	t.Parallel()
	stale := []spicedb.Tuple{
		{Resource: "account:gone", Relation: "household", Subject: "household:home"},
		{Resource: "account:gone", Relation: "sleeve", Subject: "sleeve:core"},
		{Resource: "account:bonds", Relation: "sleeve", Subject: "sleeve:old"},
		{Resource: "household:home", Relation: "owner", Subject: "person:previous"},
		{Resource: "household:home", Relation: "reader", Subject: "demo_session:*"},
		// A grant that ended, one written by hand and one of a person's viewer relation: none is in the table.
		{Resource: "account:bonds", Relation: "viewer", Subject: "script_token:ended"},
		{Resource: "household:home", Relation: "reader", Subject: "view_plugin:by-hand"},
		{Resource: "sleeve:core", Relation: "viewer", Subject: "person:ended"},
	}
	// SpiceDB holds all the grants but one, which the table has: drift in both directions.
	f := &fake{tuples: append(grantTuples()[:3:3], stale...)}
	s := spicedb.NewSyncer(f)
	ctx := t.Context()
	d := spicedb.Desired{
		Accounts: []string{"bonds", "brokerage", "IKE (PL)"},
		Sleeves:  map[string]string{"bonds": "bonds", "brokerage": "core"},
		Grants:   grants(),
	}
	if err := s.Reconcile(ctx, fixed(d)); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	want := []string{
		"account:IKE=20=28PL=29#household@household:home",
		"account:bonds#household@household:home",
		"account:bonds#sleeve@sleeve:bonds",
		"account:bonds#viewer@person:" + personID,
		"account:brokerage#household@household:home",
		"account:brokerage#sleeve@sleeve:core",
		"household:home#owner@person:owner",
		"household:home#reader@person:" + personID,
		"household:home#reader@script_token:" + tokenID,
		"sleeve:core#viewer@view_plugin:ui",
	}
	if diff := cmp.Diff(want, f.held()); diff != "" {
		t.Errorf("held relationships (-want +got):\n%s", diff)
	}

	writes := len(f.writes)
	if err := s.Reconcile(ctx, fixed(d)); err != nil {
		t.Fatalf("Reconcile again: %v", err)
	}
	if len(f.writes) != writes {
		t.Errorf("a second Reconcile wrote %d times, want none", len(f.writes)-writes)
	}

	d.DemoSessions = true
	if err := s.Reconcile(ctx, fixed(d)); err != nil {
		t.Fatalf("Reconcile with demo sessions: %v", err)
	}
	demo := spicedb.Tuple{Resource: "household:home", Relation: "reader", Subject: "demo_session:*"}
	if !f.has(demo) {
		t.Error("demo sessions have no household grant")
	}
	d.DemoSessions = false
	if err := s.Reconcile(ctx, fixed(d)); err != nil {
		t.Fatalf("Reconcile without demo sessions: %v", err)
	}
	if f.has(demo) {
		t.Error("demo sessions keep their household grant")
	}
}

// TestSyncerProjectsGrants checks that a grant is the relationship of its
// target and holder (D-209): a household grant makes the holder a reader of
// the household, an account's or a sleeve's a viewer of it; that an ID
// SpiceDB can't take is escaped as every other; and that projecting and
// unprojecting are idempotent and touch nothing else.
func TestSyncerProjectsGrants(t *testing.T) {
	t.Parallel()
	f := &fake{tuples: []spicedb.Tuple{{Resource: "household:home", Relation: "owner", Subject: "person:owner"}}}
	s := spicedb.NewSyncer(f)
	ctx := t.Context()
	at := time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC)
	odd := identity.Grant{
		Holder:    identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ui"},
		Target:    identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "IKE (PL)"},
		CreatedAt: at,
	}
	all := append(grants(), odd)
	for range 2 {
		if err := s.ProjectGrants(ctx, all); err != nil {
			t.Fatalf("ProjectGrants: %v", err)
		}
	}
	want := append([]string{"household:home#owner@person:owner", "account:IKE=20=28PL=29#viewer@view_plugin:ui"}, tupleLines(grantTuples())...)
	sortStrings(want)
	if diff := cmp.Diff(want, f.held()); diff != "" {
		t.Errorf("held after ProjectGrants (-want +got):\n%s", diff)
	}
	for range 2 {
		if err := s.UnprojectGrants(ctx, all); err != nil {
			t.Fatalf("UnprojectGrants: %v", err)
		}
	}
	if diff := cmp.Diff([]string{"household:home#owner@person:owner"}, f.held()); diff != "" {
		t.Errorf("held after UnprojectGrants (-want +got):\n%s", diff)
	}
	for _, bad := range []identity.Grant{
		{Holder: identity.GrantHolder{Kind: "agent", ID: "x"}, Target: identity.Household()},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: tokenID}, Target: identity.GrantTarget{Kind: "fief", ID: "x"}},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: tokenID}, Target: identity.GrantTarget{Kind: identity.GrantTargetAccount}},
	} {
		if err := s.ProjectGrants(ctx, []identity.Grant{bad}); err == nil {
			t.Errorf("ProjectGrants(%+v) succeeded", bad)
		}
	}
	if len(f.writes) != 4 {
		t.Errorf("%d writes, want only the four of the two good calls each: a bad grant writes nothing", len(f.writes))
	}
}

// tupleLines returns ts in zed's notation.
func tupleLines(ts []spicedb.Tuple) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, t.String())
	}
	return out
}

func sortStrings(s []string) { slices.Sort(s) }

func TestSyncerWritesInBatches(t *testing.T) {
	t.Parallel()
	f := &fake{}
	d := spicedb.Desired{}
	for i := range 700 {
		d.Accounts = append(d.Accounts, fmt.Sprintf("acc-%d", i))
	}
	if err := spicedb.NewSyncer(f).Reconcile(t.Context(), fixed(d)); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(f.writes) < 2 {
		t.Errorf("%d writes for 701 relationships, want batches", len(f.writes))
	}
	for _, w := range f.writes {
		if n := len(w.GetUpdates()); n > 500 {
			t.Errorf("a write has %d updates, want at most 500", n)
		}
	}
	if got := len(f.held()); got != 701 {
		t.Errorf("%d relationships held, want 701", got)
	}
}

// TestSyncerFails checks that every failure comes back, so the caller can
// say the change is not in SpiceDB yet.
func TestSyncerFails(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	calls := map[string]func(s *spicedb.Syncer) error{
		"AccountPlaced":   func(s *spicedb.Syncer) error { return s.AccountPlaced(ctx, "a", "core") },
		"AccountUnplaced": func(s *spicedb.Syncer) error { return s.AccountUnplaced(ctx, "a") },
		"Reconcile": func(s *spicedb.Syncer) error {
			return s.Reconcile(ctx, fixed(spicedb.Desired{Accounts: []string{"a"}}))
		},
	}
	for name, call := range calls {
		t.Run(name+" error", func(t *testing.T) {
			t.Parallel()
			if err := call(spicedb.NewSyncer(&fake{err: errSpiceDB})); !errors.Is(err, errSpiceDB) {
				t.Errorf("%s = %v, want the error", name, err)
			}
		})
		t.Run(name+" timeout", func(t *testing.T) {
			t.Parallel()
			if err := call(spicedb.NewSyncer(&fake{block: true}, spicedb.WithTimeout(10*time.Millisecond))); !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("%s with a blocked SpiceDB = %v, want a deadline", name, err)
			}
		})
	}
}

// TestSyncerRefusesAnIDItCantTake checks that an empty ID writes nothing.
func TestSyncerRefusesAnIDItCantTake(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := &fake{}
	s := spicedb.NewSyncer(f)
	if err := s.AccountPlaced(ctx, "", "core"); !errors.Is(err, spicedb.ErrInvalidID) {
		t.Errorf("AccountPlaced of no ID = %v, want ErrInvalidID", err)
	}
	if err := s.Reconcile(ctx, fixed(spicedb.Desired{Accounts: []string{"a", ""}})); !errors.Is(err, spicedb.ErrInvalidID) {
		t.Errorf("Reconcile with no ID = %v, want ErrInvalidID", err)
	}
	if len(f.writes) != 0 || strings.Join(f.held(), "") != "" {
		t.Errorf("wrote %d times, holds %v", len(f.writes), f.held())
	}
}

// TestSyncerReportsAFailedWrite checks a write that fails after a good read.
func TestSyncerReportsAFailedWrite(t *testing.T) {
	t.Parallel()
	err := spicedb.NewSyncer(&writeFails{}).AccountPlaced(t.Context(), "a", "core")
	if !errors.Is(err, errSpiceDB) {
		t.Errorf("AccountPlaced = %v, want the write's error", err)
	}
}

// TestSyncerReportsAFailedLoad checks that a Reconcile whose loader fails
// writes nothing.
func TestSyncerReportsAFailedLoad(t *testing.T) {
	t.Parallel()
	f := &fake{tuples: grantTuples()}
	err := spicedb.NewSyncer(f).Reconcile(t.Context(), func(context.Context) (spicedb.Desired, error) {
		return spicedb.Desired{}, errSpiceDB
	})
	if !errors.Is(err, errSpiceDB) || len(f.writes) != 0 || len(f.held()) != len(grantTuples()) {
		t.Errorf("Reconcile = %v after %d writes holding %v, want the error and no change", err, len(f.writes), f.held())
	}
}

// schemaLog is a spicedb.Schemas that records the schemas written, and fails
// them all with err.
type schemaLog struct {
	written []string
	err     error
}

func (l *schemaLog) WriteSchema(_ context.Context, in *v1.WriteSchemaRequest, _ ...grpc.CallOption) (*v1.WriteSchemaResponse, error) {
	if l.err != nil {
		return nil, l.err
	}
	l.written = append(l.written, in.GetSchema())
	return &v1.WriteSchemaResponse{}, nil
}

// TestSyncerWritesTheSchemaFirst checks that every Reconcile writes the
// schema it was given before it reads or writes a relationship, and that a
// schema SpiceDB refuses stops it with nothing written.
func TestSyncerWritesTheSchemaFirst(t *testing.T) {
	t.Parallel()
	d := spicedb.Desired{Accounts: []string{"a"}}
	t.Run("written once per reconcile", func(t *testing.T) {
		t.Parallel()
		f, l := &fake{}, &schemaLog{}
		s := spicedb.NewSyncer(f, spicedb.WithSchema(l, "definition person {}"))
		for range 2 {
			if err := s.Reconcile(t.Context(), fixed(d)); err != nil {
				t.Fatalf("Reconcile: %v", err)
			}
		}
		if len(l.written) != 2 || l.written[0] != "definition person {}" {
			t.Errorf("schemas written = %q, want the given schema twice", l.written)
		}
		if len(f.writes) != 1 {
			t.Errorf("relationship writes = %d, want 1: the second reconcile changes nothing", len(f.writes))
		}
	})
	t.Run("refused", func(t *testing.T) {
		t.Parallel()
		f := &fake{}
		err := spicedb.NewSyncer(f, spicedb.WithSchema(&schemaLog{err: errSpiceDB}, "x")).Reconcile(t.Context(), fixed(d))
		if !errors.Is(err, errSpiceDB) || f.reads != 0 || len(f.writes) != 0 {
			t.Errorf("Reconcile = %v after %d reads and %d writes, want the schema's error and nothing else", err, f.reads, len(f.writes))
		}
	})
	t.Run("not asked for", func(t *testing.T) {
		t.Parallel()
		if err := spicedb.NewSyncer(&fake{}).Reconcile(t.Context(), fixed(d)); err != nil {
			t.Errorf("Reconcile without a schema = %v", err)
		}
	})
}

// writeFails reads as the fake does and fails every write.
type writeFails struct{ fake }

func (w *writeFails) WriteRelationships(_ context.Context, _ *v1.WriteRelationshipsRequest, _ ...grpc.CallOption) (*v1.WriteRelationshipsResponse, error) {
	return nil, errSpiceDB
}
