package postgres

import (
	"context"
	"errors"

	"github.com/riverqueue/river/rivertype"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// JobError is what a worker hands River for a failed job: the job's kind
// and the failure described without any value it carried (GR-12). River
// logs a failed job's error in full and keeps it in river_job.errors, and a
// revaluation's failure can quote a lot's cost and quantity, a transaction's
// sums or an ISIN, so neither ever sees more than this.
type JobError struct {
	// Job names the job, such as "revaluation job".
	Job string
	// Safe is kernel.LogSafe's description of the failure: the known
	// errors in its chain, or the innermost error's type.
	Safe string
}

// Error names the job and the failure by its kind only.
func (e *JobError) Error() string { return e.Job + ": " + e.Safe }

// jobErrors are the errors whose own text names a rule or a failure and
// never a value, so a job's error may name them.
func jobErrors() []error {
	return []error{
		context.Canceled, context.DeadlineExceeded,
		ErrCorrupt, ErrInvalidConfig,
		app.ErrInvalidInput,
		valuation.ErrBondLot, valuation.ErrNoRate, valuation.ErrConflictingRates, valuation.ErrNoPrice,
		valuation.ErrUnknownIndexKind, valuation.ErrUnknownBasis, valuation.ErrUnknownMissingReason,
		ledger.ErrInvalidTransaction,
	}
}

// jobError returns err as a worker hands it to River: nil as nil, River's
// own snooze, which carries no value and River must see as such, as it is,
// and anything else as a *JobError of job (GR-12).
func jobError(job string, err error) error {
	if err == nil {
		return nil
	}
	if snooze, ok := errors.AsType[*rivertype.JobSnoozeError](err); ok {
		return snooze
	}
	return &JobError{Job: job, Safe: kernel.LogSafe(err, jobErrors()...)}
}
