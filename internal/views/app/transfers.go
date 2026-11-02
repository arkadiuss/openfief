package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// Transfers are the transfers between the owner's own accounts (D-105):
// those suggested for the owner to confirm or reject, the pairs the owner
// confirmed or rejected, each in the order views.SuggestTransfers and the
// links give, and the confirmed pairs a correction broke (D-131).
type Transfers struct {
	Suggested []views.Transfer
	Confirmed []views.Transfer
	Rejected  []views.Transfer
	// Broken are the confirmed pairs whose legs no longer match after a
	// correction (ledger.BrokenTransferLinks): they count as cash flows
	// until the owner rejects them. A leg that is no longer a deposit or
	// withdrawal carries only its key, account and date; a deleted leg
	// only its key.
	Broken []views.Transfer
}

// Transfers returns the suggested, confirmed, rejected and broken transfers
// among the current deposits and withdrawals of the accounts the caller may
// view (D-163). Only the confirmed ones that still match are left out of
// the time-weighted return's cash flows (Performance). Whether a confirmed
// pair is broken is judged over every account's transactions, so a leg in
// an account the caller may not view doesn't look deleted; such a pair is
// left out.
func (v *Views) Transfers(ctx context.Context) (Transfers, error) {
	all, err := v.ledger.Accounts(ctx)
	if err != nil {
		return Transfers{}, fmt.Errorf("loading accounts: %w", err)
	}
	accounts, err := v.viewable(ctx, all)
	if err != nil {
		return Transfers{}, err
	}
	shown := make(map[ledger.AccountID]bool, len(accounts))
	for i := range accounts {
		shown[accounts[i].ID] = true
	}
	var txs, allTxs []ledger.Transaction
	for i := range all {
		got, loadErr := v.ledger.CurrentTransactions(ctx, all[i].ID)
		if loadErr != nil {
			return Transfers{}, fmt.Errorf("loading transactions: %w", loadErr)
		}
		allTxs = append(allTxs, got...)
		if shown[all[i].ID] {
			txs = append(txs, got...)
		}
	}
	links, err := v.ledger.TransferLinks(ctx)
	if err != nil {
		return Transfers{}, fmt.Errorf("loading transfer links: %w", err)
	}
	flows := views.ExternalFlows(txs)
	broken := shownLinks(ledger.BrokenTransferLinks(links, allTxs), allTxs, shown)
	return Transfers{
		Suggested: views.SuggestTransfers(flows, links),
		Confirmed: views.LinkedTransfers(flows, links, ledger.TransferConfirmed),
		Rejected:  views.LinkedTransfers(flows, links, ledger.TransferRejected),
		Broken:    brokenTransfers(flows, txs, broken),
	}, nil
}

// shownLinks returns those of links with a leg among txs, the current
// transactions of every account, and every such leg in an account in
// shown, in their order.
func shownLinks(links []ledger.TransferLink, txs []ledger.Transaction, shown map[ledger.AccountID]bool) []ledger.TransferLink {
	accountOf := make(map[ledger.TransactionKey]ledger.AccountID, len(txs))
	for i := range txs {
		accountOf[txs[i].Key] = txs[i].Account
	}
	out := make([]ledger.TransferLink, 0, len(links))
	for _, l := range links {
		current, hidden := 0, false
		for _, k := range []ledger.TransactionKey{l.Pair.Withdrawal, l.Pair.Deposit} {
			if a, ok := accountOf[k]; ok {
				current++
				hidden = hidden || !shown[a]
			}
		}
		if current > 0 && !hidden {
			out = append(out, l)
		}
	}
	return out
}

// brokenTransfers returns broken's pairs with their legs: the cash flow of
// each leg among flows, or, for a leg that is no longer one, its key with
// the account and date of its current transaction among txs, if any.
func brokenTransfers(flows []views.CashFlow, txs []ledger.Transaction, broken []ledger.TransferLink) []views.Transfer {
	byKey := make(map[ledger.TransactionKey]views.CashFlow, len(flows)+len(txs))
	for i := range txs {
		byKey[txs[i].Key] = views.CashFlow{Account: txs[i].Account, Key: txs[i].Key, Date: txs[i].Date}
	}
	for i := range flows {
		byKey[flows[i].Key] = flows[i]
	}
	leg := func(k ledger.TransactionKey) views.CashFlow {
		if f, ok := byKey[k]; ok {
			return f
		}
		return views.CashFlow{Key: k}
	}
	out := make([]views.Transfer, 0, len(broken))
	for _, l := range broken {
		out = append(out, views.Transfer{Withdrawal: leg(l.Pair.Withdrawal), Deposit: leg(l.Pair.Deposit)})
	}
	return out
}
