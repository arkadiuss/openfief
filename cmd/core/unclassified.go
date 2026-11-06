package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// revaluationClient returns the River client of the revaluation workers
// (D-36), raising the attention items they find, such as an instrument the
// owner should classify for OKI's exemptions (D-132), through attn.
func revaluationClient(pool *pgxpool.Pool, vstore *valuationpg.Store, cfg *serveConfig, attn valuationapp.AttentionRaiser,
	now func() time.Time, log *slog.Logger,
) (*river.Client[pgx.Tx], error) {
	client, err := valuationpg.NewClient(pool, vstore, ledgerpg.New(pool), &valuationpg.Config{
		Concurrency: cfg.Workers,
		EndOfDay:    cfg.EndOfDay,
		Pack:        jurisdictionPack(),
		Today:       valuationpg.TodayIn(cfg.EndOfDay.Location, now),
		Logger:      log,
		Attention:   attn,
		Plugins:     jobTracing(),
	})
	if err != nil {
		return nil, fmt.Errorf("revaluation workers: %w", err)
	}
	return client, nil
}

// attentionItems is the attention context's port for a check that raises
// and clears one condition at a time (D-130); *attentionapp.Items
// implements it.
type attentionItems interface {
	Raise(ctx context.Context, cond attentiondomain.Condition) error
	Clear(ctx context.Context, kind attentiondomain.Kind, subject string) error
}

// valuationAttention adapts the attention items to valuation's
// AttentionRaiser, so the items a revaluation raises and clears, such as
// an instrument to classify (D-132), land in the attention store; ingest's
// port, ingest/app.AttentionRaiser, has the same shape, so it serves
// ingest's undeclared cross-check prices (D-141) and bond terms and index
// levels (D-179) too. The
// valuation context declares its port with plain strings and the
// attention context its own condition, so neither imports the other.
type valuationAttention struct {
	items attentionItems
}

// Raise raises the item of kind and subject with detail, or opens it
// again when its detail changed.
func (a valuationAttention) Raise(ctx context.Context, kind, subject, detail string) error {
	if err := a.items.Raise(ctx, attentiondomain.Condition{Kind: attentiondomain.Kind(kind), Subject: subject, Detail: detail}); err != nil {
		return fmt.Errorf("raising a valuation attention item: %w", err)
	}
	return nil
}

// Clear resolves the open or acknowledged item of kind and subject.
func (a valuationAttention) Clear(ctx context.Context, kind, subject string) error {
	if err := a.items.Clear(ctx, attentiondomain.Kind(kind), subject); err != nil {
		return fmt.Errorf("clearing a valuation attention item: %w", err)
	}
	return nil
}
