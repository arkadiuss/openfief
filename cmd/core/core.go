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

	attentionpg "github.com/arkadiuss/openfief/internal/attention/postgres"
	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	portfoliopg "github.com/arkadiuss/openfief/internal/portfolio/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewskube "github.com/arkadiuss/openfief/internal/views/kube"
)

// deps holds a database pool for one command.
type deps struct {
	pool *pgxpool.Pool
}

// withPool connects to dsn, runs fn and closes the pool.
func withPool(ctx context.Context, dsn string, fn func(d *deps) error) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer pool.Close()
	return fn(&deps{pool: pool})
}

// migrate applies every context's migrations: the ledger's, valuation's with
// River's, identity's, ingest's, portfolio's and attention's. Each takes its own advisory lock, so replicas
// starting at once apply each migration once.
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ledgerpg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	if err := valuationpg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	if err := identitypg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	if err := ingestpg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	if err := portfoliopg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	if err := attentionpg.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	return nil
}

// newTokens returns the script token use cases over pool.
func newTokens(pool *pgxpool.Pool, now func() time.Time) *identityapp.Tokens {
	return identityapp.NewTokens(identitypg.New(pool), now, nil)
}

// core is the wired core service: what each of its ports serves (D-115),
// the River client that runs the revaluation workers, and what the
// connector scheduler is built from.
type core struct {
	// browserPort serves the read API and, in a cluster, the front door
	// (D-81); the health endpoints are the server's own.
	browserPort http.Handler
	// pluginPort serves the read API, the AuthService and the owner
	// services, which only plugins call.
	pluginPort http.Handler
	river      *river.Client[pgx.Tx]
	deps       *connectorDeps
	// connectors is closed when the connector scheduler has stopped; nil
	// when it doesn't run.
	connectors chan struct{}
	// plugins reads the installed plugins for the read API (D-46, D-58);
	// nil outside a cluster.
	plugins *viewskube.Plugins
	// browser reads the view plugins that serve browsers through the
	// core's front door (D-81); nil outside a cluster.
	browser *viewskube.Browser
	// attention runs the checks that raise attention items and publishes
	// the outbox's events (D-130).
	attention *attention
}

// ingestDeps returns the ledger, with valuation's import hook (D-36), and
// valuation's input use cases over pool.
func ingestDeps(pool *pgxpool.Pool, vstore *valuationpg.Store) (*ledgerapp.Ledger, *valuationapp.InputRecorder) {
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	return l, valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore, jurisdictionPack())
}

// jurisdictionPack returns the jurisdiction pack (D-70): the Polish one,
// the only pack there is.
func jurisdictionPack() *pl.Pack { return pl.NewPack() }

// valuationSettings records the version of the jurisdiction pack's rules,
// which revalues everything when it changed (D-36), and returns the
// cross-check use cases (D-22, D-110..D-112), with the owner's sessions
// checked by owners.
func valuationSettings(ctx context.Context, recorder *valuationapp.InputRecorder, vstore *valuationpg.Store, owners valuationapp.Owners) (*valuationapp.CrossChecks, error) {
	if _, err := recorder.ApplyRules(ctx); err != nil {
		return nil, fmt.Errorf("rules version: %w", err)
	}
	return valuationapp.NewCrossChecks(vstore, owners), nil
}

