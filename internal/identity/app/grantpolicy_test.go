package app_test

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// fakeGrantReader returns fixed grants of each holder.
type fakeGrantReader struct {
	grants []identity.Grant
	err    error
}

func (f fakeGrantReader) Grants(_ context.Context, m identity.GrantMatch, _ bool) ([]identity.Grant, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []identity.Grant
	for i := range f.grants {
		if m.Matches(&f.grants[i]) {
			out = append(out, f.grants[i])
		}
	}
	return out, nil
}

// fakeCover covers the household with three accounts, two of them in the
// sleeve "core".
type fakeCover struct{ err error }

func (f fakeCover) TargetAccounts(_ context.Context, t identity.GrantTarget) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	switch t.Kind {
	case identity.GrantTargetAccount:
		return []string{t.ID}, nil
	case identity.GrantTargetSleeve:
		if t.ID == "core" {
			return []string{"a1", "a2"}, nil
		}
		return nil, nil
	case identity.GrantTargetHousehold:
		return []string{"a1", "a2", "a3"}, nil
	}
	return nil, nil
}

// listCover covers the household with the accounts it holds now, so a test can
// add one after a session started.
type listCover struct{ accounts *[]string }

func (c listCover) TargetAccounts(_ context.Context, t identity.GrantTarget) ([]string, error) {
	if t.Kind == identity.GrantTargetHousehold {
		return slices.Clone(*c.accounts), nil
	}
	return nil, nil
}

// TestGrantPolicy checks that under the policy a viewer, a script token and a
// view plugin view exactly the accounts their grants cover and change none,
// while the owner is answered as identity.MayAct answers, and that errors
// refuse (D-65, D-209).
func TestGrantPolicy(t *testing.T) {
	t.Parallel()
	const alice, bob, tokenOne, tokenHousehold, tokenNone = "0123456789abcdef", "fedcba9876543210", "1111111111111111", "2222222222222222", "3333333333333333"
	holder := func(id string) identity.GrantHolder {
		return identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: id}
	}
	grants := []identity.Grant{
		{Holder: holder(alice), Target: sleeve("core")},
		{Holder: holder(alice), Target: account("a3")},
		{Holder: holder(bob), Target: identity.Household()},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: tokenOne}, Target: account("a1")},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: tokenHousehold}, Target: identity.Household()},
		{Holder: identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: "ghostfolio"}, Target: sleeve("core")},
	}
	token := func(id string) identity.Caller { return identity.Caller{Holder: identity.HolderScriptToken, ID: id} }
	p := app.NewGrantPolicy(fakeGrantReader{grants: grants}, fakeCover{})
	viewer := func(id string) identity.Caller { return identity.Caller{Holder: identity.HolderViewer, ID: id} }
	tests := []struct {
		name    string
		caller  identity.Caller
		action  identity.Action
		account string
		want    bool
	}{
		{"a viewer's sleeve grant", viewer(alice), identity.ActionView, "a1", true},
		{"a viewer's account grant", viewer(alice), identity.ActionView, "a3", true},
		{"a viewer's household grant", viewer(bob), identity.ActionView, "a2", true},
		{"an account no grant covers", viewer(alice), identity.ActionView, "a9", false},
		{"a viewer changes nothing", viewer(bob), identity.ActionChange, "a1", false},
		{"a viewer with no grants", viewer("0000000000000000"), identity.ActionView, "a1", false},
		{"a viewer with no ID", viewer(""), identity.ActionView, "a1", false},
		{"the owner views", identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionView, "a9", true},
		{"the owner changes", identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionChange, "a9", true},
		{"a token granted one account views it", token(tokenOne), identity.ActionView, "a1", true},
		{"a token granted one account views no other", token(tokenOne), identity.ActionView, "a2", false},
		{"a token with the household grant views every account", token(tokenHousehold), identity.ActionView, "a3", true},
		{"a token with no grant views none", token(tokenNone), identity.ActionView, "a1", false},
		{"a token changes none", token(tokenHousehold), identity.ActionChange, "a1", false},
		{"a token with no ID", token(""), identity.ActionView, "a1", false},
		{"a view plugin views its sleeve", identity.Caller{Holder: identity.HolderViewPlugin, ID: "ghostfolio"}, identity.ActionView, "a2", true},
		{"a view plugin views nothing outside its grant", identity.Caller{Holder: identity.HolderViewPlugin, ID: "ghostfolio"}, identity.ActionView, "a3", false},
		{"a view plugin with no grant", identity.Caller{Holder: identity.HolderViewPlugin, ID: "other"}, identity.ActionView, "a1", false},
		{"a demo session views as before", identity.Caller{Holder: identity.HolderDemoSession, ID: "d"}, identity.ActionView, "a3", true},
		{"nobody", identity.Caller{}, identity.ActionView, "a1", false},
	}
	for _, tt := range tests {
		got, err := p.MayAct(t.Context(), tt.caller, tt.action, tt.account)
		if err != nil || got != tt.want {
			t.Errorf("MayAct(%s) = %v, %v, want %v", tt.name, got, err, tt.want)
		}
		// The list gives exactly what MayAct gives each account.
		set, err := p.ActableAccounts(t.Context(), tt.caller, tt.action)
		if err != nil {
			t.Fatalf("ActableAccounts(%s): %v", tt.name, err)
		}
		_, listed := set[tt.account]
		_, known := map[string]bool{"a1": true, "a2": true, "a3": true}[tt.account]
		if listed != (tt.want && known) {
			t.Errorf("ActableAccounts(%s) lists %q = %v, want %v", tt.name, tt.account, listed, tt.want && known)
		}
	}
	set, err := p.ActableAccounts(t.Context(), viewer(alice), identity.ActionView)
	if err != nil || !slices.Equal(slices.Sorted(maps.Keys(set)), []string{"a1", "a2", "a3"}) {
		t.Errorf("ActableAccounts(alice) = %v, %v, want a1, a2, a3", set, err)
	}

	checkDemoHoldsHousehold(t)
}

