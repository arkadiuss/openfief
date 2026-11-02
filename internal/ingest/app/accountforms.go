package app

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors of an account's form and earlier payments, each also wrapping
// ledger.ErrInvalidAccount.
var (
	// ErrAccountForm reports a form the account's wrapper doesn't offer, no
	// form for a wrapper that has forms, or a change to a form already set,
	// which the owner sets when creating the account (D-147).
	ErrAccountForm = errors.New("invalid account form")
	// ErrEarlierPayments reports earlier payments for a wrapper that takes
	// none, below zero or in another currency than the pack's (D-149).
	ErrEarlierPayments = errors.New("invalid earlier payments")
)

// changedTaxDetails checks entry's form and earlier payments as c changes
// them, entry already in its new wrapper, and returns the form to record
// (changedForm).
func (a *Accounts) changedTaxDetails(entry *AccountEntry, c *AccountChange) (jurisdiction.AccountForm, error) {
	form, err := a.changedForm(entry, c.Form)
	if err != nil {
		return "", err
	}
	if err := a.checkEarlierPayments(entry.Wrapper, c.EarlierPayments); err != nil {
		return "", err
	}
	return form, nil
}

// checkForm checks the form a new account in w is given: one of w's when
// it has forms, such as OKI (D-147), none otherwise.
func (a *Accounts) checkForm(w jurisdiction.Wrapper, form jurisdiction.AccountForm) error {
	forms := a.Pack.AccountForms(w)
	switch {
	case len(forms) == 0 && form != "":
		return fmt.Errorf("%w: %w: a %s account has no form", ledger.ErrInvalidAccount, ErrAccountForm, w)
	case len(forms) > 0 && form == "":
		return fmt.Errorf("%w: %w: a %s account needs its form, one of %v", ledger.ErrInvalidAccount, ErrAccountForm, w, forms)
	case len(forms) > 0 && !slices.Contains(forms, form):
		return fmt.Errorf("%w: %w: %q is not a form of a %s account", ledger.ErrInvalidAccount, ErrAccountForm, form, w)
	}
	return nil
}

// changedForm returns the form to record for entry, already changed to its
// new wrapper, when the owner gives form: none when its wrapper has no
// forms or it keeps the one it has, which never changes; form for an
// account in a wrapper with forms that has none yet, such as one moved into
// OKI, which then needs one. entry takes the new form.
func (a *Accounts) changedForm(entry *AccountEntry, form jurisdiction.AccountForm) (jurisdiction.AccountForm, error) {
	forms := a.Pack.AccountForms(entry.Wrapper)
	switch {
	case len(forms) == 0:
		if form != "" {
			return "", fmt.Errorf("%w: %w: a %s account has no form", ledger.ErrInvalidAccount, ErrAccountForm, entry.Wrapper)
		}
		return "", nil
	case entry.Form != "":
		if form != "" && form != entry.Form {
			return "", fmt.Errorf("%w: %w: the form is set when the account is created and never changes", ledger.ErrInvalidAccount, ErrAccountForm)
		}
		return "", nil
	}
	if err := a.checkForm(entry.Wrapper, form); err != nil {
		return "", err
	}
	entry.Form = form
	return form, nil
}

// checkEarlierPayments checks paid, the earlier payments given for an
// account in w, when not nil: only for a wrapper that taxes what is
// withdrawn above what was paid in, such as IKE (D-149), in the pack's
// currency and not below zero.
func (a *Accounts) checkEarlierPayments(w jurisdiction.Wrapper, paid *kernel.Money) error {
	if paid == nil {
		return nil
	}
	if !a.Pack.TaxesAbovePaidIn(w) {
		return fmt.Errorf("%w: %w: a %s account takes no earlier payments", ledger.ErrInvalidAccount, ErrEarlierPayments, w)
	}
	if paid.Sign() < 0 || !paid.Currency().Equal(a.Pack.ReportingCurrency()) {
		return fmt.Errorf("%w: %w: %s must be an amount of %s, not below zero", ledger.ErrInvalidAccount, ErrEarlierPayments, paid, a.Pack.ReportingCurrency())
	}
	return nil
}

// recordTaxDetails records entry's form when form is not empty and its
// earlier payments when paid is not nil, which entry then holds.
func (a *Accounts) recordTaxDetails(ctx context.Context, entry *AccountEntry, form jurisdiction.AccountForm, paid *kernel.Money) error {
	if form != "" {
		if err := a.Terms.RecordAccountForm(ctx, entry.ID, entry.Wrapper, form); err != nil {
			return fmt.Errorf("recording the form: %w", err)
		}
	}
	if paid != nil {
		if err := a.Terms.RecordEarlierPayments(ctx, entry.ID, entry.Wrapper, *paid); err != nil {
			return fmt.Errorf("recording the earlier payments: %w", err)
		}
		entry.EarlierPayments = paid
	}
	return nil
}
