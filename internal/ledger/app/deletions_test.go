package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// DeleteAccount refuses an account with a transaction version or a check,
// like the Postgres store.
func (s *memStore) DeleteAccount(_ context.Context, id ledger.AccountID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.holdsLocked(id, "") {
		return fmt.Errorf("account %s: %w", id, ledger.ErrAccountHasTransactions)
	}
	for k := range s.checks {
		if s.checks[k].Assertion.Account == id {
			return fmt.Errorf("account %s: %w", id, ledger.ErrAccountHasTransactions)
		}
	}
	delete(s.accounts, id)
	return nil
}

func (s *memStore) DeleteSleeve(_ context.Context, id portfolio.SleeveID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.accounts {
		if a.Sleeve == id {
			return fmt.Errorf("sleeve %s: %w", id, portfolio.ErrSleeveHasAccounts)
		}
	}
	delete(s.sleeves, id)
	return nil
}

func (s *memStore) DeleteKind(_ context.Context, key portfolio.KindKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sl := range s.sleeves {
		if sl.Kind == key {
			return fmt.Errorf("kind %s: %w", key, portfolio.ErrKindInUse)
		}
	}
	delete(s.kinds, key)
	return nil
}

func (s *memStore) HasTransactionsFrom(_ context.Context, account ledger.AccountID, prefix string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.holdsLocked(account, prefix), nil
}

// holdsLocked reports whether account holds a version keyed from a source
// starting with prefix; s.mu is held.
func (s *memStore) holdsLocked(account ledger.AccountID, prefix string) bool {
	for key, vs := range s.versions {
		for i := range vs {
			if vs[i].Transaction.Account == account && strings.HasPrefix(key.Source, prefix) {
				return true
			}
		}
	}
	return false
}

func TestDeleteRegistered(t *testing.T) {
	t.Parallel()
	l, store := seeded(t)
	ctx := t.Context()
	deposit := must(t)(ledger.NewDeposit(header(t, acc1, "d-1", "2026-03-02"), money(t, "100", "PLN")))
	if _, err := l.Import(ctx, []ledger.Transaction{deposit}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	// D-120: an account with transactions, a sleeve with accounts and a kind in use stay.
	steps := []struct {
		name    string
		del     func() error
		wantErr error
	}{
		{"an account with a transaction", func() error { return l.DeleteAccount(ctx, acc1) }, ledger.ErrAccountHasTransactions},
		{"a sleeve with accounts", func() error { return l.DeleteSleeve(ctx, core) }, portfolio.ErrSleeveHasAccounts},
		{"a kind a sleeve has", func() error { return l.DeleteKind(ctx, portfolio.KindCore) }, portfolio.ErrKindInUse},
		{"an account with none", func() error { return l.DeleteAccount(ctx, acc2) }, nil},
		{"an account deleted already", func() error { return l.DeleteAccount(ctx, acc2) }, nil},
		{"an unused built-in kind", func() error { return l.DeleteKind(ctx, portfolio.KindSatellite) }, nil},
	}
	for _, s := range steps {
		if err := s.del(); !errors.Is(err, s.wantErr) {
			t.Errorf("deleting %s = %v, want %v", s.name, err, s.wantErr)
		}
	}
	if _, ok := store.accounts[acc1]; !ok {
		t.Error("the account with a transaction is gone")
	}
	if _, ok := store.accounts[acc2]; ok {
		t.Error("the account with no transactions is still stored")
	}
	if _, ok := store.kinds[portfolio.KindSatellite]; ok {
		t.Error("the unused kind is still stored")
	}
	for _, tt := range []struct {
		prefix string
		want   bool
	}{{"", true}, {"te", true}, {"plugin:ibkr-main:", false}} {
		got, err := l.HasTransactionsFrom(ctx, acc1, tt.prefix)
		if err != nil || got != tt.want {
			t.Errorf("HasTransactionsFrom(%q) = %v, %v; want %v", tt.prefix, got, err, tt.want)
		}
	}
}
