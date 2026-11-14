package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	portfolioconnect "github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	"github.com/arkadiuss/openfief/internal/portfolio/jobs"
	portfoliopg "github.com/arkadiuss/openfief/internal/portfolio/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// maintainerCallerID names the owner as the maintainer reads the portfolio for
// them: the weights it sends its model are the owner's whole portfolio.
const maintainerCallerID = "core-maintainer"

// maintainerJobs is the maintainer's River job: its weekly run and its owner's
// runs, on a queue of their own, which late fills with the maintainer once the
// core has built it (D-204).
func maintainerJobs(late *jobs.Late, cfg *serveConfig, log *slog.Logger) (valuationpg.Job, error) {
	worker, periodic, err := jobs.New(late, jobs.WeeklyTime(cfg.EndOfDay.Location), log)
	if err != nil {
		return valuationpg.Job{}, fmt.Errorf("the maintainer's job: %w", err)
	}
	return valuationpg.Job{
		Queue: jobs.Queue, MaxWorkers: 1,
		Register: func(w *river.Workers) error { return river.AddWorkerSafely(w, worker) },
		Periodic: []*river.PeriodicJob{periodic},
	}, nil
}

// maintainerGate is the maintainer's Gate over the models settings: the one
// switch and the maintainer's toggle (D-212).
type maintainerGate struct {
	settings modelsapp.SettingsStore
}

// MaintainerOn reports whether the switch and the maintainer's toggle are on.
func (g maintainerGate) MaintainerOn(ctx context.Context) (bool, error) {
	s, err := g.settings.Settings(ctx)
	if err != nil {
		return false, fmt.Errorf("loading the models settings: %w", err)
	}
	return s.Switch.Allows(modelrouter.FeatureMaintainer), nil
}

// maintainerAsker sends the maintainer's request through the model router, for
// its role: the instructions as public data (T2) and the figures as
// whole-portfolio data (T0, D-206), so only a model the owner allowed for
// both receives it, and the router audits the call and refuses what the
// switch, the toggle or the tiers do not allow.
type maintainerAsker struct {
	router *modelsapp.Router
}

// Ask asks the router, and describes a failure by its class alone.
func (a maintainerAsker) Ask(ctx context.Context, req *portfolioapp.AskRequest) (portfolioapp.AskResult, error) {
	actor := modelsapp.Actor{Kind: modelsapp.ActorSystem}
	if req.Actor != "" {
		actor = modelsapp.Actor{Kind: modelsapp.ActorOwner, ID: req.Actor}
	}
	res, err := a.router.Generate(ctx, modelrouter.RoleMaintainer, actor, &modelrouter.Request{
		Messages: []modelrouter.Message{
			{Role: modelrouter.MessageSystem, Content: req.Instructions, DataTier: modelrouter.DataTierT2},
			{Role: modelrouter.MessageUser, Content: req.Figures, DataTier: modelrouter.DataTierT0},
		},
		ResponseSchemaJSON: req.Schema,
	})
	if err != nil {
		return portfolioapp.AskResult{}, &portfolioapp.AskError{Class: askClass(err)}
	}
	return portfolioapp.AskResult{Content: res.Content, Model: res.Model.String()}, nil
}

// askClass is the class a failed call of the router ends a maintainer run in:
// the router's refusals are "denied", and the others are the classes of the
// audit log (identity.ModelErrorClasses). It reads the error's kind and never
// its text.
func askClass(err error) string {
	switch {
	case errors.Is(err, modelrouter.ErrSwitchOff), errors.Is(err, modelrouter.ErrFeatureOff), errors.Is(err, modelrouter.ErrNoRoute),
		errors.Is(err, modelrouter.ErrDataTierDenied), errors.Is(err, modelrouter.ErrNoCapableModel), errors.Is(err, modelrouter.ErrUntagged):
		return portfolioapp.ClassDenied
	case errors.Is(err, modelsapp.ErrProviderMissing), errors.Is(err, modelsapp.ErrNotModelProvider):
		return "misconfigured"
	case errors.Is(err, modelsapp.ErrProviderRejected):
		return "rejected"
	case errors.Is(err, modelsapp.ErrNoAnswer), errors.Is(err, modelsapp.ErrProviderUnavailable), errors.Is(err, context.DeadlineExceeded):
		return "unavailable"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}
	return portfolioapp.ClassInternal
}

// maintainerHoldings reads what the portfolio holds from the views, as the
// owner (identity.HolderOwner): the whole portfolio the model portfolio is
// compared with. Holdings left out of the total for want of a price would
// skew every weight, so an incomplete valuation is no figures at all.
type maintainerHoldings struct {
	views heldValuesReader
}

