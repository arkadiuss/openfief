// Package connectapi serves the ingest context's BondLotService
// (proto/openfief/ingest/v1/bond_lots.proto) and StatementService
// (statements.proto) to view plugins over Connect: the built-in UI hands
// each treasury bond lot the owner enters, and streams each statement file
// the owner uploads, to the core through them (D-7, D-9, D-49). Callers are
// admitted by the interceptor the caller passes, cmd/core's view-plugin
// check; the owner's session travels in the request and the use case
// checks its scope. Errors map to Connect codes: no valid session is
// unauthenticated, a session that is not the owner's permission_denied, a
// lot that breaks a rule or a file the connector refuses invalid_argument
// with the reason, a file over the limit resource_exhausted, an account
// without its one connector failed_precondition, a connector that can't be
// reached unavailable, anything else internal with nothing of the error.
package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// errInternal is what a caller sees of an internal failure.
var errInternal = errors.New("internal error")

// percent is the factor that turns a percentage into a fraction: 6.80%
// is 0.068.
func percent() kernel.Decimal { return kernel.NewDecimal(1, -2) }

// New returns the BondLotService's path and HTTP handler over lots, with
// callers in front of every call to admit view plugins only, logging to
// log.
func New(lots *app.BondLots, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewBondLotServiceHandler(&service{lots: lots, log: log}, connect.WithInterceptors(callers))
}

// service implements ingestv1connect.BondLotServiceHandler.
type service struct {
	lots *app.BondLots
	log  *slog.Logger
}

// AddBondLot hands one lot the owner entered to the account's connector.
func (s *service) AddBondLot(ctx context.Context, req *connect.Request[ingestv1.AddBondLotRequest]) (*connect.Response[ingestv1.AddBondLotResponse], error) {
	lot, err := lotFrom(req.Msg.GetLot())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	b, err := s.lots.Add(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), ledger.AccountID(req.Msg.GetAccountId()), &lot)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.AddBondLotResponse{
		BatchId: string(b.ID), New: count(b.Imported.New), NewVersions: count(b.Imported.NewVersions), Unchanged: count(b.Imported.Unchanged),
	}), nil
}

// count returns a batch's count as the message's int32; a count beyond
// int32 is capped, which one lot's batch never reaches.
func count(n int) int32 {
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	if n < 0 {
		return 0
	}
	return int32(n)
}

// fail maps err to its Connect error. A broken rule or an account without
// its connector carries its text, which names the rule, the series and the
// account, so the owner can fix it; it never carries an amount.
func (s *service) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ingest.ErrInvalidBondLot), errors.Is(err, app.ErrEntryRefused):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrNoConnector), errors.Is(err, app.ErrNoEntries), errors.Is(err, ingest.ErrAccountNotBound):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, app.ErrConnectorUnavailable):
		return connect.NewError(connect.CodeUnavailable, app.ErrConnectorUnavailable)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "bond lot call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// lotFrom reads a lot message: rates in percent become fractions, the fee
// is in PLN. A field that doesn't parse is an error wrapping
// ingest.ErrInvalidBondLot; the rules are checked by the use case.
func lotFrom(pb *ingestv1.BondLot) (ingest.BondLot, error) {
	series, err := pl.ParseSeries(pb.GetSeries())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: series: %w", ingest.ErrInvalidBondLot, err)
	}
	q, err := kernel.ParseQuantity(pb.GetQuantity())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: quantity: %w", ingest.ErrInvalidBondLot, err)
	}
	purchased, err := kernel.ParseDate(pb.GetPurchaseDate())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: purchase date: %w", ingest.ErrInvalidBondLot, err)
	}
	first, err := kernel.ParseDecimal(pb.GetFirstPeriodRatePercent())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: first-period rate: %w", ingest.ErrInvalidBondLot, err)
	}
	margin, err := kernel.ParseDecimal(pb.GetMarginPercent())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: margin: %w", ingest.ErrInvalidBondLot, err)
	}
	fee, err := kernel.ParseMoney(pb.GetEarlyRedemptionFeePln(), pl.Nominal().Currency())
	if err != nil {
		return ingest.BondLot{}, fmt.Errorf("%w: early-redemption fee: %w", ingest.ErrInvalidBondLot, err)
	}
	return ingest.BondLot{
		Terms: pl.Terms{
			Series: series, FirstPeriodRate: first.Mul(percent()), Margin: margin.Mul(percent()), EarlyRedemptionFee: fee,
		},
		Quantity: q, Purchased: purchased,
	}, nil
}
