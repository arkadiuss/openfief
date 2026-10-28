package app_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"sync"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// errVersionGap reports an append that skips a version number.
var errVersionGap = errors.New("previous version missing")

// errNoTx reports a LockAccount outside InTx.
var errNoTx = errors.New("not in a transaction")

// memStore is an in-memory app.Store with the same rules as the Postgres
// one: versions are only appended, numbered 1..n per key with no repeat.
type memStore struct {
	mu       sync.Mutex
	versions map[ledger.TransactionKey][]ledger.VersionedTransaction
	kinds    map[portfolio.KindKey]portfolio.Kind
	sleeves  map[portfolio.SleeveID]portfolio.Sleeve
	accounts map[ledger.AccountID]ledger.Account
	checks   map[checkKey]*app.AssertionCheck
	// beforeAppend, when set, runs at the start of every Append, standing in
	// for another import that appends between this one's read and write.
	beforeAppend func(s *memStore, v *ledger.VersionedTransaction) error
	// outside are versions that another import committed, which a rollback
	// of this one keeps.
	outside []ledger.VersionedTransaction
	// locks holds each account's LockAccount lock.
	locks map[ledger.AccountID]*sync.Mutex
}

// memTx is the app.Store that memStore.InTx passes in: the store itself,
// the state a rollback restores, and the account locks it holds until InTx
// returns.
type memTx struct {
	*memStore

	versions map[ledger.TransactionKey][]ledger.VersionedTransaction
	checks   map[checkKey]*app.AssertionCheck
	held     []*sync.Mutex
}

// checkKey identifies a stored balance assertion check.
type checkKey struct {
	account ledger.AccountID
	date    kernel.Date
	book    ledger.Book
}

func newMemStore() *memStore {
	return &memStore{
		versions: map[ledger.TransactionKey][]ledger.VersionedTransaction{},
		kinds:    map[portfolio.KindKey]portfolio.Kind{},
		sleeves:  map[portfolio.SleeveID]portfolio.Sleeve{},
		accounts: map[ledger.AccountID]ledger.Account{},
		checks:   map[checkKey]*app.AssertionCheck{},
		locks:    map[ledger.AccountID]*sync.Mutex{},
	}
}

func (s *memStore) Append(_ context.Context, v *ledger.VersionedTransaction) error {
	if s.beforeAppend != nil {
		if err := s.beforeAppend(s, v); err != nil {
			return err
		}
	}
	return s.appendVersion(v)
}

// InTx runs fn on s itself and, if fn fails, restores what s held before,
// keeping versions appended outside. Unlike a database it does not isolate
// fn from concurrent calls, so tests that roll back run alone, except for
// transactions serialized by LockAccount.
func (s *memStore) InTx(_ context.Context, fn func(tx app.Store) error) error {
	tx := &memTx{memStore: s}
	tx.snapshot()
	defer tx.unlock()
	if err := fn(tx); err != nil {
		s.mu.Lock()
		s.versions, s.checks = tx.versions, tx.checks
		outside := s.outside
		s.mu.Unlock()
		for i := range outside {
			if v := &outside[i]; v.Version > len(s.versions[v.Transaction.Key]) {
				if appendErr := s.appendVersion(v); appendErr != nil {
					return appendErr
				}
			}
		}
		return err
	}
	return nil
}

// LockAccount fails outside InTx, like the Postgres store.
func (s *memStore) LockAccount(_ context.Context, id ledger.AccountID) error {
	return fmt.Errorf("locking account %s: %w", id, errNoTx)
}

// InTx runs fn in tx's own transaction.
func (tx *memTx) InTx(_ context.Context, fn func(tx app.Store) error) error {
	return fn(tx)
}

// LockAccount waits for and holds id's lock until InTx returns. Then it
// retakes the rollback snapshot, so that a rollback keeps what the
// transaction it waited for committed; that is right only for a lock taken
// before the transaction's first write, as Seed takes it.
func (tx *memTx) LockAccount(ctx context.Context, id ledger.AccountID) error {
	if _, err := tx.Account(ctx, id); err != nil {
		return err
	}
	tx.mu.Lock()
	m, ok := tx.locks[id]
	if !ok {
		m = &sync.Mutex{}
		tx.locks[id] = m
	}
	tx.mu.Unlock()
	m.Lock()
	tx.held = append(tx.held, m)
	tx.snapshot()
	return nil
}

