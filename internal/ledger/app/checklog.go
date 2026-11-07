package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// CheckLog checks the whole transaction log, every stored version of every
// account's transactions, current or superseded (ledger.CheckLog, D-19,
// D-200 criterion 2). It reads each stored account's log; a version a
// correction moved between two accounts is in both logs and counts once.
func (l *Ledger) CheckLog(ctx context.Context) (ledger.LogCheck, error) {
	accounts, err := l.store.Accounts(ctx)
	if err != nil {
		return ledger.LogCheck{}, fmt.Errorf("loading accounts: %w", err)
	}
	var all []ledger.VersionedTransaction
	for i := range accounts {
		log, err := l.store.Log(ctx, accounts[i].ID)
		if err != nil {
			return ledger.LogCheck{}, fmt.Errorf("loading the log of %s: %w", accounts[i].ID, err)
		}
		all = append(all, log...)
	}
	return ledger.CheckLog(all), nil
}
