package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// fakeAskers accepts the session tokens it knows, as the actor of each.
type fakeAskers map[string]app.Actor

func (f fakeAskers) Asker(_ context.Context, _, token string) (app.Actor, error) {
	if who, ok := f[token]; ok {
		return who, nil
	}
	return app.Actor{}, app.ErrNoSession
}

const (
	askerToken  = "asker-token"
	readerToken = "read-token"
)

func newAskFixture(t *testing.T) (*routerFixture, *app.Ask) {
	t.Helper()
	f := newRouterFixture(t)
	askers := fakeAskers{
		askerToken:  {Kind: app.ActorOwner, ID: "session-owner"},
		readerToken: {Kind: app.ActorViewer, ID: "session-read"},
	}
	return f, app.NewAsk(askers, f.store, f.router)
}

func TestAskAvailable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mod  func(s *modelrouter.Settings)
		want bool
	}{
		{"switch on, toggle on, a route", func(*modelrouter.Settings) {}, true},
		{"nothing saved", func(s *modelrouter.Settings) { *s = modelrouter.Settings{} }, false},
		{"switch off", func(s *modelrouter.Settings) { s.Switch.Enabled = false }, false},
		{"toggle off", func(s *modelrouter.Settings) { s.Switch.Toggles = nil }, false},
		{"only the maintainer's toggle on", func(s *modelrouter.Settings) {
			s.Switch.Toggles = map[modelrouter.Feature]bool{modelrouter.FeatureMaintainer: true}
		}, false},
		{"no route for the role", func(s *modelrouter.Settings) { s.Routes = nil }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, ask := newAskFixture(t)
			tt.mod(&f.store.settings)
			got, err := ask.Available(t.Context(), "https://ui.example.com", readerToken)
			if err != nil {
				t.Fatalf("Available: %v", err)
			}
			if got != tt.want {
				t.Errorf("Available = %v, want %v", got, tt.want)
			}
			if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.providers.called) != 0 {
				t.Errorf("asking whether there is a model wrote audit entries %v %v or called %v", f.audit.sent, f.audit.results, f.providers.called)
			}
		})
	}
}

func TestAskNeedsAValidSession(t *testing.T) {
	t.Parallel()
	f, ask := newAskFixture(t)
	ctx := t.Context()
	if _, err := ask.Available(ctx, "o", "nobody"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Available with no session: %v, want ErrNoSession", err)
	}
	if _, err := ask.Permits(ctx, "o", "nobody", []string{"T2"}); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Permits with no session: %v, want ErrNoSession", err)
	}
	if _, err := ask.Generate(ctx, "o", "nobody", request(modelrouter.DataTierT2)); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Generate with no session: %v, want ErrNoSession", err)
	}
	if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.providers.called) != 0 {
		t.Errorf("a call without a session wrote audit entries or called a model: %v %v %v", f.audit.sent, f.audit.results, f.providers.called)
	}
}

func TestAskPermits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tiers   []string
		mod     func(s *modelrouter.Settings)
		want    bool
		wantErr error
	}{
		{name: "public data goes to the cloud model", tiers: []string{"T2"}, want: true},
		{name: "the portfolio's data goes to the local model", tiers: []string{"T0", "T2"}, want: true},
		{
			name: "no model of the route may receive the tier", tiers: []string{"T0"},
			mod: func(s *modelrouter.Settings) { s.Policy = modelrouter.DataTierPolicy{} },
		},
		{
			name: "every tier together or none: the model allowed for one is not enough", tiers: []string{"T0", "T2"},
			mod: func(s *modelrouter.Settings) {
				s.Policy = modelrouter.DataTierPolicy{
					modelrouter.DataTierT0: {{Instance: "ollama", Model: "qwen3:8b"}},
					modelrouter.DataTierT2: {{Instance: "openrouter", Model: "qwen"}},
				}
			},
		},
		{name: "an unknown tier", tiers: []string{"T9"}, wantErr: app.ErrInvalidInput},
		{name: "no tier at all", wantErr: app.ErrInvalidInput},
		{
			name: "the switch went off", tiers: []string{"T2"},
			mod:     func(s *modelrouter.Settings) { s.Switch.Enabled = false },
			wantErr: modelrouter.ErrSwitchOff,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, ask := newAskFixture(t)
			if tt.mod != nil {
				tt.mod(&f.store.settings)
			}
			got, err := ask.Permits(t.Context(), "o", readerToken, tt.tiers)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Permits error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Permits = %v, want %v", got, tt.want)
			}
			if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.providers.called) != 0 {
				t.Errorf("asking whether a tier may go wrote audit entries or called a model: %v %v %v", f.audit.sent, f.audit.results, f.providers.called)
			}
		})
	}
}

