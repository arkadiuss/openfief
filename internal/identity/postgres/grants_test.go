package postgres_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/postgres"
)

// Synthetic IDs of the grant tests.
const (
	grantTokenID  = "0011223344556677"
	grantOtherID  = "8899aabbccddeeff"
	grantPersonID = "123456789abcdef0"
)

func grantTime() time.Time { return time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC) }

func newGrant(t *testing.T, kind identity.GrantHolderKind, holder string, target identity.GrantTarget, at time.Time) identity.Grant {
	t.Helper()
	g, err := identity.NewGrant(identity.GrantHolder{Kind: kind, ID: holder}, target, at)
	if err != nil {
		t.Fatalf("NewGrant: %v", err)
	}
	return g
}

// mainAccount is the target of the account main.
func mainAccount() identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main"}
}

func slv(id string) identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: id}
}

// seededGrants stores six grants a minute apart, in this order, and returns
// the store: the token's grant of the account main, of the sleeve core and of
// the household, the plugin ui's of the household, a person's of main and
// another token's of main.
func seededGrants(t *testing.T) *postgres.Store {
	t.Helper()
	store := postgres.New(newPool(t))
	tok := identity.GrantHolderScriptToken
	for i, g := range []identity.Grant{
		newGrant(t, tok, grantTokenID, mainAccount(), grantTime()),
		newGrant(t, tok, grantTokenID, slv("core"), grantTime().Add(time.Minute)),
		newGrant(t, tok, grantTokenID, identity.Household(), grantTime().Add(2*time.Minute)),
		newGrant(t, identity.GrantHolderViewPlugin, "ui", identity.Household(), grantTime().Add(3*time.Minute)),
		newGrant(t, identity.GrantHolderPerson, grantPersonID, mainAccount(), grantTime().Add(4*time.Minute)),
		newGrant(t, tok, grantOtherID, mainAccount(), grantTime().Add(5*time.Minute)),
	} {
		if saved, err := store.SaveGrant(t.Context(), &g); err != nil || !saved {
			t.Fatalf("SaveGrant #%d = %v, %v", i, saved, err)
		}
	}
	return store
}

// listed returns the grants m selects as "kind:first two of the ID target".
func listed(t *testing.T, store *postgres.Store, m identity.GrantMatch, inForce bool) []string {
	t.Helper()
	got, err := store.Grants(t.Context(), m, inForce)
	must(t, "Grants", err)
	out := make([]string, 0, len(got))
	for i := range got {
		out = append(out, string(got[i].Holder.Kind)+":"+got[i].Holder.ID[:2]+" "+string(got[i].Target.Kind)+":"+got[i].Target.ID)
	}
	return out
}

// TestGrantStoreSelects checks every selection of grants: all of them oldest
// first, by holder, by target, and one in force at a time per holder and
// target.
func TestGrantStoreSelects(t *testing.T) {
	t.Parallel()
	store := seededGrants(t)
	again := newGrant(t, identity.GrantHolderScriptToken, grantTokenID, mainAccount(), grantTime().Add(time.Hour))
	if saved, err := store.SaveGrant(t.Context(), &again); err != nil || saved {
		t.Errorf("SaveGrant of a grant in force = %v, %v, want nothing stored: one grant in force of a target to a holder", saved, err)
	}
	everyone := []string{
		"script_token:00 account:main", "script_token:00 sleeve:core", "script_token:00 household:home", "view_plugin:ui household:home",
		"person:12 account:main", "script_token:88 account:main",
	}
	if diff := cmp.Diff(everyone, listed(t, store, identity.GrantMatch{}, false)); diff != "" {
		t.Errorf("every grant, oldest first (-want +got):\n%s", diff)
	}
	byHolder := identity.GrantMatch{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: grantTokenID}}
	if diff := cmp.Diff(everyone[:3], listed(t, store, byHolder, false)); diff != "" {
		t.Errorf("a holder's grants (-want +got):\n%s", diff)
	}
	byTarget := identity.GrantMatch{Target: mainAccount()}
	if diff := cmp.Diff([]string{everyone[0], everyone[4], everyone[5]}, listed(t, store, byTarget, false)); diff != "" {
		t.Errorf("a target's grants (-want +got):\n%s", diff)
	}
}

