package ingest

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownSyncErrorClass reports a stored or requested sync error class
// that is not one of the SyncErrorClass names.
var ErrUnknownSyncErrorClass = errors.New("unknown sync error class")

// SyncErrorClass says, in one word, why a connector sync attempt failed,
// without any of the error's own text, which may quote an amount, an ISIN
// or an account number (GR-12). It is what the history of an instance's
// syncs keeps of a failure.
type SyncErrorClass int

// The sync error classes. The zero class is none: the attempt succeeded.
const (
	// SyncErrorTimeout: the attempt ran past its timeout, such as a
	// connector that didn't answer in time.
	SyncErrorTimeout SyncErrorClass = iota + 1
	// SyncErrorUnavailable: the connector couldn't be reached, such as an
	// instance whose pod isn't running.
	SyncErrorUnavailable
	// SyncErrorConnector: the connector answered with an error of its own,
	// such as a broker's API refusing its credentials.
	SyncErrorConnector
	// SyncErrorInvalidDelivery: the connector's answer doesn't translate
	// into a delivery, such as a malformed amount.
	SyncErrorInvalidDelivery
	// SyncErrorRefused: the core refused the delivery, such as one naming
	// an institution its manifest doesn't declare (D-182, D-185).
	SyncErrorRefused
	// SyncErrorBooking: the ledger refused to book what was delivered.
	SyncErrorBooking
	// SyncErrorInterrupted: the core stopped while the attempt ran.
	SyncErrorInterrupted
	// SyncErrorOther: any other failure, such as the database.
	SyncErrorOther
)

// String returns the class's name, as stored and served, such as
// "timeout"; "" for the zero class, or a placeholder for an invalid one.
func (c SyncErrorClass) String() string {
	switch c {
	case SyncErrorTimeout:
		return "timeout"
	case SyncErrorUnavailable:
		return "unavailable"
	case SyncErrorConnector:
		return "connector-error"
	case SyncErrorInvalidDelivery:
		return "invalid-delivery"
	case SyncErrorRefused:
		return "refused"
	case SyncErrorBooking:
		return "booking"
	case SyncErrorInterrupted:
		return "interrupted"
	case SyncErrorOther:
		return "other"
	}
	if c == 0 {
		return ""
	}
	return fmt.Sprintf("SyncErrorClass(%d)", int(c))
}

// ParseSyncErrorClass returns the class named s, the zero class for "", or
// an error wrapping ErrUnknownSyncErrorClass.
func ParseSyncErrorClass(s string) (SyncErrorClass, error) {
	if s == "" {
		return 0, nil
	}
	for c := SyncErrorTimeout; c <= SyncErrorOther; c++ {
		if c.String() == s {
			return c, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownSyncErrorClass, s)
}

// SyncAttempt is one attempt to sync a connector instance (D-49): the batch
// it ran, when it started and ended, and, for a failed one, its error class,
// how many attempts in a row have failed with it, and when the next one is
// due. Nothing in it is a value from the delivery (GR-12).
type SyncAttempt struct {
	// ID numbers the attempts in the order they ended.
	ID       int64
	Instance string
	Batch    BatchID
	Mode     Mode
	Started  time.Time
	Ended    time.Time
	// Class is why the attempt failed; zero for an attempt that succeeded.
	Class SyncErrorClass
	// Failures counts the attempts in a row that failed, this one
	// included; zero after a success.
	Failures int
	// NextSync is when the attempt scheduled the instance's next sync.
	NextSync time.Time
}

// Failed reports whether the attempt failed.
func (a *SyncAttempt) Failed() bool { return a.Class != 0 }

// SyncStatus is where an instance's syncs stand: its latest attempt, its
// latest success, and, while it fails, how many attempts in a row have
// failed since when, and why the latest did.
type SyncStatus struct {
	Instance string
	// LastAttempt is when the latest attempt ended; zero before any.
	LastAttempt time.Time
	// LastSuccess is when the latest successful attempt ended; zero before
	// any.
	LastSuccess time.Time
	// Failures counts the failed attempts since the latest success, or
	// since the first attempt.
	Failures int
	// FailingSince is when the first of those failed attempts started;
	// zero while Failures is zero.
	FailingSince time.Time
	// LastClass is the latest failed attempt's class while Failures is
	// positive; zero otherwise.
	LastClass SyncErrorClass
	// NextSync is when the next attempt is due; zero when it is due at once.
	NextSync time.Time
}

// What "keeps failing" means (KeepsFailing): enough failed attempts in a
// row, over long enough that one restart of a plugin or a short outage of
// the institution is retried without bothering the owner. With the core's
// default retries (5 minutes, doubling) the third failure in a row ends
// about 15 minutes after the first one started.
const (
	// KeepsFailingAttempts is how many attempts in a row must fail.
	KeepsFailingAttempts = 3
	// KeepsFailingFor is how long, from the start of the first of them to
	// the end of the latest, they must span.
	KeepsFailingFor = 15 * time.Minute
)

// KeepsFailing reports whether s is a sync that keeps failing (D-196): at
// least KeepsFailingAttempts attempts in a row failed since the latest
// success, and they span at least KeepsFailingFor from the first one's
// start to the latest one's end. A single success clears it.
func (s *SyncStatus) KeepsFailing() bool {
	if s.Failures < KeepsFailingAttempts || s.FailingSince.IsZero() {
		return false
	}
	return s.LastAttempt.Sub(s.FailingSince) >= KeepsFailingFor
}

// AttemptQuery selects sync attempts, newest first: by when they ended,
// then by ID.
type AttemptQuery struct {
	// Instance is the one instance whose attempts to return; empty for
	// every instance.
	Instance string
	// FailedOnly returns only failed attempts.
	FailedOnly bool
	// Before, unless zero, returns only the attempts that ended before it,
	// or at it with an ID below BeforeID: those after the last one of a
	// previous page.
	Before   time.Time
	BeforeID int64
	// Limit is the most attempts to return.
	Limit int
}

// BatchRecord is a completed import batch as the batch history keeps it
// (D-49): its connector instance, the ingest path that ran it, such as
// "sync" or "upload", when it finished, and what it counted. Only IDs,
// counts and times (GR-12).
type BatchRecord struct {
	ID          BatchID
	Instance    string
	Source      string
	Mode        Mode
	Finished    time.Time
	Institution InstitutionID
	// Transactions counts what the mapping produced, and New, NewVersions
	// and Unchanged what the ledger did with them.
	Transactions int
	New          int
	NewVersions  int
	Unchanged    int
	// FailedAssertions counts balance assertions that didn't match.
	FailedAssertions int
	// Problems counts the rows the mapping left out, and Unsupported those
	// of them OpenFief can't book yet.
	Problems    int
	Unsupported int
	// Accounts are the accounts it booked into, in ID order: every mapped
	// account a statement of several accounts books (D-109, D-140); none
	// for a batch completed before the history kept them.
	Accounts []ledger.AccountID
}

// BatchQuery selects completed import batches, newest first: by when they
// finished, then by ID in byte order.
type BatchQuery struct {
	// Before, unless zero, returns only the batches that finished before
	// it, or at it with an ID before BeforeID: those after the last one of
	// a previous page.
	Before   time.Time
	BeforeID BatchID
	// Limit is the most batches to return.
	Limit int
}
