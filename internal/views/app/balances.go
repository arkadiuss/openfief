package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrOwnerOnly reports a balance check asked for by anyone but the owner.
var ErrOwnerOnly = errors.New("only the owner may check balances")

// batchPage is how many completed batches the balance check reads at a
// time.
const batchPage = 500

// BalanceReport returns the owner's balance check (D-200, criterion 1):
// every balance assertion of the accounts the authorizer lets the owner
// view (D-163), checked with views.CheckBalance against the balance the
// ledger derived at its latest check, or only the latest of each account
// and ledger account unless all is set, and the rows each connector
// instance's latest batch left out as unsupported, such as open XTB CFDs
// (D-24). The report is the owner's alone: any other caller, a script
// token, a view plugin or a demo session included, gets ErrOwnerOnly. No
// read API method serves it.
func (v *Views) BalanceReport(ctx context.Context, all bool) (views.BalanceReport, error) {
	if identityapp.CallerFrom(ctx).Holder != identity.HolderOwner {
		return views.BalanceReport{}, ErrOwnerOnly
	}
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return views.BalanceReport{}, err
	}
	checks := []views.BalanceCheck{}
	for i := range accounts {
		stored, loadErr := v.ledger.AssertionChecks(ctx, accounts[i].ID)
		if loadErr != nil {
			return views.BalanceReport{}, fmt.Errorf("loading balance assertions of %s: %w", accounts[i].ID, loadErr)
		}
		for j := range stored {
			checks = append(checks, views.CheckBalance(&stored[j].Assertion, stored[j].Derived))
		}
	}
	if all {
		views.SortBalanceChecks(checks)
	} else {
		checks = views.LatestBalanceChecks(checks)
	}
	unsupported, err := v.latestUnsupported(ctx)
	if err != nil {
		return views.BalanceReport{}, err
	}
	return views.BalanceReport{Checks: checks, Unsupported: unsupported}, nil
}

// latestUnsupported returns views.LatestUnsupported of every completed
// batch, those tied to an account the caller may view; none without
// SyncHistory.
func (v *Views) latestUnsupported(ctx context.Context) ([]views.UnsupportedRows, error) {
	if v.syncs == nil {
		return []views.UnsupportedRows{}, nil
	}
	var batches []ingest.BatchRecord
	q := ingest.BatchQuery{Limit: batchPage}
	for {
		page, err := v.syncs.CompletedBatches(ctx, &q)
		if err != nil {
			return nil, fmt.Errorf("loading import batches: %w", err)
		}
		batches = append(batches, page...)
		if len(page) < batchPage {
			break
		}
		last := &page[len(page)-1]
		q.Before, q.BeforeID = last.Finished, last.ID
	}
	rows := views.LatestUnsupported(batches)
	out := make([]views.UnsupportedRows, 0, len(rows))
	for i := range rows {
		ok, err := v.mayViewAny(ctx, rows[i].Accounts)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, rows[i])
		}
	}
	return out, nil
}
