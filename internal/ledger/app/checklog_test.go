package app_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestCheckLog checks the whole-log check over a stored ledger: every
// account's versions, the deposit corrected from acc-2 to acc-1 counted
// once as a superseded version in acc-2 and a current one in acc-1, and a
// balanced log whose postings sum to zero per currency.
func TestCheckLog(t *testing.T) {
	t.Parallel()
	f := newQueryFixture(t)
	got, err := f.ledger.CheckLog(t.Context())
	if err != nil {
		t.Fatalf("CheckLog: %v", err)
	}
	counts := make([]string, 0, len(got.Accounts)+1+len(got.Sums))
	for _, a := range got.Accounts {
		counts = append(counts, fmt.Sprintf("%s %d/%d/%d", a.Account, a.Transactions, a.Versions, a.Superseded))
	}
	counts = append(counts, fmt.Sprintf("all %d/%d/%d", got.Transactions, got.Versions, got.Superseded))
	for _, s := range got.Sums {
		counts = append(counts, "sum "+s.String())
	}
	want := []string{"acc-1 6/6/0", "acc-2 1/1/1", "all 6/7/1", "sum 0 EUR", "sum 0 USD"}
	if diff := cmp.Diff(want, counts); diff != "" {
		t.Errorf("CheckLog (-want +got):\n%s", diff)
	}
	if !got.Balanced() || len(got.Unbalanced) > 0 {
		t.Errorf("CheckLog = unbalanced %+v, want a balanced log", got.Unbalanced)
	}
}
