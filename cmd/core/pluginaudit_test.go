package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// scriptedPlugins answers each call of InstalledPlugins with the next step's
// records or error, and ends the test's context after the last.
type scriptedPlugins struct {
	steps  []pluginsStep
	cancel context.CancelFunc
	calls  int
}

type pluginsStep struct {
	records []views.PluginRecord
	err     error
}

func (s *scriptedPlugins) InstalledPlugins(context.Context) ([]views.PluginRecord, error) {
	step := s.steps[s.calls]
	s.calls++
	if s.calls == len(s.steps) {
		s.cancel()
	}
	return step.records, step.err
}

// TestObservePluginsEntersWhatTheOperatorDid checks the core's loop over the
// Plugin resources (D-211): before the plugin cache has synced it asks again
// and enters nothing, then it enters the install, the upgrade and the removal
// of a plugin as the operator's, once each however often it looks, with the
// generation of the resource.
func TestObservePluginsEntersWhatTheOperatorDid(t *testing.T) {
	t.Parallel()
	pool, err := pgxpool.New(t.Context(), emptyDatabase(t))
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if migrateErr := migrate(t.Context(), pool); migrateErr != nil {
		t.Fatalf("migrate: %v", migrateErr)
	}
	record := func(generation int64, deleting bool) []views.PluginRecord {
		return []views.PluginRecord{
			{Instance: "xtb-main", Generation: generation, Deleting: deleting},
			{Instance: "nogeneration"},
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	installed := &scriptedPlugins{cancel: cancel, steps: []pluginsStep{
		{err: fmt.Errorf("not synced: %w", viewsapp.ErrPluginsUnknown)},
		{records: record(1, false)},
		{records: record(1, false)},
		{records: record(2, false)},
		{records: record(3, true)},
		{records: nil},
		{records: nil},
	}}
	store := identitypg.New(pool)
	observePlugins(ctx, identityapp.NewPluginObservations(store, time.Now), installed, time.Millisecond, quiet())
	if installed.calls != len(installed.steps) {
		t.Fatalf("the loop looked %d times, want %d", installed.calls, len(installed.steps))
	}
	page, err := store.AuditEntries(t.Context(), &identity.AuditFilter{})
	if err != nil {
		t.Fatalf("AuditEntries: %v", err)
	}
	var got []string
	for _, e := range slices.Backward(page.Entries) {
		if e.Actor.Kind != identity.ActorOperator {
			t.Errorf("%s is an entry of %s, want the operator", e.Action, e.Actor.Kind)
		}
		got = append(got, fmt.Sprintf("%s %s@%s", e.Action, e.Refs[0].ID, e.Refs[1].ID))
	}
	want := []string{"plugin_installed xtb-main@1", "plugin_upgraded xtb-main@2", "plugin_removed xtb-main@2"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the operator's entries, oldest first (-want +got):\n%s", diff)
	}
}

// TestOwnerSessionsRefuseToAuditWhatTheyCannot checks Audited's refusals
// (D-211): a session that is not the owner's is refused as every owner call
// is, an action or a reference kind the identity context doesn't know, and a
// session check that can't record, call no change.
func TestOwnerSessionsRefuseToAuditWhatTheyCannot(t *testing.T) {
	t.Parallel()
	origins, err := identity.NewOrigins([]string{localUI})
	if err != nil {
		t.Fatal(err)
	}
	o := &ownerSessions{auth: sessionsByToken{}, origins: origins, noSession: errNoSessionTest, notOwner: errNotOwnerTest}
	tests := []struct {
		name, token, action string
		refs                []string
		want                error
	}{
		{name: "a demo session", token: "demo", action: "settings_changed", refs: []string{"setting:tax:marginal_rate"}, want: errNotOwnerTest},
		{name: "no session", token: "nope", action: "settings_changed", refs: []string{"setting:tax:marginal_rate"}, want: errNoSessionTest},
		{name: "an action nobody knows", token: "owner", action: "everything_changed", want: identity.ErrInvalidAuditEntry},
		{name: "a kind of record nobody knows", token: "owner", action: "settings_changed", refs: []string{"amount:100"}, want: identity.ErrInvalidAuditEntry},
	}
	for _, tt := range tests {
		called := false
		err := o.Audited(t.Context(), localUI, tt.token, tt.action, tt.refs, func(context.Context) error {
			called = true
			return nil
		})
		if !errors.Is(err, tt.want) || called {
			t.Errorf("%s: Audited = %v, change called %v; want %v and no change", tt.name, err, called, tt.want)
		}
	}
	// A session check that can't record, as a fake one can't, is a log that is unavailable.
	if err := o.Audited(t.Context(), localUI, "owner", "settings_changed", []string{"setting:tax:marginal_rate"}, func(context.Context) error {
		t.Error("the change was made though nothing could enter it")
		return nil
	}); err == nil {
		t.Error("Audited with an auth that can't record = nil, want an error")
	}
}

var (
	errNoSessionTest = errors.New("no session")
	errNotOwnerTest  = errors.New("not the owner")
)
