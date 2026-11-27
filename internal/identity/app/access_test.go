package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

func TestPolicyMayAct(t *testing.T) {
	t.Parallel()
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s1"}
	token := identity.Caller{Holder: identity.HolderScriptToken, ID: "t1"}
	tests := []struct {
		name    string
		caller  identity.Caller
		action  identity.Action
		account string
		want    bool
	}{
		{name: "the owner views any account", caller: owner, action: identity.ActionView, account: "brokerage", want: true},
		{name: "the owner changes any account", caller: owner, action: identity.ActionChange, account: "bonds", want: true},
		{name: "a script token views any account", caller: token, action: identity.ActionView, account: "bonds", want: true},
		{name: "a script token changes none", caller: token, action: identity.ActionChange, account: "bonds"},
		{name: "a view plugin views", caller: identity.Caller{Holder: identity.HolderViewPlugin, ID: "ui"}, action: identity.ActionView, account: "brokerage", want: true},
		{name: "a demo session changes none", caller: identity.Caller{Holder: identity.HolderDemoSession, ID: "d1"}, action: identity.ActionChange, account: "brokerage"},
		{name: "nobody views nothing", caller: identity.Caller{}, action: identity.ActionView, account: "brokerage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := app.Policy{}.MayAct(t.Context(), tt.caller, tt.action, tt.account)
			if err != nil {
				t.Fatalf("MayAct: %v", err)
			}
			if got != tt.want {
				t.Errorf("MayAct(%+v, %s, %s) = %v, want %v", tt.caller, tt.action, tt.account, got, tt.want)
			}
		})
	}
}

func TestCallerInContext(t *testing.T) {
	t.Parallel()
	if got := app.CallerFrom(t.Context()); got != (identity.Caller{}) {
		t.Errorf("CallerFrom(no caller) = %+v, want nobody", got)
	}
	want := identity.Caller{Holder: identity.HolderViewPlugin, ID: "ui"}
	if diff := cmp.Diff(want, app.CallerFrom(app.WithCaller(t.Context(), want))); diff != "" {
		t.Errorf("CallerFrom(WithCaller) mismatch (-want +got):\n%s", diff)
	}
}

func TestPrincipalCaller(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		p    app.Principal
		want identity.Caller
	}{
		{name: "an owner's session is the owner", p: app.Principal{Session: "s1", Owner: true}, want: identity.Caller{Holder: identity.HolderOwner, ID: "s1"}},
		{name: "a session of nobody in particular is nobody", p: app.Principal{Session: "s1"}, want: identity.Caller{}},
		{name: "a viewer's session is the viewer, whatever else it says", p: app.Principal{Session: "v1", Viewer: "p1", Owner: true}, want: identity.Caller{Holder: identity.HolderViewer, ID: "p1"}},
		{name: "a demo session", p: app.Principal{Session: "d1", Demo: true}, want: identity.Caller{Holder: identity.HolderDemoSession, ID: "d1"}},
		{name: "a script token", p: app.Principal{ScriptToken: "t1"}, want: identity.Caller{Holder: identity.HolderScriptToken, ID: "t1"}},
		{name: "nothing is nobody", p: app.Principal{}, want: identity.Caller{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.p.Caller()); diff != "" {
				t.Errorf("Caller mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPrincipalOfSessionNeverMakesAViewerTheOwner checks every kind of
// session through the one constructor: only the owner's own is the owner, a
// viewer's is the viewer in every scope, and none but the owner's is let
// through to what the owner may do.
func TestPrincipalOfSessionNeverMakesAViewerTheOwner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    identity.Session
		want identity.Caller
	}{
		{name: "the owner's session", s: identity.Session{ID: "s1", Scope: identity.SessionScopeOwner}, want: identity.Caller{Holder: identity.HolderOwner, ID: "s1"}},
		{name: "the owner's read session", s: identity.Session{ID: "s2", Scope: identity.SessionScopeRead}, want: identity.Caller{Holder: identity.HolderOwner, ID: "s2"}},
		{name: "a viewer's read session", s: identity.Session{ID: "s3", Scope: identity.SessionScopeRead, PersonID: "p1"}, want: identity.Caller{Holder: identity.HolderViewer, ID: "p1"}},
		{name: "a viewer's session with owner scope", s: identity.Session{ID: "s4", Scope: identity.SessionScopeOwner, PersonID: "p1"}, want: identity.Caller{Holder: identity.HolderViewer, ID: "p1"}},
		{name: "a viewer's session from the demo", s: identity.Session{ID: "s5", Scope: identity.SessionScopeRead, PersonID: "p1", Authenticator: identity.AuthenticatorDemo}, want: identity.Caller{Holder: identity.HolderViewer, ID: "p1"}},
		{name: "a demo session", s: identity.Session{ID: "s6", Scope: identity.SessionScopeRead, Authenticator: identity.AuthenticatorDemo}, want: identity.Caller{Holder: identity.HolderDemoSession, ID: "s6"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := app.PrincipalOfSession(&tt.s).Caller()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Caller mismatch (-want +got):\n%s", diff)
			}
			if tt.want.Holder == identity.HolderViewer && identity.MayAct(got, identity.ActionView) {
				t.Errorf("a viewer may view every account by the rule that holds until SpiceDB answers")
			}
		})
	}
}
