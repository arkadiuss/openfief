package app_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
)

var errModelDown = errors.New("authorization model unreachable")

// The synthetic holders and targets of the tests.
const (
	pluginUI     = "ui"
	pluginGhost  = "ghostfolio"
	accountMain  = "main-account"
	accountOther = "other-account"
	sleeveCore   = "core"
)

// SaveGrant implements app.GrantStore over memory.
func (m *memAuthStore) SaveGrant(_ context.Context, g *identity.Grant) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.grants {
		if m.grants[i].Active() && m.grants[i].Holder == g.Holder && m.grants[i].Target == g.Target {
			return false, nil
		}
	}
	m.grants = append(m.grants, *g)
	return true, nil
}

// RevokeGrants implements app.GrantStore over memory.
func (m *memAuthStore) RevokeGrants(_ context.Context, match identity.GrantMatch, at time.Time) ([]identity.Grant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var ended []identity.Grant
	for i := range m.grants {
		if m.grants[i].Active() && match.Matches(&m.grants[i]) {
			m.grants[i] = m.grants[i].Revoked(at)
			ended = append(ended, m.grants[i])
		}
	}
	return ended, nil
}

// Grants implements app.GrantStore over memory.
func (m *memAuthStore) Grants(_ context.Context, match identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []identity.Grant
	for i := range m.grants {
		if match.Matches(&m.grants[i]) && (!inForceOnly || m.grants[i].Active()) {
			out = append(out, m.grants[i])
		}
	}
	return out, nil
}

// GrantCutoverDone implements app.GrantStore over memory.
func (m *memAuthStore) GrantCutoverDone(context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cutover, nil
}

// MarkGrantCutover implements app.GrantStore over memory.
func (m *memAuthStore) MarkGrantCutover(context.Context, time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	first := !m.cutover
	m.cutover = true
	return first, nil
}

// fakeDirectory is an app.Directory over fixed sets.
type fakeDirectory struct {
	mu       sync.Mutex
	accounts map[string]bool
	sleeves  map[string]bool
	plugins  map[string]bool
	// unknown makes every plugin question fail as the core can't tell.
	unknown bool
}

func newFakeDirectory() *fakeDirectory {
	return &fakeDirectory{
		accounts: map[string]bool{accountMain: true, accountOther: true},
		sleeves:  map[string]bool{sleeveCore: true},
		plugins:  map[string]bool{pluginUI: true, pluginGhost: true},
	}
}

func (d *fakeDirectory) TargetExists(_ context.Context, t identity.GrantTarget) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch t.Kind {
	case identity.GrantTargetHousehold:
		return true, nil
	case identity.GrantTargetAccount:
		return d.accounts[t.ID], nil
	case identity.GrantTargetSleeve:
		return d.sleeves[t.ID], nil
	}
	return false, nil
}

func (d *fakeDirectory) ViewPluginInstalled(_ context.Context, instance string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.unknown {
		return false, app.ErrDirectoryUnknown
	}
	return d.plugins[instance], nil
}

func (d *fakeDirectory) ViewPlugins(context.Context) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.unknown {
		return nil, app.ErrDirectoryUnknown
	}
	var out []string
	for name := range d.plugins {
		out = append(out, name)
	}
	slices.Sort(out)
	return out, nil
}

// AccessViewPlugins lists the plugins that need grants: every installed one but
// the built-in UI, which reads as the signed-in person.
func (d *fakeDirectory) AccessViewPlugins(ctx context.Context) ([]string, error) {
	names, err := d.ViewPlugins(ctx)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(names, func(name string) bool { return name == pluginUI }), nil
}

// fakeModel is an app.GrantProjector that records what it was told, one line
// per grant as "project|unproject holder target".
type fakeModel struct {
	mu   sync.Mutex
	log  []string
	fail error
}

func (f *fakeModel) ProjectGrants(_ context.Context, grants []identity.Grant) error {
	return f.record("project", grants)
}

func (f *fakeModel) UnprojectGrants(_ context.Context, grants []identity.Grant) error {
	return f.record("unproject", grants)
}

func (f *fakeModel) record(op string, grants []identity.Grant) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail != nil {
		return f.fail
	}
	for i := range grants {
		f.log = append(f.log, fmt.Sprintf("%s %s:%s %s:%s", op, grants[i].Holder.Kind, grants[i].Holder.ID, grants[i].Target.Kind, grants[i].Target.ID))
	}
	return nil
}

