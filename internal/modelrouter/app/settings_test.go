package app_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

const (
	origin     = "https://openfief.example"
	ownerToken = "owner-session"
)

func newSettings() (*app.Settings, *fakeStore) {
	store := &fakeStore{}
	return app.NewSettings(store, fakeOwners{token: ownerToken}), store
}

func TestSettingsDefaultsAreOff(t *testing.T) {
	t.Parallel()
	s, _ := newSettings()
	got, err := s.Owner(t.Context(), origin, ownerToken)
	if err != nil {
		t.Fatal(err)
	}
	if got.Switch.Enabled || len(got.Switch.Toggles) != 0 || len(got.Routes) != 0 || len(got.Policy) != 0 {
		t.Errorf("default settings = %+v, want everything off", got)
	}
	sw, err := s.Switch(t.Context())
	if err != nil || sw.Allows(modelrouter.FeatureAsk) {
		t.Errorf("Switch = %+v, %v", sw, err)
	}
}

func TestSettingsOnlyTheOwnerChangesThem(t *testing.T) {
	t.Parallel()
	changes := map[string]func(s *app.Settings, token string) error{
		"switch": func(s *app.Settings, token string) error { return s.SetEnabled(t.Context(), origin, token, true) },
		"toggle": func(s *app.Settings, token string) error {
			return s.SetFeature(t.Context(), origin, token, "ask", true)
		},
		"route": func(s *app.Settings, token string) error {
			return s.SetRoute(t.Context(), origin, token, "ask", []string{"p/m"})
		},
		"tier": func(s *app.Settings, token string) error {
			return s.SetDataTierModels(t.Context(), origin, token, "T2", []string{"p/m"})
		},
		"reading": func(s *app.Settings, token string) error {
			if _, err := s.Owner(t.Context(), origin, token); err != nil {
				return fmt.Errorf("reading: %w", err)
			}
			return nil
		},
	}
	callers := []struct {
		name  string
		token string
		want  error
	}{
		{"no session", "nobody", app.ErrNoSession},
		{"a read session", "read", app.ErrNotOwner},
	}
	for what, change := range changes {
		for _, c := range callers {
			t.Run(what+", "+c.name, func(t *testing.T) {
				t.Parallel()
				s, store := newSettings()
				if err := change(s, c.token); !errors.Is(err, c.want) {
					t.Fatalf("error = %v, want %v", err, c.want)
				}
				if diff := cmp.Diff(modelrouter.Settings{}, store.settings, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("a refused caller changed the settings (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestSettingsSetByTheOwner(t *testing.T) {
	t.Parallel()
	s, _ := newSettings()
	ctx := t.Context()
	steps := []error{
		s.SetEnabled(ctx, origin, ownerToken, true),
		s.SetFeature(ctx, origin, ownerToken, "ask", true),
		s.SetFeature(ctx, origin, ownerToken, "mcp", true),
		s.SetFeature(ctx, origin, ownerToken, "mcp", false),
		s.SetRoute(ctx, origin, ownerToken, "ask", []string{"openrouter/openai/gpt-oss-120b", "ollama/qwen3:8b"}),
		s.SetDataTierModels(ctx, origin, ownerToken, "T2", []string{"openrouter/openai/gpt-oss-120b"}),
	}
	for i, err := range steps {
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	got, err := s.Owner(ctx, origin, ownerToken)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Switch.Enabled || !got.Switch.Allows(modelrouter.FeatureAsk) || got.Switch.Allows(modelrouter.FeatureMCP) {
		t.Errorf("switch = %+v", got.Switch)
	}
	checkRouteAndPolicy(t, &got)
}

// checkRouteAndPolicy checks the route and the tier lists of
// TestSettingsSetByTheOwner.
func checkRouteAndPolicy(t *testing.T, got *modelrouter.Settings) {
	t.Helper()
	route := got.Routes[modelrouter.RoleAsk]
	if len(route.Models) != 2 || route.Models[0].String() != "openrouter/openai/gpt-oss-120b" {
		t.Fatalf("route = %+v", route)
	}
	if !got.Policy.Allows(modelrouter.DataTierT2, route.Models[0]) || got.Policy.Allows(modelrouter.DataTierT0, route.Models[0]) {
		t.Errorf("policy = %+v", got.Policy)
	}
}

func TestSettingsRouteCleared(t *testing.T) {
	t.Parallel()
	s, _ := newSettings()
	ctx := t.Context()
	if err := s.SetRoute(ctx, origin, ownerToken, "ask", []string{"p/m"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRoute(ctx, origin, ownerToken, "ask", nil); err != nil {
		t.Fatal(err)
	}
	got, err := s.Owner(ctx, origin, ownerToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Routes) != 0 {
		t.Errorf("routes after clearing = %+v", got.Routes)
	}
}

func TestSettingsInvalidInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		do   func(s *app.Settings) error
	}{
		{"unknown feature", func(s *app.Settings) error { return s.SetFeature(t.Context(), origin, ownerToken, "trade", true) }},
		{"unknown role", func(s *app.Settings) error {
			return s.SetRoute(t.Context(), origin, ownerToken, "trader", []string{"p/m"})
		}},
		{"a reference without a provider", func(s *app.Settings) error { return s.SetRoute(t.Context(), origin, ownerToken, "ask", []string{"m"}) }},
		{"a model twice in a route", func(s *app.Settings) error {
			return s.SetRoute(t.Context(), origin, ownerToken, "ask", []string{"p/m", "p/m"})
		}},
		{"unknown tier", func(s *app.Settings) error {
			return s.SetDataTierModels(t.Context(), origin, ownerToken, "T9", []string{"p/m"})
		}},
		{"a bad reference for a tier", func(s *app.Settings) error {
			return s.SetDataTierModels(t.Context(), origin, ownerToken, "T2", []string{"P/m"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, store := newSettings()
			if err := tt.do(s); !errors.Is(err, app.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
			if diff := cmp.Diff(modelrouter.Settings{}, store.settings, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("settings changed (-want +got):\n%s", diff)
			}
		})
	}
}