// newCore wires the core over a migrated database: the ledger with
// valuation's import hook (D-36), the River revaluation workers, the views
// and the read API with sessions, script tokens and the view plugins callers
// accepts as its verifier (readTokens, D-27). It stores
// the built-in sleeve kinds and sleeves when cfg asks, and records the
// jurisdiction pack's rules version, which revalues everything when it
// changed. The browser port serves the read API alone; the plugin port
// serves it with the AuthService, the ModelPortfolioService (D-52), the
// BondLotService (D-9, D-49), the StatementService (D-7, D-49), the
// MarketDataService (D-94), the PriceCheckService (D-110, D-112), the
// onboarding services, the SettingsService (D-78, D-113), the
// TransferService (D-105) and the AttentionService (D-130), to the view
// plugins callers accepts (D-69, D-79, D-115). The read API
// lists the installed plugins from installed, and without it, outside a
// cluster, answers that they are unknown. Start the River client to run
// the workers.
func newCore(ctx context.Context, pool *pgxpool.Pool, cfg *serveConfig, now func() time.Time, log *slog.Logger,
	callers identityconnect.Callers, installed *viewskube.Plugins,
) (*core, error) {
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	l, recorder := ingestDeps(pool, vstore)
	if defErr := ensureDefaults(ctx, l, cfg); defErr != nil {
		return nil, defErr
	}
	auth, err := newAuth(pool, &cfg.Auth, now)
	if err != nil {
		return nil, err
	}
	checks, err := valuationSettings(ctx, recorder, vstore, valuationOwners(auth, cfg.Auth.Origins))
	if err != nil {
		return nil, err
	}
	attn := newAttention(pool, l, vstore, checks, attentionOwners(auth, cfg.Auth.Origins), now, log)
	client, err := revaluationClient(pool, vstore, cfg, valuationAttention{items: attn.items}, now, log)
	if err != nil {
		return nil, err
	}
	models, err := newModels(pool, vstore, auth, cfg.Auth.Origins, cfg.Reporting)
	if err != nil {
		return nil, err
	}
	// authz is the one authorizer every read path and owner service asks
	// (D-163), until SpiceDB answers (D-164).
	authz := identityapp.Policy{}
	opts := append(viewOptions(models, checks, installed), viewsapp.WithConnections(ingestpg.New(pool)),
		viewsapp.WithJurisdiction(jurisdictionPack(), vstore), viewsapp.WithAttention(attn.items), viewsapp.WithAuthorizer(authz))
	views, err := viewsapp.New(l, vstore, vstore, vstore, cfg.Reporting, opts...)
	if err != nil {
		return nil, fmt.Errorf("views: %w", err)
	}
	api, err := connectapi.New(views, readTokens{auth: auth, plugins: callers}, connectapi.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("read API: %w", err)
	}
	deps := &connectorDeps{
		pool: pool, ledger: l, recorder: recorder, crossChecks: checks, now: now, log: log,
		today:   valuationpg.TodayIn(cfg.EndOfDay.Location, now),
		entries: &entryCaller{}, bonds: vstore, model: models,
		rankings: newRankings(pool, l, vstore, valuationOwners(auth, cfg.Auth.Origins)),
	}
	plugin := withModelService(withAuthService(api, auth, callers, log), models, callers, log)
	owners := ingestOwners(auth, cfg.Auth.Origins, authz)
	plugin = withBondLotService(plugin, newBondLots(deps, owners), callers, log)
	plugin = withStatementService(plugin, newUploads(deps, owners), newTickers(deps, owners), callers, log)
	plugin = withMarketDataService(plugin, deps.rankings, callers, log)
	plugin = withPriceCheckService(plugin, checks, callers, log)
	plugin = withOnboardingServices(plugin, newSleeves(l, auth, cfg.Auth.Origins),
		newAccounts(deps, owners, accountTerms{recorder, vstore}), callers, log)
	plugin = withSettingsService(plugin, newSettings(recorder, vstore, valuationOwners(auth, cfg.Auth.Origins), cfg), callers, log)
	plugin = withAttentionService(withTransferService(plugin, newTransfers(l, ledgerOwners(auth, cfg.Auth.Origins, authz)), callers, log), attn.items, callers, log)
	return &core{browserPort: api, pluginPort: plugin, river: client, deps: deps, plugins: installed, attention: attn}, nil
}

// ensureDefaults stores the built-in sleeve kinds and sleeves when cfg
// asks.
func ensureDefaults(ctx context.Context, l *ledgerapp.Ledger, cfg *serveConfig) error {
	if !cfg.EnsureDefaults {
		return nil
	}
	if err := l.EnsureDefaults(ctx); err != nil {
		return fmt.Errorf("storing the default sleeves: %w", err)
	}
	return nil
}

// viewOptions configures the views with the model portfolio's reports, the
// price checks (D-22) and, when installed is not nil, the installed plugins
// (D-46, D-58).
func viewOptions(models viewsapp.ModelReports, checks viewsapp.PriceChecks, installed *viewskube.Plugins) []viewsapp.Option {
	opts := []viewsapp.Option{viewsapp.WithModel(models), viewsapp.WithPriceChecks(checks)}
	if installed != nil {
		opts = append(opts, viewsapp.WithPlugins(installed))
	}
	return opts
}

// tokenAuthenticator checks the owner's bearer tokens;
// *identityapp.Auth implements it.
type tokenAuthenticator interface {
	Authenticate(ctx context.Context, token string) (identityapp.Principal, error)
}

// readTokens is the read API's token verifier: the owner's session, of any
// scope, or a read-scoped script token, both through
// identityapp.Auth.Authenticate; otherwise a view plugin's projected
// ServiceAccount token with audience openfief-core (D-27), the same callers
// the AuthService serves. The read API only reads, so a view plugin can
// only read. It returns the caller the token stands for, whom the views
// ask the authorizer about (D-163).
type readTokens struct {
	auth    tokenAuthenticator
	plugins identityconnect.Callers
}

// VerifyToken accepts a session, a read-scoped script token or a view
// plugin's token, returning its caller (identityapp.Principal.Caller, or the
// view plugin by name); an invalid one fails with an error wrapping
// viewsapp.ErrInvalidToken.
func (r readTokens) VerifyToken(ctx context.Context, token string) (identity.Caller, error) {
	p, err := r.auth.Authenticate(ctx, token)
	switch {
	case err == nil:
		return p.Caller(), nil
	case !errors.Is(err, identityapp.ErrInvalidToken):
		return identity.Caller{}, fmt.Errorf("verifying a token: %w", err)
	case r.plugins == nil:
		return identity.Caller{}, fmt.Errorf("%w: %w", viewsapp.ErrInvalidToken, err)
	}
	plugin, perr := r.plugins.VerifyCaller(ctx, token)
	switch {
	case perr == nil:
		return identity.Caller{Holder: identity.HolderViewPlugin, ID: plugin}, nil
	case errors.Is(perr, identityconnect.ErrCallerRefused):
		return identity.Caller{}, fmt.Errorf("%w: %w", viewsapp.ErrInvalidToken, err)
	}
	return identity.Caller{}, fmt.Errorf("verifying a plugin token: %w", perr)
}