func (f *fakeModel) lines() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.log)
}

func (f *fakeModel) failWith(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fail = err
}

func (f *fakeModel) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log = nil
}

// grantFixture is the auth use cases with grants over in-memory stores.
type grantFixture struct {
	*authFixture

	grants *app.Grants
	tokens *app.Tokens
	dir    *fakeDirectory
	model  *fakeModel
}

func newGrantFixture(t *testing.T, demo demoConfig) *grantFixture {
	t.Helper()
	origins, err := identity.NewOrigins([]string{local, devo})
	if err != nil {
		t.Fatalf("NewOrigins: %v", err)
	}
	clk := &clock{t: authStart()}
	store := newMemAuthStore()
	tokens := app.NewTokens(newMemStore(), clk.now, nil)
	dir, model := newFakeDirectory(), &fakeModel{}
	grants := app.NewGrants(store, tokens, dir, model, clk.now)
	cfg := app.AuthConfig{Origins: origins, Sessions: identity.DefaultSessionPolicy(), DemoSessions: demo.on}
	auth, err := app.NewAuth(store, webauthn.New("OpenFief"), tokens, demoMarker(demo.marked), cfg, clk.now, nil, app.WithGrants(grants))
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	return &grantFixture{
		authFixture: &authFixture{auth: auth, store: store, clock: clk, origins: origins},
		grants:      grants, tokens: tokens, dir: dir, model: model,
	}
}

// token issues a script token and returns its ID as a holder.
func (f *grantFixture) token(t *testing.T) identity.GrantHolder {
	t.Helper()
	tok, _, err := f.tokens.Issue(t.Context(), "reader", time.Hour)
	must(t, "Issue", err)
	return identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(tok.ID)}
}

func plugin(name string) identity.GrantHolder {
	return identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: name}
}

func account(id string) identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: id}
}

func sleeve(id string) identity.GrantTarget {
	return identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: id}
}

// grantTrail returns the audit entries of grants as "actor action holder target",
// each side as kind:ID, oldest first.
func (f *grantFixture) grantTrail() []string {
	var out []string
	for _, e := range f.store.entries() {
		if e.Action != identity.ActionGrantCreated && e.Action != identity.ActionGrantRevoked {
			continue
		}
		words := []string{string(e.Actor.Kind), string(e.Action)}
		for _, r := range e.Refs {
			words = append(words, string(r.Kind)+":"+r.ID)
		}
		out = append(out, strings.Join(words, " "))
	}
	return out
}

// created makes the grant as the owner and fails the test when it can't.
func (f *grantFixture) created(t *testing.T, owner *app.NewSession, holder identity.GrantHolder, target identity.GrantTarget) identity.Grant {
	t.Helper()
	g, err := f.auth.CreateGrant(t.Context(), local, owner.Token, holder, target)
	must(t, "CreateGrant", err)
	return g
}

// listed returns the grants m selects as the owner, in force only or with the
// history, and fails the test when it can't.
func (f *grantFixture) listed(t *testing.T, owner *app.NewSession, m identity.GrantMatch, inForceOnly bool) []identity.Grant {
	t.Helper()
	grants, err := f.auth.ListGrants(t.Context(), local, owner.Token, m, inForceOnly)
	must(t, "ListGrants", err)
	return grants
}

