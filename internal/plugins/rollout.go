package plugins

import (
	"errors"
	"fmt"
)

// RolloutOutcome says which version of a plugin runs, if any, and why.
type RolloutOutcome string

// The rollout outcomes.
const (
	// RolloutCurrent runs the version the Plugin resource asks for.
	RolloutCurrent RolloutOutcome = "Current"
	// RolloutPrevious keeps the version that runs now: the current one waits
	// for approval or cannot be deployed. It runs under the current grant.
	RolloutPrevious RolloutOutcome = "Previous"
	// RolloutNoGrant runs nothing: the owner has granted the plugin nothing.
	RolloutNoGrant RolloutOutcome = "NoGrant"
	// RolloutAwaitingApproval runs nothing: the current version requests
	// permissions the grant lacks, and no earlier version runs.
	RolloutAwaitingApproval RolloutOutcome = "AwaitingApproval"
	// RolloutNotDeployable runs nothing: the current version cannot be
	// deployed, and no earlier version runs.
	RolloutNotDeployable RolloutOutcome = "NotDeployable"
	// RolloutNothingGranted runs nothing: the version that would run has no
	// granted capability (ErrNoCapabilityGranted).
	RolloutNothingGranted RolloutOutcome = "NothingGranted"
)

// RolloutInput is what a rollout decision depends on.
type RolloutInput struct {
	// Current is the version the Plugin resource asks for; nil when it is
	// invalid.
	Current *Plugin
	// CurrentDeployable is false when a check outside this package refused
	// Current, such as its image signature or its config schema.
	CurrentDeployable bool
	// Previous is the version that runs now; nil when none does.
	Previous *Plugin
	// Grant is what the owner approved; nil when there is no grant.
	Grant *Grant
}

// Rollout is the decision: what runs, with which permissions, and what the
// current version still needs approved.
type Rollout struct {
	Outcome RolloutOutcome
	// Run is the effective permissions of the version that runs, or nil when
	// nothing runs.
	Run *EffectivePermissions
	// Pending is what Current requests and the grant does not approve, sorted:
	// everything it requests when there is no grant. Empty when Current is nil.
	Pending Permissions
}

// PlanRollout decides which plugin version runs under a grant (D-16):
//   - no grant: nothing runs, and everything the current version requests is
//     pending;
//   - the current version is deployable and requests nothing beyond the
//     grant: it runs;
//   - otherwise the previous version, if any, keeps running, limited to what
//     the current grant approves, so a narrowed grant takes effect at once;
//   - otherwise nothing runs.
//
// Whatever runs gets only what its manifest requests and the grant approves
// (Effective); when that is no capability at all, nothing runs. It fails when
// the grant is invalid or belongs to another plugin, or when the previous
// version is invalid.
func PlanRollout(in RolloutInput) (Rollout, error) {
	if in.Grant == nil {
		r := Rollout{Outcome: RolloutNoGrant}
		if in.Current != nil {
			r.Pending = in.Current.Manifest.Requested()
		}
		return r, nil
	}
	pending, err := in.pending()
	if err != nil {
		return Rollout{}, err
	}
	r := Rollout{Pending: pending}
	switch {
	case in.currentDeployable() && r.Pending.IsEmpty():
		err = r.runUnder(in.Current, in.Grant, RolloutCurrent)
	case in.Previous != nil:
		err = r.runUnder(in.Previous, in.Grant, RolloutPrevious)
	case in.currentDeployable():
		r.Outcome = RolloutAwaitingApproval
	default:
		r.Outcome = RolloutNotDeployable
	}
	if err != nil {
		return Rollout{}, err
	}
	return r, nil
}

// pending validates the grant and returns what the current version requests
// beyond it; nothing when there is no current version.
func (in *RolloutInput) pending() (Permissions, error) {
	if err := in.Grant.Validate(); err != nil {
		return Permissions{}, fmt.Errorf("grant for %q: %w", in.Grant.Plugin, err)
	}
	if in.Current == nil {
		return Permissions{}, nil
	}
	return NeedsApproval(in.Grant, in.Current)
}

// currentDeployable reports whether the current version is valid and passed
// the checks outside this package.
func (in *RolloutInput) currentDeployable() bool { return in.Current != nil && in.CurrentDeployable }

// runUnder sets r to run p under g with the given outcome, or to run nothing
// when g grants none of p's capabilities.
func (r *Rollout) runUnder(p *Plugin, g *Grant, outcome RolloutOutcome) error {
	e, err := Effective(p, g)
	switch {
	case errors.Is(err, ErrNoCapabilityGranted):
		r.Outcome = RolloutNothingGranted
		return nil
	case err != nil:
		return err
	}
	r.Outcome, r.Run = outcome, e
	return nil
}
