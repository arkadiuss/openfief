package ledger

import (
	"errors"
	"fmt"
	"slices"

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

// PayoutPeriod is a payout setting in force from Since on, until the next
// one takes over; a zero Since is from the start of the account's history.
type PayoutPeriod struct {
	Payout Payout
	Since  kernel.Date
}

// PayoutSetting is an account's payout setting, the day it took effect,
// and the settings in force before it (D-162). An account's first setting
// covers its whole history, so its Since is zero. A change takes effect
// from the day the owner gives (Changed): what falls due from then on
// follows it, and what fell due before follows the setting in force on
// its day.
type PayoutSetting struct {
	Payout Payout
	// Since is the day the setting took effect; zero for an account's
	// first setting, which covers its whole history.
	Since kernel.Date
	// Earlier are the settings in force before Since, oldest first; empty
	// for an account that never changed its setting.
	Earlier []PayoutPeriod
}

// On returns the payout setting in force on day: the latest one that took
// effect on or before it, or the earliest one for a day before all of
// them, since an account's first setting covers its whole history.
func (s PayoutSetting) On(day kernel.Date) Payout {
	all := s.periods()
	for _, period := range slices.Backward(all) {
		if period.Since.IsZero() || !day.Before(period.Since) {
			return period.Payout
		}
	}
	return all[0].Payout
}

// PaidOut reports whether what falls due on day is paid out: the setting
// in force on that day (On) is PayoutPaidOut.
func (s PayoutSetting) PaidOut(day kernel.Date) bool {
	return s.On(day) == PayoutPaidOut
}

// Changed returns the setting with p in force from the day from on
// (D-162), which may be earlier than today: the settings that took effect
// on or after from are replaced, and those before it stay in force until
// from. When the setting in force just before from is p already, nothing
// new takes effect.
func (s PayoutSetting) Changed(p Payout, from kernel.Date) PayoutSetting {
	var keep []PayoutPeriod
	for _, period := range s.periods() {
		if period.Since.IsZero() || period.Since.Before(from) {
			keep = append(keep, period)
		}
	}
	if len(keep) == 0 || keep[len(keep)-1].Payout != p {
		keep = append(keep, PayoutPeriod{Payout: p, Since: from})
	}
	last := keep[len(keep)-1]
	out := PayoutSetting{Payout: last.Payout, Since: last.Since}
	if len(keep) > 1 {
		out.Earlier = keep[:len(keep)-1]
	}
	return out
}

// Validate reports whether every setting of s is a declared one and they
// took effect in order: each one after the first on a day later than the
// one before it.
func (s PayoutSetting) Validate() error {
	all := s.periods()
	for i, period := range all {
		switch period.Payout {
		case PayoutCash, PayoutPaidOut:
		default:
			return fmt.Errorf("%w: %d", ErrUnknownPayout, int(period.Payout))
		}
		if i == 0 {
			continue
		}
		if prev := all[i-1].Since; period.Since.IsZero() || (!prev.IsZero() && !prev.Before(period.Since)) {
			return fmt.Errorf("%w: settings out of order on %s", ErrUnknownPayout, period.Since)
		}
	}
	return nil
}

// periods returns every setting of s, oldest first, the current one last.
func (s PayoutSetting) periods() []PayoutPeriod {
	return append(slices.Clone(s.Earlier), PayoutPeriod{Payout: s.Payout, Since: s.Since})
}
