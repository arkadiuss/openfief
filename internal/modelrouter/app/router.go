package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// catalogTTL is how long a provider's model entries are kept before the
// router lists them again: entries change when the owner reconfigures a
// plugin, which restarts it.
const catalogTTL = 5 * time.Minute

// Router is the model router (D-205): every call to a model goes through it.
// It gives each role its route, refuses what the switch (D-212) or the data
// tiers (D-206) do not allow, writes each call to the audit log before
// sending it (D-211), and tries the route's models in order until one
// answers.
//
// It holds no request text beyond a call's duration, and records and returns
// no text in an error (GR-12).
type Router struct {
	settings  SettingsStore
	providers ProviderInstances
	audit     Auditor
	attention AttentionRaiser
	callIDs   CallIDs
	now       func() time.Time

	mu       sync.Mutex
	catalogs map[string]catalog
}

// catalog is one provider's listed entries and when they were listed.
type catalog struct {
	entries map[string]modelrouter.ModelEntry
	listed  time.Time
}

// RouterPorts is what a Router reads and writes.
type RouterPorts struct {
	// Settings are the owner's switch, routes and tier lists.
	Settings SettingsStore
	// Providers reaches the provider instances.
	Providers ProviderInstances
	// Audit is the append-only log of calls.
	Audit Auditor
	// Attention takes the items of broken routes.
	Attention AttentionRaiser
	// CallIDs makes the IDs that link a call's two audit entries.
	CallIDs CallIDs
	// Now reads the time.
	Now func() time.Time
}

// NewRouter returns a router over ports.
func NewRouter(p *RouterPorts) *Router {
	return &Router{
		settings: p.Settings, providers: p.Providers, audit: p.Audit, attention: p.Attention, callIDs: p.CallIDs, now: p.Now,
		catalogs: map[string]catalog{},
	}
}

// Generate asks the models of role's route for an answer to req, on behalf
// of actor.
//
// With the switch off, or role's feature off, nothing is called: it records
// the refusal and returns an error wrapping modelrouter.ErrSwitchOff or
// modelrouter.ErrFeatureOff, before it reads the request or reaches any
// provider. A request with an untagged piece of data, or for which no model
// of the route is allowed for every tier it holds, is refused the same way
// (modelrouter.ErrUntagged, modelrouter.ErrDataTierDenied, modelrouter.ErrNoCapableModel).
//
// A call is two audit entries under one call ID (D-211). Before each call
// the router writes the entry "sent"; if that fails the call is not made and
// the error wraps ErrAuditUnavailable. After it, the entry of the result: ok,
// failed with an error class, or, for a call refused before a model was
// chosen, only that entry, refused. A model that is unavailable
// (ErrProviderUnavailable) passes the request to the next model of the
// route, each with its own pair of entries; a provider rejecting the request
// ends the call at once. A model whose provider instance is missing or is no
// model provider (ErrProviderMissing, ErrNotModelProvider) also passes it on,
// and raises an attention item for the broken route entry (D-196); a
// provider that is only down raises none. When every candidate failed the
// error wraps ErrNoAnswer. If the result entry of an answered call can't be
// written, the answer is withheld and the error wraps ErrAuditUnavailable.
func (r *Router) Generate(ctx context.Context, role modelrouter.Role, actor Actor, req *modelrouter.Request) (modelrouter.Result, error) {
	stored, err := r.settings.Settings(ctx)
	if err != nil {
		// Without the owner's settings the router can't know the switch is
		// on: it fails closed.
		return modelrouter.Result{}, fmt.Errorf("loading the models settings: %w", err)
	}
	if offErr := r.checkSwitch(stored.Switch, role); offErr != nil {
		return modelrouter.Result{}, r.refuse(ctx, role, actor, offErr)
	}
	if invalid := req.Validate(); invalid != nil {
		return modelrouter.Result{}, fmt.Errorf("validating the request: %w", invalid)
	}
	tiers, err := req.DataTiers()
	if err != nil {
		return modelrouter.Result{}, r.refuse(ctx, role, actor, err)
	}
	plan, err := r.plan(ctx, &stored, role, tiers, req.Needs())
	if err != nil {
		return modelrouter.Result{}, r.refuse(ctx, role, actor, err)
	}
	return r.try(ctx, role, actor, plan.Candidates, req)
}

