package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsconnect "github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsrpc "github.com/arkadiuss/openfief/internal/modelrouter/connectrpc"
	modelspg "github.com/arkadiuss/openfief/internal/modelrouter/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// modelProviderInstances is the model router's ProviderInstances (D-205): the model-provider
// plugins' client, which runConnectors sets once discovery has found the
// instances. Until then no provider is reachable, which the router treats
// as an unavailable one.
type modelProviderInstances struct {
	client atomic.Pointer[modelsrpc.Client]
}

var _ modelsapp.ProviderInstances = (*modelProviderInstances)(nil)

// Instance reports whether instance is a running model-provider instance,
// without calling it. Until discovery has found the plugins it can't say, and
// the error wraps modelsapp.ErrPluginsNotFound.
func (p *modelProviderInstances) Instance(ctx context.Context, instance string) error {
	c, err := p.current(instance)
	if err != nil {
		return err
	}
	if err := c.Instance(ctx, instance); err != nil {
		return fmt.Errorf("checking provider instance %s: %w", instance, err)
	}
	return nil
}

// Describe returns where provider's models run and its retention terms.
func (p *modelProviderInstances) Describe(ctx context.Context, provider string) (modelsdomain.ProviderInstance, error) {
	c, err := p.current(provider)
	if err != nil {
		return modelsdomain.ProviderInstance{}, err
	}
	got, err := c.Describe(ctx, provider)
	if err != nil {
		return modelsdomain.ProviderInstance{}, fmt.Errorf("describing provider %s: %w", provider, err)
	}
	return got, nil
}

