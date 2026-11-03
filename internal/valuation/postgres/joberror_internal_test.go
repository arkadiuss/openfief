package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// errStoreDown is a store's failure that carries no value of its own.
var errStoreDown = errors.New("store down")

// TestJobError checks what a worker hands River (GR-12): nothing for no
// failure, River's own snooze as it is, so River still snoozes the job, and
// any other failure as a *JobError naming the job and the failure's kind,
// never the values it carried.
func TestJobError(t *testing.T) {
	t.Parallel()
	if err := jobError("revaluation job", nil); err != nil {
		t.Errorf("jobError(nil) = %v", err)
	}
	snoozed := jobError("revaluation job", fmt.Errorf("revaluation job: changes kept arriving: %w", river.JobSnooze(0)))
	if _, ok := errors.AsType[*rivertype.JobSnoozeError](snoozed); !ok {
		t.Errorf("a snooze became %v", snoozed)
	}
	got := jobError("revaluation job", fmt.Errorf("lot of PLTEST000010 at 1234.56 PLN: %w", ErrCorrupt))
	if want := "revaluation job: " + ErrCorrupt.Error(); got.Error() != want {
		t.Errorf("jobError = %q, want %q", got, want)
	}
	unknown := jobError("end-of-day job", fmt.Errorf("row 1234.56: %w", errStoreDown))
	if want := "end-of-day job: unclassified error of type *errors.errorString"; unknown.Error() != want {
		t.Errorf("jobError = %q, want %q", unknown, want)
	}
}