// TestGrantStoreKeepsHistory checks that ending a grant keeps the row, never
// before the grant was made, and lets the same grant be made again.
func TestGrantStoreKeepsHistory(t *testing.T) {
	t.Parallel()
	store := seededGrants(t)
	ctx := t.Context()
	holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: grantTokenID}
	one := identity.GrantMatch{Holder: holder, Target: mainAccount()}
	ended, err := store.RevokeGrants(ctx, one, grantTime().Add(-time.Hour))
	must(t, "RevokeGrants", err)
	if len(ended) != 1 || ended[0].Active() || !ended[0].RevokedAt.Equal(grantTime()) {
		t.Fatalf("RevokeGrants = %+v, want the one grant, ended at its own creation: an ending never precedes it", ended)
	}
	if again, revokeErr := store.RevokeGrants(ctx, one, grantTime()); revokeErr != nil || len(again) != 0 {
		t.Errorf("RevokeGrants of an ended grant = %v, %v, want none", again, revokeErr)
	}
	byHolder := identity.GrantMatch{Holder: holder}
	if got := listed(t, store, byHolder, true); len(got) != 2 {
		t.Errorf("a holder's grants in force = %v, want two", got)
	}
	if got := listed(t, store, byHolder, false); len(got) != 3 {
		t.Errorf("a holder's grants with the history = %v, want three", got)
	}
}

// TestGrantStoreMakesAnEndedGrantAgain checks that a grant that ended can be
// made again, as a new row next to the history.
func TestGrantStoreMakesAnEndedGrantAgain(t *testing.T) {
	t.Parallel()
	store := seededGrants(t)
	ctx := t.Context()
	holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: grantTokenID}
	_, err := store.RevokeGrants(ctx, identity.GrantMatch{Holder: holder, Target: mainAccount()}, grantTime().Add(time.Minute))
	must(t, "RevokeGrants", err)
	remade := newGrant(t, identity.GrantHolderScriptToken, grantTokenID, mainAccount(), grantTime().Add(time.Hour))
	if saved, saveErr := store.SaveGrant(ctx, &remade); saveErr != nil || !saved {
		t.Errorf("SaveGrant of an ended grant again = %v, %v, want a new one", saved, saveErr)
	}
	if got := listed(t, store, identity.GrantMatch{Holder: holder}, false); len(got) != 4 {
		t.Errorf("a holder's grants after it was made again = %v, want four rows", got)
	}
	if got := listed(t, store, identity.GrantMatch{Holder: holder}, true); len(got) != 3 {
		t.Errorf("a holder's grants in force after it was made again = %v, want three", got)
	}
}

// TestGrantStoreEndsByHolderAndTarget checks that ending a holder's grants
// returns each, oldest first, and ending a target's those of every holder.
func TestGrantStoreEndsByHolderAndTarget(t *testing.T) {
	t.Parallel()
	store := seededGrants(t)
	ctx := t.Context()
	byHolder := identity.GrantMatch{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: grantTokenID}}
	gone, err := store.RevokeGrants(ctx, byHolder, grantTime().Add(2*time.Hour))
	must(t, "RevokeGrants of a holder", err)
	order := make([]string, 0, len(gone))
	for i := range gone {
		order = append(order, string(gone[i].Target.Kind))
	}
	if diff := cmp.Diff([]string{"account", "sleeve", "household"}, order); diff != "" {
		t.Errorf("a holder's ended grants, oldest first (-want +got):\n%s", diff)
	}
	byAccount, err := store.RevokeGrants(ctx, identity.GrantMatch{Target: mainAccount()}, grantTime().Add(2*time.Hour))
	must(t, "RevokeGrants of a target", err)
	if len(byAccount) != 2 {
		t.Errorf("a target's ended grants = %d, want the person's and the other token's: the first token's was ended already", len(byAccount))
	}
	left, err := store.Grants(ctx, identity.GrantMatch{}, true)
	must(t, "Grants", err)
	if len(left) != 1 || left[0].Holder.ID != "ui" {
		t.Errorf("grants in force = %+v, want the plugin's alone", left)
	}
}

