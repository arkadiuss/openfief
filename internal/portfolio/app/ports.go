package app

import (
	"context"
	"errors"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// Errors an Owners implementation wraps, so the use cases can tell a caller
// who must log in again from one who may not do this.
var (
	// ErrNoSession reports a session token that is missing, unknown,
	// expired or revoked.
	ErrNoSession = errors.New("no valid session")
	// ErrNotOwner reports a valid session without owner scope, or an origin
	// the core does not serve the owner on.
	ErrNotOwner = errors.New("only the owner may do this")
)

// ModelStore keeps the model portfolio. *portfolio/postgres.Store
// implements it.
type ModelStore interface {
	// Model returns the stored model portfolio; one never saved is the zero
	// Model: off, with no targets (D-52).
	Model(ctx context.Context) (portfolio.Model, error)
	// SaveModel replaces the stored model portfolio, whether it is on and
	// all its targets, in one database transaction.
	SaveModel(ctx context.Context, m portfolio.Model) error
}

// MarketData reads the stored end-of-day history a backtest needs (D-53).
// *valuation/postgres.Store implements it.
type MarketData interface {
	// Prices returns every price of instrument ("ISIN/symbol") dated on or
	// before through, in date order.
	Prices(ctx context.Context, instrument string, through kernel.Date) ([]valuation.Price, error)
	// IndexLevels returns every level of the market index named index dated
	// on or before through, in date order.
	IndexLevels(ctx context.Context, index string, through kernel.Date) ([]valuation.IndexLevel, error)
	// FXRates returns every rate to PLN of currency dated on or before
	// through, in date order.
	FXRates(ctx context.Context, currency kernel.Currency, through kernel.Date) ([]valuation.FXRate, error)
}

// Owners checks that a session is the owner's, with owner scope, on the
// origin the browser uses (author decision, auth split). cmd/core
// implements it over the identity context.
type Owners interface {
	// CheckOwner returns nil for an owner-scoped session token valid on
	// origin, an error wrapping ErrNoSession or ErrNotOwner when it is not,
	// and any other error when it can't tell.
	CheckOwner(ctx context.Context, origin, sessionToken string) error
}
