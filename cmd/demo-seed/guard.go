package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// errBadMarker reports a demo marker whose seed or day doesn't parse.
var errBadMarker = errors.New("malformed demo marker")

// guardLock is the advisory lock key that serializes demo loaders on one
// database: "demo" in ASCII.
const guardLock int64 = 0x64656d6f

// dataTables are the tables whose rows count as ledger or valuation data:
// a database with any of them is not empty.
func dataTables() []string {
	return []string{
		"ledger.sleeves", "ledger.accounts", "ledger.transaction_versions",
		"valuation.prices", "valuation.fx_rates", "valuation.index_readings",
		"valuation.bond_terms", "valuation.tax_wrappers",
	}
}

// markerSQL creates the demo schema and its marker table. The table holds
// at most one row, demo = true, and only this loader writes it.
const markerSQL = `
CREATE SCHEMA IF NOT EXISTS demo;
CREATE TABLE IF NOT EXISTS demo.marker (
    demo      boolean PRIMARY KEY CHECK (demo),
    seed      text NOT NULL,
    loaded_on date NOT NULL,
    loaded_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE demo.marker ADD COLUMN IF NOT EXISTS loader text NOT NULL DEFAULT '';`

// resetSQL drops everything the demo loaded, the ledger and valuation
// schemas, which migrate creates again, and the revaluations still queued
// for it. River's own schema stays, so a core running on the database
// keeps its workers while the demo is reseeded under it (D-161); the loads
// that follow enqueue the revaluations again.
const resetSQL = `DROP SCHEMA IF EXISTS ledger, valuation CASCADE;
DO $$ BEGIN
    IF to_regclass('river.river_job') IS NOT NULL THEN
        DELETE FROM river.river_job WHERE state::text IN ('available', 'scheduled', 'retryable');
    END IF;
END $$;`

// prepare decides, under the guard lock and before anything is written,
// what to do with the database (demo.DecideLoad), and does its part: it
// marks an empty database as a demo database, or empties a marked one for
// a reset. A refused database is left untouched.
func prepare(ctx context.Context, pool *pgxpool.Pool, cfg *config) (demo.LoadAction, error) {
	var action demo.LoadAction
	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", guardLock); err != nil {
			return fmt.Errorf("taking the demo lock: %w", err)
		}
		st, err := inspect(ctx, tx)
		if err != nil {
			return err
		}
		if action, err = demo.DecideLoad(st, cfg.Reset, cfg.Loader); err != nil {
			return fmt.Errorf("demo guard: %w", err)
		}
		return mark(ctx, tx, action, cfg)
	})
	if err != nil {
		return 0, fmt.Errorf("checking the database: %w", err)
	}
	return action, nil
}

// inspect reads what prepare decides on: the demo marker and the rows of
// ledger and valuation data.
func inspect(ctx context.Context, tx pgx.Tx) (demo.DatabaseState, error) {
	var st demo.DatabaseState
	exists, err := tableExists(ctx, tx, "demo.marker")
	if err != nil {
		return st, err
	}
	if exists {
		// to_jsonb reads a marker written before it recorded its loader too.
		err := tx.QueryRow(ctx, `SELECT count(*) > 0, coalesce(max(to_jsonb(m) ->> 'loader'), '')
			FROM demo.marker m WHERE demo`).Scan(&st.Marked, &st.Loader)
		if err != nil {
			return st, fmt.Errorf("reading the demo marker: %w", err)
		}
	}
	for _, table := range dataTables() {
		exists, err := tableExists(ctx, tx, table)
		if err != nil {
			return st, err
		}
		if !exists {
			continue
		}
		var n int64
		// table comes from dataTables, never from input.
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			return st, fmt.Errorf("counting %s: %w", table, err)
		}
		st.Rows += n
	}
	return st, nil
}

// tableExists reports whether the schema-qualified table exists.
func tableExists(ctx context.Context, tx pgx.Tx, table string) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists); err != nil {
		return false, fmt.Errorf("looking for %s: %w", table, err)
	}
	return exists, nil
}

// mark does action's part in tx: it writes the demo marker for a fresh
// load, and for a reset also drops the demo's data.
func mark(ctx context.Context, tx pgx.Tx, action demo.LoadAction, cfg *config) error {
	switch action {
	case demo.LoadAgain:
		return nil
	case demo.ResetAndLoad:
		if _, err := tx.Exec(ctx, resetSQL); err != nil {
			return fmt.Errorf("resetting the demo database: %w", err)
		}
	case demo.LoadFresh:
	}
	if _, err := tx.Exec(ctx, markerSQL); err != nil {
		return fmt.Errorf("creating the demo marker: %w", err)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO demo.marker (demo, seed, loaded_on, loader) VALUES (true, $1, $2, $3)
		ON CONFLICT (demo) DO UPDATE SET seed = EXCLUDED.seed, loaded_on = EXCLUDED.loaded_on, loader = EXCLUDED.loader,
		    loaded_at = now()`,
		strconv.FormatUint(cfg.Seed, 10), cfg.Today.String(), cfg.Loader)
	if err != nil {
		return fmt.Errorf("writing the demo marker: %w", err)
	}
	return nil
}

// loadedDataset returns the seed and day of the dataset a marked database
// holds, as its marker records them.
func loadedDataset(ctx context.Context, pool *pgxpool.Pool) (uint64, kernel.Date, error) {
	var seed, day string
	if err := pool.QueryRow(ctx, "SELECT seed, loaded_on::text FROM demo.marker WHERE demo").Scan(&seed, &day); err != nil {
		return 0, kernel.Date{}, fmt.Errorf("reading the demo marker: %w", err)
	}
	s, err1 := strconv.ParseUint(seed, 10, 64)
	d, err2 := kernel.ParseDate(day)
	if err := errors.Join(err1, err2); err != nil {
		return 0, kernel.Date{}, fmt.Errorf("%w: %w", errBadMarker, err)
	}
	return s, d, nil
}
