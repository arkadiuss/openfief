package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsconnect "github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsrpc "github.com/arkadiuss/openfief/internal/modelrouter/connectrpc"
	modelspg "github.com/arkadiuss/openfief/internal/modelrouter/postgres"
)

// modelProviders is the model router's Providers (D-205): the model-provider
// plugins' client, which runConnectors sets once discovery has found the
// instances. Until then no provider is reachable, which the router treats
// as an unavailable one.
type modelProviders struct {
	client atomic.Pointer[modelsrpc.Client]
}

var _ modelsapp.Providers = (*modelProviders)(nil)

// Describe returns where provider's models run and its retention terms.
func (p *modelProviders) Describe(ctx context.Context, provider string) (modelsdomain.Provider, error) {
	c, err := p.current(provider)
	if err != nil {
		return modelsdomain.Provider{}, err
	}
	got, err := c.Describe(ctx, provider)
	if err != nil {
		return modelsdomain.Provider{}, fmt.Errorf("describing provider %s: %w", provider, err)
	}
	return got, nil
}

// Models lists provider's model entries.
func (p *modelProviders) Models(ctx context.Context, provider string) ([]modelsdomain.Entry, error) {
	c, err := p.current(provider)
	if err != nil {
		return nil, err
	}
	got, err := c.Models(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("listing the models of provider %s: %w", provider, err)
	}
	return got, nil
}

// Generate asks provider's model for an answer.
func (p *modelProviders) Generate(ctx context.Context, provider, model string, req *modelsdomain.Request) (modelsdomain.Result, error) {
	c, err := p.current(provider)
	if err != nil {
		return modelsdomain.Result{}, err
	}
	got, err := c.Generate(ctx, provider, model, req)
	if err != nil {
		return modelsdomain.Result{}, fmt.Errorf("generating with provider %s: %w", provider, err)
	}
	return got, nil
}

// set makes c the client of the providers.
func (p *modelProviders) set(c *modelsrpc.Client) { p.client.Store(c) }

// current returns the client, or an error wrapping ErrProviderUnavailable
// while discovery has not started.
func (p *modelProviders) current(provider string) (*modelsrpc.Client, error) {
	c := p.client.Load()
	if c == nil {
		return nil, fmt.Errorf("%w: provider %s: the core has not found the plugins yet", modelsapp.ErrProviderUnavailable, provider)
	}
	return c, nil
}

// modelAudit is the model router's Auditor over the identity context's
// audit log (D-211): each call is an entry of the action model_call naming
// its role and model, never a request, an answer or an amount.
type modelAudit struct {
	log auditAppender
}

// auditAppender appends an entry to the audit log; *identitypg.Store
// implements it.
type auditAppender interface {
	AppendAudit(ctx context.Context, e *identity.AuditEntry) error
}

var _ modelsapp.Auditor = modelAudit{}

// RecordCall appends rec to the audit log, failing when it can't.
func (a modelAudit) RecordCall(ctx context.Context, rec *modelsapp.CallRecord) error {
	actor := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	if rec.Actor.Kind == modelsapp.ActorOwner {
		actor = identity.Actor{Kind: identity.ActorOwner, ID: rec.Actor.ID}
	}
	refs := []identity.Ref{{Kind: identity.RefModelRole, ID: string(rec.Role)}}
	if !rec.Model.IsZero() {
		refs = append(refs, identity.Ref{Kind: identity.RefModel, ID: rec.Model.String()})
	}
	var outcome identity.AuditOutcome
	switch rec.Outcome {
	case modelsapp.OutcomeCalled:
		outcome = identity.OutcomeOK
	case modelsapp.OutcomeRefused:
		outcome = identity.OutcomeRefused
	case modelsapp.OutcomeFailed:
		outcome = identity.OutcomeFailed
	default:
		return fmt.Errorf("%w: an outcome of a model call", identity.ErrInvalidAuditEntry)
	}
	e, err := identity.NewAuditEntry(rec.At, actor, identity.ActionModelCall, outcome, refs...)
	if err != nil {
		return fmt.Errorf("building the audit entry of a model call: %w", err)
	}
	if err := a.log.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("writing the audit entry of a model call: %w", err)
	}
	return nil
}

// agentModels is the wired models context: the owner's settings (D-212) and the
// router (D-205), which the Ask box (D-203) and the maintainer (D-204) call
// for a model's answer once they are built; nothing calls it yet.
type agentModels struct {
	settings  *modelsapp.Settings
	router    *modelsapp.Router
	providers *modelProviders
}

// newAgentModels wires the models context over pool: the settings in the
// models schema, owner-only (D-212); the router over them, the plugins
// through providers, and the audit log (D-211).
func newAgentModels(pool *pgxpool.Pool, auth *identityapp.Auth, origins identity.Origins, now func() time.Time) *agentModels {
	store := modelspg.New(pool)
	owners := &ownerSessions{auth: auth, origins: origins, noSession: modelsapp.ErrNoSession, notOwner: modelsapp.ErrNotOwner}
	providers := &modelProviders{}
	return &agentModels{
		settings:  modelsapp.NewSettings(store, owners),
		router:    modelsapp.NewRouter(store, providers, modelAudit{log: identitypg.New(pool)}, now),
		providers: providers,
	}
}

// withModelSettingsService returns h with the ModelSettingsService mounted
// next to it (D-212): its calls go to the models settings' handler,
// admitted like the SettingsService's, view plugins only, every other path
// to h. Nothing may store its responses.
func withModelSettingsService(h http.Handler, settings *modelsapp.Settings, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := modelsconnect.NewSettings(settings, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
