package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// errNoAnswer is what the fake model's ping says when it is down.
var errNoAnswer = errors.New("no answer")

// TestCreateScriptTokenWithGrants checks that the owner creates a token with
// the MCP flag and the grants chosen (a repeated one made once), that the token
// lists them with their flag, that each grant is an audit entry and a
// relationship in the model, and that the secret is not in any of it (D-210,
// D-211).
func TestCreateScriptTokenWithGrants(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	spec := app.NewToken{
		Name: "nightly", TTL: time.Hour, MCPAllowed: true,
		Targets: []identity.GrantTarget{sleeve(sleeveCore), account(accountMain), sleeve(sleeveCore)},
	}
	tok, secret, err := f.auth.CreateScriptToken(t.Context(), local, owner.Token, spec)
	must(t, "CreateScriptToken", err)
	if secret == "" || !tok.MCPAllowed {
		t.Errorf("token = %+v, secret empty = %v, want a secret and the MCP flag", tok, secret == "")
	}
	listed, err := f.auth.ScriptTokens(t.Context(), local, owner.Token)
	must(t, "ScriptTokens", err)
	if len(listed) != 1 || !listed[0].Token.MCPAllowed || len(listed[0].Grants) != 2 {
		t.Fatalf("ScriptTokens = %+v, want the token with its flag and two grants", listed)
	}
	if diff := cmp.Diff([]string{"owner script_token_created script_token:" + string(tok.ID)}, f.tokenTrail(identity.ActionScriptTokenCreated)); diff != "" {
		t.Errorf("token entries (-want +got):\n%s", diff)
	}
	wantGrants := []string{
		"owner grant_created script_token:" + string(tok.ID) + " sleeve:" + sleeveCore,
		"owner grant_created script_token:" + string(tok.ID) + " account:" + accountMain,
	}
	if diff := cmp.Diff(wantGrants, f.grantTrail()); diff != "" {
		t.Errorf("grant entries (-want +got):\n%s", diff)
	}
	if got := f.model.lines(); len(got) != 2 {
		t.Errorf("model = %v, want the two grants", got)
	}
}

// tokenTrail returns the audit entries of action on a script token as "actor
// action ref", oldest first.
func (f *grantFixture) tokenTrail(action identity.AuditAction) []string {
	var out []string
	for _, e := range f.store.entries() {
		if e.Action == action {
			out = append(out, string(e.Actor.Kind)+" "+string(e.Action)+" "+string(e.Refs[0].Kind)+":"+e.Refs[0].ID)
		}
	}
	return out
}

// TestCreateScriptTokenGrantRefusals checks that a token with a target that
// isn't there or isn't valid is refused before anything is stored.
func TestCreateScriptTokenGrantRefusals(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	ctx := t.Context()
	for name, tt := range map[string]struct {
		target identity.GrantTarget
		want   error
	}{
		"an unknown sleeve":  {sleeve("nowhere"), app.ErrNotFound},
		"an unknown account": {account("nowhere"), app.ErrNotFound},
		"a malformed target": {identity.GrantTarget{Kind: identity.GrantTargetHousehold, ID: "away"}, app.ErrInvalidRequest},
	} {
		_, secret, err := f.auth.CreateScriptToken(ctx, local, owner.Token, app.NewToken{Name: "x", Targets: []identity.GrantTarget{tt.target}})
		if !errors.Is(err, tt.want) || secret != "" {
			t.Errorf("%s: err = %v, secret %q, want %v and no secret", name, err, secret, tt.want)
		}
	}
	listed, err := f.auth.ScriptTokens(ctx, local, owner.Token)
	must(t, "ScriptTokens", err)
	if len(listed) != 0 || len(f.grantTrail()) != 0 {
		t.Errorf("tokens = %+v, grant entries = %v, want nothing stored", listed, f.grantTrail())
	}
}

