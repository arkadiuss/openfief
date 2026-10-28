package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Sleeves returns every sleeve, in ID order.
func (l *Ledger) Sleeves(ctx context.Context) ([]portfolio.Sleeve, error) {
	s, err := l.store.Sleeves(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading sleeves: %w", err)
	}
	return s, nil
}

// Accounts returns every account, in ID order.
func (l *Ledger) Accounts(ctx context.Context) ([]ledger.Account, error) {
	a, err := l.store.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	return a, nil
}

// Account returns the account id, or an error wrapping ErrUnknownAccount.
func (l *Ledger) Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error) {
	return lookupAccount(ctx, l.store, id)
}

// CurrentTransactions returns the current version of each of account's
// transactions, in date then key order. Superseded versions are left out.
func (l *Ledger) CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error) {
	log, err := l.log(ctx, account)
	if err != nil {
		return nil, err
	}
	return currentIn(log, account)
}

// HoldingsOn returns account's non-zero holdings at the end of on, in ISIN
// then venue symbol order.
func (l *Ledger) HoldingsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error) {
	log, err := l.log(ctx, account)
	if err != nil {
		return nil, err
	}
	h, err := ledger.HoldingsOn(log, account, on)
	if err != nil {
		return nil, fmt.Errorf("holdings of %s on %s: %w", account, on, err)
	}
	return h, nil
}

// LotsOn returns account's open lots at the end of on: per instrument in
// ISIN then venue symbol order, oldest first within each.
func (l *Ledger) LotsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Lot, error) {
	log, err := l.log(ctx, account)
	if err != nil {
		return nil, err
	}
	holdings, err := ledger.HoldingsOn(log, account, on)
	if err != nil {
		return nil, fmt.Errorf("lots of %s on %s: %w", account, on, err)
	}
	var lots []ledger.Lot
	for i := range holdings {
		in, err := ledger.LotsOn(log, account, holdings[i].Instrument, on)
		if err != nil {
			return nil, fmt.Errorf("lots of %s on %s: %w", account, on, err)
		}
		lots = append(lots, in...)
	}
	return lots, nil
}

// CashOn returns account's non-zero cash at the end of on, one amount per
// currency, in currency code order.
func (l *Ledger) CashOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]kernel.Money, error) {
	log, err := l.log(ctx, account)
	if err != nil {
		return nil, err
	}
	txs, err := currentIn(log, account)
	if err != nil {
		return nil, err
	}
	var out []kernel.Money
	for _, c := range cashCurrencies(txs) {
		m, err := ledger.CashOn(log, account, c, on)
		if err != nil {
			return nil, fmt.Errorf("cash of %s on %s: %w", account, on, err)
		}
		if !m.IsZero() {
			out = append(out, m)
		}
	}
	return out, nil
}

// AssertionChecks returns account's balance assertions with the result of
// their latest check, in date then ledger account order.
func (l *Ledger) AssertionChecks(ctx context.Context, account ledger.AccountID) ([]AssertionCheck, error) {
	if _, err := lookupAccount(ctx, l.store, account); err != nil {
		return nil, err
	}
	c, err := l.store.AssertionChecks(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("loading balance assertions of %s: %w", account, err)
	}
	return c, nil
}

// log returns the stored log of account, which must be stored.
func (l *Ledger) log(ctx context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error) {
	if _, err := lookupAccount(ctx, l.store, account); err != nil {
		return nil, err
	}
	log, err := l.store.Log(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("loading the log of %s: %w", account, err)
	}
	return log, nil
}

// currentIn returns the current versions in log that are in account: a
// correction can move a transaction to another account.
func currentIn(log []ledger.VersionedTransaction, account ledger.AccountID) ([]ledger.Transaction, error) {
	all, err := ledger.CurrentTransactions(log)
	if err != nil {
		return nil, fmt.Errorf("transactions of %s: %w", account, err)
	}
	out := all[:0]
	for i := range all {
		if all[i].Account == account {
			out = append(out, all[i])
		}
	}
	return out, nil
}

// cashCurrencies returns the currencies of txs' cash postings, in code order.
func cashCurrencies(txs []ledger.Transaction) []kernel.Currency {
	seen := map[kernel.Currency]bool{}
	for i := range txs {
		for j := range txs[i].Postings {
			if b := txs[i].Postings[j].Book; b.Kind == ledger.BookCash {
				seen[b.Currency] = true
			}
		}
	}
	out := make([]kernel.Currency, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
