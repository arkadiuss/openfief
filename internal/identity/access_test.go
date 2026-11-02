package identity_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/identity"
)

func TestMayAct(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		caller     identity.Caller
		view, edit bool
	}{
		{name: "the owner views and changes", caller: identity.Caller{Holder: identity.HolderOwner, ID: "s1"}, view: true, edit: true},
		{name: "a script token only views", caller: identity.Caller{Holder: identity.HolderScriptToken, ID: "t1"}, view: true},
		{name: "a view plugin only views", caller: identity.Caller{Holder: identity.HolderViewPlugin, ID: "ui"}, view: true},
		{name: "a demo session only views", caller: identity.Caller{Holder: identity.HolderDemoSession, ID: "d1"}, view: true},
		{name: "nobody does nothing", caller: identity.Caller{}},
		{name: "an unknown holder does nothing", caller: identity.Caller{Holder: identity.Holder(99), ID: "x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := identity.MayAct(tt.caller, identity.ActionView); got != tt.view {
				t.Errorf("MayAct(%v, view) = %v, want %v", tt.caller, got, tt.view)
			}
			if got := identity.MayAct(tt.caller, identity.ActionChange); got != tt.edit {
				t.Errorf("MayAct(%v, change) = %v, want %v", tt.caller, got, tt.edit)
			}
			if got := identity.MayAct(tt.caller, identity.Action(0)); got {
				t.Errorf("MayAct(%v, no action) = true, want false", tt.caller)
			}
		})
	}
}

func TestActionAndHolderNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		got, want string
	}{
		{identity.ActionView.String(), "view"},
		{identity.ActionChange.String(), "change"},
		{identity.Action(7).String(), "Action(7)"},
		{identity.HolderOwner.String(), "owner"},
		{identity.HolderScriptToken.String(), "script_token"},
		{identity.HolderViewPlugin.String(), "view_plugin"},
		{identity.HolderDemoSession.String(), "demo_session"},
		{identity.Holder(0).String(), "Holder(0)"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("name = %q, want %q", tt.got, tt.want)
		}
	}
}
