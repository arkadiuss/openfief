package app_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// TestRouteProblems checks D-196's check of the routes in use: one problem
// per route entry whose provider instance is missing or is no model provider,
// none for a provider that is only down or not running, none for a route that
// is not in use, and the error of a core that has not found the plugins yet.
func TestRouteProblems(t *testing.T) {
	t.Parallel()
	missing := fmt.Errorf("%w: cloud", app.ErrProviderMissing)
	notProvider := fmt.Errorf("%w: cloud", app.ErrNotModelProvider)
	tests := []struct {
		name      string
		mod       func(f *routerFixture)
		instances map[string]error
		want      []modelrouter.RouteProblem
		wantErr   error
	}{
		{name: "every instance runs"},
		{
			name: "a missing instance", instances: map[string]error{"openrouter": missing},
			want: []modelrouter.RouteProblem{{Role: modelrouter.RoleAsk, Model: mustRef(t, "openrouter/qwen"), Reason: modelrouter.ProblemMissing}},
		},
		{
			name: "a plugin that is no model provider", instances: map[string]error{"ollama": notProvider},
			want: []modelrouter.RouteProblem{{Role: modelrouter.RoleAsk, Model: mustRef(t, "ollama/qwen3:8b"), Reason: modelrouter.ProblemNotModelProvider}},
		},
		{
			name: "one problem per entry, though two share an instance",
			mod: func(f *routerFixture) {
				route, err := modelrouter.NewRoute(modelrouter.RoleAsk, []modelrouter.Ref{mustRef(t, "ollama/a"), mustRef(t, "ollama/b")})
				if err != nil {
					t.Fatal(err)
				}
				f.store.settings.Routes[modelrouter.RoleAsk] = route
			},
			instances: map[string]error{"ollama": missing},
			want: []modelrouter.RouteProblem{
				{Role: modelrouter.RoleAsk, Model: mustRef(t, "ollama/a"), Reason: modelrouter.ProblemMissing},
				{Role: modelrouter.RoleAsk, Model: mustRef(t, "ollama/b"), Reason: modelrouter.ProblemMissing},
			},
		},
		{name: "a provider that is down is no problem", instances: map[string]error{"openrouter": app.ErrProviderUnavailable}},
		{
			name: "the switch off, no route is in use", instances: map[string]error{"openrouter": missing},
			mod: func(f *routerFixture) { f.store.settings.Switch.Enabled = false },
		},
		{
			name: "the role's toggle off, its route is not in use", instances: map[string]error{"openrouter": missing},
			mod: func(f *routerFixture) { f.store.settings.Switch.Toggles = nil },
		},
		{
			name: "the plugins not found yet", instances: map[string]error{"openrouter": fmt.Errorf("%w: %w", app.ErrProviderUnavailable, app.ErrPluginsNotFound)},
			wantErr: app.ErrPluginsNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newRouterFixture(t)
			if tt.mod != nil {
				tt.mod(f)
			}
			f.providers.instances = tt.instances
			got, err := app.RouteProblems(t.Context(), f.store, f.providers)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("problems (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRouteProblemsFailsWithoutSettings checks that the check fails, and so
// clears nothing, when the settings can't be read.
func TestRouteProblemsFailsWithoutSettings(t *testing.T) {
	t.Parallel()
	f := newRouterFixture(t)
	f.store.err = errBoom
	if _, err := app.RouteProblems(t.Context(), f.store, f.providers); !errors.Is(err, errBoom) {
		t.Errorf("error = %v, want the store's", err)
	}
}

// TestRouteProblemText checks the item's subject and detail: they name the
// route and the provider instance, and hold nothing else.
func TestRouteProblemText(t *testing.T) {
	t.Parallel()
	p := modelrouter.RouteProblem{Role: modelrouter.RoleMaintainer, Model: mustRef(t, "openrouter/openai/gpt-oss-120b"), Reason: modelrouter.ProblemMissing}
	if got, want := p.Subject(), "maintainer/openrouter/openai/gpt-oss-120b"; got != want {
		t.Errorf("Subject = %q, want %q", got, want)
	}
	if got, want := p.Detail(), "the maintainer route names provider instance openrouter: no plugin has that name"; got != want {
		t.Errorf("Detail = %q, want %q", got, want)
	}
}
