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
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
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

// withPool connects to dsn (newPool), runs fn and closes the pool.
func withPool(ctx context.Context, dsn string, fn func(d *deps) error) error {
	pool, err := newPool(ctx, dsn)
	if err != nil {
		return err
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

// historyOptions has the views read the connectors' sync history and import
// batches from the ingest store, the coupons and corrections from the
// ledger store and instruments' prices from vstore, for the owner's pages
// (D-193, D-195, D-196).
func historyOptions(pool *pgxpool.Pool, vstore *valuationpg.Store) viewsapp.Option {
	opts := []viewsapp.Option{
		viewsapp.WithSyncHistory(ingestpg.New(pool)), viewsapp.WithRecordedChanges(ledgerpg.New(pool)),
		viewsapp.WithInstrumentPrices(vstore),
	}
	return func(v *viewsapp.Views) {
		for _, o := range opts {
			o(v)
		}
	}
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
// TransferService (D-105), the AttentionService (D-130) and the
// CashCheckpointService (D-158), to the view
// plugins callers accepts (D-69, D-79, D-115). The read API
// lists the installed plugins from installed, and without it, outside a
// cluster, answers that they are unknown. Start the River client to run
// the workers.
func newCore(ctx context.Context, pool *pgxpool.Pool, cfg *serveConfig, now func() time.Time, log *slog.Logger,
	callers identityconnect.Callers, installed *viewskube.Plugins,
) (*core, error) {
	vstore, err := valuationpg.NewStore(pool, jobTracing()...)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	l, recorder := ingestDeps(pool, vstore)
	sleeves := portfolioapp.NewRegistry(portfoliopg.New(pool), l)
	if defErr := ensureDefaults(ctx, sleeves, cfg); defErr != nil {
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
	names := &ledgerapp.Names{Store: ledgerpg.New(pool), Owners: ledgerOwners(auth, cfg.Auth.Origins, authz), Postings: l}
	opts := append(viewOptions(models, checks, installed), viewsapp.WithConnections(ingestpg.New(pool)), viewsapp.WithInstrumentNames(names),
		viewsapp.WithJurisdiction(jurisdictionPack(), vstore), viewsapp.WithAttention(attn.items), viewsapp.WithAuthorizer(authz),
		viewsapp.WithPortfolio(sleeves), viewsapp.WithPayoutSettings(ingestpg.New(pool)), historyOptions(pool, vstore))
	views, err := viewsapp.New(l, vstore, vstore, vstore, jurisdictionPack().ReportingCurrency(), cfg.Reporting, opts...)
	if err != nil {
		return nil, fmt.Errorf("views: %w", err)
	}
	apis, err := readAPIs(views, auth, callers, log)
	if err != nil {
		return nil, err
	}
	deps := &connectorDeps{
		pool: pool, ledger: l, sleeves: sleeves, recorder: recorder, crossChecks: checks, now: now, log: log,
		today:   valuationpg.TodayIn(cfg.EndOfDay.Location, now),
		entries: &entryCaller{}, syncs: &syncTrigger{}, bonds: vstore, model: models,
		rankings:  newRankings(pool, l, vstore, valuationOwners(auth, cfg.Auth.Origins)),
		attention: attn.items, names: names,
	}
	plugin := withModelService(withAuthService(apis.plugin, auth, callers, log), models, callers, log)
	owners := ingestOwners(auth, cfg.Auth.Origins, authz)
	plugin = withBondLotService(plugin, newBondLots(deps, owners), callers, log)
	plugin = withStatementService(plugin, newUploads(deps, owners), newTickers(deps, owners), newRetries(deps, owners), callers, log)
	plugin = withMarketDataService(plugin, deps.rankings, callers, log)
	plugin = withPriceCheckService(plugin, checks, callers, log)
	plugin = withOnboardingServices(plugin, newSleeves(sleeves, auth, cfg.Auth.Origins),
		newAccounts(deps, owners, accountTerms{recorder, vstore}), auth, callers, log)
	plugin = withSettingsService(plugin, newSettings(recorder, vstore, valuationOwners(auth, cfg.Auth.Origins), cfg), callers, log)
	plugin = withLedgerEntries(plugin, l, ledgerOwners(auth, cfg.Auth.Origins, authz), attn.items, names, callers, log)
	return &core{browserPort: apis.browser, pluginPort: plugin, river: client, deps: deps, plugins: installed, attention: attn}, nil
}

// ensureDefaults stores the built-in sleeve kinds and sleeves when cfg
// asks.
func ensureDefaults(ctx context.Context, sleeves *portfolioapp.Registry, cfg *serveConfig) error {
	if !cfg.EnsureDefaults {
		return nil
	}
	if err := sleeves.EnsureDefaults(ctx); err != nil {
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

// readAPIs returns the read API for each port. The browser port's takes
// only sessions and script tokens; view plugins' tokens, the only ones a
// TokenReview checks, are taken on the plugin port alone, which only plugin
// pods reach (D-115, D-184).
func readAPIs(views *viewsapp.Views, auth tokenAuthenticator, callers identityconnect.Callers, log *slog.Logger) (portAPIs, error) {
	browser, err := connectapi.New(views, browserReadTokens(auth), connectapi.WithLogger(log))
	if err != nil {
		return portAPIs{}, fmt.Errorf("read API: %w", err)
	}
	plugin, err := connectapi.New(views, readTokens{auth: auth, plugins: callers}, connectapi.WithLogger(log))
	if err != nil {
		return portAPIs{}, fmt.Errorf("read API: %w", err)
	}
	return portAPIs{browser: browser, plugin: plugin}, nil
}

// portAPIs are the read API of the browser port and of the plugin port.
type portAPIs struct {
	browser, plugin http.Handler
}

// browserReadTokens is the browser port's read API verifier: the owner's
// sessions and script tokens alone. A view plugin calls the read API on the
// plugin port, so a token on the browser port never reaches a TokenReview,
// and nothing that reaches the browser port can make the core ask the API
// server (D-115, D-184).
func browserReadTokens(auth tokenAuthenticator) readTokens {
	return readTokens{auth: auth}
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
