package views

import (
	"cmp"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// cashToleranceExponent makes the cash tolerance 0.01 of a balance
// assertion's currency (D-200, criterion 1).
const cashToleranceExponent = -2

// CashTolerance returns how far a cash balance assertion's derived balance
// may be from the reported one and still pass the balance check: 0.01 of
// its currency (D-200, criterion 1).
func CashTolerance() kernel.Decimal { return kernel.NewDecimal(1, cashToleranceExponent) }

// BalanceCheck is a balance assertion checked as the owner's balance check
// checks it (D-200, criterion 1): the reported balance, the balance derived
// from the transaction log, their difference and whether it passes.
type BalanceCheck struct {
	Assertion ledger.BalanceAssertion
	// Derived is the balance derived from the log at the end of the
	// assertion's date, in its unit: an amount of the cash currency, or
	// units of the instrument.
	Derived kernel.Decimal
	// Difference is the reported balance minus the derived one.
	Difference kernel.Decimal
	// Passed reports whether the difference is within CashTolerance for
	// cash, or zero for a holding.
	Passed bool
}

// CheckBalance checks a against the derived balance: a cash assertion
// passes when the two differ by at most CashTolerance of its currency, a
// holding assertion only when the units are equal, since its units carry
// no currency to measure 0.01 in.
func CheckBalance(a *ledger.BalanceAssertion, derived kernel.Decimal) BalanceCheck {
	diff := a.Balance.Sub(derived)
	passed := diff.IsZero()
	if a.Book.Kind == ledger.BookCash {
		abs := diff
		if abs.Sign() < 0 {
			abs = abs.Neg()
		}
		passed = abs.Cmp(CashTolerance()) <= 0
	}
	return BalanceCheck{Assertion: *a, Derived: derived, Difference: diff, Passed: passed}
}

// LatestBalanceChecks returns, of checks, the latest one of each account
// and ledger account: what the institution reported last for each
// position and cash balance. The result is in account, then ledger account
// order.
func LatestBalanceChecks(checks []BalanceCheck) []BalanceCheck {
	type key struct {
		account ledger.AccountID
		book    ledger.Book
	}
	latest := map[key]BalanceCheck{}
	for i := range checks {
		a := &checks[i].Assertion
		k := key{a.Account, a.Book}
		if have, ok := latest[k]; !ok || have.Assertion.Date.Before(a.Date) {
			latest[k] = checks[i]
		}
	}
	out := make([]BalanceCheck, 0, len(latest))
	for k := range latest {
		out = append(out, latest[k])
	}
	SortBalanceChecks(out)
	return out
}

// SortBalanceChecks orders checks by account, then date, then ledger
// account.
func SortBalanceChecks(checks []BalanceCheck) {
	slices.SortFunc(checks, func(x, y BalanceCheck) int {
		a, b := &x.Assertion, &y.Assertion
		return cmp.Or(
			cmp.Compare(a.Account, b.Account),
			a.Date.Compare(b.Date),
			cmp.Compare(a.Book.String(), b.Book.String()),
		)
	})
}

// UnsupportedRows is what a connector instance's latest completed import
// batch left out because OpenFief can't book it yet, such as open XTB CFD
// positions (D-24): flagged and unsupported, never checked. Only counts
// (GR-12).
type UnsupportedRows struct {
	Instance    string
	Institution ingest.InstitutionID
	Batch       ingest.BatchID
	Finished    time.Time
	// Accounts are the accounts the batch booked into, in ID order.
	Accounts []ledger.AccountID
	// Rows is how many rows it left out as unsupported.
	Rows int
}

// LatestUnsupported returns, for each connector instance whose latest
// completed batch among batches left rows out as unsupported, those rows,
// in instance order. The latest batch is the one that finished last, then
// the greatest ID; an instance whose latest batch left none out is not
// listed, whatever its earlier batches did.
func LatestUnsupported(batches []ingest.BatchRecord) []UnsupportedRows {
	latest := map[string]*ingest.BatchRecord{}
	for i := range batches {
		b := &batches[i]
		have, ok := latest[b.Instance]
		if !ok || have.Finished.Before(b.Finished) || (have.Finished.Equal(b.Finished) && have.ID < b.ID) {
			latest[b.Instance] = b
		}
	}
	out := []UnsupportedRows{}
	for _, b := range latest {
		if b.Unsupported > 0 {
			out = append(out, UnsupportedRows{
				Instance: b.Instance, Institution: b.Institution, Batch: b.ID, Finished: b.Finished,
				Accounts: slices.Clone(b.Accounts), Rows: b.Unsupported,
			})
		}
	}
	slices.SortFunc(out, func(x, y UnsupportedRows) int { return cmp.Compare(x.Instance, y.Instance) })
	return out
}

// BalanceReport is the owner's balance check (D-200, criterion 1): the
// balance assertions with their checks, and the rows left out as
// unsupported that no assertion covers.
type BalanceReport struct {
	Checks      []BalanceCheck
	Unsupported []UnsupportedRows
}

// Failed counts the checks that did not pass.
func (r *BalanceReport) Failed() int {
	n := 0
	for i := range r.Checks {
		if !r.Checks[i].Passed {
			n++
		}
	}
	return n
}
