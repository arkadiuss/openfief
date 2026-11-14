package postgres_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
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
		{name: "the administrator", kind: "admin", id: "admin"},
		{name: "the administrator as the core", kind: "admin", id: "core", wantErr: true},
		{name: "the core as the administrator", kind: "system", id: "admin", wantErr: true},
	} {
		_, err := pool.Exec(t.Context(), insert, tc.kind, tc.id, "plugin_installed", "ok")
		if gotErr := err != nil; gotErr != tc.wantErr {
			t.Errorf("%s: error = %v, want an error: %v", tc.name, err, tc.wantErr)
		}
	}
}

// TestAppendAuditOnceWritesAnEntryOncePerKey checks, over the real store, the
// outbox's write (D-211): an entry keyed by its job's ID is written once
// however often the worker runs, with the time of the change and not of the
// write; a second job's entry is written; and an entry the domain refuses is
// not.
func TestAppendAuditOnceWritesAnEntryOncePerKey(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	at := time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC)
	first := deletedAccountEntry(t, at, "xtb-main")
	for i, want := range []bool{true, false, false} {
		again := first
		wrote, err := store.AppendAuditOnce(t.Context(), 41, &again)
		if err != nil || wrote != want {
			t.Fatalf("run %d of job 41: AppendAuditOnce = %v, %v; want %v", i+1, wrote, err, want)
		}
	}
	other := deletedAccountEntry(t, at, "ibkr-main")
	if wrote, err := store.AppendAuditOnce(t.Context(), 42, &other); err != nil || !wrote {
		t.Fatalf("job 42: AppendAuditOnce = %v, %v; want it written", wrote, err)
	}
	page, err := store.AuditEntries(t.Context(), &identity.AuditFilter{Action: identity.ActionAccountDeleted})
	if err != nil || len(page.Entries) != 2 {
		t.Fatalf("entries = %d, %v; want one per job", len(page.Entries), err)
	}
	for _, e := range page.Entries {
		if !e.At.Equal(at) {
			t.Errorf("an entry's time = %s, want the change's, %s", e.At, at)
		}
	}
}

// TestAppendAuditOnceRefusesWhatTheDomainRefusesAndTheTableKeepsTheKeyUnique
// checks that an entry the domain refuses is not written, and that the key is
// unique in the table itself, not only in the statement.
func TestAppendAuditOnceRefusesWhatTheDomainRefusesAndTheTableKeepsTheKeyUnique(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	at := time.Date(2026, time.March, 2, 10, 0, 0, 0, time.UTC)
	bad := deletedAccountEntry(t, at, "xtb-main")
	bad.Refs = []identity.Ref{{Kind: identity.RefAccount, ID: "an account number: 12 3456"}}
	if _, err := store.AppendAuditOnce(t.Context(), 43, &bad); !errors.Is(err, identity.ErrInvalidAuditEntry) {
		t.Errorf("an entry the domain refuses: %v, want ErrInvalidAuditEntry", err)
	}
	good := deletedAccountEntry(t, at, "xtb-main")
	if _, err := store.AppendAuditOnce(t.Context(), 41, &good); err != nil {
		t.Fatalf("AppendAuditOnce: %v", err)
	}
	if _, err := pool.Exec(t.Context(), `
		INSERT INTO identity.audit_log (at, actor_kind, actor_id, action, outcome, refs, outbox_job_id)
		VALUES (now(), 'system', 'core', 'sign_out_everywhere', 'ok', '[]', 41)`); err == nil {
		t.Error("a second row under the key of job 41: no error, want the unique constraint to refuse it")
	}
}

// deletedAccountEntry returns the entry of the owner deleting account at at.
func deletedAccountEntry(t *testing.T, at time.Time, account string) identity.AuditEntry {
	t.Helper()
	e, err := identity.NewAuditEntry(at, identity.Actor{Kind: identity.ActorOwner, ID: auditSession}, identity.ActionAccountDeleted,
		identity.OutcomeOK, identity.Ref{Kind: identity.RefAccount, ID: account})
	if err != nil {
		t.Fatalf("NewAuditEntry: %v", err)
	}
	return e
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