// TestCreateScriptTokenTakenBackWhenTheModelFails checks that a model whose
// write fails after the token and its grants committed takes the token back,
// with its own audit entry and no secret, and that a token with no grants needs
// nothing of the model.
func TestCreateScriptTokenTakenBackWhenTheModelFails(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	ctx := t.Context()
	f.model.failWith(errModelDown)
	_, secret, err := f.auth.CreateScriptToken(ctx, local, owner.Token, app.NewToken{Name: "x", Targets: []identity.GrantTarget{identity.Household()}})
	if !errors.Is(err, app.ErrTokenRevoked) || !errors.Is(err, app.ErrGrantNotProjected) || secret != "" {
		t.Fatalf("CreateScriptToken with the model failing = %v, secret %q, want ErrTokenRevoked and no secret", err, secret)
	}
	listed, listErr := f.auth.ScriptTokens(ctx, local, owner.Token)
	must(t, "ScriptTokens", listErr)
	if len(listed) != 1 || listed[0].Token.Status(f.clock.now()) != "revoked" || len(listed[0].Grants) != 0 {
		t.Errorf("tokens = %+v, want the one token revoked with no grant in force", listed)
	}
	if n := len(f.tokenTrail(identity.ActionScriptTokenRevoked)); n != 1 {
		t.Errorf("%d revocation entries, want 1", n)
	}

	// A token with no grants asks nothing of the model.
	_, secret, err = f.auth.CreateScriptToken(ctx, local, owner.Token, app.NewToken{Name: "plain"})
	if err != nil || secret == "" {
		t.Errorf("CreateScriptToken without grants = %v, secret empty %v, want it made while the model fails", err, secret == "")
	}
}

// TestCreateScriptTokenAsksTheModelFirst checks that a token that is to hold
// grants is not stored when the model doesn't answer a probe, which leaves no
// token to revoke.
func TestCreateScriptTokenAsksTheModelFirst(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	f.grants.WithPing(func(context.Context) error { return errNoAnswer })
	entries := len(f.store.entries())
	_, secret, err := f.auth.CreateScriptToken(t.Context(), local, owner.Token, app.NewToken{Name: "x", Targets: []identity.GrantTarget{identity.Household()}})
	if !errors.Is(err, app.ErrGrantNotProjected) || !errors.Is(err, errNoAnswer) || secret != "" {
		t.Fatalf("CreateScriptToken = %v, secret %q, want ErrGrantNotProjected and no secret", err, secret)
	}
	listed, listErr := f.auth.ScriptTokens(t.Context(), local, owner.Token)
	must(t, "ScriptTokens", listErr)
	if len(listed) != 0 || len(f.store.entries()) != entries {
		t.Errorf("tokens = %+v, entries = %d, want no token and no new entry (%d)", listed, len(f.store.entries()), entries)
	}
}

// TestUpdateGrants checks that one call adds and removes grants of a token or
// a view plugin with an entry for each change, returns what is held after, and
// changes nothing when asked to add what is held or remove what is not.
func TestUpdateGrants(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.created(t, &owner, tok, sleeve(sleeveCore))
	f.model.reset()
	ctx := t.Context()

	got, err := f.auth.UpdateGrants(ctx, local, owner.Token, tok, []identity.GrantTarget{account(accountMain), sleeve(sleeveCore)}, []identity.GrantTarget{identity.Household(), sleeve(sleeveCore)})
	must(t, "UpdateGrants", err)
	held := make([]identity.GrantTarget, 0, len(got))
	for i := range got {
		held = append(held, got[i].Target)
	}
	if diff := cmp.Diff([]identity.GrantTarget{account(accountMain), sleeve(sleeveCore)}, held); diff != "" {
		t.Errorf("held (-want +got):\n%s", diff)
	}
	wantTrail := []string{
		"owner grant_created script_token:" + tok.ID + " sleeve:" + sleeveCore,
		"owner grant_revoked script_token:" + tok.ID + " sleeve:" + sleeveCore,
		"owner grant_created script_token:" + tok.ID + " account:" + accountMain,
		"owner grant_created script_token:" + tok.ID + " sleeve:" + sleeveCore,
	}
	if diff := cmp.Diff(wantTrail, f.grantTrail()); diff != "" {
		t.Errorf("grant entries (-want +got):\n%s", diff)
	}

	before := len(f.grantTrail())
	_, err = f.auth.UpdateGrants(ctx, local, owner.Token, plugin(pluginGhost), []identity.GrantTarget{identity.Household()}, []identity.GrantTarget{account(accountOther)})
	must(t, "UpdateGrants of a plugin", err)
	if n := len(f.grantTrail()); n != before+1 {
		t.Errorf("%d entries, want one more for the plugin's household grant (the removal of a grant it never held is no entry)", n)
	}
	_, err = f.auth.UpdateGrants(ctx, local, owner.Token, plugin(pluginGhost), []identity.GrantTarget{identity.Household()}, nil)
	must(t, "UpdateGrants again", err)
	if n := len(f.grantTrail()); n != before+1 {
		t.Errorf("%d entries, want adding what is held to change nothing", n)
	}
}

