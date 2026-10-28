package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// seeding returns a synthetic seeding of acc-1 at go-live on day: one
// opening lot and the opening cash, keyed by day so that seedings on
// different days don't share a key.
func seeding(t *testing.T, day string) []ledger.Transaction {
	t.Helper()
	in := instrument(t, isinA, "TSTA")
	cash := money(t, "250", "USD")
	return []ledger.Transaction{
		must(t)(ledger.NewOpeningLot(header(t, acc1, "opening-lot:"+isinA+":2025-06-02:"+day, day), in, qty(t, "10"), money(t, "1500", "USD"), date(t, "2025-06-02"))),
		must(t)(ledger.NewOpeningCash(header(t, acc1, ledger.OpeningCashID(cash.Currency(), date(t, day)), day), cash)),
	}
}

func TestConcurrentSeedsOverPostgres(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	l := app.New(store)
	ctx := t.Context()
	days := []string{"2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28", "2026-03-29", "2026-03-30", "2026-03-31"}
	errs := make([]error, len(days))
	done := make(chan int)
	for i, day := range days {
		go func() {
			_, errs[i] = l.Seed(ctx, seeding(t, day), nil)
			done <- i
		}()
	}
	for range days {
		<-done
	}
	won := ""
	for i, err := range errs {
		if err == nil {
			if won != "" {
				t.Fatalf("seedings on %s and %s both succeeded", won, days[i])
			}
			won = days[i]
			continue
		}
		if seededErr, ok := errors.AsType[*ledger.AlreadySeededError](err); !ok || seededErr.Account != acc1 {
			t.Fatalf("Seed on %s: %v, want an *AlreadySeededError for acc-1", days[i], err)
		}
	}
	if won == "" {
		t.Fatal("no seeding succeeded")
	}
	w := seeding(t, won)
	log, err := store.Log(ctx, acc1)
	checkResult(t, "Log", []ledger.VersionedTransaction{versioned(&w[1], 1), versioned(&w[0], 1)}, log, err)

	again, err := l.Seed(ctx, w, nil)
	checkResult(t, "same Seed again", app.ImportSummary{Unchanged: 2, AlreadySeeded: []ledger.AccountID{acc1}}, again, err)
}

func TestLockAccount(t *testing.T) {
	t.Parallel()
	outside := postgres.New(nil).LockAccount(t.Context(), acc1)
	if diff := cmp.Diff(postgres.ErrNoTx, outside, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("LockAccount outside InTx (-want +got):\n%s", diff)
	}
	store, _ := seededStore(t)
	err := store.InTx(t.Context(), func(tx app.Store) error {
		if err := tx.LockAccount(t.Context(), acc1); err != nil {
			t.Errorf("LockAccount(%s): %v", acc1, err)
		}
		return tx.LockAccount(t.Context(), "acc-9")
	})
	if diff := cmp.Diff(app.ErrNotFound, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("LockAccount of an unknown account (-want +got):\n%s", diff)
	}
}