// snapshot records the state a rollback of tx restores.
func (tx *memTx) snapshot() {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.versions, tx.checks = maps.Clone(tx.memStore.versions), maps.Clone(tx.memStore.checks)
}

// unlock releases the account locks tx holds.
func (tx *memTx) unlock() {
	for _, m := range tx.held {
		m.Unlock()
	}
}

func (s *memStore) Versions(_ context.Context, key ledger.TransactionKey) ([]ledger.VersionedTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.versions[key]), nil
}

func (s *memStore) Log(_ context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var keys []ledger.TransactionKey
	for key, vs := range s.versions {
		if slices.ContainsFunc(vs, func(v ledger.VersionedTransaction) bool { return v.Transaction.Account == account }) {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	var out []ledger.VersionedTransaction
	for _, key := range keys {
		out = append(out, s.versions[key]...)
	}
	return out, nil
}

// SaveKind keeps a stored kind's default role, like the Postgres store.
func (s *memStore) SaveKind(_ context.Context, k portfolio.Kind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.kinds[k.Key]; ok {
		k.DefaultRole = old.DefaultRole
	}
	s.kinds[k.Key] = k
	return nil
}

func (s *memStore) Kinds(_ context.Context) ([]portfolio.Kind, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := slices.Collect(maps.Values(s.kinds))
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// SaveSleeve refuses a sleeve whose kind is not stored, like the Postgres
// store's foreign key.
func (s *memStore) SaveSleeve(_ context.Context, sl portfolio.Sleeve) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.kinds[sl.Kind]; !ok {
		return fmt.Errorf("sleeve %s: kind %q: %w", sl.ID, sl.Kind, app.ErrNotFound)
	}
	s.sleeves[sl.ID] = sl
	return nil
}

func (s *memStore) Sleeves(_ context.Context) ([]portfolio.Sleeve, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]portfolio.Sleeve, 0, len(s.sleeves))
	for _, sl := range s.sleeves {
		out = append(out, sl)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *memStore) SaveAccount(_ context.Context, a ledger.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[a.ID] = a
	return nil
}

func (s *memStore) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.accounts[id]
	if !ok {
		return ledger.Account{}, fmt.Errorf("account %s: %w", id, app.ErrNotFound)
	}
	return a, nil
}

func (s *memStore) Accounts(_ context.Context) ([]ledger.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ledger.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *memStore) SaveAssertionCheck(_ context.Context, c *app.AssertionCheck) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := &c.Assertion
	stored := *c
	s.checks[checkKey{a.Account, a.Date, a.Book}] = &stored
	return nil
}

func (s *memStore) AssertionChecks(_ context.Context, account ledger.AccountID) ([]app.AssertionCheck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []app.AssertionCheck
	for k, c := range s.checks {
		if k.account == account {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := &out[i].Assertion, &out[j].Assertion
		if c := a.Date.Compare(b.Date); c != 0 {
			return c < 0
		}
		return a.Book.String() < b.Book.String()
	})
	return out, nil
}

// appendVersion appends v unless its key already has v.Version.
func (s *memStore) appendVersion(v *ledger.VersionedTransaction) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := v.Transaction.Key
	if v.Version <= len(s.versions[key]) {
		return fmt.Errorf("%s version %d: %w", key, v.Version, app.ErrVersionConflict)
	}
	if v.Version != len(s.versions[key])+1 {
		return fmt.Errorf("%s version %d: %w", key, v.Version, errVersionGap)
	}
	c := *v
	c.Transaction.Postings = slices.Clone(v.Transaction.Postings)
	s.versions[key] = append(s.versions[key], c)
	return nil
}

// appendOutside appends v as another import's committed work.
func (s *memStore) appendOutside(v *ledger.VersionedTransaction) error {
	if err := s.appendVersion(v); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outside = append(s.outside, *v)
	return nil
}
