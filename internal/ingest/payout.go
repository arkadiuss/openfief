package ingest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TreasuryBonds is the institution of hand-entered treasury bond accounts,
// which the treasury-bonds connector feeds from the lots the owner enters
// (D-98, D-118).
const TreasuryBonds InstitutionID = "treasury-bonds"

// ErrPayoutNotBondAccount reports a paid-out payout setting on an account
// that is not a hand-entered bond account.
var ErrPayoutNotBondAccount = errors.New("only a hand-entered bond account can pay out its coupons and maturities")

// CheckPayout reports whether an account at institution may have payout
// (D-135): any account may keep its payouts as cash, and only a
// hand-entered bond account, at TreasuryBonds, may pay them out, matching
// the option the bond registry offers.
func CheckPayout(institution InstitutionID, payout ledger.Payout) error {
	switch payout {
	case ledger.PayoutCash:
		return nil
	case ledger.PayoutPaidOut:
		if institution != TreasuryBonds {
			return fmt.Errorf("%w: %s is at %s", ErrPayoutNotBondAccount, payout, institution)
		}
		return nil
	}
	return fmt.Errorf("%w: %d", ledger.ErrUnknownPayout, int(payout))
}

// PayoutSuffix ends the external ID of the withdrawal that pays out a
// coupon or maturity: the coupon's or maturity's own ID plus this suffix
// (D-135).
const PayoutSuffix = ":payout"

// ErrPayoutFromFuture reports a payout change that would take effect after
// today: the owner gives today or an earlier day (D-162).
var ErrPayoutFromFuture = errors.New("a payout setting takes effect today or earlier")

// CheckPayoutFrom reports whether a payout change may take effect from the
// day from, given today: today or any earlier day (D-162).
func CheckPayoutFrom(from, today kernel.Date) error {
	if from.After(today) {
		return fmt.Errorf("%w: %s is after %s", ErrPayoutFromFuture, from, today)
	}
	return nil
}

// StalePayouts returns the deletions that supersede the payout withdrawals
// among current, an account's current transactions, that its payout
// setting s no longer makes after a change in force from the day from
// (D-162): each withdrawal keyed with PayoutSuffix dated on or after from,
// on a day s keeps what falls due as cash. The next sync books the coupons
// and maturities from then on anew, with a withdrawal where s pays them
// out, so a change in either direction is complete once both happened.
// Withdrawals dated before from were booked under the setting then in
// force and stay.
func StalePayouts(current []ledger.Transaction, s ledger.PayoutSetting, from kernel.Date) ([]ledger.Transaction, error) {
	var out []ledger.Transaction
	for i := range current {
		t := &current[i]
		if t.Kind != ledger.KindWithdrawal || !strings.HasSuffix(t.Key.ExternalID, PayoutSuffix) ||
			t.Date.Before(from) || s.PaidOut(t.Date) {
			continue
		}
		d, err := ledger.NewDeletion(ledger.TransactionHeader{Key: t.Key, Account: t.Account, Date: t.Date})
		if err != nil {
			return nil, fmt.Errorf("superseding %s: %w", t.Key, err)
		}
		d.Description = "Payout setting changed: " + t.Description
		out = append(out, d)
	}
	return out, nil
}