// Permits reports whether a request with data of tiers, asking what needs
// says, could be sent to a model of role's route now: nil, or the error
// Generate would refuse it with (modelrouter.ErrSwitchOff,
// modelrouter.ErrFeatureOff, modelrouter.ErrNoRoute,
// modelrouter.ErrDataTierDenied or modelrouter.ErrNoCapableModel). It calls
// no provider and writes no audit entry: it is a question about the
// settings, not a call. A caller that holds data it may not be allowed to
// send, such as a tool's result, asks it before adding the data to a
// request, so a refusal names that data and not the whole request.
func (r *Router) Permits(ctx context.Context, role modelrouter.Role, tiers modelrouter.DataTierSet, needs modelrouter.Needs) error {
	stored, err := r.settings.Settings(ctx)
	if err != nil {
		return fmt.Errorf("loading the models settings: %w", err)
	}
	if _, planErr := r.plan(ctx, &stored, role, tiers, needs); planErr != nil {
		return planErr
	}
	return nil
}

// plan decides which models of role's route may be sent a request with
// tiers and needs, under the settings stored.
func (r *Router) plan(ctx context.Context, stored *modelrouter.Settings, role modelrouter.Role, tiers modelrouter.DataTierSet, needs modelrouter.Needs) (modelrouter.Plan, error) {
	route := stored.Routes[role]
	plan, err := modelrouter.MakePlan(&modelrouter.PlanInput{
		Switch: stored.Switch, Role: role, Route: route, Policy: stored.Policy,
		DataTiers: tiers, Needs: needs, Catalog: r.catalogOf(ctx, route),
	})
	if err != nil {
		return plan, fmt.Errorf("planning a call: %w", err)
	}
	return plan, nil
}

// checkSwitch returns the error for role's feature being off, or nil.
func (r *Router) checkSwitch(sw modelrouter.Switch, role modelrouter.Role) error {
	switch {
	case !sw.Enabled:
		return modelrouter.ErrSwitchOff
	case !sw.Allows(role.Feature()):
		return fmt.Errorf("%w: %s", modelrouter.ErrFeatureOff, role.Feature())
	}
	return nil
}

// try calls the candidates in order.
func (r *Router) try(ctx context.Context, role modelrouter.Role, actor Actor, candidates []modelrouter.Ref, req *modelrouter.Request) (modelrouter.Result, error) {
	var last error
	for _, m := range candidates {
		res, class, err := r.attempt(ctx, role, actor, m, req)
		if err == nil {
			return res, nil
		}
		if class != ClassUnavailable && class != ClassMisconfigured {
			return modelrouter.Result{}, err
		}
		last = err
	}
	return modelrouter.Result{}, fmt.Errorf("%w: %w", ErrNoAnswer, last)
}

// attempt sends req to m: the entry "sent", the call, then the entry of its
// result. On failure it returns the error and its class; an audit failure,
// which must end the call, has the class ClassInternal.
func (r *Router) attempt(ctx context.Context, role modelrouter.Role, actor Actor, m modelrouter.Ref, req *modelrouter.Request) (modelrouter.Result, ErrorClass, error) {
	id, idErr := r.callIDs()
	if idErr != nil {
		return modelrouter.Result{}, ClassInternal, fmt.Errorf("%w: a call ID: %w", ErrAuditUnavailable, idErr)
	}
	sent := CallSent{At: r.now(), Actor: actor, Role: role, Model: m, CallID: id}
	if sentErr := r.audit.RecordSent(ctx, &sent); sentErr != nil {
		return modelrouter.Result{}, ClassInternal, fmt.Errorf("%w: %w", ErrAuditUnavailable, sentErr)
	}
	res, callErr := r.providers.Generate(ctx, m.Instance, m.Model, req)
	if callErr == nil {
		res.Model = m
		if resErr := r.result(ctx, &sent, OutcomeOK, ""); resErr != nil {
			return modelrouter.Result{}, ClassInternal, fmt.Errorf("%w: %w", ErrAuditUnavailable, resErr)
		}
		return res, "", nil
	}
	class := classify(callErr)
	err := fmt.Errorf("model %s: %w", m, callErr)
	if resErr := r.result(ctx, &sent, OutcomeFailed, class); resErr != nil {
		err = fmt.Errorf("%w (the failure could not be audited: %w)", err, resErr)
	}
	if class == ClassMisconfigured {
		if raiseErr := r.flag(ctx, role, m, callErr); raiseErr != nil {
			err = fmt.Errorf("%w (its attention item could not be raised: %w)", err, raiseErr)
		}
	}
	return modelrouter.Result{}, class, err
}