// TestCreateGrant checks that the owner grants a script token or a view
// plugin an account, a sleeve or the household, that each grant is an audit
// entry naming holder and target by ID, and that what can't be granted is
// refused with nothing stored or recorded (D-209, D-211).
func TestCreateGrant(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	absent := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: "0123456789abcdef"}
	person := identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: "0123456789abcdef"}
	tests := []struct {
		name   string
		holder identity.GrantHolder
		target identity.GrantTarget
		want   error
	}{
		{"a token an account", tok, account(accountMain), nil},
		{"a token a sleeve", tok, sleeve(sleeveCore), nil},
		{"a plugin the household", plugin(pluginUI), identity.Household(), nil},
		{"an unknown token", absent, identity.Household(), app.ErrNotFound},
		{"a plugin not installed", plugin("gone"), identity.Household(), app.ErrNotFound},
		{"an unknown account", tok, account("nowhere"), app.ErrNotFound},
		{"an unknown sleeve", tok, sleeve("nowhere"), app.ErrNotFound},
		{"a malformed token ID", identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: "short"}, identity.Household(), app.ErrInvalidRequest},
		{"a secret as a target", tok, account(identity.SecretPrefix + "abc"), app.ErrInvalidRequest},
		{"a household that is not home", tok, identity.GrantTarget{Kind: identity.GrantTargetHousehold, ID: "away"}, app.ErrInvalidRequest},
		{"no holder kind", identity.GrantHolder{ID: "x"}, identity.Household(), app.ErrInvalidRequest},
		{"a person, who can't exist yet", person, identity.Household(), app.ErrInvalidRequest},
	}
	for _, tt := range tests {
		g, err := f.auth.CreateGrant(t.Context(), local, owner.Token, tt.holder, tt.target)
		checkCreated(t, tt.name, &g, err, tt.holder, tt.target, tt.want)
	}
	want := []string{
		"owner grant_created script_token:" + tok.ID + " account:" + accountMain,
		"owner grant_created script_token:" + tok.ID + " sleeve:" + sleeveCore,
		"owner grant_created view_plugin:" + pluginUI + " household:home",
	}
	if diff := cmp.Diff(want, f.grantTrail()); diff != "" {
		t.Errorf("grant entries (-want +got):\n%s", diff)
	}
	wantModel := []string{
		"project script_token:" + tok.ID + " account:" + accountMain,
		"project script_token:" + tok.ID + " sleeve:" + sleeveCore,
		"project view_plugin:" + pluginUI + " household:home",
	}
	if diff := cmp.Diff(wantModel, f.model.lines()); diff != "" {
		t.Errorf("passed to the authorization model (-want +got):\n%s", diff)
	}
}

// checkCreated fails the test unless a CreateGrant that answered g and err
// did what want says: made the grant of target to holder when it is nil, and
// otherwise failed with an error wrapping it.
func checkCreated(t *testing.T, name string, g *identity.Grant, err error, holder identity.GrantHolder, target identity.GrantTarget, want error) {
	t.Helper()
	switch {
	case want != nil && !errors.Is(err, want):
		t.Errorf("%s: err = %v, want %v", name, err, want)
	case want == nil && err != nil:
		t.Errorf("%s: %v", name, err)
	case want == nil && (g.Holder != holder || g.Target != target || !g.Active()):
		t.Errorf("%s: grant = %+v", name, g)
	}
}

// TestCreateGrantTwice checks that granting what is granted changes and
// records nothing, and returns the grant in force.
func TestCreateGrantTwice(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	first := f.created(t, &owner, tok, account(accountMain))
	f.clock.set(authStart().Add(30 * time.Minute))
	second := f.created(t, &owner, tok, account(accountMain))
	if !second.CreatedAt.Equal(first.CreatedAt) {
		t.Errorf("the second grant is made at %v, want the first's %v", second.CreatedAt, first.CreatedAt)
	}
	if n := len(f.grantTrail()); n != 1 {
		t.Errorf("%d grant entries, want 1", n)
	}
	if grants := f.listed(t, &owner, identity.GrantMatch{}, false); len(grants) != 1 {
		t.Errorf("%d grants stored, want 1", len(grants))
	}
}

// TestGrantFailsClosedWithoutItsEntry checks that a grant whose audit entry
// can't be written is not made, and is not passed to the authorization model
// (D-211).
func TestGrantFailsClosedWithoutItsEntry(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.store.failAudit(errAuditDown)
	if _, err := f.auth.CreateGrant(t.Context(), local, owner.Token, tok, account(accountMain)); !errors.Is(err, errAuditDown) {
		t.Errorf("CreateGrant with the audit log down = %v, want %v", err, errAuditDown)
	}
	f.store.failAudit(nil)
	if grants := f.listed(t, &owner, identity.GrantMatch{}, false); len(grants) != 0 || len(f.model.lines()) != 0 {
		t.Errorf("grants = %v, passed on = %v, want none", grants, f.model.lines())
	}
}

