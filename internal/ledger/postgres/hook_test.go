package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// errQueueDown stands in for a job queue that refuses an insert.
var errQueueDown = errors.New("queue down")

// errNotPostgres reports a hook given another Store than *postgres.Store.
var errNotPostgres = errors.New("hook did not get a *postgres.Store")

// enqueueHook inserts one row per affected account into public.jobs, a
// stand-in for River's job table, on the import's transaction, then fails
// with fail if set.
func enqueueHook(fail error) app.ImportHook {
	return func(ctx context.Context, tx app.Store, sum *app.ImportSummary) error {
		store, ok := tx.(*postgres.Store)
		if !ok {
			return errNotPostgres
		}
		err := store.WithTx(func(tx pgx.Tx) error {
			for account, from := range sum.EarliestAffected {
				if _, err := tx.Exec(ctx, "INSERT INTO public.jobs (account, from_date) VALUES ($1, $2)",
					string(account), from.String()); err != nil {
					return fmt.Errorf("enqueueing: %w", err)
				}
			}
			return fail
		})
		if err != nil {
			return fmt.Errorf("hook: %w", err)
		}
		return nil
	}
}

func TestImportHookEnqueuesInTheSameTransaction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fail     error
		wantErr  error
		wantJobs []string
		wantLog  int
	}{
		{name: "the job commits with the import", wantJobs: []string{"acc-1 2026-03-02"}, wantLog: 1},
		{name: "a failing enqueue rolls the import back", fail: errQueueDown, wantErr: errQueueDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store, pool := seededStore(t)
			if _, err := pool.Exec(t.Context(), "CREATE TABLE public.jobs (account text, from_date date)"); err != nil {
				t.Fatalf("creating jobs: %v", err)
			}
			deposit := must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1000", "USD")))
			l := app.New(store, enqueueHook(tt.fail))
			_, err := l.Import(t.Context(), []ledger.Transaction{deposit}, nil)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Import error (-want +got):\n%s", diff)
			}
			rows, err := pool.Query(t.Context(), "SELECT account || ' ' || to_char(from_date, 'YYYY-MM-DD') FROM public.jobs")
			if err != nil {
				t.Fatalf("reading jobs: %v", err)
			}
			jobs, err := pgx.CollectRows(rows, pgx.RowTo[string])
			checkResult(t, "jobs", tt.wantJobs, jobs, err)
			log, err := store.Log(t.Context(), acc1)
			if err != nil || len(log) != tt.wantLog {
				t.Errorf("Log = %d versions, %v; want %d", len(log), err, tt.wantLog)
			}
		})
	}
}

func TestWithTxNeedsATransaction(t *testing.T) {
	t.Parallel()
	store := postgres.New(nil)
	err := store.WithTx(func(pgx.Tx) error { return nil })
	if diff := cmp.Diff(postgres.ErrNoTx, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("WithTx outside InTx (-want +got):\n%s", diff)
	}
}
