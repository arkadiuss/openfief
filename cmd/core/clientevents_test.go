package main

import (
	"testing"

	"connectrpc.com/connect"

	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
)

// clientEvents reads the switch for the UI's client events as the view plugin does,
// with no session.
func clientEvents(t *testing.T, oc ownerCore, token string) (bool, error) {
	t.Helper()
	resp, err := settingsClient(oc.srv, token).GetClientEvents(t.Context(), connect.NewRequest(&valuationv1.GetClientEventsRequest{}))
	if err != nil {
		return false, err //nolint:wrapcheck // the test reads the code of the core's own error
	}
	return resp.Msg.GetEnabled(), nil
}

// TestClientEventsSettingOnARealCore runs the owner's switch for the UI's client
// events (D-227) through the real core and its database: on until the owner turns
// it off, read by the view plugin with no session and by no other caller, changed
// by the owner's session alone and stored, not remembered by the plugin. That the
// change is entered in the audit log, once, as a settings change that names the
// setting and no value, is TestEveryOwnerChangeWritesExactlyOneEntry's.
func TestClientEventsSettingOnARealCore(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	wantEvents(t, oc, uiCaller, true, "a new core's client events (D-227: on by default)")
	for _, token := range []string{"", "anything"} {
		if _, err := clientEvents(t, oc, token); connect.CodeOf(err) == 0 {
			t.Errorf("a caller with the token %q read the setting", token)
		}
	}
	if err := setClientEvents(t, oc, "ofsess_nope_nope", false); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("SetClientEvents with an unknown session = %v, want unauthenticated", err)
	}
	wantEvents(t, oc, uiCaller, true, "after a refused change the client events")
	for _, enabled := range []bool{false, true} {
		if err := setClientEvents(t, oc, oc.session, enabled); err != nil {
			t.Fatalf("SetClientEvents(%v): %v", enabled, err)
		}
		wantEvents(t, oc, uiCaller, enabled, "after the owner's change the client events")
	}
}

// setClientEvents calls SetClientEvents as the view plugin with session.
func setClientEvents(t *testing.T, oc ownerCore, session string, enabled bool) error {
	t.Helper()
	_, err := settingsClient(oc.srv, uiCaller).SetClientEvents(t.Context(), connect.NewRequest(&valuationv1.SetClientEventsRequest{
		Origin: localUI, SessionToken: session, Enabled: enabled,
	}))
	return err //nolint:wrapcheck // the test reads the code of the core's own error
}

// wantEvents fails t unless the setting reads as want for the caller with token.
func wantEvents(t *testing.T, oc ownerCore, token string, want bool, what string) {
	t.Helper()
	if on, err := clientEvents(t, oc, token); err != nil || on != want {
		t.Fatalf("%s = %v, %v; want %v", what, on, err, want)
	}
}