// TestRevokeFailsClosedWithoutItsEntry checks that a revocation whose audit
// entry can't be written ends nothing in the table.
func TestRevokeFailsClosedWithoutItsEntry(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.created(t, &owner, tok, account(accountMain))
	f.store.failAudit(errAuditDown)
	if err := f.auth.RevokeGrant(t.Context(), local, owner.Token, tok, account(accountMain)); !errors.Is(err, errAuditDown) {
		t.Errorf("RevokeGrant with the audit log down = %v, want %v", err, errAuditDown)
	}
	f.store.failAudit(nil)
	if inForce := f.listed(t, &owner, identity.GrantMatch{}, true); len(inForce) != 1 {
		t.Errorf("%d grants in force after a revocation that could not be recorded, want 1", len(inForce))
	}
}

// TestRevokeGrant checks that a revoked grant stays as history with its end,
// each revocation is an entry, a second revocation changes and records
// nothing, one that never existed is not found, and the relationship is
// removed from the authorization model before the table changes.
func TestRevokeGrant(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	ctx := t.Context()
	f.created(t, &owner, tok, sleeve(sleeveCore))
	f.model.reset()
	f.clock.set(authStart().Add(time.Hour))
	must(t, "RevokeGrant", f.auth.RevokeGrant(ctx, local, owner.Token, tok, sleeve(sleeveCore)))
	must(t, "RevokeGrant again", f.auth.RevokeGrant(ctx, local, owner.Token, tok, sleeve(sleeveCore)))
	if err := f.auth.RevokeGrant(ctx, local, owner.Token, tok, account(accountMain)); !errors.Is(err, app.ErrNotFound) {
		t.Errorf("RevokeGrant of a grant never made = %v, want ErrNotFound", err)
	}
	history := f.listed(t, &owner, identity.GrantMatch{Holder: tok}, false)
	if len(history) != 1 || history[0].Active() || !history[0].RevokedAt.Equal(authStart().Add(time.Hour)) {
		t.Errorf("history = %+v, want the one grant, ended an hour in", history)
	}
	if inForce := f.listed(t, &owner, identity.GrantMatch{Holder: tok}, true); len(inForce) != 0 {
		t.Errorf("%d grants in force, want none", len(inForce))
	}
	want := []string{
		"owner grant_created script_token:" + tok.ID + " sleeve:" + sleeveCore,
		"owner grant_revoked script_token:" + tok.ID + " sleeve:" + sleeveCore,
	}
	if diff := cmp.Diff(want, f.grantTrail()); diff != "" {
		t.Errorf("grant entries (-want +got):\n%s", diff)
	}
	// The second revocation removes the relationship again, which is idempotent and repairs a model that missed it.
	wantModel := []string{"unproject script_token:" + tok.ID + " sleeve:" + sleeveCore, "unproject script_token:" + tok.ID + " sleeve:" + sleeveCore}
	if diff := cmp.Diff(wantModel, f.model.lines()); diff != "" {
		t.Errorf("passed to the authorization model (-want +got):\n%s", diff)
	}
}

// TestRevokeGrantKeepsItWhenTheModelFails checks that a revocation the
// authorization model can't take changes nothing: the grant stays in force and
// no entry is written, so what the owner sees is what SpiceDB answers.
func TestRevokeGrantKeepsItWhenTheModelFails(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.created(t, &owner, tok, identity.Household())
	f.model.failWith(errModelDown)
	if err := f.auth.RevokeGrant(t.Context(), local, owner.Token, tok, identity.Household()); !errors.Is(err, errModelDown) {
		t.Fatalf("RevokeGrant with the model down = %v, want %v", err, errModelDown)
	}
	inForce := f.listed(t, &owner, identity.GrantMatch{}, true)
	if len(inForce) != 1 || len(f.grantTrail()) != 1 {
		t.Errorf("%d grants in force and %d entries, want 1 and 1", len(inForce), len(f.grantTrail()))
	}
}

// TestCreateGrantReportsAModelThatFails checks that a grant the model can't
// take is stored and recorded, and says so, and that granting it again,
// which is idempotent, passes it on.
func TestCreateGrantReportsAModelThatFails(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.model.failWith(errModelDown)
	_, err := f.auth.CreateGrant(t.Context(), local, owner.Token, tok, identity.Household())
	if !errors.Is(err, app.ErrGrantNotProjected) || !errors.Is(err, errModelDown) {
		t.Fatalf("CreateGrant with the model down = %v, want ErrGrantNotProjected wrapping its error", err)
	}
	f.model.failWith(nil)
	f.created(t, &owner, tok, identity.Household())
	if got := f.model.lines(); len(got) != 1 {
		t.Errorf("passed on %v, want the grant once", got)
	}
	if n := len(f.grantTrail()); n != 1 {
		t.Errorf("%d grant entries, want 1", n)
	}
}

