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

	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
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
// River's, identity's and ingest's. Each takes its own advisory lock, so replicas
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
	return nil
}

// newTokens returns the script token use cases over pool.
func newTokens(pool *pgxpool.Pool, now func() time.Time) *identityapp.Tokens {
	return identityapp.NewTokens(identitypg.New(pool), now, nil)
}

// core is the wired core service: the read API's handler, the River
// client that runs the revaluation workers, and what the connector
// scheduler is built from.
type core struct {
	api   http.Handler
	river *river.Client[pgx.Tx]
	deps  *connectorDeps
	// connectors is closed when the connector scheduler has stopped; nil
	// when it doesn't run.
	connectors chan struct{}
}

// ingestDeps returns the ledger, with valuation's import hook (D-36), and
// valuation's input use cases over pool.
func ingestDeps(pool *pgxpool.Pool, vstore *valuationpg.Store) (*ledgerapp.Ledger, *valuationapp.InputRecorder) {
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	return l, valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore)
}

// newCore wires the core over a migrated database: the ledger with
// valuation's import hook (D-36), the River revaluation workers, the views
// and the read API with script tokens as its verifier (D-27). It stores
// the built-in sleeve kinds and sleeves when cfg asks, and records the bond
// basis, which revalues everything when it changed. Next to the read API it
// serves the AuthService to the view plugins callers accepts (author
// decision, auth split). Start the River client to run the workers.
func newCore(ctx context.Context, pool *pgxpool.Pool, cfg *serveConfig, now func() time.Time, log *slog.Logger, callers identityconnect.Callers) (*core, error) {
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	l, recorder := ingestDeps(pool, vstore)
	if cfg.EnsureDefaults {
		if defErr := l.EnsureDefaults(ctx); defErr != nil {
			return nil, fmt.Errorf("storing the default sleeves: %w", defErr)
		}
	}
	if _, basisErr := recorder.ApplyBondBasis(ctx, cfg.BondBasis); basisErr != nil {
		return nil, fmt.Errorf("bond basis: %w", basisErr)
	}
	client, err := valuationpg.NewClient(pool, vstore, ledgerpg.New(pool), &valuationpg.Config{
		Concurrency: cfg.Workers,
		EndOfDay:    cfg.EndOfDay,
		BondBasis:   cfg.BondBasis,
		Today:       valuationpg.TodayIn(cfg.EndOfDay.Location, now),
		Logger:      log,
	})
	if err != nil {
		return nil, fmt.Errorf("revaluation workers: %w", err)
	}
	views, err := viewsapp.New(l, vstore, vstore, vstore, cfg.Reporting)
	if err != nil {
		return nil, fmt.Errorf("views: %w", err)
	}
	auth, err := newAuth(pool, &cfg.Auth, now)
	if err != nil {
		return nil, err
	}
	api, err := connectapi.New(views, readTokens{auth: auth}, connectapi.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("read API: %w", err)
	}
	deps := &connectorDeps{
		pool: pool, ledger: l, recorder: recorder, now: now, log: log,
		today: valuationpg.TodayIn(cfg.EndOfDay.Location, now),
	}
	return &core{api: withAuthService(api, auth, callers, log), river: client, deps: deps}, nil
}

// readTokens is the read API's token verifier: the owner's session, of any
// scope, or a read-scoped script token (D-27), both through
// identityapp.Auth.Authenticate. Plugins' projected ServiceAccount tokens
// aren't accepted yet.
type readTokens struct {
	auth *identityapp.Auth
}

// VerifyToken accepts a session or a read-scoped script token; an invalid
// one fails with an error wrapping viewsapp.ErrInvalidToken.
func (r readTokens) VerifyToken(ctx context.Context, token string) error {
	_, err := r.auth.Authenticate(ctx, token)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, identityapp.ErrInvalidToken):
		return fmt.Errorf("%w: %w", viewsapp.ErrInvalidToken, err)
	}
	return fmt.Errorf("verifying a token: %w", err)
}
