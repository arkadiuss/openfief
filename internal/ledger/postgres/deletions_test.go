package postgres_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestDeleteRegistered checks D-120 in PostgreSQL: an account goes only
// while nothing refers to it, and nothing of the transaction
// log goes with it (D-18).
func TestDeleteRegistered(t *testing.T) {
	t.Parallel()
	store, pool := seededStore(t)
	ctx := t.Context()
	deposit := must(t)(ledger.NewDeposit(header(t, acc1, "d-1", "2026-03-02"), money(t, "100", "PLN")))
	appendAll(t, store, []ledger.VersionedTransaction{versioned(&deposit, 1)})
	steps := []struct {
		name    string
		del     func() error
		wantErr error
	}{
		{"an account with a transaction", func() error { return store.DeleteAccount(ctx, acc1) }, ledger.ErrAccountHasTransactions},
		{"an account with none", func() error { return store.DeleteAccount(ctx, acc2) }, nil},
		{"an account deleted already", func() error { return store.DeleteAccount(ctx, acc2) }, nil},
	}
	for _, s := range steps {
		if err := s.del(); !errors.Is(err, s.wantErr) {
			t.Errorf("deleting %s = %v, want %v", s.name, err, s.wantErr)
		}
	}
	accounts, err := store.Accounts(ctx)
	checkResult(t, "Accounts", []ledger.Account{{ID: acc1, Institution: "test-broker", Name: string(acc1)}}, accounts, err)
	var versions int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM ledger.transaction_versions").Scan(&versions); err != nil || versions != 1 {
		t.Errorf("transaction versions = %d, %v; want the deposit kept", versions, err)
	}
	// The account with a transaction stays on a second try too.
	if err := store.DeleteAccount(ctx, acc1); !errors.Is(err, ledger.ErrAccountHasTransactions) {
		t.Fatalf("DeleteAccount = %v", err)
	}
	for _, tt := range []struct {
		prefix string
		want   bool
	}{{"", true}, {"te", true}, {"plugin:ibkr-main:", false}} {
		got, err := store.HasTransactionsFrom(ctx, acc1, tt.prefix)
		if err != nil || got != tt.want {
			t.Errorf("HasTransactionsFrom(%q) = %v, %v; want %v", tt.prefix, got, err, tt.want)
		}
	}
}