// TestGrantsAreOwnerOnly checks that a demo session, an invalid session and
// a script token's secret are refused every grant call (D-209).
func TestGrantsAreOwnerOnly(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{on: true, marked: true})
	_, owner := f.enroll(t)
	tok := f.token(t)
	ctx := t.Context()
	demo, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
	must(t, "demo Login", err)
	f.created(t, &owner, tok, identity.Household())
	for _, tt := range []struct {
		name  string
		token string
		want  error
	}{
		{"a demo session", demo.Token, app.ErrForbidden},
		{"no session", "", app.ErrInvalidSession},
		{"a script token's secret", identity.SecretPrefix + "x", app.ErrInvalidSession},
	} {
		if _, createErr := f.auth.CreateGrant(ctx, local, tt.token, tok, account(accountMain)); !errors.Is(createErr, tt.want) {
			t.Errorf("CreateGrant by %s = %v, want %v", tt.name, createErr, tt.want)
		}
		if revokeErr := f.auth.RevokeGrant(ctx, local, tt.token, tok, identity.Household()); !errors.Is(revokeErr, tt.want) {
			t.Errorf("RevokeGrant by %s = %v, want %v", tt.name, revokeErr, tt.want)
		}
		if _, listErr := f.auth.ListGrants(ctx, local, tt.token, identity.GrantMatch{}, false); !errors.Is(listErr, tt.want) {
			t.Errorf("ListGrants by %s = %v, want %v", tt.name, listErr, tt.want)
		}
	}
	if n := len(f.grantTrail()); n != 1 {
		t.Errorf("%d grant entries, want only the owner's", n)
	}
}

// TestRevokeScriptTokenEndsItsGrants checks that revoking a script token ends
// its grants in the same transaction, each with an entry, and takes them out
// of the authorization model.
func TestRevokeScriptTokenEndsItsGrants(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok, other := f.token(t), f.token(t)
	f.created(t, &owner, tok, identity.Household())
	f.created(t, &owner, other, identity.Household())
	f.created(t, &owner, tok, account(accountMain))
	f.model.reset()
	must(t, "RevokeScriptToken", f.auth.RevokeScriptToken(t.Context(), local, owner.Token, tok.ID))
	inForce := f.listed(t, &owner, identity.GrantMatch{}, true)
	if len(inForce) != 1 || inForce[0].Holder != other {
		t.Errorf("grants in force = %+v, want only the other token's", inForce)
	}
	trail := f.grantTrail()
	if len(trail) != 5 || !strings.HasPrefix(trail[3], "owner grant_revoked script_token:"+tok.ID) || !strings.HasPrefix(trail[4], "owner grant_revoked script_token:"+tok.ID) {
		t.Errorf("grant entries = %v, want the two grants of the revoked token ended by the owner", trail)
	}
	if got := f.model.lines(); len(got) != 2 {
		t.Errorf("removed from the model: %v, want its two grants", got)
	}
}

// TestRevokeScriptTokenReportsAModelThatFails checks that a token whose
// grants the model can't remove is revoked all the same, and says so: it no
// longer authenticates, so nothing it was granted is viewable.
func TestRevokeScriptTokenReportsAModelThatFails(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.created(t, &owner, tok, identity.Household())
	f.model.failWith(errModelDown)
	err := f.auth.RevokeScriptToken(t.Context(), local, owner.Token, tok.ID)
	if !errors.Is(err, app.ErrGrantNotProjected) {
		t.Errorf("revoking a token with the model down = %v, want ErrGrantNotProjected: the token is revoked all the same", err)
	}
	got, getErr := f.tokens.Get(t.Context(), identity.ScriptTokenID(tok.ID))
	if getErr != nil || got.Status(f.clock.now()) != "revoked" {
		t.Errorf("token after the failed projection = %+v, %v, want revoked", got, getErr)
	}
}

