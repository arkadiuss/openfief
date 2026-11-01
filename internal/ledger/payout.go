package ledger

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrUnknownPayout reports a payout setting outside the declared ones.
var ErrUnknownPayout = errors.New("unknown payout setting")

// Payout says what happens to what a hand-entered bond account's bonds pay,
// its coupons and maturities, matching the option the bond registry offers
// (D-135). The zero Payout is PayoutCash, the default.
type Payout int

// The payout settings.
const (
	// PayoutCash keeps coupons and maturities as the account's cash.
	PayoutCash Payout = iota
	// PayoutPaidOut pays them out of the account: each is followed by a
	// withdrawal of the same amount.
	PayoutPaidOut
)

// String returns the setting's key: "cash" or "paid-out".
func (p Payout) String() string {
	switch p {
	case PayoutCash:
		return "cash"
	case PayoutPaidOut:
		return "paid-out"
	}
	return fmt.Sprintf("Payout(%d)", int(p))
}

// ParsePayout parses a payout setting's key as String writes it.
func ParsePayout(key string) (Payout, error) {
	for _, p := range []Payout{PayoutCash, PayoutPaidOut} {
		if p.String() == key {
			return p, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownPayout, key)
}

// PayoutSetting is an account's payout setting and the day it took effect.
// A coupon or maturity follows the setting in force on the day it falls
// due, so changing it never rewrites what was booked before.
type PayoutSetting struct {
	Payout Payout
	// Since is the day the setting took effect; zero for an account that
	// never changed it.
	Since kernel.Date
}

// PaidOut reports whether what falls due on day is paid out: the setting
// is PayoutPaidOut and took effect on or before that day.
func (s PayoutSetting) PaidOut(day kernel.Date) bool {
	return s.Payout == PayoutPaidOut && !day.Before(s.Since)
}

// Changed returns the setting with p in force from today; setting the
// payout it already has keeps the day it took effect.
func (s PayoutSetting) Changed(p Payout, today kernel.Date) PayoutSetting {
	if p == s.Payout {
		return s
	}
	return PayoutSetting{Payout: p, Since: today}
}

// Validate reports whether the payout is a declared one and a paid-out
// setting says since when.
func (s PayoutSetting) Validate() error {
	switch s.Payout {
	case PayoutCash:
		return nil
	case PayoutPaidOut:
		if s.Since.IsZero() {
			return fmt.Errorf("%w: paid out since no day", ErrUnknownPayout)
		}
		return nil
	}
	return fmt.Errorf("%w: %d", ErrUnknownPayout, int(s.Payout))
}
