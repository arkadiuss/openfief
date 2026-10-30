package postgres_test

import (
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestMigrateIsIdempotentAndOwnsTheLedgerSchema(t *testing.T) {
	t.Parallel()
	pool := newPool(t) // migrated once already
	const runs = 4
	errs := make([]error, runs)
	var wg sync.WaitGroup
	for i := range runs {
		wg.Go(func() {
			errs[i] = postgres.Migrate(t.Context(), pool)
		})
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent Migrate %d: %v", i, err)
		}
	}
	rows, err := pool.Query(t.Context(), "SELECT name FROM ledger.schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("reading migrations: %v", err)
	}
	applied, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("reading migrations: %v", err)
	}
	if diff := cmp.Diff([]string{"0001_ledger.sql", "0002_sleeve_kinds.sql", "0003_role_override.sql", "0004_interest_and_deletion.sql"}, applied); diff != "" {
		t.Errorf("applied migrations (-want +got):\n%s", diff)
	}
	rows, err = pool.Query(t.Context(), `
		SELECT table_name::text FROM information_schema.tables
		WHERE table_schema = 'ledger' ORDER BY table_name COLLATE "C"`)
	if err != nil {
		t.Fatalf("listing tables: %v", err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("listing tables: %v", err)
	}
	want := []string{"accounts", "assertion_checks", "postings", "schema_migrations", "sleeve_kinds", "sleeves", "transaction_versions"}
	if diff := cmp.Diff(want, tables); diff != "" {
		t.Errorf("ledger tables (-want +got):\n%s", diff)
	}
}

func TestRegistryRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	ctx := t.Context()
	play := portfolio.Sleeve{ID: "play", Name: "Play", Kind: portfolio.KindRetirement, Role: portfolio.RoleOffensive, RoleOverride: true}
	renamed := portfolio.Sleeve{ID: core, Name: "Core renamed", Kind: portfolio.KindBonds, Role: portfolio.RoleDefensive}
	for _, s := range []portfolio.Sleeve{play, renamed} {
		if err := store.SaveSleeve(ctx, s); err != nil {
			t.Fatalf("SaveSleeve(%s): %v", s.ID, err)
		}
	}
	sleeves, err := store.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve{renamed, play}, sleeves, err)

	moved := ledger.Account{ID: acc2, Institution: "other-broker", Name: "Moved", Sleeve: "play"}
	if saveErr := store.SaveAccount(ctx, moved); saveErr != nil {
		t.Fatalf("SaveAccount: %v", saveErr)
	}
	accounts, err := store.Accounts(ctx)
	checkResult(t, "Accounts", []ledger.Account{{ID: acc1, Institution: "test-broker", Name: "acc-1", Sleeve: core}, moved}, accounts, err)
	one, err := store.Account(ctx, acc2)
	checkResult(t, "Account", moved, one, err)

	_, err = store.Account(ctx, "acc-9")
	if diff := cmp.Diff(app.ErrNotFound, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Account of an unknown ID (-want +got):\n%s", diff)
	}
	if err := store.SaveAccount(ctx, ledger.Account{ID: "acc-9", Institution: "x", Name: "x", Sleeve: "nowhere"}); err == nil {
		t.Errorf("SaveAccount with an unknown sleeve succeeded, want a foreign key error")
	}
}

// logFixture is one synthetic transaction of every kind, a correction, and a
// deposit corrected from acc-2 into acc-1.
type logFixture struct {
	all                          []ledger.VersionedTransaction // in append order
	acc1Log, acc2Log, buyHistory []ledger.VersionedTransaction
}