// classify returns the error class of a failed call, by the error's kind and
// never its text.
func classify(err error) ErrorClass {
	switch {
	case errors.Is(err, ErrProviderMissing), errors.Is(err, ErrNotModelProvider):
		return ClassMisconfigured
	case errors.Is(err, ErrProviderUnavailable):
		return ClassUnavailable
	case errors.Is(err, ErrProviderRejected):
		return ClassRejected
	case errors.Is(err, context.Canceled):
		return ClassCanceled
	}
	return ClassInternal
}

// flag raises the attention item of the route entry m of role, when err says
// its provider instance is missing or not a model provider. A failure to
// raise it does not end the call: the next model still answers, and the
// periodic check (RouteProblems) raises the item again.
func (r *Router) flag(ctx context.Context, role modelrouter.Role, m modelrouter.Ref, err error) error {
	reason := modelrouter.ProblemMissing
	if errors.Is(err, ErrNotModelProvider) {
		reason = modelrouter.ProblemNotModelProvider
	}
	p := modelrouter.RouteProblem{Role: role, Model: m, Reason: reason}
	if raiseErr := r.attention.Raise(ctx, modelrouter.AttentionKind, p.Subject(), p.Detail()); raiseErr != nil {
		return fmt.Errorf("raising the item of a broken route: %w", raiseErr)
	}
	return nil
}

// refuse records a refusal, as a result entry alone: no call was sent. It
// returns err, joined with the failure to record it when that fails: the
// refusal stands either way.
func (r *Router) refuse(ctx context.Context, role modelrouter.Role, actor Actor, err error) error {
	id, idErr := r.callIDs()
	if idErr != nil {
		return fmt.Errorf("%w (the refusal could not be audited: %w)", err, idErr)
	}
	res := CallResult{At: r.now(), Actor: actor, Role: role, CallID: id, Outcome: OutcomeRefused}
	if auditErr := r.audit.RecordResult(ctx, &res); auditErr != nil {
		return fmt.Errorf("%w (the refusal could not be audited: %w)", err, auditErr)
	}
	return err
}

// result appends the entry that says how the call sent ended.
func (r *Router) result(ctx context.Context, sent *CallSent, outcome Outcome, class ErrorClass) error {
	res := CallResult{
		At: r.now(), Actor: sent.Actor, Role: sent.Role, Model: sent.Model, CallID: sent.CallID, Outcome: outcome, Class: class,
	}
	// The result is recorded even when the caller gave up: the call was sent.
	if err := r.audit.RecordResult(context.WithoutCancel(ctx), &res); err != nil {
		return fmt.Errorf("recording the result of a model call: %w", err)
	}
	return nil
}

// catalogOf returns the entries of the providers in route, by model, from
// the cache or the providers. A provider that can't list its entries
// contributes none: the router then leaves its models to answer for
// themselves.
func (r *Router) catalogOf(ctx context.Context, route modelrouter.Route) map[modelrouter.Ref]modelrouter.ModelEntry {
	out := map[modelrouter.Ref]modelrouter.ModelEntry{}
	seen := map[string]bool{}
	for _, m := range route.Models {
		if seen[m.Instance] {
			continue
		}
		seen[m.Instance] = true
		for id, e := range r.entriesOf(ctx, m.Instance) {
			out[modelrouter.Ref{Instance: m.Instance, Model: id}] = e
		}
	}
	return out
}

// entriesOf returns provider's entries by id, listing them again when the
// cache is older than catalogTTL.
func (r *Router) entriesOf(ctx context.Context, provider string) map[string]modelrouter.ModelEntry {
	now := r.now()
	r.mu.Lock()
	c, ok := r.catalogs[provider]
	r.mu.Unlock()
	if ok && now.Sub(c.listed) < catalogTTL {
		return c.entries
	}
	listed, err := r.providers.Models(ctx, provider)
	if err != nil {
		return c.entries // the stale entries, or none
	}
	c = catalog{entries: make(map[string]modelrouter.ModelEntry, len(listed)), listed: now}
	for _, e := range listed {
		c.entries[e.ID] = e
	}
	r.mu.Lock()
	r.catalogs[provider] = c
	r.mu.Unlock()
	return c.entries
}
