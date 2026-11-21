package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// The failures the fakes make.
var (
	errAuditDown             = errors.New("the log is down")
	errClientEventsStoreDown = errors.New("the store is down")
)

// fakeClientEvents keeps the switch for the UI's client events in memory.
type fakeClientEvents struct {
	off bool
	err error
}

func (f *fakeClientEvents) ClientEventsOff(context.Context) (bool, error) { return f.off, f.err }

func (f *fakeClientEvents) SaveClientEventsOff(_ context.Context, off bool) error {
	if f.err != nil {
		return f.err
	}
	f.off = off
	return nil
}

// recordingAuditor records the changes it was asked to enter, as the audit log
// would, and runs them; refuse makes it enter nothing and fail.
type recordingAuditor struct {
	actions []string
	refs    [][]string
	refuse  error
}

func (a *recordingAuditor) Audited(ctx context.Context, _, _, action string, refs []string, change func(context.Context) error) error {
	if a.refuse != nil {
		return a.refuse
	}
	a.actions, a.refs = append(a.actions, action), append(a.refs, refs)
	return change(ctx)
}

func TestClientEventsAreOnUntilTheOwnerTurnsThemOff(t *testing.T) {
	t.Parallel()
	events := &fakeClientEvents{}
	audit := &recordingAuditor{}
	s := newSettings(t, newFakeStore(), &fakeRecorder{}).WithClientEvents(events).WithAudit(audit)

	// On by default, read with no session at all.
	wantClientEvents(t, s, true, "by default")
	for _, step := range []struct {
		enabled bool
		what    string
	}{{false, "after turning them off"}, {true, "after turning them on"}} {
		if err := s.SetClientEventsEnabled(t.Context(), "o", ownerToken, step.enabled); err != nil {
			t.Fatalf("SetClientEventsEnabled(%v): %v", step.enabled, err)
		}
		wantClientEvents(t, s, step.enabled, step.what)
	}
	// Each change is entered in the audit log as a settings change naming the setting, and nothing else.
	wantActions := []string{"settings_changed", "settings_changed"}
	wantRefs := [][]string{{"setting:ui:client_events"}, {"setting:ui:client_events"}}
	if diff := cmp.Diff(wantActions, audit.actions); diff != "" {
		t.Errorf("audit actions (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantRefs, audit.refs); diff != "" {
		t.Errorf("audit refs (-want +got):\n%s", diff)
	}
}

// wantClientEvents fails t unless the client events read as want, without an error.
func wantClientEvents(t *testing.T, s *app.Settings, want bool, when string) {
	t.Helper()
	if on, err := s.ClientEventsEnabled(t.Context()); err != nil || on != want {
		t.Errorf("%s: ClientEventsEnabled = %v, %v, want %v", when, on, err, want)
	}
}

func TestClientEventsCanBeChangedByTheOwnerAlone(t *testing.T) {
	t.Parallel()
	events := &fakeClientEvents{}
	audit := &recordingAuditor{}
	s := newSettings(t, newFakeStore(), &fakeRecorder{}).WithClientEvents(events).WithAudit(audit)
	err := s.SetClientEventsEnabled(t.Context(), "o", "read-session", false)
	if !errors.Is(err, app.ErrNotOwner) {
		t.Errorf("a read session: error = %v, want ErrNotOwner", err)
	}
	if events.off || len(audit.actions) != 0 {
		t.Errorf("a read session changed the setting (off = %v) or entered %v", events.off, audit.actions)
	}
	// A change the audit log can't take is not made.
	audit.refuse = errAuditDown
	if err := s.SetClientEventsEnabled(t.Context(), "o", ownerToken, false); err == nil {
		t.Error("a change the audit log refused was made")
	}
	if events.off {
		t.Error("the setting changed although the audit log took no entry")
	}
}

func TestClientEventsErrors(t *testing.T) {
	t.Parallel()
	boom := errClientEventsStoreDown
	s := newSettings(t, newFakeStore(), &fakeRecorder{}).WithClientEvents(&fakeClientEvents{err: boom})
	if _, err := s.ClientEventsEnabled(t.Context()); !errors.Is(err, boom) {
		t.Errorf("ClientEventsEnabled over a failing store = %v, want its error", err)
	}
	if err := s.SetClientEventsEnabled(t.Context(), "o", ownerToken, false); !errors.Is(err, boom) {
		t.Errorf("SetClientEventsEnabled over a failing store = %v, want its error", err)
	}
	// Settings built without the store say so, and never claim the pages may report.
	bare := newSettings(t, newFakeStore(), &fakeRecorder{})
	on, err := bare.ClientEventsEnabled(t.Context())
	if err == nil || on {
		t.Errorf("ClientEventsEnabled without a store = %v, %v, want off and an error", on, err)
	}
	if diff := cmp.Diff(app.ErrNotOwner, bare.SetClientEventsEnabled(t.Context(), "o", "read-session", false), cmpopts.EquateErrors()); diff != "" {
		t.Errorf("the owner check comes first (-want +got):\n%s", diff)
	}
}
