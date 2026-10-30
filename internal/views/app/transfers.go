package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// Transfers are the transfers between the owner's own accounts (D-105):
// those suggested for the owner to confirm or reject, and the pairs the
// owner confirmed or rejected, each in the order views.SuggestTransfers and
// the links give.
type Transfers struct {
	Suggested []views.Transfer
	Confirmed []views.Transfer
	Rejected  []views.Transfer
}

// Transfers returns the suggested, confirmed and rejected transfers among
// the current deposits and withdrawals of every account. Only the confirmed
// ones are left out of the time-weighted return's cash flows
// (Performance).
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
	}, nil
}