// TestGrantsTableRefusesWhatTheDomainRefuses checks that the table's
// constraints name the same shapes identity.Grant validates, so a statement
// that bypasses the store still can't put an unfit ID, a free-text name or a
// household that isn't home into a grant.
func TestGrantsTableRefusesWhatTheDomainRefuses(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := t.Context()
	insert := func(hk, hid, tk, tid string) string {
		_, err := pool.Exec(ctx, `INSERT INTO identity.grants (holder_kind, holder_id, target_kind, target_id, created_at) VALUES ($1, $2, $3, $4, now())`, hk, hid, tk, tid)
		return sqlState(err)
	}
	if state := insert("script_token", grantTokenID, "account", "main-1"); state != "" {
		t.Fatalf("a valid grant: SQLSTATE %s", state)
	}
	for name, args := range map[string][4]string{
		"an unknown holder kind":      {"agent", grantTokenID, "household", "home"},
		"a token named by a sentence": {"script_token", "my broker token", "household", "home"},
		"a token in capitals":         {"script_token", "0011223344556ABC", "household", "home"},
		"a plugin in capitals":        {"view_plugin", "UI", "household", "home"},
		"a plugin name too long":      {"view_plugin", strings.Repeat("a", 64), "household", "home"},
		"an unknown target kind":      {"script_token", grantTokenID, "fief", "x"},
		"an account with a space":     {"script_token", grantTokenID, "account", "my account"},
		"an account with no ID":       {"script_token", grantTokenID, "account", ""},
		"a household that is away":    {"script_token", grantTokenID, "household", "away"},
	} {
		if state := insert(args[0], args[1], args[2], args[3]); state != "23514" {
			t.Errorf("%s: SQLSTATE %q, want 23514, a check violation", name, state)
		}
	}
	_, err := pool.Exec(ctx, `INSERT INTO identity.grants (holder_kind, holder_id, target_kind, target_id, created_at, revoked_at)
		VALUES ('script_token', $1, 'household', 'home', now(), now() - interval '1 hour')`, grantTokenID)
	if sqlState(err) != "23514" {
		t.Errorf("a grant revoked before it was made: err = %v, want a check violation", err)
	}
	if state := insert("script_token", grantTokenID, "account", "main-1"); state != "23505" {
		t.Errorf("a second grant in force: SQLSTATE %q, want 23505, a unique violation", state)
	}
}

// TestAuditLogTakesGrantActions checks that the audit log's action list
// includes the two actions of a grant, and still refuses any other.
func TestAuditLogTakesGrantActions(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	store := postgres.New(pool)
	ctx := t.Context()
	holder := identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ui"}
	for _, a := range []identity.AuditAction{identity.ActionGrantCreated, identity.ActionGrantRevoked} {
		e, err := identity.NewAuditEntry(grantTime(), identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}, a, identity.OutcomeOK, holder.Ref(), identity.Household().Ref())
		must(t, "NewAuditEntry", err)
		must(t, "AppendAudit "+string(a), store.AppendAudit(ctx, &e))
	}
	page, err := store.AuditEntries(ctx, &identity.AuditFilter{Action: identity.ActionGrantRevoked})
	must(t, "AuditEntries", err)
	want := []identity.Ref{{Kind: identity.RefViewPlugin, ID: "ui"}, {Kind: identity.RefHousehold, ID: "home"}}
	if len(page.Entries) != 1 || !slices.Equal(page.Entries[0].Refs, want) {
		t.Errorf("grant_revoked entries = %+v, want one naming %v", page.Entries, want)
	}
	_, err = pool.Exec(ctx, `INSERT INTO identity.audit_log (at, actor_kind, actor_id, action, outcome) VALUES (now(), 'system', 'core', 'grant_changed', 'ok')`)
	if sqlState(err) != "23514" {
		t.Errorf("an action outside the list: err = %v, want a check violation", err)
	}
}

// grantDirectory is an app.Directory that knows an account, a sleeve and a
// plugin.
type grantDirectory struct{}

func (grantDirectory) TargetExists(_ context.Context, t identity.GrantTarget) (bool, error) {
	return t.Kind == identity.GrantTargetHousehold || t.ID == "main" || t.ID == "core", nil
}

func (grantDirectory) ViewPluginInstalled(_ context.Context, name string) (bool, error) {
	return name == "ui", nil
}

func (grantDirectory) ViewPlugins(context.Context) ([]string, error) { return []string{"ui"}, nil }

// grantCases is the grant use cases over a real store with a settable clock.
type grantCases struct {
	store  *postgres.Store
	tokens *app.Tokens
	grants *app.Grants
	now    time.Time
}

func newGrantCases(t *testing.T) *grantCases {
	t.Helper()
	c := &grantCases{store: postgres.New(newPool(t)), now: grantTime()}
	clock := func() time.Time { return c.now }
	c.tokens = app.NewTokens(c.store, clock, nil)
	c.grants = app.NewGrants(c.store, c.tokens, grantDirectory{}, nil, clock)
	return c
}

