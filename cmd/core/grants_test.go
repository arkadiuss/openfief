package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// fakeAccounts is an accountLookup over a set of IDs.
type fakeAccounts map[string]bool

func (f fakeAccounts) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	if !f[string(id)] {
		return ledger.Account{}, ledgerapp.ErrUnknownAccount
	}
	return ledger.Account{ID: id}, nil
}

// fakeSleeves is a sleeveLister over a list of IDs.
type fakeSleeves []string

func (f fakeSleeves) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	out := make([]portfolio.Sleeve, 0, len(f))
	for _, id := range f {
		out = append(out, portfolio.Sleeve{ID: portfolio.SleeveID(id)})
	}
	return out, nil
}

// fakePlugins is an installedPlugins over records, or one whose cache has
// not synced.
type fakePlugins struct {
	records  []views.PluginRecord
	unsynced bool
}

func (f *fakePlugins) InstalledPlugins(context.Context) ([]views.PluginRecord, error) {
	if f.unsynced {
		return nil, viewsapp.ErrPluginsUnknown
	}
	return slices.Clone(f.records), nil
}

// directoryPlugins are the installed plugins the directory tests read: two
// that hold or ask for view, a connector and one being removed.
func directoryPlugins() *fakePlugins {
	return &fakePlugins{records: []views.PluginRecord{
		{Instance: "ui", Effective: []string{"view"}},
		{Instance: "asks-for-view", Requested: []string{"view"}},
		{Instance: "ibkr", Requested: []string{"connector"}, Effective: []string{"connector"}},
		{Instance: "removing", Effective: []string{"view"}, Deleting: true},
	}}
}

func TestGrantDirectoryTargets(t *testing.T) {
	t.Parallel()
	d := &grantDirectory{accounts: fakeAccounts{"main": true}, sleeves: fakeSleeves{"core"}, plugins: directoryPlugins()}
	for _, tc := range []struct {
		target identity.GrantTarget
		want   bool
	}{
		{identity.Household(), true},
		{identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main"}, true},
		{identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "other"}, false},
		{identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "core"}, true},
		{identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "main"}, false},
	} {
		if got, err := d.TargetExists(t.Context(), tc.target); err != nil || got != tc.want {
			t.Errorf("TargetExists(%v) = %v, %v, want %v", tc.target, got, err, tc.want)
		}
	}
}

func TestGrantDirectoryPlugins(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	plugins := directoryPlugins()
	d := &grantDirectory{accounts: fakeAccounts{}, sleeves: fakeSleeves{}, plugins: plugins}
	for name, want := range map[string]bool{"ui": true, "asks-for-view": true, "ibkr": false, "removing": false, "absent": false} {
		if got, err := d.ViewPluginInstalled(ctx, name); err != nil || got != want {
			t.Errorf("ViewPluginInstalled(%s) = %v, %v, want %v", name, got, err, want)
		}
	}
	got, err := d.ViewPlugins(ctx)
	if err != nil || !slices.Equal(got, []string{"asks-for-view", "ui"}) {
		t.Errorf("ViewPlugins = %v, %v, want the two that hold or ask for view, sorted", got, err)
	}
	plugins.unsynced = true
	if _, err := d.ViewPluginInstalled(ctx, "ui"); !errors.Is(err, identityapp.ErrDirectoryUnknown) {
		t.Errorf("ViewPluginInstalled before the cache synced = %v, want ErrDirectoryUnknown", err)
	}
	if _, err := d.ViewPlugins(ctx); !errors.Is(err, identityapp.ErrDirectoryUnknown) {
		t.Errorf("ViewPlugins before the cache synced = %v, want ErrDirectoryUnknown", err)
	}
}

// TestGrantDirectoryOutsideACluster checks that outside a cluster no plugin
// is installed, so there is none to give a grant, and one can't be told of.
func TestGrantDirectoryOutsideACluster(t *testing.T) {
	t.Parallel()
	outside := &grantDirectory{accounts: fakeAccounts{}, sleeves: fakeSleeves{}}
	if got, err := outside.ViewPlugins(t.Context()); err != nil || len(got) != 0 {
		t.Errorf("ViewPlugins outside a cluster = %v, %v, want none", got, err)
	}
	if _, err := outside.ViewPluginInstalled(t.Context(), "ui"); !errors.Is(err, identityapp.ErrDirectoryUnknown) {
		t.Errorf("ViewPluginInstalled outside a cluster = %v, want ErrDirectoryUnknown", err)
	}
}

