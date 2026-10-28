package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// MaxAppendAttempts is how many times Import decides and appends one
// transaction before it gives up on a version conflict. Each retry reloads
// the key's versions, so content another import already appended becomes
// unchanged instead of a duplicate version.
const MaxAppendAttempts = 3

// ImportSummary is what an import did.
type ImportSummary struct {
	// Unchanged, New and NewVersions count the incoming transactions by
	// import decision.
	Unchanged   int
	New         int
	NewVersions int
	// EarliestAffected is, per account whose current transactions the import
	// changed, the earliest date it changed: the date of every appended
	// version and of every version it superseded, in that version's account.
	// Holdings, lots and valuations from that date on may differ (D-36).
	EarliestAffected map[ledger.AccountID]kernel.Date
	// AlreadySeeded are the accounts a Seed found already seeded with the
	// same content, in ID order: for them it changed nothing.
	AlreadySeeded []ledger.AccountID
	// Checks are the balance assertions of every account the import touched,
	// stored and incoming, as checked after it, in account, date and ledger
	// account order.
	Checks []AssertionCheck
}

// Flagged returns the checks that found a mismatch.
func (s *ImportSummary) Flagged() []AssertionCheck {
	var out []AssertionCheck
	for i := range s.Checks {
		if s.Checks[i].Flagged() {
			out = append(out, s.Checks[i])
		}
	}
	return out
}

// count adds one transaction with decision d.
func (s *ImportSummary) count(d ledger.ImportDecision) {
	switch d {
	case ledger.ImportUnchanged:
		s.Unchanged++
	case ledger.ImportNew:
		s.New++
	case ledger.ImportNewVersion:
		s.NewVersions++
	}
}

// affect records that account's current transactions changed on day.
func (s *ImportSummary) affect(account ledger.AccountID, day kernel.Date) {
	if s.EarliestAffected == nil {
		s.EarliestAffected = map[ledger.AccountID]kernel.Date{}
	}
	if earliest, ok := s.EarliestAffected[account]; !ok || day.Before(earliest) {
		s.EarliestAffected[account] = day
	}
}

// ImportHook is work that must commit or roll back with an import, such as
// enqueueing a revaluation from sum.EarliestAffected (D-36). Import calls it
// after appending and checking, inside the import's database transaction,
// with the Store of that transaction; an error rolls the whole import back.
type ImportHook func(ctx context.Context, tx Store, sum *ImportSummary) error

// VersionConflictError reports a transaction that lost the race for its next
// version MaxAppendAttempts times in a row. The import was rolled back;
// importing it again is safe.
type VersionConflictError struct {
	Key      ledger.TransactionKey
	Attempts int
}

// Error names the key and the number of attempts.
func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("importing %s: version conflict after %d attempts", e.Key, e.Attempts)
}

// Unwrap returns ErrVersionConflict.
func (e *VersionConflictError) Unwrap() error { return ErrVersionConflict }

// Import adds txs to the log and checks balance assertions (D-18), all in
// one database transaction: it commits whole or not at all. Every
// transaction and assertion is validated, and its account must be stored.
// Then each transaction, in key order (the given order within one key), is
// decided against all stored versions of its key: unchanged content is a
// no-op, a new key is appended as version 1, and changed content is appended
// as the next version, superseding the current one, which is kept. Importing
// the same input again changes nothing.
//
// Then every account the import touched has its balance assertions checked,
// the stored ones and assertions, which replace stored ones for the same
// date and ledger account. A mismatch is stored as a flagged check and does
// not fail the import. Last, the Ledger's import hooks run in the same
// transaction.
func (l *Ledger) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ImportSummary, error) {
	return l.importAll(ctx, txs, assertions, false)
}

