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

// TestBondLotsOverPostgres stores what a manually entered bond account
// receives through the real store: opening lots one at a time, interest on
// one of them (D-101), and the deletion of another (D-100), whose superseded
// version stays in the log.
func TestBondLotsOverPostgres(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	l := app.New(store)
	ctx := t.Context()
	in := instrument(t, isinA, "COI0530")
	lot := func(opened, units string) ledger.Transaction {
		return must(t)(ledger.NewOpeningLot(header(t, acc1, "lot:COI0530:"+opened, "2026-06-01"), in, qty(t, units), money(t, units+"00", "PLN"), date(t, opened)))
	}
	first, second := lot("2026-05-04", "30"), lot("2026-05-18", "10")
	for _, tx := range []ledger.Transaction{first, second} {
		if _, err := l.AddOpeningLots(ctx, []ledger.Transaction{tx}, nil); err != nil {
			t.Fatalf("AddOpeningLots %s: %v", tx.Key, err)
		}
	}
	coupon := must(t)(ledger.NewInterest(header(t, acc1, "coupon:COI0530:2027-05-04", "2027-05-04"), in, money(t, "186.00", "PLN"), money(t, "35.34", "PLN")))
	gone := must(t)(ledger.NewDeletion(ledger.TransactionHeader{Key: second.Key, Account: acc1, Date: second.Date}))
	if _, err := l.Import(ctx, []ledger.Transaction{coupon, gone}, nil); err != nil {
		t.Fatalf("Import of the coupon and the deletion: %v", err)
	}
	versions, err := store.Versions(ctx, second.Key)
	if err != nil {
		t.Fatalf("Versions: %v", err)
	}
	if diff := cmp.Diff([]ledger.VersionedTransaction{versioned(&second, 1), versioned(&gone, 2)}, versions, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("versions of the deleted lot (-want +got):\n%s", diff)
	}
	current, err := l.CurrentTransactions(ctx, acc1)
	if err != nil {
		t.Fatalf("CurrentTransactions: %v", err)
	}
	if diff := cmp.Diff([]ledger.Transaction{first, coupon}, current, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("current transactions (-want +got):\n%s", diff)
	}
	late := must(t)(ledger.NewOpeningLot(header(t, acc1, "lot:COI0530:2026-05-25", "2026-06-02"), in, qty(t, "1"), money(t, "100", "PLN"), date(t, "2026-05-25")))
	if _, err := l.AddOpeningLots(ctx, []ledger.Transaction{late}, nil); !errors.Is(err, ledger.ErrGoLiveMismatch) {
		t.Errorf("AddOpeningLots off go-live: %v, want ErrGoLiveMismatch", err)
	}
}
