package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of an account's start (D-136).
var (
	// ErrStartedFromHistory reports an opening upload for an account that
	// isn't seeded but already holds transactions: it started from a
	// statement of its whole history, and seeding it now would count its
	// positions twice.
	ErrStartedFromHistory = errors.New("the account started from its whole history: upload its statements in transactions mode")
	// ErrOpeningCashOutsideOpening reports an opening cash balance sent with
	// an upload that doesn't seed the account.
	ErrOpeningCashOutsideOpening = errors.New("an opening cash balance goes only with an opening upload")
	// ErrOpeningCashEntered reports an opening cash balance sent for an
	// account already seeded: the owner enters it once, with the upload
	// that seeds the account.
	ErrOpeningCashEntered = errors.New("the account is already seeded: its opening cash is entered once, with the upload that seeds it")
)

// StartMode returns the ingest mode of a statement upload for an account,
// as UploadMode chooses it, for an account that starts either from its
// institution's current positions or from a statement of its whole history
// (D-136, D-109). seeded says the account holds a seeding transaction, and
// booked that it holds any transaction. An account that is not seeded but
// is booked started from its whole history: the core chooses transactions
// for it, and an opening override fails with ErrStartedFromHistory.
// Otherwise it is UploadMode's choice.
func StartMode(seeded, booked bool, override Mode) (Mode, error) {
	if seeded || !booked {
		return UploadMode(seeded, override), nil
	}
	if override == ModeOpening {
		return 0, ErrStartedFromHistory
	}
	return ModeTransactions, nil
}

// CheckOpeningCash checks the cash balance the owner entered with an upload
// in mode for an account, seeded or not: none (nil) goes with any upload,
// and an entered one, even of zero, only with the opening upload of an
// account not seeded yet, which books it as the account's opening cash:
// the owner enters it once (D-136).
func CheckOpeningCash(cash *kernel.Money, mode Mode, seeded bool) error {
	switch {
	case cash == nil:
		return nil
	case mode != ModeOpening:
		return fmt.Errorf("%w, not a %s one", ErrOpeningCashOutsideOpening, mode)
	case seeded:
		return ErrOpeningCashEntered
	}
	return nil
}
