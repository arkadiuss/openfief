package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

func TestImportFlagsAssertionMismatches(t *testing.T) {
	t.Parallel()
	f := newImportFixture(t)
	a := instrument(t, isinA, "TSTA")
	cash := func(on, amount string) ledger.BalanceAssertion {
		return ledger.CashAssertion(acc1, date(t, on), money(t, amount, "USD"))
	}
	check := func(assertion ledger.BalanceAssertion, status app.AssertionStatus, derived string) app.AssertionCheck {
		return app.AssertionCheck{Assertion: assertion, Status: status, Derived: dec(t, derived)}
	}
	units10 := ledger.HoldingAssertion(acc1, date(t, "2026-03-03"), a, qty(t, "10"))
	depositCorrected := must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "999", "USD")))

	l, _ := seeded(t)
	steps := []struct {
		name       string
		txs        []ledger.Transaction
		assertions []ledger.BalanceAssertion
		want       app.ImportSummary
	}{
		{
			name:       "a mismatch is flagged and the import is kept",
			txs:        []ledger.Transaction{f.deposit, f.buy},
			assertions: []ledger.BalanceAssertion{cash("2026-03-02", "999"), cash("2026-03-03", "499"), units10},
			want: app.ImportSummary{New: 2, EarliestAffected: affected(t, "acc-1", "2026-03-02"), Checks: []app.AssertionCheck{
				check(cash("2026-03-02", "999"), app.AssertionFlagged, "1000"),
				check(cash("2026-03-03", "499"), app.AssertionMatched, "499"),
				check(units10, app.AssertionMatched, "10"),
			}},
		},
		{
			name: "a correction re-checks the stored assertions",
			txs:  []ledger.Transaction{depositCorrected},
			want: app.ImportSummary{NewVersions: 1, EarliestAffected: affected(t, "acc-1", "2026-03-02"), Checks: []app.AssertionCheck{
				check(cash("2026-03-02", "999"), app.AssertionMatched, "999"),
				check(cash("2026-03-03", "499"), app.AssertionFlagged, "498"),
				check(units10, app.AssertionMatched, "10"),
			}},
		},
		{
			name:       "a new report replaces the stored assertion",
			assertions: []ledger.BalanceAssertion{cash("2026-03-03", "498")},
			want: app.ImportSummary{Checks: []app.AssertionCheck{
				check(cash("2026-03-02", "999"), app.AssertionMatched, "999"),
				check(cash("2026-03-03", "498"), app.AssertionMatched, "498"),
				check(units10, app.AssertionMatched, "10"),
			}},
		},
	}
	// The steps build on each other, so they run in order.
	for _, step := range steps {
		got, err := l.Import(t.Context(), step.txs, step.assertions)
		if err != nil {
			t.Fatalf("%s: Import: %v", step.name, err)
		}
		if diff := cmp.Diff(step.want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("%s: ImportSummary (-want +got):\n%s", step.name, diff)
		}
		stored, err := l.AssertionChecks(t.Context(), acc1)
		if err != nil {
			t.Fatalf("%s: AssertionChecks: %v", step.name, err)
		}
		if diff := cmp.Diff(step.want.Checks, stored); diff != "" {
			t.Errorf("%s: stored checks (-want +got):\n%s", step.name, diff)
		}
	}
}

func TestImportSummaryFlagged(t *testing.T) {
	t.Parallel()
	matched := app.AssertionCheck{Status: app.AssertionMatched}
	flagged := app.AssertionCheck{Status: app.AssertionFlagged, Derived: dec(t, "1")}
	sum := app.ImportSummary{Checks: []app.AssertionCheck{matched, flagged, matched}}
	if diff := cmp.Diff([]app.AssertionCheck{flagged}, sum.Flagged()); diff != "" {
		t.Errorf("Flagged (-want +got):\n%s", diff)
	}
	if got := (&app.ImportSummary{Checks: []app.AssertionCheck{matched}}).Flagged(); len(got) != 0 {
		t.Errorf("Flagged = %v, want none", got)
	}
}

func TestImportRejectsBadAssertions(t *testing.T) {
	t.Parallel()
	feeAssertion := ledger.CashAssertion(acc1, date(t, "2026-03-02"), money(t, "0", "USD"))
	feeAssertion.Book = ledger.FeeBook(feeAssertion.Book.Currency)
	tests := []struct {
		name      string
		assertion ledger.BalanceAssertion
		wantErr   error
	}{
		{
			name:      "unknown account",
			assertion: ledger.CashAssertion("acc-9", date(t, "2026-03-02"), money(t, "0", "USD")),
			wantErr:   app.ErrUnknownAccount,
		},
		{name: "fee ledger account", assertion: feeAssertion, wantErr: ledger.ErrInvalidAssertion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := seeded(t)
			deposit := must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1", "USD")))
			_, err := l.Import(t.Context(), []ledger.Transaction{deposit}, []ledger.BalanceAssertion{tt.assertion})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Import error (-want +got):\n%s", diff)
			}
			txs, err := l.CurrentTransactions(t.Context(), acc1)
			if err != nil {
				t.Fatalf("CurrentTransactions: %v", err)
			}
			if len(txs) != 0 {
				t.Errorf("a rejected import wrote %d transactions", len(txs))
			}
		})
	}
}

func TestAssertionStatusNames(t *testing.T) {
	t.Parallel()
	for s, want := range map[app.AssertionStatus]string{
		app.AssertionMatched: "matched",
		app.AssertionFlagged: "flagged",
		0:                    "AssertionStatus(0)",
	} {
		if got := s.String(); got != want {
			t.Errorf("AssertionStatus(%d).String() = %q, want %q", int(s), got, want)
		}
	}
}
