package app_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// seeding returns the synthetic seeding of account at go-live on day: one
// opening lot of cost and the opening cash, keyed by a source per account,
// as connectors key them.
func seeding(t *testing.T, account ledger.AccountID, day, cost, cash string) []ledger.Transaction {
	t.Helper()
	in := instrument(t, "IETEST000026", "TSTB")
	lotHeader := header(t, account, "opening-lot:IETEST000026:2025-06-02:n1", day)
	cashHeader := header(t, account, ledger.OpeningCashID(money(t, cash, "USD").Currency(), date(t, day)), day)
	lotHeader.Key.Source = "test/" + string(account)
	cashHeader.Key.Source = lotHeader.Key.Source
	lot := must(t)(ledger.NewOpeningLot(lotHeader, in, qty(t, "10"), money(t, cost, "USD"), date(t, "2025-06-02")))
	opening := must(t)(ledger.NewOpeningCash(cashHeader, money(t, cash, "USD")))
	return []ledger.Transaction{lot, opening}
}

// checkSeedRefused fails t unless err refuses to seed account.
func checkSeedRefused(t *testing.T, account ledger.AccountID, err error) {
	t.Helper()
	seededErr, ok := errors.AsType[*ledger.AlreadySeededError](err)
	if !ok || !errors.Is(err, ledger.ErrAlreadySeeded) {
		t.Fatalf("Seed error = %v, want an *AlreadySeededError", err)
	}
	if diff := cmp.Diff(&ledger.AlreadySeededError{Account: account}, seededErr); diff != "" {
		t.Errorf("AlreadySeededError (-want +got):\n%s", diff)
	}
}

func TestSeed(t *testing.T) {
	t.Parallel()
	l, store := seeded(t)
	ctx := t.Context()
	first := seeding(t, acc1, "2026-03-31", "1500", "250")
	assertions := []ledger.BalanceAssertion{ledger.CashAssertion(acc1, date(t, "2026-03-31"), money(t, "250", "USD"))}
	sum, err := l.Seed(ctx, first, assertions)
	if err != nil {
		t.Fatalf("first Seed: %v", err)
	}
	checks := []app.AssertionCheck{{Assertion: assertions[0], Status: app.AssertionMatched, Derived: dec(t, "250")}}
	checkSummary(t, &app.ImportSummary{New: 2, EarliestAffected: affected(t, "acc-1", "2026-03-31"), Checks: checks}, &sum)
	wantLog := []ledger.VersionedTransaction{versioned(&first[1], 1), versioned(&first[0], 1)}
	checkLog(t, store, wantLog)

	again, err := l.Seed(ctx, first, assertions)
	if err != nil {
		t.Fatalf("same Seed again: %v", err)
	}
	checkSummary(t, &app.ImportSummary{Unchanged: 2, AlreadySeeded: []ledger.AccountID{acc1}, Checks: checks}, &again)

	changed := seeding(t, acc1, "2026-03-31", "1510", "250")
	_, err = l.Seed(ctx, changed, nil)
	checkSeedRefused(t, acc1, err)
	checkLog(t, store, wantLog)

	// A seeding of two accounts, one already seeded, is refused whole.
	other := seeding(t, acc2, "2026-03-31", "100", "5")
	_, err = l.Seed(ctx, append(seeding(t, acc1, "2026-04-30", "1500", "250"), other...), nil)
	checkSeedRefused(t, acc1, err)
	checkLog(t, store, wantLog)
	if log, logErr := store.Log(ctx, acc2); logErr != nil || len(log) != 0 {
		t.Errorf("log of acc-2 after a refused seeding = %v, %v; want it empty", log, logErr)
	}

	sum, err = l.Seed(ctx, other, nil)
	if err != nil {
		t.Fatalf("Seed of acc-2: %v", err)
	}
	checkSummary(t, &app.ImportSummary{New: 2, EarliestAffected: affected(t, "acc-2", "2026-03-31")}, &sum)
}

func TestGoLive(t *testing.T) {
	t.Parallel()
	l, _ := seeded(t)
	ctx := t.Context()
	if _, err := l.Seed(ctx, seeding(t, acc1, "2026-03-31", "1500", "250"), nil); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	day, ok, err := l.GoLive(ctx, acc1)
	if err != nil || !ok || !day.Equal(date(t, "2026-03-31")) {
		t.Errorf("GoLive(acc-1) = %s, %v, %v; want 2026-03-31, true", day, ok, err)
	}
	if _, ok, err := l.GoLive(ctx, acc2); err != nil || ok {
		t.Errorf("GoLive(acc-2) = %v, %v; want not seeded", ok, err)
	}
	if _, _, err := l.GoLive(ctx, "acc-9"); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("GoLive(acc-9) error = %v, want ErrUnknownAccount", err)
	}
}

func TestSeedOfAnUnknownAccount(t *testing.T) {
	t.Parallel()
	l, store := seeded(t)
	_, err := l.Seed(t.Context(), seeding(t, "acc-9", "2026-03-31", "1", "1"), nil)
	if diff := cmp.Diff(app.ErrUnknownAccount, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Seed error (-want +got):\n%s", diff)
	}
	if err := store.LockAccount(t.Context(), acc1); err == nil {
		t.Error("LockAccount outside InTx succeeded, want an error")
	}
}

