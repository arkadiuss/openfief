package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// errEnqueue stands in for a failing job-queue insert.
var errEnqueue = errors.New("enqueue failed")

// errHookBlind reports a hook that does not see the import's appends.
var errHookBlind = errors.New("hook does not see the appended version")

// recordingHook returns a hook that checks it sees key's one appended
// version, then fails with fail or records the summary in seen.
func recordingHook(key ledger.TransactionKey, fail error, seen *[]app.ImportSummary) app.ImportHook {
	return func(ctx context.Context, tx app.Store, sum *app.ImportSummary) error {
		versions, err := tx.Versions(ctx, key)
		if err != nil {
			return fmt.Errorf("hook: %w", err)
		}
		if len(versions) != 1 {
			return errHookBlind
		}
		if fail != nil {
			return fail
		}
		*seen = append(*seen, *sum)
		return nil
	}
}

func TestImportHooksRunInTheImportTransaction(t *testing.T) {
	t.Parallel()
	f := newImportFixture(t)
	assertion := ledger.CashAssertion(acc1, date(t, "2026-03-02"), money(t, "1000", "USD"))
	matched := app.AssertionCheck{Assertion: assertion, Status: app.AssertionMatched, Derived: dec(t, "1000")}
	tests := []struct {
		name     string
		hookErr  error
		wantErr  error
		wantLog  []ledger.VersionedTransaction
		wantSeen []app.ImportSummary
		// wantChecks are the stored checks after the import.
		wantChecks []app.AssertionCheck
	}{
		{
			name:    "a hook sees the summary and the appended versions",
			wantLog: []ledger.VersionedTransaction{versioned(&f.deposit, 1)},
			wantSeen: []app.ImportSummary{{
				New:              1,
				EarliestAffected: affected(t, "acc-1", "2026-03-02"),
				Checks:           []app.AssertionCheck{matched},
			}},
			wantChecks: []app.AssertionCheck{matched},
		},
		{
			name:     "a failing hook rolls the import back",
			hookErr:  errEnqueue,
			wantErr:  errEnqueue,
			wantSeen: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, store := seeded(t)
			var seen []app.ImportSummary
			l := app.New(store, recordingHook(f.deposit.Key, tt.hookErr, &seen))
			_, err := l.Import(t.Context(), []ledger.Transaction{f.deposit}, []ledger.BalanceAssertion{assertion})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Import error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSeen, seen, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("summaries the hook saw (-want +got):\n%s", diff)
			}
			checkLog(t, store, tt.wantLog)
			checks, err := store.AssertionChecks(t.Context(), acc1)
			checkResult(t, "stored checks", tt.wantChecks, checks, err)
		})
	}
}
