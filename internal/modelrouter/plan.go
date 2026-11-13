package modelrouter

import (
	"errors"
	"fmt"
)

// Why a request is refused. Each is a business outcome, not a failure: the
// router records it as a refusal and no model is called.
var (
	// ErrSwitchOff reports that the one switch for agents and models is off
	// (D-212): nothing calls a model.
	ErrSwitchOff = errors.New("agents and models are switched off")
	// ErrFeatureOff reports that the role's feature is off under the switch.
	ErrFeatureOff = errors.New("the feature's toggle is off")
	// ErrNoRoute reports a role with no models in its route.
	ErrNoRoute = errors.New("the role has no models")
	// ErrDataTierDenied reports that no model of the role's route is allowed to
	// receive every tier of the request's data (D-206).
	ErrDataTierDenied = errors.New("no model of the route may receive the request's data tiers")
	// ErrNoCapableModel reports that no model allowed for the data supports
	// what the request asks, such as tools or structured output.
	ErrNoCapableModel = errors.New("no allowed model supports what the request asks")
)

// PlanInput is what Plan decides over.
type PlanInput struct {
	Switch Switch
	Role   Role
	// Route is the role's route; its zero value is no route.
	Route  Route
	Policy DataTierPolicy
	// DataTiers are the data tiers of the request.
	DataTiers DataTierSet
	Needs     Needs
	// Catalog holds the entries the providers listed. A model of the route
	// that is not in it is not skipped for what it can do: the provider
	// itself answers for a model it does not have.
	Catalog map[Ref]ModelEntry
}

// SkipReason is why a model of a route is not a candidate.
type SkipReason string

// The reasons a model of a route is skipped.
const (
	// SkipDataTierDenied: the owner has not allowed the model for every tier of
	// the request's data.
	SkipDataTierDenied SkipReason = "tier_denied"
	// SkipIncapable: the model's entry lacks what the request asks.
	SkipIncapable SkipReason = "incapable"
)

// Skip is a model of the route that Plan left out, and why. It names the
// model and the reason only.
type Skip struct {
	Model  Ref
	Reason SkipReason
	// DataTiers are the tiers the model is denied, for SkipDataTierDenied.
	DataTiers []DataTier
}

// Plan is the models a request may be sent to, in the order they are tried.
type Plan struct {
	Candidates []Ref
	Skipped    []Skip
}

// MakePlan decides which models of the role's route may be sent a request:
// the switch and the role's toggle must be on (D-212), the route must have
// models, and a model is a candidate only when the owner allowed it for
// every tier of the request's data, denying by default (D-206), and its
// entry supports what the request asks. The candidates keep the route's
// order: the primary, then the fallbacks. It returns an error wrapping
// ErrSwitchOff, ErrFeatureOff, ErrNoRoute, ErrDataTierDenied or
// ErrNoCapableModel, and ErrUntagged when the request has no tiers.
func MakePlan(in *PlanInput) (Plan, error) {
	if err := in.mayRun(); err != nil {
		return Plan{}, err
	}
	var plan Plan
	allowed := 0
	for _, m := range in.Route.Models {
		if denied := in.Policy.Denied(m, in.DataTiers); len(denied) > 0 {
			plan.Skipped = append(plan.Skipped, Skip{Model: m, Reason: SkipDataTierDenied, DataTiers: denied})
			continue
		}
		allowed++
		if e, listed := in.Catalog[m]; listed && !e.Serves(in.Needs) {
			plan.Skipped = append(plan.Skipped, Skip{Model: m, Reason: SkipIncapable})
			continue
		}
		plan.Candidates = append(plan.Candidates, m)
	}
	switch {
	case allowed == 0:
		return plan, ErrDataTierDenied
	case len(plan.Candidates) == 0:
		return plan, ErrNoCapableModel
	}
	return plan, nil
}

// mayRun returns the error for a request that no model may be asked at all:
// the switch or the role's toggle is off, the role has no route, or the
// request has no tier.
func (in *PlanInput) mayRun() error {
	switch {
	case !in.Switch.Enabled:
		return ErrSwitchOff
	case !in.Switch.Allows(in.Role.Feature()):
		return fmt.Errorf("%w: %s", ErrFeatureOff, in.Role.Feature())
	case len(in.Route.Models) == 0:
		return fmt.Errorf("%w: %s", ErrNoRoute, in.Role)
	case len(in.DataTiers) == 0:
		return ErrUntagged
	}
	return nil
}