// importAll is Import, and with seed set, Seed.
func (l *Ledger) importAll(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion, seed bool) (ImportSummary, error) {
	ordered := slices.Clone(txs)
	slices.SortStableFunc(ordered, func(a, b ledger.Transaction) int { return compareKeys(a.Key, b.Key) })
	var sum ImportSummary
	err := l.store.InTx(ctx, func(tx Store) error {
		var err error
		sum, err = importTx(ctx, tx, ordered, assertions, seed)
		if err != nil {
			return err
		}
		for _, hook := range l.onImport {
			if err := hook(ctx, tx, &sum); err != nil {
				return fmt.Errorf("import hook: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		if seed {
			return ImportSummary{}, fmt.Errorf("seeding: %w", err)
		}
		return ImportSummary{}, fmt.Errorf("importing: %w", err)
	}
	return sum, nil
}

// compareKeys orders transaction keys by source, then external ID. Appending
// in one order everywhere keeps two concurrent imports from waiting on each
// other's keys in a cycle.
func compareKeys(a, b ledger.TransactionKey) int {
	if c := cmp.Compare(a.Source, b.Source); c != 0 {
		return c
	}
	return cmp.Compare(a.ExternalID, b.ExternalID)
}

// importTx is Import's work inside its database transaction tx, and with
// seed set, Seed's: it first refuses to seed an account that is already
// seeded.
func importTx(ctx context.Context, tx Store, txs []ledger.Transaction, assertions []ledger.BalanceAssertion, seed bool) (ImportSummary, error) {
	accounts, err := validateImport(ctx, tx, txs, assertions)
	if err != nil {
		return ImportSummary{}, err
	}
	var sum ImportSummary
	if seed {
		if sum.AlreadySeeded, err = guardSeed(ctx, tx, txs); err != nil {
			return ImportSummary{}, err
		}
	}
	for i := range txs {
		t := &txs[i]
		d, superseded, err := importOne(ctx, tx, t)
		if err != nil {
			return ImportSummary{}, err
		}
		sum.count(d)
		if d != ledger.ImportUnchanged {
			sum.affect(t.Account, t.Date)
		}
		if superseded != nil {
			sum.affect(superseded.Account, superseded.Date)
		}
	}
	for _, id := range accounts {
		checks, err := checkAccount(ctx, tx, id, assertions)
		if err != nil {
			return ImportSummary{}, err
		}
		sum.Checks = append(sum.Checks, checks...)
	}
	return sum, nil
}

// validateImport validates txs and assertions and returns their accounts, in
// ID order, after checking that each is stored.
func validateImport(ctx context.Context, store Registry, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) ([]ledger.AccountID, error) {
	seen := map[ledger.AccountID]bool{}
	for i := range txs {
		if err := txs[i].Validate(); err != nil {
			return nil, fmt.Errorf("importing: %w", err)
		}
		seen[txs[i].Account] = true
	}
	for i := range assertions {
		if err := assertions[i].Validate(); err != nil {
			return nil, fmt.Errorf("importing: %w", err)
		}
		seen[assertions[i].Account] = true
	}
	accounts := make([]ledger.AccountID, 0, len(seen))
	for id := range seen {
		accounts = append(accounts, id)
	}
	slices.Sort(accounts)
	for _, id := range accounts {
		if _, err := lookupAccount(ctx, store, id); err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// importOne decides t against its key's stored versions and appends it
// unless it is unchanged. It returns the decision and, for a new version,
// the version it superseded. On a version conflict it reloads and decides
// again, up to MaxAppendAttempts times.
func importOne(ctx context.Context, store TransactionLog, t *ledger.Transaction) (ledger.ImportDecision, *ledger.Transaction, error) {
	for range MaxAppendAttempts {
		versions, err := store.Versions(ctx, t.Key)
		if err != nil {
			return 0, nil, fmt.Errorf("loading versions of %s: %w", t.Key, err)
		}
		res, err := ledger.DecideImport(versions, t)
		if err != nil {
			return 0, nil, fmt.Errorf("importing %s: %w", t.Key, err)
		}
		if res.Decision == ledger.ImportUnchanged {
			return res.Decision, nil, nil
		}
		err = store.Append(ctx, &ledger.VersionedTransaction{Transaction: *t, Version: res.Version})
		if err == nil {
			return res.Decision, supersededBy(versions, res.Supersedes), nil
		}
		if !errors.Is(err, ErrVersionConflict) {
			return 0, nil, fmt.Errorf("importing %s: %w", t.Key, err)
		}
	}
	return 0, nil, &VersionConflictError{Key: t.Key, Attempts: MaxAppendAttempts}
}

// supersededBy returns version n among versions, or nil for 0.
func supersededBy(versions []ledger.VersionedTransaction, n int) *ledger.Transaction {
	for i := range versions {
		if n != 0 && versions[i].Version == n {
			return &versions[i].Transaction
		}
	}
	return nil
}

// assertionKey identifies one balance assertion: its date and ledger
// account, within one account.
type assertionKey struct {
	date kernel.Date
	book ledger.Book
}

// checkAccount checks every balance assertion of account in store, its
// stored ones and those among incoming, against its log, stores each result
// and returns them in date and ledger account order.
func checkAccount(ctx context.Context, store Store, account ledger.AccountID, incoming []ledger.BalanceAssertion) ([]AssertionCheck, error) {
	log, err := store.Log(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("loading the log of %s: %w", account, err)
	}
	stored, err := store.AssertionChecks(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("loading balance assertions of %s: %w", account, err)
	}
	byKey := map[assertionKey]ledger.BalanceAssertion{}
	for i := range stored {
		a := &stored[i].Assertion
		byKey[assertionKey{a.Date, a.Book}] = *a
	}
	for i := range incoming {
		if a := &incoming[i]; a.Account == account {
			byKey[assertionKey{a.Date, a.Book}] = *a
		}
	}
	checks := make([]AssertionCheck, 0, len(byKey))
	for _, a := range byKey {
		c, err := checkAssertion(log, &a)
		if err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	sortChecks(checks)
	for i := range checks {
		if err := store.SaveAssertionCheck(ctx, &checks[i]); err != nil {
			return nil, fmt.Errorf("saving the balance assertion check of %s: %w", account, err)
		}
	}
	return checks, nil
}

// sortChecks orders checks by date, then ledger account.
func sortChecks(checks []AssertionCheck) {
	sort.Slice(checks, func(i, j int) bool {
		a, b := &checks[i].Assertion, &checks[j].Assertion
		if c := a.Date.Compare(b.Date); c != 0 {
			return c < 0
		}
		return a.Book.String() < b.Book.String()
	})
}