func (c *grantCases) issue(t *testing.T, name string) identity.ScriptToken {
	t.Helper()
	tok, _, err := c.tokens.Issue(t.Context(), name, time.Hour)
	must(t, "Issue "+name, err)
	return tok
}

func systemActor() identity.Actor {
	return identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
}

// TestGrantFailsClosedOverPostgres checks that a grant and its audit entry
// are one transaction: with an audit log that refuses the entry, the grant is
// not stored.
func TestGrantFailsClosedOverPostgres(t *testing.T) {
	t.Parallel()
	c := newGrantCases(t)
	ctx := t.Context()
	holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(c.issue(t, "nightly").ID)}
	down := app.NewGrants(auditDown{c.store}, c.tokens, grantDirectory{}, nil, func() time.Time { return c.now })
	if _, err := down.Create(ctx, systemActor(), holder, mainAccount()); err == nil {
		t.Fatal("Create with the audit log down succeeded")
	}
	if got, err := c.grants.List(ctx, identity.GrantMatch{}, false); err != nil || len(got) != 0 {
		t.Errorf("grants after the failed create = %v, %v, want none", got, err)
	}
}

// TestGrantAuditTrailOverPostgres checks the entries the grant use cases
// leave over the real store, and that a token's revocation ends its grants in
// the same transaction, each with its entry.
func TestGrantAuditTrailOverPostgres(t *testing.T) {
	t.Parallel()
	c := newGrantCases(t)
	ctx := t.Context()
	tok := c.issue(t, "nightly")
	holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(tok.ID)}
	for _, target := range []identity.GrantTarget{mainAccount(), slv("core"), identity.Household()} {
		if _, err := c.grants.Create(ctx, systemActor(), holder, target); err != nil {
			t.Fatalf("Create(%v): %v", target, err)
		}
	}
	must(t, "Revoke", c.grants.Revoke(ctx, systemActor(), holder, slv("core")))
	must(t, "RevokeToken", c.grants.RevokeToken(ctx, systemActor(), tok.ID))
	if got, err := c.grants.InForce(ctx); err != nil || len(got) != 0 {
		t.Errorf("grants in force after the token was revoked = %v, %v, want none", got, err)
	}
	page, err := c.store.AuditEntries(ctx, &identity.AuditFilter{})
	must(t, "AuditEntries", err)
	trail := make([]string, 0, len(page.Entries))
	for _, e := range slices.Backward(page.Entries) {
		words := make([]string, 0, 1+len(e.Refs))
		words = append(words, string(e.Action))
		for _, r := range e.Refs {
			words = append(words, string(r.Kind))
		}
		trail = append(trail, strings.Join(words, " "))
	}
	want := []string{
		"grant_created script_token account", "grant_created script_token sleeve", "grant_created script_token household",
		"grant_revoked script_token sleeve",
		"script_token_revoked script_token", "grant_revoked script_token account", "grant_revoked script_token household",
	}
	if diff := cmp.Diff(want, trail); diff != "" {
		t.Errorf("audit trail (-want +got):\n%s", diff)
	}
}

// TestCutoverOverPostgres checks that the cutover gives the token that has no
// grant and the plugin the household, once, and none to a token made after it,
// with each grant and entry in one transaction.
func TestCutoverOverPostgres(t *testing.T) {
	t.Parallel()
	c := newGrantCases(t)
	ctx := t.Context()
	fresh := c.issue(t, "fresh")
	if n, err := c.grants.Cutover(ctx, systemActor()); err != nil || n != 2 {
		t.Fatalf("Cutover = %d, %v, want 2: the token and the ui plugin", n, err)
	}
	later := c.issue(t, "after")
	if n, err := c.grants.Cutover(ctx, systemActor()); err != nil || n != 0 {
		t.Errorf("a second Cutover = %d, %v, want 0", n, err)
	}
	got, err := c.grants.InForce(ctx)
	must(t, "InForce", err)
	holders := make([]string, 0, len(got))
	for i := range got {
		holders = append(holders, got[i].Holder.ID)
	}
	slices.Sort(holders)
	want := []string{string(fresh.ID), "ui"}
	slices.Sort(want)
	if diff := cmp.Diff(want, holders); diff != "" {
		t.Errorf("holders with a grant (-want +got):\n%s", diff)
	}
	if slices.Contains(holders, string(later.ID)) {
		t.Error("a token made after the cutover was given a grant")
	}
}
