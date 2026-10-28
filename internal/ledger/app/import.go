package app

import (
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

// VersionConflictError reports a transaction that lost the race for its next
// version MaxAppendAttempts times in a row. Nothing of it was appended;
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

// Import adds txs to the log and checks balance assertions (D-18). Every
// transaction and assertion is validated, and its account must be stored,
// before anything is written. Then each transaction, in order, is decided
// against all stored versions of its key: unchanged content is a no-op, a new
// key is appended as version 1, and changed content is appended as the next
// version, superseding the current one, which is kept. Importing the same
// input again changes nothing.
//
// Finally every account the import touched has its balance assertions
// checked, the stored ones and assertions, which replace stored ones for the
// same date and ledger account. A mismatch is stored as a flagged check and
// does not fail the import. If Import fails part way, what it appended stays
// and running it again completes it.
func (l *Ledger) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ImportSummary, error) {
	accounts, err := l.validateImport(ctx, txs, assertions)
	if err != nil {
		return ImportSummary{}, err
	}
	var sum ImportSummary
	for i := range txs {
		d, err := l.importOne(ctx, &txs[i])
		if err != nil {
			return ImportSummary{}, err
		}
		sum.count(d)
	}
	for _, id := range accounts {
		checks, err := l.checkAccount(ctx, id, assertions)
		if err != nil {
			return ImportSummary{}, err
		}
		sum.Checks = append(sum.Checks, checks...)
	}
	return sum, nil
}

// validateImport validates txs and assertions and returns their accounts, in
// ID order, after checking that each is stored.
func (l *Ledger) validateImport(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) ([]ledger.AccountID, error) {
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
		if _, err := l.account(ctx, id); err != nil {
			return nil, err
		}
	}
	return accounts, nil
}

// importOne decides t against its key's stored versions and appends it
// unless it is unchanged. On a version conflict it reloads and decides
// again, up to MaxAppendAttempts times.
func (l *Ledger) importOne(ctx context.Context, t *ledger.Transaction) (ledger.ImportDecision, error) {
	for range MaxAppendAttempts {
		versions, err := l.store.Versions(ctx, t.Key)
		if err != nil {
			return 0, fmt.Errorf("loading versions of %s: %w", t.Key, err)
		}
		res, err := ledger.DecideImport(versions, t)
		if err != nil {
			return 0, fmt.Errorf("importing %s: %w", t.Key, err)
		}
		if res.Decision == ledger.ImportUnchanged {
			return res.Decision, nil
		}
		err = l.store.Append(ctx, &ledger.VersionedTransaction{Transaction: *t, Version: res.Version})
		if err == nil {
			return res.Decision, nil
		}
		if !errors.Is(err, ErrVersionConflict) {
			return 0, fmt.Errorf("appending %s version %d: %w", t.Key, res.Version, err)
		}
	}
	return 0, &VersionConflictError{Key: t.Key, Attempts: MaxAppendAttempts}
}

// assertionKey identifies one balance assertion: its date and ledger
// account, within one account.
type assertionKey struct {
	date kernel.Date
	book ledger.Book
}

// checkAccount checks every balance assertion of account, its stored ones
// and those among incoming, against its log, stores each result and returns
// them in date and ledger account order.
func (l *Ledger) checkAccount(ctx context.Context, account ledger.AccountID, incoming []ledger.BalanceAssertion) ([]AssertionCheck, error) {
	log, err := l.store.Log(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("loading the log of %s: %w", account, err)
	}
	stored, err := l.store.AssertionChecks(ctx, account)
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
		if err := l.store.SaveAssertionCheck(ctx, &c); err != nil {
			return nil, fmt.Errorf("saving the balance assertion check of %s: %w", account, err)
		}
		checks = append(checks, c)
	}
	sortChecks(checks)
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