// TestRevokeHolderAndTarget checks that the core ends the grants of an
// uninstalled plugin or of a deleted sleeve, as itself, each with an entry,
// and only those of that holder or target.
func TestRevokeHolderAndTarget(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	ctx := t.Context()
	f.created(t, &owner, plugin(pluginUI), identity.Household())
	f.created(t, &owner, plugin(pluginUI), sleeve(sleeveCore))
	f.created(t, &owner, plugin(pluginGhost), identity.Household())
	f.created(t, &owner, tok, sleeve(sleeveCore))
	f.created(t, &owner, tok, account(accountMain))
	if n, err := f.grants.RevokeHolder(ctx, testSystem(), plugin(pluginUI)); err != nil || n != 2 {
		t.Fatalf("RevokeHolder = %d, %v, want 2", n, err)
	}
	if n, err := f.grants.RevokeTarget(ctx, testSystem(), sleeve(sleeveCore)); err != nil || n != 1 {
		t.Fatalf("RevokeTarget = %d, %v, want the token's grant only: the plugin's was ended already", n, err)
	}
	inForce := mustInForce(t, f)
	got := make([]string, 0, len(inForce))
	for _, g := range inForce {
		got = append(got, g.Holder.ID+" "+string(g.Target.Kind))
	}
	if diff := cmp.Diff([]string{pluginGhost + " household", tok.ID + " account"}, got); diff != "" {
		t.Errorf("grants in force (-want +got):\n%s", diff)
	}
	trail := f.grantTrail()
	if len(trail) != 8 || !strings.HasPrefix(trail[5], "system grant_revoked view_plugin:"+pluginUI) || !strings.HasPrefix(trail[7], "system grant_revoked script_token:") {
		t.Errorf("grant entries = %v, want the core's three revocations after the five grants", trail)
	}
}

// TestCutover checks the step D-209 describes for the day SpiceDB takes
// over: every script token that may be used and every view plugin with no
// grant gets exactly the household grant, each with its own entry by the core;
// a revoked token, an expired one, and a holder with a grant, ended or not,
// get none; and the next call, however many tokens were made since, does
// nothing.
func TestCutover(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	ctx := t.Context()
	expired, revoked := f.token(t), f.token(t) // every token lasts an hour
	must(t, "Revoke", f.tokens.Revoke(ctx, identity.ScriptTokenID(revoked.ID)))
	f.clock.set(authStart().Add(2 * time.Hour))
	active, granted := f.token(t), f.token(t)
	f.created(t, &owner, granted, account(accountMain))
	f.created(t, &owner, plugin(pluginGhost), sleeve(sleeveCore))
	must(t, "RevokeGrant", f.auth.RevokeGrant(ctx, local, owner.Token, plugin(pluginGhost), sleeve(sleeveCore)))
	f.model.reset()
	n, err := f.grants.Cutover(ctx, testSystem())
	if err != nil || n != 2 {
		t.Fatalf("Cutover = %d, %v, want 2 grants: the token that may be used, and the ui plugin", n, err)
	}
	f.checkHousehold(t, []identity.GrantHolder{active, plugin(pluginUI)}, []identity.GrantHolder{expired, revoked, granted, plugin(pluginGhost)})
	if got := len(f.model.lines()); got != 2 {
		t.Errorf("passed %d grants to the model, want 2", got)
	}
	var byTheCore int
	for _, line := range f.grantTrail() {
		if strings.HasPrefix(line, "system grant_created ") && strings.HasSuffix(line, " household:home") {
			byTheCore++
		}
	}
	if byTheCore != 2 {
		t.Errorf("%d household entries by the core, want one for each of the 2 grants", byTheCore)
	}
}

// checkHousehold fails the test unless each of with holds the household grant
// in force and none of without does.
func (f *grantFixture) checkHousehold(t *testing.T, with, without []identity.GrantHolder) {
	t.Helper()
	have := map[identity.GrantHolder]bool{}
	for _, g := range mustInForce(t, f) {
		if g.Target == identity.Household() {
			have[g.Holder] = true
		}
	}
	for _, h := range with {
		if !have[h] {
			t.Errorf("%v has no household grant", h)
		}
	}
	for _, h := range without {
		if have[h] {
			t.Errorf("%v has a household grant, but it was expired, revoked or had a grant already", h)
		}
	}
}

