package views_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// pairLink is a confirmed link of the withdrawal w and the deposit d, both
// keyed in source "test".
func pairLink(w, d string) ledger.TransferLink {
	return ledger.TransferLink{
		Pair: ledger.TransferPair{
			Withdrawal: ledger.TransactionKey{Source: "test", ExternalID: w},
			Deposit:    ledger.TransactionKey{Source: "test", ExternalID: d},
		},
		Decision: ledger.TransferConfirmed, Version: 1,
	}
}

// legIn is a current transaction with key id in account.
func legIn(account ledger.AccountID, id string) ledger.Transaction {
	return ledger.Transaction{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: account}
}

func TestPairAccounts(t *testing.T) {
	t.Parallel()
	txs := []ledger.Transaction{legIn("acc-1", "w-1"), legIn("acc-2", "d-1"), legIn("acc-1", "w-2"), legIn("acc-1", "d-3"), legIn("acc-1", "w-3")}
	tests := []struct {
		name string
		link ledger.TransferLink
		want []ledger.AccountID
	}{
		{name: "legs in two accounts, withdrawal first", link: pairLink("w-1", "d-1"), want: []ledger.AccountID{"acc-1", "acc-2"}},
		{name: "a deleted deposit", link: pairLink("w-2", "d-gone"), want: []ledger.AccountID{"acc-1"}},
		{name: "both legs in one account after a correction", link: pairLink("w-3", "d-3"), want: []ledger.AccountID{"acc-1"}},
		{name: "both legs deleted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			link := tt.link
			if link.Version == 0 {
				link = pairLink("w-gone", "d-gone")
			}
			got, ok := views.PairAccounts([]ledger.TransferLink{link}, txs)[link.Pair.String()]
			if ok != (tt.want != nil) {
				t.Fatalf("pair listed = %v, want %v", ok, tt.want != nil)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("accounts (-want +got):\n%s", diff)
			}
		})
	}
}
