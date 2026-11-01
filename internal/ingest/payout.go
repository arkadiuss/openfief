package ingest

import (
	"errors"
	"fmt"

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
