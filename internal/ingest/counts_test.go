package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestCountPerAccount checks a batch's counts per account (D-199): each
// account its mapping booked into, in ID order, with its transactions and
// the ledger's counts of it; an account with only a balance assertion has
// none of the first; one the ledger names but the mapping doesn't is left
// out.
func TestCountPerAccount(t *testing.T) {
	t.Parallel()
	tx := func(account ledger.AccountID) ledger.Transaction { return ledger.Transaction{Account: account} }
	m := ingest.Mapping{
		Transactions: []ledger.Transaction{tx("acc-2"), tx("acc-1"), tx("acc-2")},
		Assertions:   []ledger.BalanceAssertion{{Account: "acc-3"}},
	}
	imported := ingest.ImportSummary{Accounts: []ingest.AccountCounts{
		{Account: "acc-1", New: 1},
		{Account: "acc-2", New: 1, NewVersions: 1},
		{Account: "acc-3", FailedAssertions: 1},
		{Account: "acc-9", New: 5},
	}}
	want := []ingest.AccountCounts{
		{Account: "acc-1", Transactions: 1, New: 1},
		{Account: "acc-2", Transactions: 2, New: 1, NewVersions: 1},
		{Account: "acc-3", FailedAssertions: 1},
	}
	if diff := cmp.Diff(want, ingest.CountPerAccount(&m, &imported)); diff != "" {
		t.Errorf("CountPerAccount (-want +got):\n%s", diff)
	}
}
