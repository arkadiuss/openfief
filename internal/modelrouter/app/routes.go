package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// RouteProblems returns the broken entries of the routes in use (D-196): each
// model a route names whose provider instance no plugin has, or whose plugin
// is not a model provider, in role order and then route order. Only the roles
// that may run count: with the switch off, or the role's toggle off, no route
// is in use and none is broken. An instance that exists but is not ready, or
// that can't be looked up now, is not broken: the owner has nothing to fix,
// and the router's fallback covers it. Before the core has found the plugins
// it can't tell, and returns an error wrapping ErrPluginsNotFound.
//
// The periodic attention check turns each into an item and clears the items
// of entries that were fixed or removed; the router raises the same items at
// once when a call meets one.
func RouteProblems(ctx context.Context, settings SettingsStore, instances ProviderInstances) ([]modelrouter.RouteProblem, error) {
	stored, err := settings.Settings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the models settings: %w", err)
	}
	return routeProblems(ctx, &stored, instances, true)
}

// routeProblems finds the broken entries of stored's routes: of the roles in
// use only when inUse is true, of every role otherwise.
func routeProblems(ctx context.Context, stored *modelrouter.Settings, instances ProviderInstances, inUse bool) ([]modelrouter.RouteProblem, error) {
	checked := map[string]error{}
	var out []modelrouter.RouteProblem
	for _, role := range modelrouter.Roles() {
		if inUse && (!stored.Switch.Enabled || !stored.Switch.Allows(role.Feature())) {
			continue
		}
		for _, m := range stored.Routes[role].Models {
			reason, err := instanceProblem(ctx, instances, m.Instance, checked)
			if err != nil {
				return nil, err
			}
			if reason != "" {
				out = append(out, modelrouter.RouteProblem{Role: role, Model: m, Reason: reason})
			}
		}
	}
	return slices.Clip(out), nil
}

// instanceProblem returns why a route entry naming instance is broken, or the
// zero reason when it is not. It asks each instance once, remembering the
// answer in checked, and returns an error wrapping ErrPluginsNotFound when the
// core can't tell yet.
func instanceProblem(ctx context.Context, instances ProviderInstances, instance string, checked map[string]error) (modelrouter.ProblemReason, error) {
	got, seen := checked[instance]
	if !seen {
		if err := instances.Instance(ctx, instance); err != nil {
			got = fmt.Errorf("provider instance %s: %w", instance, err)
		}
		checked[instance] = got
	}
	switch {
	case errors.Is(got, ErrPluginsNotFound):
		return "", got
	case errors.Is(got, ErrProviderMissing):
		return modelrouter.ProblemMissing, nil
	case errors.Is(got, ErrNotModelProvider):
		return modelrouter.ProblemNotModelProvider, nil
	}
	return "", nil
}
