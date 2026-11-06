package views

import (
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// PairAccounts returns, for each pair links name, keyed by the pair as
// ledger.TransferPair.String writes it, which is a transfer-broken
// attention item's subject (D-131), the accounts its legs are in among txs,
// the current transactions of every account: the withdrawal's first, then
// the deposit's when it is another. A deleted leg is in no account, and a
// pair with neither leg among txs is left out.
func PairAccounts(links []ledger.TransferLink, txs []ledger.Transaction) map[string][]ledger.AccountID {
	accountOf := make(map[ledger.TransactionKey]ledger.AccountID, len(txs))
	for i := range txs {
		accountOf[txs[i].Key] = txs[i].Account
	}
	out := map[string][]ledger.AccountID{}
	for _, l := range links {
		var accounts []ledger.AccountID
		for _, k := range []ledger.TransactionKey{l.Pair.Withdrawal, l.Pair.Deposit} {
			if a, ok := accountOf[k]; ok && !slices.Contains(accounts, a) {
				accounts = append(accounts, a)
			}
		}
		if len(accounts) > 0 {
			out[l.Pair.String()] = accounts
		}
	}
	return out
}

// SuggestedPairAccounts returns, for each transfer SuggestTransfers
// suggests among the cash flows of txs, the current transactions of every
// account, given the owner's links, keyed by its pair as
// ledger.TransferPair.String writes it, which is a transfer-suggested
// attention item's subject (D-105, D-130): the withdrawal's account, then
// the deposit's.
func SuggestedPairAccounts(txs []ledger.Transaction, links []ledger.TransferLink) map[string][]ledger.AccountID {
	out := map[string][]ledger.AccountID{}
	suggested := SuggestTransfers(ExternalFlows(txs), links)
	for i := range suggested {
		t := &suggested[i]
		out[t.Pair().String()] = []ledger.AccountID{t.Withdrawal.Account, t.Deposit.Account}
	}
	return out
}
