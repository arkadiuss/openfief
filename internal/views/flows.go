package views

import (
	"cmp"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TransferWindowDays is how many calendar days apart a withdrawal and a
// deposit may be dated and still be matched as one transfer between the
// owner's own accounts (MatchTransfers). Five covers a domestic transfer
// sent before a weekend and a foreign one that takes two business days.
const TransferWindowDays = 5

// CashFlow is money entering the portfolio from outside it (a positive
// amount) or leaving it (a negative one): a deposit or a withdrawal of one
// account, in the currency of its cash, on its date.
type CashFlow struct {
	Account ledger.AccountID
	Key     ledger.TransactionKey
	Date    kernel.Date
	Amount  kernel.Money
	// Seeding marks opening cash (ledger.IsSeeding): what an account held
	// at go-live, never one leg of a transfer.
	Seeding bool
}

// Transfer is a withdrawal from one of the owner's accounts matched with a
// deposit of the same amount into another (D-61): money moving inside the
// portfolio, so neither leg is a cash flow.
type Transfer struct {
	Withdrawal CashFlow
	Deposit    CashFlow
}

// InTransitOn returns the money of t that is between the two accounts at the
// end of day: its amount from the withdrawal's date until the day before the
// deposit's, or minus its amount from the deposit's date until the day before
// the withdrawal's when the deposit is booked first, so the portfolio's value
// counts the money exactly once. ok is false on any other day.
func (t *Transfer) InTransitOn(day kernel.Date) (kernel.Money, bool) {
	w, d := t.Withdrawal.Date, t.Deposit.Date
	switch {
	case w.Before(d) && !day.Before(w) && day.Before(d):
		return t.Deposit.Amount, true
	case d.Before(w) && !day.Before(d) && day.Before(w):
		return t.Deposit.Amount.Neg(), true
	}
	return kernel.Money{}, false
}

// ExternalFlows returns the deposits and withdrawals among txs, opening cash
// included, as cash flows in date, account, then key order: a deposit's
// amount is what it added to cash, a withdrawal's is minus what it took.
func ExternalFlows(txs []ledger.Transaction) []CashFlow {
	var out []CashFlow
	for i := range txs {
		t := &txs[i]
		if t.Kind != ledger.KindDeposit && t.Kind != ledger.KindWithdrawal {
			continue
		}
		for _, p := range t.Postings {
			if p.Book.Kind == ledger.BookCash {
				out = append(out, CashFlow{Account: t.Account, Key: t.Key, Date: t.Date, Amount: p.Amount, Seeding: ledger.IsSeeding(t)})
			}
		}
	}
	slices.SortFunc(out, func(a, b CashFlow) int { return compareFlows(&a, &b) })
	return out
}

// compareFlows orders cash flows by date, account, then key.
func compareFlows(a, b *CashFlow) int {
	return cmp.Or(a.Date.Compare(b.Date), cmp.Compare(a.Account, b.Account),
		cmp.Compare(a.Key.Source, b.Key.Source), cmp.Compare(a.Key.ExternalID, b.Key.ExternalID))
}

// MatchTransfers pairs withdrawals with deposits that are transfers between
// the owner's own accounts (D-61) and returns them with the flows left over,
// the portfolio's external cash flows, in the order of flows. A withdrawal
// matches a deposit into another account of exactly the same amount and
// currency dated at most window days before or after it. Opening cash never
// matches. Withdrawals are taken in the order of flows, each matched with the
// unmatched deposit nearest in date, the earliest of equally near ones, and
// each flow is matched at most once.
func MatchTransfers(flows []CashFlow, window int) ([]Transfer, []CashFlow) {
	used := make([]bool, len(flows))
	var transfers []Transfer
	for i := range flows {
		w := &flows[i]
		if w.Seeding || w.Amount.Sign() >= 0 {
			continue
		}
		if j := nearestDeposit(flows, used, w, window); j >= 0 {
			used[i], used[j] = true, true
			transfers = append(transfers, Transfer{Withdrawal: *w, Deposit: flows[j]})
		}
	}
	external := make([]CashFlow, 0, len(flows))
	for i := range flows {
		if !used[i] {
			external = append(external, flows[i])
		}
	}
	return transfers, external
}

// nearestDeposit returns the index in flows of the unused deposit that
// matches withdrawal w (MatchTransfers) nearest in date, the first of equally
// near ones, or -1 when none does.
func nearestDeposit(flows []CashFlow, used []bool, w *CashFlow, window int) int {
	best, bestGap := -1, 0
	for j := range flows {
		d := &flows[j]
		if used[j] || d.Seeding || d.Account == w.Account || !d.Amount.Equal(w.Amount.Neg()) {
			continue
		}
		gap := max(w.Date.DaysUntil(d.Date), d.Date.DaysUntil(w.Date))
		if gap <= window && (best < 0 || gap < bestGap) {
			best, bestGap = j, gap
		}
	}
	return best
}
