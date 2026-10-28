package app_test

import (
	"context"
	"errors"
	"fmt"
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

// memStore is an in-memory app.Store with the same rules as the Postgres
// one: versions are only appended, numbered 1..n per key with no repeat.
type memStore struct {
	mu       sync.Mutex
	versions map[ledger.TransactionKey][]ledger.VersionedTransaction
	sleeves  map[portfolio.SleeveID]portfolio.Sleeve
	accounts map[ledger.AccountID]ledger.Account
	checks   map[checkKey]*app.AssertionCheck
	// beforeAppend, when set, runs at the start of every Append, standing in
	// for another import that appends between this one's read and write.
	beforeAppend func(s *memStore, v *ledger.VersionedTransaction) error
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
		sleeves:  map[portfolio.SleeveID]portfolio.Sleeve{},
		accounts: map[ledger.AccountID]ledger.Account{},
		checks:   map[checkKey]*app.AssertionCheck{},
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

func (s *memStore) SaveSleeve(_ context.Context, sl portfolio.Sleeve) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
