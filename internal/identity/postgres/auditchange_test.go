package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
)

// TestAuditLogTakesEveryDomainAction checks that the database's closed set of
// actions is the domain's (migration 0014): a row of every action
// identity.AuditActions lists is accepted, and the operator is an actor with
// its own ID and no other.
func TestAuditLogTakesEveryDomainAction(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	const insert = `INSERT INTO identity.audit_log (at, actor_kind, actor_id, action, outcome, refs) VALUES (now(), $1, $2, $3, $4, '[]')`
	for _, a := range identity.AuditActions() {
		if _, err := pool.Exec(t.Context(), insert, "system", "core", string(a), "ok"); err != nil {
			t.Errorf("an entry of %s: %v", a, err)
		}
	}
	for _, tc := range []struct {
		name, kind, id string
		wantErr        bool
	}{
		{name: "the operator", kind: "operator", id: "operator"},
		{name: "the operator as the core", kind: "operator", id: "core", wantErr: true},
		{name: "the operator with no ID", kind: "operator", id: "", wantErr: true},
		{name: "the core as the operator", kind: "system", id: "operator", wantErr: true},
		{name: "an owner called the operator", kind: "owner", id: "operator", wantErr: true},
	} {
		_, err := pool.Exec(t.Context(), insert, tc.kind, tc.id, "plugin_installed", "ok")
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: error = %v, want an error: %v", tc.name, err, tc.wantErr)
		}
	}
}

var errChangeBroke = errors.New("the change broke")

// ownerChangeAuth returns the auth use cases over the real store of pool.
func ownerChangeAuth(t *testing.T, store *postgres.Store) *app.Auth {
	t.Helper()
	origins, err := identity.NewOrigins([]string{localUI})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	cfg := app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy()}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), app.NewTokens(store, issuedAt, nil), nil, cfg, issuedAt, nil)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	return auth
}

// TestOwnerChangeEntryCommitsOnlyWithTheChange checks, over the real store,
// RecordOwnerChange (D-211): the entry of a change that was made is in the
// log, the entry of one that failed is not, and neither is a change made when
// the log can't take its entry.
func TestOwnerChangeEntryCommitsOnlyWithTheChange(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	auth := ownerChangeAuth(t, store)
	ctx := t.Context()
	refs := []identity.Ref{{Kind: identity.RefAccount, ID: "xtb-main"}}
	count := func() int {
		t.Helper()
		page, err := store.AuditEntries(ctx, &identity.AuditFilter{Action: identity.ActionAccountDeleted})
		if err != nil {
			t.Fatalf("AuditEntries: %v", err)
		}
		return len(page.Entries)
	}

	made := 0
	if err := auth.RecordOwnerChange(ctx, auditSession, identity.ActionAccountDeleted, refs, func(context.Context) error { made++; return nil }); err != nil {
		t.Fatalf("RecordOwnerChange: %v", err)
	}
	if made != 1 || count() != 1 {
		t.Errorf("after a change: made %d times, %d entries; want 1 and 1", made, count())
	}
	err := auth.RecordOwnerChange(ctx, auditSession, identity.ActionAccountDeleted, refs, func(context.Context) error { return errChangeBroke })
	if !errors.Is(err, errChangeBroke) || count() != 1 {
		t.Errorf("a failing change = %v with %d entries, want its error and the one entry", err, count())
	}
	if _, execErr := pool.Exec(ctx, "ALTER TABLE identity.audit_log RENAME TO audit_log_away"); execErr != nil {
		t.Fatalf("taking the log away: %v", execErr)
	}
	made = 0
	err = auth.RecordOwnerChange(ctx, auditSession, identity.ActionAccountDeleted, refs, func(context.Context) error { made++; return nil })
	if err == nil || made != 0 {
		t.Errorf("a change with the log gone = %v, made %d times; want an error and no change", err, made)
	}
}

// TestPluginGenerationsAreEnteredOncePerGeneration checks the observation of
// plugins over the real store: an install, an upgrade and a removal are
// entries by the operator, seeing the same again writes nothing, and the
// generation and the entry commit together.
func TestPluginGenerationsAreEnteredOncePerGeneration(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	ctx := t.Context()
	o := app.NewPluginObservations(store, issuedAt)
	steps := []struct {
		observed []app.ObservedPlugin
		want     int
	}{
		{[]app.ObservedPlugin{{Instance: "xtb-main", Generation: 1}}, 1},
		{[]app.ObservedPlugin{{Instance: "xtb-main", Generation: 1}}, 0},
		{[]app.ObservedPlugin{{Instance: "xtb-main", Generation: 2}}, 1},
		{[]app.ObservedPlugin{{Instance: "xtb-main", Generation: 2, Deleting: true}}, 0},
		{nil, 1},
		{nil, 0},
		{[]app.ObservedPlugin{{Instance: "xtb-main", Generation: 1}}, 1},
	}
	for i, s := range steps {
		n, err := o.Observe(ctx, s.observed)
		if err != nil || n != s.want {
			t.Fatalf("step %d: Observe = %d, %v; want %d", i, n, err, s.want)
		}
	}
	page, err := store.AuditEntries(ctx, &identity.AuditFilter{ActorKind: identity.ActorOperator, ActorID: identity.OperatorActorID})
	if err != nil {
		t.Fatalf("AuditEntries: %v", err)
	}
	var got []string
	for _, e := range slices.Backward(page.Entries) {
		got = append(got, string(e.Action)+" "+e.Refs[0].ID+"@"+e.Refs[1].ID)
	}
	want := []string{
		"plugin_installed xtb-main@1", "plugin_upgraded xtb-main@2", "plugin_removed xtb-main@2", "plugin_installed xtb-main@1",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the operator's entries, oldest first (-want +got):\n%s", diff)
	}
}
