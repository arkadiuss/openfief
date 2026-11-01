package postgres_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// TestDeleteRegistered checks D-120 in PostgreSQL: an account, sleeve or
// kind goes only while nothing refers to it, and nothing of the transaction
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
		{"a sleeve with accounts", func() error { return store.DeleteSleeve(ctx, core) }, portfolio.ErrSleeveHasAccounts},
		{"a kind a sleeve has", func() error { return store.DeleteKind(ctx, portfolio.KindCore) }, portfolio.ErrKindInUse},
		{"an account with none", func() error { return store.DeleteAccount(ctx, acc2) }, nil},
		{"an account deleted already", func() error { return store.DeleteAccount(ctx, acc2) }, nil},
		{"an unused built-in kind", func() error { return store.DeleteKind(ctx, portfolio.KindSatellite) }, nil},
	}
	for _, s := range steps {
		if err := s.del(); !errors.Is(err, s.wantErr) {
			t.Errorf("deleting %s = %v, want %v", s.name, err, s.wantErr)
		}
	}
	accounts, err := store.Accounts(ctx)
	checkResult(t, "Accounts", []ledger.Account{{ID: acc1, Institution: "test-broker", Name: string(acc1), Sleeve: core}}, accounts, err)
	var versions int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM ledger.transaction_versions").Scan(&versions); err != nil || versions != 1 {
		t.Errorf("transaction versions = %d, %v; want the deposit kept", versions, err)
	}
	// With its only account gone, the sleeve goes, and then its kind.
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

// TestDeleteSleeveThenKind deletes a sleeve once its account is gone, then
// its kind.
func TestDeleteSleeveThenKind(t *testing.T) {
	t.Parallel()
	store, _ := seededStore(t)
	ctx := t.Context()
	for _, id := range []ledger.AccountID{acc1, acc2} {
		if err := store.DeleteAccount(ctx, id); err != nil {
			t.Fatalf("DeleteAccount(%s): %v", id, err)
		}
	}
	if err := store.DeleteSleeve(ctx, core); err != nil {
		t.Fatalf("DeleteSleeve: %v", err)
	}
	if err := store.DeleteKind(ctx, portfolio.KindCore); err != nil {
		t.Fatalf("DeleteKind: %v", err)
	}
	sleeves, err := store.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve(nil), sleeves, err)
}
