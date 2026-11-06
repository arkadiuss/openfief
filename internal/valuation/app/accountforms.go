package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// RecordAccountForm records account's form (D-147), one the pack offers
// for its wrapper w (Pack.AccountForms), such as an OKI account's brokerage
// account. No stored value depends on it, so nothing is revalued. A form w
// doesn't offer is an error wrapping ErrInvalidInput.
func (r *InputRecorder) RecordAccountForm(ctx context.Context, account ledger.AccountID, w jurisdiction.Wrapper, form jurisdiction.AccountForm) error {
	if !slices.Contains(r.pack.AccountForms(w), form) {
		return fmt.Errorf("%w: %q is not a form of a %s account", ErrInvalidInput, form, w)
	}
	if err := r.recorder.SaveAccountForm(ctx, account, form); err != nil {
		return fmt.Errorf("recording the form of %s: %w", account, err)
	}
	return nil
}

// RecordEarlierPayments records the total paid into account before its
// history is known (D-149), which its early return is taxed above when its
// wrapper w taxes what is withdrawn above what was paid in
// (Pack.TaxesAbovePaidIn), such as IKE. Totals are taxed when read, so
// nothing is revalued. Another wrapper, a negative amount, or an amount
// not in the pack's reporting currency is an error wrapping
// ErrInvalidInput.
func (r *InputRecorder) RecordEarlierPayments(ctx context.Context, account ledger.AccountID, w jurisdiction.Wrapper, paid kernel.Money) error {
	if !r.pack.TaxesAbovePaidIn(w) {
		return fmt.Errorf("%w: a %s account takes no earlier payments", ErrInvalidInput, w)
	}
	if paid.Sign() < 0 || !paid.Currency().Equal(r.pack.ReportingCurrency()) {
		return fmt.Errorf("%w: earlier payments of %s must be an amount of %s, not below zero", ErrInvalidInput, paid, r.pack.ReportingCurrency())
	}
	if err := r.recorder.SaveEarlierPayments(ctx, account, paid); err != nil {
		return fmt.Errorf("recording the earlier payments of %s: %w", account, err)
	}
	return nil
}
