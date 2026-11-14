package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Errors of the owner's early-redemption fee.
var (
	// ErrUnknownBond reports a bond whose series terms are not stored.
	ErrUnknownBond = errors.New("no stored terms of the treasury bond")
	// errNoBondFees reports settings built without WithBondFees.
	errNoBondFees = errors.New("no store of bond fees")
)

// BondFees reads a treasury bond's stored series terms and changes its
// early-redemption fee (D-97); *valuation/postgres.Store implements it.
type BondFees interface {
	// BondTerms returns the series terms of the bond with isin, and false
	// when none are stored.
	BondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error)
	// SaveBondFee stores fee, from source, as the early-redemption fee of
	// the bond with isin and enqueues revals in one transaction.
	SaveBondFee(ctx context.Context, isin ledger.ISIN, fee kernel.Money, source pl.FeeSource, revals []valuation.Revaluation) error
}

// WithBondFees has s confirm and override bonds' early-redemption fees in
// fees (SetEarlyRedemptionFee), and returns s.
func (s *Settings) WithBondFees(fees BondFees) *Settings {
	s.fees = fees
	return s
}

// SetEarlyRedemptionFee confirms or overrides, for the owner, the
// early-redemption fee of the treasury bond with isin (D-97). An empty fee
// confirms the fee assumed from the reviewed table (pl.Terms.ConfirmFee):
// the same fee becomes the owner's and nothing is revalued. Any other fee,
// a PLN decimal of zero or more, replaces the series' fee whatever it came
// from (pl.Terms.OverrideFee), and every account that held the series is
// revalued when the amount changed. An ISIN whose terms are not stored is
// ErrUnknownBond; confirming a fee that is not assumed wraps
// pl.ErrFeeNotAssumed; a malformed ISIN or fee wraps ErrInvalidInput.
func (s *Settings) SetEarlyRedemptionFee(ctx context.Context, origin, sessionToken, isin, fee string) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if s.fees == nil {
		return errNoBondFees
	}
	parsed, err := ledger.ParseISIN(isin)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	terms, ok, err := s.fees.BondTerms(ctx, parsed)
	if err != nil {
		return fmt.Errorf("setting the early-redemption fee: %w", err)
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownBond, parsed)
	}
	next, err := feeChange(&terms, fee)
	if err != nil {
		return err
	}
	var revals []valuation.Revaluation
	if !next.EarlyRedemptionFee.Equal(terms.EarlyRedemptionFee) {
		if revals, err = s.recorder.planner.ForISIN(ctx, parsed); err != nil {
			return fmt.Errorf("setting the early-redemption fee: %w", err)
		}
	}
	err = auditedSetting(ctx, s.audit, origin, sessionToken, settingEarlyRedemption, func(ctx context.Context) error {
		return s.fees.SaveBondFee(ctx, parsed, next.EarlyRedemptionFee, next.FeeSource, revals)
	}, auditInstrumentRef+parsed.String())
	if err != nil {
		return fmt.Errorf("setting the early-redemption fee of %s: %w", parsed, err)
	}
	return nil
}

// feeChange returns terms with the owner's fee: the assumed one confirmed
// when fee is empty, otherwise fee, a PLN decimal, in its place.
func feeChange(terms *pl.Terms, fee string) (pl.Terms, error) {
	if fee == "" {
		next, err := terms.ConfirmFee()
		if err != nil {
			return pl.Terms{}, fmt.Errorf("confirming the early-redemption fee: %w", err)
		}
		return next, nil
	}
	m, err := kernel.ParseMoney(fee, pl.Nominal().Currency())
	if err != nil {
		return pl.Terms{}, fmt.Errorf("%w: early-redemption fee: %w", ErrInvalidInput, err)
	}
	next, err := terms.OverrideFee(m)
	if err != nil {
		return pl.Terms{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	return next, nil
}
