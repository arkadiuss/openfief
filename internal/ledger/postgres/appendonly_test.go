package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// SQLSTATEs the ledger schema's triggers raise.
const (
	appendOnlyViolation = "OFA01"
	versionGap          = "OFA02"
)

func TestAppendRefusesAStoredVersion(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	f := newLogFixture(t)
	appendAll(t, store, f.all[:2])
	again := f.all[1]
	err := store.Append(t.Context(), &again)
	if diff := cmp.Diff(app.ErrVersionConflict, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("Append of a stored version (-want +got):\n%s", diff)
	}
	gap := versioned(&f.all[8].Transaction, 3)
	err = store.Append(t.Context(), &gap)
	if errors.Is(err, app.ErrVersionConflict) || sqlState(err) != versionGap {
		t.Errorf("Append of version 3 without version 2 = %v, want SQLSTATE %s", err, versionGap)
	}
	versions, err := store.Versions(t.Context(), again.Transaction.Key)
	checkResult(t, "Versions after refused appends", f.all[1:2], versions, err)
}

func TestLogIsAppendOnlyInTheDatabase(t *testing.T) {
	t.Parallel()
	store, pool := seededStore(t)
	f := newLogFixture(t)
	appendAll(t, store, f.all)
	statements := []string{
		"UPDATE ledger.transaction_versions SET description = 'changed'",
		"UPDATE ledger.transaction_versions SET version = version + 10",
		"DELETE FROM ledger.transaction_versions WHERE version = 1",
		"TRUNCATE ledger.transaction_versions CASCADE",
		"UPDATE ledger.postings SET amount = 0",
		"DELETE FROM ledger.postings",
		"TRUNCATE ledger.postings",
		`INSERT INTO ledger.postings (version_id, position, book_kind, amount, currency)
		 SELECT min(id), 99, 'cash', 1, 'USD' FROM ledger.transaction_versions`,
	}
	for _, sql := range statements {
		if _, err := pool.Exec(t.Context(), sql); sqlState(err) != appendOnlyViolation {
			t.Errorf("%s: error %v, want SQLSTATE %s", sql, err, appendOnlyViolation)
		}
	}
	log, err := store.Log(t.Context(), acc1)
	checkResult(t, "Log after refused changes", f.acc1Log, log, err)
}

func TestLoadDetectsAChangedVersion(t *testing.T) {
	t.Parallel()
	store, pool := seededStore(t)
	f := newLogFixture(t)
	appendAll(t, store, f.all[:1])
	// The schema's owner can switch the trigger off; the stored fingerprint
	// still catches the change.
	if _, err := pool.Exec(t.Context(), `
		ALTER TABLE ledger.postings DISABLE TRIGGER postings_append_only;
		UPDATE ledger.postings SET amount = amount * 2;
		ALTER TABLE ledger.postings ENABLE TRIGGER postings_append_only;`); err != nil {
		t.Fatalf("tampering: %v", err)
	}
	_, logErr := store.Log(t.Context(), acc1)
	_, versionsErr := store.Versions(t.Context(), f.all[0].Transaction.Key)
	for name, err := range map[string]error{"Log": logErr, "Versions": versionsErr} {
		if diff := cmp.Diff(postgres.ErrCorruptLog, err, cmpopts.EquateErrors()); diff != "" {
			t.Errorf("%s of a changed version (-want +got):\n%s", name, diff)
		}
	}
}

func TestImportOverPostgres(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	l := app.New(store)
	ctx := t.Context()
	a := instrument(t, isinA, "TSTA")
	deposit := must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1000", "USD")))
	buy := must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), a, qty(t, "10"), money(t, "500", "USD"), money(t, "1", "USD")))
	corrected := must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), a, qty(t, "12"), money(t, "600", "USD"), money(t, "1", "USD")))
	assertions := []ledger.BalanceAssertion{
		ledger.CashAssertion(acc1, date(t, "2026-03-03"), money(t, "499", "USD")),
		ledger.HoldingAssertion(acc1, date(t, "2026-03-03"), a, qty(t, "12")),
	}
	first, err := l.Import(ctx, []ledger.Transaction{deposit, buy}, assertions)
	checkResult(t, "first Import", app.ImportSummary{New: 2, EarliestAffected: map[ledger.AccountID]kernel.Date{acc1: date(t, "2026-03-02")}, Checks: []app.AssertionCheck{
		{Assertion: assertions[0], Status: app.AssertionMatched, Derived: dec(t, "499")},
		{Assertion: assertions[1], Status: app.AssertionFlagged, Derived: dec(t, "10")},
	}}, first, err)
	again, err := l.Import(ctx, []ledger.Transaction{deposit, buy}, nil)
	checkResult(t, "re-import", app.ImportSummary{Unchanged: 2, Checks: first.Checks}, again, err)

	const imports = 8
	sums := make([]app.ImportSummary, imports)
	errs := make([]error, imports)
	done := make(chan int)
	for i := range imports {
		go func() {
			sums[i], errs[i] = l.Import(ctx, []ledger.Transaction{corrected}, nil)
			done <- i
		}()
	}
	var total app.ImportSummary
	for range imports {
		i := <-done
		if errs[i] != nil {
			t.Fatalf("concurrent Import %d: %v", i, errs[i])
		}
		total.Unchanged += sums[i].Unchanged
		total.NewVersions += sums[i].NewVersions
	}
	checkResult(t, "concurrent imports", app.ImportSummary{Unchanged: imports - 1, NewVersions: 1}, total, nil)
	if got := sums[winner(sums)].EarliestAffected; !got[acc1].Equal(date(t, "2026-03-03")) || len(got) != 1 {
		t.Errorf("EarliestAffected of the import that appended = %v, want acc-1 from 2026-03-03", got)
	}
	versions, err := store.Versions(ctx, buy.Key)
	checkResult(t, "Versions", []ledger.VersionedTransaction{versioned(&buy, 1), versioned(&corrected, 2)}, versions, err)
	checks, err := l.AssertionChecks(ctx, acc1)
	checkResult(t, "AssertionChecks", []app.AssertionCheck{
		{Assertion: assertions[0], Status: app.AssertionFlagged, Derived: dec(t, "399")},
		{Assertion: assertions[1], Status: app.AssertionMatched, Derived: dec(t, "12")},
	}, checks, err)
}

// winner returns the index of the summary that appended a new version.
func winner(sums []app.ImportSummary) int {
	for i := range sums {
		if sums[i].NewVersions > 0 {
			return i
		}
	}
	return 0
}