// heldValuesReader is what the maintainer reads of the views;
// *viewsapp.Views implements it.
type heldValuesReader interface {
	HeldValues(ctx context.Context) (viewsapp.HeldValues, error)
}

// Holdings returns the valued holdings, or an error wrapping
// portfolioapp.ErrNoHoldings when nothing is recorded or part of it is not
// valued.
func (h maintainerHoldings) Holdings(ctx context.Context) (portfolioapp.Holdings, error) {
	owner := identity.Caller{Holder: identity.HolderOwner, ID: maintainerCallerID}
	held, err := h.views.HeldValues(identityapp.WithCaller(ctx, owner))
	switch {
	case errors.Is(err, viewsapp.ErrNothingHeld):
		return portfolioapp.Holdings{}, fmt.Errorf("%w: %w", portfolioapp.ErrNoHoldings, err)
	case err != nil:
		return portfolioapp.Holdings{}, fmt.Errorf("reading the holdings: %w", err)
	case held.Unvalued > 0:
		return portfolioapp.Holdings{}, fmt.Errorf("%w: %d holdings or balances lack a value", portfolioapp.ErrNoHoldings, held.Unvalued)
	}
	out := portfolioapp.Holdings{
		AsOf: held.On, Currency: held.Total.Currency(), Total: held.Total.Amount(),
		Values: make(map[string]kernel.Decimal, len(held.ByInstrument)),
	}
	for k, v := range held.ByInstrument {
		out.Values[k] = v.Amount()
	}
	return out, nil
}

// maintainerAudit writes the maintainer's audit entries (D-211) over the
// identity context's log: the run, a proposal dropped, and the owner applying
// or dismissing one. They name the role and the run, a class for a refusal or
// a failure, and never a note, a reason or a figure.
type maintainerAudit struct {
	log auditAppender
}

// Record appends the entry of r.
func (a maintainerAudit) Record(ctx context.Context, r *portfolioapp.AuditRecord) error {
	who := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	if r.Actor != "" {
		who = identity.Actor{Kind: identity.ActorOwner, ID: r.Actor}
	}
	action, refs, ok := maintainerEntry(r)
	if !ok {
		return fmt.Errorf("%w: an event of the maintainer", identity.ErrInvalidAuditEntry)
	}
	outcome, err := identity.ParseAuditOutcome(string(r.Outcome))
	if err != nil {
		return fmt.Errorf("the maintainer's audit entry: %w", err)
	}
	e, err := identity.NewAuditEntry(r.At, who, action, outcome, refs...)
	if err != nil {
		return fmt.Errorf("building the maintainer's audit entry: %w", err)
	}
	if err := a.log.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("writing the maintainer's audit entry: %w", err)
	}
	return nil
}

// maintainerEntry returns the audit action of r's event and the records its
// entry names: the run, and for a run its role and, when it was refused or
// failed, its class. ok is false for an event the maintainer has no entry for.
func maintainerEntry(r *portfolioapp.AuditRecord) (identity.AuditAction, []identity.Ref, bool) {
	actions := map[portfolioapp.AuditEvent]identity.AuditAction{
		portfolioapp.AuditRun:       identity.ActionMaintainerRun,
		portfolioapp.AuditDropped:   identity.ActionMaintainerProposalDropped,
		portfolioapp.AuditApplied:   identity.ActionMaintainerProposalApplied,
		portfolioapp.AuditDismissed: identity.ActionMaintainerProposalDismissed,
	}
	action, ok := actions[r.Event]
	run := identity.Ref{Kind: identity.RefMaintainerRun, ID: r.RunID}
	if r.Event != portfolioapp.AuditRun {
		return action, []identity.Ref{run}, ok
	}
	refs := []identity.Ref{{Kind: identity.RefModelRole, ID: string(modelrouter.RoleMaintainer)}, run}
	if r.Class != "" {
		refs = append(refs, identity.Ref{Kind: identity.RefModelError, ID: r.Class})
	}
	return action, refs, ok
}

// maintainerSessions tells who a session is to the maintainer's note: the
// owner, or someone who may view every account (D-163), since the note speaks
// of the whole portfolio. The caller a session stands for is the read API's
// (identityapp.Principal.Caller), asked of the authorizer account by account.
type maintainerSessions struct {
	auth    sessions
	origins identity.Origins
	ledger  accountLister
	authz   identityapp.Authorizer
}

