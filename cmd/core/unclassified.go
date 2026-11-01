package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// revaluationClient returns the River client of the revaluation workers
// (D-36), raising the attention items they find through loggedAttention.
func revaluationClient(pool *pgxpool.Pool, vstore *valuationpg.Store, cfg *serveConfig, now func() time.Time, log *slog.Logger) (*river.Client[pgx.Tx], error) {
	client, err := valuationpg.NewClient(pool, vstore, ledgerpg.New(pool), &valuationpg.Config{
		Concurrency: cfg.Workers,
		EndOfDay:    cfg.EndOfDay,
		Pack:        jurisdictionPack(),
		Today:       valuationpg.TodayIn(cfg.EndOfDay.Location, now),
		Logger:      log,
		Attention:   loggedAttention{log: log},
	})
	if err != nil {
		return nil, fmt.Errorf("revaluation workers: %w", err)
	}
	return client, nil
}

// loggedAttention stands in for the attention store (D-130) until it is
// built: it only logs the items revaluations raise and clear, such as an
// instrument the owner should classify for OKI's exemptions (D-132). The
// subject is an ISIN, never an amount or an account number.
type loggedAttention struct {
	log *slog.Logger
}

// Raise logs an item to raise.
func (a loggedAttention) Raise(ctx context.Context, kind, subject, _ string) error {
	a.log.LogAttrs(ctx, slog.LevelInfo, "attention item", slog.String("kind", kind), slog.String("subject", subject))
	return nil
}

// Clear logs an item to clear.
func (a loggedAttention) Clear(ctx context.Context, kind, subject string) error {
	a.log.LogAttrs(ctx, slog.LevelDebug, "attention item cleared", slog.String("kind", kind), slog.String("subject", subject))
	return nil
}
