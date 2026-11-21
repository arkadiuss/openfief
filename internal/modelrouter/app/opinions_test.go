package app_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

var (
	errAuditDown = errors.New("the audit log is down")
	errStoreDown = errors.New("the settings store is down")
)

func TestSetOpinionsRefusesANonOwnerAndStoresNothing(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{}
	s, store := auditedSettings(audit)
	if store.settings.Opinions {
		t.Fatal("opinions start on")
	}
	if err := s.SetOpinions(t.Context(), origin, "nobody", true); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("SetOpinions with no session = %v, want ErrNoSession", err)
	}
	if err := s.SetOpinions(t.Context(), origin, "read", true); !errors.Is(err, app.ErrNotOwner) {
		t.Errorf("SetOpinions with a read session = %v, want ErrNotOwner", err)
	}
	if store.settings.Opinions || len(audit.changes) != 0 {
		t.Errorf("a refused caller changed the setting or wrote an entry: %v %v", store.settings.Opinions, audit.changes)
	}
}

func TestSetOpinionsIsAuditedAsOnAndOff(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{}
	s, _ := auditedSettings(audit)
	for _, on := range []bool{true, false} {
		if err := s.SetOpinions(t.Context(), origin, ownerToken, on); err != nil {
			t.Fatalf("SetOpinions(%v): %v", on, err)
		}
	}
	got := make([]string, 0, len(audit.changes))
	for _, c := range audit.changes {
		got = append(got, c.Setting)
	}
	if len(got) != 2 || got[0] != "opinions:on" || got[1] != "opinions:off" {
		t.Errorf("audited settings = %v, want opinions:on then opinions:off", got)
	}
}

func TestSetOpinionsOnANeverStoresWhenTheLogFails(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{err: errAuditDown}
	s, store := auditedSettings(audit)
	if err := s.SetOpinions(t.Context(), origin, ownerToken, true); !errors.Is(err, errAuditDown) {
		t.Fatalf("SetOpinions = %v, want the log's error", err)
	}
	if store.settings.Opinions {
		t.Error("a change whose entry failed was stored")
	}
}

func TestADemoNeverTurnsOpinionsOn(t *testing.T) {
	t.Parallel()
	audit := &fakeSettingsAudit{}
	s, store := auditedSettings(audit)
	s = s.WithDemo(true)
	if !s.Demo() {
		t.Fatal("Demo() = false after WithDemo(true)")
	}
	if err := s.SetOpinions(t.Context(), origin, ownerToken, true); !errors.Is(err, app.ErrOpinionsOnDemo) {
		t.Fatalf("SetOpinions(on) on a demo = %v, want ErrOpinionsOnDemo", err)
	}
	if store.settings.Opinions || len(audit.changes) != 0 {
		t.Errorf("a refused change was stored or audited: %v %v", store.settings.Opinions, audit.changes)
	}
	// Turning them off is always allowed.
	if err := s.SetOpinions(t.Context(), origin, ownerToken, false); err != nil {
		t.Errorf("SetOpinions(off) on a demo = %v, want it made", err)
	}
}

func TestAskOpinions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		demo  bool
		token string
		mod   func(f *routerFixture)
		want  modelrouter.OpinionsStatus
	}{
		{"off by default", false, askerToken, func(*routerFixture) {}, modelrouter.OpinionsOff},
		{"on, the owner, a route that may receive every tier", false, askerToken, func(f *routerFixture) { f.store.settings.Opinions = true }, modelrouter.OpinionsOn},
		{"on, but a viewer asks", false, readerToken, func(f *routerFixture) { f.store.settings.Opinions = true }, modelrouter.OpinionsOff},
		{"on, a demo", true, askerToken, func(f *routerFixture) { f.store.settings.Opinions = true }, modelrouter.OpinionsOff},
		{"on, the switch off", false, askerToken, func(f *routerFixture) {
			f.store.settings.Opinions = true
			f.store.settings.Switch.Enabled = false
		}, modelrouter.OpinionsOff},
		{"on, no model may receive the thesis's tier", false, askerToken, func(f *routerFixture) {
			f.store.settings.Opinions = true
			delete(f.store.settings.Policy, modelrouter.DataTierT1)
		}, modelrouter.OpinionsRouteIneligible},
		{"on, no model may receive the question's tier", false, askerToken, func(f *routerFixture) {
			f.store.settings.Opinions = true
			delete(f.store.settings.Policy, modelrouter.DataTierT0)
		}, modelrouter.OpinionsRouteIneligible},
		{"on, no model may receive public data", false, askerToken, func(f *routerFixture) {
			f.store.settings.Opinions = true
			delete(f.store.settings.Policy, modelrouter.DataTierT2)
		}, modelrouter.OpinionsRouteIneligible},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, ask := newAskFixture(t)
			tt.mod(f)
			ask = ask.WithDemo(tt.demo)
			got, err := ask.Opinions(t.Context(), "https://ui.example.com", tt.token)
			if err != nil || got != tt.want {
				t.Errorf("Opinions = %q, %v; want %q", got, err, tt.want)
			}
			if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.providers.called) != 0 {
				t.Errorf("deciding wrote audit entries or called a model")
			}
		})
	}
	f, ask := newAskFixture(t)
	if _, err := ask.Opinions(t.Context(), "o", "nobody"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Opinions with no session = %v, want ErrNoSession", err)
	}
	f.store.err = errStoreDown
	if _, err := ask.Opinions(t.Context(), "o", askerToken); err == nil {
		t.Error("Opinions with the settings unreadable succeeded")
	}
}