func newLogFixture(t *testing.T) logFixture {
	t.Helper()
	a, b := instrument(t, isinA, "TSTA"), instrument(t, isinB, "TSTB")
	zero := kernel.Money{}
	buy := must(t)(ledger.NewBuy(header(t, acc1, "02-buy", "2026-03-03"), a, qty(t, "1.5"), money(t, "123.4500", "USD"), money(t, "1", "EUR")))
	buy.Description = "Zakup – synthetic, ąę"
	corrected := must(t)(ledger.NewBuy(header(t, acc1, "02-buy", "2026-03-03"), a, qty(t, "2"), money(t, "164.6", "USD"), money(t, "1", "EUR")))
	misbooked := must(t)(ledger.NewDeposit(header(t, acc2, "09-moved", "2026-03-04"), money(t, "50", "PLN")))
	moved := must(t)(ledger.NewDeposit(header(t, acc1, "09-moved", "2026-03-04"), money(t, "50", "PLN")))
	txs := []ledger.Transaction{
		must(t)(ledger.NewDeposit(header(t, acc1, "01-deposit", "2026-03-02"), money(t, "1000.50", "USD"))),
		buy,
		must(t)(ledger.NewSell(header(t, acc1, "03-sell", "2026-03-04"), a, qty(t, "0.000001"), money(t, "0.01", "USD"), zero)),
		must(t)(ledger.NewDividend(header(t, acc1, "04-dividend", "2026-03-05"), a, money(t, "7", "USD"), money(t, "1.05", "USD"))),
		must(t)(ledger.NewFee(header(t, acc1, "05-fee", "2026-03-05"), money(t, "2.5", "PLN"))),
		must(t)(ledger.NewFXConversion(header(t, acc1, "06-fx", "2026-03-05"), money(t, "100", "USD"), money(t, "90.12", "EUR"), money(t, "0.3", "EUR"))),
		must(t)(ledger.NewWithdrawal(header(t, acc1, "07-withdrawal", "2026-03-06"), money(t, "10", "USD"))),
		must(t)(ledger.NewOpeningLot(header(t, acc1, "08-opening", "2026-03-01"), b, qty(t, "3"), money(t, "300", "EUR"), date(t, "2019-05-02"))),
	}
	var f logFixture
	for i := range txs {
		f.all = append(f.all, versioned(&txs[i], 1))
	}
	f.all = append(f.all, versioned(&corrected, 2), versioned(&misbooked, 1), versioned(&moved, 2))
	f.buyHistory = []ledger.VersionedTransaction{f.all[1], f.all[8]}
	f.acc1Log = []ledger.VersionedTransaction{f.all[0], f.all[1], f.all[8]}
	f.acc1Log = append(f.acc1Log, f.all[2:8]...)
	f.acc1Log = append(f.acc1Log, f.all[9], f.all[10])
	f.acc2Log = []ledger.VersionedTransaction{f.all[9], f.all[10]}
	return f
}

func TestLogRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	f := newLogFixture(t)
	appendAll(t, store, f.all)
	ctx := t.Context()
	versions, err := store.Versions(ctx, f.all[1].Transaction.Key)
	checkResult(t, "Versions", f.buyHistory, versions, err)
	none, err := store.Versions(ctx, ledger.TransactionKey{Source: "test", ExternalID: "never"})
	checkResult(t, "Versions of an unknown key", []ledger.VersionedTransaction(nil), none, err)
	log1, err := store.Log(ctx, acc1)
	checkResult(t, "Log(acc-1)", f.acc1Log, log1, err)
	log2, err := store.Log(ctx, acc2)
	checkResult(t, "Log(acc-2)", f.acc2Log, log2, err)
}

func TestAssertionChecksRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	ctx := t.Context()
	a := instrument(t, isinA, "TSTA")
	cashUSD := ledger.CashAssertion(acc1, date(t, "2026-03-03"), money(t, "499.00", "USD"))
	cashEUR := ledger.CashAssertion(acc1, date(t, "2026-03-03"), money(t, "5", "EUR"))
	units := ledger.HoldingAssertion(acc1, date(t, "2026-03-02"), a, qty(t, "10.5"))
	checks := []app.AssertionCheck{
		{Assertion: cashUSD, Status: app.AssertionFlagged, Derived: dec(t, "498")},
		{Assertion: units, Status: app.AssertionMatched, Derived: dec(t, "10.5")},
		{Assertion: cashEUR, Status: app.AssertionMatched, Derived: dec(t, "5")},
		{Assertion: cashUSD, Status: app.AssertionMatched, Derived: dec(t, "499")}, // replaces the first
		{Assertion: ledger.CashAssertion(acc2, date(t, "2026-03-03"), money(t, "1", "USD")), Status: app.AssertionMatched, Derived: dec(t, "1")},
	}
	for i := range checks {
		if err := store.SaveAssertionCheck(ctx, &checks[i]); err != nil {
			t.Fatalf("SaveAssertionCheck %d: %v", i, err)
		}
	}
	got, err := store.AssertionChecks(ctx, acc1)
	checkResult(t, "AssertionChecks", []app.AssertionCheck{checks[1], checks[2], checks[3]}, got, err)
}

// checkResult fails t on err, or unless got is want; empty and nil slices
// are equal.
func checkResult[T any](t *testing.T, name string, want, got T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("%s (-want +got):\n%s", name, diff)
	}
}
