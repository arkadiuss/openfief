package app

import (
	"context"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TransactionLog reads an account's versioned transaction log, which the
// ledger owns. The ledger's store (ledger/app.TransactionLog) satisfies it.
type TransactionLog interface {
	// Log returns every stored version of every key that has a version in
	// account.
	Log(ctx context.Context, account ledger.AccountID) ([]ledger.VersionedTransaction, error)
}

// AccountLister lists the ledger's accounts; ledger/app.Ledger satisfies it.
type AccountLister interface {
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
}

// Inputs reads the stored inputs of a revaluation besides the log.
type Inputs interface {
	// Prices returns every price of instrument (ledger.Instrument.String())
	// dated on or before through.
	Prices(ctx context.Context, instrument string, through kernel.Date) ([]valuation.Price, error)
	// FXRates returns every rate of currency dated on or before through.
	FXRates(ctx context.Context, currency kernel.Currency, through kernel.Date) ([]valuation.FXRate, error)
	// IndexReadings returns every period index reading.
	IndexReadings(ctx context.Context) ([]valuation.IndexReading, error)
	// BondTerms returns the series terms of the retail treasury bond with
	// isin, and false when isin is not a bond with stored terms.
	BondTerms(ctx context.Context, isin ledger.ISIN) (pl.Terms, bool, error)
	// TaxWrapper returns account's tax wrapper: pl.Taxable unless one was
	// recorded.
	TaxWrapper(ctx context.Context, account ledger.AccountID) (pl.TaxWrapper, error)
}

// ValuationStore stores valuations and each account's valued-through date.
type ValuationStore interface {
	// ValuedThrough returns account's valued-through date, and false when
	// the account was never revalued.
	ValuedThrough(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error)
	// ReplaceValuations removes account's valuations dated on or after from,
	// stores rows and records through as its valued-through date, all in one
	// database transaction.
	ReplaceValuations(ctx context.Context, account ledger.AccountID, from kernel.Date, rows []valuation.HoldingValuation, through kernel.Date) error
}

// Store is what Revalue reads and writes.
type Store interface {
	Inputs
	ValuationStore
}

// Recorder stores changed inputs and enqueues the revaluations they cause
// in the same database transaction (D-36): both commit, or neither does.
// Upserting replaces a stored row with the same key.
type Recorder interface {
	// SavePrices upserts prices, keyed by instrument and date.
	SavePrices(ctx context.Context, prices []valuation.Price, revals []valuation.Revaluation) error
	// SaveFXRates upserts rates, keyed by currency and date.
	SaveFXRates(ctx context.Context, rates []valuation.FXRate, revals []valuation.Revaluation) error
	// SaveIndexReadings upserts readings, keyed by kind, year and month.
	SaveIndexReadings(ctx context.Context, readings []valuation.IndexReading, revals []valuation.Revaluation) error
	// SaveBondTerms upserts the series terms of the bond with isin.
	SaveBondTerms(ctx context.Context, isin ledger.ISIN, terms *pl.Terms, revals []valuation.Revaluation) error
	// SaveTaxWrapper records account's tax wrapper.
	SaveTaxWrapper(ctx context.Context, account ledger.AccountID, wrapper pl.TaxWrapper, revals []valuation.Revaluation) error
	// SaveBondBasis records the configured bond basis and, only when it
	// differs from the one recorded before, enqueues revals. It reports
	// whether it enqueued them.
	SaveBondBasis(ctx context.Context, basis valuation.Basis, revals []valuation.Revaluation) (bool, error)
	// Enqueue enqueues revals.
	Enqueue(ctx context.Context, revals []valuation.Revaluation) error
}
