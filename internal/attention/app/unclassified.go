package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Holdings reads what the check of unclassified instruments needs: each
// account's stored valuations and the owner's classes;
// *valuation/postgres.Store implements it.
type Holdings interface {
	// ValuedThrough returns the last day account's holdings are valued
	// through; ok is false before its first revaluation.
	ValuedThrough(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error)
	// Valuations returns account's stored valuations dated from..to.
	Valuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) ([]valuation.HoldingValuation, error)
	// ExemptionClasses returns the owner's class of each instrument they
	// classified, by ISIN (D-132).
	ExemptionClasses(ctx context.Context) (map[string]jurisdiction.ExemptionClass, error)
}

// UnclassifiedInstruments is the check of instruments for a yearly tax's
// exemptions (D-132, D-148): each instrument priced at market that an
// account in a wrapper with a yearly tax, such as OKI, held in the tax year
// of the day it is valued through, which pack can't classify and the owner
// hasn't, is a condition, its subject the ISIN, once however many accounts
// hold it. The owner's classification clears it, so the item resolves
// itself. An account without a recorded wrapper takes the pack's default.
func UnclassifiedInstruments(l Ledger, v Values, h Holdings, pack jurisdiction.Pack) Check {
	return Check{Kind: attention.KindInstrumentUnclassified, Find: func(ctx context.Context) ([]attention.Condition, error) {
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading accounts: %w", err)
		}
		owner, err := h.ExemptionClasses(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading exemption classes: %w", err)
		}
		detail := map[string]string{}
		for i := range accounts {
			if err := unclassifiedIn(ctx, v, h, pack, accounts[i].ID, owner, detail); err != nil {
				return nil, err
			}
		}
		out := make([]attention.Condition, 0, len(detail))
		for isin, d := range detail {
			out = append(out, attention.Condition{Kind: attention.KindInstrumentUnclassified, Subject: isin, Detail: d})
		}
		slices.SortFunc(out, func(a, b attention.Condition) int { return cmp.Compare(a.Subject, b.Subject) })
		return out, nil
	}}
}

// unclassifiedIn adds to detail, by ISIN, the instruments account held in
// the tax year it is valued through that its wrapper's yearly tax needs a
// class for and nobody gave one; an instrument already in detail keeps the
// detail it has.
func unclassifiedIn(ctx context.Context, v Values, h Holdings, pack jurisdiction.Pack, account ledger.AccountID,
	owner map[string]jurisdiction.ExemptionClass, detail map[string]string,
) error {
	w, recorded, err := v.TaxWrapper(ctx, account)
	if err != nil {
		return fmt.Errorf("loading the tax wrapper of %s: %w", account, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(pack)
	}
	if len(pack.ExemptionClasses(w)) == 0 {
		return nil
	}
	through, valued, err := h.ValuedThrough(ctx, account)
	if err != nil {
		return fmt.Errorf("loading the valued-through date of %s: %w", account, err)
	}
	if !valued {
		return nil
	}
	first, err := kernel.NewDate(through.Year(), time.January, 1)
	if err != nil {
		return fmt.Errorf("the tax year of %s: %w", account, err)
	}
	rows, err := h.Valuations(ctx, account, first, through)
	if err != nil {
		return fmt.Errorf("loading the valuations of %s: %w", account, err)
	}
	for _, isin := range valuation.Unclassified(pack, w, rows, owner) {
		if _, seen := detail[isin]; !seen {
			detail[isin] = "the jurisdiction pack can't tell this instrument's exemption class for " + string(w) + "'s yearly tax"
		}
	}
	return nil
}