// errDown is what the fakes say when they can't be read.
var errDown = errors.New("down")

// TestGrantPolicyFailsClosed checks that a viewer is refused, with the error,
// when their grants or the accounts they cover can't be read.
func TestGrantPolicyFailsClosed(t *testing.T) {
	t.Parallel()
	caller := identity.Caller{Holder: identity.HolderViewer, ID: "0123456789abcdef"}
	grants := []identity.Grant{{Holder: identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: caller.ID}, Target: identity.Household()}}
	for name, p := range map[string]*app.GrantPolicy{
		"grants down":   app.NewGrantPolicy(fakeGrantReader{err: errDown}, fakeCover{}),
		"accounts down": app.NewGrantPolicy(fakeGrantReader{grants: grants}, fakeCover{err: errDown}),
	} {
		if ok, err := p.MayAct(t.Context(), caller, identity.ActionView, "a1"); ok || !errors.Is(err, errDown) {
			t.Errorf("MayAct with %s = %v, %v, want a refusal with the error", name, ok, err)
		}
		if set, err := p.ActableAccounts(t.Context(), caller, identity.ActionView); len(set) != 0 || !errors.Is(err, errDown) {
			t.Errorf("ActableAccounts with %s = %v, %v, want none with the error", name, set, err)
		}
	}
}

// checkDemoHoldsHousehold checks that a demo session holds the household grant
// (D-163, D-209): it views every account of the household, one that exists only
// after its session started too, and changes none.
func checkDemoHoldsHousehold(t *testing.T) {
	t.Helper()
	accounts := make([]string, 0, 3)
	accounts = append(accounts, "a1", "a2")
	live := app.NewGrantPolicy(fakeGrantReader{}, listCover{accounts: &accounts})
	demo := identity.Caller{Holder: identity.HolderDemoSession, ID: "d"}
	checkDemoAccounts(t, live, demo, accounts, "at the start")
	accounts = append(accounts, "a3")
	checkDemoAccounts(t, live, demo, accounts, "after an account was added")
}

// checkDemoAccounts checks that demo views each of accounts and changes none.
func checkDemoAccounts(t *testing.T, p *app.GrantPolicy, demo identity.Caller, accounts []string, step string) {
	t.Helper()
	for _, id := range accounts {
		if ok, err := p.MayAct(t.Context(), demo, identity.ActionView, id); err != nil || !ok {
			t.Errorf("a demo session views %q %s = %v, %v, want true", id, step, ok, err)
		}
		if ok, err := p.MayAct(t.Context(), demo, identity.ActionChange, id); err != nil || ok {
			t.Errorf("a demo session changes %q %s = %v, %v, want false", id, step, ok, err)
		}
	}
	set, err := p.ActableAccounts(t.Context(), demo, identity.ActionView)
	if err != nil || !slices.Equal(slices.Sorted(maps.Keys(set)), accounts) {
		t.Errorf("ActableAccounts(demo) %s = %v, %v, want %v", step, set, err, accounts)
	}
	if set, err := p.ActableAccounts(t.Context(), demo, identity.ActionChange); err != nil || len(set) != 0 {
		t.Errorf("ActableAccounts(demo, change) %s = %v, %v, want none", step, set, err)
	}
}