// Reader returns the session of sessionToken on origin as the maintainer sees
// it, or an error wrapping portfolioapp.ErrNoSession for an origin off the
// allow-list and a session that is missing, unknown, expired or revoked.
func (s maintainerSessions) Reader(ctx context.Context, origin, sessionToken string) (portfolioapp.Reader, error) {
	allowed, err := s.origins.Allowed(origin)
	if err != nil {
		return portfolioapp.Reader{}, fmt.Errorf("%w: %w", portfolioapp.ErrNoSession, err)
	}
	sess, err := s.auth.Session(ctx, allowed, sessionToken, identity.SessionScopeRead)
	switch {
	case errors.Is(err, identityapp.ErrInvalidSession):
		return portfolioapp.Reader{}, fmt.Errorf("%w: %w", portfolioapp.ErrNoSession, err)
	case err != nil:
		return portfolioapp.Reader{}, fmt.Errorf("checking a session: %w", err)
	}
	out := portfolioapp.Reader{SessionID: string(sess.ID), Owner: sess.Scope == identity.SessionScopeOwner}
	if out.Owner {
		out.ViewsAll = true
		return out, nil
	}
	accounts, err := s.ledger.Accounts(ctx)
	if err != nil {
		return portfolioapp.Reader{}, fmt.Errorf("loading the accounts: %w", err)
	}
	principal := identityapp.Principal{Session: sess.ID, Demo: sess.Authenticator == identity.AuthenticatorDemo}
	out.ViewsAll = s.viewsEvery(ctx, accounts, principal.Caller())
	return out, nil
}

// viewsEvery reports whether the authorizer lets c view every one of
// accounts. An authorizer that can't tell is a refusal: the note is not shown.
func (s maintainerSessions) viewsEvery(ctx context.Context, accounts []ledger.Account, c identity.Caller) bool {
	if lister, ok := s.authz.(identityapp.AccountLister); ok {
		set, err := lister.ActableAccounts(ctx, c, identity.ActionView)
		if err != nil {
			return false
		}
		for i := range accounts {
			if _, ok := set[string(accounts[i].ID)]; !ok {
				return false
			}
		}
		return true
	}
	for i := range accounts {
		if ok, err := s.authz.MayAct(ctx, c, identity.ActionView, string(accounts[i].ID)); err != nil || !ok {
			return false
		}
	}
	return true
}

// revaluation is the core's River client with the maintainer's job on it, and
// the late-bound maintainer that job calls.
type revaluation struct {
	client *river.Client[pgx.Tx]
	late   *jobs.Late
}

// newRevaluation builds the River client of the revaluation workers (D-36)
// with the maintainer's job on its own queue (D-204); newMaintainer fills in
// the maintainer the job calls.
func newRevaluation(pool *pgxpool.Pool, vstore *valuationpg.Store, cfg *serveConfig, attn valuationapp.AttentionRaiser,
	now func() time.Time, log *slog.Logger,
) (revaluation, error) {
	late := &jobs.Late{}
	job, err := maintainerJobs(late, cfg, log)
	if err != nil {
		return revaluation{}, err
	}
	// The audit entries the changes enqueued are written to the log by a
	// worker on this client too (D-211).
	client, err := revaluationClient(pool, vstore, cfg, attn, now, log, job, auditJob(identitypg.New(pool)))
	if err != nil {
		return revaluation{}, err
	}
	return revaluation{client: client, late: late}, nil
}

// newMaintainer wires the model-portfolio maintainer (D-204) over its ports and
// hands it to the job that runs it.
func newMaintainer(pool *pgxpool.Pool, rv revaluation, agents *agentModels, attn *attention, views *viewsapp.Views,
	auth *identityapp.Auth, origins identity.Origins, l *ledgerapp.Ledger, authz identityapp.Authorizer, now func() time.Time,
) *portfolioapp.Maintainer {
	store := portfoliopg.New(pool)
	m := portfolioapp.NewMaintainer(&portfolioapp.MaintainerPorts{
		Runs: store, Proposals: store, Model: store, Gate: maintainerGate{settings: agents.store},
		Holdings: maintainerHoldings{views: views}, Asker: maintainerAsker{router: agents.router},
		Decisions: valuationAttention{items: attn.items}, Audit: maintainerAudit{log: identitypg.New(pool)},
		Queue:    jobs.Queued{Client: rv.client},
		Sessions: maintainerSessions{auth: auth, origins: origins, ledger: l, authz: authz},
		IDs:      newCallID, Now: now,
	})
	rv.late.Set(m)
	return m
}

// withPortfolioServices returns h with the portfolio context's two services
// mounted next to it: the ModelPortfolioService (D-52) and the
// MaintainerService (D-204).
func withPortfolioServices(h http.Handler, models *portfolioapp.Models, m *portfolioapp.Maintainer, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	return withMaintainerService(withModelService(h, models, callers, log), m, callers, log)
}

// withMaintainerService returns h with the MaintainerService mounted next to
// it (D-204): its calls go to the maintainer, admitted like the
// ModelPortfolioService's, view plugins only, every other path to h. Nothing
// may store its responses.
func withMaintainerService(h http.Handler, m *portfolioapp.Maintainer, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := portfolioconnect.NewMaintainer(m, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}
