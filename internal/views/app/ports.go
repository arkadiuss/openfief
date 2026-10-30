package app

import (
	"context"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
)

// Ledger is what the views read from the ledger context: its queries over
// the transaction log. *ledgerapp.Ledger implements it. Queries for an
// account that is not stored fail with an error wrapping
// ledgerapp.ErrUnknownAccount.
type Ledger interface {
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// Kinds returns every sleeve kind, in key order.
	Kinds(ctx context.Context) ([]portfolio.Kind, error)
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

// Valuations reads the stored holding valuations (D-36).
// *valuation/postgres.Store implements it.
type Valuations interface {
	// Valuations returns account's holding valuations dated from to to, both
	// inclusive, in date then instrument order, valued or lacking an input,
	// each with its value in its original currency and in PLN. An account
	// with none returns none.
	Valuations(ctx context.Context, account ledger.AccountID, from, to kernel.Date) ([]valuation.HoldingValuation, error)
}

// ValuationStatuses reads where each account's revaluation stands (D-36).
// *valuation/postgres.Store implements it.
type ValuationStatuses interface {
	// Status returns account's valuation status: the date it is valued
	// through, zero when never valued, and whether a revaluation is pending
	// or running.
	Status(ctx context.Context, account ledger.AccountID) (valuationapp.ValuationStatus, error)
}

// FXRates reads stored FX rates to PLN (D-20).
// *valuation/postgres.Store implements it.
type FXRates interface {
	// FXRates returns every rate of currency dated on or before through, in
	// date order.
	FXRates(ctx context.Context, currency kernel.Currency, through kernel.Date) ([]valuation.FXRate, error)
}