// TestUpdateGrantsChecksFirst checks that a holder or target that isn't there,
// or isn't valid, fails the whole call before any change, and that a demo
// session, no session and a viewer holder are refused.
func TestUpdateGrantsChecksFirst(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{on: true, marked: true})
	_, owner := f.enroll(t)
	tok := f.token(t)
	f.created(t, &owner, tok, sleeve(sleeveCore))
	ctx := t.Context()
	demo, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
	must(t, "demo Login", err)
	absent := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: "0123456789abcdef"}
	person := identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: "0123456789abcdef"}
	for _, tt := range []struct {
		name   string
		token  string
		holder identity.GrantHolder
		add    []identity.GrantTarget
		remove []identity.GrantTarget
		want   error
	}{
		{"an unknown sleeve to add", owner.Token, tok, []identity.GrantTarget{account(accountMain), sleeve("nowhere")}, []identity.GrantTarget{sleeve(sleeveCore)}, app.ErrNotFound},
		{"an unknown holder", owner.Token, absent, []identity.GrantTarget{identity.Household()}, nil, app.ErrNotFound},
		{"a plugin not installed", owner.Token, plugin("gone"), []identity.GrantTarget{identity.Household()}, nil, app.ErrNotFound},
		{"a malformed target to remove", owner.Token, tok, nil, []identity.GrantTarget{{Kind: identity.GrantTargetHousehold, ID: "away"}}, app.ErrInvalidRequest},
		{"a viewer that does not exist", owner.Token, person, []identity.GrantTarget{identity.Household()}, nil, app.ErrNotFound},
		{"a demo session", demo.Token, tok, []identity.GrantTarget{identity.Household()}, nil, app.ErrForbidden},
		{"no session", "", tok, []identity.GrantTarget{identity.Household()}, nil, app.ErrInvalidSession},
	} {
		if _, updateErr := f.auth.UpdateGrants(ctx, local, tt.token, tt.holder, tt.add, tt.remove); !errors.Is(updateErr, tt.want) {
			t.Errorf("UpdateGrants with %s = %v, want %v", tt.name, updateErr, tt.want)
		}
	}
	if n := len(f.grantTrail()); n != 1 {
		t.Errorf("%d grant entries, want only the first grant's: no refused call changed anything", n)
	}
}

// TestViewPluginsAreListedForTheOwner checks that the Access page's view
// plugins are those that read with their own identity (not the built-in UI),
// each with its grants in force, for the owner only, and that an installed
// plugins list the core can't read is an error naming that.
func TestViewPluginsAreListedForTheOwner(t *testing.T) {
	t.Parallel()
	f := newGrantFixture(t, demoConfig{on: true, marked: true})
	_, owner := f.enroll(t)
	f.created(t, &owner, plugin(pluginGhost), account(accountMain))
	f.created(t, &owner, plugin(pluginGhost), sleeve(sleeveCore))
	ctx := t.Context()
	got, err := f.auth.ViewPlugins(ctx, local, owner.Token)
	must(t, "ViewPlugins", err)
	if len(got) != 1 || got[0].Instance != pluginGhost || len(got[0].Grants) != 2 {
		t.Errorf("ViewPlugins = %+v, want ghostfolio with its two grants and not the built-in ui", got)
	}
	demo, err := f.auth.Login(ctx, f.auth.Demo(), local, app.Proof{})
	must(t, "demo Login", err)
	if _, demoErr := f.auth.ViewPlugins(ctx, local, demo.Token); !errors.Is(demoErr, app.ErrForbidden) {
		t.Errorf("ViewPlugins by a demo session = %v, want ErrForbidden", demoErr)
	}
	f.dir.mu.Lock()
	f.dir.unknown = true
	f.dir.mu.Unlock()
	if _, unknownErr := f.auth.ViewPlugins(ctx, local, owner.Token); !errors.Is(unknownErr, app.ErrDirectoryUnknown) {
		t.Errorf("ViewPlugins with the plugins unreadable = %v, want ErrDirectoryUnknown", unknownErr)
	}
}