// grantsFixture is the grant use cases over a fresh PostgreSQL, with a syncer
// over a recording SpiceDB and a directory with the plugins and accounts a test
// sets.
type grantsFixture struct {
	pool    *pgxpool.Pool
	grants  *identityapp.Grants
	tokens  tokenAdmin
	plugins *fakePlugins
	spicedb *writeLog
	dir     *grantDirectory
	now     time.Time
}

func newGrantsFixture(t *testing.T) *grantsFixture {
	t.Helper()
	dsn := emptyDatabase(t)
	ctx := t.Context()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	if migrateErr := identitypg.Migrate(ctx, pool); migrateErr != nil {
		t.Fatalf("Migrate: %v", migrateErr)
	}
	f := &grantsFixture{pool: pool, plugins: &fakePlugins{}, spicedb: &writeLog{}, now: time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC)}
	clock := func() time.Time { return f.now }
	store := identitypg.New(pool)
	toks := identityapp.NewTokens(store, clock, nil)
	f.dir = &grantDirectory{accounts: fakeAccounts{"main": true}, sleeves: fakeSleeves{"core"}, plugins: f.plugins}
	f.grants = identityapp.NewGrants(store, toks, f.dir, spicedb.NewSyncer(f.spicedb), clock)
	f.tokens = tokenAdmin{store: store, tokens: toks, grants: identityapp.NewGrants(store, toks, householdOnly{}, nil, clock)}
	return f
}

// trail returns the audit log as one line per entry, oldest first.
func (f *grantsFixture) trail(t *testing.T) []string {
	t.Helper()
	rows, err := f.pool.Query(t.Context(), `SELECT actor_kind || ' ' || action || ' ' || outcome FROM identity.audit_log ORDER BY id`)
	if err != nil {
		t.Fatalf("reading the audit log: %v", err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("reading the audit log: %v", err)
	}
	return got
}

// TestGrantsStartupCutsOverOnce checks D-209's step for the day SpiceDB takes
// over, as the core runs it at start: every script token and every installed
// view plugin without a grant gets the household grant, each with the core's
// own audit entry, written to SpiceDB; a second start writes nothing; and under
// the policy the start gives nobody a grant.
func TestGrantsStartupCutsOverOnce(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	ctx := t.Context()
	f.plugins.records = []views.PluginRecord{{Instance: "ui", Effective: []string{"view"}}, {Instance: "ibkr", Effective: []string{"connector"}}}
	tok, _, err := f.tokens.Issue(ctx, "nightly", time.Hour, false)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	log := quiet()

	grantsStartup(ctx, f.grants, false, nil, log)
	issued := []string{"admin script_token_created ok"}
	if got := f.trail(t); !slices.Equal(got, issued) {
		t.Fatalf("a start under the policy left %v, want only the token's creation", got)
	}

	grantsStartup(ctx, f.grants, true, nil, log)
	want := []string{
		"OPERATION_TOUCH household:home#reader@script_token:" + string(tok.ID),
		"OPERATION_TOUCH household:home#reader@view_plugin:ui",
	}
	if !slices.Equal(f.spicedb.writes, want) {
		t.Errorf("written to SpiceDB = %v, want %v", f.spicedb.writes, want)
	}
	if diff := cmp.Diff(append(slices.Clone(issued), "system grant_created ok", "system grant_created ok"), f.trail(t)); diff != "" {
		t.Errorf("audit log (-want +got):\n%s", diff)
	}

	// A token made afterwards, and a second start: nothing is written.
	if _, _, issueErr := f.tokens.Issue(ctx, "later", time.Hour, false); issueErr != nil {
		t.Fatalf("Issue: %v", issueErr)
	}
	grantsStartup(ctx, f.grants, true, nil, log)
	if len(f.spicedb.writes) != len(want) {
		t.Errorf("a second start wrote %v", f.spicedb.writes[len(want):])
	}
	inForce, err := f.grants.InForce(ctx)
	if err != nil || len(inForce) != 2 {
		t.Errorf("grants in force = %v, %v, want the two of the cutover", inForce, err)
	}
}

// TestGrantsStartupWaitsForThePlugins checks that a start whose plugin cache
// has not synced does not finish the cutover and retries until it has,
// instead of giving the household grant to some holders and marking it done.
func TestGrantsStartupWaitsForThePlugins(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	f.plugins.unsynced = true
	f.plugins.records = []views.PluginRecord{{Instance: "ui", Effective: []string{"view"}}}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	grantsStartup(ctx, f.grants, true, nil, quiet())
	if got := f.trail(t); len(got) != 0 {
		t.Errorf("a cutover that couldn't read the plugins wrote %v", got)
	}
	var done bool
	if err := f.pool.QueryRow(t.Context(), `SELECT EXISTS (SELECT 1 FROM identity.grant_cutover)`).Scan(&done); err != nil || done {
		t.Errorf("the cutover is marked done = %v, %v, want not", done, err)
	}
}

// grantRows are the grants TestGrantsFollowTheirRecords makes: the plugin ui's
// and the plugin ghost's of the household, and a token's of the sleeve core and
// of the account main.
func grantRows(token identity.GrantHolder) []struct {
	holder identity.GrantHolder
	target identity.GrantTarget
} {
	ui := identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ui"}
	ghost := identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ghost"}
	return []struct {
		holder identity.GrantHolder
		target identity.GrantTarget
	}{
		{ui, identity.Household()},
		{ghost, identity.Household()},
		{token, identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "core"}},
		{token, identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main"}},
	}
}

// newFollowing is a fixture with the plugins ui and ghost installed, a token
// and the grants of grantRows made.
func newFollowing(t *testing.T) *grantsFixture {
	t.Helper()
	f := newGrantsFixture(t)
	f.plugins.records = []views.PluginRecord{{Instance: "ui", Effective: []string{"view"}}, {Instance: "ghost", Effective: []string{"view"}}}
	tok, _, err := f.tokens.Issue(t.Context(), "nightly", time.Hour, false)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(tok.ID)}
	for _, g := range grantRows(holder) {
		if _, createErr := f.grants.Create(t.Context(), systemActor(), g.holder, g.target); createErr != nil {
			t.Fatalf("Create(%v, %v): %v", g.holder, g.target, createErr)
		}
	}
	return f
}

