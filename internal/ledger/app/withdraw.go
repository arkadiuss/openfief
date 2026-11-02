package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrNotHoldingAssertion reports a withdrawal of a balance assertion of
// anything but a holding: cash assertions, the owner's cash checkpoints
// among them, are never withdrawn.
var ErrNotHoldingAssertion = errors.New("only a holding's balance assertion is withdrawn")

// WithdrawAssertions removes the checks of holding assertions that no
// longer hold, in one transaction: those of the instrument the owner
// corrected a ticker's ISIN from (D-153), whose holding the correction
// moved to the corrected instrument. The transaction log is untouched; an
// assertion not stored is skipped. A cash assertion is refused with
// ErrNotHoldingAssertion and nothing is removed.
func (l *Ledger) WithdrawAssertions(ctx context.Context, assertions []ledger.BalanceAssertion) error {
	for i := range assertions {
		if a := &assertions[i]; a.Book.Kind != ledger.BookHolding {
			return fmt.Errorf("withdrawing %s of %s on %s: %w", a.Book, a.Account, a.Date, ErrNotHoldingAssertion)
		}
	}
	err := l.store.InTx(ctx, func(tx Store) error {
		for i := range assertions {
			a := &assertions[i]
			if err := tx.DeleteAssertionCheck(ctx, a.Account, a.Date, a.Book); err != nil {
				return fmt.Errorf("withdrawing %s of %s on %s: %w", a.Book, a.Account, a.Date, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("withdrawing balance assertions: %w", err)
	}
	return nil
}
