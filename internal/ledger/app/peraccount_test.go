package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// TestImportSummaryPerAccount checks what an import did in each account
// (D-199): its transactions by decision, and its flagged checks, an
// account with only a flagged check included, in ID order.
func TestImportSummaryPerAccount(t *testing.T) {
	t.Parallel()
	flagged := func(account ledger.AccountID) app.AssertionCheck {
		return app.AssertionCheck{Assertion: ledger.BalanceAssertion{Account: account}, Status: app.AssertionFlagged}
	}
	matched := app.AssertionCheck{Assertion: ledger.BalanceAssertion{Account: "acc-1"}, Status: app.AssertionMatched}
	tests := []struct {
		name string
		sum  app.ImportSummary
		want []app.AccountResult
	}{
		{"nothing imported", app.ImportSummary{}, nil},
		{
			"two accounts and a flagged check elsewhere",
			app.ImportSummary{
				Accounts: map[ledger.AccountID]app.AccountImport{"acc-2": {New: 1}, "acc-1": {Unchanged: 2, NewVersions: 1}},
				Checks:   []app.AssertionCheck{flagged("acc-3"), matched, flagged("acc-1")},
			},
			[]app.AccountResult{
				{Account: "acc-1", Unchanged: 2, NewVersions: 1, Flagged: 1},
				{Account: "acc-2", New: 1},
				{Account: "acc-3", Flagged: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, tt.sum.PerAccount(), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("PerAccount (-want +got):\n%s", diff)
			}
		})
	}
}
