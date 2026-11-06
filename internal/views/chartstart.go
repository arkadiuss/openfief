package views

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// AccountHistory is how far back an account's record goes (D-138): its
// first day, whether and on which day it was seeded from a starting
// balance (D-23), before which its history is unknown, and whether the
// owner marked it as opened on its seeding day (D-168), when nothing
// before its seeding is unknown.
type AccountHistory struct {
	// First is the date of its earliest current transaction.
	First kernel.Date
	// Seeded reports a current seeding transaction, an opening lot or
	// opening cash (ledger.IsSeeding).
	Seeded bool
	// SeededOn is the latest date of a seeding transaction; zero when not
	// Seeded.
	SeededOn kernel.Date
	// OpenedOnSeedingDay is the owner's mark that the account was opened on
	// its seeding day (ledger.Account.OpenedOnSeedingDay): its seeding
	// then doesn't move the portfolio's default start.
	OpenedOnSeedingDay bool
}

// HistoryOf returns the history of an account with the current
// transactions txs, which are not empty.
func HistoryOf(txs []ledger.Transaction) AccountHistory {
	h := AccountHistory{First: txs[0].Date}
	for i := range txs {
		t := &txs[i]
		if t.Date.Before(h.First) {
			h.First = t.Date
		}
		if ledger.IsSeeding(t) && (!h.Seeded || t.Date.After(h.SeededOn)) {
			h.Seeded, h.SeededOn = true, t.Date
		}
	}
	return h
}

// ChartStart returns the portfolio's charts' default start (D-138), the day
// from which every account's history is known: the latest day an account
// was seeded from a starting balance (D-23) and not marked as opened on
// its seeding day (D-168). An account booked from its whole history,
// opened later, or so marked doesn't move it; with no account that moves
// it, it is the first day of any. It is also where tracking began, which
// the charts mark. It returns false for no accounts.
func ChartStart(accounts []AccountHistory) (kernel.Date, bool) {
	var start, first kernel.Date
	seeded, found := false, false
	for _, a := range accounts {
		if !found || a.First.Before(first) {
			first = a.First
		}
		found = true
		if a.Seeded && !a.OpenedOnSeedingDay && (!seeded || a.SeededOn.After(start)) {
			start, seeded = a.SeededOn, true
		}
	}
	if !seeded {
		return first, found
	}
	return start, true
}

// ChosenStart resolves the start the owner chose for a chart whose record
// begins on first and whose default start is def (D-138): a zero chosen is
// the default, and a day before first starts on first, since nothing is
// recorded earlier.
func ChosenStart(chosen, first, def kernel.Date) kernel.Date {
	if chosen.IsZero() {
		return def
	}
	if chosen.Before(first) {
		return first
	}
	return chosen
}
