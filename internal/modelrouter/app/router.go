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
	providers Providers
	audit     Auditor
	now       func() time.Time

	mu       sync.Mutex
	catalogs map[string]catalog
}

// catalog is one provider's listed entries and when they were listed.
type catalog struct {
	entries map[string]modelrouter.Entry
	listed  time.Time
}

// NewRouter returns a router over the owner's settings, the model-provider
// plugins and the audit log, reading the time from now.
func NewRouter(settings SettingsStore, providers Providers, audit Auditor, now func() time.Time) *Router {
	return &Router{settings: settings, providers: providers, audit: audit, now: now, catalogs: map[string]catalog{}}
}

// Generate asks the models of role's route for an answer to req, on behalf
// of actor.
//
// With the switch off, or role's feature off, nothing is called: it records
// the refusal and returns an error wrapping modelrouter.ErrSwitchOff or
// modelrouter.ErrFeatureOff, before it reads the request or reaches any
// provider. A request with an untagged piece of data, or for which no model
// of the route is allowed for every tier it holds, is refused the same way
// (modelrouter.ErrUntagged, modelrouter.ErrTierDenied, modelrouter.ErrNoCapableModel).
//
// Before each call it appends the call to the audit log; if that fails the
// call is not made and the error wraps ErrAuditUnavailable. A model that
// is unavailable (ErrProviderUnavailable) passes the request to the next
// model of the route, each with its own entry; a provider rejecting the
// request ends the call at once. When every candidate was unavailable the
// error wraps ErrNoAnswer.
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
	tiers, err := req.Tiers()
	if err != nil {
		return modelrouter.Result{}, r.refuse(ctx, role, actor, err)
	}
	route := stored.Routes[role]
	plan, err := modelrouter.MakePlan(&modelrouter.PlanInput{
		Switch: stored.Switch, Role: role, Route: route, Policy: stored.Policy,
		Tiers: tiers, Needs: req.Needs(), Catalog: r.catalogOf(ctx, route),
	})
	if err != nil {
		return modelrouter.Result{}, r.refuse(ctx, role, actor, err)
	}
	return r.try(ctx, role, actor, plan.Candidates, req)
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
		if err := r.record(ctx, role, actor, m, OutcomeCalled); err != nil {
			return modelrouter.Result{}, fmt.Errorf("%w: %w", ErrAuditUnavailable, err)
		}
		res, err := r.providers.Generate(ctx, m.Provider, m.Model, req)
		if err == nil {
			res.Model = m
			return res, nil
		}
		err = fmt.Errorf("model %s: %w", m, err)
		if auditErr := r.record(ctx, role, actor, m, OutcomeFailed); auditErr != nil {
			err = fmt.Errorf("%w (the failure could not be audited: %w)", err, auditErr)
		}
		if !errors.Is(err, ErrProviderUnavailable) {
			return modelrouter.Result{}, err
		}
		last = err
	}
	return modelrouter.Result{}, fmt.Errorf("%w: %w", ErrNoAnswer, last)
}

// refuse records a refusal and returns err, joined with the failure to
// record it when that fails: the refusal stands either way.
func (r *Router) refuse(ctx context.Context, role modelrouter.Role, actor Actor, err error) error {
	if auditErr := r.record(ctx, role, actor, modelrouter.Ref{}, OutcomeRefused); auditErr != nil {
		return fmt.Errorf("%w (the refusal could not be audited: %w)", err, auditErr)
	}
	return err
}

// record appends one call to the audit log.
func (r *Router) record(ctx context.Context, role modelrouter.Role, actor Actor, m modelrouter.Ref, outcome Outcome) error {
	rec := CallRecord{At: r.now(), Actor: actor, Role: role, Model: m, Outcome: outcome}
	if err := r.audit.RecordCall(ctx, &rec); err != nil {
		return fmt.Errorf("recording a model call: %w", err)
	}
	return nil
}

// catalogOf returns the entries of the providers in route, by model, from
// the cache or the providers. A provider that can't list its entries
// contributes none: the router then leaves its models to answer for
// themselves.
func (r *Router) catalogOf(ctx context.Context, route modelrouter.Route) map[modelrouter.Ref]modelrouter.Entry {
	out := map[modelrouter.Ref]modelrouter.Entry{}
	seen := map[string]bool{}
	for _, m := range route.Models {
		if seen[m.Provider] {
			continue
		}
		seen[m.Provider] = true
		for id, e := range r.entriesOf(ctx, m.Provider) {
			out[modelrouter.Ref{Provider: m.Provider, Model: id}] = e
		}
	}
	return out
}

// entriesOf returns provider's entries by id, listing them again when the
// cache is older than catalogTTL.
func (r *Router) entriesOf(ctx context.Context, provider string) map[string]modelrouter.Entry {
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
	c = catalog{entries: make(map[string]modelrouter.Entry, len(listed)), listed: now}
	for _, e := range listed {
		c.entries[e.ID] = e
	}
	r.mu.Lock()
	r.catalogs[provider] = c
	r.mu.Unlock()
	return c.entries
}