// Models lists provider's model entries.
func (p *modelProviderInstances) Models(ctx context.Context, provider string) ([]modelsdomain.ModelEntry, error) {
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
func (p *modelProviderInstances) Generate(ctx context.Context, provider, model string, req *modelsdomain.Request) (modelsdomain.Result, error) {
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
func (p *modelProviderInstances) set(c *modelsrpc.Client) { p.client.Store(c) }

// providerAddress returns base, the discovery's address of a model-provider
// instance, with the discovery's errors named as the models context's: a name
// no plugin has is a missing provider instance, and a plugin that is no model
// provider is one too, which the router answers by trying the next model and
// raising an attention item (D-196); a provider that does not run now stays
// unavailable.
func providerAddress(base func(instance string) (string, error)) func(instance string) (string, error) {
	return func(instance string) (string, error) {
		url, err := base(instance)
		switch {
		case errors.Is(err, ingestkube.ErrNoSuchInstance):
			return "", fmt.Errorf("%w: %w", modelsapp.ErrProviderMissing, err)
		case errors.Is(err, ingestkube.ErrNotModelProvider):
			return "", fmt.Errorf("%w: %w", modelsapp.ErrNotModelProvider, err)
		}
		return url, err
	}
}

// current returns the client, or an error wrapping ErrProviderUnavailable
// while discovery has not started.
func (p *modelProviderInstances) current(provider string) (*modelsrpc.Client, error) {
	c := p.client.Load()
	if c == nil {
		return nil, fmt.Errorf("%w: %w: provider %s", modelsapp.ErrProviderUnavailable, modelsapp.ErrPluginsNotFound, provider)
	}
	return c, nil
}

// modelAudit is the model router's Auditor over the identity context's audit
// log (D-211): each call is an entry of the action model_call, written before
// it, and an entry of the action model_call_result after it, linked by the
// call's ID. They name the role, the model and the call, and an error class
// for a failure, never a request, an answer or an amount.
type modelAudit struct {
	log auditAppender
}

// auditAppender appends an entry to the audit log; *identitypg.Store
// implements it.
type auditAppender interface {
	AppendAudit(ctx context.Context, e *identity.AuditEntry) error
}

var _ modelsapp.Auditor = modelAudit{}

// RecordSent appends the entry of a call about to be sent, failing when it
// can't: the router then does not make the call.
func (a modelAudit) RecordSent(ctx context.Context, c *modelsapp.CallSent) error {
	refs := []identity.Ref{
		{Kind: identity.RefModelRole, ID: string(c.Role)},
		{Kind: identity.RefModel, ID: c.Model.String()},
		{Kind: identity.RefModelCall, ID: c.CallID},
	}
	return a.append(ctx, c.At, c.Actor, identity.ActionModelCall, identity.OutcomeSent, refs)
}

// RecordResult appends the entry of how a call ended.
func (a modelAudit) RecordResult(ctx context.Context, c *modelsapp.CallResult) error {
	refs := []identity.Ref{{Kind: identity.RefModelRole, ID: string(c.Role)}}
	if !c.Model.IsZero() {
		refs = append(refs, identity.Ref{Kind: identity.RefModel, ID: c.Model.String()})
	}
	refs = append(refs, identity.Ref{Kind: identity.RefModelCall, ID: c.CallID})
	var outcome identity.AuditOutcome
	switch c.Outcome {
	case modelsapp.OutcomeOK:
		outcome = identity.OutcomeOK
	case modelsapp.OutcomeRefused:
		outcome = identity.OutcomeRefused
	case modelsapp.OutcomeFailed:
		outcome = identity.OutcomeFailed
		refs = append(refs, identity.Ref{Kind: identity.RefModelError, ID: string(c.Class)})
	default:
		return fmt.Errorf("%w: an outcome of a model call", identity.ErrInvalidAuditEntry)
	}
	return a.append(ctx, c.At, c.Actor, identity.ActionModelCallResult, outcome, refs)
}

// append writes one entry of actor's call.
func (a modelAudit) append(ctx context.Context, at time.Time, actor modelsapp.Actor, action identity.AuditAction,
	outcome identity.AuditOutcome, refs []identity.Ref,
) error {
	who := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	if actor.Kind == modelsapp.ActorOwner {
		who = identity.Actor{Kind: identity.ActorOwner, ID: actor.ID}
	}
	e, err := identity.NewAuditEntry(at, who, action, outcome, refs...)
	if err != nil {
		return fmt.Errorf("building the audit entry of a model call: %w", err)
	}
	if err := a.log.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("writing the audit entry of a model call: %w", err)
	}
	return nil
}

// newCallID returns a random call ID, 16 lower-case hex digits.
func newCallID() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// agentModels is the wired models context: the owner's settings (D-212) and the
// router (D-205), which the Ask box (D-203) and the maintainer (D-204) call
// for a model's answer once they are built; nothing calls it yet.
type agentModels struct {
	settings  *modelsapp.Settings
	router    *modelsapp.Router
	providers *modelProviderInstances

	store *modelspg.Store
	pool  *pgxpool.Pool
	now   func() time.Time
}

// newAgentModels wires the models context over pool: the settings in the
// models schema, owner-only (D-212), and the plugins through providers. The
// router needs the attention items, which need the route check this returns,
// so attend builds it.
func newAgentModels(pool *pgxpool.Pool, auth *identityapp.Auth, origins identity.Origins, now func() time.Time) *agentModels {
	store := modelspg.New(pool)
	owners := &ownerSessions{auth: auth, origins: origins, noSession: modelsapp.ErrNoSession, notOwner: modelsapp.ErrNotOwner}
	return &agentModels{
		settings: modelsapp.NewSettings(store, owners), providers: &modelProviderInstances{}, store: store, pool: pool, now: now,
	}
}

// newAgentsAndAttention wires the models context and the attention items, which
// need each other: the items run the models' route check (D-196), and the
// router raises its items on them. extra are the other checks that need the
// owner, such as the view plugins without a grant (D-196, D-209).
func newAgentsAndAttention(pool *pgxpool.Pool, auth *identityapp.Auth, origins identity.Origins, l *ledgerapp.Ledger,
	vstore *valuationpg.Store, checks *valuationapp.CrossChecks, now func() time.Time, log *slog.Logger, extra ...attentionapp.Check,
) (*agentModels, *attention) {
	agents := newAgentModels(pool, auth, origins, now)
	attn := newAttention(pool, l, vstore, checks, attentionOwners(auth, origins), now, log, append([]attentionapp.Check{agents.routeCheck()}, extra...)...)
	agents.attend(attn.items)
	return agents, attn
}

// routeCheck is the attention check of the routes in use (D-196): each route
// entry naming a provider instance no plugin has, or one that is no model
// provider, is a condition of the kind model-route-broken, which clears when
// the owner fixes the route or installs the plugin.
func (a *agentModels) routeCheck() attentionapp.Check { return routeCheckOver(a.store, a.providers) }

// routeCheckOver is routeCheck over settings and instances.
func routeCheckOver(settings modelsapp.SettingsStore, instances modelsapp.ProviderInstances) attentionapp.Check {
	kind := attentiondomain.Kind(modelsdomain.AttentionKind)
	return attentionapp.Check{Kind: kind, Find: func(ctx context.Context) ([]attentiondomain.Condition, error) {
		problems, err := modelsapp.RouteProblems(ctx, settings, instances)
		if err != nil {
			return nil, fmt.Errorf("checking the model routes: %w", err)
		}
		out := make([]attentiondomain.Condition, 0, len(problems))
		for _, p := range problems {
			out = append(out, attentiondomain.Condition{Kind: kind, Subject: p.Subject(), Detail: p.Detail()})
		}
		return out, nil
	}}
}

// attend builds the router over the settings, the plugins, the audit log and
// the attention items it raises broken routes on.
func (a *agentModels) attend(items attentionItems) {
	a.router = modelsapp.NewRouter(&modelsapp.RouterPorts{
		Settings: a.store, Providers: a.providers, Audit: modelAudit{log: identitypg.New(a.pool)},
		Attention: valuationAttention{items: items}, CallIDs: newCallID, Now: a.now,
	})
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
