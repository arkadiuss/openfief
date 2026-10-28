package app

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// AssertionStatus is the outcome of checking a balance assertion.
type AssertionStatus int

// The assertion statuses. The zero status is not a status.
const (
	// AssertionMatched means the derived balance equals the reported one.
	AssertionMatched AssertionStatus = iota + 1
	// AssertionFlagged means they differ: a flagged error for the owner to
	// resolve (D-18). The import that found it is not rejected.
	AssertionFlagged
)

// String returns "matched" or "flagged", or a placeholder for an invalid
// status.
func (s AssertionStatus) String() string {
	switch s {
	case AssertionMatched:
		return "matched"
	case AssertionFlagged:
		return "flagged"
	}
	return fmt.Sprintf("AssertionStatus(%d)", int(s))
}

// AssertionCheck is a balance assertion with the result of its latest check
// against the balance derived from the log.
type AssertionCheck struct {
	Assertion ledger.BalanceAssertion
	Status    AssertionStatus
	// Derived is the balance derived from the current versions at the end of
	// the assertion's date, in the assertion's unit.
	Derived kernel.Decimal
}

// Flagged reports whether the check found a mismatch.
func (c *AssertionCheck) Flagged() bool { return c.Status == AssertionFlagged }

// checkAssertion checks a against log: matched, or flagged with the derived
// balance on a mismatch. Any other failure, such as an invalid assertion or
// log, is an error.
func checkAssertion(log []ledger.VersionedTransaction, a *ledger.BalanceAssertion) (AssertionCheck, error) {
	err := ledger.CheckBalanceAssertion(log, a)
	if err == nil {
		return AssertionCheck{Assertion: *a, Status: AssertionMatched, Derived: a.Balance}, nil
	}
	if mismatch, ok := errors.AsType[*ledger.AssertionMismatchError](err); ok {
		return AssertionCheck{Assertion: *a, Status: AssertionFlagged, Derived: mismatch.Derived}, nil
	}
	return AssertionCheck{}, fmt.Errorf("checking balance assertion of %s: %w", a.Account, err)
}
