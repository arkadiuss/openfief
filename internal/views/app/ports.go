package app

import (
	"context"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/views"
)

// Ledger is what the views read from the ledger context: its queries over
// the transaction log. *ledgerapp.Ledger implements it. Queries for an
// account that is not stored fail with an error wrapping
// ledgerapp.ErrUnknownAccount.
type Ledger interface {
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
	// Account returns the account id.
	Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error)
	// CurrentTransactions returns the current version of each of account's
	// transactions, in date then key order.
	CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error)
	// HoldingsOn returns account's non-zero holdings at the end of on.
	HoldingsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error)
	// LotsOn returns account's open lots at the end of on.
	LotsOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Lot, error)
	// CashOn returns account's non-zero cash at the end of on, one amount per
	// currency, in currency code order.
	CashOn(ctx context.Context, account ledger.AccountID, on kernel.Date) ([]kernel.Money, error)
	// AssertionChecks returns account's balance assertions with the result
	// of their latest check, in date then ledger account order.
	AssertionChecks(ctx context.Context, account ledger.AccountID) ([]ledgerapp.AssertionCheck, error)
}

// Valuations reads the stored valuations (D-36).
type Valuations interface {
	// HoldingValuations returns account's stored valuations dated from to to,
	// both inclusive, in date then instrument order, with each value in its
	// original currency and in the reporting currency (D-62). An account with
	// none returns none.
	HoldingValuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) ([]views.HoldingValuation, error)
}

// ValuationStatus says how current an account's stored valuations are
// (D-36).
type ValuationStatus struct {
	Account ledger.AccountID
	// ValuedThrough is the last date the account's valuations are complete
	// for; zero until it has been valued.
	ValuedThrough kernel.Date
	// Pending reports a revaluation queued and not started.
	Pending bool
	// Running reports a revaluation in progress.
	Running bool
}

// ValuationStatuses reads where each account's revaluation stands (D-36).
type ValuationStatuses interface {
	// ValuationStatus returns account's valuation status; an account never
	// valued has a zero ValuedThrough.
	ValuationStatus(ctx context.Context, account ledger.AccountID) (ValuationStatus, error)
}

// FXRates reads stored FX rates to PLN (D-20).
type FXRates interface {
	// LatestRates returns, for each currency that has one, the latest rate
	// dated on or before on.
	LatestRates(ctx context.Context, on kernel.Date) ([]valuation.FXRate, error)
}
