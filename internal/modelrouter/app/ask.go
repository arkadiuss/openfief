package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// Askers checks the session of someone asking the Ask box a question (D-203):
// any valid session on an allowed origin, a read-scoped one included, since
// the demo's visitors ask too and what the model sees is limited by what the
// session may read. It returns who asked, as the audit entries name the actor:
// the owner or a viewer, by the session's ID; a viewer's question is the
// viewer's, never the owner's. cmd/core answers it over the identity context,
// with errors wrapping ErrNoSession.
type Askers interface {
	Asker(ctx context.Context, origin, sessionToken string) (Actor, error)
}

// Ask is the Ask box's door to the model router (D-203): for the Ask box's
// role alone, for the person whose session is valid. The tool loop runs in
// the built-in UI, which holds the read-only answerers the model may call
// and reads with the asking person's session; the router stays here, so
// every round goes through the switch, the route, the data tiers and the
// audit log.
//
// It holds no question or answer beyond a call's duration, and records and
// returns no text in an error (GR-12).
type Ask struct {
	askers   Askers
	settings SettingsStore
	router   *Router
}

// NewAsk returns the Ask box's door over router, with sessions checked by
// askers and the switch and route read from settings.
func NewAsk(askers Askers, settings SettingsStore, router *Router) *Ask {
	return &Ask{askers: askers, settings: settings, router: router}
}

// Available reports whether the Ask box has a model for the person with the
// session: the one switch and the Ask box's toggle are on and its route has a
// model (D-212). The data tiers decide per request. It writes no audit entry.
func (a *Ask) Available(ctx context.Context, origin, sessionToken string) (bool, error) {
	if _, err := a.asker(ctx, origin, sessionToken); err != nil {
		return false, err
	}
	stored, err := a.settings.Settings(ctx)
	if err != nil {
		return false, fmt.Errorf("loading the models settings: %w", err)
	}
	return stored.HasModel(modelrouter.RoleAsk), nil
}

// Permits reports whether a model of the Ask box's route may receive data of
// every one of tiers together, with tools (D-206), without calling any model
// or writing an audit entry. A tier that is none of T0, TF, T1 and T2 is an
// error wrapping ErrInvalidInput. A route that may not is false and no error:
// that is the answer, a refusal of the data and not a failure.
func (a *Ask) Permits(ctx context.Context, origin, sessionToken string, tiers []string) (bool, error) {
	if _, err := a.asker(ctx, origin, sessionToken); err != nil {
		return false, err
	}
	parsed := make([]modelrouter.DataTier, 0, len(tiers))
	for _, t := range tiers {
		tier, err := modelrouter.ParseDataTier(t)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		parsed = append(parsed, tier)
	}
	set, err := modelrouter.NewDataTierSet(parsed)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	err = a.router.Permits(ctx, modelrouter.RoleAsk, set, modelrouter.Needs{Tools: true})
	if refusal(err) {
		return false, nil
	}
	return err == nil, err
}

// Generate asks the Ask box's models for the next message of req on behalf of
// the person with the session, through the router (Router.Generate): the
// router's refusals, audit failures and the exhaustion of the route come back
// as its own errors.
func (a *Ask) Generate(ctx context.Context, origin, sessionToken string, req *modelrouter.Request) (modelrouter.Result, error) {
	who, err := a.asker(ctx, origin, sessionToken)
	if err != nil {
		return modelrouter.Result{}, err
	}
	res, err := a.router.Generate(ctx, modelrouter.RoleAsk, who, req)
	if err != nil {
		return modelrouter.Result{}, fmt.Errorf("asking the Ask box's model: %w", err)
	}
	return res, nil
}

// asker returns who asked with the session of sessionToken on origin.
func (a *Ask) asker(ctx context.Context, origin, sessionToken string) (Actor, error) {
	who, err := a.askers.Asker(ctx, origin, sessionToken)
	if err != nil {
		return Actor{}, fmt.Errorf("the Ask box: %w", err)
	}
	return who, nil
}

// refusal reports whether err is the router saying the data, not the
// settings or the providers, may not go: no model of the route may receive it
// or serves what it asks.
func refusal(err error) bool {
	return errors.Is(err, modelrouter.ErrDataTierDenied) || errors.Is(err, modelrouter.ErrNoCapableModel)
}