func TestConcurrentSeedsSeedOnce(t *testing.T) {
	t.Parallel()
	const seeds = 8
	l, store := seeded(t)
	days := []string{"2026-03-24", "2026-03-25", "2026-03-26", "2026-03-27", "2026-03-28", "2026-03-29", "2026-03-30", "2026-03-31"}
	statements := make([][]ledger.Transaction, seeds)
	for i := range seeds {
		statements[i] = seeding(t, acc1, days[i], "1500", "250")
	}
	errs := make([]error, seeds)
	var wg sync.WaitGroup
	for i := range seeds {
		wg.Go(func() {
			_, errs[i] = l.Seed(t.Context(), statements[i], nil)
		})
	}
	wg.Wait()
	winner := -1
	for i, err := range errs {
		if err == nil {
			if winner >= 0 {
				t.Fatalf("seedings %d and %d both succeeded", winner, i)
			}
			winner = i
			continue
		}
		checkSeedRefused(t, acc1, err)
	}
	if winner < 0 {
		t.Fatal("no seeding succeeded")
	}
	w := statements[winner]
	checkLog(t, store, []ledger.VersionedTransaction{versioned(&w[1], 1), versioned(&w[0], 1)})
}

// bondLot returns a synthetic opening lot of account at go-live on day,
// opened on opened, keyed per series and purchase date as a manually entered
// account's lots are.
func bondLot(t *testing.T, account ledger.AccountID, series, opened, day, units string) ledger.Transaction {
	t.Helper()
	h := header(t, account, "lot:"+series+":"+opened, day)
	h.Key.Source = "test/" + string(account)
	in := instrument(t, "PLEDO0536000", series)
	return must(t)(ledger.NewOpeningLot(h, in, qty(t, units), money(t, units+"00", "PLN"), date(t, opened)))
}

func TestAddOpeningLots(t *testing.T) {
	t.Parallel()
	l, store := seeded(t)
	ctx := t.Context()
	first := bondLot(t, acc1, "EDO0536", "2026-05-12", "2026-06-01", "10")
	sum, err := l.AddOpeningLots(ctx, []ledger.Transaction{first}, nil)
	if err != nil {
		t.Fatalf("first lot: %v", err)
	}
	checkSummary(t, &app.ImportSummary{New: 1, EarliestAffected: affected(t, "acc-1", "2026-06-01")}, &sum)

	second := bondLot(t, acc1, "EDO0536", "2026-05-20", "2026-06-01", "5")
	if sum, err = l.AddOpeningLots(ctx, []ledger.Transaction{second}, nil); err != nil {
		t.Fatalf("second lot: %v", err)
	}
	checkSummary(t, &app.ImportSummary{New: 1, EarliestAffected: affected(t, "acc-1", "2026-06-01")}, &sum)

	if sum, err = l.AddOpeningLots(ctx, []ledger.Transaction{second}, nil); err != nil {
		t.Fatalf("second lot again: %v", err)
	}
	checkSummary(t, &app.ImportSummary{Unchanged: 1}, &sum)

	// A lot on another date than go-live would move it: refused, nothing imported.
	late := bondLot(t, acc1, "EDO0536", "2026-05-21", "2026-06-15", "1")
	_, err = l.AddOpeningLots(ctx, []ledger.Transaction{late}, nil)
	if diff := cmp.Diff(ledger.ErrGoLiveMismatch, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("lot off go-live error (-want +got):\n%s", diff)
	}

	// Deleting the first lot supersedes it and keeps it in the log.
	gone := must(t)(ledger.NewDeletion(ledger.TransactionHeader{Key: first.Key, Account: acc1, Date: first.Date}))
	if sum, err = l.AddOpeningLots(ctx, []ledger.Transaction{gone}, nil); err != nil {
		t.Fatalf("deleting the first lot: %v", err)
	}
	checkSummary(t, &app.ImportSummary{NewVersions: 1, EarliestAffected: affected(t, "acc-1", "2026-06-01")}, &sum)
	checkLog(t, store, []ledger.VersionedTransaction{versioned(&first, 1), versioned(&gone, 2), versioned(&second, 1)})
	current, err := l.CurrentTransactions(ctx, acc1)
	if err != nil {
		t.Fatalf("CurrentTransactions: %v", err)
	}
	if diff := cmp.Diff([]ledger.Transaction{second}, current); diff != "" {
		t.Errorf("current transactions after the deletion (-want +got):\n%s", diff)
	}

	never := must(t)(ledger.NewDeletion(header(t, acc1, "lot:EDO0536:2026-05-01", "2026-06-01")))
	_, err = l.Import(ctx, []ledger.Transaction{never}, nil)
	if diff := cmp.Diff(ledger.ErrNothingToDelete, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("deleting a lot never entered (-want +got):\n%s", diff)
	}
}