// TestGrantsFollowTheirRecords checks that the core ends, as itself, the
// grants of a plugin that was uninstalled, of a sleeve that was deleted and
// of an account that was deleted, each with an entry.
func TestGrantsFollowTheirRecords(t *testing.T) {
	t.Parallel()
	f := newFollowing(t)
	ctx := t.Context()
	grantsOnUninstall(ctx, f.grants, quiet())("ui")
	grantsOnUninstall(ctx, f.grants, quiet())("not a name")
	cleanup := grantCleanup{f.grants}
	for name, err := range map[string]error{
		"SleeveRemoved":  cleanup.SleeveRemoved(ctx, "core"),
		"AccountRemoved": cleanup.AccountRemoved(ctx, "main"),
		"AccountPlaced":  cleanup.AccountPlaced(ctx, "main", "core"),
	} {
		if err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	inForce, err := f.grants.InForce(ctx)
	if err != nil || len(inForce) != 1 || inForce[0].Holder.ID != "ghost" {
		t.Fatalf("grants in force = %v, %v, want only the ghost plugin's", inForce, err)
	}
	revoked := 0
	for _, line := range f.trail(t) {
		if line == "system grant_revoked ok" {
			revoked++
		}
	}
	if revoked != 3 {
		t.Errorf("%d revocations by the core, want 3 (a plugin's, a sleeve's and an account's)", revoked)
	}
}

// TestGrantsStartupSweepsWhatWentWhileTheCoreWasDown checks that the start's
// sweep ends the grant of a plugin that is gone from the cluster.
func TestGrantsStartupSweepsWhatWentWhileTheCoreWasDown(t *testing.T) {
	t.Parallel()
	f := newFollowing(t)
	ctx := t.Context()
	f.plugins.records = f.plugins.records[:1] // the ghost plugin is gone behind the core's back
	grantsStartup(ctx, f.grants, false, nil, quiet())
	inForce, err := f.grants.InForce(ctx)
	if err != nil {
		t.Fatalf("InForce: %v", err)
	}
	for i := range inForce {
		if inForce[i].Holder.ID == "ghost" {
			t.Errorf("the ghost plugin's grant is in force after the sweep: %+v", inForce[i])
		}
	}
}

// TestTokenIssueGrantsTheHousehold checks that token issue --grant-household
// gives the token the household grant, with its entry.
func TestTokenIssueGrantsTheHousehold(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	ctx := t.Context()
	tok, secret, err := f.tokens.Issue(ctx, "nightly", time.Hour, true)
	if err != nil || !strings.HasPrefix(secret, "ofs_") {
		t.Fatalf("Issue = %v, %v", tok, err)
	}
	inForce, err := f.grants.InForce(ctx)
	if err != nil || len(inForce) != 1 || inForce[0].Target != identity.Household() || inForce[0].Holder.ID != string(tok.ID) {
		t.Errorf("grants in force = %v, %v, want the token's household grant", inForce, err)
	}
	want := []string{"admin script_token_created ok", "admin grant_created ok"}
	if diff := cmp.Diff(want, f.trail(t)); diff != "" {
		t.Errorf("audit log (-want +got):\n%s", diff)
	}
}

// TestTokenRevokeEndsItsGrants checks that token revoke ends the token's
// grants, with their entries.
func TestTokenRevokeEndsItsGrants(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	ctx := t.Context()
	tok, _, err := f.tokens.Issue(ctx, "nightly", time.Hour, true)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if revokeErr := f.tokens.Revoke(ctx, tok.ID); revokeErr != nil {
		t.Fatalf("Revoke: %v", revokeErr)
	}
	if left, listErr := f.grants.InForce(ctx); listErr != nil || len(left) != 0 {
		t.Errorf("grants in force after the revocation = %v, %v, want none", left, listErr)
	}
	want := []string{"admin script_token_created ok", "admin grant_created ok", "admin script_token_revoked ok", "admin grant_revoked ok"}
	if diff := cmp.Diff(want, f.trail(t)); diff != "" {
		t.Errorf("audit log (-want +got):\n%s", diff)
	}
}

// wantUngranted fails t unless check finds the view plugins named by want, in
// order, and none other.
func wantUngranted(t *testing.T, check attentionapp.Check, when string, want ...string) {
	t.Helper()
	got, err := check.Find(t.Context())
	if err != nil {
		t.Fatalf("Find %s: %v", when, err)
	}
	subjects := make([]string, 0, len(got))
	for i := range got {
		subjects = append(subjects, got[i].Subject)
	}
	if !slices.Equal(subjects, want) && (len(subjects) != 0 || len(want) != 0) {
		t.Errorf("view plugins without a grant %s = %v, want %v", when, subjects, want)
	}
}

// TestOwnerGrantsAnUngrantedViewPluginTheHousehold checks D-196's one action
// for a view plugin that can't read anything: the owner's household grant is
// the existing CreateGrant, which writes one audit entry by the owner and the
// relationship to SpiceDB, and afterwards the check of ungranted view plugins
// finds the plugin no more; it finds none either once the plugin is uninstalled.
func TestOwnerGrantsAnUngrantedViewPluginTheHousehold(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	f.plugins.records = []views.PluginRecord{readyView("ghostfolio")}
	check := ungrantedViewPlugins(true, f.plugins, f.grants)
	wantUngranted(t, check, "before the grant", "ghostfolio")

	owner := identity.Actor{Kind: identity.ActorOwner, ID: "0123456789abcdef"}
	holder := identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ghostfolio"}
	if _, err := f.grants.Create(t.Context(), owner, holder, identity.Household()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got, want := f.spicedb.writes, []string{"OPERATION_TOUCH household:home#reader@view_plugin:ghostfolio"}; !slices.Equal(got, want) {
		t.Errorf("written to SpiceDB = %v, want %v", got, want)
	}
	if got, want := f.trail(t), []string{"owner grant_created ok"}; !slices.Equal(got, want) {
		t.Errorf("audit log = %v, want %v", got, want)
	}
	wantUngranted(t, check, "after the grant")

	f.plugins.records = []views.PluginRecord{readyView("second")}
	wantUngranted(t, check, "with another plugin", "second")
	f.plugins.records = nil
	wantUngranted(t, check, "after the uninstall")
}

// TestGrantDirectoryAccessViewPlugins checks the view plugins the Access page
// lists (D-210): those that hold or ask for the view capability, less the one
// that serves browsers, which reads as the signed-in person, and any being
// removed; none outside a cluster, and an error while the cache has not
// synced.
func TestGrantDirectoryAccessViewPlugins(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	plugins := &fakePlugins{records: []views.PluginRecord{
		{Instance: "ui", Effective: []string{"view"}, BrowserPath: "/"},
		{Instance: "ghostfolio", Effective: []string{"view"}},
		{Instance: "asks-for-view", Requested: []string{"view"}},
		{Instance: "ibkr", Effective: []string{"connector"}},
		{Instance: "removing", Effective: []string{"view"}, Deleting: true},
	}}
	d := &grantDirectory{accounts: fakeAccounts{}, sleeves: fakeSleeves{}, plugins: plugins}
	got, err := d.AccessViewPlugins(ctx)
	if err != nil || !slices.Equal(got, []string{"asks-for-view", "ghostfolio"}) {
		t.Errorf("AccessViewPlugins = %v, %v, want the two that read as themselves, sorted", got, err)
	}
	if got, err := (&grantDirectory{}).AccessViewPlugins(ctx); err != nil || len(got) != 0 {
		t.Errorf("AccessViewPlugins outside a cluster = %v, %v, want none", got, err)
	}
	plugins.unsynced = true
	if _, err := d.AccessViewPlugins(ctx); !errors.Is(err, identityapp.ErrDirectoryUnknown) {
		t.Errorf("AccessViewPlugins before the cache synced = %v, want ErrDirectoryUnknown", err)
	}
}
