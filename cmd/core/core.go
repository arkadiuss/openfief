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
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
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
// River's, and identity's. Each takes its own advisory lock, so replicas
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
	return nil
}

// newTokens returns the script token use cases over pool.
func newTokens(pool *pgxpool.Pool, now func() time.Time) *identityapp.Tokens {
	return identityapp.NewTokens(identitypg.New(pool), now, nil)
}

// core is the wired core service: the read API's handler and the River
// client that runs the revaluation workers.
type core struct {
	api   http.Handler
	river *river.Client[pgx.Tx]
}

// newCore wires the core over a migrated database: the ledger with
// valuation's import hook (D-36), the River revaluation workers, the views
// and the read API with script tokens as its verifier (D-27). It stores
// the built-in sleeve kinds and sleeves when cfg asks, and records the bond
// basis, which revalues everything when it changed. Start the River client
// to run the workers.
func newCore(ctx context.Context, pool *pgxpool.Pool, cfg *serveConfig, now func() time.Time, log *slog.Logger) (*core, error) {
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	if cfg.EnsureDefaults {
		if defErr := l.EnsureDefaults(ctx); defErr != nil {
			return nil, fmt.Errorf("storing the default sleeves: %w", defErr)
		}
	}
	recorder := valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore)
	if _, basisErr := recorder.ApplyBondBasis(ctx, cfg.BondBasis); basisErr != nil {
		return nil, fmt.Errorf("bond basis: %w", basisErr)
	}
	client, err := valuationpg.NewClient(pool, vstore, store, &valuationpg.Config{
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
	api, err := connectapi.New(views, readTokens{tokens: newTokens(pool, now)}, connectapi.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("read API: %w", err)
	}
	return &core{api: api, river: client}, nil
}

// readTokens is the read API's token verifier: a read-scoped script token
// (D-27). Plugins' projected ServiceAccount tokens aren't accepted yet.
type readTokens struct {
	tokens *identityapp.Tokens
}

// VerifyToken accepts a script token with the read scope; an invalid one
// fails with an error wrapping viewsapp.ErrInvalidToken.
func (r readTokens) VerifyToken(ctx context.Context, token string) error {
	err := r.tokens.Verify(ctx, token, identity.ScopeRead)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, identityapp.ErrInvalidToken):
		return fmt.Errorf("%w: %w", viewsapp.ErrInvalidToken, err)
	}
	return fmt.Errorf("verifying a token: %w", err)
}