// TestCutoverRunsOnce checks that a second cutover, after a token was made
// that has no grant, gives it none.
func TestCutoverRunsOnce(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	ctx := t.Context()
	if n, err := f.grants.Cutover(ctx, testSystem()); err != nil || n != 2 {
		t.Fatalf("Cutover = %d, %v, want 2: both plugins", n, err)
	}
	after := f.token(t)
	if n, err := f.grants.Cutover(ctx, testSystem()); err != nil || n != 0 {
		t.Errorf("a second Cutover = %d, %v, want 0", n, err)
	}
	f.checkHousehold(t, []identity.GrantHolder{plugin(pluginUI), plugin(pluginGhost)}, []identity.GrantHolder{after})
}

func mustInForce(t *testing.T, f *grantFixture) []identity.Grant {
	t.Helper()
	g, err := f.grants.InForce(t.Context())
	must(t, "InForce", err)
	return g
}

// TestCutoverWaitsForThePlugins checks that a cutover that can't list the
// installed plugins does nothing and is not recorded as done, so a later try
// gives every holder its grant.
func TestCutoverWaitsForThePlugins(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	ctx := t.Context()
	f.token(t)
	f.dir.unknown = true
	if n, err := f.grants.Cutover(ctx, testSystem()); !errors.Is(err, app.ErrDirectoryUnknown) || n != 0 {
		t.Fatalf("Cutover with the plugins unknown = %d, %v, want ErrDirectoryUnknown", n, err)
	}
	if got := mustInForce(t, f); len(got) != 0 {
		t.Fatalf("grants made without knowing the plugins: %v", got)
	}
	f.dir.unknown = false
	if n, err := f.grants.Cutover(ctx, testSystem()); err != nil || n != 3 {
		t.Errorf("Cutover once they are known = %d, %v, want 3 (the token and both plugins)", n, err)
	}
}

// TestCutoverReportsAModelThatFails checks that a cutover whose grants the
// model can't take has made them, and says so, for the core to reconcile.
func TestCutoverReportsAModelThatFails(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	f.token(t)
	f.model.failWith(errModelDown)
	n, err := f.grants.Cutover(t.Context(), testSystem())
	if !errors.Is(err, app.ErrGrantNotProjected) || n != 3 {
		t.Errorf("Cutover with the model down = %d, %v, want 3 and ErrGrantNotProjected", n, err)
	}
	if got := mustInForce(t, f); len(got) != 3 {
		t.Errorf("%d grants stored, want 3", len(got))
	}
}

// TestSweep checks that the core ends the grants that outlived their holder
// or their target, each with an entry, and leaves what it can't tell about.
func TestSweep(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	ctx := t.Context()
	revoked, live := f.token(t), f.token(t)
	f.created(t, &owner, revoked, identity.Household())
	f.created(t, &owner, live, identity.Household())
	f.created(t, &owner, live, account(accountMain))
	f.created(t, &owner, live, sleeve(sleeveCore))
	f.created(t, &owner, plugin(pluginUI), identity.Household())
	f.created(t, &owner, plugin(pluginGhost), account(accountOther))
	// The token is revoked, the plugin uninstalled and the account and the sleeve deleted behind the grants' back.
	must(t, "Revoke", f.tokens.Revoke(ctx, identity.ScriptTokenID(revoked.ID)))
	delete(f.dir.plugins, pluginGhost)
	delete(f.dir.accounts, accountMain)
	delete(f.dir.sleeves, sleeveCore)

	f.dir.unknown = true // the plugins can't be read: nothing of a plugin is swept
	n, err := f.grants.Sweep(ctx, testSystem())
	must(t, "Sweep", err)
	if n != 3 {
		t.Errorf("Sweep with the plugins unknown ended %d grants, want 3 (the revoked token's, the account's and the sleeve's)", n)
	}
	f.dir.unknown = false
	if n, err = f.grants.Sweep(ctx, testSystem()); err != nil || n != 1 {
		t.Errorf("Sweep = %d, %v, want 1 (the plugin that is gone)", n, err)
	}
	inForce := mustInForce(t, f)
	left := make([]string, 0, len(inForce))
	for _, g := range inForce {
		left = append(left, g.Holder.ID+" "+string(g.Target.Kind))
	}
	slices.Sort(left)
	want := []string{live.ID + " household", pluginUI + " household"}
	slices.Sort(want)
	if diff := cmp.Diff(want, left); diff != "" {
		t.Errorf("grants left (-want +got):\n%s", diff)
	}
	if n, err := f.grants.Sweep(ctx, testSystem()); err != nil || n != 0 {
		t.Errorf("a second Sweep = %d, %v, want 0", n, err)
	}
}
