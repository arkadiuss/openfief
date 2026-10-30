package views

import (
	"cmp"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

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

// Transfer is a withdrawal from one of the owner's accounts paired with a
// deposit of the same amount into another (D-105): a suggestion, or a pair
// the owner confirmed or rejected. A confirmed one is money moving inside
// the portfolio, so neither leg is a cash flow (D-61).
type Transfer struct {
	Withdrawal CashFlow
	Deposit    CashFlow
}

// Pair returns the transaction keys of t's legs.
func (t *Transfer) Pair() ledger.TransferPair {
	return ledger.TransferPair{Withdrawal: t.Withdrawal.Key, Deposit: t.Deposit.Key}
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

// LinkedTransfers returns the pairs among links whose current decision is
// decision (D-105), with their legs found among flows, in the order of
// links. A link counts only while both legs are among flows as a
// withdrawal and a deposit of the opposite amount in different accounts,
// neither opening cash: a correction of a leg that breaks this leaves the
// link out.
func LinkedTransfers(flows []CashFlow, links []ledger.TransferLink, decision ledger.TransferDecision) []Transfer {
	byKey := make(map[ledger.TransactionKey]int, len(flows))
	for i := range flows {
		byKey[flows[i].Key] = i
	}
	var out []Transfer
	for _, l := range links {
		if l.Decision != decision {
			continue
		}
		wi, okW := byKey[l.Pair.Withdrawal]
		di, okD := byKey[l.Pair.Deposit]
		if okW && okD && legsMatch(&flows[wi], &flows[di]) {
			out = append(out, Transfer{Withdrawal: flows[wi], Deposit: flows[di]})
		}
	}
	return out
}

// ConfirmedTransfers returns the transfers between the owner's own accounts
// that the owner confirmed (D-105) and the flows left over, the portfolio's
// external cash flows, in the order of flows. Only confirmed transfers stay
// out of the cash flows: a suggested or rejected pair is two cash flows.
func ConfirmedTransfers(flows []CashFlow, links []ledger.TransferLink) ([]Transfer, []CashFlow) {
	transfers := LinkedTransfers(flows, links, ledger.TransferConfirmed)
	legs := legKeys(transfers)
	external := make([]CashFlow, 0, len(flows))
	for i := range flows {
		if !legs[flows[i].Key] {
			external = append(external, flows[i])
		}
	}
	return transfers, external
}

// SuggestTransfers suggests transfers between the owner's own accounts
// (D-105) among flows, given the owner's links: each withdrawal that is not
// a leg of a confirmed transfer, in the order of flows, paired with the
// deposit into another account of exactly the same amount and currency,
// dated at most ledger.TransferWindowDays before or after it, that is
// nearest in date, the earliest of equally near ones. The deposit must not
// be a leg of a confirmed transfer or of an earlier suggestion, and the
// pair must not be rejected. Opening cash is never a leg.
func SuggestTransfers(flows []CashFlow, links []ledger.TransferLink) []Transfer {
	used := legKeys(LinkedTransfers(flows, links, ledger.TransferConfirmed))
	rejected := map[ledger.TransferPair]bool{}
	for _, l := range links {
		if l.Decision == ledger.TransferRejected {
			rejected[l.Pair] = true
		}
	}
	var out []Transfer
	for i := range flows {
		w := &flows[i]
		if w.Seeding || w.Amount.Sign() >= 0 || used[w.Key] {
			continue
		}
		if j := nearestDeposit(flows, w, func(d *CashFlow) bool {
			return used[d.Key] || rejected[ledger.TransferPair{Withdrawal: w.Key, Deposit: d.Key}]
		}); j >= 0 {
			used[w.Key], used[flows[j].Key] = true, true
			out = append(out, Transfer{Withdrawal: *w, Deposit: flows[j]})
		}
	}
	return out
}

// nearestDeposit returns the index in flows of the deposit that can be the
// other leg of withdrawal w (legsMatch), dated at most
// ledger.TransferWindowDays from it and not skipped, nearest in date, the
// first of equally near ones, or -1 when there is none.
func nearestDeposit(flows []CashFlow, w *CashFlow, skip func(d *CashFlow) bool) int {
	best, bestGap := -1, 0
	for j := range flows {
		d := &flows[j]
		if !legsMatch(w, d) || skip(d) {
			continue
		}
		gap := max(w.Date.DaysUntil(d.Date), d.Date.DaysUntil(w.Date))
		if gap <= ledger.TransferWindowDays && (best < 0 || gap < bestGap) {
			best, bestGap = j, gap
		}
	}
	return best
}

// legsMatch reports whether w and d can be the withdrawal and the deposit
// of one transfer: neither opening cash, in different accounts, w taking out
// exactly what d puts in, in the same currency.
func legsMatch(w, d *CashFlow) bool {
	return !w.Seeding && !d.Seeding && w.Account != d.Account && w.Amount.Sign() < 0 && d.Amount.Equal(w.Amount.Neg())
}

// legKeys returns the keys of the legs of transfers.
func legKeys(transfers []Transfer) map[ledger.TransactionKey]bool {
	out := make(map[ledger.TransactionKey]bool, len(transfers))
	for i := range transfers {
		out[transfers[i].Withdrawal.Key], out[transfers[i].Deposit.Key] = true, true
	}
	return out
}
