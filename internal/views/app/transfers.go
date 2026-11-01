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
// among the current deposits and withdrawals of every account. Only the
// confirmed ones that still match are left out of the time-weighted
// return's cash flows (Performance).
func (v *Views) Transfers(ctx context.Context) (Transfers, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return Transfers{}, err
	}
	var txs []ledger.Transaction
	for i := range accounts {
		got, loadErr := v.ledger.CurrentTransactions(ctx, accounts[i].ID)
		if loadErr != nil {
			return Transfers{}, fmt.Errorf("loading transactions: %w", loadErr)
		}
		txs = append(txs, got...)
	}
	links, err := v.ledger.TransferLinks(ctx)
	if err != nil {
		return Transfers{}, fmt.Errorf("loading transfer links: %w", err)
	}
	flows := views.ExternalFlows(txs)
	return Transfers{
		Suggested: views.SuggestTransfers(flows, links),
		Confirmed: views.LinkedTransfers(flows, links, ledger.TransferConfirmed),
		Rejected:  views.LinkedTransfers(flows, links, ledger.TransferRejected),
		Broken:    brokenTransfers(flows, txs, ledger.BrokenTransferLinks(links, txs)),
	}, nil
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
