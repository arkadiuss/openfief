package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// TestScopeBatch checks D-199 as a caller sees a batch: every account it
// may view keeps the batch's totals; only some of them, the sums of their
// counts and none of the rows tied to no account; a batch that booked into
// no account, or one recorded before its counts per account were, as it is.
func TestScopeBatch(t *testing.T) {
	t.Parallel()
	a1 := ingest.AccountCounts{Account: "acc-1", Transactions: 3, New: 2, NewVersions: 1, FailedAssertions: 1}
	a2 := ingest.AccountCounts{Account: "acc-2", Transactions: 5, New: 4, Unchanged: 1}
	both := ingest.BatchRecord{
		ID: "u-1", Transactions: 8, New: 6, NewVersions: 1, Unchanged: 1, FailedAssertions: 1, Problems: 2, Unsupported: 1,
		Accounts: []ledger.AccountID{"acc-1", "acc-2"}, PerAccount: []ingest.AccountCounts{a1, a2},
	}
	legacy := ingest.BatchRecord{ID: "u-0", Transactions: 8, New: 6, Problems: 2, Accounts: []ledger.AccountID{"acc-1", "acc-2"}}
	none := ingest.BatchRecord{ID: "n-0", Transactions: 4, New: 4, Problems: 1}
	tests := []struct {
		name    string
		batch   ingest.BatchRecord
		visible []ledger.AccountID
		want    ingest.BatchRecord
	}{
		{"every account viewable keeps the totals", both, []ledger.AccountID{"acc-1", "acc-2"}, both},
		{
			"one account viewable sums its counts alone", both,
			[]ledger.AccountID{"acc-2"},
			ingest.BatchRecord{
				ID: "u-1", Transactions: 5, New: 4, Unchanged: 1, Accounts: []ledger.AccountID{"acc-2"}, PerAccount: []ingest.AccountCounts{a2},
			},
		},
		{"none viewable names none", both, nil, ingest.BatchRecord{ID: "u-1"}},
		{"counts not recorded per account show none", legacy, []ledger.AccountID{"acc-1"}, ingest.BatchRecord{ID: "u-0", Accounts: []ledger.AccountID{"acc-1"}}},
		{"no account booked stays as it is", none, nil, none},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, views.ScopeBatch(&tt.batch, tt.visible)); diff != "" {
				t.Errorf("ScopeBatch (-want +got):\n%s", diff)
			}
		})
	}
}