func TestAskPermitsNeedsAModelThatServesTools(t *testing.T) {
	t.Parallel()
	f, ask := newAskFixture(t)
	f.providers.entries = map[string][]modelrouter.ModelEntry{
		"openrouter": {{ID: "qwen"}}, "ollama": {{ID: "qwen3:8b"}},
	}
	got, err := ask.Permits(t.Context(), "o", askerToken, []string{"T2"})
	if err != nil || got {
		t.Errorf("Permits with no model that serves tools = %v, %v; want false and no error", got, err)
	}
}

// TestAskGenerateIsTheAskRolesCallForTheSession checks that the Ask box's door
// calls the router for the Ask box's role only, for the asking session, and
// that the audit entries hold that and no text (GR-12, D-211).
func TestAskGenerateIsTheAskRolesCallForTheSession(t *testing.T) {
	t.Parallel()
	f, ask := newAskFixture(t)
	res, err := ask.Generate(t.Context(), "o", readerToken, request(modelrouter.DataTierT2))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Model != f.cloud {
		t.Errorf("answered by %s, want the primary %s", res.Model, f.cloud)
	}
	// A viewer's question is the viewer's, never the owner's (D-211).
	wantActor := app.Actor{Kind: app.ActorViewer, ID: "session-read"}
	if len(f.audit.sent) != 1 || len(f.audit.results) != 1 {
		t.Fatalf("audit entries: %d sent, %d results; want one pair", len(f.audit.sent), len(f.audit.results))
	}
	if diff := cmp.Diff(wantActor, f.audit.sent[0].Actor); diff != "" {
		t.Errorf("sent entry's actor (-want +got):\n%s", diff)
	}
	if f.audit.sent[0].Role != modelrouter.RoleAsk || f.audit.results[0].Role != modelrouter.RoleAsk {
		t.Errorf("roles = %s, %s; want ask", f.audit.sent[0].Role, f.audit.results[0].Role)
	}
	if got := f.audit.results[0].Outcome; got != app.OutcomeOK {
		t.Errorf("outcome = %s, want ok", got)
	}
}

func TestAskGenerateRefusesDataTheTiersDeny(t *testing.T) {
	t.Parallel()
	f, ask := newAskFixture(t)
	f.store.settings.Policy = modelrouter.DataTierPolicy{}
	_, err := ask.Generate(t.Context(), "o", askerToken, request(modelrouter.DataTierT0))
	if !errors.Is(err, modelrouter.ErrDataTierDenied) {
		t.Fatalf("Generate error = %v, want ErrDataTierDenied", err)
	}
	if len(f.audit.sent) != 0 || len(f.audit.results) != 1 || f.audit.results[0].Outcome != app.OutcomeRefused {
		t.Errorf("audit = %v %v, want only a refused result", f.audit.sent, f.audit.results)
	}
	if len(f.providers.called) != 0 {
		t.Errorf("called %v, want no model", f.providers.called)
	}
	if strings.Contains(err.Error(), "SECRET-QUESTION-TEXT") {
		t.Errorf("the error quotes the question: %v", err)
	}
}

func TestRouterPermitsCallsNothing(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	tiers, err := modelrouter.NewDataTierSet([]modelrouter.DataTier{modelrouter.DataTierT0})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.router.Permits(t.Context(), modelrouter.RoleAsk, tiers, modelrouter.Needs{Tools: true}); err != nil {
		t.Errorf("Permits(T0) = %v, want nil: the local model may receive it", err)
	}
	f.store.settings.Policy = modelrouter.DataTierPolicy{}
	if err := f.router.Permits(t.Context(), modelrouter.RoleAsk, tiers, modelrouter.Needs{}); !errors.Is(err, modelrouter.ErrDataTierDenied) {
		t.Errorf("Permits(T0) with no model allowed = %v, want ErrDataTierDenied", err)
	}
	f.store.err = errBoom
	if err := f.router.Permits(t.Context(), modelrouter.RoleAsk, tiers, modelrouter.Needs{}); !errors.Is(err, errBoom) {
		t.Errorf("Permits without settings = %v, want the store's error: it fails closed", err)
	}
	if len(f.audit.sent)+len(f.audit.results) != 0 || len(f.providers.called) != 0 {
		t.Errorf("Permits wrote audit entries or called a model: %v %v %v", f.audit.sent, f.audit.results, f.providers.called)
	}
}
