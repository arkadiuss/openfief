package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Planner decides which accounts an input change affects, and from which
// date. The rules are fixed, not configurable (D-36):
//
//   - a price: every account whose log ever held the instrument, from the
//     price's date;
//   - an FX rate or a period index reading: every account, from the rate's
//     date or the first day of the reading's month;
//   - a bond's series terms: every account whose log ever held an
//     instrument with that ISIN, from its first transaction;
//   - an account's tax wrapper: that account, from its first transaction;
//   - the bond basis: every account, from its first transaction;
//   - the end-of-day schedule: every account, from its valued-through date,
//     or its first transaction if it was never valued;
//   - an import: every account whose current transactions it changed, from
//     the earliest changed date (ledger/app.ImportSummary.EarliestAffected).
type Planner struct {
	accounts AccountLister
	log      TransactionLog
}

// NewPlanner returns a Planner over the ledger's accounts and log.
func NewPlanner(accounts AccountLister, log TransactionLog) *Planner {
	return &Planner{accounts: accounts, log: log}
}

// accountLog is one account's log.
type accountLog struct {
	id  ledger.AccountID
	log []ledger.VersionedTransaction
}

// ForInstrument returns a revaluation, from the date from, of every account whose log
// has a holding posting of instrument (ledger.Instrument.String()).
func (p *Planner) ForInstrument(ctx context.Context, instrument string, from kernel.Date) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		if held(a.log, func(in ledger.Instrument) bool { return in.String() == instrument }) {
			out = append(out, valuation.Revaluation{Account: a.id, From: from})
		}
	}
	return out, nil
}

// ForISIN returns a revaluation from its first transaction of every
// account whose log has a holding posting of an instrument with isin.
func (p *Planner) ForISIN(ctx context.Context, isin ledger.ISIN) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		if held(a.log, func(in ledger.Instrument) bool { return in.ISIN.Equal(isin) }) {
			first, _ := firstDate(a.log)
			out = append(out, valuation.Revaluation{Account: a.id, From: first})
		}
	}
	return out, nil
}

// ForAll returns a revaluation, from the date from, of every account with a log.
func (p *Planner) ForAll(ctx context.Context, from kernel.Date) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		if len(a.log) > 0 {
			out = append(out, valuation.Revaluation{Account: a.id, From: from})
		}
	}
	return out, nil
}

// FromStart returns a revaluation of every account with a log, or only of
// account when it is not empty, from its first transaction.
func (p *Planner) FromStart(ctx context.Context, account ledger.AccountID) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		if first, ok := firstDate(a.log); ok && (account == "" || a.id == account) {
			out = append(out, valuation.Revaluation{Account: a.id, From: first})
		}
	}
	return out, nil
}

// EndOfDay returns a revaluation of every account with a log from its
// valued-through date in store, or from its first transaction when it was
// never valued.
func (p *Planner) EndOfDay(ctx context.Context, store ValuationStore) ([]valuation.Revaluation, error) {
	logs, err := p.logs(ctx)
	if err != nil {
		return nil, err
	}
	var out []valuation.Revaluation
	for _, a := range logs {
		first, ok := firstDate(a.log)
		if !ok {
			continue
		}
		through, valued, err := store.ValuedThrough(ctx, a.id)
		if err != nil {
			return nil, fmt.Errorf("valued-through date of %s: %w", a.id, err)
		}
		if valued && through.After(first) {
			first = through
		}
		out = append(out, valuation.Revaluation{Account: a.id, From: first})
	}
	return out, nil
}

// logs returns every account with its log.
func (p *Planner) logs(ctx context.Context) ([]accountLog, error) {
	accounts, err := p.accounts.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing accounts: %w", err)
	}
	out := make([]accountLog, 0, len(accounts))
	for i := range accounts {
		log, err := p.log.Log(ctx, accounts[i].ID)
		if err != nil {
			return nil, fmt.Errorf("loading the log of %s: %w", accounts[i].ID, err)
		}
		out = append(out, accountLog{id: accounts[i].ID, log: log})
	}
	return out, nil
}

// ForImport returns a revaluation of each account in earliest from its
// date: what an import changed (ledger/app.ImportSummary.EarliestAffected).
func ForImport(earliest map[ledger.AccountID]kernel.Date) []valuation.Revaluation {
	out := make([]valuation.Revaluation, 0, len(earliest))
	for account, from := range earliest {
		out = append(out, valuation.Revaluation{Account: account, From: from})
	}
	return valuation.Coalesce(out)
}

// held reports whether log has a holding posting of an instrument that
// match accepts.
func held(log []ledger.VersionedTransaction, match func(ledger.Instrument) bool) bool {
	for i := range log {
		for _, p := range log[i].Transaction.Postings {
			if p.Book.Kind == ledger.BookHolding && match(p.Book.Instrument) {
				return true
			}
		}
	}
	return false
}
