package app

import (
	"context"
	"errors"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrConnectorFailed reports a connector that answered a sync with an error
// of its own, such as a broker's API refusing its credentials; the error
// carries the connector's answer, which the core never stores or logs
// (GR-12).
var ErrConnectorFailed = errors.New("the connector answered the sync with an error")

// ErrDeliveryTimeout reports a delivery whose database transaction ran past its
// timeout (D-18, D-198): a statement of it did not finish in time, so the
// transaction rolled back and nothing of the delivery was written. A delivery
// is all or nothing, so it is never split to fit.
var ErrDeliveryTimeout = errors.New("the delivery's transaction timed out")

// SyncErrorClassOf returns the class of err, a failed sync attempt's error,
// from the known errors in its chain and never from its text (GR-12): a
// delivery's transaction that ran past its timeout, a timeout before an
// interruption, then an instance that can't be reached,
// a connector's own error, an answer that doesn't translate, a delivery
// the core refuses, and one the ledger won't book; anything else, such as
// the database, is SyncErrorOther. A nil error has the zero class.
func SyncErrorClassOf(err error) ingest.SyncErrorClass {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrDeliveryTimeout):
		return ingest.SyncErrorDeliveryTimeout
	case errors.Is(err, context.DeadlineExceeded):
		return ingest.SyncErrorTimeout
	case errors.Is(err, context.Canceled):
		return ingest.SyncErrorInterrupted
	case errors.Is(err, ErrConnectorUnavailable):
		return ingest.SyncErrorUnavailable
	case errors.Is(err, ErrConnectorFailed):
		return ingest.SyncErrorConnector
	case errors.Is(err, ErrInvalidDelivery):
		return ingest.SyncErrorInvalidDelivery
	case errors.Is(err, ingest.ErrUndeclaredInstitution), errors.Is(err, ingest.ErrAccountNotBound),
		errors.Is(err, ingest.ErrInvalidBatch), errors.Is(err, ingest.ErrInvalidInstitution):
		return ingest.SyncErrorRefused
	case isBookingError(err):
		return ingest.SyncErrorBooking
	}
	return ingest.SyncErrorOther
}

// isBookingError reports whether err is the ledger refusing what a delivery
// asked it to book.
func isBookingError(err error) bool {
	for _, target := range []error{
		ledger.ErrInvalidTransaction, ledger.ErrInvalidAccount, ledger.ErrInvalidBook, ledger.ErrInvalidAssertion,
		ledger.ErrInvalidISIN, ledger.ErrInvalidInstrument, ledger.ErrAlreadySeeded, ledger.ErrBookedBeforeGoLive,
		ledger.ErrGoLiveMismatch, ledger.ErrNothingToDelete, ErrSeedingOutsideOpening,
	} {
		if errors.Is(err, target) {
			return true
		}
	}
	_, unbalanced := errors.AsType[*ledger.UnbalancedError](err)
	return unbalanced
}
