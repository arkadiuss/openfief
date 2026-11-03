package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// loggableErrors are the errors whose own text names a rule, a refusal or a
// failure and never a value, so LogSafe may write them. What wraps them,
// such as a connector's reason, a transaction's amount or an ISIN, is left
// out.
func loggableErrors() []error {
	return []error{
		context.Canceled, context.DeadlineExceeded,
		ErrSyncRunning, ErrEntryRefused, ErrNoEntries, ErrConnectorUnavailable, ErrInvalidDelivery,
		ErrNoConnector, ErrNoCrossCheckStore, ErrInvalidUpload, ErrFileTooLarge, ErrSeedingOutsideOpening,
		ingest.ErrUndeclaredInstitution, ingest.ErrAccountNotBound, ingest.ErrInvalidBatch, ingest.ErrInvalidInstitution,
		ingest.ErrInvalidInstanceName, ingest.ErrStartedFromHistory, ingest.ErrOpeningCashEntered,
		ingest.ErrOpeningCashOutsideOpening,
		ledger.ErrInvalidTransaction, ledger.ErrInvalidAccount, ledger.ErrInvalidBook, ledger.ErrInvalidAssertion,
		ledger.ErrInvalidISIN, ledger.ErrInvalidInstrument, ledger.ErrAlreadySeeded, ledger.ErrBookedBeforeGoLive,
		ledger.ErrGoLiveMismatch, ledger.ErrNothingToDelete,
		valuation.ErrUnknownMarketDataKind,
		kernel.ErrInvalidDecimal, kernel.ErrInvalidCurrency, kernel.ErrInvalidDate,
	}
}

// LogSafe describes err for a log line without any value it may carry
// (GR-12): the messages of the known errors in its chain, which name a rule
// and never an amount, a holding, an ISIN or an account number, joined by
// "; ", or else only the innermost error's Go type, such as
// "*pgconn.PgError", whose text may quote a row. Nil is "".
func LogSafe(err error) string {
	if err == nil {
		return ""
	}
	var known []string
	for _, k := range loggableErrors() {
		if errors.Is(err, k) {
			known = append(known, k.Error())
		}
	}
	if len(known) > 0 {
		return strings.Join(known, "; ")
	}
	return fmt.Sprintf("unclassified error of type %T", innermost(err))
}

// innermost follows err's chain, the first of joined errors at each step,
// to the error that wraps nothing.
func innermost(err error) error {
	for {
		if next := errors.Unwrap(err); next != nil {
			err = next
			continue
		}
		var joined interface{ Unwrap() []error }
		if errors.As(err, &joined) {
			if errs := joined.Unwrap(); len(errs) > 0 && errs[0] != nil {
				err = errs[0]
				continue
			}
		}
		return err
	}
}
